package structs

import (
	. "github.com/dave/jennifer/jen"
	"github.com/fatih/structtag"
	"github.com/pschichtel/auto-marshal/pkg/api"
	"go/types"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

	file := generateFile(structType, structObject)
	return file.Save(api.DeriveOutputFileName(sourceFile, structObject))
}

func generateFile(structType *types.Struct, structObject *types.Object) *File {
	file := api.CreateJenFile(structObject)

	generateFieldsEncoderFunction(file, structType, structObject)
	generateEncoderFunction(file, structObject)
	generateMarshalFunction(file, structObject)

	return file
}

func StructFieldEncoderFunctionNameForNamedType(typeName string) string {
	return api.EncoderFunctionNameForNamedType(typeName) + "Fields"
}

func generateFieldsEncoderFunction(file *File, structType *types.Struct, structObject *types.Object) {

	var body []Code
	emittedFields := 0
	hasError := false

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

			body = append(body, Err().Op(errorAssignOp).Id(api.EncoderFunctionNameForNamedType(namedType.Obj().Name())).Call(valueParam, Id(api.WriterVariableName)))
			body = append(body, If(Err().Op("!=").Nil()).Block(Return(Err())))
			continue
		}

		basicType, isBasic := fieldType.(*types.Basic)
		if isBasic {
			basicWriterFunctionName := cases.Title(language.English, cases.Compact).String(basicType.Name())
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

		println("can't handle this type!")
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

func generateEncoderFunction(file *File, structObject *types.Object) {
	structName := (*structObject).Name()
	file.Func().Id(api.EncoderFunctionNameForNamedType(structName)).Params(
		api.EncoderFunctionParams(structName)...,
	).Params(Error()).Block(
		If(Id(api.ValueVariableName).Op("==").Nil()).Block(
			Id(api.WriterVariableName).Dot("RawString").Call(Lit("null")),
			Return(Nil()),
		),
		Id(api.WriterVariableName).Dot("RawString").Call(Lit("{")),
		Err().Op(":=").Id(StructFieldEncoderFunctionNameForNamedType(structName)).Call(Id(api.ValueVariableName), Id(api.WriterVariableName), True()),
		If(Err().Op("!=").Nil()).Block(
			Return(Err()),
		),
		Id(api.WriterVariableName).Dot("RawString").Call(Lit("}")),
		Return(Nil()),
	).Line()
}

func generateMarshalFunction(file *File, structObject *types.Object) {
	receiverName := "subject"
	structName := (*structObject).Name()
	file.Func().Params(Id(receiverName).Op("*").Id(structName)).Id("MarshalJSON").Params().Params(Op("[]").Byte(), Error()).Block(
		Return(Qual("github.com/pschichtel/auto-marshal/pkg/api/encoder", "EncodeJson").Call(Id(receiverName), Id(api.EncoderFunctionNameForNamedType(structName)))),
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
