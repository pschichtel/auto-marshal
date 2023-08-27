package simple

import (
	. "github.com/dave/jennifer/jen"
	"github.com/pschichtel/auto-marshal/pkg/api"
	"github.com/pschichtel/auto-marshal/pkg/api/encoder"
	"go/types"
)

func GenerateCode(targetFile string, basicType *types.Type, object *types.Object, pointer bool) error {
	file, err := generateFile(basicType, object, pointer)
	if err != nil {
		return err
	}
	return file.Save(targetFile)
}

func generateFile(theType *types.Type, object *types.Object, pointer bool) (*File, error) {
	file := api.CreateJenFile(object)

	basicType, isBasic := (*theType).(*types.Basic)
	namedType, isNamed := (*theType).(*types.Named)
	sliceType, isSlice := (*theType).(*types.Slice)
	arrayType, isArray := (*theType).(*types.Array)
	mapType, isMap := (*theType).(*types.Map)

	// TODO pointer dereferencing is currently not nil-safe

	var err error = nil
	if isBasic {
		generateEncoderFunctionForBasic(file, basicType, object, pointer)
	} else if isNamed {
		generateEncoderFunctionForNamed(file, namedType, object, pointer)
	} else if isSlice {
		err = generateEncoderFunctionForSlice(file, sliceType, object, pointer)
	} else if isArray {
		err = generateEncoderFunctionForArray(file, arrayType, object, pointer)
	} else if isMap {
		err = generateEncoderFunctionForMap(file, mapType, object, pointer)
	} else {
		err = encoder.JsonEncodingError("Unsupported simple type: " + (*theType).String())
	}
	if err != nil {
		return nil, err
	}

	if !pointer {
		api.GenerateMarshalFunction(file, object)
	}

	return file, nil
}

// func DereferenValue()

func GenerateBasicType(basicType *types.Basic, pointer bool) Code {
	derefOp := "*"
	if pointer {
		derefOp = "**"
	}
	basicWriterFunctionName := api.WriterFunctionForBasicType(basicType)
	return Id(api.WriterVariableName).Dot(basicWriterFunctionName).Call(Id(basicType.Name()).Params(Op(derefOp).Id(api.ValueVariableName)))
}

func generateEncoderFunctionForBasic(file *File, basicType *types.Basic, object *types.Object, pointer bool) {
	typeName := (*object).Name()
	file.Func().Id(api.EncoderFunctionNameForNamedType(typeName)).Params(
		api.EncoderFunctionParams(typeName)...,
	).Params(Error()).Block(
		api.ReturnNilIfValueIsNil(),
		GenerateBasicType(basicType, pointer),
		Return(Nil()),
	).Line()
}

func GenerateNamedType(namedType *types.Named, pointer bool) Code {
	var value Code
	if pointer {
		value = Op("*").Id(api.ValueVariableName)
	} else {
		value = Id(api.ValueVariableName)
	}
	return Id(api.EncoderFunctionNameForNamedType(namedType.Obj().Name())).Call(value, Id(api.WriterVariableName))
}

func generateEncoderFunctionForNamed(file *File, namedType *types.Named, object *types.Object, pointer bool) {
	typeName := (*object).Name()
	file.Func().Id(api.EncoderFunctionNameForNamedType(typeName)).Params(
		api.EncoderFunctionParams(typeName)...,
	).Params(Error()).Block(
		Return(GenerateNamedType(namedType, pointer)),
	).Line()
}

func generateEncoderFunctionForSlice(file *File, namedType *types.Slice, object *types.Object, pointer bool) error {
	return encoder.JsonEncodingError("Meh!")
}

func generateEncoderFunctionForArray(file *File, namedType *types.Array, object *types.Object, pointer bool) error {
	return encoder.JsonEncodingError("Meh!")
}

func generateEncoderFunctionForMap(file *File, namedType *types.Map, object *types.Object, pointer bool) error {
	return encoder.JsonEncodingError("Meh!")
}
