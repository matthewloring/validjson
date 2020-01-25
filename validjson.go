// Package validjson defines an Analyzer that checks that struct fields
// with json tags have json serializable types.
package validjson

import (
	"go/ast"
	"go/types"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:             "validjson",
	Doc:              "check that struct fields with json tags have json-compatible types.",
	Requires:         []*analysis.Analyzer{inspect.Analyzer},
	RunDespiteErrors: true,
	Run:              run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.StructType)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		styp, ok := pass.TypesInfo.Types[n.(*ast.StructType)].Type.(*types.Struct)
		// Type information may be incomplete.
		if !ok {
			return
		}
		for i := 0; i < styp.NumFields(); i++ {
			f := styp.Field(i)
			if isNonSkipJSONTag(styp.Tag(i)) && !isJSONSerializable(f.Type()) {
				pass.Reportf(f.Pos(), "struct field has json tag but non-serializable type %v", f.Type())
			}
		}
	})
	return nil, nil
}

// See: https://blog.golang.org/json-and-go
func isJSONSerializable(t types.Type) bool {
	switch fieldType := t.(type) {
	case *types.Basic:
		return fieldType.Kind() != types.Complex64 && fieldType.Kind() != types.Complex128
	case *types.Chan:
		return false
	case *types.Map:
		return isJSONSerializableAsMapKey(fieldType.Key())
	case *types.Named:
		return isJSONSerializable(fieldType.Underlying())
	case *types.Signature:
		return false
	}
	return true
}

func isJSONSerializableAsMapKey(t types.Type) bool {
	switch keyType := t.(type) {
	case *types.Basic:
		return keyType.Info()&types.IsInteger != 0 || keyType.Kind() == types.String
	case *types.Named:
		return isJSONSerializableAsMapKey(keyType.Underlying())
	}
	return false
}

func isNonSkipJSONTag(tag string) bool {
	val, found := (reflect.StructTag)(tag).Lookup("json")
	return found && val != "-"
}
