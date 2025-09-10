package routers

import (
	"net/http"
	"quiz_app/internal/api/handlers"
)

func QuizRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /createquiz", handlers.CreateNewQuizHandler)
	mux.HandleFunc("GET /listmyquizzes", handlers.ListMyQuizzes)
	mux.HandleFunc("GET /savequestion", handlers.SaveOneQuestion)
	return mux
}