package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/vuezy/go-ask-and-answer/internal/database"
	"github.com/vuezy/go-ask-and-answer/internal/middlewares"
	"github.com/vuezy/go-ask-and-answer/internal/utils"
)

func Login(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	ctx := r.Context()
	credentials, ok := ctx.Value(utils.VALIDATED_CTX).(*middlewares.LoginPayload)
	if !ok {
		log.Println("Error casting the validated request payload")
		utils.RespondWith500Error(w)
		return
	}

	user, err := db.Login(ctx, credentials.Email)
	if err != nil {
		log.Println("Error from db.Login method.", err)
		utils.RespondWithJSON(w, 400, map[string]any{
			"type": "error",
			"msg":  "Invalid email or password",
		})
		return
	}

	err = utils.CheckPasswordHash(credentials.Password, user.Password)
	if err != nil {
		log.Println("Error from utils.CheckPasswordHash function.", err)
		utils.RespondWithJSON(w, 400, map[string]any{
			"type": "error",
			"msg":  "Invalid email or password",
		})
		return
	}

	jti, accessTokenStr, refreshTokenStr, err := utils.IssueJWT(user.ID)
	if err != nil {
		utils.RespondWith500Error(w)
		return
	}

	err = db.SetActiveToken(ctx, database.SetActiveTokenParams{
		Sub: user.ID,
		Jti: jti,
	})
	if err != nil {
		log.Println("Error from db.SetActiveToken method.", err)
		utils.RespondWith500Error(w)
		return
	}

	utils.RespondWithJSON(w, 200, map[string]any{
		"type": "success",
		"user": map[string]any{
			"id":      user.ID,
			"name":    user.Name,
			"email":   user.Email,
			"points":  user.Points,
			"credits": user.Credits,
		},
		"access_token":  accessTokenStr,
		"refresh_token": refreshTokenStr,
	})
}

func Register(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	ctx := r.Context()
	data, ok := ctx.Value(utils.VALIDATED_CTX).(*middlewares.RegisterPayload)
	if !ok {
		log.Println("Error casting the validated request payload")
		utils.RespondWith500Error(w)
		return
	}

	count, err := db.CheckIfEmailExists(ctx, data.Email)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error from db.CheckIfEmailExists method.", err)
		utils.RespondWith500Error(w)
		return
	}
	if count >= 1 {
		utils.RespondWithJSON(w, 400, map[string]any{
			"type": "validation_error",
			"msg": map[string]any{
				"email": "Email is already in use",
			},
		})
		return
	}

	hashedPassword, err := utils.HashPassword(data.Password)
	if err != nil {
		log.Println("Error hashing password.", err)
		utils.RespondWithJSON(w, 400, map[string]any{
			"type": "validation_error",
			"msg": map[string]any{
				"password": "Bad password",
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

	err = qtx.RegisterUser(ctx, database.RegisterUserParams{
		Name:      data.Name,
		Email:     data.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Println("Error from qtx.RegisterUser method.", err)
		tx.Rollback()
		utils.RespondWith500Error(w)
		return
	}

	user, err := qtx.Login(ctx, data.Email)
	if err != nil {
		log.Println("Error from qtx.Login method.", err)
		tx.Rollback()
		utils.RespondWith500Error(w)
		return
	}

	jti, accessTokenStr, refreshTokenStr, err := utils.IssueJWT(user.ID)
	if err != nil {
		tx.Rollback()
		utils.RespondWith500Error(w)
		return
	}

	err = qtx.SetActiveToken(ctx, database.SetActiveTokenParams{
		Sub: user.ID,
		Jti: jti,
	})
	if err != nil {
		log.Println("Error from qtx.SetActiveToken method.", err)
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
		"user": map[string]any{
			"id":      user.ID,
			"name":    user.Name,
			"email":   user.Email,
			"points":  user.Points,
			"credits": user.Credits,
		},
		"access_token":  accessTokenStr,
		"refresh_token": refreshTokenStr,
	})
}

func GetUserPointsAndCredits(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	ctx := r.Context()
	userId := utils.GetTokenSubject(ctx)
	if userId == 0 {
		utils.RespondWith401Error(w)
		return
	}

	user, err := db.GetUserPointsAndCredits(ctx, userId)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error from db.GetUserPointsAndCredits method.", err)
		}
		utils.RespondWith401Error(w)
		return
	}

	utils.RespondWithJSON(w, 200, map[string]any{
		"type":    "success",
		"points":  user.Points,
		"credits": user.Credits,
	})
}
