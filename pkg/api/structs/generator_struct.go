package structs

import (
	. "github.com/dave/jennifer/jen"
	"github.com/fatih/structtag"
	"github.com/pschichtel/auto-marshal/pkg/api"
	"github.com/pschichtel/auto-marshal/pkg/api/encoder"
	"go/types"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

func GenerateCode(sourceFile string, structType *types.Struct, structObject *types.Object) error {
	hasInterface := false
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		fieldType := field.Type()
		fieldPointerType, fieldIsPointer := fieldType.(*types.Pointer)
		if fieldIsPointer {
			fieldType = fieldPointerType.Elem()
		}
		underlyingType := fieldType.Underlying()
		_, hasInterface = underlyingType.(*types.Interface)
		if hasInterface {
			break
		}
	}

	if !hasInterface {
		return nil
	}

	file, err := generateFile(structType, structObject)
	if err != nil {
		return err
	}
	return file.Save(api.DeriveOutputFileName(sourceFile, structObject))
}

func generateFile(structType *types.Struct, structObject *types.Object) (*File, error) {
	file := api.CreateJenFile(structObject)

	err := generateFieldsEncoderFunction(file, structType, structObject)
	if err != nil {
		return nil, err
	}
	generateEncoderFunction(file, structObject)
	api.GenerateMarshalFunction(file, structObject)

	return file, nil
}

func StructFieldEncoderFunctionNameForNamedType(typeName string) string {
	return api.EncoderFunctionNameForNamedType(typeName) + "Fields"
}

func generateFieldsEncoderFunction(file *File, structType *types.Struct, structObject *types.Object) error {

	var body []Code
	emittedFields := 0
	hasError := false
	var usedFieldNames []string

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		tag := structType.Tag(i)
		tags, err := structtag.Parse(tag)
		if err != nil {
			panic(err)
		}

		if !field.Exported() {
			continue
		}

		if emittedFields > 0 {
			body = append(body, Id(api.WriterVariableName).Dot("RawString").Call(Lit(",")))
		}
		emittedFields++
		fieldName := field.Name()
		jsonFieldName := firstToLower(fieldName)
		jsonFieldNameFromTags := findJsonFieldName(tags)
		if jsonFieldNameFromTags != nil {
			jsonFieldName = *jsonFieldNameFromTags
		}
		if slices.Contains(usedFieldNames, jsonFieldName) {
			return encoder.JsonEncodingError("Field name '" + jsonFieldName + "' is not unique in struct!")
		}
		usedFieldNames = append(usedFieldNames, jsonFieldName)

		body = append(body, Id(api.WriterVariableName).Dot("String").Call(Lit(jsonFieldName)))
		body = append(body, Id(api.WriterVariableName).Dot("RawString").Call(Lit(":")))

		fieldType := field.Type()
		fieldPointerType, fieldIsPointer := fieldType.(*types.Pointer)
		if fieldIsPointer {
			fieldType = fieldPointerType.Elem()
		}
		stmt := Id(fieldName)
		if fieldIsPointer {
			stmt.Op("*")
		}

		namedType, isNamedType := fieldType.(*types.Named)
		if isNamedType {
			var valueParam Code
			if fieldIsPointer {
				valueParam = Id(api.ValueVariableName).Dot(fieldName)
			} else {
				valueParam = Op("&").Id(api.ValueVariableName).Dot(fieldName)
			}
			errorAssignOp := "="
			if !hasError {
				errorAssignOp = ":="
				hasError = true
			}

			generatorFunc := generatorRefFromTags(tags)
			if generatorFunc == nil {
				generatorFunc = Id(api.EncoderFunctionNameForNamedType(namedType.Obj().Name()))
			}
			body = append(body, appendStatement(Err().Op(errorAssignOp), generatorFunc).Call(valueParam, Id(api.WriterVariableName)))
			body = append(body, If(Err().Op("!=").Nil()).Block(Return(Err())))
			continue
		}

		arrayType, isArray := fieldType.(*types.Array)
		if isArray {
			return encoder.JsonEncodingError("can't handle array types: " + arrayType.String())
		}

		sliceType, isSlice := fieldType.(*types.Slice)
		if isSlice {
			return encoder.JsonEncodingError("can't handle slice types: " + sliceType.String())
		}

		mapType, isMap := fieldType.(*types.Map)
		if isMap {
			keyType := mapType.Key()
			basicKeyType, keyIsBasic := keyType.(*types.Basic)
			if !keyIsBasic || basicKeyType.Kind() != types.String {
				return encoder.JsonEncodingError("maps can only have strings as keys, but got: " + keyType.String())
			}
			return encoder.JsonEncodingError("can't handle map types: " + mapType.String())
		}

		basicType, isBasic := fieldType.(*types.Basic)
		if isBasic {
			basicWriterFunctionName := api.WriterFunctionForBasicType(basicType)
			if fieldIsPointer {
				stmt := If(Id(api.ValueVariableName).Dot(fieldName).Op("!=").Nil()).Block(
					Id(api.WriterVariableName).Dot(basicWriterFunctionName).Call(Op("*").Id(api.ValueVariableName).Dot(fieldName)),
				).Else().Block(
					Id(api.WriterVariableName).Dot("RawString").Call(Lit("null")),
				)
				body = append(body, stmt)
			} else {
				body = append(body, Id(api.WriterVariableName).Dot(basicWriterFunctionName).Call(Id(api.ValueVariableName).Dot(fieldName)))
			}
			continue
		}

		return encoder.JsonEncodingError("can't handle this type: " + fieldType.String())
	}

	body = append(body, Return(Nil()))

	if emittedFields > 0 {
		body = append([]Code{
			If(Op("!").Id(api.FirstVariableName)).Block(Id(api.WriterVariableName).Dot("RawString").Call(Lit(","))),
		}, body...)
	}

	structName := (*structObject).Name()
	file.Func().Id(StructFieldEncoderFunctionNameForNamedType(structName)).Params(
		append(api.EncoderFunctionParams(structName), Id(api.FirstVariableName).Bool())...,
	).Params(Id("error")).Block(body...).Line()

	return nil
}

