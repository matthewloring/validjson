// Package validjson defines an Analyzer that checks that struct fields
// with json tags have json serializable types.
package validjson

import (
	"fmt"
	"go/ast"
	"go/importer"
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
	interfaces, err := loadInterfaces()
	if err != nil {
		return nil, err
	}
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
			if isNonSkipJSONTag(styp.Tag(i)) && !isJSONSerializable(f.Type(), interfaces) {
				pass.Reportf(f.Pos(), "struct field has json tag but non-serializable type %v", f.Type())
			}
		}
	})
	return nil, nil
}

type interfaces struct {
	TextMarshaler   *types.Interface
	TextUnmarshaler *types.Interface
}

func loadInterfaces() (*interfaces, error) {
	encodingPkg, err := importer.Default().Import("encoding")
	if err != nil {
		return nil, fmt.Errorf("unable to import 'encoding/json' for analysis: %w", err)
	}
	return &interfaces{
		TextMarshaler:   encodingPkg.Scope().Lookup("TextMarshaler").Type().Underlying().(*types.Interface),
		TextUnmarshaler: encodingPkg.Scope().Lookup("TextUnmarshaler").Type().Underlying().(*types.Interface),
	}, nil
}

// See: https://blog.golang.org/json-and-go
func isJSONSerializable(t types.Type, ifaces *interfaces) bool {
	switch fieldType := t.(type) {
	case *types.Basic:
		return fieldType.Kind() != types.Complex64 && fieldType.Kind() != types.Complex128
	case *types.Chan:
		return false
	case *types.Map:
		return isJSONSerializableAsMapKey(fieldType.Key(), ifaces)
	case *types.Named:
		return isJSONSerializable(fieldType.Underlying(), ifaces)
	case *types.Signature:
		return false
	}
	return true
}

func isJSONSerializableAsMapKey(t types.Type, ifaces *interfaces) bool {
	switch keyType := t.(type) {
	case *types.Basic:
		return keyType.Info()&types.IsInteger != 0 || keyType.Kind() == types.String
	case *types.Named:
		return isJSONSerializableAsMapKey(keyType.Underlying(), ifaces) ||
			(types.Implements(t, ifaces.TextMarshaler) && types.Implements(types.NewPointer(t), ifaces.TextUnmarshaler))
	}
	return false
}

func isNonSkipJSONTag(tag string) bool {
	val, found := (reflect.StructTag)(tag).Lookup("json")
	return found && val != "-"
}
