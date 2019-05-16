package laboratory

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	runner "github.com/Synthace/antha-runner/export"
	"github.com/antha-lang/antha/codegen"
	"github.com/antha-lang/antha/composer"
	"github.com/antha-lang/antha/instructions"
	"github.com/antha-lang/antha/laboratory/effects"
	"github.com/antha-lang/antha/logger"
	"github.com/antha-lang/antha/target"
	"github.com/antha-lang/antha/target/human"
	"github.com/antha-lang/antha/target/mixer"
	"github.com/antha-lang/antha/target/qpcrdevice"
	"github.com/antha-lang/antha/target/shakerincubator"
	"github.com/antha-lang/antha/target/woplatereader"
	"github.com/antha-lang/antha/utils"
	"github.com/antha-lang/antha/workflow"
)

type Element interface {
	Name() workflow.ElementInstanceName
	TypeName() workflow.ElementTypeName

	Setup(*Laboratory) error
	Steps(*Laboratory) error
	Analysis(*Laboratory) error
	Validation(*Laboratory) error
}

type LaboratoryBuilder struct {
	inDir    string
	outDir   string
	Workflow *workflow.Workflow

	// elemLock must be taken to access/mutate elemsRunning and elementToLab fields
	elemLock       sync.Mutex
	elemsRunning   int
	elementToLab   map[Element]*Laboratory
	nextElementId  uint64
	elemsCompleted []*Laboratory

	// This lock is here to serialize access to errors (append and
	// Pack()), and to avoid races around closing errored chan
	errLock sync.Mutex
	errors  utils.ErrorSlice
	// errored is closed as soon as an error is recorded. This can
	// happen before Completed is closed.
	errored chan struct{}

	throttle chan struct{}

	// Completed is closed when all the elements within the workflow
	// have stopped (some may never have even started), regardless of
	// whether any error occurred.
	Completed chan struct{}

	*lineMapManager
	Logger *logger.Logger
	logFH  *os.File

	effects *effects.LaboratoryEffects

	instrs instructions.Insts
}

func EmptyLaboratoryBuilder() *LaboratoryBuilder {
	// Hopefully temporary. See AC-72
	const concurrencyLevel = 1

	labBuild := &LaboratoryBuilder{
		elementToLab: make(map[Element]*Laboratory),
		errored:      make(chan struct{}),
		throttle:     make(chan struct{}, concurrencyLevel),
		Completed:    make(chan struct{}),

		lineMapManager: NewLineMapManager(),
		Logger:         logger.NewLogger(),
	}
	token := struct{}{}
	for c := 0; c < concurrencyLevel; c++ {
		labBuild.throttle <- token
	}
	return labBuild
}

func parseFlags() (inDir, outDir string) {
	flag.StringVar(&inDir, "indir", "", "Path to directory from which to read input files")
	flag.StringVar(&outDir, "outdir", "", "Path to directory in which to write output files")
	flag.Parse()
	return
}

func NewLaboratoryBuilder(fh io.ReadCloser) *LaboratoryBuilder {
	labBuild := EmptyLaboratoryBuilder()
	inDir, outDir := parseFlags()
	labBuild.Setup(fh, inDir, outDir)
	return labBuild
}

func (labBuild *LaboratoryBuilder) Setup(fh io.ReadCloser, inDir, outDir string) {
	err := utils.ErrorFuncs{
		// We sort out the paths first so that we have somewhere to
		// write out errors to if the workflow is invalid.
		func() error { return labBuild.SetupPaths(inDir, outDir) },
		func() error { return labBuild.SetupWorkflow(fh) },
		func() error { return labBuild.SetupEffects() },
	}.Run()
	if err != nil {
		labBuild.RecordError(err, true)
	}
}

func (labBuild *LaboratoryBuilder) SetupWorkflow(fh io.ReadCloser) error {
	if wf, err := workflow.WorkflowFromReaders(fh); err != nil {
		return err
	} else if err := wf.Validate(); err != nil {
		return err
	} else if err := wf.NewSimulation(); err != nil {
		return err
	} else {
		labBuild.Logger = labBuild.Logger.With("simulationId", wf.Simulation.SimulationId)
		if anthaMod := composer.AnthaModule(); anthaMod != nil && len(anthaMod.Version) != 0 {
			wf.Simulation.Version = anthaMod.Version
			labBuild.Logger.Log("simulatorVersion", anthaMod.Version)
		} else {
			wf.Simulation.Version = "unknown"
			labBuild.Logger.Log("simulatorVersion", "unknown")
		}
		wf.Simulation.Start = time.Now().Format(time.RFC3339Nano)

		labBuild.Workflow = wf
		return nil
	}
}

