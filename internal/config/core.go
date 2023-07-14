package config

import (
	"github.com/google/uuid"
)

type Core struct {
	CoreURL      string    `env:"CORE_URL" envDefault:""`
	SubscriberID uuid.UUID `env:"CORE_SUBSCRIBER_ID" envDefault:""`
}
