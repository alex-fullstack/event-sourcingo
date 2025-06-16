package tests

import (
	"errors"
	"github.com/google/uuid"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/commands"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/events"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/subscriptions"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/transactions"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/mocks/entities"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/mocks/repositories"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/mocks/services"
)

var (
	eventStoreMock           *repositories.MockEventStore
	saverMock                *repositories.MockProjectionSaver
	aggregateProviderMock    *entities.MockAggregateProvider
	publisherMock            *repositories.MockPublisher
	eventHandlerMock         *services.MockEventHandler
	expectedError                  = errors.New("test error")
	expectedExecutor               = &struct{}{}
	expectedId                     = uuid.New()
	expectedEvents                 = []events.Event{{AggregateId: expectedId}, {AggregateId: expectedId}}
	expectedCommandEvent           = commands.CommandEvent{Type: 1, Payload: &struct{}{}}
	expectedCommand                = commands.Command{Events: []commands.CommandEvent{expectedCommandEvent, expectedCommandEvent}}
	expectedProjection             = &struct{}{}
	expectedIntegrationEvent       = events.IntegrationEvent{}
	expectedTransactionId          = uuid.New()
	expectedLastSequenceID   int64 = 24
	expectedTransaction            = &transactions.Transaction{
		Id:          expectedTransactionId,
		SequenceId:  expectedLastSequenceID + 1,
		AggregateId: expectedId,
	}
	expectedSubscription = &subscriptions.Subscription{LastSequenceID: expectedLastSequenceID}
)
