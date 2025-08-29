package routers

import (
	"net/http"
	"quiz_app/internal/api/handlers"
)

func AdminRouters() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /signup", handlers.SignupHandler)
	mux.HandleFunc("POST /login", handlers.LoginHandler)
	mux.HandleFunc("POST /test", )
	return mux
}
