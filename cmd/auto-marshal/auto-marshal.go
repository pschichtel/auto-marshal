package main

import (
	"fmt"
	"github.com/pschichtel/auto-marshal/internal/app/util"
	"github.com/pschichtel/auto-marshal/pkg/api"
	"github.com/pschichtel/auto-marshal/pkg/api/interfaces"
	"github.com/pschichtel/auto-marshal/pkg/api/structs"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s <symbol>\n", os.Args[0])
		os.Exit(1)
	}
	pwd := util.ResolvedPwd()
	moduleRoot := util.FindModuleRoot(pwd)
	packagePath := util.DetectPackagePath(pwd, moduleRoot)
	println("Generating code for package: " + packagePath)
	symbolName := os.Args[1]
	config := packages.Config{Mode: packages.NeedTypes | packages.NeedDeps | packages.NeedImports}
	packageList, err := packages.Load(&config, packagePath)
	if err != nil {
		panic(err)
	}
	if len(packageList) != 1 {
		panic("did not get exactly one resolved package!")
	}

	p := packageList[0]
	if len(p.Errors) > 0 {
		for i, e := range p.Errors {
			_, _ = fmt.Fprintf(os.Stderr, "%d. Error: %s\n", i+1, e.Error())
		}
		panic("Processing of package (" + packagePath + ")" + p.PkgPath + " " + p.Name + " failed")
	}

	scope := p.Types.Scope()
	obj := scope.Lookup(symbolName)
	if obj == nil {
		_, _ = fmt.Fprintf(os.Stderr, "Type %s not found!\n", symbolName)
		os.Exit(1)
	}
	sourceFile := p.Fset.File(obj.Pos()).Name()

	if _, ok := obj.(*types.TypeName); !ok {
		_, _ = fmt.Fprintf(os.Stderr, "Type '%s' not found in package '%s'!\n", symbolName, packagePath)
		os.Exit(1)
		return
	}
	println(symbolName, " -> ")

	underlying := obj.Type().Underlying()
	switch kind := underlying.(type) {
	case *types.Struct:
		println("a struct!", kind.NumFields())
		err = structs.GenerateCode(sourceFile, kind, &obj)
		if err != nil {
			panic(err)
		}
	case *types.Basic:
		println("a primitive!", kind.Name())
	case *types.Interface:
		implementations := api.FindImplementations(kind, p)
		println("an interface!", obj.Name(), len(implementations))
		for _, impl := range implementations {
			println("  - ", impl.Name())
		}

		err = interfaces.GenerateCode(sourceFile, &obj, implementations, "json")
		if err != nil {
			panic(err)
		}
	default:
		println("unknown kind", kind.String())
	}
}
