package clickTracker

import (
	"testing"
	"time"
)

func TestNewClickTracker(t *testing.T) {
	ct := NewClickTracker()
	if ct == nil {
		t.Fatal("NewClickTracker вернул nil")
	}
	if ct.todayClicks == nil {
		t.Error("todayClicks не инициализирован")
	}
	if ct.yesterdayClicks == nil {
		t.Error("yesterdayClicks не инициализирован")
	}
	if ct.currentDate == "" {
		t.Error("currentDate не установлен")
	}
	// Проверяем, что текущая дата соответствует сегодняшнему дню
	expected := time.Now().Format("2006-01-02")
	if ct.currentDate != expected {
		t.Errorf("currentDate = %s, ожидается %s", ct.currentDate, expected)
	}
}

func TestRecordClickAndGetStatus(t *testing.T) {
	ct := NewClickTracker()

	// Записываем клики
	ct.RecordClick("author_1", "user_1")
	ct.RecordClick("author_1", "user_2")
	ct.RecordClick("author_1", "user_3")
	ct.RecordClick("author_2", "user_1")
	ct.RecordClick("author_1", "user_1") // дубликат

	// Получаем статус
	authors := []string{"author_1", "author_2", "author_3"}
	status := ct.GetAuthorsStatus(authors)

	expected := map[string]int{
		"author_1": 3, // user_1, user_2, user_3 (уникальные)
		"author_2": 1,
		"author_3": 0,
	}

	for _, author := range authors {
		if status[author] != expected[author] {
			t.Errorf("для автора %s ожидается %d, получено %d", author, expected[author], status[author])
		}
	}
}

func TestDataShift(t *testing.T) {
	ct := NewClickTracker()

	// Записываем клики в "сегодня" (дата в ct.currentDate)
	ct.RecordClick("author_1", "user_1")
	ct.RecordClick("author_1", "user_2")
	ct.RecordClick("author_2", "user_3")

	// Искусственно меняем текущую дату на вчерашнюю
	// (используем доступ к неэкспортируемому полю, т.к. мы в том же пакете)
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	ct.currentDate = yesterday

	// Теперь записываем новый клик – должен произойти сдвиг
	ct.RecordClick("author_1", "user_4")
	ct.RecordClick("author_2", "user_5")

	// Проверяем, что старые клики переместились в yesterdayClicks
	if len(ct.yesterdayClicks) != 2 {
		t.Errorf("yesterdayClicks должно содержать 2 авторов, получено %d", len(ct.yesterdayClicks))
	}
	if len(ct.yesterdayClicks["author_1"]) != 2 { // user_1, user_2
		t.Errorf("yesterdayClicks[author_1] должно быть 2 элемента, получено %d", len(ct.yesterdayClicks["author_1"]))
	}
	if len(ct.yesterdayClicks["author_2"]) != 1 { // user_3
		t.Errorf("yesterdayClicks[author_2] должно быть 1 элемент, получено %d", len(ct.yesterdayClicks["author_2"]))
	}

	// Проверяем, что todayClicks содержит только новые клики
	if len(ct.todayClicks) != 2 {
		t.Errorf("todayClicks должно содержать 2 авторов, получено %d", len(ct.todayClicks))
	}
	if len(ct.todayClicks["author_1"]) != 1 { // user_4
		t.Errorf("todayClicks[author_1] должно быть 1 элемент, получено %d", len(ct.todayClicks["author_1"]))
	}
	if len(ct.todayClicks["author_2"]) != 1 { // user_5
		t.Errorf("todayClicks[author_2] должно быть 1 элемент, получено %d", len(ct.todayClicks["author_2"]))
	}

	// Проверяем, что currentDate обновилась на сегодня (после RecordClick)
	expectedToday := time.Now().Format("2006-01-02")
	if ct.currentDate != expectedToday {
		t.Errorf("after shift currentDate = %s, ожидается %s", ct.currentDate, expectedToday)
	}

	// Проверяем, что GetAuthorsStatus возвращает только сегодняшние данные
	status := ct.GetAuthorsStatus([]string{"author_1", "author_2", "author_3"})
	if status["author_1"] != 1 {
		t.Errorf("GetAuthorsStatus для author_1 ожидается 1, получено %d", status["author_1"])
	}
	if status["author_2"] != 1 {
		t.Errorf("GetAuthorsStatus для author_2 ожидается 1, получено %d", status["author_2"])
	}
	if status["author_3"] != 0 {
		t.Errorf("GetAuthorsStatus для author_3 ожидается 0, получено %d", status["author_3"])
	}
}

// Тест на конкурентную запись (для проверки блокировок)
func TestConcurrentAccess(t *testing.T) {
	ct := NewClickTracker()
	const goroutines = 100
	const iterations = 100

	done := make(chan bool)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				author := "author"
				user := "user"
				ct.RecordClick(author, user)
			}
			done <- true
		}(i)
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// После всех записей должно быть ровно 1 уникальный пользователь для этого автора
	status := ct.GetAuthorsStatus([]string{"author"})
	if status["author"] != 1 {
		t.Errorf("ожидается 1 уникальный пользователь, получено %d", status["author"])
	}
}
