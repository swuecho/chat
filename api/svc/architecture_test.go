package svc

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServicesDoNotExposeQueryEscapeHatch(t *testing.T) {
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
		for _, declaration := range file.Decls {
			fn, ok := declaration.(*ast.FuncDecl)
			if ok && fn.Recv != nil && fn.Name.Name == "Q" {
				position := files.Position(fn.Pos())
				t.Errorf("%s:%d exposes Q(); add an explicit service operation instead", path, position.Line)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
