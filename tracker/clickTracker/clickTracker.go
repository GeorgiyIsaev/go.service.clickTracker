package clickTracker

import (
	"fmt"
	"sync"
	"time"

	"go.service.clickTracker/tracker/storage"
)

// Структура для хранения кликов
type ClickTracker struct {
	mu             sync.RWMutex
	todayClicks    map[string]map[string]struct{} // authorID -> set of userID
	yesterdayStats map[string]int                 //только количество
	currentDate    string                         // "YYYY-MM-DD"
	repo           storage.ClickRepository
}

// Экзмпляр класс ClickTracker
func NewClickTracker(repo storage.ClickRepository) (*ClickTracker, error) {
	now := time.Now()
	ct := &ClickTracker{
		todayClicks:    make(map[string]map[string]struct{}),
		yesterdayStats: make(map[string]int),
		currentDate:    now.Format("2006-01-02"),
		repo:           repo,
	}
	// 1. Загружаем клики за сегодня, чтобы восстановить todayClicks
	if err := ct.loadTodayFromDB(); err != nil {
		return nil, err
	}

	// 2. Загружаем агрегированную статистику за вчерашний день
	yesterday := now.AddDate(0, 0, -1)
	stats, err := repo.GetDailyStats(yesterday)
	if err != nil {
		return nil, fmt.Errorf("load yesterday stats: %w", err)
	}
	ct.yesterdayStats = stats // если stats == nil, останется пустая мапа

	return ct, nil
}

// Регисрация клика пользователя
func (ct *ClickTracker) RecordClick(authorID, userID string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.dataShift()

	users, ok := ct.todayClicks[authorID]
	if !ok {
		users = make(map[string]struct{})
		ct.todayClicks[authorID] = users
	}
	users[userID] = struct{}{}

	// Сохраняем в БД (синхронно для гарантии, в реальности можно асинхронно с буфером)
	now := time.Now()
	if err := ct.repo.SaveClick(authorID, userID, now); err != nil {
		// Логируем ошибку, но не роняем сервис
		fmt.Printf("ERROR saving click: %v\n", err)
	}
}

// Возаращает количество уникальных пользователей для всех авторов из списка
func (ct *ClickTracker) GetAuthorsStatus(authorsIDs []string) map[string]int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	result := make(map[string]int, len(authorsIDs))
	for _, authorID := range authorsIDs {
		if users, ok := ct.todayClicks[authorID]; ok {
			result[authorID] = len(users)
		} else {
			result[authorID] = 0
		}
	}
	return result
}

// Сдвиг даты при наступлении новых суток
func (ct *ClickTracker) dataShift() {
	today := time.Now().Format("2006-01-02")
	if today == ct.currentDate {
		return
	}

	// Агрегируем статистику за завершившийся день
	previousDate, _ := time.Parse("2006-01-02", ct.currentDate)
	stats := make(map[string]int, len(ct.todayClicks))
	for author, users := range ct.todayClicks {
		stats[author] = len(users)
	}

	// Сохраняем агрегированную статистику в БД
	if err := ct.repo.SaveDailyStats(previousDate, stats); err != nil {
		fmt.Printf("ERROR saving daily stats: %v\n", err)
	}

	// Сдвигаем данные
	ct.yesterdayStats = stats
	ct.todayClicks = make(map[string]map[string]struct{})
	ct.currentDate = today
}

// Загружаем всех пользователей за сегодня
func (ct *ClickTracker) loadTodayFromDB() error {
	today := time.Now()
	clicks, err := ct.repo.GetAllClicksForDate(today)
	if err != nil {
		return fmt.Errorf("load today clicks: %w", err)
	}
	for author, users := range clicks {
		userSet := make(map[string]struct{}, len(users))
		for _, u := range users {
			userSet[u] = struct{}{}
		}
		ct.todayClicks[author] = userSet
	}
	return nil
}
