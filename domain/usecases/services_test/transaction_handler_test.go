package services_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/alex-fullstack/event-sourcingo/domain/entities"
	"github.com/alex-fullstack/event-sourcingo/domain/events"
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
	var (
		eventStoreMock        *repositories.MockEventStore
		aggregateProviderMock *mockEntities.MockAggregateProvider
		eventHandlerMock      *mockServices.MockEventHandler
		errExpected           = errors.New("test error")
		expectedExecutor      = &struct{}{}
		expectedID            = uuid.New()
		expectedEvents        = []events.Event{
			{AggregateID: expectedID},
			{AggregateID: expectedID},
		}
		expectedTransactionID        = uuid.New()
		expectedLastSequenceID int64 = 24
		expectedTransaction          = &transactions.Transaction{
			ID:          expectedTransactionID,
			SequenceID:  expectedLastSequenceID + 1,
			AggregateID: expectedID,
		}
		expectedSubscription = &subscriptions.Subscription{LastSequenceID: expectedLastSequenceID}
	)
	testCases := []TransactionHandlerTestCase{
		{
			description: "Если при вызове метода Handle не удалось открыть транзакцию, то должна вернуться ошибка",
			ctx:         context.Background(),
			mockAssertion: func(tc TransactionHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(nil, errExpected)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, errExpected, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось получить подписки, то должна вернуться ошибка с откатом транзакции", //nolint:lll
			ctx:         context.Background(),
			mockAssertion: func(tc TransactionHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				eventStoreMock.EXPECT().
					GetSubscription(tc.ctx, expectedExecutor).
					Return(nil, errExpected)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, errExpected, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось получить данные агрегата, то должна вернуться ошибка с откатом транзакции", //nolint:lll
			ctx:         context.Background(),
			transaction: expectedTransaction,
			mockAssertion: func(tc TransactionHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				eventStoreMock.EXPECT().
					GetSubscription(tc.ctx, expectedExecutor).
					Return(expectedSubscription, nil)
				eventStoreMock.EXPECT().
					GetNewEventsAndHistory(tc.ctx, expectedID, expectedLastSequenceID, tc.transaction.SequenceID, expectedExecutor).
					Return(nil, nil, errExpected)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, errExpected, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось обработать события агрегата, то должна вернуться ошибка с откатом транзакции", //nolint:lll
			ctx:         context.Background(),
			transaction: expectedTransaction,
			mockAssertion: func(tc TransactionHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				eventStoreMock.EXPECT().
					GetSubscription(tc.ctx, expectedExecutor).
					Return(expectedSubscription, nil)
				eventStoreMock.EXPECT().
					GetNewEventsAndHistory(tc.ctx, expectedID, expectedLastSequenceID, tc.transaction.SequenceID, expectedExecutor).
					Return(expectedEvents, expectedEvents, nil)
				eventHandlerMock.EXPECT().
					HandleEvents(tc.ctx, expectedEvents, expectedEvents, aggregateProviderMock).
					Return(errExpected)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, errExpected, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось сохранить новую подписку, то должна вернуться ошибка с откатом транзакции", //nolint:lll
			ctx:         context.Background(),
			transaction: expectedTransaction,
			mockAssertion: func(tc TransactionHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				eventStoreMock.EXPECT().
					GetSubscription(tc.ctx, expectedExecutor).
					Return(expectedSubscription, nil)
				eventStoreMock.EXPECT().
					GetNewEventsAndHistory(tc.ctx, expectedID, expectedLastSequenceID, tc.transaction.SequenceID, expectedExecutor).
					Return(expectedEvents, expectedEvents, nil)
				eventHandlerMock.EXPECT().
					HandleEvents(tc.ctx, expectedEvents, expectedEvents, aggregateProviderMock).
					Return(nil)
				eventStoreMock.EXPECT().
					UpdateSubscription(
						tc.ctx,
						&subscriptions.Subscription{
							LastSequenceID: tc.transaction.SequenceID,
						}, expectedExecutor).
					Return(errExpected)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, errExpected, actual)
			},
		},
		{
			description: "При успешном вызове метода Handle должнен возвращаться пустой результат без ошибки",
			ctx:         context.Background(),
			transaction: expectedTransaction,
			mockAssertion: func(tc TransactionHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				eventStoreMock.EXPECT().
					GetSubscription(tc.ctx, expectedExecutor).
					Return(expectedSubscription, nil)
				eventStoreMock.EXPECT().
					GetNewEventsAndHistory(
						tc.ctx,
						expectedID,
						expectedLastSequenceID,
						tc.transaction.SequenceID,
						expectedExecutor,
					).
					Return(expectedEvents, expectedEvents, nil)
				eventHandlerMock.EXPECT().
					HandleEvents(tc.ctx, expectedEvents, expectedEvents, aggregateProviderMock).
					Return(nil)
				eventStoreMock.EXPECT().
					UpdateSubscription(
						tc.ctx,
						&subscriptions.Subscription{
							LastSequenceID: tc.transaction.SequenceID,
						}, expectedExecutor).
					Return(nil)
				eventStoreMock.EXPECT().Commit(tc.ctx, expectedExecutor).Return(nil)
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
				aggregateProviderMock = mockEntities.NewMockAggregateProvider(t)
				eventStoreMock = repositories.NewMockEventStore(t)
				eventHandlerMock = mockServices.NewMockEventHandler(t)
				tc.mockAssertion(tc)

				handler := services.NewTransactionHandler(
					eventStoreMock,
					eventHandlerMock,
					slog.Default(),
				)
				err := handler.Handle(
					tc.ctx,
					tc.transaction,
					func(_ uuid.UUID) entities.AggregateProvider { return aggregateProviderMock },
				)

				if tc.dataAssertion != nil {
					tc.dataAssertion(err)
				}
			})
	}
}
