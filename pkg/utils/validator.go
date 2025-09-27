package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"quiz_app/internal/models"
	"reflect"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

func IsAnyUserFieldEmpty(admin models.Admin) bool {
	adminVal := reflect.ValueOf(admin)

	for i := 1; i < adminVal.NumField(); i++ {
		if adminVal.Field(i) == reflect.ValueOf("") {
			return true
		}
	}
	return false
}

func ValidateCookie(r *http.Request) (*models.Admin, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	cookie, err := r.Cookie("Bearer")
	if err != nil {
		return nil, errors.New("cookie not found")
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	fmt.Println("Decode claims:", claims)
	user := &models.Admin{
		ID:       int(claims["id"].(float64)),
		FullName: claims["full_name"].(string),
		Email:    claims["email"].(string),
	}

	fmt.Println("Printing decoded user:", user)

	return user, nil
}

func ValidateCookiePlayers2(tokenString string, conn *websocket.Conn) (*models.DecodePlayer) {
	fmt.Println("Inside Validate Cookie for Players")
	tokenString = strings.TrimPrefix(tokenString, "player_token=")
	fmt.Println("Token Before Decoding;", tokenString)
	tokenByte, err := base64.StdEncoding.DecodeString(tokenString)
	if err != nil {
		errObj := models.SocketError{
			Type:    "error",
			Message: "failed to decode token",
		}
		ErrorSocket(err, conn, &errObj)

	}
	fmt.Println("Token after decoding", string(tokenByte))

	//variable to store decoded player obj 
	playerObj := models.DecodePlayer{}
	err = json.Unmarshal(tokenByte, &playerObj)
	if err != nil {
		errObj := CreateSocketErrorObj("failed to Unmarshal decoded player")
		ErrorSocket(err, conn, errObj)
	}

	return &playerObj

}

func CreatePlayerCookie(playerObj models.PlayerIdName, w http.ResponseWriter) string {
	//convert playerObj to []byte
	playerObjByte, err := json.Marshal(playerObj)
	if err != nil {
		ErrorHttp(err, w, "failed marshalling", http.StatusInternalServerError)
		return ""
	}

	encodedString := base64.StdEncoding.EncodeToString(playerObjByte)
	fmt.Println("Player Cookie Token")
	fmt.Println(encodedString)
	return encodedString
}
