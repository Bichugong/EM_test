package controller

import (
	"net/http"
	"strconv"

	"em_test/internal/logger"
	"em_test/internal/model"
	"em_test/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PersonController struct {
	service service.PersonService
}

func NewPersonController(s service.PersonService) *PersonController {
	logger.Logger.Debug("Creating new person controller", zap.Any("service", s))
	// Инициализируем контроллер с сервисом
	return &PersonController{service: s}
}

// CreatePerson - обработчик для создания новой персоны
// @Summary Создать новую запись о человеке
// @Description Принимает ФИО и сохраняет в БД с доп. данными (возраст, пол, национальность)
// @Tags persons
// @Accept  json
// @Produce  json
// @Param input body model.PersonInput true "ФИО человека"
// @Success 201 {object} model.Person
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /persons [post]
func (c *PersonController) CreatePerson(ctx *gin.Context) {
	logger.Logger.Debug("Incoming request",
		zap.String("method", ctx.Request.Method),
		zap.String("path", ctx.FullPath()),
	)
	logger.Logger.Info("Starting request processing",
		zap.String("method", ctx.Request.Method),
		zap.String("path", ctx.FullPath()),
	)

	var input model.PersonInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	person, err := c.service.Create(ctx, input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Logger.Info("Person created successfully",
		zap.Int("id", person.ID),
		zap.String("name", person.Name),
		zap.String("surname", person.Surname),
	)

	ctx.JSON(http.StatusCreated, person)
}

// GetPersons godoc
// @Summary Get filtered persons list
// @Description Get list of persons with filtering and pagination
// @Tags persons
// @Accept json
// @Produce json
// @Param gender query string false "Gender filter"
// @Param nationality query string false "Nationality filter"
// @Param age_from query int false "Minimum age"
// @Param age_to query int false "Maximum age"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {array} model.Person
// @Failure 500 {object} map[string]string
// @Router /persons [get]
func (c *PersonController) GetPersons(ctx *gin.Context) {
	logger.Logger.Debug("Incoming request",
		zap.String("method", ctx.Request.Method),
		zap.String("path", ctx.FullPath()),
	)

	gender := ctx.Query("gender")
	nationality := ctx.Query("nationality")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	ageFrom, _ := strconv.Atoi(ctx.Query("age_from"))
	ageTo, _ := strconv.Atoi(ctx.Query("age_to"))

	filter := model.PersonFilter{
		Gender:      gender,
		Nationality: nationality,
		AgeFrom:     ageFrom,
		AgeTo:       ageTo,
		Page:        page,
		Limit:       limit,
	}

	persons, err := c.service.GetAll(ctx, filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Logger.Info("Persons retrieved successfully",
		zap.Int("count", len(persons)),
		zap.Any("filter", filter),
	)

	ctx.JSON(http.StatusOK, persons)
}

// UpdatePerson - обработчик для обновления данных о человеке
// @Summary Обновить данные о человеке
// @Description Обновляет данные о человеке по его ID
// @Tags persons
// @Accept  json
// @Produce  json
// @Param id path int true "ID человека"
// @Param input body model.Person true "Обновленные данные"
// @Success 200 {object} model.Person
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /persons/{id} [put]
func (c *PersonController) UpdatePerson(ctx *gin.Context) {
	logger.Logger.Debug("Incoming request",
		zap.String("method", ctx.Request.Method),
		zap.String("path", ctx.FullPath()),
	)

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var person model.Person
	if err := ctx.ShouldBindJSON(&person); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedPerson, err := c.service.Update(ctx, id, person)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Logger.Info("Person updated successfully",
		zap.Int("id", updatedPerson.ID),
		zap.String("name", updatedPerson.Name),
		zap.String("surname", updatedPerson.Surname),
	)

	ctx.JSON(http.StatusOK, updatedPerson)
}

// DeletePerson - обработчик для удаления человека по ID
// @Summary Удалить человека
// @Description Удаляет запись о человеке по его ID
// @Tags persons
// @Accept  json
// @Produce  json
// @Param id path int true "ID человека"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /persons/{id} [delete]
func (c *PersonController) DeletePerson(ctx *gin.Context) {
	logger.Logger.Debug("Incoming request",
		zap.String("method", ctx.Request.Method),
		zap.String("path", ctx.FullPath()),
	)

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.service.Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Logger.Info("Person deleted successfully",
		zap.Int("id", id),
	)

	ctx.Status(http.StatusNoContent)
}
