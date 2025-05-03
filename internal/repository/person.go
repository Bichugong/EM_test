package repository

import (
	"context"
	"fmt"
	"time"

	"em_test/internal/logger"
	"em_test/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

type PersonRepository interface {
	Create(ctx context.Context, person model.Person) (model.Person, error)
	GetAll(ctx context.Context, filter model.PersonFilter, offset int) ([]model.Person, error)
	Update(ctx context.Context, id int, person model.Person) (model.Person, error)
	Delete(ctx context.Context, id int) error
}

type personRepository struct {
	pool *pgxpool.Pool
}

func NewPersonRepository(pool *pgxpool.Pool) PersonRepository {
	return &personRepository{pool: pool}
}

func (r *personRepository) Create(ctx context.Context, person model.Person) (model.Person, error) {
	logger.Logger.Debug("Creating person in repository", zap.Any("person", person))
	query := `
		INSERT INTO persons 
			(name, surname, patronymic, age, gender, nationality, created_at, updated_at) 
		VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		person.Name,
		person.Surname,
		person.Patronymic,
		person.Age,
		person.Gender,
		person.Nationality,
		time.Now(),
		time.Now(),
	).Scan(&person.ID)

	if err == nil {
		logger.Logger.Info("Person created successfully", zap.Int("id", person.ID))
	}

	return person, err
}

func (r *personRepository) GetAll(ctx context.Context, filter model.PersonFilter, offset int) ([]model.Person, error) {
	logger.Logger.Debug("Getting all persons with filter", zap.Any("filter", filter))
	query := `
		SELECT 
			id, name, surname, patronymic, age, gender, nationality, created_at, updated_at
		FROM 
			persons
		WHERE 
			($1 = '' OR gender = $1) AND
			($2 = '' OR nationality = $2) AND
			($3 = 0 OR age >= $3) AND
			($4 = 0 OR age <= $4)
		ORDER BY id
		LIMIT $5 OFFSET $6
	`

	rows, err := r.pool.Query(ctx, query,
		filter.Gender,
		filter.Nationality,
		filter.AgeFrom,
		filter.AgeTo,
		filter.Limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []model.Person
	for rows.Next() {
		var p model.Person
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Surname,
			&p.Patronymic,
			&p.Age,
			&p.Gender,
			&p.Nationality,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		persons = append(persons, p)
	}

	if err == nil {
		logger.Logger.Info("Persons retrieved successfully", zap.Int("count", len(persons)))
	}

	return persons, nil
}

func (r *personRepository) Update(ctx context.Context, id int, person model.Person) (model.Person, error) {
	logger.Logger.Debug("Updating person in repository", zap.Int("id", id), zap.Any("person", person))
	query := `
		UPDATE persons SET
			name = $1,
			surname = $2,
			patronymic = $3,
			age = $4,
			gender = $5,
			nationality = $6,
			updated_at = $7
		WHERE id = $8
		RETURNING created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		person.Name,
		person.Surname,
		person.Patronymic,
		person.Age,
		person.Gender,
		person.Nationality,
		time.Now(),
		id,
	).Scan(&person.CreatedAt, &person.UpdatedAt)

	if err == nil {
		logger.Logger.Info("Person updated successfully", zap.Int("id", id))
	}

	person.ID = id
	return person, err
}

func (r *personRepository) Delete(ctx context.Context, id int) error {
	logger.Logger.Debug("Deleting person in repository", zap.Int("id", id))
	query := `DELETE FROM persons WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)

	if err == nil {
		logger.Logger.Info("Person deleted successfully", zap.Int("id", id))
	}

	return err
}

type Migrator struct {
	logger *zap.Logger
}

func (m *Migrator) Apply(ctx context.Context, conn *pgx.Conn, migrationsDir string) error {
	logger.Logger.Debug("Starting migration", zap.String("dir", migrationsDir))

	// Настройка Goose
	goose.SetBaseFS(nil)
	goose.SetDialect("postgres")
	goose.SetLogger(m)
	goose.SetVerbose(true)

	m.logger.Info("Applying database migrations", zap.String("dir", migrationsDir))
	db := stdlib.OpenDB(*conn.Config())
	defer db.Close()

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	m.logger.Info("Migrations applied successfully")
	return nil
}

// Реализация goose.Logger interface
func (m *Migrator) Fatal(v ...interface{}) {
	m.logger.Fatal("Goose fatal", zap.Any("error", v))
}

func (m *Migrator) Fatalf(format string, v ...interface{}) {
	m.logger.Fatal("Goose fatal", zap.String("error", fmt.Sprintf(format, v...)))
}

func (m *Migrator) Print(v ...interface{}) {
	m.logger.Info("Goose log", zap.Any("message", v))
}

func (m *Migrator) Printf(format string, v ...interface{}) {
	m.logger.Info("Goose log", zap.String("message", fmt.Sprintf(format, v...)))
}
