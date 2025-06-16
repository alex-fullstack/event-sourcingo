package v1

import (
	"api/internal/domain/dto"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type HTTPError struct {
	Error string `json:"error"`
}

type Converter interface {
	ConvertUsers(req *http.Request) (uuid.UUID, error)
	WriteUsersResult(w http.ResponseWriter, result []dto.UserDocument)
	WriteError(w http.ResponseWriter, err error, status int)
}

type header struct {
	userID string
}

type converter struct {
}

func NewConverter() Converter {
	return &converter{}
}

func (c *converter) ConvertUsers(req *http.Request) (uuid.UUID, error) {
	h := c.convertHeader(req)
	return uuid.Parse(h.userID)
}

func (c *converter) WriteUsersResult(w http.ResponseWriter, result []dto.UserDocument) {
	c.write(w, result, http.StatusOK)
}

func (c *converter) WriteError(w http.ResponseWriter, err error, statusCode int) {
	c.write(w, HTTPError{Error: err.Error()}, statusCode)
}

func (c *converter) write(w http.ResponseWriter, result interface{}, statusCode int) {
	data, _ := json.Marshal(result)
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(statusCode)
	_, _ = w.Write(data)
}

func (c *converter) convertHeader(req *http.Request) header {
	return header{userID: req.Header.Get("User-Id")}
}
