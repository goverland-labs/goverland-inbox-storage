package config

type Zerion struct {
	BaseURL string `env:"ZERION_API_BASE_URL" require:"true"`
	Key     string `env:"ZERION_API_KEY" require:"true"`
}
