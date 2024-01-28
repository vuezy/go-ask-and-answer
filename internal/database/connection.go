package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var dbPool *Queries
var dbConn *sql.DB

func ConnectToDatabase() {
	var err error
	dsn := os.Getenv("DSN")
	dbConn, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalln("Error connecting to the database.", err)
	}

	err = dbConn.Ping()
	if err != nil {
		log.Fatalln("Error connecting to the database.", err)
	}

	log.Println("Connected to the database")
	dbPool = New(dbConn)
}

func GetDB() *Queries {
	return dbPool
}

func GetDBConn() *sql.DB {
	return dbConn
}
