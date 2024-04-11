package petstore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackvonhouse/product-catalog/parser/petstore/config"
	"github.com/jackvonhouse/product-catalog/parser/petstore/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"net/http"
)

type serviceStorage interface {
	Save(context.Context, dto.Pet) error
}

type PetStore struct {
	service serviceStorage

	source   string
	interval int

	logger log.Logger
}

func New(
	storage serviceStorage,
	config config.ExternalAPI,
	logger log.Logger,
) PetStore {

	return PetStore{
		service:  storage,
		source:   config.Source,
		interval: config.Duration,
		logger:   logger.WithField("unit", "petstore"),
	}
}

func (p PetStore) Get(
	ctx context.Context,
) error {

	pets, err := p.fetch(ctx)
	if err != nil {
		return err
	}

	p.logger.Infof("fetched %d pets from petstore", len(pets))

	for _, pet := range pets {
		if err := p.service.Save(ctx, pet); err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func (p PetStore) fetch(
	ctx context.Context,
) ([]dto.Pet, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.source, nil)
	if err != nil {
		return []dto.Pet{}, fmt.Errorf("creating request failed: %w", err)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []dto.Pet{}, fmt.Errorf("get request to failed: %w", err)
	}

	p.logger.Info("get request to petstore was successful")

	defer resp.Body.Close()

	pets := make([]dto.Pet, 0)

	if err := json.NewDecoder(resp.Body).Decode(&pets); err != nil {
		return []dto.Pet{}, fmt.Errorf("decode json error: %w", err)
	}

	p.logger.Infof("received pets from petstore: %d", len(pets))

	return pets, nil
}
