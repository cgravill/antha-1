package laboratory

import "github.com/antha-lang/antha/workflow"

type ElementTypeMeta struct {
	Name                 workflow.ElementTypeName
	GoSrcPath            string
	AnthaSrcPath         string
	LineMap              map[int]int
	InputsFieldTypes     map[string]string
	OutputsFieldTypes    map[string]string
	ParametersFieldTypes map[string]string
	DataFieldTypes       map[string]string
}

func (labBuild *LaboratoryBuilder) RegisterElementType(elemType *ElementTypeMeta, repoName workflow.RepositoryName, elemPath workflow.ElementPath) {
	labBuild.RegisterLineMap(elemType)
	labBuild.Workflow.Simulation.Elements.Types[elemType.Name] = workflow.SimulatedElementType{
		ElementType: workflow.ElementType{
			RepositoryName: repoName,
			ElementPath:    elemPath,
		},
		GoSrcPath:            elemType.GoSrcPath,
		AnthaSrcPath:         elemType.AnthaSrcPath,
		InputsFieldTypes:     elemType.InputsFieldTypes,
		OutputsFieldTypes:    elemType.OutputsFieldTypes,
		ParametersFieldTypes: elemType.ParametersFieldTypes,
		DataFieldTypes:       elemType.DataFieldTypes,
	}
}
