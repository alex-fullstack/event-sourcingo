package tests

import (
	"errors"

	"github.com/alex-fullstack/event-sourcingo/domain/commands"
	"github.com/alex-fullstack/event-sourcingo/domain/events"
	"github.com/alex-fullstack/event-sourcingo/domain/subscriptions"
	"github.com/alex-fullstack/event-sourcingo/domain/transactions"
	"github.com/alex-fullstack/event-sourcingo/mocks/entities"
	"github.com/alex-fullstack/event-sourcingo/mocks/repositories"
	"github.com/alex-fullstack/event-sourcingo/mocks/services"
	"github.com/google/uuid"
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
