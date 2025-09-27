package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"quiz_app/internal/models"

	"github.com/gorilla/websocket"
)

func ErrorLogger(err error) {
	log.Println(err)
}

// error handler function for http requests
func ErrorHttp(err error, w http.ResponseWriter, msg string, statusCode int) {
	log.Println(err)
	http.Error(w, msg, statusCode)
}

// error handler function for websocket connection
func ErrorSocket(err error, conn *websocket.Conn, errObj *models.SocketError) {
	//Marshall the errObj before sending
	fmt.Println(err)
	errObjBytes, err := json.Marshal(errObj)
	if err != nil {
		log.Println(err)
	}
	conn.WriteMessage(websocket.TextMessage, errObjBytes)
}

func CreateSocketErrorObj(msg string) *models.SocketError {
	errObj := models.SocketError{
		Type:    "error",
		Message: msg,
	}
	return &errObj
}
