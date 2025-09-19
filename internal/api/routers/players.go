package routers

import (
	"net/http"
	"quiz_app/internal/api/handlers"
)

func PlayersRouters() *http.ServeMux{
	mux := http.NewServeMux()
	mux.HandleFunc("POST /players/save", handlers.SavePlayers)
	return mux
}