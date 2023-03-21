package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/swuecho/chatgpt_backend/sqlc_queries"
)

var OPENAI_API_KEY string
var JWT_SECRET string
var JWT_AUD string
var logger *log.Logger

func main() {
	var exists bool
	if OPENAI_API_KEY, exists = os.LookupEnv("OPENAI_API_KEY"); !exists {
		log.Fatal("OPENAI_API_KEY not set")
	}
	OPENAI_API_KEY = os.Getenv("OPENAI_API_KEY")

	if JWT_SECRET, exists = os.LookupEnv("JWT_SECRET"); !exists {
		log.Fatal("JWT_SECRET not set")
	}
	JWT_SECRET = os.Getenv("JWT_SECRET")

	if JWT_AUD, exists = os.LookupEnv("JWT_AUD"); !exists {
		log.Fatal("JWT_AUD not set")
	}
	JWT_AUD = os.Getenv("JWT_AUD")

	// Create a new logger instance, configure it as desired
	logger = log.New()
	logger.Formatter = &log.JSONFormatter{}

	host := os.Getenv("PG_HOST")
	port := os.Getenv("PG_PORT")
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASS")
	dbname := os.Getenv("PG_DB")

	// Establish a database connection
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	pgdb, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer pgdb.Close()

	// Get current executable file path
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	// Get current project directory
	projectDir := filepath.Dir(ex)

	// Print project directory
	fmt.Println(projectDir)

	// Read SQL file
	schemaPath := filepath.Join(projectDir, "schema.sql")
	// check if file exists
	// Check if file exists
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		fmt.Println("File does not exist")
	} else {
		sqlFile, err := os.Open(schemaPath)
		if err != nil {
			panic(err.Error())
		}
		defer sqlFile.Close()

		// Get SQL statements
		sqlBytes, err := io.ReadAll(sqlFile)
		if err != nil {
			panic(err.Error())
		}
		sqlStatements := string(sqlBytes)

		// Execute SQL statements
		_, err = pgdb.Exec(sqlStatements)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("SQL statements executed successfully")
	}

	// create a new Gorilla Mux router instance
	// Create a new router
	router := mux.NewRouter()
	sqlc_q := sqlc_queries.New(pgdb)

	// create a new AuthUserService instance
	userService := NewAuthUserService(sqlc_q)

	// create a new AuthUserHandler instance
	userHandler := NewAuthUserHandler(userService)

	// register the AuthUserHandler with the router
	userHandler.Register(router)

	// create a new ChatPromptService instance
	promptService := NewChatPromptService(sqlc_q)

	// create a new ChatPromptHandler instance
	promptHandler := NewChatPromptHandler(promptService)

	// register the ChatPromptHandler with the router
	promptHandler.Register(router)

	// create a new ChatSessionService instance
	chatSessionService := NewChatSessionService(sqlc_q)

	// create a new ChatSessionHandler instance
	chatSessionHandler := NewChatSessionHandler(chatSessionService)

	// register the ChatSessionHandler with the router
	chatSessionHandler.Register(router)

	// create a new ChatMessageService instance
	chatMessageService := NewChatMessageService(sqlc_q)

	// create a new ChatMessageHandler instance
	chatMessageHandler := NewChatMessageHandler(chatMessageService)

	// register the ChatMessageHandler with the router
	chatMessageHandler.Register(router)

	// create a new UserActiveChatSessionService instance
	activeSessionService := NewUserActiveChatSessionService(sqlc_q)

	// UserActiveChatSessionHandler
	activeSessionHandler := NewUserActiveChatSessionHandler(activeSessionService)

	// register the UserActiveChatSessionHandler with the router
	activeSessionHandler.Register(router)

	// create a new ChatService instance
	chatService := NewChatService(sqlc_q)

	// create a new ChatHandler instance
	chatHandler := NewChatHandler(chatService)

	// regiser the ChatHandler with the router
	chatHandler.Register(router)

	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		tpl, err1 := route.GetPathTemplate()
		met, err2 := route.GetMethods()
		fmt.Println(tpl, err1, met, err2)
		return nil
	})
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.Use(IsAuthorizedMiddleware)
	limitedRouter := RateLimitByUserID(sqlc_q)
	router.Use(limitedRouter)
	// Wrap the router with the logging middleware
	// 10 min < 100 requests
	// loggedMux := loggingMiddleware(router, logger)
	loggedRouter := handlers.LoggingHandler(logger.Out, router)
	err = http.ListenAndServe(":8077", loggedRouter)
	if err != nil {
		log.Fatal(err)
	}
}
