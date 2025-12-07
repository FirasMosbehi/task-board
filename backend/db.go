// Package main contains the backend server logic for TaskBoard, including
// database initialization, environment variable handling, and application startup.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global GORM database connection used throughout the backend.
var DB *gorm.DB

// initDB initializes the PostgreSQL connection using environment variables
// and runs automatic migrations for all database models.
func initDB() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "taskboard")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	// Create a custom logger for GORM that records metrics
	customLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // Slow SQL threshold
			LogLevel:                  logger.Info,            // Log level
			IgnoreRecordNotFoundError: false,                  // Log record not found error
			Colorful:                  true,                   // Disable color
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: customLogger,
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(&Task{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Add OpenTelemetry instrumentation to GORM
	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		log.Fatalf("failed to add OTEL instrumentation to GORM: %v", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get database connection: %v", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	log.Println("✅ Connected to PostgreSQL and migrated")

	// Record initial DB metrics if meter is initialized
	if meter != nil {
		ctx := context.Background()
		dbConnectionsOpen.Add(ctx, int64(sqlDB.Stats().OpenConnections),
			metric.WithAttributes(
				attribute.String("db_name", dbName),
				attribute.String("db_host", dbHost),
			),
		)
	}
}

// getEnv returns an environment variable value or a default
// value if the variable is not set.
func getEnv(key, def string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return def
	}
	return val
}
