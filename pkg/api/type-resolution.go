package api

import (
	"go/types"
	"golang.org/x/tools/go/packages"
)

func FindImplementations(target *types.Interface, p *packages.Package) []*types.TypeName {
	var out []*types.TypeName
	scope := p.Types.Scope()
	for _, symbolName := range scope.Names() {
		obj := scope.Lookup(symbolName)

		// only named types can implement interfaces
		typeName, isTypeName := obj.(*types.TypeName)
		if !isTypeName {
			continue
		}

		_, isInterface := obj.Type().Underlying().(*types.Interface)
		if isInterface {
			continue
		}

		if types.Implements(obj.Type(), target) {
			out = append(out, typeName)
		}
	}

	return out
}
