package utils

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/vuezy/go-ask-and-answer/internal/database"
)

var jwtAuth *jwtauth.JWTAuth

func UseJWTAuthentication() {
	secret := os.Getenv("JWT_SECRET")
	jwtAuth = jwtauth.New("HS256", []byte(secret), nil)
}

func GetJWTAuth() *jwtauth.JWTAuth {
	return jwtAuth
}

func IssueJWT(userId int32) (string, string, string, error) {
	jti := uuid.New().String()
	sub := strconv.Itoa(int(userId))
	iat := time.Now().UTC().Unix()
	exp := time.Now().UTC().Unix()

	accessTokenClaims := map[string]interface{}{
		"jti":     jti,
		"sub":     sub,
		"iat":     iat,
		"exp":     exp + int64((time.Hour * 1).Seconds()),
		"refresh": false,
	}

	_, accessToken, err := jwtAuth.Encode(accessTokenClaims)
	if err != nil {
		log.Println("Error issuing access token.", err)
		return "", "", "", err
	}

	refreshTokenClaims := map[string]interface{}{
		"jti":     jti,
		"sub":     sub,
		"iat":     iat,
		"exp":     exp + int64((time.Hour * 24 * 3).Seconds()),
		"refresh": true,
	}

	_, refreshToken, err := jwtAuth.Encode(refreshTokenClaims)
	if err != nil {
		log.Println("Error issuing refresh token.", err)
		return "", "", "", err
	}
	return jti, accessToken, refreshToken, nil
}

func VerifyToken(
	ctx context.Context,
	tokenStr string,
	isRefreshToken bool,
	clock jwt.Clock,
) (token jwt.Token, claims map[string]any, code int) {
	db := database.GetDB()

	token, err := jwtAuth.Decode(tokenStr)
	if err != nil || token == nil {
		if err != nil {
			log.Println("Error validating token.", err)
		}
		return nil, nil, 401
	}

	if clock != nil {
		if err = jwt.Validate(token, jwt.WithClock(clock)); err != nil {
			log.Println("Error validating token.", err)
			if err.Error() != "iat not satisfied" {
				return nil, nil, 401
			}
		}
	} else {
		if err = jwt.Validate(token); err != nil {
			log.Println("Error validating token.", err)
			return nil, nil, 401
		}
	}

	if token.JwtID() == "" ||
		token.Subject() == "" ||
		(token.IssuedAt().Equal(time.Time{})) ||
		(token.Expiration().Equal(time.Time{})) {
		return nil, nil, 401
	}

	sub, err := strconv.ParseInt(token.Subject(), 10, 32)
	if err != nil {
		log.Println("Error converting token.Subject() to integer.", err)
		return nil, nil, 401
	}

	count, err := db.CheckIfTokenIsActive(ctx, database.CheckIfTokenIsActiveParams{
		Sub: int32(sub),
		Jti: token.JwtID(),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, 401
		}
		log.Println("Error from db.CheckIfTokenIsActive method.", err)
		return nil, nil, 500
	}
	if count == 0 {
		return nil, nil, 401
	}

	claims, err = token.AsMap(ctx)
	if err != nil {
		log.Println("Error validating token.", err)
		return nil, nil, 401
	}

	if isRefreshToken && claims["refresh"] == true {
		return token, claims, 200
	}
	if !isRefreshToken && claims["refresh"] == false {
		return token, claims, 200
	}
	return nil, nil, 401
}

func ExtractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && strings.ToUpper(authHeader[0:6]) == "BEARER" {
		return authHeader[7:]
	} else {
		return ""
	}
}

func GetTokenSubject(ctx context.Context) int32 {
	_, token, _ := JWTAuthFromContext(ctx)
	sub, err := strconv.ParseInt(token.Subject(), 10, 32)
	if err != nil {
		log.Println("Error converting token.Subject() to integer.", err)
		return 0
	}
	return int32(sub)
}

func JWTAuthFromContext(ctx context.Context) (string, jwt.Token, map[string]any) {
	tokenStr, ok := ctx.Value(TOKEN_STR_CTX).(string)
	if !ok {
		return "", nil, nil
	}
	jwtToken, ok := ctx.Value(JWT_TOKEN_CTX).(jwt.Token)
	if !ok {
		return tokenStr, nil, nil
	}
	claims, ok := ctx.Value(CLAIMS_CTX).(map[string]any)
	if !ok {
		return tokenStr, jwtToken, nil
	}

	return tokenStr, jwtToken, claims
}

func AccessTokenFromContext(ctx context.Context) (string, jwt.Token) {
	tokenStr, ok := ctx.Value(ACCESS_TOKEN_STR_CTX).(string)
	if !ok {
		return "", nil
	}
	jwtToken, ok := ctx.Value(ACCESS_TOKEN_CTX).(jwt.Token)
	if !ok {
		return tokenStr, nil
	}

	return tokenStr, jwtToken
}
