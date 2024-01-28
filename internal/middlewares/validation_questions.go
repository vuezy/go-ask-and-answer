package middlewares

import (
	"net/http"
)

type QuestionPayload struct {
	Title         string `json:"title" validate:"required,min=1,max=50"`
	Body          string `json:"body" validate:"required,min=1,max=300"`
	PriorityLevel int32  `json:"priority_level" validate:"omitnil,numeric,gte=0"`
}

func ValidateQuestionPayload(next http.Handler) http.Handler {
	return performValidation(next, &QuestionPayload{}, func(msg map[string]any, field string) {
		if field == "Title" {
			msg["title"] = "Title is required (max length: 50)"
		} else if field == "Body" {
			msg["body"] = "Body is required (max length: 300)"
		} else if field == "PriorityLevel" {
			msg["priority_level"] = "Priority level must be a number not greater than your credits"
		}
	})
}
