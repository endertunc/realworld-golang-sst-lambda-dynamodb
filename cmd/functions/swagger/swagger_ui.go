package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/api/openapi"
)

// reference: https://github.com/swagger-api/swagger-ui/blob/HEAD/docs/usage/installation.md#unpkg
var (
	spec, _   = openapi.GenerateAPISpec().MarshalJSON()
	swaggerUI = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
		<meta name="description" content="SwaggerUI" />
		<title>SwaggerUI</title>
		<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
  	</head>
	<body>
  		<div id="swagger-ui"></div>
  		<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
  		<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js" crossorigin></script>
  		<script>
			window.onload = () => {
		  		window.ui = SwaggerUIBundle({
					url: '/docs/spec.json',
					dom_id: '#swagger-ui',
					presets: [
			  			SwaggerUIBundle.presets.apis,
			  			SwaggerUIStandalonePreset
					],
					layout: "StandaloneLayout",
		  		});
			};
  		</script>
  	</body>
</html>`
)

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/docs/spec.json" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(spec)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(swaggerUI))
	})
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}
