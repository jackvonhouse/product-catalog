package category

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

type useCaseCategory interface {
	Create(context.Context, dto.CreateCategory) (int, error)

	Get(context.Context, dto.GetCategory) ([]dto.Category, error)

	Update(context.Context, dto.UpdateCategory) (int, error)

	Delete(context.Context, int) (int, error)
}

type useCaseAccessToken interface {
	Verify(context.Context, string) error
}

type Transport struct {
	useCase useCaseCategory

	mw     middleware.Middleware
	logger log.Logger
}

func New(
	category useCaseCategory,
	accessToken useCaseAccessToken,
	logger log.Logger,
) Transport {

	return Transport{
		useCase: category,
		mw:      middleware.New(accessToken, logger),
		logger:  logger.WithField("unit", "category"),
	}
}

func (t Transport) Handle(
	router *mux.Router,
) {
	authorizedOnly := router.PathPrefix("").Subrouter()
	authorizedOnly.Use(t.mw.AuthorizedOnly)

	authorizedOnly.HandleFunc("", t.Create).
		Methods(http.MethodPost)

	router.HandleFunc("", t.Get).
		Methods(http.MethodGet)

	authorizedOnly.HandleFunc("/{id:[0-9]+}", t.Update).
		Methods(http.MethodPut)

	authorizedOnly.HandleFunc("/{id:[0-9]+}", t.Delete).
		Methods(http.MethodDelete)
}

// Create godoc
// @Summary			Создать категорию
// @Description		Создание категории
// @Security		Bearer
// @Accept			json
// @Produce			json
// @Param			request body dto.CreateCategory true "Данные о категории"
// @Success			200 {object} object{id=int}
// @Failure			401 {object} object{error=string} "Пользователь не авторизован"
// @Failure			409 {object} object{error=string} "Категория уже существует"
// @Failure			500 {object} object{error=string} "Неизвестная ошибка"
// @Tags			Категория
// @Router /category [post]
func (t Transport) Create(
	w http.ResponseWriter,
	r *http.Request,
) {

	data := dto.CreateCategory{}
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

// Get godoc
// @Summary			Получить категории
// @Description		Получение категорий
// @Accept			json
// @Produce			json
// @Param			limit path int false "Лимит"
// @Param			offset path int false "Смещение"
// @Success			200 {array} dto.Category
// @Failure			404 {object} object{error=string} "Категории отсутствуют"
// @Failure			500 {object} object{error=string} "Неизвестная ошибка"
// @Tags			Категория
// @Router /category [get]
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

	data := dto.GetCategory{
		Limit:  limit,
		Offset: offset,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	categories, err := t.useCase.Get(ctx, data)
	if err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	transport.Response(w, categories)
}

// Update godoc
// @Summary			Обновить категорию
// @Description		Обновление категории
// @Security		Bearer
// @Accept			json
// @Produce			json
// @Param			request body dto.UpdateCategory true "Данные о категории"
// @Param			id path int true "Идентификатор категории"
// @Success			200 {object} object{id=int}
// @Failure			401 {object} object{error=string} "Пользователь не авторизован"
// @Failure			404 {object} object{error=string} "Категория не найдена"
// @Failure			409 {object} object{error=string} "Категория уже существует"
// @Failure			500 {object} object{error=string} "Неизвестная ошибка"
// @Tags			Категория
// @Router /category/{id} [put]
func (t Transport) Update(
	w http.ResponseWriter,
	r *http.Request,
) {

	vars := mux.Vars(r)

	categoryId, err := transport.StringToInt(vars["id"])
	if err != nil || categoryId <= 0 {
		transport.Error(w, http.StatusBadRequest, "invalid category id")

		return
	}

	data := dto.UpdateCategory{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		transport.Error(
			w,
			http.StatusInternalServerError,
			"invalid json structure",
		)

		return
	}

	data.ID = categoryId

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

// Delete godoc
// @Summary			Удалить категорию
// @Description		Удаление категории
// @Security		Bearer
// @Accept			json
// @Produce			json
// @Param			id path int true "Идентификатор категории"
// @Success			200 {object} object{id=int}
// @Failure			401 {object} object{error=string} "Пользователь не авторизован"
// @Failure			404 {object} object{error=string} "Категория не найдена"
// @Failure			500 {object} object{error=string} "Неизвестная ошибка"
// @Tags			Категория
// @Router /category/{id} [delete]
func (t Transport) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {

	vars := mux.Vars(r)

	categoryId, err := transport.StringToInt(vars["id"])
	if err != nil || categoryId <= 0 {
		transport.Error(w, http.StatusBadRequest, "invalid category id")

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, err := t.useCase.Delete(ctx, categoryId)
	if err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	transport.Response(w, map[string]any{"id": id})
}
