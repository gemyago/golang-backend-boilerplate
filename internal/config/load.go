package config

import (
	"embed"
	"fmt"

	"github.com/spf13/viper"
)

//go:embed *.json
var resources embed.FS

func mergeResourceCfg(cfg *viper.Viper, resourceName string) error {
	resourceStream, err := resources.Open(resourceName)
	if err != nil {
		return fmt.Errorf("failed to read config %v: %w", resourceName, err)
	}
	defer resourceStream.Close()

	if err = cfg.MergeConfig(resourceStream); err != nil {
		return fmt.Errorf("failed to load config %v: %w", resourceName, err)
	}
	return nil
}

type LoadOpts struct {
	env string
}

func (opts *LoadOpts) WithEnv(val string) *LoadOpts {
	if val != "" {
		opts.env = val
	}
	return opts
}

func NewLoadOpts() *LoadOpts {
	return &LoadOpts{
		env: "local",
	}
}

func Load(opts *LoadOpts) (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigType("json")

	if err := mergeResourceCfg(cfg, "default.json"); err != nil {
		return nil, err
	}

	if err := mergeResourceCfg(cfg, opts.env+".json"); err != nil {
		return nil, err
	}

	return cfg, nil
}
