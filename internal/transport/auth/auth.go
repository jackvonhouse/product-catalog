package auth

import (
	"context"
	"encoding/json"
	"github.com/jackvonhouse/product-catalog/internal/transport/validator"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/internal/transport"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type useCaseAuth interface {
	SignUp(context.Context, dto.Credentials) (dto.TokenPair, error)
	SignIn(context.Context, dto.Credentials) (dto.TokenPair, error)
	Refresh(context.Context, dto.TokenPair) (dto.TokenPair, error)
}

type Transport struct {
	useCase useCaseAuth

	logger log.Logger
}

func New(
	auth useCaseAuth,
	logger log.Logger,
) Transport {

	return Transport{
		useCase: auth,
		logger:  logger.WithField("layer", "transport"),
	}
}

func (t Transport) Handle(
	router *mux.Router,
) {

	router.HandleFunc("/sign-in", t.SignIn).
		Methods(http.MethodPost)

	router.HandleFunc("/sign-up", t.SignUp).
		Methods(http.MethodPost)

	router.HandleFunc("/refresh", t.Refresh).
		Methods(http.MethodPost)
}

// SignUp godoc
// @Summary			Регистрация
// @Description		Регистрация нового пользователя
// @Accept			json
// @Produce			json
// @Param			request body object{username=string,password=string} true "Данные пользователя"
// @Success			200 {object} dto.TokenPair
// @Failure			409 {object} object{error=string} "Пользователь уже существует"
// @Failure			500 {object} object{error=string} "Неизвестная ошибка"
// @Tags			Авторизация
// @Router /user/sign-up [post]
func (t Transport) SignUp(
	w http.ResponseWriter,
	r *http.Request,
) {

	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		t.logger.Warn(err)

		transport.Error(w, http.StatusBadRequest, "invalid json structure")

		return
	}

	if err := validator.IsValidCredentials(data.Username, data.Password); err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	signUp := dto.Credentials{
		Username: data.Username,
		Password: data.Password,
	}

	tokenPair, err := t.useCase.SignUp(ctx, signUp)
	if err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	transport.Response(w, tokenPair)
}

// SignIn godoc
// @Summary			Авторизация
// @Description		Авторизация пользователя
// @Accept			json
// @Produce			json
// @Param			request body object{username=string,password=string} true "Данные пользователя"
// @Success			200 {object} dto.TokenPair
// @Failure			500 {object} object{error=string} "Неизвестная ошибка"
// @Tags			Авторизация
// @Router /user/sign-in [post]
func (t Transport) SignIn(
	w http.ResponseWriter,
	r *http.Request,
) {

	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		t.logger.Warn(err)

		transport.Error(w, http.StatusBadRequest, "invalid json structure")

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	signIn := dto.Credentials{
		Username: data.Username,
		Password: data.Password,
	}

	tokenPair, err := t.useCase.SignIn(ctx, signIn)
	if err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	transport.Response(w, tokenPair)
}

// Refresh godoc
// @Summary			Обновление токенов
// @Description		Обновление токенов
// @Accept			json
// @Produce			json
// @Param			request body object{access_token=string,refresh_token=string} true "Пара токенов"
// @Success			200 {object} dto.TokenPair
// @Failure			500 {object} object{error=string} "Неизвестная ошибка"
// @Tags			Авторизация
// @Router /user/refresh [post]
func (t Transport) Refresh(
	w http.ResponseWriter,
	r *http.Request,
) {

	data := dto.TokenPair{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		t.logger.Warn(err)

		transport.Error(w, http.StatusBadRequest, "invalid json structure")

		return
	}

	if data.AccessToken == "" {
		transport.Error(w, http.StatusBadRequest, "access token can't be empty")

		return
	}

	if data.RefreshToken == "" {
		transport.Error(w, http.StatusBadRequest, "refresh token can't be empty")

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tokenPair, err := t.useCase.Refresh(ctx, data)
	if err != nil {
		t.logger.Warn(err)

		code, msg := transport.ErrorToHttpResponse(err)

		transport.Error(w, code, msg)

		return
	}

	transport.Response(w, tokenPair)
}
