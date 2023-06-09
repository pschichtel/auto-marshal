package interfaces

import (
	. "github.com/dave/jennifer/jen"
	"github.com/pschichtel/auto-marshal/pkg/api"
	"go/types"
)

const ContainerTypeValueFieldName = "Value"

func GenerateCode(sourceFile string, interfaceObject *types.Object, implementations []*types.TypeName, jsonTypeTagName string) error {
	file := generateFile(interfaceObject, implementations, jsonTypeTagName)
	return file.Save(api.DeriveOutputFileName(sourceFile, interfaceObject))
}

func ContainerTypeName(interfaceName string) string {
	return interfaceName + "JsonContainer"
}

func generateFile(interfaceObject *types.Object, implementations []*types.TypeName, jsonTypeTagName string) *File {
	interfaceName := (*interfaceObject).Name()
	typeTagFieldName := "JsonTypeTag" // TODO derive a unique name from the implementations
	containerTypeName := ContainerTypeName(interfaceName)
	typeTagStructName := "aux" + interfaceName + "JsonTypeTag"

	file := api.CreateJenFile(interfaceObject)
	file.PackageComment("// Code generated by github.com/pschichtel/auto-marshal! DO NOT EDIT.")
	errorTypeName := api.GenerateAuxErrorType(file, interfaceObject)

	file.Type().Id(containerTypeName).Struct(
		Id(ContainerTypeValueFieldName).Op("*").Id(interfaceName),
	).Line()

	file.Type().Id(typeTagStructName).Struct(
		Id(typeTagFieldName).String().Tag(map[string]string{"json": jsonTypeTagName}),
	).Line()

	for _, implementation := range implementations {
		structRef := Op("*")
		structRef.Id(implementation.Name())
		file.Type().Id(marshalAuxStructName(interfaceObject, implementation)).Struct(
			Id(typeTagFieldName).String().Tag(map[string]string{"json": jsonTypeTagName}),
			structRef,
		).Line()
	}

	receiverName := "container"
	generateMarshalCode(file, interfaceObject, receiverName, containerTypeName, typeTagFieldName, errorTypeName, implementations)
	generateUnmarshalCode(file, interfaceObject, receiverName, containerTypeName, typeTagStructName, typeTagFieldName, errorTypeName, implementations)

	return file
}

func mapSlice[I any, O any](slice []I, f func(I) O) []O {
	var output []O
	for _, value := range slice {
		output = append(output, f(value))
	}
	return output
}

func generateMarshalCode(file *File, interfaceObject *types.Object, receiverName string, containerTypeName string, typeTagFieldName string, errorTypeName string, implementations []*types.TypeName) {
	valueName := "value"
	actualName := "actual"
	cases := mapSlice(implementations, func(i *types.TypeName) Code {
		return Case(Id(i.Name())).Block(generateMarshalSwitchCase(interfaceObject, i, typeTagFieldName, actualName)...)
	})
	cases = append(cases, Default().Block(Return(Nil(), Id(errorTypeName).Call(Lit("Unknown type: ").Op("+").Qual("reflect", "TypeOf").Call(Id(valueName)).Dot("Name").Call()))))
	file.Func().Params(Id(receiverName).Op("*").Id(containerTypeName)).Id("MarshalJSON").Params().Params(Op("[]").Byte(), Error()).Block(
		Id(valueName).Op(":=").Op("*").Id(receiverName).Dot("Value"),
		Switch(Id(actualName).Op(":=").Id(valueName).Assert(Type())).Block(cases...),
	).Line()
}

func generateMarshalSwitchCase(interfaceObject *types.Object, implementation *types.TypeName, typeTagFieldName string, actualName string) []Code {
	auxVarName := "aux"
	return []Code{
		Id(auxVarName).Op(":=").Id(marshalAuxStructName(interfaceObject, implementation)).Values(
			Id(typeTagFieldName).Op(":").Lit(implementation.Name()),
			Id(implementation.Name()).Op(":").Op("&").Id(actualName),
		),
		Return(Qual("encoding/json", "Marshal").Call(Id(auxVarName))),
	}
}

func marshalAuxStructName(interfaceObject *types.Object, implementation *types.TypeName) string {
	return "aux" + (*interfaceObject).Name() + "TaggedJson" + implementation.Name()
}

func generateUnmarshalCode(file *File, interfaceObject *types.Object, receiverName string, containerTypeName string, typeTagStructName string, typeTagFieldName string, errorTypeName string, implementations []*types.TypeName) {
	typeTagStructInstanceName := "typeTag"
	errName := "err"
	dataName := "data"
	cases := mapSlice(implementations, func(i *types.TypeName) Code {
		return Case(Lit(i.Name())).Block(generateUnmarshalSwitchCase(interfaceObject, i, receiverName, dataName, errName)...)
	})
	cases = append(cases, Default().Block(Return(Id(errorTypeName).Call(Lit("Unknown type: ").Op("+").Id(typeTagStructInstanceName).Dot(typeTagFieldName)))))

	file.Func().Params(Id(receiverName).Op("*").Id(containerTypeName)).Id("UnmarshalJSON").Params(Id(dataName).Op("[]").Byte()).Params(Error()).Block(
		Id(typeTagStructInstanceName).Op(":=").Id(typeTagStructName).Values(),
		Id(errName).Op(":=").Qual("encoding/json", "Unmarshal").Call(Id(dataName), Op("&").Id(typeTagStructInstanceName)),
		If(Id(errName).Op("!=").Nil()).Block(
			Return(Id(errName)),
		),
		Switch(Id(typeTagStructInstanceName).Dot(typeTagFieldName)).Block(cases...),
	).Line()
}

func generateUnmarshalSwitchCase(interfaceObject *types.Object, implementation *types.TypeName, receiverName string, dataName string, errName string) []Code {
	valueName := "value"
	copyName := "valueCopy"
	return []Code{
		Id(valueName).Op(":=").Id(implementation.Name()).Values(),
		Id(errName).Op("=").Qual("encoding/json", "Unmarshal").Call(Id(dataName), Op("&").Id(valueName)),
		If(Id(errName).Op("!=").Nil()).Block(
			Return(Id(errName)),
		),
		Var().Id(copyName).Id((*interfaceObject).Name()).Op("=").Id(valueName),
		Id(receiverName).Dot(ContainerTypeValueFieldName).Op("=").Op("&").Id(copyName),
		Return(Nil()),
	}
}
