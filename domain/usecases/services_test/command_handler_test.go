package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alex-fullstack/event-sourcingo/domain/commands"
	"github.com/alex-fullstack/event-sourcingo/domain/events"
	"github.com/alex-fullstack/event-sourcingo/domain/usecases/services"
	"github.com/alex-fullstack/event-sourcingo/mocks/entities"
	"github.com/alex-fullstack/event-sourcingo/mocks/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type CommandHandlerTestCase struct {
	description   string
	mockAssertion func(tc CommandHandlerTestCase)
	dataAssertion func(actual error)
	ctx           context.Context
	cmd           commands.Command[*struct{}]
}

func TestCommandHandler_HandleMethod(t *testing.T) {
	var (
		eventStoreMock        *repositories.MockEventStore[*struct{}, *struct{}, *struct{}]
		saverMock             *repositories.MockProjectionStore[*struct{}]
		aggregateProviderMock *entities.MockAggregateProvider[*struct{}, *struct{}, *struct{}, *struct{}]
		errExpected           = errors.New("test error")
		expectedExecutor      = &struct{}{}
		expectedID            = uuid.New()
		expectedEvents        = []events.Event[*struct{}]{
			{AggregateID: expectedID},
			{AggregateID: expectedID},
		}
		expectedCommandEvent = commands.CommandEvent[*struct{}]{
			Type:    1,
			Payload: &struct{}{},
		}
		expectedCommand = commands.Command[*struct{}]{
			Events: []commands.CommandEvent[*struct{}]{
				expectedCommandEvent,
				expectedCommandEvent,
			},
		}
		expectedSnapshot   = &struct{}{}
		expectedProjection = &struct{}{}
	)
	testCases := []CommandHandlerTestCase{
		{
			description: "Если при вызове метода Handle не удалось открыть транзакцию, то должна вернуться ошибка",
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(nil, errExpected)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, errExpected, actual)
			},
			ctx: context.Background(),
		},
		{
			description: "Если при вызове метода Handle не удалось получить список событий агрегата, то должна вернуться ошибка с откатом транзакции", //nolint:lll
			ctx:         context.Background(),
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				aggregateProviderMock.EXPECT().ID().Return(expectedID)
				eventStoreMock.EXPECT().
					GetSnapshot(tc.ctx, expectedID, func() *int { return nil }(), expectedExecutor).
					Return(0, nil, nil)
				eventStoreMock.EXPECT().
					GetEvents(tc.ctx, expectedID, 0, func() *int { return nil }(), expectedExecutor).
					Return(nil, errExpected)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, errExpected, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось построить агрегат, то должна вернуться ошибка с откатом транзакции", //nolint:lll
			ctx:         context.Background(),
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				aggregateProviderMock.EXPECT().ID().Return(expectedID)
				eventStoreMock.EXPECT().
					GetSnapshot(tc.ctx, expectedID, func() *int { return nil }(), expectedExecutor).
					Return(0, nil, nil)
				eventStoreMock.EXPECT().
					GetEvents(tc.ctx, expectedID, 0, func() *int { return nil }(), expectedExecutor).
					Return(expectedEvents, nil)
				aggregateProviderMock.EXPECT().Build(expectedEvents).Return(errExpected)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, errExpected, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось обновить агрегат, то должна вернуться ошибка с откатом транзакции", //nolint:lll
			ctx:         context.Background(),
			cmd:         expectedCommand,
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				aggregateProviderMock.EXPECT().ID().Return(expectedID).Times(4)
				eventStoreMock.EXPECT().
					GetSnapshot(tc.ctx, expectedID, func() *int { return nil }(), expectedExecutor).
					Return(0, nil, nil)
				eventStoreMock.EXPECT().
					GetEvents(tc.ctx, expectedID, 0, func() *int { return nil }(), expectedExecutor).
					Return(expectedEvents, nil)
				aggregateProviderMock.EXPECT().Build(expectedEvents).Return(nil)
				aggregateProviderMock.EXPECT().Version().Return(0)
				aggregateProviderMock.EXPECT().ApplyChanges(mock.Anything).Return(errExpected)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, errExpected, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось сохранить агрегат, то должна вернуться ошибка с откатом транзакции", //nolint:lll
			ctx:         context.Background(),
			cmd:         expectedCommand,
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				aggregateProviderMock.EXPECT().ID().Return(expectedID).Times(4)
				eventStoreMock.EXPECT().
					GetSnapshot(tc.ctx, expectedID, func() *int { return nil }(), expectedExecutor).
					Return(0, nil, nil)
				eventStoreMock.EXPECT().
					GetEvents(tc.ctx, expectedID, 0, func() *int { return nil }(), expectedExecutor).
					Return(expectedEvents, nil)
				aggregateProviderMock.EXPECT().Build(expectedEvents).Return(nil)
				aggregateProviderMock.EXPECT().Version().Return(0)
				aggregateProviderMock.EXPECT().ApplyChanges(mock.Anything).Return(nil)
				aggregateProviderMock.EXPECT().Snapshot().Return(nil)
				eventStoreMock.EXPECT().
					UpdateOrCreateAggregate(
						tc.ctx,
						mock.Anything,
						aggregateProviderMock,
						mock.Anything,
						expectedExecutor,
					).
					Return(errExpected)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, errExpected, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось сохранить проекцию агрегата, то должна вернуться ошибка с откатом транзакции", //nolint:lll
			ctx:         context.Background(),
			cmd:         expectedCommand,
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				aggregateProviderMock.EXPECT().ID().Return(expectedID).Times(4)
				eventStoreMock.EXPECT().
					GetSnapshot(tc.ctx, expectedID, func() *int { return nil }(), expectedExecutor).
					Return(0, nil, nil)
				eventStoreMock.EXPECT().
					GetEvents(tc.ctx, expectedID, 0, func() *int { return nil }(), expectedExecutor).
					Return(expectedEvents, nil)
				aggregateProviderMock.EXPECT().Build(expectedEvents).Return(nil)
				aggregateProviderMock.EXPECT().Version().Return(0)
				aggregateProviderMock.EXPECT().ApplyChanges(mock.Anything).Return(nil)
				eventStoreMock.EXPECT().
					UpdateOrCreateAggregate(
						tc.ctx,
						mock.Anything,
						aggregateProviderMock,
						mock.Anything,
						expectedExecutor,
					).
					Return(nil)
				aggregateProviderMock.EXPECT().Snapshot().Return(expectedSnapshot)
				aggregateProviderMock.EXPECT().Projection().Return(expectedProjection)
				saverMock.EXPECT().Save(tc.ctx, expectedProjection).Return(errExpected)
				eventStoreMock.EXPECT().Rollback(tc.ctx, expectedExecutor).Return(nil)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, errExpected, actual)
			},
		},
		{
			description: "Если при вызове метода Handle не удалось выполнить транзакцию, то должна вернуться ошибка",
			ctx:         context.Background(),
			cmd:         expectedCommand,
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				aggregateProviderMock.EXPECT().ID().Return(expectedID).Times(4)
				eventStoreMock.EXPECT().
					GetSnapshot(tc.ctx, expectedID, func() *int { return nil }(), expectedExecutor).
					Return(0, nil, nil)
				eventStoreMock.EXPECT().
					GetEvents(tc.ctx, expectedID, 0, func() *int { return nil }(), expectedExecutor).
					Return(expectedEvents, nil)
				aggregateProviderMock.EXPECT().Build(expectedEvents).Return(nil)
				aggregateProviderMock.EXPECT().Version().Return(0)
				aggregateProviderMock.EXPECT().ApplyChanges(mock.Anything).Return(nil)
				eventStoreMock.EXPECT().
					UpdateOrCreateAggregate(
						tc.ctx,
						mock.Anything,
						aggregateProviderMock,
						mock.Anything,
						expectedExecutor,
					).
					Return(nil)
				aggregateProviderMock.EXPECT().Snapshot().Return(expectedSnapshot)
				aggregateProviderMock.EXPECT().Projection().Return(expectedProjection)
				saverMock.EXPECT().Save(tc.ctx, expectedProjection).Return(nil)
				eventStoreMock.EXPECT().Commit(tc.ctx, expectedExecutor).Return(errExpected)
			},
			dataAssertion: func(actual error) {
				assert.Equal(t, errExpected, actual)
			},
		},
		{
			description: "При успешном выполнении транзакция метод Handle должнен возвращать пустой результат без ошибки",
			ctx:         context.Background(),
			cmd:         expectedCommand,
			mockAssertion: func(tc CommandHandlerTestCase) {
				eventStoreMock.EXPECT().Begin(tc.ctx).Return(expectedExecutor, nil)
				aggregateProviderMock.EXPECT().ID().Return(expectedID).Times(4)
				eventStoreMock.EXPECT().
					GetSnapshot(tc.ctx, expectedID, func() *int { return nil }(), expectedExecutor).
					Return(0, nil, nil)
				eventStoreMock.EXPECT().
					GetEvents(tc.ctx, expectedID, 0, func() *int { return nil }(), expectedExecutor).
					Return(expectedEvents, nil)
				aggregateProviderMock.EXPECT().Build(expectedEvents).Return(nil)
				aggregateProviderMock.EXPECT().Version().Return(0)
				aggregateProviderMock.EXPECT().ApplyChanges(mock.Anything).Return(nil)
				eventStoreMock.EXPECT().UpdateOrCreateAggregate(
					tc.ctx,
					mock.Anything,
					aggregateProviderMock,
					mock.Anything,
					expectedExecutor,
				).Return(nil)
				aggregateProviderMock.EXPECT().Snapshot().Return(expectedSnapshot)
				aggregateProviderMock.EXPECT().Projection().Return(expectedProjection)
				saverMock.EXPECT().Save(tc.ctx, expectedProjection).Return(nil)
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
				eventStoreMock = repositories.NewMockEventStore[*struct{}, *struct{}, *struct{}](
					t,
				)
				saverMock = repositories.NewMockProjectionStore[*struct{}](t)
				aggregateProviderMock = entities.NewMockAggregateProvider[*struct{}, *struct{}, *struct{}, *struct{}](
					t,
				)
				tc.mockAssertion(tc)

				handler := services.NewCommandHandler[*struct{}, *struct{}, *struct{}, *struct{}, *struct{}](
					eventStoreMock,
					saverMock,
				)
				err := handler.Handle(tc.ctx, tc.cmd, aggregateProviderMock)

				if tc.dataAssertion != nil {
					tc.dataAssertion(err)
				}
			})
	}
}
