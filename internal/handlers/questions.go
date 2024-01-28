package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vuezy/go-ask-and-answer/internal/database"
	"github.com/vuezy/go-ask-and-answer/internal/middlewares"
	"github.com/vuezy/go-ask-and-answer/internal/utils"
)

var questionMutex sync.Mutex

func GetQuestions(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	ctx := r.Context()
	query := r.URL.Query()

	if userIdStr := query.Get("user_id"); userIdStr != "" {
		userIdInt, err := strconv.ParseInt(userIdStr, 10, 32)
		if err != nil {
			log.Println("Error parsing user_id from query param.", err)
			utils.RespondWithJSON(w, 200, map[string]any{
				"type":      "success",
				"questions": []any{},
			})
			return
		}

		questions, err := db.GetQuestionsByUserId(ctx, int32(userIdInt))
		if err != nil || len(questions) == 0 {
			if err != nil {
				log.Println("Error from db.GetQuestionsByUserId method.", err)
			}
			utils.RespondWithJSON(w, 200, map[string]any{
				"type":      "success",
				"questions": []any{},
			})
			return
		}

		utils.RespondWithJSON(w, 200, map[string]any{
			"type":      "success",
			"questions": utils.ConvertStructToMap(questions),
		})
	} else {
		title := query.Get("title")
		title = "%" + strings.ReplaceAll(title, " ", "%") + "%"

		questions, err := db.SearchQuestions(ctx, title)
		if err != nil || len(questions) == 0 {
			if err != nil {
				log.Println("Error from db.SearchQuestions method.", err)
			}
			utils.RespondWithJSON(w, 200, map[string]any{
				"type":      "success",
				"questions": []any{},
			})
			return
		}

		utils.RespondWithJSON(w, 200, map[string]any{
			"type":      "success",
			"questions": utils.ConvertStructToMap(questions),
		})
	}
}

func GetQuestionById(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	ctx := r.Context()
	id := int32(ctx.Value(utils.PARSED_ID_CTX).(int64))

	question, err := db.GetQuestionById(ctx, id)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error from db.GetQuestionById method.", err)
		}
		utils.RespondWith404Error(w)
		return
	}

	utils.RespondWithJSON(w, 200, map[string]any{
		"type":     "success",
		"question": utils.ConvertStructToMap(question),
	})
}

func CreateQuestion(w http.ResponseWriter, r *http.Request) {
	questionMutex.Lock()
	defer questionMutex.Unlock()

	db := database.GetDB()
	ctx := r.Context()
	userId := utils.GetTokenSubject(ctx)
	if userId == 0 {
		utils.RespondWith401Error(w)
		return
	}

	question, ok := ctx.Value(utils.VALIDATED_CTX).(*middlewares.QuestionPayload)
	if !ok {
		log.Println("Error casting the validated request payload")
		utils.RespondWith500Error(w)
		return
	}

	user, err := db.GetUserPointsAndCredits(ctx, userId)
	if err != nil || question.PriorityLevel > user.Credits {
		if err != nil && err != sql.ErrNoRows {
			log.Println("Error from db.GetUserPointsAndCredits method.", err)
		}
		utils.RespondWithJSON(w, 400, map[string]any{
			"type": "validation_error",
			"msg": map[string]any{
				"priority_level": "You don't have enough credits to set the priority level",
			},
		})
		return
	}

	conn := database.GetDBConn()
	tx, err := conn.Begin()
	if err != nil {
		log.Println("Error starting a transaction.", err)
		utils.RespondWith500Error(w)
		return
	}
	qtx := db.WithTx(tx)

	err = qtx.CreateQuestion(ctx, database.CreateQuestionParams{
		Title:         question.Title,
		Body:          question.Body,
		PriorityLevel: question.PriorityLevel,
		UserID:        userId,
		RespondedAt:   time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})
	if err != nil {
		log.Println("Error from qtx.CreateQuestion method.", err)
		tx.Rollback()
		utils.RespondWith500Error(w)
		return
	}

	err = qtx.RemoveUserCredit(ctx, database.RemoveUserCreditParams{
		ID:        userId,
		Credits:   question.PriorityLevel,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Println("Error from qtx.RemoveUserCredit method.", err)
		tx.Rollback()
		utils.RespondWith500Error(w)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error commiting the transaction.", err)
		utils.RespondWith500Error(w)
		return
	}

	utils.RespondWithJSON(w, 201, map[string]any{
		"type": "success",
		"msg":  "The question has been created",
	})
}

func UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	questionMutex.Lock()
	defer questionMutex.Unlock()

	db := database.GetDB()
	ctx := r.Context()
	userId := utils.GetTokenSubject(ctx)
	if userId == 0 {
		utils.RespondWith401Error(w)
		return
	}

	questionId := int32(ctx.Value(utils.PARSED_ID_CTX).(int64))
	question, err := db.GetQuestionById(ctx, questionId)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error from db.GetQuestionById method.", err)
		}
		utils.RespondWith404Error(w)
		return
	}
	if question.UserID != userId {
		utils.RespondWith403Error(w)
		return
	}

	questionPayload, ok := ctx.Value(utils.VALIDATED_CTX).(*middlewares.QuestionPayload)
	if !ok {
		log.Println("Error casting the validated request payload")
		utils.RespondWith500Error(w)
		return
	}

	priorityLevel := max(0, questionPayload.PriorityLevel-question.PriorityLevel)
	user, err := db.GetUserPointsAndCredits(ctx, userId)
	if err != nil || priorityLevel > user.Credits {
		if err != nil && err != sql.ErrNoRows {
			log.Println("Error from db.GetUserPointsAndCredits method.", err)
		}
		utils.RespondWithJSON(w, 400, map[string]any{
			"type": "validation_error",
			"msg": map[string]any{
				"priority_level": "You don't have enough credits to set this priority level",
			},
		})
		return
	}

	conn := database.GetDBConn()
	tx, err := conn.Begin()
	if err != nil {
		log.Println("Error starting a transaction.", err)
		utils.RespondWith500Error(w)
		return
	}
	qtx := db.WithTx(tx)

	err = qtx.UpdateQuestion(ctx, database.UpdateQuestionParams{
		ID:            questionId,
		UserID:        userId,
		Title:         questionPayload.Title,
		Body:          questionPayload.Body,
		PriorityLevel: priorityLevel,
		UpdatedAt:     time.Now(),
	})
	if err != nil {
		log.Println("Error from qtx.UpdateQuestion method.", err)
		tx.Rollback()
		utils.RespondWith500Error(w)
		return
	}

	err = qtx.RemoveUserCredit(ctx, database.RemoveUserCreditParams{
		ID:        userId,
		Credits:   priorityLevel,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Println("Error from qtx.RemoveUserCredit method.", err)
		tx.Rollback()
		utils.RespondWith500Error(w)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error commiting the transaction.", err)
		utils.RespondWith500Error(w)
		return
	}

	utils.RespondWithJSON(w, 200, map[string]any{
		"type": "success",
		"msg":  "The question has been updated",
	})
}

func DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	ctx := r.Context()
	userId := utils.GetTokenSubject(ctx)
	if userId == 0 {
		utils.RespondWith401Error(w)
		return
	}

	questionId := int32(ctx.Value(utils.PARSED_ID_CTX).(int64))
	question, err := db.GetQuestionById(ctx, questionId)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error from db.GetQuestionById method.", err)
		}
		utils.RespondWith404Error(w)
		return
	}
	if question.UserID != userId {
		utils.RespondWith403Error(w)
		return
	}

	err = db.DeleteQuestion(ctx, database.DeleteQuestionParams{
		ID:     questionId,
		UserID: userId,
	})
	if err != nil {
		log.Println("Error from db.DeleteQuestion method.", err)
		utils.RespondWith500Error(w)
		return
	}

	utils.RespondWithJSON(w, 200, map[string]any{
		"type": "success",
		"msg":  "The question has been deleted",
	})
}

func CloseQuestion(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	ctx := r.Context()
	userId := utils.GetTokenSubject(ctx)
	if userId == 0 {
		utils.RespondWith401Error(w)
		return
	}

	questionId := int32(ctx.Value(utils.PARSED_ID_CTX).(int64))
	question, err := db.GetQuestionById(ctx, questionId)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error from db.GetQuestionById method.", err)
		}
		utils.RespondWith404Error(w)
		return
	}
	if question.UserID != userId {
		utils.RespondWith403Error(w)
		return
	}

	conn := database.GetDBConn()
	tx, err := conn.Begin()
	if err != nil {
		log.Println("Error starting a transaction.", err)
		utils.RespondWith500Error(w)
		return
	}
	qtx := db.WithTx(tx)

	err = qtx.CloseQuestion(ctx, database.CloseQuestionParams{
		ID:        questionId,
		UserID:    userId,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Println("Error from qtx.CloseQuestion method.", err)
		tx.Rollback()
		utils.RespondWith500Error(w)
		return
	}

	answers, err := qtx.GetAnswersByQuestionId(ctx, questionId)
	if err != nil {
		log.Println("Error from qtx.GetAnswersByQuestionId method.", err)
		tx.Rollback()
		utils.RespondWith500Error(w)
		return
	}

	for _, answer := range answers {
		err = qtx.AddUserCredit(ctx, database.AddUserCreditParams{
			ID:        answer.UserID,
			Credits:   answer.Votes,
			UpdatedAt: time.Now(),
		})
		if err != nil {
			log.Println("Error from qtx.AddUserCredit method.", err)
			tx.Rollback()
			utils.RespondWith500Error(w)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error commiting the transaction.", err)
		utils.RespondWith500Error(w)
		return
	}

	utils.RespondWithJSON(w, 200, map[string]any{
		"type": "success",
		"msg":  "The question has been closed",
	})
}
