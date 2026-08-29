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

const dtoImportPath = "github.com/swuecho/chat_backend/dto"

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

func TestServicesDoNotImportTransportDTOs(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files := token.NewFileSet()
		file, err := parser.ParseFile(files, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imported := range file.Imports {
			if imported.Path.Value == `"`+dtoImportPath+`"` {
				position := files.Position(imported.Pos())
				t.Errorf("%s:%d imports transport DTOs; use an application type and map it in the handler", path, position.Line)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceInputsDoNotHaveJSONTags(t *testing.T) {
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
			gen, ok := declaration.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || (!strings.HasSuffix(typeSpec.Name.Name, "Input") && !strings.HasSuffix(typeSpec.Name.Name, "Command")) {
					continue
				}
				structure, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for _, field := range structure.Fields.List {
					if field.Tag != nil && strings.Contains(field.Tag.Value, "json:") {
						position := files.Position(field.Pos())
						t.Errorf("%s:%d application input %s has an HTTP JSON tag", path, position.Line, typeSpec.Name.Name)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
