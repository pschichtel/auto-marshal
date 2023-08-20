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
	println("Generating code for package: " + p.String())

	symbolName := (*obj).Name()
	underlying := namedType.Type().Underlying()
	switch kind := underlying.(type) {
	case *types.Interface:
		implementations := api.FindImplementations(kind, p)
		println(symbolName + " is an interface with " + strconv.Itoa(len(implementations)) + " implementations")
		for _, impl := range implementations {
			println("  - ", impl.Name())
		}

		return interfaces.GenerateCode(sourceFile, obj, implementations, "json")
	case *types.Struct:
		println(symbolName+" is a struct!", kind.NumFields())
		return structs.GenerateCode(sourceFile, kind, obj)
	case *types.Pointer:
		// TODO the elem can itself be a pointer again, it should be recursively dereferenced
		// TODO pointers should probably be handled as part of the simple type default case, since they aren't less simple than e.g. slices
		target := kind.Elem()
		println(symbolName + " is a pointer!")
		return simple.GenerateCode(sourceFile, &target, obj, true)
	default:
		println(symbolName+" is a simple type!", kind.String())
		return simple.GenerateCode(sourceFile, &underlying, obj, false)
	}
}

func main() {
	if len(os.Args) < 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s <symbol>\n", os.Args[0])
		os.Exit(1)
	}
	pwd := util.ResolvedPwd()
	moduleRoot := util.FindModuleRoot(pwd)
	packagePath := util.DetectPackagePath(pwd, moduleRoot)
	symbolName := os.Args[1]
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
		for i, e := range p.Errors {
			_, _ = fmt.Fprintf(os.Stderr, "%d. Error: %s\n", i+1, e.Error())
		}
		_, _ = fmt.Fprintf(os.Stderr, "Processing of package (%s) %s %s failed\n", packagePath, p.PkgPath, p.Name)
		os.Exit(1)
	}

	scope := p.Types.Scope()
	obj := scope.Lookup(symbolName)
	if obj == nil {
		_, _ = fmt.Fprintf(os.Stderr, "Type %s not found!\n", symbolName)
		os.Exit(1)
	}
	sourceFile := p.Fset.File(obj.Pos()).Name()
	namedType, isNamedType := obj.(*types.TypeName)

	if !isNamedType {
		_, _ = fmt.Fprintf(os.Stderr, "Type '%s' not found in package '%s'!\n", symbolName, packagePath)
		os.Exit(1)
		return
	}

	err = generate(sourceFile, p, &obj, namedType)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Code generation failed for type %s: %s\n", symbolName, err.Error())
		os.Exit(1)
	}
}
