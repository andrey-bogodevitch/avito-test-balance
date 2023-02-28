package main

import (
	"database/sql"
	"fmt"
	"log"

	"balance/internal/api"
	"balance/internal/dal"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 8000
	user     = "postgres"
	password = "dev"
	dbname   = "postgres"
)

func main() {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	userStorage := dal.NewUserStorage(db)
	userHandler := api.NewHandler(userStorage)
	serv := api.NewServer("8080", userHandler)
	err = serv.Run()
	if err != nil {
		log.Fatal("server run: ", err)
	}
}
