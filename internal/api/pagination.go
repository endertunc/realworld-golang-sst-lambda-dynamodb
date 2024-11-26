package api

import (
	"github.com/caarlos0/env/v11"
	"log"
)

type PaginationConfig struct {
	DefaultLimit int `env:"DEFAULT_LIMIT,notEmpty" envDefault:"10"`
	MinLimit     int `env:"MIN_LIMIT,notEmpty" envDefault:"1"`
	MaxLimit     int `env:"MAX_LIMIT,notEmpty" envDefault:"20"`
}

// one could define an empty variable declaration and later initialize it inside init() or main()
// but I don't like to create an empty variable declaration and then later assign a value to it.
func GetPaginationConfig() PaginationConfig {
	var cfg PaginationConfig
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}
	return cfg
}
