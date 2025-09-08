package routers

import (
	"net/http"
	"quiz_app/internal/api/handlers"
)

func AdminRouters() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /signup", handlers.SignupHandler)
	mux.HandleFunc("POST /login", handlers.LoginHandler)
	mux.HandleFunc("GET /validatecookie", handlers.ValidateCookie)
	mux.HandleFunc("GET /logout", handlers.LogoutHandler)
	mux.HandleFunc("/test", handlers.TestHandler)
	return mux
}
