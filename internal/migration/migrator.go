package migration

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

type Migrator struct {
	logger *zap.Logger
}

func NewMigrator(logger *zap.Logger) *Migrator {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Migrator{
		logger: logger,
	}
}

func (m *Migrator) Apply(ctx context.Context, conn *pgx.Conn, migrationsDir string) error {
    m.logger.Info("Applying database migrations", zap.String("dir", migrationsDir))
    
    // Используем правильный способ подключения для Goose
    db := stdlib.OpenDB(*conn.Config())
    defer db.Close()

	goose.SetTableName("goose_migrations")
    goose.SetLogger(m)
    goose.SetVerbose(true)

    // Устанавливаем dialect перед выполнением миграций
    if err := goose.SetDialect("postgres"); err != nil {
        return fmt.Errorf("failed to set dialect: %w", err)
    }

    // Явно указываем тип миграций как SQL
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