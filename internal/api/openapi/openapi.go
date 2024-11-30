package openapi

import (
	"log"
	"reflect"
	"strings"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi3"
)

var BearerAuthSecurityName = "BearerAuth"
var NoAuthSecurityName = "NoAuth" // ToDo @ender this doesn't seem to be working in generated openapi spec

func GenerateAPISpec() *openapi3.Spec {
	reflector := openapi3.Reflector{}
	reflector.DefaultOptions = append(reflector.DefaultOptions, jsonschema.InterceptDefName(
		func(t reflect.Type, defaultDefName string) string {
			// remove prefixed package names from the definition name
			modifiedDefName := strings.TrimPrefix(defaultDefName, "Dto")
			modifiedDefName = strings.TrimPrefix(modifiedDefName, "Errutil")
			return modifiedDefName
		},
	))

	reflector.Spec = &openapi3.Spec{Openapi: "3.0.0"}
	reflector.Spec.Info.WithTitle("Realworld API Specification")
	reflector.Spec.Servers = []openapi3.Server{{}} // ToDo @ender TBD

	buildUser(&reflector)
	buildProfile(&reflector)
	buildArticle(&reflector)
	buildComment(&reflector)

	reflector.SpecEns().SetHTTPBearerTokenSecurity(BearerAuthSecurityName, "JWT", "")

	// ToDo @ender this is for debugging purposes for now delete it later
	schema, err := reflector.Spec.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(schema))

	return reflector.Spec
}
