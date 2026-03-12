package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// ConnectWithRetry tries to connect to database with retry and exponential backoff
func ConnectWithRetry(dsn string, maxRetries int) (*sql.DB, error) {

	var db *sql.DB
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {

		log.Printf("🔄 Database connection attempt %d/%d...", attempt, maxRetries)

		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open db: %w", err)
		}

		err = db.Ping()
		if err == nil {
			log.Println("✅ Database connected successfully!")
			return db, nil
		}

		if attempt == maxRetries {
			break
		}

		backoff := time.Duration(1<<uint(attempt-1)) * time.Second

		log.Printf("⚠️  Connection failed: %v. Retrying in %v...", err, backoff)

		time.Sleep(backoff)
	}

	return nil, fmt.Errorf("failed to connect database after %d attempts: %w", maxRetries, err)
}