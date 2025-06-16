package tests

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/events"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/usecases/services"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/mocks/entities"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/mocks/repositories"
	"testing"
)

type EventHandlerTestCase struct {
	description   string
	ctx           context.Context
	history       []events.Event
	newEvents     []events.Event
	mockAssertion func(tc EventHandlerTestCase)
	dataAssertion func(actual error)
}

func TestEventHandler_HandleMethod(t *testing.T) {
	testCases := []EventHandlerTestCase{
		{
			description: "Если при вызове метода HandleEvents не удалось собрать агрегат, то должна вернуться ошибка",
			ctx:         context.Background(),
			history:     expectedEvents,
			mockAssertion: func(tc EventHandlerTestCase) {
				aggregateProviderMock.EXPECT().Build(tc.history).Return(expectedError)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "Если при вызове метода HandleEvents не удалось обновить агрегат, то должна вернуться ошибка",
			ctx:         context.Background(),
			history:     expectedEvents,
			newEvents:   expectedEvents,
			mockAssertion: func(tc EventHandlerTestCase) {
				aggregateProviderMock.EXPECT().Build(tc.history).Return(nil)
				aggregateProviderMock.EXPECT().ApplyChanges([]events.Event{tc.newEvents[0]}).Return(expectedError)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "Если при вызове метода HandleEvents не удалось сгенерировать интеграционное событие, то должна вернуться ошибка",
			ctx:         context.Background(),
			history:     expectedEvents,
			newEvents:   expectedEvents,
			mockAssertion: func(tc EventHandlerTestCase) {
				aggregateProviderMock.EXPECT().Build(tc.history).Return(nil)
				aggregateProviderMock.EXPECT().ApplyChanges([]events.Event{tc.newEvents[0]}).Return(nil)
				aggregateProviderMock.EXPECT().ApplyChanges([]events.Event{tc.newEvents[1]}).Return(nil)
				aggregateProviderMock.EXPECT().IntegrationEvent(0).Return(events.IntegrationEvent{}, false, expectedError)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "Если при вызове метода HandleEvents не удалось опубликовать интеграционные события, то должна вернуться ошибка",
			ctx:         context.Background(),
			history:     expectedEvents,
			newEvents:   expectedEvents,
			mockAssertion: func(tc EventHandlerTestCase) {
				aggregateProviderMock.EXPECT().Build(tc.history).Return(nil)
				aggregateProviderMock.EXPECT().ApplyChanges([]events.Event{tc.newEvents[0]}).Return(nil)
				aggregateProviderMock.EXPECT().IntegrationEvent(0).Return(expectedIntegrationEvent, true, nil)
				aggregateProviderMock.EXPECT().IntegrationEvent(0).Return(expectedIntegrationEvent, true, nil)
				aggregateProviderMock.EXPECT().ApplyChanges([]events.Event{tc.newEvents[1]}).Return(nil)
				publisherMock.EXPECT().Publish(tc.ctx, []events.IntegrationEvent{expectedIntegrationEvent, expectedIntegrationEvent}).Return(expectedError)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "При успешной публикации интеграционных событий метод HandleEvents должнен возвращать пустой результат без ошибки",
			ctx:         context.Background(),
			history:     expectedEvents,
			newEvents:   expectedEvents,
			mockAssertion: func(tc EventHandlerTestCase) {
				aggregateProviderMock.EXPECT().Build(tc.history).Return(nil)
				aggregateProviderMock.EXPECT().ApplyChanges([]events.Event{tc.newEvents[0]}).Return(nil)
				aggregateProviderMock.EXPECT().IntegrationEvent(0).Return(expectedIntegrationEvent, true, nil)
				aggregateProviderMock.EXPECT().IntegrationEvent(0).Return(expectedIntegrationEvent, true, nil)
				aggregateProviderMock.EXPECT().ApplyChanges([]events.Event{tc.newEvents[1]}).Return(nil)
				publisherMock.EXPECT().Publish(tc.ctx, []events.IntegrationEvent{expectedIntegrationEvent, expectedIntegrationEvent}).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, nil, actual)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.description,
			func(t *testing.T) {
				publisherMock = repositories.NewMockPublisher(t)
				aggregateProviderMock = entities.NewMockAggregateProvider(t)
				tc.mockAssertion(tc)

				handler := services.NewEventHandler(publisherMock)
				err := handler.HandleEvents(tc.ctx, tc.history, tc.newEvents, aggregateProviderMock)

				if tc.dataAssertion != nil {
					tc.dataAssertion(err)
				}
			})
	}

}
