package consumers

import (
	"context"
	"encoding/json"
	"log/slog"
	"policy/internal/domain/dto"
	"policy/internal/domain/usecase"
	"policy/internal/infrastructure/kafka"

	"github.com/alex-fullstack/event-sourcingo/domain/events"
	"github.com/alex-fullstack/event-sourcingo/endpoints"
	"github.com/google/uuid"
	kafkaGo "github.com/segmentio/kafka-go"
)

type consumer struct {
	*endpoints.Endpoint
}

func NewMSUserConsumer(
	baseContext context.Context,
	cases usecase.MSUserCases,
	cfg kafkaGo.ReaderConfig,
	log *slog.Logger,
) endpoints.EndpointStarter {
	reader := kafka.NewReader(
		cfg,
		func(ctx context.Context, message kafkaGo.Message) error {
			userID, evType, err := convert(message)
			if err != nil {
				slog.Error("MSUser consumer conversation error", ": ", err.Error())
				return err
			}
			switch evType {
			case dto.CredentialsCreated.Index():
				err = cases.OnUserCreate(ctx, userID)
			case
				dto.ConfirmationCreated.Index(),
				dto.UserSigned.Index(),
				dto.UserAuthenticated.Index(),
				dto.ConfirmationUpdated.Index(),
				dto.EmailConfirmed.Index(),
				dto.UserConfirmationRequested.Index(),
				dto.UserConfirmationRejected.Index(),
				dto.UserConfirmationCompleted.Index():
				return nil
			}

			if err != nil {
				slog.Error("MSUser consumer handle error", ": ", err.Error())
				return err
			}
			return nil
		},
		func() context.Context {
			return baseContext
		},
	)
	return &consumer{
		Endpoint: endpoints.NewEndpoint(
			reader.StartListen,
			reader.Shutdown,
			log,
		),
	}
}

func convert(message kafkaGo.Message) (uuid.UUID, int, error) {
	userID, err := uuid.FromBytes(message.Key)
	if err != nil {
		return [16]byte{}, -1, err
	}
	var msg events.IntegrationEvent
	err = json.Unmarshal(message.Value, &msg)
	return userID, msg.Type, err
}
