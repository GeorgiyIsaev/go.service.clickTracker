package main

import "fmt"
import (
	"go.service.clickTracker/tracker/clickTracker"
)

func main() {
	tracker := clickTracker.NewClickTracker()

	//Симуляция кликов 1
	tracker.RecordClick("author_1", "user_1")
	tracker.RecordClick("author_1", "user_2")
	tracker.RecordClick("author_1", "user_3")
	tracker.RecordClick("author_2", "user_1")
	tracker.RecordClick("author_1", "user_1") //дубль

	status := tracker.GetAuthorsStatus([]string{"author_1", "author_2", "author_3"})
	fmt.Println(status)

}
