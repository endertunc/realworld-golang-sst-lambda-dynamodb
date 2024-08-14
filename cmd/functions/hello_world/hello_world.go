package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	requestsigner "github.com/opensearch-project/opensearch-go/v4/signer/awsv2"
	"io"
	"log"
	"net/http"
	"strings"
)

// e.g. https://opensearch-domain.region.com or Amazon OpenSearch Serverless endpoint

func Handler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	return localTest()
}

func localTest() (events.APIGatewayProxyResponse, error) {
	//fmt.Println(os.Getenv("AWS_ACCESS_KEY")[0:4])
	//fmt.Println(os.Getenv("AWS_ACCESS_KEY_ID"))
	//fmt.Println(os.Getenv("AWS_SECRET_ACCESS_KEY"))
	//fmt.Println(os.Getenv("AWS_SESSION_TOKEN"))

	endpoint := "https://ka9ehtys9fgwac4oju9g.eu-west-1.aoss.amazonaws.com"
	//endpoint := os.Getenv("OPENSEARCH_ENDPOINT")
	fmt.Println("OPENSEARCH_ENDPOINT from environment is", endpoint)

	ctx := context.Background()

	//export AWS_ACCESS_KEY_ID="ASIA6IY35YVFXY647N7Y"
	//export AWS_SECRET_ACCESS_KEY="mA+pJiHCEQbdSknzTxaIUV2Y/Nnsghw5boqDY35O"
	//export AWS_SESSION_TOKEN="IQoJb3JpZ2luX2VjEGgaCWV1LXdlc3QtMSJHMEUCIQDaQQtPBpz8HvWG+assSWH500N/2SbuepZ06y99gVVZSQIgNMISNJ4J0agSKLvS5ThJ8gPE7KAQNC8CzXmmiutnRxUq+wIIYBAAGgw5ODA5MjE3MzAzNzkiDCAxNc60jZHttI2n8irYAry4Fm4AfyTUpOcNL9tVOuSXaaY92FZwGjHxzF4qNtLuOg/fKI4LWWotFeTp2w+GhP9U5oqCb+HhmKPypE/dDquC7HBj/Yo3cSGUIwgW4F8+Eyv4C1ckrTuGSkdigZbfCArZ6rutMj95xCmLJXLg/9jc7P7MhGru5cmTvDQ5uKYJFcANaCo4ZQrpXSQCvNdr9c1zpFagLzgZcqFLbWBUD4AfeBfYfgh+x0IPWHRkZnxeo/eRj3jUtn4krGKKp/yX99ZY3HrqDdu2Jj5hWtl0YrZAMc2Y1JsTQ25J4EBq5cpxCP6n86V12ky5COG/KRqC7twDOJUuXubZxY3exDv2X7Kvbsnx4huZsMMj+f7iurfIx2wFc5760mfXzOfEbMC9SrKr4cv8sXsYC29dga1klG7rr49kpwvMTt+KJCCSnvFteombqTm00XaBDAsgC7icINHG8iw8XqA0MMmG3rUGOqcB0sRp/BVtRIhbw/dRy9DRfD3+hUVyrudp0bQPplwzEVQHuywsEAMKJPqe6iA1kG8HfgAl1GAKVdjP3bHFaRE1RxdMS/C3mN5mEGYvLLdI+Ez1zA0U0HFO070JU19I270BdIXzsHN9mcp+lJXl4MQuOEjL/8ErH8tJqBRhkEIHUSPPb5DPKOcpX3WjE5J7YAhr2MXK2JF3nbhmxlwd0ZsKHGvO0exjtuQ="

	//os.Setenv("AWS_ACCESS_KEY_ID", "ASIA6IY35YVFXY647N7Y")
	//os.Setenv("AWS_SECRET_ACCESS_KEY", "mA+pJiHCEQbdSknzTxaIUV2Y/Nnsghw5boqDY35O")
	//os.Setenv("AWS_SESSION_TOKEN", "IQoJb3JpZ2luX2VjEGgaCWV1LXdlc3QtMSJHMEUCIQDaQQtPBpz8HvWG+assSWH500N/2SbuepZ06y99gVVZSQIgNMISNJ4J0agSKLvS5ThJ8gPE7KAQNC8CzXmmiutnRxUq+wIIYBAAGgw5ODA5MjE3MzAzNzkiDCAxNc60jZHttI2n8irYAry4Fm4AfyTUpOcNL9tVOuSXaaY92FZwGjHxzF4qNtLuOg/fKI4LWWotFeTp2w+GhP9U5oqCb+HhmKPypE/dDquC7HBj/Yo3cSGUIwgW4F8+Eyv4C1ckrTuGSkdigZbfCArZ6rutMj95xCmLJXLg/9jc7P7MhGru5cmTvDQ5uKYJFcANaCo4ZQrpXSQCvNdr9c1zpFagLzgZcqFLbWBUD4AfeBfYfgh+x0IPWHRkZnxeo/eRj3jUtn4krGKKp/yX99ZY3HrqDdu2Jj5hWtl0YrZAMc2Y1JsTQ25J4EBq5cpxCP6n86V12ky5COG/KRqC7twDOJUuXubZxY3exDv2X7Kvbsnx4huZsMMj+f7iurfIx2wFc5760mfXzOfEbMC9SrKr4cv8sXsYC29dga1klG7rr49kpwvMTt+KJCCSnvFteombqTm00XaBDAsgC7icINHG8iw8XqA0MMmG3rUGOqcB0sRp/BVtRIhbw/dRy9DRfD3+hUVyrudp0bQPplwzEVQHuywsEAMKJPqe6iA1kG8HfgAl1GAKVdjP3bHFaRE1RxdMS/C3mN5mEGYvLLdI+Ez1zA0U0HFO070JU19I270BdIXzsHN9mcp+lJXl4MQuOEjL/8ErH8tJqBRhkEIHUSPPb5DPKOcpX3WjE5J7YAhr2MXK2JF3nbhmxlwd0ZsKHGvO0exjtuQ=")

	//staticProvider := credentials.NewStaticCredentialsProvider(
	//	"ASIA6IY35YVFWHLH5X4H",
	//	"+IayElZfudPaDOa79EZTgXK8y+8yZnCOJRzygMmy",
	//	"IQoJb3JpZ2luX2VjEGgaCWV1LXdlc3QtMSJGMEQCIDLQj/jMcZT6PIFRGChu5oAXqEvhJTfU/FRnEkn+PAwwAiABz75tgu0ZkZE4l882xwtjJxlEMhHYmtpTV9lHZ5u9ISrTAwhhEAAaDDk4MDkyMTczMDM3OSIM+RzfaRYRE+o46BmBKrADvB+5Hu8OjBuY07EXNXKuAfYzM+0ZdxElBAJ5tS+HOejAYwKTndeevtjrbEmp5tW+4siUCt7x59j/RNVWJMJqyVNH87yhDw3lodiYFvyNx07eFCPW3cR8JjanFiladc8X2VTbpvFOFDBKyleiAV8fncgPYXzcLyrxs97sg2aGtFg8wKeT8fHgolNAVrAdRJB7DkMEeVlZf0KsnELfCEls3ofVtdwutDsubxcnszSEClZjK4Tx5l0gPvsMPVhKChQ7Bw3jfsy/FAkD4T00ziIgW5PMPNYV9XCRyxKseeSGuV40nmtcKgXcKE3m6jI5Q3owyMW67bpSYQgIVsv2b8tfYDTPfht1/fL36+CG4lIwqnJNyDHeU4qQk3/lMVOmCdkAcdgZhasHRJu9YqS8t8QMvSszMwc8aNVBs/dZANTttmH9lNs9kECA9w8a6OkUtpURS5g6vfoYkGWijKtWR8+Iy3US1VWUZE4s027xcEzCh0cdPl+3uDAtWt6R8v+cuo22it3Oi6GO6HLKEAhKHORybUGGrkeP/HRRyvjsMOgy7Fa56OnjyOO+VqjHwGQpCVLmMN2Q3rUGOp8BAYXkJLfub1+9hQRGRQmsddHUhl1vXWTJltlSNteGe0WStxlpyL9jC2xIrZHuPdEAiEukX9JLqo09p4RsMhEWroPptVblVzWJFs6N9+Tg5vOR9NTYVO3wK70pdfRLj8SvoV5wYyNSRXaoupqoPLRRu2BZmZtrbeQPyOs29O+G/Jup4M1N+lXzz/h25k7H8NxmCK0u8/k20G4aUltE5F9R",
	//)

	//cfg, _ := config.LoadSharedConfigProfile(ctx, "real-world-login") //config.WithRegion("eu-west-1"),
	//config.WithCredentialsProvider(staticProvider),

	//awsCfg, err := config.LoadDefaultConfig(ctx,
	//	config.WithRegion("eu-west-1"),
	//)

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-west-1"),
	)

	//awsCfg, err := config.LoadDefaultConfig(ctx,
	//	config.WithRegion("eu-west-1"),
	//	config.WithCredentialsProvider(staticProvider),
	//)

	if err != nil {
		log.Println("Error loading AWS configuration:", err)
		log.Fatal(err)
	}

	// Create an AWS request Signer and load AWS configuration using default config folder or env vars.
	signer, err := requestsigner.NewSignerWithService(awsCfg, "aoss") // Use "aoss" for Amazon OpenSearch Serverless
	if err != nil {
		log.Println("Error creating request signer:", err)
		log.Fatal(err)
	}

	// Create an opensearch client and use the request-signer.
	client, err := opensearchapi.NewClient(
		opensearchapi.Config{
			Client: opensearch.Config{
				Addresses: []string{endpoint},
				Signer:    signer,
			},
		},
	)
	if err != nil {
		log.Println("Error creating OpenSearch client:", err)
		log.Fatal(err)
	}

	indexName := "go-test-index"

	// Define index mapping.
	mapping := strings.NewReader(`{
	 "settings": {
	   "index": {
	        "number_of_shards": 4
	        }
	      }
	 }`)

	// Create an index with non-default settings.
	createResp, err := client.Indices.Create(
		ctx,
		opensearchapi.IndicesCreateReq{
			Index: indexName,
			Body:  mapping,
		},
	)
	if err != nil {
		log.Println("Error creating index: ", err)
		log.Fatal(err)
	}

	fmt.Println("created index: %s", createResp.Index)

	delResp, err := client.Indices.Delete(ctx, opensearchapi.IndicesDeleteReq{Indices: []string{indexName}})
	if err != nil {
		log.Println("Error deleting index: ", err)
		log.Fatal(err)
	}

	fmt.Println("deleted index: %#v", delResp.Acknowledged)

	return events.APIGatewayProxyResponse{
		Body:       "Atta boy! See the logs!!!",
		StatusCode: 200,
	}, nil
}

func connectsToInternet() events.APIGatewayProxyResponse {
	url := "https://jsonplaceholder.typicode.com/posts/1"

	// Make the GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making request:", err)
		return events.APIGatewayProxyResponse{
			Body:       "Error making request",
			StatusCode: 200,
		}
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: Status code", resp.StatusCode)
		return events.APIGatewayProxyResponse{
			Body:       "Error: Status code",
			StatusCode: 200,
		}
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return events.APIGatewayProxyResponse{
			Body:       "Error reading response body:",
			StatusCode: 200,
		}
	}

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: 200,
	}
}

func main() {
	lambda.Start(Handler)
	//_, _ = localTest()
}
