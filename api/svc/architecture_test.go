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

func TestWorkspaceServiceAPIDoesNotExposeSQLCRecords(t *testing.T) {
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, "chat_workspace_service.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Recv == nil {
			continue
		}
		star, ok := fn.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		receiver, ok := star.X.(*ast.Ident)
		if !ok || receiver.Name != "ChatWorkspaceService" {
			continue
		}
		ast.Inspect(fn.Type, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := selector.X.(*ast.Ident)
			if ok && pkg.Name == "sqlc_queries" {
				position := files.Position(selector.Pos())
				t.Errorf("%s:%d workspace method %s exposes generated SQLC type %s", position.Filename, position.Line, fn.Name.Name, selector.Sel.Name)
			}
			return true
		})
	}
}

func TestCoreSessionAPIDoesNotExposeSQLCRecords(t *testing.T) {
	coreMethods := map[string]bool{
		"CreateChatSession": true, "GetChatSessionByID": true,
		"UpdateChatSession": true, "GetAllChatSessions": true,
		"GetChatSessionsByUserID": true, "GetChatSessionByUUID": true,
		"UpdateChatSessionByUUID": true, "UpdateChatSessionTopicByUUID": true,
		"CreateOrUpdateChatSessionByUUID": true, "UpdateSessionMaxLength": true,
		"GetChatSessionByUUIDWithInActive": true,
	}
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, "chat_session_service.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || !coreMethods[fn.Name.Name] {
			continue
		}
		ast.Inspect(fn.Type, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := selector.X.(*ast.Ident)
			if ok && pkg.Name == "sqlc_queries" {
				position := files.Position(selector.Pos())
				t.Errorf("%s:%d core session method %s exposes generated SQLC type %s", position.Filename, position.Line, fn.Name.Name, selector.Sel.Name)
			}
			return true
		})
	}
}
