package middlewares

import (
	"net/http"
)

type LoginPayload struct {
	Email    string `json:"email" validate:"required,email,max=50"`
	Password string `json:"password" validate:"required,alphanum,min=8,max=32"`
}

type RegisterPayload struct {
	Name     string `json:"name" validate:"required,alphaspace,min=3,max=20"`
	Email    string `json:"email" validate:"required,email,max=50"`
	Password string `json:"password" validate:"required,alphanum,min=8,max=32"`
}

func ValidateLoginPayload(next http.Handler) http.Handler {
	return performValidation(next, &LoginPayload{}, func(msg map[string]any, field string) {
		if field == "Email" {
			msg["email"] = "Email must be valid (max length: 50)"
		} else if field == "Password" {
			msg["password"] = "Password must contain 8-32 alphanumeric characters"
		}
	})
}

func ValidateRegisterPayload(next http.Handler) http.Handler {
	return performValidation(next, &RegisterPayload{}, func(msg map[string]any, field string) {
		if field == "Name" {
			msg["name"] = "Name must contain 3-20 alphabet characters"
		} else if field == "Email" {
			msg["email"] = "Email must be valid (max length: 50)"
		} else if field == "Password" {
			msg["password"] = "Password must contain 8-32 alphanumeric characters"
		}
	})
}
