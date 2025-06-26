package nuke

import (
	"errors"
	"fmt"
	"github.com/go-playground/form"
	"github.com/go-playground/validator/v10"
	"mime/multipart"
	"net/http"
)

func ReadForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	f := form.NewDecoder()
	if err := f.Decode(dst, r.Form); err != nil {
		return err
	}
	v := validator.New()
	return v.Struct(dst)
}

func ReadFile(r *http.Request, fileName string, required bool, maxFileSize int64) (multipart.File, *multipart.FileHeader, error) {
	if err := r.ParseMultipartForm(maxFileSize << 20); err != nil {
		return nil, nil, err
	}
	file, handler, err := r.FormFile(fileName)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			if required == false {
				return file, handler, nil
			}
			return nil, nil, fmt.Errorf("missing required field: %s", fileName)
		}
		return nil, nil, err
	}
	defer file.Close()
	return file, handler, nil
}
