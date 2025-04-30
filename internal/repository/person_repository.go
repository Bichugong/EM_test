package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"em_test/internal/entity"
	"em_test/pkg/logging"
)

type PersonRepository struct {
	db     *sql.DB
	logger logging.Logger
}

func NewPersonRepository(db *sql.DB, logger logging.Logger) *PersonRepository {
	return &PersonRepository{
		db:     db,
		logger: logging.Logger,
	}
}

// Create создает новую запись о человеке
func (r *PersonRepository) Create(ctx context.Context, person entity.Person) (int, error) {
	query := `
		INSERT INTO people 
			(name, surname, patronymic, age, gender, nationality, created_at)
		VALUES 
			($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		person.Name,
		person.Surname,
		person.Patronymic,
		person.Age,
		person.Gender,
		person.Nationality,
		time.Now().UTC(),
	).Scan(&id)

	if err != nil {
		r.logger.Errorf("Failed to create person: %v", err)
		return 0, fmt.Errorf("repository create failed: %w", err)
	}

	r.logger.Debugf("Created person with ID: %d", id)
	return id, nil
}

// GetByFilter возвращает отфильтрованный список людей
func (r *PersonRepository) GetByFilter(ctx context.Context, filter entity.PersonFilter) ([]entity.Person, error) {
	baseQuery := `
		SELECT 
			id, name, surname, patronymic, age, gender, nationality, created_at
		FROM 
			people
		WHERE 
			1=1
	`

	args := []interface{}{}
	argPos := 1

	// Динамическое построение запроса
	if filter.ID != nil {
		baseQuery += fmt.Sprintf(" AND id = $%d", argPos)
		args = append(args, *filter.ID)
		argPos++
	}

	if filter.Name != nil {
		baseQuery += fmt.Sprintf(" AND name ILIKE $%d", argPos)
		args = append(args, "%"+*filter.Name+"%")
		argPos++
	}

	if filter.Surname != nil {
		baseQuery += fmt.Sprintf(" AND surname ILIKE $%d", argPos)
		args = append(args, "%"+*filter.Surname+"%")
		argPos++
	}

	if filter.Patronymic != nil {
		baseQuery += fmt.Sprintf(" AND patronymic ILIKE $%d", argPos)
		args = append(args, "%"+*filter.Patronymic+"%")
		argPos++
	}

	if filter.AgeMin != nil {
		baseQuery += fmt.Sprintf(" AND age >= $%d", argPos)
		args = append(args, *filter.AgeMin)
		argPos++
	}

	if filter.AgeMax != nil {
		baseQuery += fmt.Sprintf(" AND age <= $%d", argPos)
		args = append(args, *filter.AgeMax)
		argPos++
	}

	if filter.Gender != nil {
		baseQuery += fmt.Sprintf(" AND gender = $%d", argPos)
		args = append(args, *filter.Gender)
		argPos++
	}

	if filter.Nationality != nil {
		baseQuery += fmt.Sprintf(" AND nationality = $%d", argPos)
		args = append(args, *filter.Nationality)
		argPos++
	}

	// Сортировка и пагинация
	baseQuery += " ORDER BY created_at DESC"

	if filter.PageSize > 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.PageSize)
		argPos++

		if filter.Page > 1 {
			baseQuery += fmt.Sprintf(" OFFSET $%d", argPos)
			args = append(args, (filter.Page-1)*filter.PageSize)
		}
	}

	r.logger.Debugf("Executing query: %s\nWith args: %v", baseQuery, args)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		r.logger.Errorf("Query failed: %v", err)
		return nil, fmt.Errorf("repository query failed: %w", err)
	}
	defer rows.Close()

	var people []entity.Person
	for rows.Next() {
		var p entity.Person
		var createdAt time.Time
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Surname,
			&p.Patronymic,
			&p.Age,
			&p.Gender,
			&p.Nationality,
			&createdAt,
		)
		if err != nil {
			r.logger.Errorf("Row scan failed: %v", err)
			return nil, fmt.Errorf("repository scan failed: %w", err)
		}
		people = append(people, p)
	}

	if err = rows.Err(); err != nil {
		r.logger.Errorf("Rows iteration failed: %v", err)
		return nil, fmt.Errorf("repository rows iteration failed: %w", err)
	}

	r.logger.Debugf("Found %d persons", len(people))
	return people, nil
}

// Update обновляет данные человека
func (r *PersonRepository) Update(ctx context.Context, id int, person entity.Person) error {
	query := `
		UPDATE people
		SET 
			name = $1,
			surname = $2,
			patronymic = $3,
			age = $4,
			gender = $5,
			nationality = $6
		WHERE 
			id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		person.Name,
		person.Surname,
		person.Patronymic,
		person.Age,
		person.Gender,
		person.Nationality,
		id,
	)

	if err != nil {
		r.logger.Errorf("Update failed for ID %d: %v", id, err)
		return fmt.Errorf("repository update failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected failed: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Warnf("No rows affected for ID %d", id)
		return errors.New("no rows affected")
	}

	r.logger.Debugf("Updated person with ID: %d", id)
	return nil
}

// Delete удаляет запись о человеке
func (r *PersonRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM people WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Errorf("Delete failed for ID %d: %v", id, err)
		return fmt.Errorf("repository delete failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected failed: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Warnf("No rows affected for ID %d", id)
		return errors.New("no rows affected")
	}

	r.logger.Debugf("Deleted person with ID: %d", id)
	return nil
}

// WithTransaction выполняет функцию в транзакции
func (r *PersonRepository) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %w (original error: %v)", rbErr, err)
		}
		return err
	}

	return tx.Commit()
}