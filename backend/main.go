package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func main() {
	initDB()

	ctx := context.Background()

	// --- Init OTEL ---
	shutdownTracer := initTracer(ctx)
	defer shutdownTracer()

	shutdownMetrics := initMetrics(ctx)
	defer shutdownMetrics()

	// Initialize DB connection metrics
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Start tracking DB connections
	go trackDBConnections(ctx, sqlDB)

	// Initial task metrics
	UpdateTaskMetrics(ctx)

	r := gin.Default()

	// --- CORS middleware ---
	frontendOrigin := "http://localhost"
	if envOrigin := os.Getenv("FRONTEND_ORIGIN"); envOrigin != "" {
		frontendOrigin = envOrigin
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// --- Tracing middleware ---
	r.Use(otelgin.Middleware("taskboard-backend"))

	// --- Metrics middleware (must come after instrument creation) ---
	r.Use(MetricsMiddleware())

	// CORS, routes...
	api := r.Group("/api")
	{
		api.GET("/tasks", getTasks)
		api.POST("/tasks", createTask)
		api.PUT("/tasks/:id", updateTask)
		api.DELETE("/tasks/:id", deleteTask)
	}

	// Add debug endpoints to test metrics generation
	debug := r.Group("/debug")
	{
		// Basic metric test
		debug.GET("/metrics", func(c *gin.Context) {
			requestCount.Add(c.Request.Context(), 1,
				metric.WithAttributes(
					attribute.String("method", "DEBUG"),
					attribute.String("path", "/debug/metrics"),
					attribute.Int("status", 200),
				),
			)

			c.JSON(200, gin.H{
				"message": "Debug metric recorded",
				"info":    "Check Prometheus or the collector debug output",
			})
		})

		// Simulate high latency
		debug.GET("/slow", func(c *gin.Context) {
			// Sleep between 1-5 seconds
			sleepTime := 1 + rand.Intn(4)
			time.Sleep(time.Duration(sleepTime) * time.Second)

			c.JSON(200, gin.H{
				"message":       "Slow response simulated",
				"sleep_seconds": sleepTime,
			})
		})

		// Get system stats
		debug.GET("/stats", func(c *gin.Context) {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			stats := gin.H{
				"goroutines": runtime.NumGoroutine(),
				"memory": gin.H{
					"alloc_bytes":       memStats.Alloc,
					"total_alloc_bytes": memStats.TotalAlloc,
					"sys_bytes":         memStats.Sys,
					"heap_objects":      memStats.HeapObjects,
				},
				"tasks": gin.H{
					"count":     0,
					"completed": 0,
				},
			}

			// Get task counts
			var totalCount, completedCount int64
			DB.Model(&Task{}).Count(&totalCount)
			DB.Model(&Task{}).Where("completed = ?", true).Count(&completedCount)

			stats["tasks"].(gin.H)["count"] = totalCount
			stats["tasks"].(gin.H)["completed"] = completedCount

			c.JSON(200, stats)
		})

		// Generate random tasks for testing
		debug.POST("/generate-tasks", func(c *gin.Context) {
			var count int = 10 // Default

			// Parse count from query if provided
			countParam := c.Query("count")
			if countParam != "" {
				parsedCount, err := strconv.Atoi(countParam)
				if err == nil && parsedCount > 0 && parsedCount <= 100 {
					count = parsedCount
				}
			}

			// Generate tasks
			for i := 0; i < count; i++ {
				title := fmt.Sprintf("Generated Task #%d", i+1)
				completed := rand.Intn(2) == 1 // 50% chance of being completed

				task := Task{
					Title:     title,
					Completed: completed,
				}

				err := TrackDBOperation(c.Request.Context(), "create_task", func() error {
					return DB.Create(&task).Error
				})

				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tasks"})
					return
				}
			}

			// Update metrics
			UpdateTaskMetrics(c.Request.Context())

			c.JSON(200, gin.H{
				"message": "Tasks generated successfully",
				"count":   count,
			})
		})

		// Clear all tasks
		debug.DELETE("/clear-tasks", func(c *gin.Context) {
			err := TrackDBOperation(c.Request.Context(), "delete_all_tasks", func() error {
				return DB.Exec("DELETE FROM tasks").Error
			})

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear tasks"})
				return
			}

			// Update metrics
			UpdateTaskMetrics(c.Request.Context())

			c.JSON(200, gin.H{
				"message": "All tasks cleared",
			})
		})
	}

	log.Println("🚀 Running backend on :8080")
	r.Run(":8080")
}

// Track database connections
func trackDBConnections(ctx context.Context, db *sql.DB) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := db.Stats()

			// Reset counter and set to current value
			dbConnectionsOpen.Add(ctx, -int64(stats.OpenConnections))
			dbConnectionsOpen.Add(ctx, int64(stats.OpenConnections))

			// Additional DB stats as attributes
			dbConnectionsOpen.Add(ctx, 0,
				metric.WithAttributes(
					attribute.Int("idle", stats.Idle),
					attribute.Int("in_use", stats.InUse),
					attribute.Int("max_open", stats.MaxOpenConnections),
				),
			)

		case <-ctx.Done():
			return
		}
	}
}
