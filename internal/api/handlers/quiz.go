package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"quiz_app/internal/models"
	"quiz_app/internal/repository/sqlconnect"
	"quiz_app/pkg/utils"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
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

	admin, err := utils.ValidateCookie(r)
	if err != nil {
		fmt.Println("Error validating cookie", err)
		http.Error(w, "Error validating cookie", http.StatusUnauthorized)
		return
	}

	fmt.Println("printing decoded admin:", admin.ID)

	query := "INSERT INTO quizzes(title, admin_id) VALUES(?, ?)"
	res, err := db.Exec(query, quizTitle["title"], admin.ID)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "Error creating quiz", http.StatusInternalServerError)
		return
	}
	insertedId, err := res.LastInsertId()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "Failed to retreive inserted id of the quiz", http.StatusInternalServerError)
		return
	}

	response := struct {
		QuizID int
	}{
		QuizID: int(insertedId),
	}
	json.NewEncoder(w).Encode(response)
}

func ListMyQuizzes(w http.ResponseWriter, r *http.Request) {

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to connect with db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	//validate the cookie/user
	admin, err := utils.ValidateCookie(r)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to validate cookie/user", http.StatusBadRequest)
		return
	}

	//get the all quizzes title
	var quizTitles []models.Quiz
	query := "SELECT id, title, active  FROM quizzes WHERE admin_id = ?"
	rows, err := db.Query(query, admin.ID)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to retreive quizzes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var quiz models.Quiz
		err := rows.Scan(&quiz.ID, &quiz.Title, &quiz.Active)
		if err != nil {
			utils.ErrorLogger(err)
			http.Error(w, "failed to scan quiz title", http.StatusInternalServerError)
			return
		}
		quizTitles = append(quizTitles, quiz)
	}

	json.NewEncoder(w).Encode(quizTitles)
}

func SaveOneQuestion(w http.ResponseWriter, r *http.Request) {
	//parse body
	var question models.Question
	err := json.NewDecoder(r.Body).Decode(&question)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to parse the question", http.StatusBadRequest)
		return
	}

	//make correct_answer to lower case
	question.CorrectAnswer = strings.ToLower(question.CorrectAnswer)

	//validate the admin/user
	admin, err := utils.ValidateCookie(r)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to validate cookie/user", http.StatusUnauthorized)
		return
	}

	//assign extracted id from cookie to question admin_id column
	question.AdminID = admin.ID
	question.PointsCorrect = 100 //default value points_correct

	//connect db
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	//save question response type
	type qResponse struct {
		IsSaved bool
	}
	response := qResponse{
		IsSaved: false,
	}

	//create the query for db insertion
	query := "INSERT INTO questions(quiz_id, question_text, option_a, option_b, option_c, option_d, correct_answer, admin_id) VALUES(?, ?, ?, ?, ?, ?, ?, ?)"
	result, err := db.Exec(query, question.QuizID, question.QuestionText, question.OptionA, question.OptionB, question.OptionC, question.OptionD, question.CorrectAnswer, question.AdminID)
	if err != nil {
		utils.ErrorLogger(err)
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.ErrorLogger(err)
		}
		http.Error(w, "failed to save question to db", http.StatusInternalServerError)
		return
	}

	questionID, err := result.LastInsertId()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to get the id of added question", http.StatusInternalServerError)
		return
	}

	fmt.Println("Saved Question ID", questionID)

	//send successful response
	response.IsSaved = true
	json.NewEncoder(w).Encode(response)

}

// get saved questions for a specific quiz_id
func GetQuestions(w http.ResponseWriter, r *http.Request) {
	//get the request sent
	var req models.GetQuizId
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "Error parsing the request", http.StatusBadRequest)
		return
	}

	//validate the user/cookie
	_, err = utils.ValidateCookie(r)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to validate the cookie/user", http.StatusUnauthorized)
		return
	}

	//connect db
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	//storage for the fetched questions
	var fetchedQuestions []models.Question

	//generate the query string
	query := "SELECT question_text, option_a, option_b, option_c, option_d, correct_answer, points_correct FROM questions WHERE quiz_id = ?"

	//query db
	rows, err := db.Query(query, req.QuizID)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "error while querying database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	//scan the values from rows fetch from db
	for rows.Next() {
		var q models.Question //instance of question model to save one question
		err := rows.Scan(&q.QuestionText, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD, &q.CorrectAnswer, &q.PointsCorrect)
		if err != nil {
			utils.ErrorLogger(err)
			http.Error(w, "failed to insert/scan values to output variable", http.StatusInternalServerError)
			return
		}
		fetchedQuestions = append(fetchedQuestions, q)
	}

	//send response
	type GetQuestionResponse struct {
		Success bool              `json:"success"`
		Payload []models.Question `json:"payload"`
	}
	response := GetQuestionResponse{
		Success: true,
		Payload: fetchedQuestions,
	}

	json.NewEncoder(w).Encode(response)
}

