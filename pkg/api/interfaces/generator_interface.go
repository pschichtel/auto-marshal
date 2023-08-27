package interfaces

import (
	. "github.com/dave/jennifer/jen"
	"github.com/pschichtel/auto-marshal/pkg/api"
	"github.com/pschichtel/auto-marshal/pkg/api/structs"
	"go/types"
)

func GenerateCode(targetFile string, interfaceObject *types.Object, implementations []*types.TypeName, jsonTypeTagName string) error {
	file := generateFile(interfaceObject, implementations, jsonTypeTagName)
	return file.Save(targetFile)
}

func generateFile(interfaceObject *types.Object, implementations []*types.TypeName, jsonTypeTagName string) *File {
	file := api.CreateJenFile(interfaceObject)

	generateEncoderFunction(file, interfaceObject, jsonTypeTagName, implementations)

	return file
}

func mapSlice[I any, O any](slice []I, f func(I) O) []O {
	var output []O
	for _, value := range slice {
		output = append(output, f(value))
	}
	return output
}

func generateEncoderFunction(file *File, interfaceObject *types.Object, typeTag string, implementations []*types.TypeName) {
	actualName := "actual"
	cases := mapSlice(implementations, func(i *types.TypeName) Code {
		return Case(Id(i.Name())).Block(generateMarshalSwitchCase(i, actualName)...)
	})
	cases = append(cases, Default().Block(Return(Qual("github.com/pschichtel/auto-marshal/pkg/api/encoder", "JsonEncodingError").Call(Lit("Unknown type: ").Op("+").Qual("reflect", "TypeOf").Call(Id(actualName)).Dot("Name").Call()))))
	interfaceName := (*interfaceObject).Name()
	body := []Code{
		api.WriteNilAndReturnIfValueIsNil(),
	}
	if len(implementations) > 0 {
		body = append(body,
			Id(api.WriterVariableName).Dot("RawString").Call(Lit("{")),
			Id(api.WriterVariableName).Dot("String").Call(Lit(typeTag)),
			Id(api.WriterVariableName).Dot("RawString").Call(Lit(":")),
			Switch(Id(actualName).Op(":=").Parens(Op("*").Id(api.ValueVariableName)).Assert(Type())).Block(cases...),
			Id(api.WriterVariableName).Dot("RawString").Call(Lit("}")),
		)
	} else {
		body = append(body,
			Id(api.WriterVariableName).Dot("RawString").Call(Lit("{}")),
		)
	}
	body = append(body,
		Return(Nil()),
	)
	file.Func().Id(api.EncoderFunctionNameForNamedType(interfaceName)).Params(
		api.EncoderFunctionParams(interfaceName)...,
	).Params(Error()).Block(body...).Line()
}

func generateMarshalSwitchCase(implementation *types.TypeName, actualName string) []Code {
	implName := implementation.Name()
	return []Code{
		Id(api.WriterVariableName).Dot("String").Call(Lit(implName)),
		Err().Op(":=").Id(structs.StructFieldEncoderFunctionNameForNamedType(implName)).Call(Op("&").Id(actualName), Id(api.WriterVariableName), False()),
		If(Err().Op("!=").Nil()).Block(
			Return(Err()),
		),
	}
}
