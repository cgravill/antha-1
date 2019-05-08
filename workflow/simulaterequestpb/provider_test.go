package simulaterequestpb

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/laboratory/effects"
	"github.com/antha-lang/antha/logger"
	"github.com/antha-lang/antha/workflow"
)

func getTestProvider(filename string, elementTypeNames ...workflow.ElementTypeName) (*Provider, error) {
	protobufFilePath := filepath.Join("testdata", filename)
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

	elementTypeMap := workflow.ElementTypeMap{}
	for _, etn := range elementTypeNames {
		ep := workflow.ElementPath("Elements/Test/" + etn)
		elementTypeMap[etn] = workflow.ElementType{
			RepositoryName: "repos.antha.com/antha-test/elements-test",
			ElementPath:    ep,
		}
	}
	repoMap := workflow.ElementTypesByRepository{}
	repoMap[repo] = elementTypeMap

	gilsonDeviceName := "testie"

	logger := logger.NewLogger()

	r, err := os.Open(protobufFilePath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return NewProvider(r, fm, repoMap, gilsonDeviceName, logger)
}

func TestGetConfig(t *testing.T) {
	p, err := getTestProvider("request.pb",
		"Define_Liquid_Set", "Define_Plate_Layout", "Setup_QPCR_Plate", "Upload_Plate_Layout_File_Single", "Upload_QPCR_Design_File")
	if err != nil {
		t.Fatal(err)
	}

	c, err := p.GetConfig()
	if err != nil {
		t.Fatal(err)
	}

	if !c.GlobalMixer.UseDriverTipTracking {
		t.Error("Expected UseDriverTipTracking to be true, got false")
	}

	if c.GlobalMixer.CustomPolicyRuleSet == nil {
		t.Error("Expected a custom policy rule set, got nil")
	}
}

func TestGetElements(t *testing.T) {
	p, err := getTestProvider("request.pb",
		"Define_Liquid_Set", "Define_Plate_Layout", "Setup_QPCR_Plate", "Upload_Plate_Layout_File_Single", "Upload_QPCR_Design_File")
	if err != nil {
		t.Fatal(err)
	}

	els, err := p.GetElements()
	if err != nil {
		t.Fatal(err)
	}

	if len(els.Types) == 0 {
		t.Error("Didn't find any element types")
	}
	if len(els.Instances) == 0 {
		t.Error("Didn't find any element instances")
	}
	if len(els.InstancesConnections) == 0 {
		t.Error("Didn't find any element connections")
	}
}

func TestEmptyFileParam(t *testing.T) {
	p, err := getTestProvider("requestEmptyFile.pb", "Upload_QPCR_Design_File")
	if err != nil {
		t.Fatal(err)
	}

	els, err := p.GetElements()
	if err != nil {
		t.Fatal(err)
	}

	if len(els.Types) != 1 {
		t.Fatalf("Expected exactly 1 element type, got %d", len(els.Types))
	} else if len(els.Instances) != 1 {
		t.Fatalf("Expected exactly 1 element instance, got %d", len(els.Instances))
	} else if elem, found := els.Instances["Upload QPCR Design File 1"]; !found {
		t.Fatal("Couldn't find element instance named 'Upload QPCR Design File 1'")
	} else if param, found := elem.Parameters["QPCRDesignFile"]; !found {
		t.Fatal("Couldn't find parameter named 'QPCRDesignFile' for element instance 'Upload QPCR Design File 1'")
	} else {
		file := &wtype.File{}
		if err := json.Unmarshal(param, file); err != nil {
			t.Fatal(err)
		} else if file.Name != "foo" {
			t.Fatalf("Expected file parameter name 'foo', but got %s'", file.Name)
		} else if bs, err := p.fm.ReadAll(file); err != nil {
			t.Fatal(err)
		} else if bs == nil || len(bs) != 0 {
			t.Fatalf("Expected to find an empty non-nil byte array from reading file. But got '%#v'", bs)
		}
	}
}
