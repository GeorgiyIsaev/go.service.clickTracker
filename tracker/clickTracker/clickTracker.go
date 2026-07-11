package clickTracker

import (
	"sync"
	"time"

	"go.service.clickTracker/tracker/clickTracker/storage"
)

// Структура для хранения кликов
type ClickTracker struct {
	mu              sync.RWMutex
	todayClicks     map[string]map[string]struct{} // authorID -> set of userID
	yesterdayClicks map[string]int                 //только количество
	currentDate     string                         // "YYYY-MM-DD"
	repo            storage.ClickRepository
}

// Экзмпляр класс ClickTracker
func NewClickTracker(repo storage.ClickRepository) *ClickTracker {
	now := time.Now()
	return &ClickTracker{
		todayClicks:     make(map[string]map[string]struct{}),
		yesterdayClicks: make(map[string]int),
		currentDate:     now.Format("2006-01-02"),
		repo:            repo,
	}
	//TODO востоновить сосотяние
	//и статистику за завтра

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

	ct.yesterdayClicks = ct.todayClicks
	ct.todayClicks = make(map[string]map[string]struct{})
	ct.currentDate = today
}
