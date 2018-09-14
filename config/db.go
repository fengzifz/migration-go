package config

import (
	"database/sql"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
)

var db *sql.DB

func Conf() *sql.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Database configuration
	conn := os.Getenv("DB_CONNECTION")
	dbName := os.Getenv("DB_DATABASE")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	str := []string{username, ":", password, "@/", dbName}
	connInfo := strings.Join(str, "")

	db, err = sql.Open(conn, connInfo)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
