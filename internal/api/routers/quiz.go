package routers

import (
	"net/http"
	"quiz_app/internal/api/handlers"
)

func QuizRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /createquiz", handlers.CreateNewQuizHandler)
	mux.HandleFunc("GET /listmyquizzes", handlers.ListMyQuizzes)
	mux.HandleFunc("POST /savequestion", handlers.SaveOneQuestion)
	mux.HandleFunc("POST /getquestions", handlers.GetQuestions)
	mux.HandleFunc("POST /quiz/delete", handlers.DeleteQuiz)
	mux.HandleFunc("POST /quiz/activate", handlers.ActivateQuiz)
	mux.HandleFunc("GET /quiz/join/{quiz_id}", handlers.StartQuiz)
	mux.HandleFunc("GET /quiz/validate/{quizId}", handlers.ValidateQuiz)
	return mux
}