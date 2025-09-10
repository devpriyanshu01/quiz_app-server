package models

type Quiz struct {
	ID int
	Title string
}

type Question struct {
	QuizID int `json:"quiz_id"`
	QuestionText string `json:"question_text"`
	OptionA string `json:"option_a"` 
	OptionB string `json:"option_b"`
	OptionC string `json:"option_c"`
	OptionD string `json:"option_d"`
	CorrectAnswer string `json:"correct_answer"`
	AdminID int `json:"admin_id"`
}