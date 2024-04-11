package config

import (
	"fmt"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

type ExternalAPI struct {
	Source   string
	Duration int
}

type ProductCatalogAPI struct {
	Username string
	Password string
	Source   string
}

type Config struct {
	External    ExternalAPI
	Internal    ProductCatalogAPI
	ParsePeriod int
}

func New(
	configPath string,
	logger log.Logger,
) (Config, error) {

	configType := strings.TrimPrefix(filepath.Ext(configPath), ".")

	viper.SetConfigType(configType)
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		logger.WithFields(map[string]any{
			"layer":       "config",
			"config_path": configPath,
		}).Warnf("error on reading config: %s", err)

		return Config{}, fmt.Errorf("error on reading config: %s", err)
	}

	externalPrefix := "api.external"
	productCatalogPrefix := "api.internal"

	return Config{
		External: ExternalAPI{
			Source:   viper.GetString(fmt.Sprintf("%s.source", externalPrefix)),
			Duration: viper.GetInt(fmt.Sprintf("%s.interval", externalPrefix)),
		},
		Internal: ProductCatalogAPI{
			Username: viper.GetString(fmt.Sprintf("%s.username", productCatalogPrefix)),
			Password: viper.GetString(fmt.Sprintf("%s.password", productCatalogPrefix)),
			Source:   viper.GetString(fmt.Sprintf("%s.source", productCatalogPrefix)),
		},
		ParsePeriod: viper.GetInt("api.parse_period"),
	}, nil
}
