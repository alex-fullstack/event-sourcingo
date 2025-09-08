package repositories

import "context"

type ProjectionStore[T any] interface {
	Save(ctx context.Context, projection T) error
}
