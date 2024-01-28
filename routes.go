package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vuezy/go-ask-and-answer/internal/handlers"
	"github.com/vuezy/go-ask-and-answer/internal/middlewares"
)

func setUpRouter() http.Handler {
	router := chi.NewRouter()
	v1Router := chi.NewRouter()

	// Public routes
	v1Router.Group(func(r chi.Router) {
		r.With(middlewares.ValidateLoginPayload).
			Post("/login", handlers.Login)
		r.With(middlewares.ValidateRegisterPayload).
			Post("/register", handlers.Register)
		r.With(middlewares.VerifyRefreshToken).
			Post("/refresh", handlers.RefreshTokens)
	})

	// Protected routes (require authentication)
	v1Router.Group(func(r chi.Router) {
		r.Use(middlewares.VerifyAccessToken)

		r.Get("/credits", handlers.GetUserPointsAndCredits)

		r.Get("/questions", handlers.GetQuestions)
		r.With(middlewares.ParseIdFromURLParam).
			Get("/question/{id}", handlers.GetQuestionById)
		r.With(middlewares.ValidateQuestionPayload).
			Post("/question", handlers.CreateQuestion)
		r.With(middlewares.ParseIdFromURLParam, middlewares.ValidateQuestionPayload).
			Put("/question/{id}", handlers.UpdateQuestion)
		r.With(middlewares.ParseIdFromURLParam).
			Delete("/question/{id}", handlers.DeleteQuestion)
		r.With(middlewares.ParseIdFromURLParam).
			Patch("/question/{id}", handlers.CloseQuestion)

		r.With(middlewares.ParseIdFromURLParam).
			Get("/question/{id}/answers", handlers.GetAnswersByQuestionId)
		r.Get("/answers", handlers.GetAnswersByUserId)
		r.With(middlewares.ParseIdFromURLParam).
			Get("/answer/{id}", handlers.GetAnswerById)
		r.With(middlewares.ValidateAnswerPayload).
			Post("/answer", handlers.CreateAnswer)
		r.With(middlewares.ParseIdFromURLParam, middlewares.ValidateAnswerPayload).
			Patch("/answer/{id}", handlers.UpdateAnswer)
		r.With(middlewares.ParseIdFromURLParam).
			Delete("/answer/{id}", handlers.DeleteAnswer)
		r.With(middlewares.ParseIdFromURLParam).
			Patch("/answer/{id}/upvote", handlers.VoteAnswer)
		r.With(middlewares.ParseIdFromURLParam).
			Patch("/answer/{id}/downvote", handlers.VoteAnswer)
	})

	router.Mount("/v1", v1Router)
	return router
}
