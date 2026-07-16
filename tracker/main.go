package main

import (
	"log"
	"net/http"

	"go.service.clickTracker/tracker/clickTracker"
	"go.service.clickTracker/tracker/handler"
	"go.service.clickTracker/tracker/storage"
	repository "go.service.clickTracker/tracker/storage/config"
)

func main() {
	//  1. Конфиг из .env
	cfg, err := repository.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}

	// 2. Подключение к БД
	if err := repository.InitDB(cfg.DatabaseURL()); err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	defer repository.Pool.Close()

	// 3. Создаём репозиторий для трекера
	repo := storage.NewPostgresClickRepo(repository.Pool)

	// 4. Запускаем трекер (восстанавливает данные из БД)
	tracker, err := clickTracker.NewClickTracker(repo)
	if err != nil {
		log.Fatalf("Ошибка создания ClickTracker: %v", err)
	}

	// 5. HTTP-сервер
	server := handler.NewServer(tracker)
	http.HandleFunc("/clickTracker/click", server.HandlerClick)
	http.HandleFunc("/clickTracker/status", server.HandlerStatus)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
