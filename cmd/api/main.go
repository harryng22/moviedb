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
	"github.com/harryng22/moviedb/internal/jsonlog"
	_ "github.com/lib/pq"
)

const version = "1.0.0"

type application struct {
	config Config
	logger *jsonlog.Logger
	model  data.Model
}

func main() {
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	config, err := LoadConfig(".env")
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// db connect
	db, err := openDB(config)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: config,
		logger: logger,
		model:  data.NewModel(db),
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      app.routes(),
		ErrorLog:     log.New(logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.PrintInfo("Starting server", map[string]string{
		"addr": server.Addr,
		"env":  config.Env,
	})

	err = server.ListenAndServe()
	logger.PrintFatal(err, nil)
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
