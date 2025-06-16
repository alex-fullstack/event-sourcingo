package backend

import (
	"context"
	"user/internal/domain/usecase"
	v1 "user/internal/endpoints/api/backend/generated/v1"
)

type server struct {
	v1.UnimplementedUserBackendServer
	converter Converter
	cases     usecase.BackendAPICases
}

func New(converter Converter, cases usecase.BackendAPICases) v1.UserBackendServer {
	return &server{converter: converter, cases: cases}
}

func (s *server) GetUserIDByEmail(ctx context.Context, req *v1.UserIDRequest) (*v1.UserIDResponse, error) {
	id, err := s.cases.GetUserIDByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, err
	}
	return s.converter.ConvertUserIDResponse(id), nil
}
