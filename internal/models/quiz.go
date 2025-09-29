package models

import "github.com/gorilla/websocket"

type Quiz struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Active bool   `json:"active"`
}

type Question struct {
	QuizID        int    `json:"quiz_id"`
	QuestionText  string `json:"question_text"`
	OptionA       string `json:"option_a"`
	OptionB       string `json:"option_b"`
	OptionC       string `json:"option_c"`
	OptionD       string `json:"option_d"`
	CorrectAnswer string `json:"correct_answer"`
	PointsCorrect int    `json:"points_correct"`
	AdminID       int    `json:"admin_id"`
}

type FetchQuestions struct {
	ID int `json:"id"`
	Question
}
type GetQuizId struct {
	QuizID int `json:"quiz_id"`
}

type DeleteQuiz struct {
	QuizId int `json:"quiz_id"`
}

type ActivateQuiz struct {
	ID int `json:"id"`
}

type JoinQuiz struct {
	QuizId int `json:"quiz_id"`
}

type ConnectedClients struct {
	Conn *websocket.Conn
	QuizId string 
	Name string
}

type QuizHub struct {
	QuizId string
	Clients map[*websocket.Conn]bool
	Broadcast chan []byte 
	Register chan *websocket.Conn
	Unregister chan *websocket.Conn
}

//Hub Logic
func (h *QuizHub) Run(){
	for {
		select {
		case conn := <- h.Register :
			h.Clients[conn] = true
		case conn := <- h.Unregister:
			delete(h.Clients, conn)
			conn.Close()
		case ques := <- h.Broadcast:
			for conn := range h.Clients {
				conn.WriteMessage(websocket.TextMessage, ques)
			}
		}
	}
}

//testing...
type TestStruct struct {
	Name string `json:"name"`
}
