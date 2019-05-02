package v1_2

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/antha-lang/antha/laboratory/effects"
	"github.com/antha-lang/antha/logger"
	"github.com/antha-lang/antha/workflow"
	"github.com/antha-lang/antha/workflow/migrate"
	"github.com/antha-lang/antha/workflow/migrate/provider"
)

func getTestV1_2WorkflowProvider() (provider.WorkflowProvider, error) {
	fixture := filepath.Join("testdata", "sample_v1_2_workflow.json")
	bytes, err := ioutil.ReadFile(fixture)
	if err != nil {
		return nil, err
	}

	wf := &workflowv1_2{}
	err = json.Unmarshal(bytes, wf)
	if err != nil {
		return nil, err
	}

	tmpDir, err := ioutil.TempDir("", "tests")
	if err != nil {
		return nil, err
	}
	fm, err := effects.NewFileManager(tmpDir, tmpDir)
	if err != nil {
		return nil, err
	}

	repo := &workflow.Repository{
		Directory: "/tmp",
	}

	elementNames := []string{"AccuracyTest", "Aliquot_Liquid"}

	elementTypeMap := workflow.ElementTypeMap{}
	for _, name := range elementNames {
		etn := workflow.ElementTypeName(name)
		ep := workflow.ElementPath("Elements/Test/" + name)
		elementTypeMap[etn] = workflow.ElementType{
			RepositoryName: "repos.antha.com/antha-test/elements-test",
			ElementPath:    ep,
		}
	}
	repoMap := workflow.ElementTypesByRepository{}
	repoMap[repo] = elementTypeMap

	gilsonDeviceName := "testie"

	p := NewV1_2WorkflowProvider(wf, fm, repoMap, gilsonDeviceName)

	return p, nil
}

func TestGetMeta(t *testing.T) {
	p, err := getTestV1_2WorkflowProvider()
	if err != nil {
		t.Fatal(err)
	}

	m, err := p.GetMeta()
	if err != nil {
		t.Fatal(err)
	}

	expectedName := "My Test Workflow"
	if m.Name != expectedName {
		t.Errorf("Expected name '%v', got '%v'", expectedName, m.Name)
	}
}

func TestGetElements(t *testing.T) {
	p, err := getTestV1_2WorkflowProvider()
	if err != nil {
		t.Fatal(err)
	}

	els, err := p.GetElements()
	if err != nil {
		t.Fatal(err)
	}

	if len(els.Instances) != 2 {
		t.Fatalf("Expected %d element instance(s), got %d", 2, len(els.Instances))
	}

	expectedNames := []string{"AccuracyTest 1", "Aliquot Liquid 1"}
	foundNames := []string{}
	for name := range els.Instances {
		foundNames = append(foundNames, string(name))
	}
	sort.Strings(foundNames)
	if !reflect.DeepEqual(expectedNames, foundNames) {
		t.Errorf("Expected element names %v, got %v", expectedNames, foundNames)
	}

	if len(els.InstancesConnections) != 1 {
		t.Fatalf("Expected %d element instance connection(s), got %d", 1, len(els.InstancesConnections))
	}

	if len(els.Types) != 2 {
		t.Fatalf("Expected %d element type(s), got %d", 2, len(els.Types))
	}
}

func TestGetConfig(t *testing.T) {
	p, err := getTestV1_2WorkflowProvider()
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := p.GetConfig()
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.GlobalMixer.UseDriverTipTracking {
		t.Fatal("Expected Config.GlobalMixer.UseDriverTipTracking to be true")
	}

	if !cfg.GlobalMixer.IgnorePhysicalSimulation {
		t.Fatal("Expected Config.GlobalMixer.IgnorePhysicalSimulation to be true")
	}

	if len(cfg.GilsonPipetMax.Devices) != 1 {
		t.Fatalf("Expected to find %d Gilson device(s), found %d", 1, len(cfg.GilsonPipetMax.Devices))
	}
}

func TestGetTesting(t *testing.T) {
	p, err := getTestV1_2WorkflowProvider()
	if err != nil {
		t.Fatal(err)
	}

	tst, err := p.GetTesting()
	if err != nil {
		t.Fatal(err)
	}

	if len(tst.MixTaskChecks) != 0 {
		t.Fatal("Got non-empty Testing from GetTesting()")
	}
}

func TestGetWorkflowID(t *testing.T) {
	p, err := getTestV1_2WorkflowProvider()
	if err != nil {
		t.Fatal(err)
	}

	id, err := p.GetWorkflowID()
	if err != nil {
		t.Fatal(err)
	}

	if string(id) == "" {
		t.Error("Got empty workflow ID")
	}
}

// TODO: DELETE THIS. Simple snippet to help sanity-check where we're up to
func TestMigrator(t *testing.T) {
	p, err := getTestV1_2WorkflowProvider()
	if err != nil {
		t.Fatal(err)
	}

	l := logger.NewLogger()
	m := migrate.NewMigrator(l, p)
	w, err := m.Workflow()

	output := "testdata/sample_v1_2_migrated.json"
	err = os.Remove(output)
	if err != nil {
		t.Log(err)
	}

	err = w.WriteToFile(output, true)
	if err != nil {
		t.Fatal(err)
	}
}
