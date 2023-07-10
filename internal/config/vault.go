package config

type Vault struct {
	Address  string `env:"VAULT_ADDR" envDefault:"http://127.0.0.1:8200"`
	Token    string `env:"VAULT_TOKEN"`
	BasePath string `env:"VAULT_BASE_PATH" envDefault:"/"`
}
