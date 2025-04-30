package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"em_test/internal/entity"
	"em_test/internal/repository"
	"em_test/pkg/logging"
)

// Сервисные ошибки
var (
	ErrPersonNotFound      = errors.New("person not found")
	ErrInvalidInput        = errors.New("invalid input data")
	ErrEnrichmentFailed    = errors.New("failed to enrich data")
	ErrRepositoryOperation = errors.New("repository operation failed")
)

type PersonService struct {
	repo      repository.PersonRepository
	enricher  EnricherService
	logger    logging.Logger
	timeout   time.Duration
}

func NewPersonService(
	repo repository.PersonRepository, 
	enricher EnricherService,
	logger logging.Logger,
) *PersonService {
	return &PersonService{
		repo:     repo,
		enricher: enricher,
		logger:   logging.Logger,
		timeout:  5 * time.Second,
	}
}

// CreatePerson создает новую запись с обогащенными данными
func (s *PersonService) CreatePerson(ctx context.Context, input entity.PersonInput) (*entity.Person, error) {
	// Валидация входа
	if err := validatePersonInput(input); err != nil {
		s.logger.Warnf("Validation failed for input %+v: %v", input, err)
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Создаем контекст с таймаутом
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Обогащаем данные
	age, gender, nationality, err := s.enricher.Enrich(timeoutCtx, input.Name)
	if err != nil {
		s.logger.Errorf("Enrichment failed for %s %s: %v", input.Name, input.Surname, err)
		return nil, fmt.Errorf("%w: %v", ErrEnrichmentFailed, err)
	}

	person := entity.Person{
		Name:        input.Name,
		Surname:     input.Surname,
		Patronymic:  input.Patronymic,
		Age:         age,
		Gender:      gender,
		Nationality: nationality,
	}

	// Сохраняем в БД
	id, err := s.repo.Create(timeoutCtx, person)
	if err != nil {
		s.logger.Errorf("Failed to create person: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrRepositoryOperation, err)
	}

	person.ID = id
	s.logger.Infof("Created person ID %d: %s %s", id, person.Name, person.Surname)
	return &person, nil
}

// GetPersons возвращает список людей с фильтрацией
func (s *PersonService) GetPersons(ctx context.Context, filter entity.PersonFilter) ([]entity.Person, error) {
	// Валидация пагинации
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 10
	}

	persons, err := s.repo.GetByFilter(ctx, filter)
	if err != nil {
		s.logger.Errorf("Failed to get persons with filter %+v: %v", filter, err)
		return nil, fmt.Errorf("%w: %v", ErrRepositoryOperation, err)
	}

	if len(persons) == 0 {
		s.logger.Debugf("No persons found with filter %+v", filter)
		return []entity.Person{}, nil
	}

	s.logger.Debugf("Retrieved %d persons with filter %+v", len(persons), filter)
	return persons, nil
}

// UpdatePerson обновляет данные человека
func (s *PersonService) UpdatePerson(ctx context.Context, id int, input entity.PersonInput) (*entity.Person, error) {
	// Валидация входа
	if err := validatePersonInput(input); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Получаем текущие данные
	persons, err := s.repo.GetByFilter(ctx, entity.PersonFilter{ID: &id, PageSize: 1})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRepositoryOperation, err)
	}
	if len(persons) == 0 {
		return nil, ErrPersonNotFound
	}

	// Обновляем только основные поля (без переобогащения)
	person := persons[0]
	person.Name = input.Name
	person.Surname = input.Surname
	person.Patronymic = input.Patronymic

	if err := s.repo.Update(ctx, id, person); err != nil {
		s.logger.Errorf("Failed to update person ID %d: %v", id, err)
		return nil, fmt.Errorf("%w: %v", ErrRepositoryOperation, err)
	}

	s.logger.Infof("Updated person ID %d", id)
	return &person, nil
}

// DeletePerson удаляет запись
func (s *PersonService) DeletePerson(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Errorf("Failed to delete person ID %d: %v", id, err)
		return fmt.Errorf("%w: %v", ErrRepositoryOperation, err)
	}

	s.logger.Infof("Deleted person ID %d", id)
	return nil
}

// validatePersonInput проверяет обязательные поля
func validatePersonInput(input entity.PersonInput) error {
	if input.Name == "" {
		return errors.New("name is required")
	}
	if input.Surname == "" {
		return errors.New("surname is required")
	}
	if len(input.Name) > 100 || len(input.Surname) > 100 {
		return errors.New("name and surname must be less than 100 characters")
	}
	if input.Patronymic != nil && len(*input.Patronymic) > 100 {
		return errors.New("patronymic must be less than 100 characters")
	}
	return nil
}