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
	"strings"
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for testing; tighten this in production
	},
}

func StartQuiz(w http.ResponseWriter, r *http.Request) {
	// var quizId models.JoinQuiz
	// err := json.NewDecoder(r.Body).Decode(&quizId)
	// if err != nil {
	// 	utils.ErrorLogger(err)
	// 	http.Error(w, "failed to start quiz as quiz_id not provided", http.StatusBadRequest)
	// 	return
	// }
	quizId := r.PathValue("quiz_id")
	fmt.Println("quizId is:", quizId)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println("Message Received:", string(message))

		time.Sleep(3 * time.Second)

		msgToClient := "I'm Backend Webscoket Server for " + quizId
		//write message back to the client
		if err := conn.WriteMessage(messageType, []byte(msgToClient)); err != nil {
			log.Println(err)
			return
		}
	}
}

func ValidateQuiz(w http.ResponseWriter, r *http.Request){
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
