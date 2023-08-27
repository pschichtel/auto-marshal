package main

import (
	"fmt"
	"github.com/pschichtel/auto-marshal/internal/app/util"
	"github.com/pschichtel/auto-marshal/pkg/api"
	"github.com/pschichtel/auto-marshal/pkg/api/interfaces"
	"github.com/pschichtel/auto-marshal/pkg/api/simple"
	"github.com/pschichtel/auto-marshal/pkg/api/structs"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
	"strconv"
)

func generate(sourceFile string, p *packages.Package, obj *types.Object, namedType *types.TypeName) error {
	symbolName := (*obj).Name()
	targetFile := api.DeriveOutputFileName(sourceFile, obj)
	println("Generating code for: " + p.String() + "." + symbolName)
	println("    From: " + sourceFile)
	println("    To:   " + targetFile)

	underlying := namedType.Type().Underlying()
	switch kind := underlying.(type) {
	case *types.Interface:
		implementations := api.FindImplementations(kind, p)
		println(symbolName + " is an interface with " + strconv.Itoa(len(implementations)) + " implementations")
		for _, impl := range implementations {
			println("  - ", impl.Name())
		}

		return interfaces.GenerateCode(targetFile, obj, implementations, "json")
	case *types.Struct:
		println(symbolName+" is a struct!", kind.NumFields())
		return structs.GenerateCode(targetFile, kind, obj)
	case *types.Pointer:
		// TODO the elem can itself be a pointer again, it should be recursively dereferenced
		// TODO pointers should probably be handled as part of the simple type default case, since they aren't less simple than e.g. slices
		target := kind.Elem()
		println(symbolName + " is a pointer!")
		return simple.GenerateCode(targetFile, &target, obj, true)
	default:
		println(symbolName+" is a simple type!", kind.String())
		return simple.GenerateCode(targetFile, &underlying, obj, false)
	}
}

func main() {
	pwd := util.ResolvedPwd()
	moduleRoot := util.FindModuleRoot(pwd)
	packagePath := util.DetectPackagePath(pwd, moduleRoot)
	config := packages.Config{Mode: packages.NeedTypes | packages.NeedDeps | packages.NeedImports}
	packageList, err := packages.Load(&config, packagePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to load package %s: %s\n", packagePath, err.Error())
		os.Exit(1)
	}
	if len(packageList) != 1 {
		_, _ = fmt.Fprintf(os.Stderr, "Received more than one package when loading package %s!\n", packagePath)
		os.Exit(1)
	}

	p := packageList[0]
	if len(p.Errors) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Package processing errors:\n")
		for i, e := range p.Errors {
			_, _ = fmt.Fprintf(os.Stderr, "  %d. Error: %s\n", i+1, e.Error())
		}
	}

	scope := p.Types.Scope()

	var objects []types.Object

	if len(os.Args) > 1 {
		for i := 1; i < len(os.Args); i++ {
			obj := scope.Lookup(os.Args[i])
			if obj != nil {
				objects = append(objects, obj)
			}
		}
	} else {
		names := scope.Names()
		for i := range names {
			name := names[i]
			obj := scope.Lookup(name)
			if obj != nil {
				objects = append(objects, obj)
			}
		}
	}
	if len(objects) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "No symbols found!\n")
		os.Exit(1)
	}

	for i := range objects {
		obj := objects[i]
		_, isFunction := obj.(*types.Func)
		if isFunction {
			continue
		}

		namedType, isNamedType := obj.(*types.TypeName)
		if !isNamedType {
			_, _ = fmt.Fprintf(os.Stderr, "    Type '%s' not found in package '%s'!\n", obj.Name(), packagePath)
			continue
		}

		sourceFile := p.Fset.File(obj.Pos()).Name()
		err = generate(sourceFile, p, &obj, namedType)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "    Code generation failed for type %s: %s\n", obj.Name(), err.Error())
			continue
		}
	}
}
