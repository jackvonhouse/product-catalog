package cache

import (
	"context"
	"github.com/jackvonhouse/product-catalog/config"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/patrickmn/go-cache"
	"time"
)

type Database struct {
	db *cache.Cache
}

func New(
	_ context.Context,
	config config.Cache,
	logger log.Logger,
) (Database, error) {

	expireDuration := time.Duration(config.ExpireDuration) * time.Minute
	cleanupDuration := time.Duration(config.CleanupInterval) * time.Minute

	c := cache.New(expireDuration, cleanupDuration)

	logger.WithFields(map[string]any{
		"duration": map[string]any{
			"expire":  expireDuration,
			"cleanup": cleanupDuration,
		},
	}).Info("cache initialized")

	return Database{
		db: c,
	}, nil
}

func (d Database) Database() *cache.Cache { return d.db }
