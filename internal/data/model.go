package data

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Model struct {
	Movie interface {
		Insert(movie *Movie) error
		Get(id int64) (*Movie, error)
		Update(movie *Movie) error
		Delete(id int64) error
	}
}

func NewModel(db *sql.DB, contextTimeout time.Duration) Model {
	return Model{
		Movie: MovieModel{
			DB:             db,
			ContextTimeout: contextTimeout},
	}
}

func NewMockModel() Model {
	return Model{
		Movie: MockMovieModel{},
	}
}
