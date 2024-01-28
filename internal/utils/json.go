package utils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func RespondWithJSON(w http.ResponseWriter, code int, body map[string]any) {
	data, err := json.Marshal(body)

	w.Header().Add("Content-Type", "application/json")
	if err != nil {
		log.Println("Error encoding the body to JSON format.", err)
		w.WriteHeader(500)
		w.Write([]byte(`{"type": "error", "msg": "An error has occured"}`))
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}

func RespondWith500Error(w http.ResponseWriter) {
	RespondWithJSON(w, 500, map[string]any{
		"type": "error",
		"msg":  "An error has occured",
	})
}

func RespondWith404Error(w http.ResponseWriter) {
	RespondWithJSON(w, 404, map[string]any{
		"type": "error",
		"msg":  "The ID does not exist",
	})
}

func RespondWith403Error(w http.ResponseWriter) {
	RespondWithJSON(w, 403, map[string]any{
		"type": "error",
		"msg":  "You do not have permission to do this action",
	})
}

func RespondWith401Error(w http.ResponseWriter) {
	RespondWithJSON(w, 401, map[string]any{
		"type": "authentication_error",
		"msg":  "You are not authenticated to access this resource",
	})
}

func AskToReauthenticate(w http.ResponseWriter) {
	RespondWithJSON(w, 401, map[string]any{
		"type": "authentication_error",
		"msg":  "Invalid refresh token! Please re-authenticate!",
	})
}

func ParseJSON(w http.ResponseWriter, body io.ReadCloser, data any) error {
	decoder := json.NewDecoder(body)
	err := decoder.Decode(data)
	if err != nil {
		log.Println("Error parsing JSON data.", err)
		RespondWithJSON(w, 400, map[string]any{
			"type": "error",
			"msg":  "Error parsing JSON data",
		})
	}
	return err
}
