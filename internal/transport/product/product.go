package product

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/internal/transport"
	"github.com/jackvonhouse/product-catalog/internal/transport/middleware"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"net/http"
	"time"
)

type productUseCase interface {
	Create(context.Context, dto.CreateProduct) (int, error)

	Get(context.Context, dto.GetProduct) ([]dto.Product, error)
	GetByCategoryId(context.Context, dto.GetProduct, int) ([]dto.Product, error)

	Update(context.Context, dto.UpdateProduct) (int, error)

	Delete(context.Context, int) (int, error)
}

type useCaseAccessToken interface {
	Verify(context.Context, string) error
}

type Transport struct {
	product     productUseCase
	accessToken useCaseAccessToken

	mw     middleware.Middleware
	logger log.Logger
}

func New(
	product productUseCase,
	accessToken useCaseAccessToken,
	logger log.Logger,
) Transport {

	return Transport{
		product:     product,
		accessToken: accessToken,
		mw:          middleware.New(accessToken, logger),
		logger:      logger.WithField("unit", "product"),
	}
}

func (t Transport) Handle(
	router *mux.Router,
) {

	authorizedOnly := router.PathPrefix("").Subrouter()
	authorizedOnly.Use(t.mw.AuthorizedOnly)

	authorizedOnly.HandleFunc("", t.Create).
		Methods(http.MethodPost)

	router.HandleFunc("", t.GetByCategoryId).
		Methods(http.MethodGet).
		Queries("category_id", "{category_id:[0-9]+}")

	router.HandleFunc("", t.Get).
		Methods(http.MethodGet)

	authorizedOnly.HandleFunc("/{id:[0-9]+}", t.Update).
		Methods(http.MethodPut)

	authorizedOnly.HandleFunc("/{id:[0-9]+}", t.Delete).
		Methods(http.MethodDelete)
}

// Create godoc
// @Summary			Создать товар
// @Description		Создание товара с определённой категорией
// @Security		Bearer
// @Accept			json
// @Produce			json
// @Param			request body dto.CreateProduct true "Данные о товаре"
// @Success			200 {object} object{id=int}
// @Failure			401 {object} object{error=string} "Пользователь не авторизован"
// @Failure			409 {object} object{error=string} "Товар уже существует"
// @Failure			500 {object} object{error=string} "Неизвестная ошибка"
// @Tags			Товар
// @Router /product [post]
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

	id, err := t.product.Create(ctx, data)
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

	categoryId, err := transport.StringToInt(r.URL.Query().Get("category_id"))
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

	products, err := t.product.GetByCategoryId(ctx, data, categoryId)
	if err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	transport.Response(w, products)
}

// Get godoc
// @Summary			Получить товары
// @Description		Получение товаров
// @Accept			json
// @Produce			json
// @Param			limit path int false "Лимит"
// @Param			offset path int false "Смещение"
// @Param			category_id path int false "Идентификатор категории"
// @Success			200 {array} dto.Product
// @Failure			404 {object} object{error=string} "Товары отсутствуют или категория не найдена"
// @Failure			500 {object} object{error=string} "Неизвестная ошибка"
// @Tags			Товар
// @Router /product [get]
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

	products, err := t.product.Get(ctx, data)
	if err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	transport.Response(w, products)
}

// Update godoc
// @Summary			Обновить товар
// @Description		Обновление товара
// @Security		Bearer
// @Accept			json
// @Produce			json
// @Param			request body dto.UpdateProduct true "Данные о товаре"
// @Param			id path int true "Идентификатор товара"
// @Success			200 {object} object{id=int}
// @Failure			401 {object} object{error=string} "Пользователь не авторизован"
// @Failure			404 {object} object{error=string} "Товар или категория не найдены"
// @Failure			409 {object} object{error=string} "Товар уже существует"
// @Failure			500 {object} object{error=string} "Неизвестная ошибка"
// @Tags			Товар
// @Router /product/{id} [put]
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

	id, err := t.product.Update(ctx, data)
	if err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	transport.Response(w, map[string]any{"id": id})
}

// Delete godoc
// @Summary			Удалить товар
// @Description		Удаление товара
// @Security		Bearer
// @Accept			json
// @Produce			json
// @Param			id path int true "Идентификатор товара"
// @Success			200 {object} object{id=int}
// @Failure			401 {object} object{error=string} "Пользователь не авторизован"
// @Failure			404 {object} object{error=string} "Товар не найден"
// @Failure			500 {object} object{error=string} "Неизвестная ошибка"
// @Tags			Товар
// @Router /product/{id} [delete]
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

	id, err := t.product.Delete(ctx, productId)
	if err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	transport.Response(w, map[string]any{"id": id})
}
