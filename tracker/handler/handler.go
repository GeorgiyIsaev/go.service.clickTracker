package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ClickTrackerService – контракт для трекера кликов
type ClickTrackerService interface {
	RecordClick(authorID, userID string)
	GetAuthorsStatus(authorIDs []string) map[string]int
}

// Структура сервер которая хранит кликер
type Server struct {
	tracker ClickTrackerService
}

// Конструктор сервера
func NewServer(tracker ClickTrackerService) *Server {
	return &Server{tracker: tracker}
}

// Запрос на регистрацию клика
type ClickRequest struct {
	AuthorID string `json:"author_id"`
	UserID   string `json:"user_id"`
}

/*Хендлер регистраиции клика*/
func (s *Server) HandlerClick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ClickRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Request JSON decode error: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.AuthorID == "" || req.UserID == "" {
		http.Error(w, "AutorID or UserID not fund ", http.StatusBadRequest)
		return
	}

	s.tracker.RecordClick(req.AuthorID, req.UserID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))

}

/**Хендлер получения статистики */
func (s *Server) HandlerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) //405
		return
	}

	idsParam := r.URL.Query().Get("author_ids")
	if idsParam == "" {
		http.Error(w, "authorIDs not fund ", http.StatusBadRequest) //400
		return
	}

	authorIDs := strings.Split(idsParam, ",")
	status := s.tracker.GetAuthorsStatus(authorIDs)

	resp, err := json.Marshal(status)
	if err != nil {
		http.Error(w, "error Marshal", http.StatusInternalServerError) //500
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}
