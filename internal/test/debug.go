package test

import "encoding/json"

// used for debugging purposes
func PrintAsJSON(obj interface{}, printer func(value ...any)) {
	bytes, _ := json.MarshalIndent(obj, "\t", "\t")
	printer(string(bytes))
}
