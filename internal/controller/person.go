package controller

import (
	"context"
	"errors"

	"em_test/internal/entity"
	"em_test/internal/service"
	"em_test/pkg/logging"
)

type PersonController struct {
	service service.PersonService
	logger  logging.Logger
}

func NewPersonController(service service.PersonService, logger logging.Logger) *PersonController {
	return &PersonController{
		service: service,
		logger:  logger,
	}
}

func (c *PersonController) Create(ctx context.Context, input entity.PersonInput) (*entity.Person, error) {
	c.logger.Debugf("Creating person: %+v", input)
	return c.service.CreatePerson(ctx, input)
}

func (c *PersonController) GetByFilter(ctx context.Context, filter entity.PersonFilter) ([]entity.Person, error) {
	c.logger.Debugf("Getting persons with filter: %+v", filter)
	return c.service.GetPersons(ctx, filter)
}

func (c *PersonController) Update(ctx context.Context, id int, input entity.PersonInput) (*entity.Person, error) {
	c.logger.Debugf("Updating person ID %d with data: %+v", id, input)
	return c.service.UpdatePerson(ctx, id, input)
}

func (c *PersonController) Delete(ctx context.Context, id int) error {
	c.logger.Debugf("Deleting person ID %d", id)
	return c.service.DeletePerson(ctx, id)
}

func (c *PersonController) ValidateInput(input entity.PersonInput) error {
	if input.Name == "" {
		return errors.New("name is required")
	}
	if input.Surname == "" {
		return errors.New("surname is required")
	}
	return nil
}
