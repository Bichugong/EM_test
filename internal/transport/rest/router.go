package rest

import (
	"database/sql"
	"em_test/internal/config"
	"em_test/internal/controller"
	"em_test/internal/repository"
	"em_test/internal/service"
	"em_test/internal/transport/rest/handler"
	"em_test/pkg/logging"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg     *config.Config
	db      *sql.DB
	logger  logging.Logger
	handler *handler.PersonHandler
}

func NewServer(cfg *config.Config, db *sql.DB, logger logging.Logger) *Server {
	// Инициализация зависимостей
	personRepo := repository.NewPersonRepository(db)
	enricher := service.NewEnricherService(logger)
	personService := service.NewPersonService(personRepo, enricher)
	personController := controller.NewPersonController(personService, logger)
	personHandler := handler.NewPersonHandler(personController, logger)

	return &Server{
		cfg:     cfg,
		db:      db,
		logger:  logger,
		handler: personHandler,
	}
}

func (s *Server) Run() error {
	router := gin.Default()

	// Маршруты
	api := router.Group("/api/v1")
	{
		persons := api.Group("/persons")
		{
			persons.POST("", s.handler.CreatePerson)
			persons.GET("", s.handler.GetPersons)
			persons.PUT("/:id", s.handler.UpdatePerson)
			persons.DELETE("/:id", s.handler.DeletePerson)
		}
	}

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router.Run(":" + s.cfg.AppPort)
}