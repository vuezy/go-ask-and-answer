package middlewares

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/vuezy/go-ask-and-answer/internal/utils"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())

	validate.RegisterValidation("alphaspace", func(fl validator.FieldLevel) bool {
		fieldValue := fl.Field().String()
		match, err := regexp.MatchString("^[A-Za-z ]+$", fieldValue)
		if err != nil {
			log.Fatalln("Error compiling regex ^[A-Za-z ]+$.", err)
		}
		return match
	})
}

type payload interface {
	*LoginPayload | *RegisterPayload | *QuestionPayload | *AnswerPayload
}

func performValidation[T payload](next http.Handler, payload T, setErrMsg func(map[string]any, string)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := utils.ParseJSON(w, r.Body, payload); err != nil {
			return
		}

		err := validate.Struct(payload)
		if err != nil {
			response := map[string]any{
				"type": "validation_error",
				"msg":  map[string]any{},
			}
			for _, err := range err.(validator.ValidationErrors) {
				msg := response["msg"].(map[string]any)
				if _, exists := msg[err.Field()]; exists {
					continue
				}
				setErrMsg(msg, err.Field())
			}
			utils.RespondWithJSON(w, 400, response)
			return
		}

		ctx := context.WithValue(r.Context(), utils.VALIDATED_CTX, payload)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ParseIdFromURLParam(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		idInt, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			log.Println("Error parsing id from URL param.", err)
			utils.RespondWith404Error(w)
			return
		}

		ctx := context.WithValue(r.Context(), utils.PARSED_ID_CTX, idInt)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
