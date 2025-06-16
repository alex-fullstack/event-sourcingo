package backend

import apiV1 "user/internal/endpoints/api/backend/generated/v1"

type Converter interface {
	ConvertUserIDResponse(id string) *apiV1.UserIDResponse
}

type converter struct{}

func NewConverter() Converter {
	return converter{}
}

func (c converter) ConvertUserIDResponse(id string) *apiV1.UserIDResponse {
	return &apiV1.UserIDResponse{Id: id}
}
