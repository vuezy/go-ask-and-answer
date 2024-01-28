package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/vuezy/go-ask-and-answer/internal/utils"
)

func VerifyAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tokenStr := utils.ExtractTokenFromHeader(r)
		if tokenStr == "" {
			utils.RespondWith401Error(w)
			return
		}

		token, claims, code := utils.VerifyToken(ctx, tokenStr, false, nil)
		if code == 401 {
			utils.RespondWith401Error(w)
			return
		} else if code == 500 {
			utils.RespondWith500Error(w)
			return
		}

		ctx = context.WithValue(ctx, utils.TOKEN_STR_CTX, tokenStr)
		ctx = context.WithValue(ctx, utils.JWT_TOKEN_CTX, token)
		ctx = context.WithValue(ctx, utils.CLAIMS_CTX, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func VerifyRefreshToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		refreshTokenStr := utils.ExtractTokenFromHeader(r)
		if refreshTokenStr == "" {
			utils.AskToReauthenticate(w)
			return
		}

		refreshToken, refreshTokenClaims, code := utils.VerifyToken(ctx, refreshTokenStr, true, nil)
		if code == 401 {
			utils.AskToReauthenticate(w)
			return
		} else if code == 500 {
			utils.RespondWith500Error(w)
			return
		}

		ctx = context.WithValue(ctx, utils.TOKEN_STR_CTX, refreshTokenStr)
		ctx = context.WithValue(ctx, utils.JWT_TOKEN_CTX, refreshToken)
		ctx = context.WithValue(ctx, utils.CLAIMS_CTX, refreshTokenClaims)

		type payload struct {
			AccessToken string `json:"access_token"`
		}
		data := payload{}
		if err := utils.ParseJSON(w, r.Body, &data); err != nil {
			return
		}

		accessTokenStr := data.AccessToken
		if accessTokenStr == "" {
			utils.AskToReauthenticate(w)
			return
		}

		clock := jwt.ClockFunc(func() time.Time {
			return time.Time{}
		})
		accessToken, _, code := utils.VerifyToken(ctx, accessTokenStr, false, clock)
		if code == 401 {
			utils.AskToReauthenticate(w)
			return
		} else if code == 500 {
			utils.RespondWith500Error(w)
			return
		}

		ctx = context.WithValue(ctx, utils.ACCESS_TOKEN_STR_CTX, accessTokenStr)
		ctx = context.WithValue(ctx, utils.ACCESS_TOKEN_CTX, accessToken)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
