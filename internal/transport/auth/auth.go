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
