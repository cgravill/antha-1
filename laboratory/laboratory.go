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

	// elemLock must be taken to access/mutate elemsUnrun and elements fields
	elemLock      sync.Mutex
	elemsUnrun    int64
	elementToLab  map[Element]*Laboratory
	nextElementId uint64

	// This lock is here to serialize access to errors (append and
	// Pack()), and to avoid races around closing Errored chan
	errLock sync.Mutex
	errors  utils.ErrorSlice
	Errored chan struct{}

	Completed chan struct{}

	*lineMapManager
	Logger *logger.Logger
	logFH  *os.File

	effects *effects.LaboratoryEffects

	instrs instructions.Insts
}

func EmptyLaboratoryBuilder() *LaboratoryBuilder {
	labBuild := &LaboratoryBuilder{
		elementToLab: make(map[Element]*Laboratory),
		Errored:      make(chan struct{}),
		Completed:    make(chan struct{}),

		lineMapManager: NewLineMapManager(),
		Logger:         logger.NewLogger(),
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
	} else if simId, err := workflow.RandomBasicId(wf.WorkflowId); err != nil {
		return err
	} else {
		wf.SimulationId = simId
		labBuild.Logger = labBuild.Logger.With("simulationId", simId)
		if anthaMod := composer.AnthaModule(); anthaMod != nil && len(anthaMod.Version) != 0 {
			if err := wf.Meta.Set("SimulatorVersion", anthaMod.Version); err != nil {
				return err
			}
			labBuild.Logger.Log("simulatorVersion", anthaMod.Version)
		} else {
			if err := wf.Meta.Set("SimulatorVersion", "unknown"); err != nil {
				return err
			}
			labBuild.Logger.Log("simulatorVersion", "unknown")
		}
		wf.Meta.Set("SimulationStart", time.Now().Format(time.RFC3339Nano))

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
		if err := os.MkdirAll(filepath.Join(labBuild.outDir, leaf), 0700); err != nil {
			return err
		}
	}

	// Switch the logger over to write to disk too:
	if logFH, err := os.OpenFile(filepath.Join(labBuild.outDir, "logs.txt"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0400); err != nil {
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
		labBuild.effects = effects.NewLaboratoryEffects(labBuild.Workflow, labBuild.Workflow.SimulationId, fm)
		return nil
	}
}

// Returns all the errors that were encountered and recorded in this lab's existence
func (labBuild *LaboratoryBuilder) Decommission() error {
	labBuild.Workflow.Meta.Set("SimulationEnd", time.Now().Format(time.RFC3339Nano))
	if err := labBuild.saveWorkflow(); err != nil {
		labBuild.RecordError(err, true)
	}

	if err := labBuild.saveErrors(); err != nil {
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

func (labBuild *LaboratoryBuilder) Export() {
	if err := export(labBuild.effects.IDGenerator, labBuild.inDir, labBuild.outDir, labBuild.instrs, labBuild.Errors()); err != nil {
		labBuild.RecordError(err, true)
	}
}

// This interface exists just to allow both the lab builder and
// laboratory itself to contain the InstallElement method. We need the
// method on both due to the dynamic inter-element calls.
type ElementInstaller interface {
	InstallElement(Element)
}

func (labBuild *LaboratoryBuilder) InstallElement(e Element) {
	labBuild.addElementLaboratory(e, labBuild.makeLab(labBuild.Logger, e))
}

func (labBuild *LaboratoryBuilder) addElementLaboratory(e Element, lab *Laboratory) {
	labBuild.elemLock.Lock()
	defer labBuild.elemLock.Unlock()
	labBuild.elementToLab[e] = lab
	labBuild.elemsUnrun++
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
		labSrc.addOnExit(func() {
			fun()
			labDst.inputReady()
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
	if labBuild.elemsUnrun == 0 {
		labBuild.elemLock.Unlock()
		close(labBuild.Completed)

	} else {
		for _, lab := range labBuild.elementToLab {
			lab.addOnExit(labBuild.elementCompleted)
		}
		for _, lab := range labBuild.elementToLab {
			go lab.run()
		}
		labBuild.elemLock.Unlock()
		<-labBuild.Completed
	}
}

func (labBuild *LaboratoryBuilder) elementCompleted() {
	labBuild.elemLock.Lock()
	defer labBuild.elemLock.Unlock()
	labBuild.elemsUnrun--
	if labBuild.elemsUnrun == 0 {
		close(labBuild.Completed)
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
	case <-labBuild.Errored:
	default:
		close(labBuild.Errored)
	}
}

// Returns any errors that have been encountered and recorded so far -
// does not block. Safe for concurrent use.
func (labBuild *LaboratoryBuilder) Errors() error {
	select {
	case <-labBuild.Errored:
		labBuild.errLock.Lock()
		defer labBuild.errLock.Unlock()
		return labBuild.errors.Pack()
	default:
		return nil
	}
}

// Laboratory. A separate laboratory exists for every element. It
// provides each element with the ability to interact with the lab,
// and to log.
type Laboratory struct {
	labBuild *LaboratoryBuilder
	element  Element

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
	onExit []func()
}

func (labBuild *LaboratoryBuilder) makeLab(logger *logger.Logger, e Element) *Laboratory {
	id := atomic.AddUint64(&labBuild.nextElementId, 1)
	return &Laboratory{
		labBuild: labBuild,
		element:  e,

		Logger:            logger.With("id", id, "name", e.Name(), "type", e.TypeName()),
		Workflow:          labBuild.Workflow,
		LaboratoryEffects: labBuild.effects,

		id:           id,
		pendingCount: 1,
		inputsReady:  make(chan struct{}),
	}
}

func (lab *Laboratory) InstallElement(e Element) {
	// take the root logger (from labBuild) and build up from there.
	logger := lab.labBuild.Logger.With("parentName", lab.element.Name(), "parentType", lab.element.TypeName())
	lab.labBuild.addElementLaboratory(e, lab.labBuild.makeLab(logger, e))
}

// CallSteps is only for use when you're in an element and want to
// call another element. Note that in this case, only the Steps are
// run.
func (lab *Laboratory) CallSteps(e Element) error {
	// it should already be in the map because the element constructor
	// will have called through to InstallElement which would have
	// added it.
	lab.labBuild.elemLock.Lock()
	eLab, found := lab.labBuild.elementToLab[e]
	if !found {
		lab.labBuild.elemLock.Unlock()
		return fmt.Errorf("CallSteps called on unknown element '%s'", e.Name())
	}

	finished := make(chan struct{})
	eLab.addOnExit(func() { close(finished) })
	eLab.addOnExit(lab.labBuild.elementCompleted)
	lab.labBuild.elemLock.Unlock()

	go eLab.run(eLab.element.Steps)
	<-finished

	return lab.labBuild.Errors()
}

func (lab *Laboratory) error(err error) {
	lab.labBuild.RecordError(err, false)
	lab.Logger.Log("error", err.Error())
}

func (lab *Laboratory) errorf(fmtStr string, args ...interface{}) {
	lab.error(fmt.Errorf(fmtStr, args...))
}

func (lab *Laboratory) run(funs ...func(*Laboratory) error) {
	lab.inputReady()

	if len(funs) == 0 {
		funs = []func(*Laboratory) error{
			lab.element.Setup,
			lab.element.Steps,
			lab.element.Analysis,
			lab.element.Validation,
		}
	}

	defer lab.exited()

	defer func() {
		if res := recover(); res != nil {
			lab.errorf("panic: %v", res)
			// Use println because of the embedded \n in the Stack Trace
			fmt.Println(lab.labBuild.lineMapManager.ElementStackTrace())
		}
	}()

	select {
	case <-lab.inputsReady:
		lab.Logger.Log("progress", "starting")
		// this defer comes here because this defer will read all our
		// inputs and parameters as part of the Save() call. This is
		// only safe (concurrency) once we know our inputs are ready.
		defer lab.save()
		for _, fun := range funs {
			select {
			case <-lab.labBuild.Errored:
				return
			default:
				if err := fun(lab); err != nil {
					lab.error(err)
					return
				}
			}
		}
	case <-lab.labBuild.Errored:
		return
	}
}

func (lab *Laboratory) exited() {
	funs := lab.onExit
	lab.onExit = nil
	for _, fun := range funs {
		fun()
	}
}

func (lab *Laboratory) save() {
	lab.Logger.Log("progress", "stopped")
	p := filepath.Join(lab.labBuild.outDir, "elements", fmt.Sprintf("%d_%s.json", lab.id, lab.element.Name()))
	if fh, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0400); err != nil {
		lab.error(err)
	} else {
		defer fh.Close()
		if err := json.NewEncoder(fh).Encode(lab.element); err != nil {
			lab.error(err)
		}
	}
}

func (lab *Laboratory) inputReady() {
	if atomic.AddInt64(&lab.pendingCount, -1) == 0 {
		// we've done the transition from 1 -> 0. By definition, we're
		// the only routine that can do that, so we don't need to be
		// careful.
		close(lab.inputsReady)
	}
}

func (lab *Laboratory) addBlockedInput() {
	atomic.AddInt64(&lab.pendingCount, 1)
}

func (lab *Laboratory) addOnExit(fun func()) {
	lab.onExit = append(lab.onExit, fun)
}
