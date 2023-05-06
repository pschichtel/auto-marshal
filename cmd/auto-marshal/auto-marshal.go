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

type B string

func (i B) Test() {

}

type C string

func (i C) Test() {

}

func main() {
	if len(os.Args) < 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s <package name>\n", os.Args[0])
		os.Exit(1)
	}
	packageName := os.Args[1]
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
		panic(p.Errors[0].Error())
	}

	scope := p.Types.Scope()
	for _, symbolName := range scope.Names() {
		obj := scope.Lookup(symbolName)

		if _, ok := obj.(*types.TypeName); !ok {
			continue
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
		default:
			println("unknown kind", kind.String())
		}
	}
}
