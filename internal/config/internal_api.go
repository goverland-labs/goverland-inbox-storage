package config

type API struct {
	Bind               string `env:"API_GRPC_SERVER_BIND" envDefault:":11000"`
	FeedAddress        string `env:"INBOX_API_FEED_ADDRESS" envDefault:"localhost:11066"`
	EnsResolverAddress string `env:"INTERNAL_API_ENS_RESOLVER_ADDRESS" envDefault:":20200"`
}
