// ~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~
//      /\_/\
//     ( o.o )
//      > ^ <
//
// Author: Johan Hanekom
// Date: December 2024
//
// ~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~

// ========================= // GET // =================================

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-redis/redis/v8"
)

// =========== // DEFINE CONSTANTS AND VARIABLES // ===========

var (
	client   *redis.Client
	headers  map[string]string
	redisUrl string
)

const (
	redisURLEnv        = "REDIS"
	notFoundResponse   = `<!DOCTYPE html><html><body><h3>Link not found, try another code</h3></body></html>`
	badRequestResponse = `<!DOCTYPE html><html><body><h3>That link format is not correct</h3></body></html>`

	// There the redirect magic happens!
	// https://www.w3schools.com/tags/att_meta_http_equiv.asp
	redirectResponse = `<!DOCTYPE html><html><head><meta http-equiv="refresh" content="0; url=%s"></head></html>`
)

// =========== // DEFINE STRUCTS // ===========

type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// See https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html
// ironic, this link is too long for me... would be nice if there was a shorter link! ha ha
type APIGatewayProxyRequest struct {
	PathParameters map[string]string `json:"pathParameters"`
	HTTPMethod     string            `json:"httpMethod"`
	Path           string            `json:"path"`
	RequestContext RequestContext    `json:"requestContext"`
}

type RequestContext struct {
	Identity Identity `json:"identity"`
}

type Identity struct {
	SourceIP  string `json:"sourceIp"`
	UserAgent string `json:"userAgent"`
}

// =========== // DEFINE LAMBDA INIT // ===========

func init() {
	log.Println("Initializing Lambda")
	redisUrl = os.Getenv(redisURLEnv)

	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		log.Println(err)
	}

	client = redis.NewClient(opt)

	headers = make(map[string]string)
	headers["Content-Type"] = "text/html"
}

// =========== // DEFINE LAMBDA HANDLER // ===========

func handleRequest(ctx context.Context, event json.RawMessage) (Response, error) {
	log.Println("Starting get lambda")

	// =========== // PARSE INCOMING REQUEST EVENT // ===========
	var request APIGatewayProxyRequest
	err := json.Unmarshal([]byte(event), &request)
	if err != nil {
		log.Println(err)
		return Response{}, err
	}
	sourceIP := request.RequestContext.Identity.SourceIP

	// =========== // EXTRACT PATH PARAMETER // ===========
	link, exists := request.PathParameters["link"]
	if !exists {
		log.Println("Could not find the link path parameter")
		return Response{
			StatusCode: 400,
			Body:       badRequestResponse,
			Headers:    headers,
		}, nil
	}
	log.Printf("[%v] Request for %v", sourceIP, link)

	// =========== // GET LINK FROM REDIS // ===========
	res, err := client.Get(ctx, link).Result()
	if err == redis.Nil {
		log.Printf("Uh.. oh.. %q could not be found in Redis!", link)
		return Response{
			StatusCode: 404,
			Body:       notFoundResponse,
			Headers:    headers,
		}, nil
	} else if err != nil {
		log.Println(err)
		return Response{}, nil
	}

	log.Printf("Found it! %q points to %q. Redirecting user.", link, res)
	return Response{
		StatusCode: 200,
		Body:       fmt.Sprintf(redirectResponse, res),
		Headers:    headers,
	}, nil
}

// =========== // GO ENTRYPOINT // ===========

func main() {
	lambda.Start(handleRequest)
}
