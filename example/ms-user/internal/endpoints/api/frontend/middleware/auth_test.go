package middleware_test

import (
	"net/http"
	"testing"
	"time"
	"user/internal/endpoints/api/frontend/middleware"
	mock "user/mocks/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	accessTokenSecret = "jdjenc746cjeu37jd3374"
	userID            = "123"
)

type testAssertion func(tc TestCase)

type TestCase struct {
	description   string
	request       *http.Request
	mockAssertion testAssertion
	dataAssertion testAssertion
}

func TestAuthMiddleware(t *testing.T) {
	var (
		wrongAccessToken = "wrongAccessToken"
		accessTokenFn    = func(claims jwt.MapClaims) string {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenStr, _ := token.SignedString([]byte(accessTokenSecret))
			return tokenStr
		}
		expiredAccessToken = accessTokenFn(jwt.MapClaims{
			"sub": "",
			"exp": time.Now().Unix(),
		})
		accessToken = accessTokenFn(jwt.MapClaims{
			"sub": userID,
			"exp": time.Now().Add(time.Hour).Unix(),
		})
	)
	handlerMock := mock.NewMockHandler(t)
	writerMock := mock.NewMockResponseWriter(t)
	viper.Set("access_token_secret", accessTokenSecret)
	testCases := []TestCase{
		{
			description: "Пустой запрос должен возвращать ошибку авторизации",
			request:     &http.Request{},
			mockAssertion: func(_ TestCase) {
				writerMock.EXPECT().WriteHeader(http.StatusUnauthorized)
				writerMock.EXPECT().Write([]byte(http.StatusText(http.StatusUnauthorized))).Return(401, nil)
			},
		},
		{
			description: "Запрос с неправильным заголовком авторизации должен возвращать ошибку авторизации",
			request:     &http.Request{Header: http.Header{"Authorization": []string{"Bea" + accessToken}}},
			mockAssertion: func(_ TestCase) {
				writerMock.EXPECT().WriteHeader(http.StatusUnauthorized)
				writerMock.EXPECT().Write([]byte(http.StatusText(http.StatusUnauthorized))).Return(401, nil)
			},
		},
		{
			description: "Запрос с неправильным токеном авторизации должен возвращать ошибку авторизации",
			request:     &http.Request{Header: http.Header{"Authorization": []string{"Bearer " + wrongAccessToken}}},
			mockAssertion: func(_ TestCase) {
				writerMock.EXPECT().WriteHeader(http.StatusUnauthorized)
				writerMock.EXPECT().Write([]byte(http.StatusText(http.StatusUnauthorized))).Return(401, nil)
			},
		},
		{
			description: "Запрос с истекшим токеном авторизации должен возвращать ошибку авторизации",
			request:     &http.Request{Header: http.Header{"Authorization": []string{"Bearer " + expiredAccessToken}}},
			mockAssertion: func(_ TestCase) {
				writerMock.EXPECT().WriteHeader(http.StatusUnauthorized)
				writerMock.EXPECT().Write([]byte(http.StatusText(http.StatusUnauthorized))).Return(401, nil)
			},
		},
		{
			description: "Запрос с правильным заголовком авторизации должен пропускаться без ошибок",
			request:     &http.Request{Header: http.Header{"Authorization": []string{"Bearer " + accessToken}}},
			mockAssertion: func(tc TestCase) {
				handlerMock.EXPECT().ServeHTTP(writerMock, tc.request)
			},
			dataAssertion: func(tc TestCase) {
				assert.Equal(t, userID, tc.request.Header.Get("User-Id"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.description,
			func(_ *testing.T) {
				tc.mockAssertion(tc)

				middleware.NewAuth(func(writer http.ResponseWriter, request *http.Request) {
					handlerMock.ServeHTTP(writer, request)
				}).ServeHTTP(writerMock, tc.request)
				if tc.dataAssertion != nil {
					tc.dataAssertion(tc)
				}
			})
	}
}
