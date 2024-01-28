package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/vuezy/go-ask-and-answer/internal/database"
	"github.com/vuezy/go-ask-and-answer/internal/middlewares"
	"github.com/vuezy/go-ask-and-answer/internal/utils"
)

var answerMutex sync.Mutex

func GetAnswersByQuestionId(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	ctx := r.Context()
	questionId := int32(ctx.Value(utils.PARSED_ID_CTX).(int64))

	answers, err := db.GetAnswersByQuestionId(ctx, questionId)
	if err != nil || len(answers) == 0 {
		if err != nil {
			log.Println("Error from db.GetAnswersByQuestionId method.", err)
		}
		utils.RespondWithJSON(w, 200, map[string]any{
			"type":    "success",
			"answers": []any{},
		})
		return
	}

	utils.RespondWithJSON(w, 200, map[string]any{
		"type":    "success",
		"answers": utils.ConvertStructToMap(answers),
	})
}

func GetAnswersByUserId(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	ctx := r.Context()
	userId := utils.GetTokenSubject(ctx)
	if userId == 0 {
		utils.RespondWith401Error(w)
		return
	}

	answers, err := db.GetAnswersByUserId(ctx, userId)
	if err != nil || len(answers) == 0 {
		if err != nil {
			log.Println("Error from db.GetAnswersByUserId method.", err)
		}
		utils.RespondWithJSON(w, 200, map[string]any{
			"type":    "success",
			"answers": []any{},
		})
		return
	}

	utils.RespondWithJSON(w, 200, map[string]any{
		"type":    "success",
		"answers": utils.ConvertStructToMap(answers),
	})
}

func GetAnswerById(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	ctx := r.Context()
	id := int32(ctx.Value(utils.PARSED_ID_CTX).(int64))

	answer, err := db.GetAnswerById(ctx, id)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error from db.GetAnswerById method.", err)
		}
		utils.RespondWith404Error(w)
		return
	}

	utils.RespondWithJSON(w, 200, map[string]any{
		"type":   "success",
		"answer": utils.ConvertStructToMap(answer),
	})
}

func CreateAnswer(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	ctx := r.Context()
	userId := utils.GetTokenSubject(ctx)
	if userId == 0 {
		utils.RespondWith401Error(w)
		return
	}

	answer, ok := ctx.Value(utils.VALIDATED_CTX).(*middlewares.AnswerPayload)
	if !ok {
		log.Println("Error casting the validated request payload")
		utils.RespondWith500Error(w)
		return
	}

	_, err := db.GetQuestionById(ctx, answer.QuestionID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error from db.GetQuestionById method.", err)
		}
		utils.RespondWith404Error(w)
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

	err = qtx.CreateAnswer(ctx, database.CreateAnswerParams{
		Body:       answer.Body,
		QuestionID: answer.QuestionID,
		UserID:     userId,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	})
	if err != nil {
		log.Println("Error from qtx.CreateAnswer method.", err)
		tx.Rollback()
		utils.RespondWith500Error(w)
		return
	}

	err = qtx.RespondToQuestion(ctx, database.RespondToQuestionParams{
		ID:          answer.QuestionID,
		RespondedAt: time.Now(),
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		log.Println("Error from qtx.RespondToQuestion method.", err)
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
		"msg":  "The answer has been created",
	})
}

func UpdateAnswer(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	ctx := r.Context()
	userId := utils.GetTokenSubject(ctx)
	if userId == 0 {
		utils.RespondWith401Error(w)
		return
	}

	answerId := int32(ctx.Value(utils.PARSED_ID_CTX).(int64))
	answer, err := db.GetAnswerById(ctx, answerId)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error from db.GetAnswerById method.", err)
		}
		utils.RespondWith404Error(w)
		return
	}
	if answer.UserID != userId {
		utils.RespondWith403Error(w)
		return
	}

	question, err := db.GetQuestionById(ctx, answer.QuestionID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error from db.GetQuestionById method.", err)
		}
		utils.RespondWith500Error(w)
		return
	}
	if question.Closed {
		utils.RespondWithJSON(w, 400, map[string]any{
			"type": "error",
			"msg":  "Cannot update the answer because the question has been closed",
		})
		return
	}

	answerPayload, ok := ctx.Value(utils.VALIDATED_CTX).(*middlewares.AnswerPayload)
	if !ok {
		log.Println("Error casting the validated request payload")
		utils.RespondWith500Error(w)
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

	err = qtx.UpdateAnswer(ctx, database.UpdateAnswerParams{
		ID:        answerId,
		UserID:    userId,
		Body:      answerPayload.Body,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Println("Error from qtx.UpdateAnswer method.", err)
		tx.Rollback()
		utils.RespondWith500Error(w)
		return
	}

	err = qtx.RespondToQuestion(ctx, database.RespondToQuestionParams{
		ID:          answer.QuestionID,
		RespondedAt: time.Now(),
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		log.Println("Error from qtx.RespondToQuestion method.", err)
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
		"msg":  "The answer has been updated",
	})
}

func DeleteAnswer(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	ctx := r.Context()
	userId := utils.GetTokenSubject(ctx)
	if userId == 0 {
		utils.RespondWith401Error(w)
		return
	}

	answerId := int32(ctx.Value(utils.PARSED_ID_CTX).(int64))
	answer, err := db.GetAnswerById(ctx, answerId)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error from db.GetAnswerById method.", err)
		}
		utils.RespondWith404Error(w)
		return
	}
	if answer.UserID != userId {
		utils.RespondWith403Error(w)
		return
	}

	question, err := db.GetQuestionById(ctx, answer.QuestionID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error from db.GetQuestionById method.", err)
		}
		utils.RespondWith500Error(w)
		return
	}
	if question.Closed {
		utils.RespondWithJSON(w, 400, map[string]any{
			"type": "error",
			"msg":  "Cannot delete the answer because the question has been closed",
		})
		return
	}

	err = db.DeleteAnswer(ctx, database.DeleteAnswerParams{
		ID:     answerId,
		UserID: userId,
	})
	if err != nil {
		log.Println("Error from db.DeleteAnswer method.", err)
		utils.RespondWith500Error(w)
		return
	}

	utils.RespondWithJSON(w, 200, map[string]any{
		"type": "success",
		"msg":  "The answer has been deleted",
	})
}

