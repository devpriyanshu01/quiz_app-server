package utils

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"quiz_app/internal/models"
	"reflect"

	"github.com/golang-jwt/jwt/v5"
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
