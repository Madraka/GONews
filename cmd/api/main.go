package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"news/cmd/api/docs" // Swagger docs
	"news/internal/cache"
	"news/internal/database"
	"news/internal/metrics"
	"news/internal/middleware"
	"news/internal/profiling"
	"news/internal/pubsub"
	"news/internal/queue"
	"news/internal/repositories"
	"news/internal/routes"
	"news/internal/server"
	"news/internal/services"
	"news/internal/tracing"
	"news/internal/version"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title News Blog API
// @version 1.0
// @description This is a comprehensive CRUD API for a news blog platform with advanced authentication and real-time features - UPDATED!
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// Note: The @host annotation will be dynamically updated based on environment PORT

// Global logger instance
var logger *middleware.Logger

// parseIntEnv parses an environment variable to int with a default value
func parseIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

func main() {
	// Handle version flag
	versionFlag := flag.Bool("version", false, "show version information")
	flag.Parse()

	if *versionFlag {
		version.PrintVersion()
		return
	}

	// Load environment variables from a .env file if it exists
	_ = godotenv.Load()

	// Initialize logger with appropriate log level
	logger = middleware.NewLogger()
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = middleware.LevelInfo
	}
	logger.LogLevel = logLevel
	logger.Info("Starting News API server with log level: " + logLevel)

	// Initialize OpenTelemetry tracing (only if enabled)
	enableTracing := os.Getenv("ENABLE_TRACING")
	if enableTracing == "true" {
		logger.Debug("Initializing OpenTelemetry tracing")
		cleanup, err := tracing.InitTracing("news-api")
		if err != nil {
			logger.Fatal("Failed to initialize tracing", err)
		}
		defer cleanup()
		logger.Debug("OpenTelemetry tracing initialized")
	} else {
		logger.Info("OpenTelemetry tracing disabled for development performance")
	}

	// Environment variables already loaded above
	logger.Debug("Environment variables loaded")

	// Connect to the database
	logger.Info("Connecting to database")

	// Build DATABASE_URL from environment variables if not already set
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Create database configuration from individual env vars
		dbHost := os.Getenv("DB_HOST")
		dbPort := parseIntEnv("DB_PORT", 5432)
		dbUser := os.Getenv("DB_USER")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")
		dbSSLMode := os.Getenv("DB_SSL_MODE")

		// Set defaults if any required DB config is missing
		if dbHost == "" {
			dbHost = "localhost"
		}
		if dbUser == "" {
			dbUser = "postgres"
		}
		if dbName == "" {
			dbName = "news"
		}
		if dbSSLMode == "" {
			dbSSLMode = "disable"
		}

		// Build DATABASE_URL
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)
		if err := os.Setenv("DATABASE_URL", dbURL); err != nil {
			log.Printf("Warning: Failed to set DATABASE_URL: %v", err)
		}
	}

	// Initialize database connection
	database.Connect()

	// Run GORM AutoMigrate for all models
	logger.Info("Running GORM AutoMigrate")
	database.AutoMigrate()
	logger.Debug("Database connection established and GORM AutoMigrate completed successfully")

	// Initialize Redis
	logger.Info("Connecting to Redis")
	if err := cache.InitRedis(); err != nil {
		logger.Fatal("Failed to connect to Redis", err)
	}
	logger.Debug("Redis connection established")

	// Initialize Ristretto cache
	logger.Info("Initializing Ristretto cache")
	if err := cache.InitRistretto(); err != nil {
		logger.Fatal("Failed to initialize Ristretto cache", err)
	}
	logger.Debug("Ristretto cache initialized")

	// Initialize Optimized Unified Cache Manager (Primary cache system)
	logger.Info("Initializing Optimized Unified Cache Manager as primary system")
	if err := cache.InitOptimizedUnifiedCache(); err != nil {
		logger.Error("Failed to initialize Optimized Unified Cache, falling back to standard cache", err)

		// Initialize standard Unified Cache Manager as fallback
		logger.Info("Initializing Standard Unified Cache Manager as fallback")
		if err := cache.InitUnifiedCache(); err != nil {
			logger.Fatal("Failed to initialize both cache systems", err)
		}
		logger.Debug("Standard Unified Cache Manager initialized as fallback (L1: Ristretto, L2: Redis)")
	} else {
		logger.Info("✅ Optimized Unified Cache Manager initialized successfully as primary system")
		logger.Debug("Primary cache features: Enhanced L1+L2 with singleflight, circuit breakers, and smart TTL")

		// Initialize standard cache as secondary system for compatibility
		logger.Debug("Initializing Standard Unified Cache Manager for compatibility")
		if err := cache.InitUnifiedCache(); err != nil {
			logger.Warning("Failed to initialize secondary standard cache (non-critical)", map[string]interface{}{"error": err.Error()})
		} else {
			logger.Debug("Secondary Standard Unified Cache Manager initialized for compatibility")
		}
	}

	// Initialize API key tiers
	logger.Debug("Initializing API key tiers")
	middleware.InitAPIKeys()
	logger.Debug("API key tiers initialized")

	// Initialize repositories
	logger.Info("Initializing repositories")
	repositories.InitializeArticleContentBlockRepository()
	logger.Debug("Article content block repository initialized")

	// Initialize translation service (keep existing working system)
	logger.Info("Initializing translation service")
	sqlDB, err := database.DB.DB()
	if err != nil {
		logger.Fatal("Failed to get SQL DB from GORM", err)
	}

	// Keep the original working TranslationService
	translationService, err := services.NewTranslationService(sqlDB, "./locales")
	if err != nil {
		logger.Fatal("Failed to initialize translation service", err)
	}

	// Also initialize UnifiedTranslationService for new features (parallel)
	unifiedTranslationService, err := services.NewUnifiedTranslationService(sqlDB, "./locales")
	if err != nil {
		logger.Error("Failed to initialize unified translation service", err)
		// Don't fail completely, just log the error
	} else {
		// Set global unified translation service for new handlers
		services.SetGlobalUnifiedTranslationService(unifiedTranslationService)
	}

	// Initialize legacy I18nService for backward compatibility
	if err := services.InitI18nService("./locales"); err != nil {
		logger.Error("Failed to initialize legacy I18n service", err)
	}

	logger.Debug("Translation services initialized")

	// Set global unified translation service
	services.SetGlobalUnifiedTranslationService(unifiedTranslationService)

	// Initialize AI service for video processing
	logger.Info("Initializing AI service for video processing")
	aiService := services.GetAIService()
	if aiService == nil {
		logger.Error("Failed to initialize AI service - video processing will be limited", fmt.Errorf("AI service is nil"))
	} else {
		logger.Debug("AI service initialized successfully")
	}

	// Initialize Video Processing Service
	logger.Info("Initializing video processing service")
	// Get storage service (already initialized in news.go init())
	storageService := services.GetStorageService() // We'll need to add this function
	if storageService == nil {
		logger.Error("Storage service not available - video processing will be disabled", fmt.Errorf("storage service is nil"))
	} else {
		videoProcessingService := services.NewVideoProcessingService(database.DB, storageService, aiService)
		videoProcessingQueue := services.NewVideoProcessingQueue(videoProcessingService)

		// Start video processing queue in background
		go func() {
			logger.Info("Starting video processing queue worker")
			if err := videoProcessingQueue.ProcessJobs(context.Background()); err != nil {
				logger.Error("Video processing queue worker stopped", err)
			}
		}()

		// Store the services globally for use in handlers
		services.SetGlobalVideoProcessingService(videoProcessingService)
		services.SetGlobalVideoProcessingQueue(videoProcessingQueue)
		logger.Debug("Video processing service initialized and queue worker started")
	}

	// Initialize Redis pub/sub notification system with translation service (original working system)
	logger.Info("Initializing Redis pub/sub notification system")
	if err := pubsub.InitNotificationHub(translationService); err != nil {
		logger.Error("Failed to initialize notification hub", err)
		// Don't fail completely, but log the error
	} else {
		logger.Debug("Redis pub/sub notification system initialized")
	}

	// Initialize AI Translation Service for the queue
	aiService = services.GetAIService()
	aiTranslationService := services.NewAITranslationService(aiService)

	// Note: Queue processing is now handled by separate worker containers
	// Initialize minimal queue manager for job enqueueing only
	logger.Info("Initializing Redis queue client for job enqueueing")

	// Create lightweight queue manager for job creation only (no workers)
	serviceContainer := &queue.ServiceContainer{
		TranslationService:     aiTranslationService,
		VideoProcessingService: services.GetGlobalVideoProcessingService(),
	}

	queueManager := queue.NewQueueManager(serviceContainer)

	// Initialize Redis connections without starting workers
	if err := queueManager.Initialize(); err != nil {
		logger.Fatal("Failed to initialize queue connections", err)
	}

	// Set global queue manager for handlers to use
	queue.SetGlobalQueueManager(queueManager)
	logger.Debug("Queue client initialized for job enqueueing")

	// Set Gin to release mode for production performance
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router with HTTP/2 optimized settings
	r := gin.New()

	// Enable HTTP/2 features
	r.UseH2C = true // Keep H2C for development

	// HTTP/2 specific optimizations
	r.Use(func(c *gin.Context) {
		// Add HTTP/2 specific headers
		c.Header("Server", "News-API-HTTP2/1.0")

		// TODO: Enable server push when static files are available
		// if pusher := c.Writer.Pusher(); pusher != nil {
		//     pusher.Push("/static/css/main.css", nil)
		//     pusher.Push("/static/js/main.js", nil)
		// }
		// c.Header("Link", "</static/css/main.css>; rel=preload; as=style")
		// c.Header("Link", "</static/js/main.js>; rel=preload; as=script")

		c.Next()
	})

	// Add custom recovery middleware to handle panics gracefully
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			logger.Error("Recovered from panic: "+err, nil)
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}))

	// Disable Gin's default logging for performance (we use custom logging)
	gin.DisableConsoleColor()

	// Set trusted proxies for security (only localhost and docker networks for development)
	if err := r.SetTrustedProxies([]string{"127.0.0.1", "::1", "172.16.0.0/12", "192.168.0.0/16", "10.0.0.0/8"}); err != nil {
		logger.Error("Failed to set trusted proxies: %v", err)
	}

	// Configure additional Gin settings for high performance
	r.MaxMultipartMemory = 16 << 20 // 16 MiB for file uploads (increased for higher throughput)

	// Add i18n middleware before other middlewares (use original working system)
	r.Use(middleware.I18nMiddleware(translationService.GetBundle()))

	// Add Prometheus metrics middleware
	r.Use(metrics.PrometheusMiddleware())

	// Setup Prometheus metrics endpoint
	metrics.SetupMetrics(r)

	// Register routes
	routes.RegisterRoutes(r)

	// Setup profiling endpoints (pprof and GC tuning)
	logger.Debug("Setting up profiling endpoints")
	profiling.SetupPprof(r)
	profiling.SetupGCTuningEndpoints(r)
	logger.Debug("Profiling endpoints configured")

	// Get the port from the environment variable or use the default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Log profiling endpoint information for developers
	if gin.Mode() == gin.DebugMode || os.Getenv("ENABLE_PPROF") == "true" {
		logger.Info("🔍 Profiling endpoints available:")
		logger.Info("  • pprof: http://localhost:" + port + "/debug/pprof/")
		logger.Info("  • GC Stats: http://localhost:" + port + "/debug/gc/stats")
		logger.Info("  • GC Analysis: http://localhost:" + port + "/debug/gc/analyze")
		logger.Info("  • Force GC: POST http://localhost:" + port + "/debug/gc/force")
		logger.Info("  • GC Tuning: POST http://localhost:" + port + "/debug/gc/tune")
	}

	// Dynamic Swagger configuration based on environment
	swaggerHost := fmt.Sprintf("localhost:%s", port)

	// Update the docs.SwaggerInfo with the correct host dynamically
	docs.SwaggerInfo.Host = swaggerHost

	// Configure Swagger with dynamic host - use the generated docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Dynamic Swagger UI endpoint that uses current port
	r.GET("/docs/swagger-ui", func(c *gin.Context) {
		currentPort := c.Request.Host
		if !strings.Contains(currentPort, ":") {
			currentPort = currentPort + ":" + port
		}

		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>News API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: 'http://%s/docs/api/swagger.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                tryItOutEnabled: true,
                requestInterceptor: function(request) {
                    request.headers['X-API-Source'] = 'swagger-ui';
                    return request;
                },
                responseInterceptor: function(response) {
                    return response;
                },
                onComplete: function() {
                    console.log('Swagger UI loaded successfully');
                }
            });
        };
    </script>
