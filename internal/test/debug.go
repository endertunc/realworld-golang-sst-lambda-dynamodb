package test

import (
	"encoding/json"
	"log"
)

// used for debugging purposes
func PrintAsJSON(obj interface{}) {
	bytes, _ := json.MarshalIndent(obj, "\t", "\t")
	log.Println(string(bytes))
}
