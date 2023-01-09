package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const version = "1.0.0"

type application struct {
	config Config
	logger *log.Logger
}

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	config, err := LoadConfig("../..")
	if err != nil {
		logger.Fatal(err)
	}

	// db connect
	db, err := openDB(config)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	logger.Println("database connection pool established")

	app := &application{
		config: config,
		logger: logger,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("Starting %s server on %s", config.Env, server.Addr)
	err = server.ListenAndServe()
	logger.Fatal(err)
}

func openDB(config Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.DbDsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