func DeleteQuiz(w http.ResponseWriter, r *http.Request) {
	var quizId models.DeleteQuiz
	err := json.NewDecoder(r.Body).Decode(&quizId)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "Invalid Quiz_ID", http.StatusBadRequest)
		return
	}

	//validate the cookie/user
	admin, err := utils.ValidateCookie(r)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to validate user/cookie", http.StatusUnauthorized)
		return
	}
	fmt.Println("Decoded Admin", admin)

	//connect database
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to connect to db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	//delete query
	query := "DELETE FROM quizzes WHERE id = ?"
	_, err = db.Exec(query, quizId.QuizId)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to delete the quiz", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("quiz deleted"))
}

func ActivateQuiz(w http.ResponseWriter, r *http.Request) {
	fmt.Println("inside activate quiz")
	var quizId models.ActivateQuiz
	err := json.NewDecoder(r.Body).Decode(&quizId)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "Invalid Quiz_ID", http.StatusBadRequest)
		return
	}

	fmt.Println("quizId", quizId)

	//validate the cookie/user
	admin, err := utils.ValidateCookie(r)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to validate user/cookie", http.StatusUnauthorized)
		return
	}
	fmt.Println("Decoded Admin", admin)

	//connect database
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to connect to db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	//activate query
	query := "UPDATE quizzes SET active = ? WHERE id = ?"
	_, err = db.Exec(query, true, quizId.ID)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to activate the quiz", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("quiz activated"))
}

// upgrader for upgrading to websocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for testing; tighten this in production
	},
}

func StartQuiz(w http.ResponseWriter, r *http.Request) {
	quizId := r.PathValue("quiz_id")
	fmt.Println("quizId is:", quizId)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	//connect db
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to connect to db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println("Message Received:", string(message))

		//save client to an struct
		var connectedClients []models.ConnectedClients
		oneClient := models.ConnectedClients{
			Conn:   conn,
			QuizId: quizId,
			Name:   string(message),
		}

		connectedClients = append(connectedClients, oneClient)
		fmt.Println("All Connected Client are ------------------")
		fmt.Println(connectedClients)

		//for quiz to start
		if string(message) == "begin quiz" {
			fmt.Println("condition to start quiz met.")
			query := "SELECT id, question_text, option_a, option_b, option_c, option_d, correct_answer, points_correct FROM questions where quiz_id = ?"
			rows, err := db.Query(query, quizId)
			if err != nil {
				utils.ErrorLogger(err)
				http.Error(w, "failed to fetch questions", http.StatusInternalServerError)
				return
			}
			//for storing questions
			var questions []models.FetchQuestions
			for rows.Next() {
				var question models.FetchQuestions
				err = rows.Scan(&question.ID, &question.QuestionText, &question.OptionA, &question.OptionB, &question.OptionC, &question.OptionD, &question.CorrectAnswer, &question.PointsCorrect)
				if err != nil {
					utils.ErrorLogger(err)
					http.Error(w, "failed to store questions", http.StatusInternalServerError)
					return
				}
				questions = append(questions, question)
			}
			//send message to client every 20s
			for _, ques := range questions {
				byteQues, err := json.Marshal(ques)
				if err != nil {
					utils.ErrorLogger(err)
					http.Error(w, "error marshalling question", http.StatusInternalServerError)
					return
				}
				err = conn.WriteMessage(messageType, byteQues)
				if err != nil {
					utils.ErrorLogger(err)
					http.Error(w, "failed to send questions", http.StatusInternalServerError)
					return
				}

				time.Sleep(15 * time.Second)
			}
		}

		msgToClient := 1
		//write message back to the client
		if err := conn.WriteMessage(messageType, []byte(strconv.Itoa(msgToClient))); err != nil {
			log.Println(err)
			return
		}
	}
}