func VoteAnswer(w http.ResponseWriter, r *http.Request) {
	answerMutex.Lock()
	defer answerMutex.Unlock()

	db := database.GetDB()
	ctx := r.Context()
	userId := utils.GetTokenSubject(ctx)
	if userId == 0 {
		utils.RespondWith401Error(w)
		return
	}

	answerId := int32(ctx.Value(utils.PARSED_ID_CTX).(int64))
	answer, err := db.GetAnswerById(ctx, answerId)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error from db.GetAnswerById method.", err)
		}
		utils.RespondWith404Error(w)
		return
	}

	if answer.UserID == userId {
		utils.RespondWithJSON(w, 400, map[string]any{
			"type": "error",
			"msg":  "You cannot vote your own answer",
		})
		return
	}

	vote, err := db.CheckIfVoteExists(ctx, database.CheckIfVoteExistsParams{
		AnswerID: answerId,
		UserID:   userId,
	})
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error from db.CheckIfVoteExists method.", err)
		utils.RespondWith500Error(w)
		return
	}

	segment := path.Base(r.URL.Path)
	doCreate := err == sql.ErrNoRows
	var newVote int32

	if segment == "upvote" && !doCreate && vote.Val == 1 {
		utils.RespondWithJSON(w, 400, map[string]any{
			"type": "error",
			"msg":  "You already upvoted this answer",
		})
		return
	} else if segment == "downvote" && !doCreate && vote.Val == -1 {
		utils.RespondWithJSON(w, 400, map[string]any{
			"type": "error",
			"msg":  "You already downvoted this answer",
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

	if segment == "upvote" {
		if doCreate {
			err = qtx.CreateVote(ctx, database.CreateVoteParams{
				Val:       1,
				AnswerID:  answerId,
				UserID:    userId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
			if err != nil {
				log.Println("Error from qtx.CreateVote method.", err)
				tx.Rollback()
				utils.RespondWith500Error(w)
				return
			}
			newVote = 1

		} else {
			err = qtx.Upvote(ctx, database.UpvoteParams{
				ID:        vote.ID,
				UpdatedAt: time.Now(),
			})
			if err != nil {
				log.Println("Error from qtx.Upvote method.", err)
				tx.Rollback()
				utils.RespondWith500Error(w)
				return
			}
			newVote = -vote.Val + 1
		}

	} else if segment == "downvote" {
		if doCreate {
			err = qtx.CreateVote(ctx, database.CreateVoteParams{
				Val:       -1,
				AnswerID:  answerId,
				UserID:    userId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
			if err != nil {
				log.Println("Error from qtx.CreateVote method.", err)
				tx.Rollback()
				utils.RespondWith500Error(w)
				return
			}
			newVote = -1

		} else {
			err = qtx.Downvote(ctx, database.DownvoteParams{
				ID:        vote.ID,
				UpdatedAt: time.Now(),
			})
			if err != nil {
				log.Println("Error from qtx.Downvote method.", err)
				tx.Rollback()
				utils.RespondWith500Error(w)
				return
			}
			newVote = -vote.Val - 1
		}
	}

	err = qtx.UpdateAnswerVotes(ctx, database.UpdateAnswerVotesParams{
		ID:        answerId,
		Votes:     newVote,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Println("Error from qtx.UpdateAnswerVotes method.", err)
		tx.Rollback()
		utils.RespondWith500Error(w)
		return
	}

	err = qtx.UpdateUserPoints(ctx, database.UpdateUserPointsParams{
		ID:        answer.UserID,
		Points:    newVote,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Println("Error from qtx.UpdateUserPoints method.", err)
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

	if segment == "upvote" {
		utils.RespondWithJSON(w, 200, map[string]any{
			"type":  "success",
			"msg":   "The answer has been upvoted",
			"votes": answer.Votes + newVote,
		})
	} else if segment == "downvote" {
		utils.RespondWithJSON(w, 200, map[string]any{
			"type":  "success",
			"msg":   "The answer has been downvoted",
			"votes": answer.Votes + newVote,
		})
	}
}
