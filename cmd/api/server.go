package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"quiz_app/internal/api/middlewares"
	"quiz_app/internal/api/routers"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("failed to load .env file contents")
		return
	}

	serverPort := os.Getenv("SERVER_PORT")

	mux := routers.MainRouters()
	secureMux := middlewares.Cors(mux)
	//create a custom server
	server := &http.Server{
		Addr:    serverPort,
		Handler: secureMux,
	}
	fmt.Println("Server is Listening on port ", serverPort)
	err = server.ListenAndServe()
	if err != nil {
		log.Println("failed to start the server")
		return
	}
}
