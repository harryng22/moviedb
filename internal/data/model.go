package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")
)

type Model struct {
	Movie interface {
		Insert(movie *Movie) error
		Get(id int64) (*Movie, error)
		Update(movie *Movie) error
		Delete(id int64) error
	}
}

func NewModel(db *sql.DB) Model {
	return Model{
		Movie: MovieModel{DB: db},
	}
}

func NewMockModel() Model {
	return Model{
		Movie: MockMovieModel{},
	}
}
