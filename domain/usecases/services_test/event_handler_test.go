package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alex-fullstack/event-sourcingo/domain/events"
	"github.com/alex-fullstack/event-sourcingo/domain/usecases/services"
	"github.com/alex-fullstack/event-sourcingo/mocks/entities"
	"github.com/alex-fullstack/event-sourcingo/mocks/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type EventHandlerTestCase struct {
	description   string
	ctx           context.Context
	newEvents     []events.Event[*struct{}]
	mockAssertion func(tc EventHandlerTestCase)
	dataAssertion func(actual error)
}

func TestEventHandler_HandleMethod(t *testing.T) {
	var (
		aggregateProviderMock *entities.MockAggregateProvider[*struct{}, *struct{}, *struct{}, *struct{}]
		publisherMock         *repositories.MockPublisher[*struct{}]
		errExpected           = errors.New("test error")
		expectedID            = uuid.New()
		expectedEvents        = []events.Event[*struct{}]{
			{AggregateID: expectedID},
			{AggregateID: expectedID},
		}
		expectedIntegrationEvent = events.IntegrationEvent[*struct{}]{}
	)
	testCases := []EventHandlerTestCase{
		{
			description: "Если при вызове метода HandleEvents не удалось опубликовать интеграционные события, то должна вернуться ошибка", //nolint:lll
			ctx:         context.Background(),
			newEvents:   expectedEvents,
			mockAssertion: func(tc EventHandlerTestCase) {
				aggregateProviderMock.EXPECT().ApplyChange(expectedEvents[0]).Return(nil)
				aggregateProviderMock.EXPECT().ApplyChange(expectedEvents[1]).Return(nil)
				aggregateProviderMock.EXPECT().
					IntegrationEvent(0).
					Return(expectedIntegrationEvent)
				aggregateProviderMock.EXPECT().
					IntegrationEvent(0).
					Return(expectedIntegrationEvent)
				publisherMock.EXPECT().
					Publish(
						tc.ctx,
						[]events.IntegrationEvent[*struct{}]{
							expectedIntegrationEvent,
							expectedIntegrationEvent,
						}).
					Return(errExpected)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, errExpected, actual)
			},
		},
		{
			description: "При успешной публикации интеграционных событий метод HandleEvents должнен возвращать пустой результат без ошибки", //nolint:lll
			ctx:         context.Background(),
			newEvents:   expectedEvents,
			mockAssertion: func(tc EventHandlerTestCase) {
				aggregateProviderMock.EXPECT().ApplyChange(expectedEvents[0]).Return(nil)
				aggregateProviderMock.EXPECT().ApplyChange(expectedEvents[1]).Return(nil)
				aggregateProviderMock.EXPECT().
					IntegrationEvent(0).
					Return(expectedIntegrationEvent)
				aggregateProviderMock.EXPECT().
					IntegrationEvent(0).
					Return(expectedIntegrationEvent)
				publisherMock.EXPECT().Publish(
					tc.ctx, []events.IntegrationEvent[*struct{}]{expectedIntegrationEvent, expectedIntegrationEvent}).
					Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.NoError(t, actual)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.description,
			func(t *testing.T) {
				publisherMock = repositories.NewMockPublisher[*struct{}](t)
				aggregateProviderMock = entities.NewMockAggregateProvider[*struct{}, *struct{}, *struct{}, *struct{}](
					t,
				)
				tc.mockAssertion(tc)

				handler := services.NewEventHandler[*struct{}, *struct{}, *struct{}, *struct{}](
					publisherMock,
				)
				err := handler.HandleEvents(
					tc.ctx,
					aggregateProviderMock,
					tc.newEvents,
				)

				if tc.dataAssertion != nil {
					tc.dataAssertion(err)
				}
			})
	}
}
