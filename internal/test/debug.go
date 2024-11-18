package test

import (
	"encoding/json"
)

// used for debugging purposes
func PrintAsJSON(obj interface{}) {
	bytes, _ := json.MarshalIndent(obj, "\t", "\t")
	println(string(bytes))
}
