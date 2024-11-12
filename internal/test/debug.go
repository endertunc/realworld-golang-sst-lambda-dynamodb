package test

import (
	"encoding/json"
	"log/slog"
)

// used for debugging purposes
func PrintAsJSON(obj interface{}, printer func(value ...any)) {
	bytes, _ := json.MarshalIndent(obj, "\t", "\t")
	printer(string(bytes))
}

func SlogAsJSON(obj interface{}) {
	bytes, _ := json.MarshalIndent(obj, "\t", "\t")
	slog.Info(string(bytes))
}
