package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/harryng22/moviedb/internal/data"
	"github.com/harryng22/moviedb/internal/validator"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
)

func (app *application) readIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid Id parameter")
	}

	return id, nil
}

type envelope map[string]interface{}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) readString(queryString url.Values, key, defaultValue string) string {
	result := queryString.Get(key)

	if result == "" {
		return defaultValue
	}

	return result
}

func (app *application) readCSV(queryString url.Values, key string, defaultValue []string) []string {
	csv := queryString.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *application) readInt(queryString url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := queryString.Get(key)
	if s == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return intValue
}

type Config struct {
	Port           int    `mapstructure:"PORT"`
	Env            string `mapstructure:"ENV"`
	DbDsn          string `mapstructure:"DB_DSN"`
	DbMaxOpenConns int    `mapstructure:"DB_MAX_OPEN_CONNS"`
	DbMaxIdleConns int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	DbMaxIdleTime  string `mapstructure:"DB_MAX_IDLE_TIME"`
}

func LoadConfig(filePath string) (config Config, err error) {
	viper.AddConfigPath(filepath.Dir(filePath))
	viper.SetConfigName(filepath.Base(filePath))
	viper.SetConfigType(strings.TrimPrefix(filepath.Ext(filePath), "."))

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}


func copyProperties(input Input, movie *data.Movie) {
	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}
}