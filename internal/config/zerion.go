package config

type Zerion struct {
	BaseURL     string `env:"ZERION_API_BASE_URL" require:"true"`
	Key         string `env:"ZERION_API_KEY" require:"true"`
	MappingPath string `env:"ZERION_MAPPING_SOURCE" require:"true"`
}
