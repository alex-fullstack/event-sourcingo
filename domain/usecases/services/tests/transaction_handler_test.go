package tests

import (
	"context"
	"testing"

	"github.com/alex-fullstack/event-sourcingo/domain/entities"
	"github.com/alex-fullstack/event-sourcingo/domain/subscriptions"
	"github.com/alex-fullstack/event-sourcingo/domain/transactions"
	"github.com/alex-fullstack/event-sourcingo/domain/usecases/services"
	mockEntities "github.com/alex-fullstack/event-sourcingo/mocks/entities"
	"github.com/alex-fullstack/event-sourcingo/mocks/repositories"
	mockServices "github.com/alex-fullstack/event-sourcingo/mocks/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type TransactionHandlerTestCase struct {
	description   string
	ctx           context.Context
	transaction   *transactions.Transaction
	mockAssertion func(tc TransactionHandlerTestCase)
	dataAssertion func(actual error)
}

func TestTransactionHandler_HandleMethod(t *testing.T) {
	testCases := []TransactionHandlerTestCase{
		{
			description: "Если при вызове метода Handle не удалось открыть транзакцию, то должна вернуться ошибка",
			ctx:         context.Background(),
			mockAssertion: func(tc TransactionHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(nil, expectedError)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось получить подписки, то должна вернуться ошибка с откатом транзакции",
			ctx:         context.Background(),
			mockAssertion: func(tc TransactionHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				eventStoreMock.EXPECT().GetSubscription(tc.ctx, expectedExecutor).Return(nil, expectedError)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось получить данные агрегата, то должна вернуться ошибка с откатом транзакции",
			ctx:         context.Background(),
			transaction: expectedTransaction,
			mockAssertion: func(tc TransactionHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				eventStoreMock.EXPECT().GetSubscription(tc.ctx, expectedExecutor).Return(expectedSubscription, nil)
				eventStoreMock.EXPECT().GetNewEventsAndHistory(tc.ctx, expectedId, expectedLastSequenceID, tc.transaction.SequenceId, expectedExecutor).Return(nil, nil, expectedError)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось обработать события агрегата, то должна вернуться ошибка с откатом транзакции",
			ctx:         context.Background(),
			transaction: expectedTransaction,
			mockAssertion: func(tc TransactionHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				eventStoreMock.EXPECT().GetSubscription(tc.ctx, expectedExecutor).Return(expectedSubscription, nil)
				eventStoreMock.EXPECT().GetNewEventsAndHistory(tc.ctx, expectedId, expectedLastSequenceID, tc.transaction.SequenceId, expectedExecutor).Return(expectedEvents, expectedEvents, nil)
				eventHandlerMock.EXPECT().HandleEvents(tc.ctx, expectedEvents, expectedEvents, aggregateProviderMock).Return(expectedError)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось сохранить новую подписку, то должна вернуться ошибка с откатом транзакции",
			ctx:         context.Background(),
			transaction: expectedTransaction,
			mockAssertion: func(tc TransactionHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				eventStoreMock.EXPECT().GetSubscription(tc.ctx, expectedExecutor).Return(expectedSubscription, nil)
				eventStoreMock.EXPECT().GetNewEventsAndHistory(tc.ctx, expectedId, expectedLastSequenceID, tc.transaction.SequenceId, expectedExecutor).Return(expectedEvents, expectedEvents, nil)
				eventHandlerMock.EXPECT().HandleEvents(tc.ctx, expectedEvents, expectedEvents, aggregateProviderMock).Return(nil)
				eventStoreMock.EXPECT().UpdateSubscription(tc.ctx, &subscriptions.Subscription{LastSequenceID: tc.transaction.SequenceId}, expectedExecutor).Return(expectedError)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, expectedError, actual)
			},
		},
		{
			description: "При успешном вызове метода Handle должнен возвращаться пустой результат без ошибки",
			ctx:         context.Background(),
			transaction: expectedTransaction,
			mockAssertion: func(tc TransactionHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				eventStoreMock.EXPECT().GetSubscription(tc.ctx, expectedExecutor).Return(expectedSubscription, nil)
				eventStoreMock.EXPECT().GetNewEventsAndHistory(tc.ctx, expectedId, expectedLastSequenceID, tc.transaction.SequenceId, expectedExecutor).Return(expectedEvents, expectedEvents, nil)
				eventHandlerMock.EXPECT().HandleEvents(tc.ctx, expectedEvents, expectedEvents, aggregateProviderMock).Return(nil)
				eventStoreMock.EXPECT().UpdateSubscription(tc.ctx, &subscriptions.Subscription{LastSequenceID: tc.transaction.SequenceId}, expectedExecutor).Return(nil)
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
				aggregateProviderMock = mockEntities.NewMockAggregateProvider(t)
				eventStoreMock = repositories.NewMockEventStore(t)
				eventHandlerMock = mockServices.NewMockEventHandler(t)
				tc.mockAssertion(tc)

				handler := services.NewTransactionHandler(eventStoreMock, eventHandlerMock)
				err := handler.Handle(tc.ctx, tc.transaction, func(uuid uuid.UUID) entities.AggregateProvider { return aggregateProviderMock })

				if tc.dataAssertion != nil {
					tc.dataAssertion(err)
				}
			})
	}

}
