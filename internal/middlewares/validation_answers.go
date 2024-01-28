package middlewares

import "net/http"

type AnswerPayload struct {
	Body       string `json:"body" validate:"required,min=1,max=300"`
	QuestionID int32  `json:"question_id" validate:"omitnil,numeric,gte=0"`
}

func ValidateAnswerPayload(next http.Handler) http.Handler {
	return performValidation(next, &AnswerPayload{}, func(msg map[string]any, field string) {
		if field == "Body" {
			msg["body"] = "Body is required (max length: 300)"
		} else if field == "QuestionID" {
			msg["question_id"] = "Question ID must be provided"
		}
	})
}
