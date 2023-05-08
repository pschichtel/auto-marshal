package structs

import (
	. "github.com/dave/jennifer/jen"
	"github.com/pschichtel/auto-marshal/pkg/api"
	"github.com/pschichtel/auto-marshal/pkg/api/interfaces"
	"go/types"
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
	structName := (*structObject).Name()

	_ = api.GenerateAuxErrorType(file, structObject)

	auxStructName := "aux" + structName

	var fields []Code

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		fieldType := field.Type()
		fieldPointerType, fieldIsPointer := fieldType.(*types.Pointer)
		if fieldIsPointer {
			fieldType = fieldPointerType.Elem()
		}
		underlyingType := fieldType.Underlying()
		stmt := Id(field.Name())
		if fieldIsPointer {
			stmt.Op("*")
		}

		_, isInterface := underlyingType.(*types.Interface)
		namedType, isNamed := fieldType.(*types.Named)
		if isInterface && isNamed {
			interfaceObj := namedType.Obj()
			interfacePackage := interfaceObj.Pkg().Path()
			interfaceContainerName := interfaces.ContainerTypeName(interfaceObj.Name())
			// if the interface container exists
			if typeExists(structObject, interfaceContainerName) {
				if (*structObject).Pkg().Path() == interfacePackage {
					stmt.Id(interfaceContainerName)
				} else {
					stmt.Qual(interfacePackage, interfaceContainerName)
				}
			} else {
				if (*structObject).Pkg().Path() == interfacePackage {
					stmt.Id(interfaceObj.Name())
				} else {
					stmt.Qual(interfacePackage, interfaceObj.Name())
				}
			}
			fields = append(fields, stmt)
			continue
		}

		basicType, isBasic := fieldType.(*types.Basic)
		if isBasic {
			fields = append(fields, stmt.Id(basicType.Name()))
			continue
		}
	}

	file.Type().Id(auxStructName).Struct(fields...).Line()

	receiverName := "subject"
	generateMarshalCode(file, structType, structObject, auxStructName, receiverName)

	return file
}

func generateMarshalCode(file *File, structType *types.Struct, structObject *types.Object, auxStructName string, receiverName string) {
	auxValueName := "auxValue"
	structName := (*structObject).Name()
	body := []Code{
		Id(auxValueName).Op(":=").Id(auxStructName).Values(),
	}
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		fieldType := field.Type()
		fieldPointerType, fieldIsPointer := fieldType.(*types.Pointer)
		if fieldIsPointer {
			fieldType = fieldPointerType.Elem()
		}
		underlyingType := fieldType.Underlying()
		stmt := Id(auxValueName).Dot(field.Name())

		namedType, isNamed := fieldType.(*types.Named)
		useContainer := false
		if types.IsInterface(underlyingType) && isNamed {
			interfaceObj := namedType.Obj()
			containerTypeName := interfaces.ContainerTypeName(interfaceObj.Name())
			Id(containerTypeName).Id(interfaces.ContainerTypeValueFieldName)
			// if the interface container exists
			if typeExists(structObject, containerTypeName) {
				useContainer = true
				if fieldIsPointer {
					body = append(body, Id(auxValueName).Dot(field.Name()).Op("=").Op("&").Id(containerTypeName).Values())
				}
			}
		}
		if useContainer {
			stmt.Dot(interfaces.ContainerTypeValueFieldName)
		}
		stmt.Op("=")
		if useContainer && !fieldIsPointer {
			stmt.Op("&")
		}
		stmt.Id(receiverName).Dot(field.Name())
		body = append(body, stmt)
	}
	body = append(body, Return(Qual("encoding/json", "Marshal").Call(Id(auxValueName))))
	file.Func().Params(Id(receiverName).Op("*").Id(structName)).Id("MarshalJSON").Params().Params(Op("[]").Byte(), Error()).Block(
		body...,
	).Line()
}

func typeExists(context *types.Object, typeName string) bool {
	return (*context).Pkg().Scope().Lookup(typeName) != nil
}
