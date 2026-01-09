package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/automation-platform/control-plane/internal/api"
	"github.com/automation-platform/control-plane/internal/auth"
	"github.com/automation-platform/control-plane/internal/store/mysql"
	"github.com/automation-platform/control-plane/internal/store/redis"
)

func main() {
	ctx := context.Background()

	// Load configuration from environment
	mysqlDSN := getEnv("MYSQL_DSN", "automation:password@tcp(localhost:3306)/automation")
	redisAddr := getEnv("VALKEY_ADDR", "localhost:6379")
	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")
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
	jobsHandler := api.NewJobsHandler(mysqlStore, rbacAuthorizer)
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
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("POST /jobs", jobsHandler.CreateJob)
	apiMux.HandleFunc("GET /jobs", jobsHandler.ListJobs)
	apiMux.HandleFunc("POST /jobs/{id}/lease", jobsHandler.LeaseJob)
	apiMux.HandleFunc("POST /jobs/{id}/complete", jobsHandler.CompleteJob)
	apiMux.HandleFunc("GET /projects", projectsHandler.ListProjects)
	apiMux.HandleFunc("POST /agents/register", agentsHandler.RegisterAgent)
	apiMux.HandleFunc("POST /agents/{id}/upgrade", agentsHandler.UpgradeAgent)
	apiMux.HandleFunc("GET /audit/logs", auditHandler.ListAuditLogs)

	// Apply auth middleware
	handler := auth.AuthMiddleware(jwtValidator)(apiMux)
	mux.Handle("/api/", http.StripPrefix("/api", handler))

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
