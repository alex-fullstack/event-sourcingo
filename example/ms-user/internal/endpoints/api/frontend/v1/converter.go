package v1

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"
	"user/internal/domain/dto"

	"github.com/google/uuid"
)

type HTTPError struct {
	Error string `json:"error"`
}

type Converter interface {
	ConvertSign(req *http.Request) (dto.UserSign, error)
	ConvertLogin(req *http.Request) (dto.UserLogin, error)
	ConvertRequestConfirm(req *http.Request) (dto.UserRequestConfirmation, error)
	ConvertConfirm(req *http.Request) (dto.ConfirmInput, error)
	WriteAuthResult(w http.ResponseWriter, result dto.AuthResult)
	WriteNoContentResult(w http.ResponseWriter)
	WriteError(w http.ResponseWriter, err error, status int)
}

type header struct {
	userID           string
	userAgent        string
	requestTimestamp time.Time
}

type converter struct {
}

func NewConverter() Converter {
	return &converter{}
}

func (c *converter) ConvertLogin(req *http.Request) (dto.UserLogin, error) {
	credentials, err := c.convertBody(req)
	if err != nil {
		return dto.UserLogin{}, err
	}
	h, err := c.convertHeader(req)
	if err != nil {
		return dto.UserLogin{}, err
	}
	var data = dto.UserLogin{}
	email, err := c.convertEmail(credentials)
	if err != nil {
		return dto.UserLogin{}, err
	}
	password, err := c.convertPassword(credentials)
	if err != nil {
		return dto.UserLogin{}, err
	}
	data.Credentials = dto.AuthInput{Email: email, Password: password}
	data.Activity = dto.ActivityInput{Device: h.userAgent, Timestamp: h.requestTimestamp}
	return data, nil
}

func (c *converter) ConvertSign(req *http.Request) (dto.UserSign, error) {
	credentials, err := c.convertBody(req)
	if err != nil {
		return dto.UserSign{}, err
	}
	h, err := c.convertHeader(req)
	if err != nil {
		return dto.UserSign{}, err
	}
	var data = dto.UserSign{}
	email, err := c.convertEmail(credentials)
	if err != nil {
		return dto.UserSign{}, err
	}
	password, err := c.convertPassword(credentials)
	if err != nil {
		return dto.UserSign{}, err
	}
	data.Credentials = dto.CredentialsCreate{Email: email, Password: password}
	data.Activity = dto.ActivityInput{Device: h.userAgent, Timestamp: h.requestTimestamp}
	return data, nil
}

func (c *converter) ConvertRequestConfirm(req *http.Request) (dto.UserRequestConfirmation, error) {
	h, err := c.convertHeader(req)
	if err != nil {
		return dto.UserRequestConfirmation{}, err
	}
	var data = dto.UserRequestConfirmation{}
	data.Activity = dto.ActivityInput{Device: h.userAgent, Timestamp: h.requestTimestamp}
	data.UserID, err = uuid.Parse(h.userID)
	return data, err
}

func (c *converter) ConvertConfirm(req *http.Request) (dto.ConfirmInput, error) {
	h, err := c.convertHeader(req)
	if err != nil {
		return dto.ConfirmInput{}, err
	}
	confirmation, err := c.convertBody(req)
	if err != nil {
		return dto.ConfirmInput{}, err
	}
	code, err := c.convertConfirmationCode(confirmation)
	if err != nil {
		return dto.ConfirmInput{}, err
	}
	var data = dto.ConfirmInput{Code: code}
	data.Activity = dto.ActivityInput{Device: h.userAgent, Timestamp: h.requestTimestamp}
	data.UserID, err = uuid.Parse(h.userID)
	return data, err
}

func (c *converter) WriteAuthResult(w http.ResponseWriter, result dto.AuthResult) {
	c.write(w, result, http.StatusOK)
}

func (c *converter) WriteNoContentResult(w http.ResponseWriter) {
	c.write(w, nil, http.StatusNoContent)
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

func (c *converter) convertBody(req *http.Request) (map[string]interface{}, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	var payload map[string]interface{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *converter) convertHeader(req *http.Request) (header, error) {
	timestamp, err := time.Parse(time.DateTime, req.Header.Get("X-Request-Timestamp"))
	if err != nil {
		return header{}, err
	}
	return header{
		userID:           req.Header.Get("User-Id"),
		userAgent:        req.Header.Get("User-Agent"),
		requestTimestamp: timestamp,
	}, nil
}

func (c *converter) convertEmail(value map[string]interface{}) (string, error) {
	emailField, ok := value["email"]
	if !ok {
		return "", errors.New("no email field found")
	}
	email, ok := emailField.(string)
	if !ok {
		return "", errors.New("wrong data type, need string")
	}
	return email, nil
}

func (c *converter) convertPassword(value map[string]interface{}) (string, error) {
	passwordField, ok := value["password"]
	if !ok {
		return "", errors.New("no password field found")
	}
	password, ok := passwordField.(string)
	if !ok {
		return "", errors.New("wrong data type, need string")
	}
	return password, nil
}

func (c *converter) convertConfirmationCode(value map[string]interface{}) (string, error) {
	codeField, ok := value["code"]
	if !ok {
		return "", errors.New("no code field found")
	}
	code, ok := codeField.(string)
	if !ok {
		return "", errors.New("wrong data type, need string")
	}
	return code, nil
}
