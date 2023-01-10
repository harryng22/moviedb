package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/harryng22/moviedb/internal/data"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

const version = "1.0.0"

type application struct {
	config Config
	logger *log.Logger
	model  data.Model
}

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	config, err := LoadConfig(".env")
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

	// Get DB Context Timeout
	contextTimeout := viper.GetDuration("DB_CONTEXT_TIMEOUT")

	app := &application{
		config: config,
		logger: logger,
		model:  data.NewModel(db, contextTimeout),
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

	db.SetMaxOpenConns(config.DbMaxOpenConns)
	db.SetMaxIdleConns(config.DbMaxIdleConns)

	duration, err := time.ParseDuration(config.DbMaxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