func ValidateQuiz(w http.ResponseWriter, r *http.Request) {
	quizId := r.PathValue("quizId")

	//validate cookie for player later

	//connect db
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to connect to db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	//query
	query := "SELECT title FROM quizzes where id = ?"
	row := db.QueryRow(query, quizId)

	//variable for storing fetched title
	var title string
	err = row.Scan(&title)
	if err == sql.ErrNoRows {
		utils.ErrorLogger(err)
		http.Error(w, "Invalid Quiz Id", http.StatusUnauthorized)
		return
	}
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "Invalid Quiz Id", http.StatusUnauthorized)
		return
	}

	response := struct {
		IsValid bool `json:"is_valid"`
	}{
		IsValid: true,
	}

	json.NewEncoder(w).Encode(response)
}

//declare a variable for storing the marks of each player
var playersMarks = make(map[string]models.PlayerDetails)
var playersInQuiz = make(map[int][]string)

// Broadcasting Logic Begins here
func BroadcastQuestions(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Broadcast Route Hits.....")
	quizId := r.PathValue("quizId")
	fmt.Println("Quiz ID:", quizId)
	conn, err := upgrader.Upgrade(w, r, nil) //upgrade to websocket
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to establish websocket connection", http.StatusInternalServerError)
		return
	}

	hub := getOrCreateHub(quizId)
	hub.Register <- conn


	questions := []models.FetchQuestions{}
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			hub.Unregister <- conn
			break
		}
		//check condition to start sending question
		if string(msg) == "start quiz" {
			fmt.Println("Condition to start Quiz met....")
			questions = fetchQuestions(quizId, w)
			fmt.Println(questions)
			fmt.Println("len(questions):", len(questions))
		}

		fmt.Println("len(questions):", len(questions))
		//print question
		for _, q := range questions {
			fmt.Println("----------------------------------------------------------------------------")
			fmt.Println(q)
		}

		//send each question after receiving trigger from fe
		if string(msg) == "next ques" {
			ticker := time.NewTicker(45 * time.Second)
			defer ticker.Stop()
			fmt.Println("----- Demanded next ques -----")

			for _, currQues := range questions {
				quesData := models.QuesData{
					Type: "question",
					FetchQuestions: currQues,
				}
				quesInByte, err := json.Marshal(quesData)
				if err != nil {
					utils.ErrorLogger(err)
					http.Error(w, "failed to marshal question", http.StatusInternalServerError)
					return
				}
				<-ticker.C
				hub.Broadcast <- quesInByte
			}
		}

		
		//handle save answer
		fmt.Println("Received Msg:", string(msg))
		var saveans = models.SaveAns{}
		fmt.Println("Before Unmarshalling.....")
		fmt.Println(string(msg))
		msgString := string(msg)
		if msgString[0] == '{' {
			fmt.Println("Inside the condn - {")
			err = json.Unmarshal(msg, &saveans)
			if err != nil {
				utils.ErrorLogger(err)
				return
			}
			SaveAnsToDb(&saveans, conn)
			//update global quiz store
		}

		//send the leaderboard details to client when asked for it.
		if string(msg) == "send leaderboard" {
			fmt.Println("............Condition matched for sending leaderboard")
			sendLeaderboardData(conn, quizId)
		}
	}
}

// Creating Global Repository for managing Hubs
var quizHubs = make(map[string]*models.QuizHub)
var mu sync.Mutex

// Initialize a hub for each request
func getOrCreateHub(quizId string) *models.QuizHub {
	mu.Lock()
	defer mu.Unlock()

	if hub, exists := quizHubs[quizId]; exists {
		return hub
	}

	hub := &models.QuizHub{
		QuizId:     quizId,
		Clients:    make(map[*websocket.Conn]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}

	quizHubs[quizId] = hub
	go hub.Run()
	return hub
}

// function for fetching all questions for a quizId
func fetchQuestions(quizId string, w http.ResponseWriter) []models.FetchQuestions {
	fmt.Println("fetchquestions hit...................")
	//connect database
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to connect with db", http.StatusInternalServerError)
		return nil
	}
	defer db.Close()

	query := "SELECT id, question_text, option_a, option_b, option_c, option_d, correct_answer, points_correct FROM questions where quiz_id = ?"
	rows, err := db.Query(query, quizId)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to fetch questions", http.StatusInternalServerError)
		return nil
	}
	//for storing questions
	var questions []models.FetchQuestions
	for rows.Next() {
		var question models.FetchQuestions
		err = rows.Scan(&question.ID, &question.QuestionText, &question.OptionA, &question.OptionB, &question.OptionC, &question.OptionD, &question.CorrectAnswer, &question.PointsCorrect)
		if err != nil {
			utils.ErrorLogger(err)
			http.Error(w, "failed to store questions", http.StatusInternalServerError)
			return nil
		}
		questions = append(questions, question)
	}
	return questions
}

