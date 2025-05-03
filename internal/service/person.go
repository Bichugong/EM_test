package service

import (
	"context"
	"em_test/internal/api"
	"em_test/internal/logger"
	"em_test/internal/model"
	"em_test/internal/repository"

	"go.uber.org/zap"
)

type PersonFilter struct {
	Gender      string
	Nationality string
	AgeFrom     int
	AgeTo       int
	Page        int
	Limit       int
}

type PersonService interface {
	Create(ctx context.Context, input model.PersonInput) (model.Person, error)
	GetAll(ctx context.Context, filter model.PersonFilter) ([]model.Person, error)
	Update(ctx context.Context, id int, person model.Person) (model.Person, error)
	Delete(ctx context.Context, id int) error
}

type personService struct {
	repo       repository.PersonRepository
	enrichment *api.EnrichmentAPI
}

func NewPersonService(repo repository.PersonRepository, enrichment *api.EnrichmentAPI) PersonService {
	logger.Logger.Debug("Creating new person service", zap.Any("repo", repo), zap.Any("enrichment", enrichment))
	return &personService{
		repo:       repo,
		enrichment: enrichment,
	}
}

func (s *personService) Create(ctx context.Context, input model.PersonInput) (model.Person, error) {
	logger.Logger.Debug("Creating person", zap.Any("input", input))

	person := model.Person{
		Name:       input.Name,
		Surname:    input.Surname,
		Patronymic: input.Patronymic,
	}

	// Обогащение данных
	age, err := s.enrichment.GetAge(input.Name)
	if err != nil {
		logger.Logger.Warn("Failed to get age", zap.Error(err))
	} else if age > 0 {
		person.Age = age
		logger.Logger.Info("Age enrichment successful", zap.Int("age", age))
	}

	gender, err := s.enrichment.GetGender(input.Name)
	if err != nil {
		logger.Logger.Warn("Failed to get gender", zap.Error(err))
	} else if gender != "" {
		person.Gender = gender
		logger.Logger.Info("Gender enrichment successful", zap.String("gender", gender))
	}

	nationality, err := s.enrichment.GetNationality(input.Name)
	if err != nil {
		logger.Logger.Warn("Failed to get nationality", zap.Error(err))
	} else if nationality != "" {
		person.Nationality = nationality
		logger.Logger.Info("Nationality enrichment successful", zap.String("nationality", nationality))
	}

	// Сохраняем в БД
	result, err := s.repo.Create(ctx, person)
	if err != nil {
		logger.Logger.Error("Failed to create person", zap.Error(err))
	} else {
		logger.Logger.Info("Person created successfully", zap.Int("id", result.ID))
	}

	return result, err
}

func (s *personService) GetAll(ctx context.Context, filter model.PersonFilter) ([]model.Person, error) {
	logger.Logger.Debug("Getting all persons", zap.Any("filter", filter))
	offset := (filter.Page - 1) * filter.Limit
	persons, err := s.repo.GetAll(ctx, filter, offset)
	if err == nil {
		logger.Logger.Info("Persons retrieved successfully", zap.Int("count", len(persons)))
	}
	return persons, err
}

func (s *personService) Update(ctx context.Context, id int, person model.Person) (model.Person, error) {
	logger.Logger.Debug("Updating person", zap.Int("id", id), zap.Any("person", person))
	updatedPerson, err := s.repo.Update(ctx, id, person)
	if err == nil {
		logger.Logger.Info("Person updated successfully", zap.Int("id", id))
	}
	return updatedPerson, err
}

func (s *personService) Delete(ctx context.Context, id int) error {
	logger.Logger.Debug("Deleting person", zap.Int("id", id))
	err := s.repo.Delete(ctx, id)
	if err == nil {
		logger.Logger.Info("Person deleted successfully", zap.Int("id", id))
	}
	return err
}
