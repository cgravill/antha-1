package laboratory

import "github.com/antha-lang/antha/workflow"

type ElementTypeMeta struct {
	Name                 workflow.ElementTypeName
	GoSrcPath            string
	AnthaSrcPath         string
	LineMap              map[int]int
	InputsFieldTypes     map[workflow.ElementParameterName]string
	OutputsFieldTypes    map[workflow.ElementParameterName]string
	ParametersFieldTypes map[workflow.ElementParameterName]string
	DataFieldTypes       map[workflow.ElementParameterName]string
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