// function to save the answer to database
func SaveAnsToDb(saveans *models.SaveAns, conn *websocket.Conn) {
	//validate cookie
	decodedPlayerDetails := utils.ValidateCookiePlayers2(saveans.Token, conn)
	if decodedPlayerDetails == nil {
		return
	}	

	//print decoded player from the cookie
	fmt.Println("Decoded Player Details:", decodedPlayerDetails)

	//current logic for maintaing a store
	//call the fn to update the global store for current quiz and current player
	updateGlobalQuizStore(*saveans, decodedPlayerDetails)

	//connect db
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		errObj := utils.CreateSocketErrorObj("failed to connect db")
		utils.ErrorSocket(err, conn, errObj)
		return
	}
	defer db.Close()

	//query string
	query := "INSERT INTO answers(player_id, question_id, quiz_id, choosen_ans, marks) VALUES(?, ?, ?, ?, ?)"
	result, err := db.Exec(query, decodedPlayerDetails.Id, saveans.QuestionId, saveans.QuizId, saveans.ChoosenAns, saveans.Marks)
	if err != nil {
		errObj := utils.CreateSocketErrorObj("Player Answer Save Failed")
		utils.ErrorSocket(err, conn, errObj)
		return
	}
	
	//check how many rows udpated
	rowsCount, err := result.RowsAffected()
	if err != nil {
		errObj := utils.CreateSocketErrorObj("failed to get number of updated player")
		utils.ErrorSocket(err, conn, errObj)
		return
	}

	fmt.Println(rowsCount, " row update for player ", decodedPlayerDetails.Name)
}

func sendLeaderboardData(conn *websocket.Conn, quizId string){
	
	fmt.Println("@@@@@@@@@@@@@@@@@@@@@ inside send leaderboard data @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	quizid, _  := strconv.Atoi(quizId)
	leaderBoardData, exists := globalQuizStore[quizid]
	if !exists {
		log.Println("======= LEADER-BOARD DATA DOESN'T EXISTS ========")
		fmt.Println("PRINTING TOTAL GLOBAL QUIZ STORE")
		fmt.Println(globalQuizStore)
		fmt.Println("QUIZ ID WAS:", quizId)
		return
	}
	//leader-board body
	leaderBoardResponse := models.LeaderBoard{
		Type: "leaderboard",
		Data: leaderBoardData,
	}
	leaderBoardDataResponseByte, err := json.Marshal(leaderBoardResponse)
	if err != nil {
		errObj := utils.CreateSocketErrorObj("failed marshal leaderboard data")	
		utils.ErrorSocket(err, conn, errObj)
	}
	//send leaderboard data to frontend using websocket.

	conn.WriteMessage(websocket.TextMessage, leaderBoardDataResponseByte)
}

//fn to update global quiz store
func updateGlobalQuizStore(quesData models.SaveAns, decodedPlayer *models.DecodePlayer){
	fmt.Println("###################################### INSIDE CURRENT PLAYER DATA ############################")
	quizId := quesData.QuizId
	
	currQuizData, exist := globalQuizStore[quizId]
	if !exist {
		log.Println("QuizId ", quizId, " doesn't exist in the global Store")
		return
	}

	currPlayerData, exists := currQuizData[decodedPlayer.Id]
	if !exists {
		log.Println("Player with id ", decodedPlayer.Id, " doesn't exists in quiz data- ", quizId)
		return
	}
	
	currPlayerData.Marks = currPlayerData.Marks + quesData.Marks

	//assign above updated data to global quiz store
	globalQuizStore[quizId][decodedPlayer.Id] = currPlayerData

	//log the updated quiz details
	fmt.Println("************************************** LOGGING UPDATED GLOBAL STORE FOR CURRENT QUIZ ID ****************************")
	fmt.Println(globalQuizStore[quizId])
}