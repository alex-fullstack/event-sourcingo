package tests

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/commands"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/usecases/services"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/mocks/entities"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/mocks/repositories"
	"testing"
)

type CommandHandlerTestCase struct {
	description   string
	mockAssertion func(tc CommandHandlerTestCase)
	dataAssertion func(actual error)
	ctx           context.Context
	cmd           commands.Command
}

func TestCommandHandler_HandleMethod(t *testing.T) {
	testCases := []CommandHandlerTestCase{
		{
			description: "Если при вызове метода Handle не удалось открыть транзакцию, то должна вернуться ошибка",
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(nil, expectedError)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
			ctx: context.Background(),
		},
		{
			description: "Если при вызове метода Handle не удалось получить список событий агрегата, то должна вернуться ошибка с откатом транзакции",
			ctx:         context.Background(),
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				aggregateProviderMock.EXPECT().ID().Return(expectedId)
				eventStoreMock.EXPECT().GetAggregateEvents(tc.ctx, expectedId, expectedExecutor).Return(nil, expectedError)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось построить агрегат, то должна вернуться ошибка с откатом транзакции",
			ctx:         context.Background(),
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				aggregateProviderMock.EXPECT().ID().Return(expectedId)
				eventStoreMock.EXPECT().GetAggregateEvents(tc.ctx, expectedId, expectedExecutor).Return(expectedEvents, nil)
				aggregateProviderMock.EXPECT().Build(expectedEvents).Return(expectedError)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось обновить агрегат, то должна вернуться ошибка с откатом транзакции",
			ctx:         context.Background(),
			cmd:         expectedCommand,
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				aggregateProviderMock.EXPECT().ID().Return(expectedId).Times(3)
				eventStoreMock.EXPECT().GetAggregateEvents(tc.ctx, expectedId, expectedExecutor).Return(expectedEvents, nil)
				aggregateProviderMock.EXPECT().Build(expectedEvents).Return(nil)
				aggregateProviderMock.EXPECT().Version().Return(0)
				aggregateProviderMock.EXPECT().ApplyChanges(mock.Anything).Return(expectedError)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось сохранить агрегат, то должна вернуться ошибка с откатом транзакции",
			ctx:         context.Background(),
			cmd:         expectedCommand,
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				aggregateProviderMock.EXPECT().ID().Return(expectedId).Times(3)
				eventStoreMock.EXPECT().GetAggregateEvents(tc.ctx, expectedId, expectedExecutor).Return(expectedEvents, nil)
				aggregateProviderMock.EXPECT().Build(expectedEvents).Return(nil)
				aggregateProviderMock.EXPECT().Version().Return(0)
				aggregateProviderMock.EXPECT().ApplyChanges(mock.Anything).Return(nil)
				eventStoreMock.EXPECT().UpdateOrCreateAggregate(tc.ctx, mock.Anything, aggregateProviderMock, expectedExecutor).Return(expectedError)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось сохранить проекцию агрегата, то должна вернуться ошибка с откатом транзакции",
			ctx:         context.Background(),
			cmd:         expectedCommand,
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				aggregateProviderMock.EXPECT().ID().Return(expectedId).Times(3)
				eventStoreMock.EXPECT().GetAggregateEvents(tc.ctx, expectedId, expectedExecutor).Return(expectedEvents, nil)
				aggregateProviderMock.EXPECT().Build(expectedEvents).Return(nil)
				aggregateProviderMock.EXPECT().Version().Return(0)
				aggregateProviderMock.EXPECT().ApplyChanges(mock.Anything).Return(nil)
				eventStoreMock.EXPECT().UpdateOrCreateAggregate(tc.ctx, mock.Anything, aggregateProviderMock, expectedExecutor).Return(nil)
				aggregateProviderMock.EXPECT().Projection().Return(expectedProjection)
				saverMock.EXPECT().Save(tc.ctx, expectedProjection).Return(expectedError)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось выполнить транзакцию, то должна вернуться ошибка",
			ctx:         context.Background(),
			cmd:         expectedCommand,
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				aggregateProviderMock.EXPECT().ID().Return(expectedId).Times(3)
				eventStoreMock.EXPECT().GetAggregateEvents(tc.ctx, expectedId, expectedExecutor).Return(expectedEvents, nil)
				aggregateProviderMock.EXPECT().Build(expectedEvents).Return(nil)
				aggregateProviderMock.EXPECT().Version().Return(0)
				aggregateProviderMock.EXPECT().ApplyChanges(mock.Anything).Return(nil)
				eventStoreMock.EXPECT().UpdateOrCreateAggregate(tc.ctx, mock.Anything, aggregateProviderMock, expectedExecutor).Return(nil)
				aggregateProviderMock.EXPECT().Projection().Return(expectedProjection)
				saverMock.EXPECT().Save(tc.ctx, expectedProjection).Return(nil)
				eventStoreMock.EXPECT().Commit(tc.ctx, expectedExecutor).Return(expectedError)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "При успешном выполнении транзакция метод Handle должнен возвращать пустой результат без ошибки",
			ctx:         context.Background(),
			cmd:         expectedCommand,
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				aggregateProviderMock.EXPECT().ID().Return(expectedId).Times(3)
				eventStoreMock.EXPECT().GetAggregateEvents(tc.ctx, expectedId, expectedExecutor).Return(expectedEvents, nil)
				aggregateProviderMock.EXPECT().Build(expectedEvents).Return(nil)
				aggregateProviderMock.EXPECT().Version().Return(0)
				aggregateProviderMock.EXPECT().ApplyChanges(mock.Anything).Return(nil)
				eventStoreMock.EXPECT().UpdateOrCreateAggregate(tc.ctx, mock.Anything, aggregateProviderMock, expectedExecutor).Return(nil)
				aggregateProviderMock.EXPECT().Projection().Return(expectedProjection)
				saverMock.EXPECT().Save(tc.ctx, expectedProjection).Return(nil)
				eventStoreMock.EXPECT().Commit(tc.ctx, expectedExecutor).Return(nil)
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
				eventStoreMock = repositories.NewMockEventStore(t)
				saverMock = repositories.NewMockProjectionSaver(t)
				aggregateProviderMock = entities.NewMockAggregateProvider(t)
				tc.mockAssertion(tc)

				handler := services.NewCommandHandler(eventStoreMock, saverMock)
				err := handler.Handle(tc.ctx, tc.cmd, aggregateProviderMock)

				if tc.dataAssertion != nil {
					tc.dataAssertion(err)
				}
			})
	}

}
