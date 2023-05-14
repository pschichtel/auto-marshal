package basics

import (
	. "github.com/dave/jennifer/jen"
	"github.com/pschichtel/auto-marshal/pkg/api"
	"go/types"
)

func GenerateCode(sourceFile string, basicType *types.Type, object *types.Object, pointer bool) error {
	file := generateFile(basicType, object, pointer)
	return file.Save(api.DeriveOutputFileName(sourceFile, object))
}

func generateFile(theType *types.Type, object *types.Object, pointer bool) *File {
	file := api.CreateJenFile(object)

	basicType, isBasic := (*theType).(*types.Basic)
	namedType, isNamed := (*theType).(*types.Named)

	// TODO pointer dereferencing is currently not nil-safe

	if isBasic {
		generateEncoderFunctionForBasic(file, basicType, object, pointer)
	} else if isNamed {
		generateEncoderFunctionForNamed(file, namedType, object, pointer)
	}
	if !pointer {
		api.GenerateMarshalFunction(file, object)
	}

	return file
}

func generateEncoderFunctionForBasic(file *File, basicType *types.Basic, object *types.Object, pointer bool) {
	typeName := (*object).Name()
	basicWriterFunctionName := api.WriterFunctionForBasicType(basicType)
	derefOp := "*"
	if pointer {
		derefOp = "**"
	}
	file.Func().Id(api.EncoderFunctionNameForNamedType(typeName)).Params(
		api.EncoderFunctionParams(typeName)...,
	).Params(Error()).Block(
		api.ReturnNilIfValueIsNil(),
		Id(api.WriterVariableName).Dot(basicWriterFunctionName).Call(Id(basicType.Name()).Params(Op(derefOp).Id(api.ValueVariableName))),
		Return(Nil()),
	).Line()
}

func generateEncoderFunctionForNamed(file *File, namedType *types.Named, object *types.Object, pointer bool) {
	typeName := (*object).Name()
	var value Code
	if pointer {
		value = Op("*").Id(api.ValueVariableName)
	} else {
		value = Id(api.ValueVariableName)
	}
	file.Func().Id(api.EncoderFunctionNameForNamedType(typeName)).Params(
		api.EncoderFunctionParams(typeName)...,
	).Params(Error()).Block(
		Return(Id(api.EncoderFunctionNameForNamedType(namedType.Obj().Name())).Call(value, Id(api.WriterVariableName))),
	).Line()
}
