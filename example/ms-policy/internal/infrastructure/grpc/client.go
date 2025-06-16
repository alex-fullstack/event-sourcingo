package grpc

import (
	"context"
	v1 "policy/internal/infrastructure/grpc/generated/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	*grpc.ClientConn
	grpc v1.UserBackendClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{grpc: v1.NewUserBackendClient(conn), ClientConn: conn}, nil
}

func (c *Client) GetUserIDByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	resp, err := c.grpc.GetUserIDByEmail(ctx, &v1.UserIDRequest{Email: email})
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(resp.GetId())
}
