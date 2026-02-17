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

	"github.com/gin-gonic/gin"
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
		PROXY_URL string
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

func main() {

	// Allow only 3000 requests per minute, with burst 500
	openAIRateLimiter = rate.NewLimiter(rate.Every(time.Minute/3000), 500)

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
		print(connStr)
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
		log.WithError(err).Fatal("Failed to get executable path")
	}

	// Get current project directory
	projectDir := filepath.Dir(ex)

	// Print project directory
	fmt.Println(projectDir)

	sqlStatements := string(schemaBytes)

	// Execute SQL statements
	_, err = pgdb.Exec(sqlStatements)
	if err != nil {
		log.WithError(err).Fatal("Failed to execute SQL schema statements")
	}
	fmt.Println("SQL statements executed successfully")

	sqlc_q := sqlc_queries.New(pgdb)
	secretService := NewJWTSecretService(sqlc_q)
	jwtSecretAndAud, err = secretService.GetOrCreateJwtSecret(context.Background(), "chat")
	if err != nil {
		log.Fatal(err)
	}

	// Create services container
	services := &Services{
		SQLC:        sqlc_q,
		SecretSvc:   secretService,
		AuthUserSvc: NewAuthUserService(sqlc_q),
	}

	// Setup Gin router
	ginRouter := SetupRouter(services, jwtSecretAndAud)

	// Setup static file serving for Gin
	setupGinStaticFiles(ginRouter)

	// Print routes for debugging
	PrintRoutes(ginRouter)

	// Fly.io idle shutdown
	if os.Getenv("FLY_APP_NAME") != "" {
		setupFlyIdleShutdown()
	}

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", ginRouter)
	if err != nil {
		log.Fatal(err)
	}
}

// setupFlyIdleShutdown configures idle shutdown for Fly.io
func setupFlyIdleShutdown() {
	restartInterval := os.Getenv("FLY_RESTART_INTERVAL_IF_IDLE")

	// If not set, default to 30 minutes
	if restartInterval == "" {
		restartInterval = "30m"
	}

	duration, err := time.ParseDuration(restartInterval)
	if err != nil {
		log.Println("Invalid FLY_RESTART_INTERVAL_IF_IDLE value. Exiting.")
		return
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

// setupGinStaticFiles configures static file serving for Gin router
func setupGinStaticFiles(r *gin.Engine) {
	// Static files are served from embedded FS
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Check if it's a static asset request
		if strings.HasPrefix(path, "/static/") {
			c.Header("Cache-Control", "max-age=31536000") // 1 year
		} else if path == "/" || path == "" {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		// Serve from embedded static files
		fs := http.FileServer(http.FS(static.StaticFiles))
		fs.ServeHTTP(c.Writer, c.Request)
	})
}
