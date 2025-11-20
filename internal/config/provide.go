package config

import (
	"fmt"

	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	"github.com/spf13/viper"
	"go.uber.org/dig"
)

type configValueProvider struct {
	cfg        *viper.Viper
	configPath string
	diPath     string
}

func provideConfigValue(cfg *viper.Viper, path string) configValueProvider {
	if !cfg.IsSet(path) {
		panic(fmt.Errorf("config key not found: %s", path))
	}
	return configValueProvider{cfg, path, "config." + path}
}

func (p configValueProvider) asInt() di.ConstructorWithOpts {
	return di.ProvideValue(p.cfg.GetInt(p.configPath), dig.Name(p.diPath))
}

func (p configValueProvider) asInt32() di.ConstructorWithOpts {
	return di.ProvideValue(p.cfg.GetInt32(p.configPath), dig.Name(p.diPath))
}

func (p configValueProvider) asString() di.ConstructorWithOpts {
	return di.ProvideValue(p.cfg.GetString(p.configPath), dig.Name(p.diPath))
}

func (p configValueProvider) asBool() di.ConstructorWithOpts {
	return di.ProvideValue(p.cfg.GetBool(p.configPath), dig.Name(p.diPath))
}

func (p configValueProvider) asDuration() di.ConstructorWithOpts {
	return di.ProvideValue(p.cfg.GetDuration(p.configPath), dig.Name(p.diPath))
}

func (p configValueProvider) asFloat64() di.ConstructorWithOpts {
	return di.ProvideValue(p.cfg.GetFloat64(p.configPath), dig.Name(p.diPath))
}

func Provide(container *dig.Container, cfg *viper.Viper) error {
	return di.ProvideAll(container,
		// env should only be used for tracing/debugging purposes
		provideConfigValue(cfg, "env").asString(),

		provideConfigValue(cfg, "gracefulShutdownTimeout").asDuration(),

		// petstore config
		provideConfigValue(cfg, "petstore.baseURL").asString(),

		// http server config
		provideConfigValue(cfg, "httpServer.host").asString(),
		provideConfigValue(cfg, "httpServer.port").asInt(),
		provideConfigValue(cfg, "httpServer.idleTimeout").asDuration(),
		provideConfigValue(cfg, "httpServer.readHeaderTimeout").asDuration(),
		provideConfigValue(cfg, "httpServer.readTimeout").asDuration(),
		provideConfigValue(cfg, "httpServer.writeTimeout").asDuration(),
		provideConfigValue(cfg, "httpServer.accessLogsLevel").asString(),

		// mcp server config
		provideConfigValue(cfg, "mcpServer.name").asString(),
		provideConfigValue(cfg, "mcpServer.version").asString(),
		provideConfigValue(cfg, "mcpServer.httpHost").asString(),
		provideConfigValue(cfg, "mcpServer.httpPort").asInt(),

		// database config
		provideConfigValue(cfg, "database.dsn").asString(),

		// opentelemetry config
		provideConfigValue(cfg, "openTelemetry.enabled").asBool(),
		provideConfigValue(cfg, "openTelemetry.endpoint").asString(),
		provideConfigValue(cfg, "openTelemetry.protocol").asString(),
		provideConfigValue(cfg, "openTelemetry.samplingRate").asFloat64(),
		provideConfigValue(cfg, "openTelemetry.enableMetrics").asBool(),
		provideConfigValue(cfg, "openTelemetry.enableLogs").asBool(),
	)
}
