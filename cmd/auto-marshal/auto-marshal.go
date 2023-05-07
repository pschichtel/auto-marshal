package main

import (
	"fmt"
	"github.com/pschichtel/auto-marshal/pkg/api"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
)

type A interface {
	Test()
}

type AContainer struct {
	a *A
}

func (a AContainer) ContainedValue() *A {
	return a.a
}

func (a AContainer) Test2() {

}

type B struct {
	S string
}

func (i B) Test() {

}

type C struct {
	I int32
}

func (i C) Test() {

}

func main() {
	if len(os.Args) < 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s <symbol>\n", os.Args[0])
		os.Exit(1)
	}
	packageName, envExists := os.LookupEnv("GOPACKAGE")
	if !envExists {
		_, _ = fmt.Fprintln(os.Stderr, "GOPACKAGE environment var was not set!")
		os.Exit(1)
	}
	symbolName := os.Args[1]
	config := packages.Config{Mode: packages.NeedTypes | packages.NeedDeps | packages.NeedImports}
	packageList, err := packages.Load(&config, packageName)
	if err != nil {
		panic(err)
	}
	if len(packageList) != 1 {
		panic("did not get exactly one resolved package!")
	}

	p := packageList[0]
	if len(p.Errors) > 0 {
		for i, e := range p.Errors {
			fmt.Fprintf(os.Stderr, "%d. Error: %s\n", i+1, e.Error())
		}
		panic("Processing of package (" + packageName + ")" + p.PkgPath + " " + p.Name + " failed")
	}

	scope := p.Types.Scope()
	obj := scope.Lookup(symbolName)
	sourceFile := p.Fset.File(obj.Pos())

	if _, ok := obj.(*types.TypeName); !ok {
		_, _ = fmt.Fprintf(os.Stderr, "Type '%s' not found in package '%s'!\n", symbolName, packageName)
		os.Exit(1)
		return
	}
	println(symbolName, " -> ")

	underlying := obj.Type().Underlying()
	switch kind := underlying.(type) {
	case *types.Struct:
		println("a struct!", kind.NumFields())
	case *types.Basic:
		println("a primitive!", kind.Name())
	case *types.Interface:
		implementations := api.FindImplementations(kind, p)
		println("an interface!", obj.Name(), len(implementations))
		for _, impl := range implementations {
			println("  - ", impl.Name())
		}

		err = api.GenerateInterfaceCode(sourceFile.Name(), kind, obj, implementations, "json")
		if err != nil {
			panic(err)
		}
	default:
		println("unknown kind", kind.String())
	}
}
