package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/static"
	"golang.org/x/time/rate"
)

var logger *log.Logger

type AppConfig struct {
	OPENAI struct {
		API_KEY   string
		RATELIMIT int
	}
	CLAUDE struct {
		API_KEY string
	}
	PG struct {
		HOST string
		PORT int
		USER string
		PASS string
		DB   string
	}
}

var appConfig AppConfig
var jwtSecretAndAud sqlc_queries.JwtSecret

func getFlattenKeys(prefix string, v reflect.Value) (keys []string) {
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			name := v.Type().Field(i).Name
			keys = append(keys, getFlattenKeys(prefix+name+".", field)...)
		}
	default:
		keys = append(keys, prefix[:len(prefix)-1])
	}
	return keys
}

func bindEnvironmentVariables() {
	appConfig = AppConfig{}
	for _, key := range getFlattenKeys("", reflect.ValueOf(appConfig)) {
		envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		err := viper.BindEnv(key, envKey)
		if err != nil {
			logger.Fatal("config: unable to bind env: " + err.Error())
		}
	}
}

//go:embed sqlc/schema.sql
var schemaBytes []byte

// lastRequest tracks the last time a request was received
var lastRequest time.Time
var openAIRateLimiter *rate.Limiter

var claudeRateLimiteToken chan struct{}

func main() {

	// Allow only 3000 requests per minute, with burst 500
	openAIRateLimiter = rate.NewLimiter(rate.Every(time.Minute/3000), 500)

	// A buffered channel with capacity 1
	// This ensures only one API call can proceed at a time
	claudeRateLimiteToken = make(chan struct{}, 1)

	lastRequest = time.Now()
	// Configure viper to read environment variables
	bindEnvironmentVariables()
	viper.AutomaticEnv()

	if err := viper.Unmarshal(&appConfig); err != nil {
		logger.Fatal("config: unable to decode into struct: " + err.Error())
	}

	log.Printf("%+v", appConfig)
	logger = log.New()
	logger.Formatter = &log.JSONFormatter{}

	// Establish a database connection
	dbURL := os.Getenv("DATABASE_URL")
	var connStr string
	if dbURL == "" {
		pg := appConfig.PG
		connStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			pg.HOST, pg.PORT, pg.USER, pg.PASS, pg.DB)
	} else {
		connStr = dbURL
	}
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

	sqlStatements := string(schemaBytes)

	// Execute SQL statements
	_, err = pgdb.Exec(sqlStatements)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("SQL statements executed successfully")

	// create a new Gorilla Mux router instance
	// Create a new router
	router := mux.NewRouter()
	sqlc_q := sqlc_queries.New(pgdb)
	secretService := NewJWTSecretService(sqlc_q)
	jwtSecretAndAud, err = secretService.GetOrCreateJwtSecret(context.Background(), "chat")
	if err != nil {
		log.Fatal(err)
	}

	ChatModelHandler := NewChatModelHandler(sqlc_q)
	ChatModelHandler.Register(router)

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

	chatSnapshotHandler := NewChatSnapshotHandler(chatMessageService)
	chatSnapshotHandler.Register(router)

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

	user_model_privilege_handler := NewUserChatModelPrivilegeHandler(sqlc_q)
	user_model_privilege_handler.Register(router)

	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		tpl, err1 := route.GetPathTemplate()
		met, err2 := route.GetMethods()
		fmt.Println(tpl, err1, met, err2)
		return nil
	})

	// Embed static/* directory
	fs := http.FileServer(http.FS(static.StaticFiles))

	// Set cache headers for static/assets files
	cacheHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "assets/") {
			w.Header().Set("Cache-Control", "max-age=31536000") // 1 year
		} else if r.URL.Path == "/index.html" {
			// Set no cache headers for index.html
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}
		fs.ServeHTTP(w, r)
	})

	// Redirect "/" to "/static/"
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/", http.StatusMovedPermanently)
	})

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", makeGzipHandler(cacheHandler)))

	// fly.io
	if os.Getenv("FLY_APP_NAME") != "" {
		router.Use(UpdateLastRequestTime)
	}
	router.Use(IsAuthorizedMiddleware)
	limitedRouter := RateLimitByUserID(sqlc_q)
	router.Use(limitedRouter)
	// Wrap the router with the logging middleware
	// 10 min < 100 requests
	// loggedMux := loggingMiddleware(router, logger)
	loggedRouter := handlers.LoggingHandler(logger.Out, router)

	// fly.io

	if os.Getenv("FLY_APP_NAME") != "" {
		// read env var FLY_RESTART_INTERVAL_IF_IDLE if not set, set to 30 minutes
		restartInterval := os.Getenv("FLY_RESTART_INTERVAL_IF_IDLE")

		// If not set, default to 30 minutes
		if restartInterval == "" {
			restartInterval = "30m"
		}

		duration, err := time.ParseDuration(restartInterval)
		if err != nil {
			log.Println("Invalid FLY_RESTART_INTERVAL_IF_IDLE value. Exiting.")
		}
		// Use a goroutine to check for inactivity and exit
		go func() {
			for {
				time.Sleep(1 * time.Minute) // Check every minute
				if time.Since(lastRequest) > duration {
					fmt.Printf("No activity for %s. Exiting.", restartInterval)
					os.Exit(0)
					return
				}
			}
		}()

	}

	err = http.ListenAndServe(":8080", loggedRouter)
	if err != nil {
		log.Fatal(err)
	}
}