func (labBuild *LaboratoryBuilder) SetupPaths(inDir, outDir string) error {
	labBuild.inDir, labBuild.outDir = inDir, outDir

	// Make sure we have a valid working outDir:
	if labBuild.outDir == "" {
		if d, err := ioutil.TempDir("", "antha-run-outputs"); err != nil {
			return err
		} else {
			labBuild.outDir = d
		}
	}
	labBuild.Logger.Log("outdir", labBuild.outDir)

	// Create subdirs within it:
	for _, leaf := range []string{"elements", "data", "tasks", "workflow"} {
		if err := utils.MkdirAll(filepath.Join(labBuild.outDir, leaf)); err != nil {
			return err
		}
	}

	// Switch the logger over to write to disk too:
	if logFH, err := utils.CreateFile(filepath.Join(labBuild.outDir, "logs.txt"), utils.ReadWrite); err != nil {
		return err
	} else {
		labBuild.logFH = logFH
		labBuild.Logger.SwapWriters(logFH, os.Stderr)
	}

	// Sort out inDir:
	if labBuild.inDir == "" {
		// We do this to make certain that we have a root path to join
		// onto so we can't permit reading arbitrary parts of the
		// filesystem.
		if d, err := ioutil.TempDir("", "antha-run-inputs"); err != nil {
			return err
		} else {
			labBuild.inDir = d
		}
	}
	labBuild.Logger.Log("indir", labBuild.inDir)
	return nil
}

func (labBuild *LaboratoryBuilder) SetupEffects() error {
	if fm, err := effects.NewFileManager(filepath.Join(labBuild.inDir, "data"), filepath.Join(labBuild.outDir, "data")); err != nil {
		return err
	} else {
		labBuild.effects = effects.NewLaboratoryEffects(labBuild.Workflow, labBuild.Workflow.Simulation.SimulationId, fm)
		return nil
	}
}

// Returns all the errors that were encountered and recorded in this lab's existence
func (labBuild *LaboratoryBuilder) Decommission() error {
	labBuild.Workflow.Simulation.End = time.Now().Format(time.RFC3339Nano)

	labBuild.elemLock.Lock()
	for _, lab := range labBuild.elemsCompleted {
		simElem := &workflow.SimulatedElement{
			ElementInstanceName: lab.element.Name(),
			ElementTypeName:     lab.element.TypeName(),
			StatePath:           lab.statePath,
		}
		if lab.parent != nil {
			simElem.ParentElementId = workflow.ElementId(fmt.Sprint(lab.parent.id))
		}
		if lab.err != nil {
			simElem.Error = lab.err.Error()
		}
		labBuild.Workflow.Simulation.Elements[workflow.ElementId(fmt.Sprint(lab.id))] = simElem
	}
	labBuild.elemLock.Unlock()

	if err := labBuild.saveWorkflow(); err != nil {
		labBuild.RecordError(err, true)
	}

	if err := labBuild.saveErrors(); err != nil {
		labBuild.RecordError(err, true)
	}

	if err := runner.Export(labBuild.effects.IDGenerator, labBuild.inDir, labBuild.outDir, labBuild.instrs, labBuild.Errors()); err != nil {
		labBuild.RecordError(err, true)
	}

	if labBuild.logFH != nil {
		labBuild.Logger.SwapWriters(os.Stderr)
		if err := labBuild.logFH.Sync(); err != nil {
			labBuild.Logger.Log("msg", "Error when syncing log file handle", "error", err)
		}
		if err := labBuild.logFH.Close(); err != nil {
			labBuild.Logger.Log("msg", "Error when closing log file handle", "error", err)
		}
		labBuild.logFH = nil
	}

	return labBuild.Errors()
}

func (labBuild *LaboratoryBuilder) saveWorkflow() error {
	return labBuild.Workflow.WriteToFile(filepath.Join(labBuild.outDir, "workflow", "workflow.json"), false)
}

// returns non-nil error iff there is an error during the *saving*
// process. I.e. this is not a reflection of whether there have been
// errors recorded.
func (labBuild *LaboratoryBuilder) saveErrors() error {
	if labBuild.Errors() != nil {
		// Because we've called Errors() we have gone through a memory
		// barrier, so direct access to labBuild.errors is now safe,
		// provided we are the only go-routine doing so, which we should
		// be.
		return labBuild.errors.WriteToFile(filepath.Join(labBuild.outDir, "errors.json"))
	} else {
		return nil
	}
}

