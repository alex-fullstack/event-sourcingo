package v1

import (
	"api/internal/domain/usecase"
	"api/internal/endpoints/api/frontend/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type router struct {
	converter Converter
	cases     usecase.FrontendAPICases
	http.Handler
}

func New(cases usecase.FrontendAPICases, converter Converter) http.Handler {
	router := &router{cases: cases, converter: converter}
	mux := chi.NewRouter()
	mux.Post("/users", func(writer http.ResponseWriter, r *http.Request) {
		middleware.NewAuth(router.Users).ServeHTTP(writer, r)
	})
	router.Handler = mux
	return router
}

func (r *router) Users(w http.ResponseWriter, req *http.Request) {
	data, err := r.converter.ConvertUsers(req)
	if err != nil {
		r.converter.WriteError(w, err, http.StatusBadRequest)
		return
	}

	res, err := r.cases.Users(req.Context(), data)
	if err != nil {
		r.converter.WriteError(w, err, http.StatusBadRequest)
		return
	}
	r.converter.WriteUsersResult(w, res)
}
