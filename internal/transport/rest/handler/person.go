package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "em_test/internal/entity"
    "em_test/internal/service"
    "em_test/pkg/logging"
)

type PersonHandler struct {
    service *service.PersonService
    logger  logging.Logger
}

func NewPersonHandler(service *service.PersonService, logger logging.Logger) *PersonHandler {
    return &PersonHandler{service: service, logger: logger}
}

// CreatePerson godoc
// @Summary Create a new person
// @Description Create a new person with the input payload
// @Tags persons
// @Accept  json
// @Produce  json
// @Param person body entity.PersonInput true "Create person"
// @Success 201 {object} entity.Person
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /persons [post]
func (h *PersonHandler) CreatePerson(c *gin.Context) {
    var input entity.PersonInput
    if err := c.ShouldBindJSON(&input); err != nil {
        h.logger.Errorf("Failed to bind input: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    person, err := h.service.CreatePerson(input)
    if err != nil {
        h.logger.Errorf("Failed to create person: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, person)
}

// GetPersons godoc
// @Summary Get persons with filtering
// @Description Get persons with filtering and pagination
// @Tags persons
// @Accept  json
// @Produce  json
// @Param name query string false "Name filter"
// @Param surname query string false "Surname filter"
// @Param patronymic query string false "Patronymic filter"
// @Param age_min query int false "Minimum age"
// @Param age_max query int false "Maximum age"
// @Param gender query string false "Gender filter"
// @Param nationality query string false "Nationality filter"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {array} entity.Person
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /persons [get]
func (h *PersonHandler) GetPersons(c *gin.Context) {
    var filter entity.PersonFilter
    if err := c.ShouldBindQuery(&filter); err != nil {
        h.logger.Errorf("Failed to bind query: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    persons, err := h.service.GetPersons(filter)
    if err != nil {
        h.logger.Errorf("Failed to get persons: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, persons)
}

// UpdatePerson godoc
// @Summary Update a person
// @Description Update a person by ID
// @Tags persons
// @Accept  json
// @Produce  json
// @Param id path int true "Person ID"
// @Param person body entity.PersonInput true "Update person"
// @Success 200 {object} entity.Person
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /persons/{id} [put]
func (h *PersonHandler) UpdatePerson(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        h.logger.Errorf("Failed to parse id: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }

    var input entity.PersonInput
    if err := c.ShouldBindJSON(&input); err != nil {
        h.logger.Errorf("Failed to bind input: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    person, err := h.service.UpdatePerson(id, input)
    if err != nil {
        h.logger.Errorf("Failed to update person: %v", err)
        if err == service.ErrPersonNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "person not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, person)
}

// DeletePerson godoc
// @Summary Delete a person
// @Description Delete a person by ID
// @Tags persons
// @Accept  json
// @Produce  json
// @Param id path int true "Person ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /persons/{id} [delete]
func (h *PersonHandler) DeletePerson(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        h.logger.Errorf("Failed to parse id: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }

    if err := h.service.DeletePerson(id); err != nil {
        h.logger.Errorf("Failed to delete person: %v", err)
        if err == service.ErrPersonNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "person not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.Status(http.StatusNoContent)
}