func (labBuild *LaboratoryBuilder) RemoveOutDir() error {
	return os.RemoveAll(labBuild.outDir)
}

func (labBuild *LaboratoryBuilder) RemoveInDir() error {
	return os.RemoveAll(labBuild.inDir)
}

func (labBuild *LaboratoryBuilder) Compile() {
	if labBuild.Errors() != nil {
		return

	} else if devices, err := labBuild.connectDevices(); err != nil {
		labBuild.RecordError(err, true)

	} else {
		defer devices.Close()

		// We have to do this this late because we need the connections
		// to the plugins established to eg figure out if the device
		// supports prompting.
		human.New(labBuild.effects.IDGenerator).DetermineRole(devices)

		tasksDir := filepath.Join(labBuild.outDir, "tasks")

		if nodes, err := labBuild.effects.Maker.MakeNodes(labBuild.effects.Trace.Instructions()); err != nil {
			labBuild.RecordError(err, true)

		} else if instrs, err := codegen.Compile(labBuild.effects, tasksDir, devices, nodes); err != nil {
			labBuild.RecordError(err, true)

		} else {
			labBuild.instrs = instrs
		}
	}
}

func (labBuild *LaboratoryBuilder) connectDevices() (*target.Target, error) {
	cfg := &labBuild.Workflow.Config
	inv := labBuild.effects.Inventory
	tgt := target.New()
	if global, err := mixer.NewGlobalMixerConfig(inv, &cfg.GlobalMixer); err != nil {
		return nil, err
	} else {
		err := utils.ErrorSlice{
			mixer.NewGilsonPipetMaxInstances(labBuild.Logger, tgt, inv, global, cfg.GilsonPipetMax),
			mixer.NewTecanInstances(labBuild.Logger, tgt, inv, global, cfg.Tecan),
			mixer.NewCyBioInstances(labBuild.Logger, tgt, inv, global, cfg.CyBio),
			mixer.NewLabcyteInstances(labBuild.Logger, tgt, inv, global, cfg.Labcyte),
			mixer.NewHamiltonInstances(labBuild.Logger, tgt, inv, global, cfg.Hamilton),
			qpcrdevice.NewQPCRInstances(tgt, cfg.QPCR),
			shakerincubator.NewShakerIncubatorsInstances(tgt, cfg.ShakerIncubator),
			woplatereader.NewWOPlateReaderInstances(tgt, cfg.PlateReader),
		}.Pack()
		if err != nil {
			return nil, err
		} else if err := tgt.Connect(labBuild.Workflow); err != nil {
			tgt.Close()
			return nil, err
		} else {
			return tgt, nil
		}
	}
}

// This interface exists just to allow both the lab builder and
// laboratory itself to contain the InstallElement method. We need the
// method on both due to the dynamic inter-element calls.
type ElementInstaller interface {
	InstallElement(Element)
}

func (labBuild *LaboratoryBuilder) InstallElement(e Element) {
	labBuild.addElementLaboratory(e, labBuild.makeLab(labBuild.Logger, e, nil))
}

func (labBuild *LaboratoryBuilder) addElementLaboratory(e Element, lab *Laboratory) {
	labBuild.elemLock.Lock()
	defer labBuild.elemLock.Unlock()
	labBuild.elementToLab[e] = lab
}

func (labBuild *LaboratoryBuilder) AddConnection(src, dst Element, fun func()) error {
	labBuild.elemLock.Lock()
	defer labBuild.elemLock.Unlock()
	if labSrc, found := labBuild.elementToLab[src]; !found {
		return fmt.Errorf("Unknown src element: %v", src)
	} else if labDst, found := labBuild.elementToLab[dst]; !found {
		return fmt.Errorf("Unknown dst element: %v", dst)
	} else {
		labDst.addBlockedInput()
		labSrc.addOnExit(func(lab *Laboratory) {
			if lab.err == nil {
				fun()
				labDst.inputReady()
			}
		})
		return nil
	}
}

