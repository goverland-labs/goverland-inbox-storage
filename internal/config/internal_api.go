package config

type API struct {
	Bind string `env:"API_GRPC_SERVER_BIND" envDefault:":11000"`
}
