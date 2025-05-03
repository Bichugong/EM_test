package main

import (
	"context"
	_ "em_test/docs"
	"em_test/internal/api"
	"em_test/internal/config"
	"em_test/internal/controller"
	"em_test/internal/logger"
	"em_test/internal/migration"
	"em_test/internal/repository"
	"em_test/internal/service"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @title People API
// @version 1.0
// @description API for managing people data with enrichment
func main() {
	// Загрузка конфигурации
	log := logger.Get()
	defer logger.Sync()

	log.Info("Starting application")
	log.Debug("Loading configuration")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config: ", zap.Error(err))
	}
	log.Info("Configuration loaded successfully", zap.Any("config", cfg))

	// Инициализация логгера
	if err := logger.Init(cfg.LogLevel); err != nil {
		log.Fatal("Failed to initialize logger", zap.Error(err))
	}
	log.Info("Logger initialized successfully", zap.String("level", cfg.LogLevel))
	defer logger.Logger.Sync()

	// Контекст для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Инициализация подключения к БД
	log.Debug("Initializing database connection")
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)
	log.Debug("Database connection string", zap.String("host", cfg.DB.Host), zap.String("port", cfg.DB.Port))

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		logger.Logger.Fatal("Failed to parse DB config", zap.Error(err))
	}
	poolConfig.MaxConns = cfg.DB.MaxConns

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Logger.Fatal("Unable to create connection pool", zap.Error(err))
	}
	log.Info("Database connection pool created successfully")
	defer pool.Close()

	// Проверка подключения
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("Unable to ping database", zap.Error(err))
	}
	log.Info("Database ping successful")

	// Применение миграций
	migrator := migration.NewMigrator(logger.Logger)
	conn, err := pool.Acquire(ctx)
	if err != nil {
		logger.Logger.Fatal("Failed to acquire connection", zap.Error(err))
	}
	defer conn.Release()

	if err := migrator.Apply(ctx, conn.Conn(), "migrations"); err != nil {
		logger.Logger.Fatal("Failed to apply migrations", zap.Error(err))
	}
	log.Info("Migrations applied successfully")

	// Инициализация клиентов для внешних API
	enrichmentAPI := api.NewEnrichmentAPI(cfg.EnrichmentAPI)
	log.Info("Enrichment API client initialized successfully")

	// Инициализация слоёв
	personRepo := repository.NewPersonRepository(pool)
	personService := service.NewPersonService(personRepo, enrichmentAPI)
	personController := controller.NewPersonController(personService)

	// Настройка роутера
	router := gin.Default()
	log.Info("Router initialized successfully")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Маршруты
	apiGroup := router.Group("/api/v1")
	{
		apiGroup.POST("/persons", personController.CreatePerson)
		apiGroup.GET("/persons", personController.GetPersons)
		apiGroup.PUT("/persons/:id", personController.UpdatePerson)
		apiGroup.DELETE("/persons/:id", personController.DeletePerson)
	}

	// Запуск сервера с graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	go func() {
		log.Info("Server is running", zap.String("port", cfg.AppPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Ожидание сигнала завершения
	<-ctx.Done()
	stop()
	log.Info("Termination signal received, shutting down gracefully")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exiting")
}
