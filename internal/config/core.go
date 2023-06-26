package config

type Core struct {
	CoreURL      string `env:"CORE_URL" envDefault:""`
	SubscriberID string `env:"CORE_SUBSCRIBER_ID" envDefault:""`
}
