package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/vuezy/go-ask-and-answer/internal/database"
	"github.com/vuezy/go-ask-and-answer/internal/utils"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("Error loading .env file.", err)
	}
}

func main() {
	database.ConnectToDatabase()
	utils.UseJWTAuthentication()
	router := setUpRouter()

	server := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: router,
	}

	log.Println("Server is running on port", os.Getenv("PORT"))
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalln("An error has occured.", err)
	}
}
