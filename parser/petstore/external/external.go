package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackvonhouse/product-catalog/parser/petstore/config"
	"github.com/jackvonhouse/product-catalog/parser/petstore/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"net/http"
)

type PetStore struct {
	config config.ProductCatalogAPI

	logger log.Logger
}

func New(
	config config.ProductCatalogAPI,
	logger log.Logger,
) PetStore {

	return PetStore{
		config: config,
		logger: logger.WithField("unit", "petstore"),
	}
}

func (s PetStore) Create(
	ctx context.Context,
	pet dto.Pet,
) (int64, error) {

	accessToken, _, err := s.getTokenPair(ctx, s.config)
	if err != nil {
		return 0, err
	}

	s.logger.Infof("received access token: %s", accessToken)

	categoryId, err := s.createCategory(ctx, pet, accessToken, s.config)
	if err != nil {
		return 0, err
	}

	pet.Category.ID = int64(categoryId)

	productId, err := s.createProduct(ctx, pet, accessToken, s.config)
	if err != nil {
		return 0, err
	}

	return int64(productId), nil
}

func (s PetStore) getTokenPair(
	ctx context.Context,
	config config.ProductCatalogAPI,
) (string, string, error) {

	url := fmt.Sprintf("%s/api/v1/user/sign-in", config.Source)
	body := fmt.Sprintf(`{"username": "%s", "password": "%s"}`,
		config.Username, config.Password,
	)

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, url,
		bytes.NewBufferString(body),
	)

	if err != nil {
		return "", "", fmt.Errorf("creating request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("get request to failed: %w", err)
	}

	defer resp.Body.Close()

	var tokenPair struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenPair); err != nil {
		return "", "", fmt.Errorf("decode json error: %w", err)
	}

	return tokenPair.AccessToken, tokenPair.RefreshToken, nil
}

func (s PetStore) createCategory(
	ctx context.Context,
	pet dto.Pet,
	accessToken string,
	config config.ProductCatalogAPI,
) (int, error) {

	url := fmt.Sprintf("%s/api/v1/category", config.Source)
	body := fmt.Sprintf(`{"name": "%s"}`, pet.Category.Name)

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, url,
		bytes.NewBufferString(body),
	)

	if err != nil {
		return 0, fmt.Errorf("creating category request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("post request failed: %w", err)
	}

	defer resp.Body.Close()

	var category struct {
		ID int `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&category); err != nil {
		return 0, fmt.Errorf("decode json error: %w", err)
	}

	if category.ID == 0 {
		return 0, fmt.Errorf("category not created")
	}

	return category.ID, nil
}

func (s PetStore) createProduct(
	ctx context.Context,
	pet dto.Pet,
	accessToken string,
	config config.ProductCatalogAPI,
) (int, error) {

	url := fmt.Sprintf("%s/api/v1/product", config.Source)
	body := fmt.Sprintf(`{"name": "%s", "category_id": %d}`,
		pet.Name, pet.Category.ID,
	)

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, url,
		bytes.NewBufferString(body),
	)

	if err != nil {
		return 0, fmt.Errorf("creating product request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("post request failed: %w", err)
	}

	defer resp.Body.Close()

	var product struct {
		ID int `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return 0, fmt.Errorf("decode json error: %w", err)
	}

	if product.ID == 0 {
		return 0, fmt.Errorf("product not created")
	}

	return product.ID, nil
}
