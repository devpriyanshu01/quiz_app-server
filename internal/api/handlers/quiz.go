package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"quiz_app/internal/models"
	"quiz_app/internal/repository/sqlconnect"
	"quiz_app/pkg/utils"
	"strings"
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
	query := "SELECT id, title FROM quizzes WHERE admin_id = ?"
	rows, err := db.Query(query, admin.ID)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to retreive quizzes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var quiz models.Quiz
		err := rows.Scan(&quiz.ID, &quiz.Title)
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
