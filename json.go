package nuke

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
	"strings"
)

func ReadJSON(r *http.Request, dst interface{}) error {
	contentType := r.Header.Get(contentType)
	if contentType != "" && strings.ToLower(contentType) != applicationJSON {
		return ErrNotJson
	}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	err := d.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return ErrInvalidJson
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return ErrEOFJson
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	if err = d.Decode(&struct{}{}); err != io.EOF {
		return ErrDuplicateJson
	}
	v := validator.New()
	return v.Struct(dst)
}

func WriteJSON(w http.ResponseWriter, statusCode int, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	w.Header().Set(contentType, applicationJSON)
	w.WriteHeader(statusCode)
	if _, err := w.Write(data); err != nil {
		panic(err)
	}
}
