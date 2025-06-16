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

func NewPolicyConsumer(
	baseContext context.Context,
	cases usecase.PolicyCases,
	cfg kafkaGo.ReaderConfig,
) endpoints.EndpointStarter {
	return newUpsertConsumer(
		baseContext,
		cfg,
		func(ctx context.Context, message kafkaGo.Message) error {
			id, data, err := convertPolicyUpsert(message)
			if err != nil {
				slog.Error("Policy consumer conversation error", ": ", err.Error())
				return err
			}
			err = cases.OnUpsert(ctx, id, data)

			if err != nil {
				slog.Error("Policy consumer handle error", ": ", err.Error())
				return err
			}
			return nil
		},
	)
}

func convertPolicyUpsert(message kafkaGo.Message) (uuid.UUID, dto.PolicyUpsert, error) {
	id, ev, err := convert(message)
	if err != nil {
		return [16]byte{}, dto.PolicyUpsert{}, err
	}
	policy, err := dto.NewPolicyUpsert(ev.Payload)
	return id, policy, err
}
