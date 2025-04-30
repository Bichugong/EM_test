package service

import (
	"context"
	"errors"

	"em_test/internal/entity"
	"em_test/internal/repository"
)

var (
	ErrPersonNotFound = errors.New("person not found")
)

type PersonService struct {
	repo     *repository.PersonRepository
	enricher *EnricherService
}

func NewPersonService(repo *repository.PersonRepository, enricher *EnricherService) *PersonService {
	return &PersonService{repo: repo, enricher: enricher}
}

func (s *PersonService) CreatePerson(ctx context.Context, input entity.PersonInput) (*entity.Person, error) {
	// Обогащение данных
	age, gender, nationality, err := s.enricher.Enrich(input.Name)
	if err != nil {
		return nil, err
	}

	p := entity.Person{
		Name:        input.Name,
		Surname:     input.Surname,
		Patronymic:  input.Patronymic,
		Age:         age,
		Gender:      gender,
		Nationality: nationality,
	}

	id, err := s.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	p.ID = id
	return &p, nil
}

func (s *PersonService) GetPersons(ctx context.Context, filter entity.PersonFilter) ([]entity.Person, error) {
	return s.repo.GetByFilter(ctx, filter)
}

func (s *PersonService) UpdatePerson(ctx context.Context, id int, input entity.PersonInput) (*entity.Person, error) {
	// Проверка существования
	persons, err := s.repo.GetByFilter(ctx, entity.PersonFilter{ID: &id, PageSize: 1})
	if err != nil || len(persons) == 0 {
		return nil, ErrPersonNotFound
	}
}