func appendStatement(statement *Statement, code Code) *Statement {
	*statement = append(*statement, code)
	return statement
}

func findJsonFieldName(tags *structtag.Tags) *string {
	if tags == nil {
		return nil
	}
	tag, _ := tags.Get("json")
	if tag == nil {
		return nil
	}

	return &tag.Name
}

func generatorRefFromTags(tags *structtag.Tags) Code {
	if tags == nil {
		return nil
	}
	tag, _ := tags.Get("generator")
	if tag == nil {
		return nil
	}
	name := tag.Name
	if !strings.Contains(name, ".") {
		return Id(name)
	}
	parts := strings.SplitN(name, ".", 2)

	return Qual(parts[0], parts[1])
}

func generateEncoderFunction(file *File, structObject *types.Object) {
	structName := (*structObject).Name()
	file.Func().Id(api.EncoderFunctionNameForNamedType(structName)).Params(
		api.EncoderFunctionParams(structName)...,
	).Params(Error()).Block(
		api.ReturnNilIfValueIsNil(),
		Id(api.WriterVariableName).Dot("RawString").Call(Lit("{")),
		Err().Op(":=").Id(StructFieldEncoderFunctionNameForNamedType(structName)).Call(Id(api.ValueVariableName), Id(api.WriterVariableName), True()),
		If(Err().Op("!=").Nil()).Block(
			Return(Err()),
		),
		Id(api.WriterVariableName).Dot("RawString").Call(Lit("}")),
		Return(Nil()),
	).Line()
}

// Source: https://stackoverflow.com/a/75989905/1827771
func firstToLower(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	lc := unicode.ToLower(r)
	if r == lc {
		return s
	}
	return string(lc) + s[size:]
}
