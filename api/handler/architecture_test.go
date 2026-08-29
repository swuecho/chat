package handler

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlersDoNotBuildUntypedJSONObjects(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files := token.NewFileSet()
		file, err := parser.ParseFile(files, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			mapType, ok := node.(*ast.MapType)
			if !ok {
				return true
			}
			untyped := false
			switch value := mapType.Value.(type) {
			case *ast.Ident:
				untyped = value.Name == "any"
			case *ast.InterfaceType:
				untyped = value.Methods != nil && len(value.Methods.List) == 0
			}
			if untyped {
				position := files.Position(mapType.Pos())
				t.Errorf("%s:%d uses an untyped JSON-shaped map; define a handler response DTO", path, position.Line)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
