package main

import (
	"log"
	"net/http"

	"go.service.clickTracker/tracker/clickTracker"
	"go.service.clickTracker/tracker/handler"
)

func main() {
	tracker := clickTracker.NewClickTracker()
	server := handler.NewServer(tracker)

	http.HandleFunc("/clickTracker/click", server.HandlerClick)
	http.HandleFunc("/clickTracker/status", server.HandlerStatus)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
