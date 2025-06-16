package backend

import (
	"context"
	"policy/internal/domain/usecase"
	v1 "policy/internal/endpoints/api/backend/generated/v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	v1.UnimplementedPolicyBackendServer
	converter Converter
	cases     usecase.BackendAPICases
}

func New(converter Converter, cases usecase.BackendAPICases) v1.PolicyBackendServer {
	return &server{converter: converter, cases: cases}
}

func (s *server) CheckRole(ctx context.Context, data *v1.RoleCheckRequest) (*emptypb.Empty, error) {
	dto, err := s.converter.ConvertRoleCheck(data)
	if err != nil {
		return nil, err
	}

	err = s.cases.CheckRole(ctx, dto)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *server) CheckPermission(ctx context.Context, data *v1.PermissionCheckRequest) (*emptypb.Empty, error) {
	dto, err := s.converter.ConvertPermissionCheck(data)
	if err != nil {
		return nil, err
	}

	err = s.cases.CheckPermission(ctx, dto)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
