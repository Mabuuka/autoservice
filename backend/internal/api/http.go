package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type DataResponse struct {
	Data any `json:"data"`
}

type MutationResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type DeleteResponse struct {
	Message   string `json:"message"`
	DeletedID int64  `json:"deleted_id"`
}

type RequestError struct {
	Status  int
	Code    string
	Message string
}

func (e *RequestError) Error() string {
	return e.Message
}

func NewRequestError(status int, code, message string) *RequestError {
	return &RequestError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

func ValidationError(message string) *RequestError {
	return NewRequestError(http.StatusBadRequest, "validation_error", message)
}

func NotFoundError(message string) *RequestError {
	return NewRequestError(http.StatusNotFound, "not_found", message)
}

func ConflictError(message string) *RequestError {
	return NewRequestError(http.StatusConflict, "conflict", message)
}

func MethodNotAllowedError(message string) *RequestError {
	return NewRequestError(http.StatusMethodNotAllowed, "method_not_allowed", message)
}

func WriteJSON(w http.ResponseWriter, status int, payload any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}

func WriteData(w http.ResponseWriter, status int, data any) error {
	return WriteJSON(w, status, DataResponse{Data: data})
}

func WriteCreated(w http.ResponseWriter, message string, data any) error {
	return WriteJSON(w, http.StatusCreated, MutationResponse{
		Message: message,
		Data:    data,
	})
}

func WriteUpdated(w http.ResponseWriter, message string, data any) error {
	return WriteJSON(w, http.StatusOK, MutationResponse{
		Message: message,
		Data:    data,
	})
}

func WriteSuccessMessage(w http.ResponseWriter, status int, message string) error {
	return WriteJSON(w, status, MutationResponse{Message: message})
}

func WriteDeleted(w http.ResponseWriter, message string, deletedID int64) error {
	return WriteJSON(w, http.StatusOK, DeleteResponse{
		Message:   message,
		DeletedID: deletedID,
	})
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	_ = WriteJSON(w, status, ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func WriteRequestError(w http.ResponseWriter, err *RequestError) {
	if err == nil {
		return
	}

	WriteError(w, err.Status, err.Code, err.Message)
}

func BadRequest(w http.ResponseWriter, message string) {
	WriteRequestError(w, ValidationError(message))
}

func NotFound(w http.ResponseWriter, message string) {
	WriteRequestError(w, NotFoundError(message))
}

func Conflict(w http.ResponseWriter, message string) {
	WriteRequestError(w, ConflictError(message))
}

func Internal(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, "internal_error", message)
}

func MethodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	WriteRequestError(w, MethodNotAllowedError("Method is not allowed for this endpoint."))
}

func DecodeJSONBody(r *http.Request, dst any) *RequestError {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return ValidationError("Request body contains badly formed JSON.")
		case errors.Is(err, io.ErrUnexpectedEOF):
			return ValidationError("Request body contains badly formed JSON.")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return ValidationError(fmt.Sprintf("Field %q has an invalid value type.", unmarshalTypeError.Field))
			}
			return ValidationError("Request body contains a value with an invalid type.")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return ValidationError(fmt.Sprintf("Unknown field %s in request body.", fieldName))
		case errors.Is(err, io.EOF):
			return ValidationError("Request body must not be empty.")
		default:
			return ValidationError("Request body must be valid JSON.")
		}
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ValidationError("Request body must contain only one JSON object.")
	}

	return nil
}

func pgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

func IsUniqueViolation(err error) bool {
	return pgErrorCode(err) == "23505"
}

func IsForeignKeyViolation(err error) bool {
	return pgErrorCode(err) == "23503"
}

func IsCheckViolation(err error) bool {
	return pgErrorCode(err) == "23514"
}

func IsNotNullViolation(err error) bool {
	return pgErrorCode(err) == "23502"
}
