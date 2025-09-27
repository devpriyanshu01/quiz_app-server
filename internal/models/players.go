package models

import (
	"time"
)

type PlayerData struct {
	Name      string    `json:"name"`
	QuizId    int       `json:"quiz_id"`
	TimeStamp time.Time `json:"time-stamp"`
}

type SaveAns struct {
	Token      string `json:"token"`
	QuestionId int    `json:"question_id"`
	ChoosenAns string `json:"choosen_ans"`
	Marks      int    `json:"marks"`
	QuizId     int    `json:"quiz_id"`
}

type DecodePlayer struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type SocketError struct {
	Type    string `json:"type"`
	Message string `json:"error"`
}

// player details encoding
type PlayerIdName struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
