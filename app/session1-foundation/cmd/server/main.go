package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	// Register routes
	// Go 1.22+ supports method in pattern: "GET /health"
	http.HandleFunc("GET /health", healthCheckHandler)
	http.HandleFunc("GET /hello/minh", helloMinhHandler)

	// Start server
	log.Println("🚀 Server starting on http://localhost:8080")
	log.Println("📍 Health check: http://localhost:8080/health")
	log.Println("📍 Hello Minh: http://localhost:8080/hello/minh")
	log.Println("Press Ctrl+C to stop")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}

}

// healthCheckHandler handles GET /health endpoint
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Set response header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Create response
	response := HealthResponse{
		Status:    "ok",
		Message:   "Mini ASM service is running",
		Timestamp: time.Now(),
	}

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If encoding fails, return 500
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Log the request
	log.Printf("Health check requested from %s", r.RemoteAddr)
}

// helloMinhHandler handles GET /hello/minh endpoint
func helloMinhHandler(w http.ResponseWriter, r *http.Request) {
	// Set response header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Create response
	response := HealthResponse{
		Status:    "ok",
		Message:   "Hello, Minh!",
		Timestamp: time.Now(),
	}

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If encoding fails, return 500
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Log the request
	log.Printf("Hello Minh requested from %s", r.RemoteAddr)
}

/*
🎓 NOTES:

1. Package main:
   - Entry point của Go application
   - Must có function main()

2. Imports:
   - encoding/json: để convert struct → JSON
   - net/http: HTTP server và handlers
   - log: logging
   - time: timestamps

3. Struct Tags:
   - `json:"status"` - định nghĩa JSON field name
   - Important cho API response format

4. HTTP Handler Function:
   - Signature: func(w http.ResponseWriter, r *http.Request)
   - w - để write response
   - r - để read request

5. Go 1.22+ Pattern Matching:
   - "GET /health" - method + path trong 1 string
   - Trước Go 1.22: phải dùng http.NewServeMux() và check method manually

6. Error Handling:
   - Always check errors
   - Return appropriate HTTP status codes
   - Don't panic in production code

📝 COMMON MISTAKES TO HIGHLIGHT:

❌ Forget to set Content-Type header
   → Browser có thể hiểu sai response format

❌ Not checking errors
   → App crash khi có unexpected issues

❌ Using wrong HTTP method
   → GET vs POST confusion
*/
