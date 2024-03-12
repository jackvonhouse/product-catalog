package product

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/internal/transport"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"net/http"
	"time"
)

type useCase interface {
	Create(context.Context, dto.CreateProduct) (int, error)

	Get(context.Context, dto.GetProduct) ([]dto.Product, error)
	GetByCategoryId(context.Context, dto.GetProduct, int) ([]dto.Product, error)

	Update(context.Context, dto.UpdateProduct) (int, error)

	Delete(context.Context, int) (int, error)
}

type Transport struct {
	useCase useCase

	logger log.Logger
}

func New(
	useCase useCase,
	logger log.Logger,
) Transport {

	return Transport{
		useCase: useCase,
		logger:  logger.WithField("unit", "product"),
	}
}

func (t Transport) Handle(
	router *mux.Router,
) {

	router.HandleFunc("", t.Create).
		Methods(http.MethodPost)

	router.HandleFunc("/", t.GetByCategoryId).
		Methods(http.MethodGet).
		Queries("category_id", "{category_id:[0-9]+}")

	router.HandleFunc("", t.Get).
		Methods(http.MethodGet)

	router.HandleFunc("/{id:[0-9]+}", t.Update).
		Methods(http.MethodPut)

	router.HandleFunc("/{id:[0-9]+}", t.Delete).
		Methods(http.MethodDelete)
}

func (t Transport) Create(
	w http.ResponseWriter,
	r *http.Request,
) {

	data := dto.CreateProduct{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		transport.Error(
			w,
			http.StatusInternalServerError,
			"invalid json structure",
		)

		return
	}

	if data.Name == "" {
		transport.Error(
			w,
			http.StatusBadRequest,
			"name can't be empty",
		)

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, err := t.useCase.Create(ctx, data)
	if err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	transport.Response(w, map[string]any{"id": id})
}

func (t Transport) GetByCategoryId(
	w http.ResponseWriter,
	r *http.Request,
) {

	queries := r.URL.Query()

	limit, err := transport.StringToInt(queries.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := transport.StringToInt(queries.Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	data := dto.GetProduct{
		Limit:  limit,
		Offset: offset,
	}

	categoryId, err := transport.StringToInt(r.URL.Query().Get("university_id"))
	if err != nil || categoryId == 0 {
		transport.Error(
			w,
			http.StatusBadRequest,
			"invalid category id",
		)

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	products, err := t.useCase.GetByCategoryId(ctx, data, categoryId)
	if err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	transport.Response(w, products)
}

func (t Transport) Get(
	w http.ResponseWriter,
	r *http.Request,
) {

	queries := r.URL.Query()

	limit, err := transport.StringToInt(queries.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := transport.StringToInt(queries.Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	data := dto.GetProduct{
		Limit:  limit,
		Offset: offset,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	products, err := t.useCase.Get(ctx, data)
	if err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	transport.Response(w, products)
}

func (t Transport) Update(
	w http.ResponseWriter,
	r *http.Request,
) {

	vars := mux.Vars(r)

	productId, err := transport.StringToInt(vars["id"])
	if err != nil || productId <= 0 {
		transport.Error(w, http.StatusBadRequest, "invalid product id")

		return
	}

	data := dto.UpdateProduct{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		transport.Error(
			w,
			http.StatusInternalServerError,
			"invalid json structure",
		)

		return
	}

	data.ID = productId

	if data.Name == "" {
		transport.Error(
			w,
			http.StatusBadRequest,
			"name can't be empty",
		)

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, err := t.useCase.Update(ctx, data)
	if err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	transport.Response(w, map[string]any{"id": id})
}

func (t Transport) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {

	vars := mux.Vars(r)

	productId, err := transport.StringToInt(vars["id"])
	if err != nil || productId <= 0 {
		transport.Error(w, http.StatusBadRequest, "invalid product id")

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, err := t.useCase.Delete(ctx, productId)
	if err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	transport.Response(w, map[string]any{"id": id})
}
