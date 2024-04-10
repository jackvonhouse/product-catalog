package main

import (
	"context"
	"flag"

	"github.com/jackvonhouse/product-catalog/app"
	"github.com/jackvonhouse/product-catalog/config"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/jackvonhouse/product-catalog/pkg/shutdown"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.NewLogrusLogger()

	var configPath string

	flag.StringVar(
		&configPath,
		"config",
		"config/config.toml",
		"The path to the configuration file",
	)

	flag.Parse()

	cfg, err := config.New(configPath, logger)
	if err != nil {
		logger.Error(err)

		return
	}

	application, err := app.New(ctx, cfg, logger)
	if err != nil {
		logger.Error(err)

		return
	}

	go application.Run()

	shutdown.Graceful(ctx, cancel, logger, application)
}
