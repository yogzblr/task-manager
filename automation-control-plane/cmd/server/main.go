package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/automation-platform/control-plane/internal/api"
	"github.com/automation-platform/control-plane/internal/auth"
	"github.com/automation-platform/control-plane/internal/centrifugo"
	"github.com/automation-platform/control-plane/internal/store/mysql"
	"github.com/automation-platform/control-plane/internal/store/redis"
)

func main() {
	ctx := context.Background()

	// Load configuration from environment
	mysqlDSN := getEnv("MYSQL_DSN", "automation:password@tcp(localhost:3306)/automation")
	redisAddr := getEnv("VALKEY_ADDR", "localhost:6379")
	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")
	centrifugoURL := getEnv("CENTRIFUGO_URL", "http://localhost:8000")
	centrifugoAPIKey := getEnv("CENTRIFUGO_API_KEY", "change-me-in-production")
	port := getEnv("PORT", "8080")

	// Initialize MySQL store
	mysqlStore, err := mysql.NewStore(ctx, mysql.Config{
		DSN:             mysqlDSN,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
	})
	if err != nil {
		log.Fatalf("Failed to initialize MySQL store: %v", err)
	}
	defer mysqlStore.Close()

	// Initialize Redis store
	redisStore, err := redis.NewStore(ctx, redis.Config{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Redis store: %v", err)
	}
	defer redisStore.Close()

	// Initialize Centrifugo client
	centrifugoClient := centrifugo.NewClient(centrifugo.Config{
		URL:    centrifugoURL,
		APIKey: centrifugoAPIKey,
	})
	log.Printf("Initialized Centrifugo client with URL: %s", centrifugoURL)

	// Initialize auth
	jwtValidator := auth.NewJWTValidator(jwtSecret)
	
	// Initialize RBAC authorizer with project roles getter
	projectRolesGetter := func(ctx context.Context, tenantID, userID string) (map[string][]auth.Role, error) {
		// TODO: Implement actual project roles lookup from database
		// For now, return empty map
		return make(map[string][]auth.Role), nil
	}
	rbacAuthorizer := auth.NewRBACAuthorizer(projectRolesGetter)

	// Initialize API handlers
	jobsHandler := api.NewJobsHandler(mysqlStore, rbacAuthorizer, centrifugoClient)
	projectsHandler := api.NewProjectsHandler(mysqlStore, rbacAuthorizer)
	agentsHandler := api.NewAgentsHandler(mysqlStore, rbacAuthorizer)
	auditHandler := api.NewAuditHandler(mysqlStore, rbacAuthorizer)

	// Setup routes
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API routes (protected by auth middleware)
	// Note: Go 1.21 compatible routing (pattern matching added in Go 1.22)
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Router] %s %s", r.Method, r.URL.Path)
		switch {
		case r.Method == "POST" && r.URL.Path == "/jobs":
			jobsHandler.CreateJob(w, r)
		case r.Method == "GET" && r.URL.Path == "/jobs":
			jobsHandler.ListJobs(w, r)
		case r.Method == "POST" && len(r.URL.Path) > 13 && r.URL.Path[:6] == "/jobs/" && r.URL.Path[len(r.URL.Path)-6:] == "/lease":
			log.Printf("[Router] Matched lease endpoint")
			jobsHandler.LeaseJob(w, r)
		case r.Method == "POST" && len(r.URL.Path) > 16 && r.URL.Path[:6] == "/jobs/" && r.URL.Path[len(r.URL.Path)-9:] == "/complete":
			log.Printf("[Router] Matched complete endpoint")
			jobsHandler.CompleteJob(w, r)
		default:
			log.Printf("[Router] No match found for %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	})
	apiMux.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Router] %s %s (exact match)", r.Method, r.URL.Path)
		switch {
		case r.Method == "POST":
			jobsHandler.CreateJob(w, r)
		case r.Method == "GET":
			jobsHandler.ListJobs(w, r)
		default:
			http.NotFound(w, r)
		}
	})
	apiMux.HandleFunc("/projects", projectsHandler.ListProjects)
	apiMux.HandleFunc("/agents/register", agentsHandler.RegisterAgent)
	apiMux.HandleFunc("/agents/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && len(r.URL.Path) > 15 && r.URL.Path[:15] == "/agents/" && r.URL.Path[len(r.URL.Path)-8:] == "/upgrade" {
			agentsHandler.UpgradeAgent(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
	apiMux.HandleFunc("/audit/logs", auditHandler.ListAuditLogs)

	// Apply auth middleware
	handler := auth.AuthMiddleware(jwtValidator)(apiMux)
	
	// Logging middleware
	loggedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Main] Incoming request: %s %s", r.Method, r.URL.Path)
		handler.ServeHTTP(w, r)
	})
	
	mux.Handle("/api/", http.StripPrefix("/api", loggedHandler))

	// Start server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Server starting on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