// Run all the installed elements.
func (labBuild *LaboratoryBuilder) RunElements() {
	if labBuild.Errors() != nil {
		return
	}

	labBuild.elemLock.Lock()
	labBuild.elemsRunning = len(labBuild.elementToLab)
	if labBuild.elemsRunning == 0 {
		labBuild.elemLock.Unlock()
		close(labBuild.Completed)

	} else {
		elemToLab := labBuild.elementToLab
		labBuild.elementToLab = make(map[Element]*Laboratory)
		for _, lab := range elemToLab {
			lab.addOnExit(labBuild.recordElementError)
			lab.addOnExit(labBuild.elementCompleted)
		}
		for _, lab := range elemToLab {
			go lab.run()
		}
		labBuild.elemLock.Unlock()
		<-labBuild.Completed
	}
}

// Our sole concern here is accounting of how many elements are still
// unrun.
func (labBuild *LaboratoryBuilder) elementCompleted(lab *Laboratory) {
	labBuild.elemLock.Lock()
	defer labBuild.elemLock.Unlock()
	labBuild.elemsCompleted = append(labBuild.elemsCompleted, lab)
	labBuild.elemsRunning--
	if labBuild.elemsRunning == 0 {
		close(labBuild.Completed)
	}
}

// for top-level (i.e. workflow-defined) elements, if the element
// produces an error then that error is fatal to the whole
// workflow. It will have already been logged though so we don't need
// to worry about that here.
func (labBuild *LaboratoryBuilder) recordElementError(lab *Laboratory) {
	if lab.err != nil {
		labBuild.RecordError(lab.err, false)
	}
}

// record that an error has happened, and optionally log it out of the
// standard logger. Safe for concurrent use.
func (labBuild *LaboratoryBuilder) RecordError(err error, log bool) {
	if log {
		labBuild.Logger.Log("error", err)
	}
	labBuild.errLock.Lock()
	defer labBuild.errLock.Unlock()
	labBuild.errors = append(labBuild.errors, err)
	select { // we keep the lock here to avoid a race to close
	case <-labBuild.errored:
	default:
		close(labBuild.errored)
	}
}

// Returns any errors that have been encountered and recorded so far -
// does not block. Safe for concurrent use.
func (labBuild *LaboratoryBuilder) Errors() error {
	select {
	case <-labBuild.errored:
		labBuild.errLock.Lock()
		defer labBuild.errLock.Unlock()
		return labBuild.errors.Pack()
	default:
		return nil
	}
}

// Laboratory. A separate laboratory exists for every element. It
// provides each element with the ability to interact with the lab
// effects, and to log.
type Laboratory struct {
	labBuild *LaboratoryBuilder
	// if non-nil then this laboratory has a parent, which means it was
	// not part of the workflow itself (i.e. CallSteps)
	parent  *Laboratory
	element Element

	Logger   *logger.Logger
	Workflow *workflow.Workflow
	*effects.LaboratoryEffects

	// every element has a unique id to ensure we don't collide on names.
	id uint64
	// count of inputs that are not yet ready (plus 1)
	pendingCount int64
	// this gets closed when all inputs become ready
	inputsReady chan struct{}
	// funcs to run when this element is completed
	onExit []func(*Laboratory)
	// closed once the the element has stopped. This closing does *not*
	// imply that inputsReady will be closed. For example, the element
	// can be blocked waiting for inputs to become ready when a
	// workflow error occurs, causing labBuild.errored to close. That
	// will be detected and will cause the element to close exited. But
	// inputsReady will still remain unclosed.
	exited chan struct{}
	// any error that was produced by the element itself
	err error

	// name of file (leaf) where we've written out the state of the
	// element once it's completed
	statePath string
}

func (labBuild *LaboratoryBuilder) makeLab(logger *logger.Logger, e Element, parent *Laboratory) *Laboratory {
	id := atomic.AddUint64(&labBuild.nextElementId, 1)
	return &Laboratory{
		labBuild: labBuild,
		parent:   parent,
		element:  e,

		Logger:            logger.With("id", id, "name", e.Name(), "type", e.TypeName()),
		Workflow:          labBuild.Workflow,
		LaboratoryEffects: labBuild.effects,

		id:           id,
		pendingCount: 1,
		inputsReady:  make(chan struct{}),
		exited:       make(chan struct{}),
	}
}

func (lab *Laboratory) InstallElement(e Element) {
	// take the root logger (from labBuild) and build up from there.
	logger := lab.labBuild.Logger.With("parentId", lab.id, "parentName", lab.element.Name(), "parentType", lab.element.TypeName())
	lab.labBuild.addElementLaboratory(e, lab.labBuild.makeLab(logger, e, lab))
}

