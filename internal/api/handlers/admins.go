package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"quiz_app/internal/models"
	"quiz_app/internal/repository/sqlconnect"
	"quiz_app/pkg/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

	//logging received input details
	fmt.Println("sent data is:-", req)

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

	//connect to db
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

func ValidateCookie(w http.ResponseWriter, r *http.Request) {
	log.Println("Inside /validate cookie handler")
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	cookie, err := r.Cookie("Bearer")	//check cookie exists with the given name
	if err != nil {
		utils.ErrorLogger(err)
		response := models.CheckLogin{
			Valid: false,
			ID: 0,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	//claims object
	claims := &models.Claims{}

	tokenStr := cookie.Value
	fmt.Println("Received Token String:", tokenStr)

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		utils.ErrorLogger(err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	fmt.Println("Printing Claims for testing", claims)
	
	fmt.Println("Decoded Claims is:", claims)

	if claims.ID > 1000 {
		fmt.Println("claims id > 1000")
		response := models.CheckLogin{
			Valid: true,
			ID: claims.ID,
		}
		json.NewEncoder(w).Encode(response)
	}else {
		fmt.Println("claims id < 1000")
		response := models.CheckLogin{
			Valid: false,
			ID: claims.ID,
		}
		json.NewEncoder(w).Encode(response)
	}
}

//logout handler
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inside /logout handler")

	//remove the Value Attribute from the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value: "",
		Path:     "/",
		Domain: "localhost",
		HttpOnly: false, // this allows JavaScript to access the cookie.
		Secure:   true, // this allows the cookie to be sent over non-HTTPS connections.
		Expires:  time.Now().Add(-1 * time.Hour),
		MaxAge: -1,
		SameSite: http.SameSiteLaxMode,
	})

	

	// cookieVal := cookie.Value

	response := struct{
		Success bool
		Message string
		CookieValue string
	}{
		Success: true,
		Message: "Logout Successful",
		CookieValue: "deleted",
	}

	json.NewEncoder(w).Encode(response)
}
