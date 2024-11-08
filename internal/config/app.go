package config

type App struct {
	LogLevel   string `env:"LOG_LEVEL" envDefault:"info"`
	Prometheus Prometheus
	Health     Health
	Nats       Nats
	DB         DB
	API        API
	Core       Core
	Vault      Vault
	Zerion     Zerion
	AI         AI
}
