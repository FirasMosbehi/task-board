package main

import (
	"context"
	"log"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	meter metric.Meter
	
	// HTTP metrics
	requestCount       metric.Int64Counter
	latency            metric.Float64Histogram
	requestSize        metric.Int64Histogram
	responseSize       metric.Int64Histogram
	activeRequests     metric.Int64UpDownCounter
	
	// Database metrics
	dbOperations       metric.Int64Counter
	dbOperationLatency metric.Float64Histogram
	dbConnectionsOpen  metric.Int64UpDownCounter
	
	// Application metrics
	taskCount          metric.Int64UpDownCounter
	completedTaskCount metric.Int64UpDownCounter
	
	// System metrics
	memoryUsage        metric.Int64UpDownCounter
	goroutineCount     metric.Int64UpDownCounter
)

func initTracer(ctx context.Context) func() {
	endpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317")
	log.Printf("Attempting to connect to OTEL collector for tracing at: %s", endpoint)
	
	conn, err := grpc.Dial(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("OTEL trace dial failed: %v", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		log.Fatalf("OTEL trace exporter failed: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("taskboard-backend"),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", "development"),
		)),
	)

	otel.SetTracerProvider(tp)
	log.Println("✅ Trace provider successfully initialized")
	return func() { _ = tp.Shutdown(ctx) }
}

func initMetrics(ctx context.Context) func() {
	endpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317")
	log.Printf("Attempting to connect to OTEL collector for metrics at: %s", endpoint)
	
	conn, err := grpc.Dial(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("OTEL metrics dial failed: %v", err)
	}

	exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		log.Fatalf("OTEL metrics exporter failed: %v", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(10*time.Second), // Reduce interval for more frequent exports
			),
		),
		sdkmetric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("taskboard-backend"),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", "development"),
		)),
	)

	otel.SetMeterProvider(mp)
	meter = mp.Meter("taskboard-backend")
	
	// Initialize all metrics
	initializeMetrics()
	
	// Start system metrics collection
	go collectSystemMetrics(ctx)
	
	log.Println("✅ Meter provider successfully initialized")
	return func() { _ = mp.Shutdown(ctx) }
}

