package service

import (
	"context"
	"errors"
	"github.com/jackvonhouse/product-catalog/parser/petstore/config"
	"github.com/jackvonhouse/product-catalog/parser/petstore/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"sync"
)

type repositoryPetStore interface {
	Create(context.Context, dto.Pet) (int64, error)

	GetById(context.Context, int64) (dto.Pet, error)
}

type repositoryAPI interface {
	Create(context.Context, dto.Pet) (int64, error)
}

type PetStore struct {
	repository repositoryPetStore
	internal   repositoryAPI

	config config.ProductCatalogAPI
	logger log.Logger
}

func New(
	repository repositoryPetStore,
	internal repositoryAPI,
	config config.ProductCatalogAPI,
	logger log.Logger,
) PetStore {

	return PetStore{
		repository: repository,
		internal:   internal,
		config:     config,
		logger:     logger.WithField("unit", "petstore"),
	}
}

func (p PetStore) Save(
	ctx context.Context,
	pet dto.Pet,
) error {

	_, err := p.repository.GetById(ctx, pet.ID)
	if err == nil {
		p.logger.Infof("pet with id %d found in storage", pet.ID)
		return nil
	}

	chResults := make(chan dto.Result, 2)
	wg := &sync.WaitGroup{}

	wg.Add(2)

	go p.saveInternal(ctx, pet, chResults, wg)
	go p.saveExternal(ctx, pet, chResults, wg)

	wg.Wait()
	close(chResults)

	var (
		rErr error
	)

	for result := range chResults {
		if !result.Success {
			if rErr == nil {
				rErr = result.Error
			} else {
				rErr = errors.Join(result.Error)
			}
		}
	}

	return rErr
}

func (p PetStore) saveInternal(
	ctx context.Context,
	pet dto.Pet,
	ch chan<- dto.Result,
	wg *sync.WaitGroup,

) {
	defer wg.Done()

	id, err := p.repository.Create(ctx, pet)
	ch <- dto.Result{
		ID:      id,
		Success: true,
		Error:   err,
		Source:  "internal",
	}
}

func (p PetStore) saveExternal(
	ctx context.Context,
	pet dto.Pet,
	ch chan<- dto.Result,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	id, err := p.internal.Create(ctx, pet)
	ch <- dto.Result{
		ID:      id,
		Success: true,
		Error:   err,
		Source:  "external",
	}
}
