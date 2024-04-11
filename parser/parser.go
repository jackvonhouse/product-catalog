package main

import (
	"context"
	"fmt"
	"github.com/jackvonhouse/product-catalog/parser/petstore"
	"github.com/jackvonhouse/product-catalog/parser/petstore/config"
	"github.com/jackvonhouse/product-catalog/parser/petstore/external"
	"github.com/jackvonhouse/product-catalog/parser/petstore/service"
	"github.com/jackvonhouse/product-catalog/parser/petstore/storage"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/patrickmn/go-cache"
	"time"
)

func main() {
	logger := log.NewLogrusLogger()

	cfg, err := config.New(
		"parser/petstore/config/config.toml",
		logger.WithField("layer", "config"),
	)

	if err != nil {
		fmt.Println(err)
	}

	cleanUpExp := 24 * time.Hour
	defaultExp := cleanUpExp / 2

	c := cache.New(defaultExp, cleanUpExp)

	internalRepo := storage.New(c, cfg.External,
		logger.WithField("layer", "cache"),
	)

	productCatalogRepo := external.New(
		cfg.Internal,
		logger.WithField("layer", "product_catalog_api"),
	)

	s := service.New(
		internalRepo,
		productCatalogRepo,
		cfg.Internal,
		logger.WithField("layer", "service"),
	)

	pet := petstore.New(
		s,
		cfg.External,
		logger.WithField("layer", "parser"),
	)

	ticker := time.Tick(
		time.Duration(cfg.ParsePeriod) * time.Minute,
	)

	for range ticker {
		if err := pet.Get(context.Background()); err != nil {
			fmt.Println(err)
		}
	}
}
