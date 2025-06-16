package v1

import (
	"net/http"
	"user/internal/domain/usecase"
	"user/internal/endpoints/api/frontend/middleware"

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
	mux.Post("/login", router.Login)
	mux.Post("/sign", router.Sign)
	mux.Post("/request-confirm", func(writer http.ResponseWriter, r *http.Request) {
		middleware.NewAuth(router.RequestConfirm).ServeHTTP(writer, r)
	})
	mux.Post("/confirm", func(writer http.ResponseWriter, r *http.Request) {
		middleware.NewAuth(router.Confirm).ServeHTTP(writer, r)
	})
	router.Handler = mux
	return router
}

func (r *router) Login(w http.ResponseWriter, req *http.Request) {
	data, err := r.converter.ConvertLogin(req)
	if err != nil {
		r.converter.WriteError(w, err, http.StatusBadRequest)
		return
	}

	res, err := r.cases.Login(req.Context(), data)
	if err != nil {
		r.converter.WriteError(w, err, http.StatusNotFound)
		return
	}
	r.converter.WriteAuthResult(w, res)
}

func (r *router) Sign(w http.ResponseWriter, req *http.Request) {
	data, err := r.converter.ConvertSign(req)
	if err != nil {
		r.converter.WriteError(w, err, http.StatusBadRequest)
		return
	}

	err = r.cases.Sign(req.Context(), data)
	if err != nil {
		r.converter.WriteError(w, err, http.StatusBadRequest)
		return
	}
	r.converter.WriteNoContentResult(w)
}

func (r *router) RequestConfirm(w http.ResponseWriter, req *http.Request) {
	data, err := r.converter.ConvertRequestConfirm(req)
	if err != nil {
		r.converter.WriteError(w, err, http.StatusBadRequest)
		return
	}

	err = r.cases.RequestConfirm(req.Context(), data)
	if err != nil {
		r.converter.WriteError(w, err, http.StatusBadRequest)
		return
	}
	r.converter.WriteNoContentResult(w)
}

func (r *router) Confirm(w http.ResponseWriter, req *http.Request) {
	data, err := r.converter.ConvertConfirm(req)
	if err != nil {
		r.converter.WriteError(w, err, http.StatusBadRequest)
		return
	}

	err = r.cases.Confirm(req.Context(), data)
	if err != nil {
		r.converter.WriteError(w, err, http.StatusBadRequest)
		return
	}
	r.converter.WriteNoContentResult(w)
}
