package storage

import (
	"database/sql"
	"go/types"
	"time"
)

type PostgresClickRepo struct {
	db *sql.DB
}

func NewPostgresClickRepo(db *sql.DB) *PostgresClickRepo {
	return &PostgresClickRepo{db: db}
}

// 1. Сохраняем клик в таблицу
func (r *PostgresClickRepo) SaveClick(authorID, userID string, date time.Time) error {
	_, err := r.db.Exec(""+
		"INSERT INTO raw_click (author_id, user_id, date) ("+
		"VALUES ($1, $2, $3)"+
		"ON CONFLICT DO NOTHING", authorID, userID, date)
	return err
}

// 2. Извелкаем клики за конкретную дату
func (r *PostgresClickRepo) GetAllClicksForDate(date time.Time) (map[string][]string, error) {
	rows, err := r.db.Query(""+
		"SELECT author_id, user_id FROM raw_clicks "+
		"WHERE created_at::date = $1", date.Format("2006-01-02"))
	if err != nil {
		return nil, err
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
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("" +
		"INSERT INTO daily_stats (date, author_id, unique_users)" +
		"VALUES ($1, $2, $3)" +
		"ON CONFLICT (date, author_id) DO UPDATE SET unique_users = $3")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for author, count := range stats {
		if _, err := stmt.Exec(date, author, count); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// 4 GetDailyStats - Получает уже сохранённую статистику для указанной даты.
func (r *PostgresClickRepo) GetDailyStats(date time.Time) (map[string]int, error) {
	rows, err := r.db.Query(""+
		"SELECT author_id, unique_users FROM daily_stats"+
		"WHERE date = $1    ", date)
	if err != nil {
		return nil, err
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
