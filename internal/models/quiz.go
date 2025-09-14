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
	PointsCorrect int `json:"points_correct"`
	AdminID int `json:"admin_id"`
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