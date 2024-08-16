package config

type AI struct {
	MonthlyRateLimit  int64  `env:"MONTHLY_RATE_LIMIT" envDefault:"10"`
	ExternalClientKey string `env:"EXTERNAL_CLIENT_KEY" require:"true"`
}