// Initialize all metrics instruments
func initializeMetrics() {
	var err error
	
	// HTTP metrics
	requestCount, err = meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		log.Fatalf("Failed to create request counter: %v", err)
	}

	latency, err = meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		log.Fatalf("Failed to create latency histogram: %v", err)
	}
	
	requestSize, err = meter.Int64Histogram(
		"http_request_size_bytes",
		metric.WithDescription("HTTP request size in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		log.Fatalf("Failed to create request size histogram: %v", err)
	}
	
	responseSize, err = meter.Int64Histogram(
		"http_response_size_bytes",
		metric.WithDescription("HTTP response size in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		log.Fatalf("Failed to create response size histogram: %v", err)
	}
	
	activeRequests, err = meter.Int64UpDownCounter(
		"http_active_requests",
		metric.WithDescription("Number of active HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		log.Fatalf("Failed to create active requests counter: %v", err)
	}
	
	// Database metrics
	dbOperations, err = meter.Int64Counter(
		"db_operations_total",
		metric.WithDescription("Total number of database operations"),
		metric.WithUnit("{operation}"),
	)
	if err != nil {
		log.Fatalf("Failed to create db operations counter: %v", err)
	}
	
	dbOperationLatency, err = meter.Float64Histogram(
		"db_operation_duration_seconds",
		metric.WithDescription("Duration of database operations in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		log.Fatalf("Failed to create db latency histogram: %v", err)
	}
	
	dbConnectionsOpen, err = meter.Int64UpDownCounter(
		"db_connections_open",
		metric.WithDescription("Number of open database connections"),
		metric.WithUnit("{connection}"),
	)
	if err != nil {
		log.Fatalf("Failed to create db connections counter: %v", err)
	}
	

	// Application metrics
	taskCount, err = meter.Int64UpDownCounter(
		"tasks_total",
		metric.WithDescription("Total number of tasks in the system"),
		metric.WithUnit("{task}"),
	)
	if err != nil {
		log.Fatalf("Failed to create task counter: %v", err)
	}
	
	completedTaskCount, err = meter.Int64UpDownCounter(
		"tasks_completed_total",
		metric.WithDescription("Number of completed tasks"),
		metric.WithUnit("{task}"),
	)
	if err != nil {
		log.Fatalf("Failed to create completed task counter: %v", err)
	}
	
	// System metrics
	memoryUsage, err = meter.Int64UpDownCounter(
		"memory_usage_bytes",
		metric.WithDescription("Memory usage of the application in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		log.Fatalf("Failed to create memory usage counter: %v", err)
	}
	
	goroutineCount, err = meter.Int64UpDownCounter(
		"goroutine_count",
		metric.WithDescription("Number of goroutines"),
		metric.WithUnit("{goroutine}"),
	)
	if err != nil {
		log.Fatalf("Failed to create goroutine counter: %v", err)
	}
	
	log.Println("✅ All metrics instruments created successfully")
}

// Periodically collect system metrics
func collectSystemMetrics(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Get memory stats
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			
			// Update memory usage
			memoryUsage.Add(ctx, int64(memStats.Alloc),
				metric.WithAttributes(
					attribute.String("type", "heap"),
				),
			)
			
			// Update goroutine count
			goroutineCount.Add(ctx, int64(runtime.NumGoroutine())-int64(runtime.NumGoroutine()),
				metric.WithAttributes(
					attribute.String("type", "user"),
				),
			)
			goroutineCount.Add(ctx, int64(runtime.NumGoroutine()),
				metric.WithAttributes(
					attribute.String("type", "user"),
				),
			)
			
		case <-ctx.Done():
			return
		}
	}
}

// --- Middleware ---

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Increment active requests
		activeRequests.Add(c.Request.Context(), 1)
		
		// Record request size if Content-Length is available
		if c.Request.ContentLength > 0 {
			requestSize.Record(c.Request.Context(), c.Request.ContentLength,
				metric.WithAttributes(
					attribute.String("method", c.Request.Method),
					attribute.String("path", c.FullPath()),
				),
			)
		}
		
		// Create a responseWriter that captures the response size
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Calculate duration
		duration := time.Since(start).Seconds()
		
		// Record metrics
		log.Printf("Recording metric: method=%s, path=%s, status=%d, duration=%.3fs", 
			c.Request.Method, c.FullPath(), c.Writer.Status(), duration)

		requestCount.Add(c.Request.Context(), 1,
			metric.WithAttributes(
				attribute.String("method", c.Request.Method),
				attribute.String("path", c.FullPath()),
				attribute.Int("status", c.Writer.Status()),
			),
		)

		latency.Record(c.Request.Context(), duration,
			metric.WithAttributes(
				attribute.String("method", c.Request.Method),
				attribute.String("path", c.FullPath()),
				attribute.Int("status", c.Writer.Status()),
			),
		)
		
		// Record response size
		responseSize.Record(c.Request.Context(), int64(c.Writer.Size()),
			metric.WithAttributes(
				attribute.String("method", c.Request.Method),
				attribute.String("path", c.FullPath()),
				attribute.Int("status", c.Writer.Status()),
			),
		)
		
		// Decrement active requests
		activeRequests.Add(c.Request.Context(), -1)
	}
}

// DatabaseMetricsMiddleware wraps database operations with metrics
func TrackDBOperation(ctx context.Context, operation string, f func() error) error {
	start := time.Now()
	
	// Increment operation counter
	dbOperations.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("operation", operation),
		),
	)
	
	// Execute the operation
	err := f()
	
	// Record duration
	duration := time.Since(start).Seconds()
	dbOperationLatency.Record(ctx, duration,
		metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.Bool("success", err == nil),
		),
	)
	
	return err
}

// UpdateTaskMetrics updates task-related metrics
func UpdateTaskMetrics(ctx context.Context) {
	// Count total tasks
	var totalCount int64
	DB.Model(&Task{}).Count(&totalCount)
	
	// Reset and update the task counter
	taskCount.Add(ctx, -totalCount)
	taskCount.Add(ctx, totalCount)
	
	// Count completed tasks
	var completedCount int64
	DB.Model(&Task{}).Where("completed = ?", true).Count(&completedCount)
	
	// Reset and update the completed task counter
	completedTaskCount.Add(ctx, -completedCount)
	completedTaskCount.Add(ctx, completedCount)
}