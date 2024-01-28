package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/vuezy/go-ask-and-answer/internal/database"
	"github.com/vuezy/go-ask-and-answer/internal/utils"
)

func RefreshTokens(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	ctx := r.Context()
	refreshTokenStr, refreshToken, _ := utils.JWTAuthFromContext(ctx)
	accessTokenStr, accessToken := utils.AccessTokenFromContext(ctx)

	refreshTokenSub := utils.GetTokenSubject(ctx)
	if refreshTokenSub == 0 {
		utils.AskToReauthenticate(w)
		return
	}

	sub, err := strconv.ParseInt(accessToken.Subject(), 10, 32)
	if err != nil {
		log.Println("Error converting accessToken.Subject() to integer.", err)
		utils.AskToReauthenticate(w)
		return
	}
	accessTokenSub := int32(sub)

	if refreshTokenSub != accessTokenSub ||
		refreshToken.JwtID() != accessToken.JwtID() ||
		!(refreshToken.IssuedAt().Equal(accessToken.IssuedAt())) {
		utils.AskToReauthenticate(w)
		return
	}

	jti, accessTokenStr, refreshTokenStr, err := utils.IssueJWT(refreshTokenSub)
	if err != nil {
		utils.RespondWith500Error(w)
		return
	}

	err = db.SetActiveToken(ctx, database.SetActiveTokenParams{
		Sub: refreshTokenSub,
		Jti: jti,
	})
	if err != nil {
		log.Println("Error from db.SetActiveToken method.", err)
		utils.RespondWith500Error(w)
		return
	}

	utils.RespondWithJSON(w, 200, map[string]any{
		"type":          "success",
		"access_token":  accessTokenStr,
		"refresh_token": refreshTokenStr,
	})
}
