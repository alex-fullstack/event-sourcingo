package consumers

import (
	"api/internal/domain/dto"
	"api/internal/domain/usecase"
	"context"
	"log/slog"

	"github.com/google/uuid"
	kafkaGo "github.com/segmentio/kafka-go"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/endpoints"
)

func NewUserConsumer(
	baseContext context.Context,
	cases usecase.UserCases,
	cfg kafkaGo.ReaderConfig,
) endpoints.EndpointStarter {
	return newUpsertConsumer(
		baseContext,
		cfg,
		func(ctx context.Context, message kafkaGo.Message) error {
			id, data, err := convertUserUpsert(message)
			if err != nil {
				slog.Error("User consumer conversation error", ": ", err.Error())
				return err
			}
			err = cases.OnUpsert(ctx, id, data)

			if err != nil {
				slog.Error("User consumer handle error", ": ", err.Error())
				return err
			}
			return nil
		},
	)
}

func convertUserUpsert(message kafkaGo.Message) (uuid.UUID, dto.UserUpsert, error) {
	id, ev, err := convert(message)
	if err != nil {
		return [16]byte{}, dto.UserUpsert{}, err
	}
	user, err := dto.NewUserUpsert(ev.Payload)
	return id, user, err
}
