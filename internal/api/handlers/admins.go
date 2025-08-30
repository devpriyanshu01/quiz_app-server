package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"quiz_app/internal/models"
	"quiz_app/internal/repository/sqlconnect"
	"quiz_app/pkg/utils"
	"time"
)
//just for testing handler
func TestHandler(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("You're are inside /test route. It means your are testing the application."))
}
// admin signup handler
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inside Signup Handler......")
	var req models.Admin

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to parse body", http.StatusBadRequest)
		return
	}

	//basic validation
	isFieldEmpty := utils.IsAnyUserFieldEmpty(req)
	if isFieldEmpty {
		http.Error(w, "All fields should have a value", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to connect with db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	//hashing password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.ErrorLogger(err)
	}
	fmt.Println("hashPassword:", hashedPassword)

	query := "INSERT INTO admins (email, password, full_name) VALUES(?, ?, ?)"
	result, err := db.Exec(query, req.Email, hashedPassword, req.FullName)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to signup", http.StatusInternalServerError)
		return
	}

	idInserted, err := result.LastInsertId()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "signup successful, but failed to get the id of the created admin", http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("User Created with id - %v", idInserted)

	w.Write([]byte(response))
}

// admins login handler
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inside Login Handler........")
	var req models.LoginCredentials
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to parse body", http.StatusBadRequest)
		return
	}

	//basic validation
	if req.Email == "" || req.Password == "" {
		http.Error(w, "All fields should have a value", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to connect with db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	//password validation
	var currUser models.Admin
	err = db.QueryRow("SELECT id, email, password, full_name FROM admins WHERE email = ?", req.Email).Scan(&currUser.ID, &currUser.Email, &currUser.Password, &currUser.FullName)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to retreive user from the server", http.StatusInternalServerError)
		return
	}

	//verify password
	isCorrect := utils.VerifyPassword(currUser.Password, req.Password)
	if !isCorrect {
		http.Error(w, "incorrect password", http.StatusForbidden)
		return
	} else {
		log.Println("password verified")
	}

	//generate jwt
	jwtTokenString, err := utils.SignToken(currUser.ID, currUser.Email, currUser.FullName)
	if err != nil {
		utils.ErrorLogger(err)
		http.Error(w, "failed to Login", http.StatusInternalServerError)
		return
	}

	//send token as a response or a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    jwtTokenString,
		Path:     "/",
		HttpOnly: false, // this allows JavaScript to access the cookie.
		Secure:   false, // this allows the cookie to be sent over non-HTTPS connections.
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Token string `json:"token"`
	}{
		Token: jwtTokenString,
	}
	json.NewEncoder(w).Encode(response)
}
