package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"quiz_app/internal/repository/sqlconnect"
	"quiz_app/pkg/utils"
)

func CreateNewQuizHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inside Create New Quiz Handler")
	quizTitle := make(map[string]string)
	err := json.NewDecoder(r.Body).Decode(&quizTitle)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to parse body", http.StatusBadRequest)
		return
	}

	fmt.Println("Quiz Title:", quizTitle)

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to connect with db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = utils.ValidateCookie(r)
	if err != nil {
		fmt.Println("Error validating cookie", err)
		http.Error(w, "Error validating cookie", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("END OF CODE........."))
}
