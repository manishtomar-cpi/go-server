package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/manishtomar-cpi/go-server/internal/storage"
	"github.com/manishtomar-cpi/go-server/internal/types"
	"github.com/manishtomar-cpi/go-server/internal/utills/response"
)

func Ready() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { // w is response , r is request
		w.Write([]byte("welcome to go server"))
	}
}

func New(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { // w is response , r is request
		var student types.Student
		err := json.NewDecoder(r.Body).Decode(&student) // what data is comimng decode it in the student var
		if errors.Is(err, io.EOF) {                     // if getting blank body
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		//any general errro
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}
		//validation of request
		validationError := validator.New().Struct(student)
		if validationError != nil {
			validateErrs := validationError.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}
		//calling function
		lastId, err := storage.CreateStudent(
			student.Name,
			student.Email,
			student.Age,
		)
		slog.Info("user created", slog.String("userId", fmt.Sprint(lastId)))
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, err)
		}
		response.WriteJson(w, http.StatusCreated, map[string]int64{"id": lastId})

	}
}
