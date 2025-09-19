package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"quiz_app/internal/models"
	"quiz_app/internal/repository/sqlconnect"
	"quiz_app/pkg/utils"
	"strconv"
	"time"
)

func SavePlayers(w http.ResponseWriter, r *http.Request) {
	var playerData models.PlayerData
	err := json.NewDecoder(r.Body).Decode(&playerData)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "Invalid sent body", http.StatusBadRequest)
		return
	}

	//set indian location
	ist, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to load location", http.StatusInternalServerError)
		return
	}

	//get timestamp in IST
	timeStamp := time.Now().In(ist)
	//convert timestamp to string
	timeStampString := timeStamp.Format("2006-01-02 15:04:05")

	id := playerData.Name + strconv.Itoa(playerData.QuizId) + timeStampString
	fmt.Println("id:", id)

	//connect db
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to connect to db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	//query
	query := "INSERT INTO players(id, quiz_id, name) VALUES(?, ?, ?)"
	result, err := db.Exec(query, id, playerData.QuizId, playerData.Name)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to save player", http.StatusInternalServerError)
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to get rows affected", http.StatusInternalServerError)
		return
	}
	fmt.Println("rows:", rows)

	//get token for sending cookie
	token, err := utils.SignTokenPlayers(id, playerData.Name)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to create cookie", http.StatusInternalServerError)
		return
	}

	fmt.Println("token:", token)

	//send token as a response or a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer Player",
		Value:    token,
		Path:     "/",
		HttpOnly: false, // this allows JavaScript to access the cookie.
		Secure:   true, // this allows the cookie to be sent over non-HTTPS connections.
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}
	json.NewEncoder(w).Encode(response)

}
