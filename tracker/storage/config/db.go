package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

// InitDB создаёт пул соединений и применяет миграции.
// Принимает готовую строку подключения (DSN) в формате, понятном pgxpool.
func InitDB(databaseURL string) error {
	var err error
	Pool, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return fmt.Errorf("не удалось создать пул: %w", err)
	}

	if err = Pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("нет соединения с БД: %w", err)
	}

	log.Println("Подключение к PostgreSQL установлено")
	return runMigrations(databaseURL)
}

// runMigrations применяет файлы миграций из папки migrations
func runMigrations(databaseURL string) error {
	m, err := migrate.New("file://tracker/storage/migrations", databaseURL)
	if err != nil {
		return fmt.Errorf("миграции не инициализированы: %w", err)
	}
	defer m.Close() // закрываем ресурс после работы

	err = m.Up()
	if err == nil {
		log.Println("✅ Миграции успешно применены")
		return nil
	}
	if err == migrate.ErrNoChange {
		log.Println("ℹ️ Нет новых миграций для применения (база уже актуальна)")
		return nil
	}
	return fmt.Errorf("❌ ошибка применения миграций: %w", err)
}
