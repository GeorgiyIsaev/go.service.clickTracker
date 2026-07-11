package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ------------- Мок трекера -------------
type mockTracker struct {
	recordedAuthor string
	recordedUser   string
	statusResult   map[string]int
}

func (m *mockTracker) RecordClick(authorID, userID string) {
	m.recordedAuthor = authorID
	m.recordedUser = userID
}

func (m *mockTracker) GetAuthorsStatus(authorIDs []string) map[string]int {
	return m.statusResult
}

// ------------- Тесты HandlerClick -------------

func TestHandlerClick_MethodNotAllowed(t *testing.T) {
	s := NewServer(&mockTracker{})
	req := httptest.NewRequest(http.MethodGet, "/click", nil)
	w := httptest.NewRecorder()

	s.HandlerClick(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandlerClick_InvalidJSON(t *testing.T) {
	s := NewServer(&mockTracker{})
	body := `{"author_id": 123` // невалидный JSON
	req := httptest.NewRequest(http.MethodPost, "/click", strings.NewReader(body))
	w := httptest.NewRecorder()

	s.HandlerClick(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandlerClick_EmptyAuthorID(t *testing.T) {
	s := NewServer(&mockTracker{})
	body := `{"author_id": "", "user_id": "user1"}`
	req := httptest.NewRequest(http.MethodPost, "/click", strings.NewReader(body))
	w := httptest.NewRecorder()

	s.HandlerClick(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	// Проверяем, что текст ошибки содержит ожидаемое сообщение
	if !strings.Contains(w.Body.String(), "AutorID or UserID not fund") {
		t.Errorf("unexpected error message: %s", w.Body.String())
	}
}

func TestHandlerClick_EmptyUserID(t *testing.T) {
	s := NewServer(&mockTracker{})
	body := `{"author_id": "author1", "user_id": ""}`
	req := httptest.NewRequest(http.MethodPost, "/click", strings.NewReader(body))
	w := httptest.NewRecorder()

	s.HandlerClick(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandlerClick_Success(t *testing.T) {
	mock := &mockTracker{}
	s := NewServer(mock)
	body := `{"author_id": "author1", "user_id": "user1"}`
	req := httptest.NewRequest(http.MethodPost, "/click", strings.NewReader(body))
	w := httptest.NewRecorder()

	s.HandlerClick(w, req)

	// Проверяем статус и тело
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	expectedBody := `{"status": "ok"}`
	if strings.TrimSpace(w.Body.String()) != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, w.Body.String())
	}
	// Проверяем, что вызов RecordClick передан правильные параметры
	if mock.recordedAuthor != "author1" || mock.recordedUser != "user1" {
		t.Errorf("expected RecordClick(author1, user1), got RecordClick(%s, %s)",
			mock.recordedAuthor, mock.recordedUser)
	}
}

// ------------- Тесты HandlerStatus -------------

func TestHandlerStatus_MethodNotAllowed(t *testing.T) {
	s := NewServer(&mockTracker{})
	req := httptest.NewRequest(http.MethodPost, "/stats", nil)
	w := httptest.NewRecorder()

	s.HandlerStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandlerStatus_MissingAuthorIDs(t *testing.T) {
	s := NewServer(&mockTracker{})
	req := httptest.NewRequest(http.MethodGet, "/stats", nil) // без параметра
	w := httptest.NewRecorder()

	s.HandlerStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandlerStatus_Success(t *testing.T) {
	mock := &mockTracker{
		statusResult: map[string]int{
			"author1": 10,
			"author2": 0,
		},
	}
	s := NewServer(mock)
	req := httptest.NewRequest(http.MethodGet, "/stats?author_ids=author1,author2", nil)
	w := httptest.NewRecorder()

	s.HandlerStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result map[string]int
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["author1"] != 10 || result["author2"] != 0 {
		t.Errorf("unexpected result: %v", result)
	}
	// Проверяем заголовок Content-Type
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
