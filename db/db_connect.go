package db

import (
	// "log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func ConnectToDB() *sqlx.DB {

	err := godotenv.Load(".env")
	if err != nil {
		// log.Fatalf("Error loading .env file")
	}

	dbConnectionString := os.Getenv("POSTGRESQL_URL")

	db, err := sqlx.Open("postgres", dbConnectionString)
	if err != nil {
		// log.Fatal("error opening database connection: ", err)
	}
	err = db.Ping()
	if err != nil {
		// log.Fatal("error while pinging database: ", err)
	}

	return db
}