</body>
</html>`, currentPort)
		c.Header("Content-Type", "text/html")
		c.String(200, html)
	})

	// Static route to serve swagger files directly
	r.Static("/swagger-docs", "./docs")

	// Set up graceful shutdown
	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine with HTTP/2 configuration
	serverErrors := make(chan error, 1)
	go func() {
		// Create HTTP/2 server configuration
		http2Config := server.DefaultHTTP2Config()
		http2Config.Port = port

		// Override config based on environment
		if os.Getenv("ENVIRONMENT") == "production" {
			logger.Info("Starting HTTPS/2 server for production with TLS")
			http2Config.CertFile = os.Getenv("TLS_CERT_FILE")
			http2Config.KeyFile = os.Getenv("TLS_KEY_FILE")
		} else {
			logger.Info(fmt.Sprintf("Starting HTTP/2 server on port %s with H2C (cleartext) for development", port))
			// Force H2C mode for development
			http2Config.CertFile = ""
			http2Config.KeyFile = ""
			http2Config.H2CEnabled = true
		}

		// Create optimized HTTP/2 server
		srv, err := server.CreateHTTP2Server(http2Config, r)
		if err != nil {
			serverErrors <- fmt.Errorf("failed to create HTTP/2 server: %w", err)
			return
		}

		// Add custom error logging
		srv.ErrorLog = log.New(os.Stderr, "HTTP2_SERVER_ERROR: ", log.LstdFlags)

		// Configure TCP socket options for HTTP/2
		srv.SetKeepAlivesEnabled(true)

		// Create optimized listener for HTTP/2
		listener, err := server.CreateOptimizedListener(port)
		if err != nil {
			serverErrors <- fmt.Errorf("failed to create optimized listener: %w", err)
			return
		}

		// Start server based on configuration
		if http2Config.CertFile != "" && http2Config.KeyFile != "" {
			logger.Info(fmt.Sprintf("HTTPS/2 server listening on port %s with TLS", port))
			serverErrors <- srv.ServeTLS(listener, http2Config.CertFile, http2Config.KeyFile)
		} else {
			logger.Info(fmt.Sprintf("HTTP/2 server listening on port %s with H2C", port))
			serverErrors <- srv.Serve(listener)
		}
	}()

	// Wait for either server error or shutdown signal
	select {
	case err := <-serverErrors:
		logger.Fatal("Server failed to start", err)
	case sig := <-sigChan:
		logger.Info(fmt.Sprintf("Received signal %v, initiating graceful shutdown", sig))

		// Stop queue manager gracefully
		if err := queueManager.Stop(); err != nil {
			logger.Error("Error stopping queue manager", err)
		}

		// Close notification hub gracefully
		if err := pubsub.Close(); err != nil {
			logger.Error("Error closing notification hub", err)
		}

		// Close Redis connection
		if err := cache.CloseRedis(); err != nil {
			logger.Error("Error closing Redis connection", err)
		}

		// Close database connection
		if sqlDB, err := database.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				logger.Error("Error closing database connection", err)
			}
		}

		logger.Info("Graceful shutdown completed")
		os.Exit(0)
	}
}
