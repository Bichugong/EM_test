package repository

import (
	"context"
	"database/sql"
	"fmt"

	"em_test/internal/entity"
)

type PersonRepository struct {
	db *sql.DB
}

func NewPersonRepository(db *sql.DB) *PersonRepository {
	return &PersonRepository{db: db}
}

func (r *PersonRepository) Create(ctx context.Context, p entity.Person) (int, error) {
	query := `
		INSERT INTO people (name, surname, patronymic, age, gender, nationality)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	var id int
	err := r.db.QueryRowContext(ctx, query,
		p.Name, p.Surname, p.Patronymic, p.Age, p.Gender, p.Nationality,
	).Scan(&id)
	return id, err
}

func (r *PersonRepository) GetByFilter(ctx context.Context, filter entity.PersonFilter) ([]entity.Person, error) {
	query := `
		SELECT id, name, surname, patronymic, age, gender, nationality
		FROM people
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if filter.Name != nil {
		query += fmt.Sprintf(" AND name = $%d", argPos)
		args = append(args, *filter.Name)
		argPos++
	}

	// ... (аналогично для других фильтров)

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []entity.Person
	for rows.Next() {
		var p entity.Person
		if err := rows.Scan(&p.ID, &p.Name, &p.Surname, &p.Patronymic, &p.Age, &p.Gender, &p.Nationality); err != nil {
			return nil, err
		}
		persons = append(persons, p)
	}

	return persons, nil
}

func (r *PersonRepository) Update(ctx context.Context, id int, p entity.Person) error {
	query := `
		UPDATE people
		SET name = $1, surname = $2, patronymic = $3, age = $4, gender = $5, nationality = $6
		WHERE id = $7
	`
	_, err := r.db.ExecContext(ctx, query,
		p.Name, p.Surname, p.Patronymic, p.Age, p.Gender, p.Nationality, id,
	)
	return err
}

func (r *PersonRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM people WHERE id = $1", id)
	return err
}