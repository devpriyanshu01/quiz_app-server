package models

import (
	"time"

)

type PlayerData struct {
	Name string `json:"name"`
	QuizId int `json:"quiz_id"`
	TimeStamp time.Time `json:"time-stamp"`
}