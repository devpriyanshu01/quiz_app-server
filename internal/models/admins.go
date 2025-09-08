package models

import "github.com/golang-jwt/jwt/v5"

type Admin struct {
	ID int `json:"id"`
	Email string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginCredentials struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type Claims struct {
	ID int `json:"id"`
	Email string `json:"email"`
	FullName string `json:"full_name"`
	jwt.RegisteredClaims
}

type CheckLogin struct {
	Valid bool `json:"valid`
	ID int `json:"id"`
}

// type LogoutResponse struct{

// }