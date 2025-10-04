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

	"github.com/gorilla/websocket"
)

// create global marks store wiz. map of quiz_ids.
var globalQuizStore = make(map[int](map[string]models.PlayerDetails))

// handler to save the player to the database.
func SavePlayers(w http.ResponseWriter, r *http.Request) {
	//extract player data
	var playerData models.PlayerData
	err := json.NewDecoder(r.Body).Decode(&playerData)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "Invalid sent body", http.StatusBadRequest)
		return
	}

	fmt.Println("Received Player Data:", playerData)

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

	playerObj := models.PlayerIdName{
		Id:   id,
		Name: playerData.Name,
	}
	fmt.Println("Encoding:", playerObj)
	token := utils.CreatePlayerCookie(playerObj, w) //create player cookie
	fmt.Println("Encoded PlayerObj:", token)

	//add current player and current quiz to global quiz store
	//for sending leaderboard data
	_, exists := globalQuizStore[playerData.QuizId]	
	if !exists {
		globalQuizStore[playerData.QuizId] = make(map[string]models.PlayerDetails)
		fmt.Println("Added QuizId - ", playerData.QuizId, " to global quiz store")
	}

	//if current quiz already exists in global quiz store, then add
	//new player in the store.
	globalQuizStore[playerData.QuizId][id] = models.PlayerDetails{
		Name: playerData.Name,
		Marks: 0,
	}

	//log this quiz id store to check if initialized or not
	fmt.Println("********************** Initializiing Current Quiz Store ******************")
	fmt.Println(globalQuizStore[playerData.QuizId])

	//send token as a response or a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "player_token",
		Value:    token,
		Path:     "/",
		HttpOnly: false, // this allows JavaScript to access the cookie.
		Secure:   false, // this allows the cookie to be sent over non-HTTPS connections.
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Token   string `json:"token"`
		Success bool   `json:"success"`
	}{
		Token:   token,
		Success: true,
	}
	json.NewEncoder(w).Encode(response)

}

// save answers for answered question for players
/*
func SaveAnswer(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Save ans route Hit")
	var saveans models.SaveAns
	err := json.NewDecoder(r.Body).Decode(&saveans)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to parse body", http.StatusBadRequest)
		return
	}

	//connect db
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to connect with db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	//validate the cookie & decode details
	decodedPlayerDetails, err := utils.ValidateCookiePlayers2()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to validate player cookie", http.StatusUnauthorized)
		return
	}

	//query string
	query := "UPDATE players SET question_id = ?, choosen_ans = ?, marks = ? WHERE id = ?"
	result, err := db.Exec(query, saveans.QuestionId, saveans.ChoosenAns, saveans.Marks, decodedPlayerDetails.Id)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to save question", http.StatusInternalServerError)
		return
	}

	//rows affected count
	rowsAff, err := result.RowsAffected()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to get rows affected", http.StatusInternalServerError)
		return
	}
	fmt.Println("Rows Affected after saving question:", rowsAff)

}
*/

// get leaderboard
func GetLeaderboard(conn *websocket.Conn, leaderboardBody *models.GetLeaderBoardBody) {
	//get player id from cookie
	decodedPlayer := utils.ValidateCookiePlayers2(leaderboardBody.Cookie, conn)

	//connect db
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		errObj := models.SocketError{
			Type:    "error",
			Message: "failed to connect db",
		}
		utils.ErrorSocket(err, conn, &errObj)
		return
	}
	defer db.Close()

	//leaderboard store
	// leaderboardStore := []models.LeaderData{}

	//query string
	query := "SELECT marks FROM answers WHERE player_id = ? and question_id = ?"
	row := db.QueryRow(query, decodedPlayer.Id, leaderboardBody.QuestionId)

	//variable for storing fetched marks
	marksData := models.MarksData{
		Role:       "leaderboard",
		PlayerId:   decodedPlayer.Id,
		PlayerName: decodedPlayer.Name,
		Marks:      0,
	}
	err = row.Scan(&marksData.Marks)
	if err != nil { //handle error
		errObj := models.SocketError{
			Type:    "error",
			Message: "failed to scan value is marks varialbe",
		}
		utils.ErrorSocket(err, conn, &errObj)
		return
	}

	//

	//convert marksData to slice of byte for sending using websocket
	marksDataBytes, err := json.Marshal(marksData)
	if err != nil {
		errObj := models.SocketError{
			Type:    "error",
			Message: "failed to marshal marks data",
		}
		utils.ErrorSocket(err, conn, &errObj)
	}
	//send marks data to client
	conn.WriteMessage(websocket.TextMessage, marksDataBytes)

}
