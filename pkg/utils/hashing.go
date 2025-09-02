package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

func HashPassword(password string) (string, error) {
	//generate salt
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		ErrorLogger(err)
		return "", fmt.Errorf("failed to feed random values in salt")
	}

	//hash the password
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	//convert hash to string
	hashBase64 := base64.StdEncoding.EncodeToString(hash)
	saltBase64 := base64.StdEncoding.EncodeToString(salt)

	encodedHash := fmt.Sprintf("%s.%s", saltBase64, hashBase64)

	return encodedHash, nil
}

func VerifyPassword(currUserPassword string, enteredPassword string) bool {
	splitPassword := strings.Split(currUserPassword, ".")

	salt, err := base64.StdEncoding.DecodeString(splitPassword[0])
	if err != nil {
		ErrorLogger(err)
		return false
	}

	hashedPassword, err := base64.StdEncoding.DecodeString(splitPassword[1])
	if err != nil {
		ErrorLogger(err)
		return false
	}

	hashedEnteredPassword := argon2.IDKey([]byte(enteredPassword), []byte(salt), 1, 64*1024, 4, 32)

	fmt.Println("hashPassword:", string(hashedPassword))
	fmt.Println("hashed Entered passowrd:", string(hashedEnteredPassword))

	if len(hashedPassword) != len(hashedEnteredPassword) {
		return false
	}

	if subtle.ConstantTimeCompare([]byte(hashedPassword), []byte(hashedEnteredPassword)) == 1 {
		//do nothing
		return true
	} else {
		return false
	}
}

func SignToken(userId int, email string, fullName string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtExpiresIn := os.Getenv("JWT_EXPIRES_IN")

	claims := jwt.MapClaims{
		"id":       userId,
		"email":     email,
		"full_name" : fullName,
	}

	log.Println("claims printing: ---------", claims)

	if jwtExpiresIn != "" {
		duration, err := time.ParseDuration(jwtExpiresIn)
		if err != nil {
			return "", err
		}
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(duration))
	} else {
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(15 * time.Minute))
	}

	//generate the token using claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//sign the token with jwt secret
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func VerifyJwtToken(jwtToken string) (jwt.MapClaims, error) {
	decodedClaims := jwt.MapClaims{}
	secretKey := os.Getenv("JWT_SECRET")
	token, err := jwt.ParseWithClaims(jwtToken, &decodedClaims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		ErrorLogger(err)
		return decodedClaims, err
	}

	if !token.Valid {
		return decodedClaims, fmt.Errorf("invalid token")
	}

	fmt.Println("Decoded Claims from JWT", decodedClaims)
	return decodedClaims, nil
}

func CheckJwtToken(r *http.Request, w http.ResponseWriter) jwt.MapClaims {
	//extract cookie
	cookie, err := r.Cookie("Bearer")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, "No cookie sent", http.StatusForbidden)
			return jwt.MapClaims{}
		}
		ErrorLogger(err)
		http.Error(w, "Error occured parsing cookie", http.StatusInternalServerError)
		return jwt.MapClaims{}
	}

	jwtToken := cookie.Value
	log.Println("JWT - Token", jwtToken)

	decodedClaims, err := VerifyJwtToken(jwtToken)
	if err != nil {
		http.Error(w, "Invalid JWT Token", http.StatusForbidden)
		return jwt.MapClaims{}
	}
	return decodedClaims
}
