package storage

import "time"

// ClickRepository описывает операции с хранилищем кликов.
type ClickRepository interface {
	// SaveClick сохраняет факт клика. Вызывается при каждом RecordClick.
	SaveClick(authorID, userID string, date time.Time) error

	// GetAllClicksForDate Извелкает клики за конкретную дату
	GetAllClicksForDate(date time.Time) (map[string][]string, error)

	// SaveDailyStats сохраняет агрегированную статистику (количество уникальных
	// пользователей по автору за конкретную дату). Используется при закрытии суток.
	SaveDailyStats(date time.Time, stats map[string]int) error

	// GetDailyStats загружает агрегированную статистику за указанную дату.
	// Возвращает map[authorID]uniqueUsers.
	GetDailyStats(date time.Time) (map[string]int, error)
}
