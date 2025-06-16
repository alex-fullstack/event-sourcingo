package grpc

import (
	v1 "api/internal/infrastructure/grpc/generated/v1"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	*grpc.ClientConn
	grpc v1.PolicyBackendClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{grpc: v1.NewPolicyBackendClient(conn), ClientConn: conn}, nil
}

func (c *Client) CheckRole(ctx context.Context, userID uuid.UUID, code string) bool {
	_, err := c.grpc.CheckRole(ctx, &v1.RoleCheckRequest{RoleCode: code, UserId: userID.String()})
	return err == nil
}
