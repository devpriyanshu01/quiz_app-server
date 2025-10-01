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

// for fetching leaderboard data
type GetLeaderBoardBody struct {
	Cookie string `json:"cookie"`
	QuestionId int `json:"question_id"`
	QuizId int `json:"quiz_id"`
}

// send marks for each player
type MarksData struct {
	Role string `json:"role"`
	PlayerId string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Marks int `json:"marks"`
}

//player Data for sending marks to client
type PlayerDetails struct {
	Name string `json:"name"`
	Marks int `json:"marks"`
}

type LeaderBoard struct {
	Type string `json:"type"`
	Data map[string]PlayerDetails
}