// CallSteps is only for use when you're in an element and want to
// call another element (a dynamically-called element). Note that in
// this case, only the Steps are run. Any error that the Steps of the
// element produces is logged and returned. However, it is up to the
// calling element to determine whether or not such an error is fatal
// to the workflow (if it is considered fatal, the calling element
// should return it).
func (lab *Laboratory) CallSteps(e Element) error {
	// it should already be in the map because the element constructor
	// will have called through to InstallElement which would have
	// added it.
	labBuild := lab.labBuild
	labBuild.elemLock.Lock()
	eLab, found := labBuild.elementToLab[e]
	if !found {
		labBuild.elemLock.Unlock()
		panic(fmt.Errorf("CallSteps called on unknown element '%s'", e.Name()))
	}
	// delete it so it can't be run more than once
	delete(labBuild.elementToLab, e)

	labBuild.elemsRunning++
	eLab.addOnExit(lab.labBuild.elementCompleted)
	labBuild.elemLock.Unlock()

	go eLab.run(eLab.element.Steps)
	<-eLab.exited

	return eLab.err
}

// run() is designed to be called from a new go-routine (i.e. `go
// lab.run()`), hence it does not return an error. If you wish to wait
// for the element to stop and be safe to access, wait for the
// lab.exited channel to be closed.
func (lab *Laboratory) run(funs ...func(*Laboratory) error) {
	lab.Logger.Log("progress", "waiting")
	lab.inputReady()

	if len(funs) == 0 {
		funs = []func(*Laboratory) error{
			lab.element.Setup,
			lab.element.Steps,
			lab.element.Analysis,
			lab.element.Validation,
		}
	}

	defer lab.stopped()

	defer func() {
		if res := recover(); res != nil {
			// A panic is always fatal to the whole workflow, regardless of the element
			fmt.Printf("panic: %v\n%s", res, lab.labBuild.lineMapManager.ElementStackTrace())
			lab.err = fmt.Errorf("panic: %v", res)
			lab.labBuild.RecordError(lab.err, false)
		}
	}()

	select {
	case <-lab.inputsReady:
		if lab.parent == nil {
			token := <-lab.labBuild.throttle
			defer func() {
				res := recover()
				lab.labBuild.throttle <- token
				if res != nil {
					panic(res)
				}
			}()
		}
		lab.Logger.Log("progress", "starting")
		// this defer comes here because this defer will read all our
		// inputs and parameters as part of the save() call. This is
		// only safe (concurrency) once we know our inputs are ready.
		defer lab.save()
		for _, fun := range funs {
			select {
			case <-lab.labBuild.errored:
				return
			default:
				if err := fun(lab); err != nil {
					lab.err = err
					return
				}
			}
		}
	case <-lab.labBuild.errored:
		return
	}
}

func (lab *Laboratory) stopped() {
	err := lab.err
	if err == nil {
		lab.Logger.Log("progress", "stopped")
	} else {
		lab.Logger.Log("progress", "stopped", "error", err)
	}
	close(lab.exited)
	for _, fun := range lab.onExit {
		fun(lab)
	}
}

func (lab *Laboratory) save() {
	leaf := fmt.Sprintf("%d_%s.json", lab.id, lab.element.Name())
	p := filepath.Join(lab.labBuild.outDir, "elements", leaf)
	if fh, err := utils.CreateFile(p, utils.ReadWrite); err != nil {
		lab.labBuild.RecordError(err, true)
	} else {
		defer fh.Close()
		if err := json.NewEncoder(fh).Encode(lab.element); err != nil {
			lab.labBuild.RecordError(err, true)
		} else {
			lab.statePath = leaf
		}
	}
}

func (lab *Laboratory) inputReady() {
	if atomic.AddInt64(&lab.pendingCount, -1) == 0 {
		// we've done the transition from 1 -> 0. By definition, we're
		// the only routine that can do that (given that wiring is
		// finished before any element is started, so we cannot be
		// racing with calls to addBlockedInput), so we don't need to be
		// careful about double-close-panics.
		close(lab.inputsReady)
	}
}

func (lab *Laboratory) addBlockedInput() {
	atomic.AddInt64(&lab.pendingCount, 1)
}

// nb, it is not safe for the callback fun to read the inputs /
// parameters of this element unless there is no error.
func (lab *Laboratory) addOnExit(fun func(*Laboratory)) {
	lab.onExit = append(lab.onExit, fun)
}
