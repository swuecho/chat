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

func TestConvertedServiceAPIsDoNotExposeSQLCTypes(t *testing.T) {
	convertedReceivers := map[string]bool{
		"ChatSessionService":           true,
		"ChatWorkspaceService":         true,
		"ChatSnapshotService":          true,
		"ChatPromptService":            true,
		"ChatMessageService":           true,
		"SessionConversationService":   true,
		"SessionRateLimitService":      true,
		"SessionSnapshotQueryService":  true,
		"SessionAdminQueryService":     true,
		"SessionBotHistoryService":     true,
		"SessionModelService":          true,
		"UserActiveChatSessionService": true,
	}
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
			if !ok || fn.Recv == nil {
				continue
			}
			receiverType := fn.Recv.List[0].Type
			if star, ok := receiverType.(*ast.StarExpr); ok {
				receiverType = star.X
			}
			receiver, ok := receiverType.(*ast.Ident)
			if !ok || !convertedReceivers[receiver.Name] {
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
					t.Errorf("%s:%d %s.%s exposes generated SQLC type %s", path, position.Line, receiver.Name, fn.Name.Name, selector.Sel.Name)
				}
				return true
			})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestApplicationServicesDoNotManageSQLCTransactionsDirectly(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || path == "transaction_manager.go" {
			return nil
		}
		files := token.NewFileSet()
		file, err := parser.ParseFile(files, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if ok && selector.Sel.Name == "InTransaction" {
				position := files.Position(selector.Pos())
				t.Errorf("%s:%d manages a SQLC transaction directly; depend on TransactionManager", path, position.Line)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestChatSessionServiceRemainsLifecycleFocused(t *testing.T) {
	allowed := map[string]bool{
		"CreateChatSession": true, "GetChatSessionByID": true, "UpdateChatSession": true,
		"DeleteChatSession": true, "GetAllChatSessions": true, "GetChatSessionsByUserID": true,
		"GetSimpleChatSessionsByUserID": true, "GetChatSessionByUUID": true,
		"UpdateChatSessionByUUID": true, "UpdateChatSessionTopicByUUID": true,
		"CreateOrUpdateChatSessionByUUID": true, "DeleteChatSessionByUUID": true,
		"UpdateSessionMaxLength": true, "GetChatSessionByUUIDWithInActive": true,
		"CreateSessionFromSnapshot": true,
	}
	files := token.NewFileSet()
	packages, err := parser.ParseDir(files, ".", func(info os.FileInfo) bool { return !strings.HasSuffix(info.Name(), "_test.go") }, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range packages["svc"].Files {
		for _, declaration := range file.Decls {
			fn, ok := declaration.(*ast.FuncDecl)
			if !ok || fn.Recv == nil {
				continue
			}
			typeExpr := fn.Recv.List[0].Type
			if star, ok := typeExpr.(*ast.StarExpr); ok {
				typeExpr = star.X
			}
			receiver, ok := typeExpr.(*ast.Ident)
			if ok && receiver.Name == "ChatSessionService" && !allowed[fn.Name.Name] {
				t.Errorf("ChatSessionService.%s is not a session lifecycle operation", fn.Name.Name)
			}
		}
	}
}

func TestApplicationModelsUseStandardInitialisms(t *testing.T) {
	targets := map[string]bool{
		"ChatSession": true, "CreateOrUpdateChatSessionInput": true,
		"ChatPrompt": true, "CreateChatPromptInput": true, "UpdateChatPromptInput": true,
		"ChatMessage": true, "CreateChatMessageInput": true, "UpdateChatMessageByUUIDInput": true,
		"CreateBotAnswerHistoryInput": true,
	}
	for _, path := range []string{"chat_session_service.go", "chat_prompt_service.go", "chat_message_service.go", "bot_answer_history_service.go"} {
		files := token.NewFileSet()
		file, err := parser.ParseFile(files, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			gen, ok := declaration.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || !targets[typeSpec.Name.Name] {
					continue
				}
				structure, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for _, field := range structure.Fields.List {
					for _, name := range field.Names {
						if strings.Contains(name.Name, "Uuid") || strings.Contains(name.Name, "Llm") {
							t.Errorf("%s.%s uses nonstandard initialism; use UUID or LLM", typeSpec.Name.Name, name.Name)
						}
					}
				}
			}
		}
	}
}
