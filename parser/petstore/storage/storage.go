package storage

import (
	"context"
	"fmt"
	"github.com/jackvonhouse/product-catalog/parser/petstore/config"
	"github.com/jackvonhouse/product-catalog/parser/petstore/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/patrickmn/go-cache"
	"time"
)

type PetStore struct {
	cache *cache.Cache

	config config.ExternalAPI

	logger log.Logger
}

func New(
	db *cache.Cache,
	config config.ExternalAPI,
	logger log.Logger,
) PetStore {

	return PetStore{
		cache:  db,
		config: config,
		logger: logger.WithField("unit", "petstore"),
	}
}

func (s PetStore) Create(
	_ context.Context,
	pet dto.Pet,
) (int64, error) {

	key := fmt.Sprintf("%d", pet.ID)
	value := map[string]any{
		"name": pet.Name,
		"category": map[string]any{
			"id":   pet.Category.ID,
			"name": pet.Category.Name,
		},
	}

	expDuration := time.Duration(s.config.Duration) * time.Minute

	if err := s.cache.Add(key, value, expDuration); err != nil {
		return 0, fmt.Errorf("failed to add item to cache: %w", err)
	}

	s.logger.Infof("pet with id %d added to cache", pet.ID)

	return 0, nil
}
func (s PetStore) GetById(
	_ context.Context,
	id int64,
) (dto.Pet, error) {

	key := fmt.Sprintf("%d", id)

	value, exp, ok := s.cache.GetWithExpiration(key)
	if !ok || exp.Before(time.Now()) {
		return dto.Pet{}, fmt.Errorf("failed to get item from cache")
	}

	m := value.(map[string]any)
	mC := m["category"].(map[string]any)

	petName := m["name"].(string)
	petCategoryID := mC["id"].(int64)
	petCategoryName := mC["name"].(string)

	pet := dto.Pet{
		ID:   id,
		Name: petName,
		Category: dto.Category{
			ID:   petCategoryID,
			Name: petCategoryName,
		},
	}

	s.logger.Infof("pet with id %d found in cache", id)

	return pet, nil
}
