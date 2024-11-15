package domain

import (
	"github.com/brianvoe/gofakeit/v7"
	"log"
	"os"
	"strconv"
)

// we use random seed by default, but we could also set static seed for debugging purposes
// this package level init() makes sure that all generators use the same seed
// ToDo @ender I might move this to somewhere else. I kind of didn't like this implicit logic here...
func init() {
	seed, ok := os.LookupEnv("GOFAKEIT_SEED")
	if ok {
		seedInt, err := strconv.ParseUint(seed, 10, 64)
		if err != nil {
			log.Fatalf("GOFAKEIT_SEED must be a valid intiger: %v", err)
		}
		gofakeit.GlobalFaker = gofakeit.New(seedInt)
	}
}
