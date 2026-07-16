package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type PostgresClickRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresClickRepo(pool *pgxpool.Pool) *PostgresClickRepo {
	return &PostgresClickRepo{pool: pool}
}

// 1. Сохраняем клик в таблицу
func (r *PostgresClickRepo) SaveClick(authorID, userID string, date time.Time) error {
	_, err := r.pool.Exec(context.Background(),
		"INSERT INTO raw_click (author_id, user_id, date) "+
			"VALUES ($1, $2, $3)"+
			"ON CONFLICT DO NOTHING",
		authorID, userID, date)
	return err
}

// 2. Извелкаем клики за конкретную дату
func (r *PostgresClickRepo) GetAllClicksForDate(date time.Time) (map[string][]string, error) {
	rows, err := r.pool.Query(context.Background(),
		"SELECT author_id, user_id FROM raw_clicks "+
			"WHERE created_at::date = $1",
		date.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("GetAllClicksForDate: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var author, user string
		if err := rows.Scan(&author, &user); err != nil {
			return nil, err
		}
		result[author] = append(result[author], user)
	}
	return result, rows.Err()
}

// 3 SaveDailyStats - Сохраняет агрегированную статистику за день в таблицу daily_stats.
func (r *PostgresClickRepo) SaveDailyStats(date time.Time, stats map[string]int) error {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for author, count := range stats {
		_, err := tx.Exec(ctx,
			"INSERT INTO daily_stats (date, author_id, unique_users)"+
				"VALUES ($1, $2, $3)"+
				"ON CONFLICT (date, author_id) DO UPDATE SET unique_users = $3",
			date, author, count,
		)
		if err != nil {
			return fmt.Errorf("insert daily stats for %s: %w", author, err)
		}
	}
	return tx.Commit(ctx)
}

// 4 GetDailyStats - Получает уже сохранённую статистику для указанной даты.
func (r *PostgresClickRepo) GetDailyStats(date time.Time) (map[string]int, error) {
	rows, err := r.pool.Query(context.Background(),
		"SELECT author_id, unique_users FROM daily_stats WHERE date = $1", date)
	if err != nil {
		return nil, fmt.Errorf("GetDailyStats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var author string
		var count int
		if err := rows.Scan(&author, &count); err != nil {
			return nil, err
		}
		stats[author] = count
	}
	return stats, rows.Err()
}
