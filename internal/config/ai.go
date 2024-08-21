package config

type AI struct {
	MonthlyRateLimit  int64  `env:"AI_MONTHLY_RATE_LIMIT" envDefault:"10"`
	ExternalClientKey string `env:"AI_EXTERNAL_CLIENT_KEY" require:"true"`
}
