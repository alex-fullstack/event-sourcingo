package v1_test

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
	"user/internal/domain/dto"
	apiV1 "user/internal/endpoints/api/frontend/v1"
	mock "user/mocks/http"
	"user/mocks/usecase"
	apiV1Mock "user/mocks/v1"

	mock2 "github.com/stretchr/testify/mock"
)

type testAssertion func(tc TestCase)

type TestCase struct {
	description   string
	request       *http.Request
	mockAssertion testAssertion
	dataAssertion testAssertion
}

func TestRouter(t *testing.T) {
	var (
		apiCasesMock  *usecase.MockFrontendAPICases
		converterMock *apiV1Mock.MockConverter
		writerMock    *mock.MockResponseWriter
		expectedErr   = errors.New("test error")
	)
	testCases := []TestCase{
		{
			description: "Пустой запрос должен возвращать статус 405",
			request:     &http.Request{URL: &url.URL{}},
			mockAssertion: func(_ TestCase) {
				writerMock.EXPECT().WriteHeader(http.StatusMethodNotAllowed)
				writerMock.EXPECT().Write([]byte(nil)).Return(0, nil)
			},
		},
		{
			description: "Произвольный POST запрос должен возвращать статус 404",
			request: &http.Request{
				Method: http.MethodPost,
				URL:    &url.URL{Path: ""},
				Header: map[string][]string{},
			},
			mockAssertion: func(_ TestCase) {
				writerMock.EXPECT().Header().Return(map[string][]string{})
				writerMock.EXPECT().WriteHeader(http.StatusNotFound)
				writerMock.EXPECT().Write([]byte("404 page not found\n")).Return(0, nil)
			},
		},
		{
			description: "При ошибке конвертации запрос авторизации должен возвращать статус 400",
			request: &http.Request{
				Method: http.MethodPost,
				URL:    &url.URL{Path: "/login"},
				Header: map[string][]string{},
			},
			mockAssertion: func(_ TestCase) {
				converterMock.EXPECT().
					ConvertLogin(mock2.Anything).
					Return(dto.UserLogin{}, expectedErr)
				converterMock.EXPECT().WriteError(writerMock, expectedErr, http.StatusBadRequest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.description,
			func(_ *testing.T) {
				apiCasesMock = usecase.NewMockFrontendAPICases(t)
				converterMock = apiV1Mock.NewMockConverter(t)
				writerMock = mock.NewMockResponseWriter(t)
				tc.mockAssertion(tc)

				router := apiV1.New(apiCasesMock, converterMock)
				router.ServeHTTP(writerMock, tc.request)

				if tc.dataAssertion != nil {
					tc.dataAssertion(tc)
				}
			})
	}
}
