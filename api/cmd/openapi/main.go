// Command openapi generates the build-time OpenAPI artifact without starting
// the application or connecting to PostgreSQL.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apiopenapi"
	"github.com/swuecho/chat_backend/handler"
)

func main() {
	output := flag.String("output", "openapi/openapi.json", "generated OpenAPI JSON path")
	flag.Parse()

	registry := apiopenapi.NewRegistry()
	router := mux.NewRouter()
	// Registration records contracts only. Services are not invoked while the
	// router is assembled, so generation has no database dependency.
	handler.NewChatSessionHandler(nil).Register(router, registry)
	handler.NewChatMessageHandler(nil, nil, nil, nil).Register(router, registry)
	handler.NewChatPromptHandler(nil).Register(router, registry)
	handler.NewUserActiveChatSessionHandler(nil, nil).Register(router, registry)
	handler.NewChatWorkspaceHandler(nil).Register(router, registry)
	handler.NewChatCommentHandler(nil).Register(router, registry)
	handler.NewChatHandler(nil, nil, nil, nil, nil, nil, nil, nil, "", "").Register(router, registry)
	handler.NewBotAnswerHistoryHandler(nil).Register(router, registry)
	handler.NewChatFileHandler(nil).Register(router, registry)
	handler.NewChatModelHandler(nil).Register(router, registry)
	handler.NewChatSnapshotHandler(nil).Register(router, registry)
	authHandler := handler.NewAuthUserHandler(nil, "", "", 0)
	authHandler.Register(router, registry)
	authHandler.RegisterPublicRoutes(router, registry)
	handler.NewUserChatModelPrivilegeHandler(nil).Register(router, registry)
	handler.NewAPIKeyHandler(nil).Register(router, registry)
	handler.NewAdminHandler(nil, nil, 0).RegisterRoutes(router, registry)
	handler.RegisterTTSContract(registry)

	document, err := registry.MarshalJSON()
	if err != nil {
		fatalf("marshal OpenAPI document: %v", err)
	}
	document = append(document, '\n')
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		fatalf("create output directory: %v", err)
	}
	if err := os.WriteFile(*output, document, 0o644); err != nil {
		fatalf("write OpenAPI document: %v", err)
	}
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
