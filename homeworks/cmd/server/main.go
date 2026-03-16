package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"mini-asm/internal/database"
	"mini-asm/internal/handler"
	"mini-asm/internal/service"
	"mini-asm/internal/storage/postgres"

	_ "github.com/lib/pq"
)

func main() {
	log.Println("🚀 Starting Mini ASM Server (Session 3 - Database)...")

	// ============================================
	// CONFIGURATION - Load from environment
	// ============================================

	// Database configuration
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "secops")
	dbPassword := getEnv("DB_PASSWORD", "secops123")
	dbName := getEnv("DB_NAME", "mini_asm")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	// Build connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode,
	)

	log.Printf("📊 Connecting to database: %s@%s:%s/%s", dbUser, dbHost, dbPort, dbName)

	// ============================================
	// DATABASE CONNECTION
	// ============================================

	// Open database connection
	// db, err := sql.Open("postgres", connStr)
	// if err != nil {
	// 	log.Fatal("❌ Failed to open database:", err)
	// }
	// defer db.Close()

	// // Verify connection with ping
	// if err := db.Ping(); err != nil {
	// 	log.Fatal("❌ Failed to ping database:", err)
	// }
	db, err := database.ConnectWithRetry(connStr, 5)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}

	log.Println("✅ Database connected successfully")

	// Optional: Configure connection pool
	db.SetMaxOpenConns(25)               // Maximum open connections
	db.SetMaxIdleConns(5)                // Maximum idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime (5 minutes)

	// ============================================
	// DEPENDENCY INJECTION - Wire up all layers
	// ============================================

	store := postgres.NewPostgresStorage(db)
	log.Println("✅ Storage initialized: PostgreSQL")

	assetService := service.NewAssetService(store)
	scanService := service.NewScanService(store, store)
	log.Println("✅ Services initialized: AssetService, ScanService")

	assetHandler := handler.NewAssetHandler(assetService)
	scanHandler := handler.NewScanHandler(scanService)
	healthHandler := handler.NewHealthHandler(db)
	log.Println("✅ Handlers initialized")

	// ============================================
	// ROUTING - Register HTTP endpoints
	// ============================================

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", healthHandler.Check)

	// Asset CRUD operations
	mux.HandleFunc("POST /assets", assetHandler.CreateAsset)
	mux.HandleFunc("POST /assets/batch", assetHandler.BatchCreateAssets)
	mux.HandleFunc("GET /assets/search", assetHandler.SearchAssets)
	mux.HandleFunc("GET /assets", assetHandler.ListAssets)
	mux.HandleFunc("GET /assets/stats", assetHandler.GetStats) // homework add
	mux.HandleFunc("GET /assets/count", assetHandler.Count)    // homework add
	mux.HandleFunc("GET /assets/{id}", assetHandler.GetAsset)
	mux.HandleFunc("PUT /assets/{id}", assetHandler.UpdateAsset)
	mux.HandleFunc("DELETE /assets/{id}", assetHandler.DeleteAsset)
	mux.HandleFunc("DELETE /assets/batch", assetHandler.BatchDeleteAssets)

	mux.HandleFunc("POST /assets/{id}/scan", scanHandler.StartScan)
	mux.HandleFunc("GET /scan-jobs/{id}", scanHandler.GetScanJob)
	mux.HandleFunc("GET /scan-jobs/{id}/results", scanHandler.GetScanResults)

	mux.HandleFunc("GET /assets/{id}/scans", scanHandler.ListAssetScans)
	mux.HandleFunc("GET /assets/{id}/dns", scanHandler.GetAssetDNS)
	mux.HandleFunc("GET /assets/{id}/whois", scanHandler.GetAssetWhois)
	mux.HandleFunc("GET /assets/{id}/subdomains", scanHandler.GetAssetSubdomains)

	log.Println("✅ Routes registered:")
	log.Println("   GET    /health")
	log.Println("   POST   /assets")
	log.Println("   POST   /assets/batch")
	log.Println("   GET    /assets")
	log.Println("   GET    /assets/stats")
	log.Println("   GET    /assets/count")
	log.Println("   GET    /assets/{id}")
	log.Println("   PUT    /assets/{id}")
	log.Println("   DELETE /assets/{id}")
	log.Println("   DELETE /assets/batch?ids={id1},{id2},...")

	// ============================================
	// START SERVER
	// ============================================

	port := getEnv("SERVER_PORT", "8080")
	addr := ":" + port

	log.Printf("🌐 Server listening on http://localhost%s\n", addr)
	log.Println("📖 API Documentation: see docs/api.yml")
	log.Println("🗄️  Database: PostgreSQL (persistent storage)")
	log.Println("Press Ctrl+C to stop")
	log.Println()

	handlerWithCORS := handler.CORSMiddleware(mux)

	if err := http.ListenAndServe(addr, handlerWithCORS); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}

// getEnv retrieves an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}