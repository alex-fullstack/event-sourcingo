package repositories

import "context"

type ProjectionSaver interface {
	Save(ctx context.Context, projection interface{}) error
}
