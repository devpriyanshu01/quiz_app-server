package models

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
