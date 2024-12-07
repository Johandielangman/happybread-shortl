// ~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~
//      /\_/\
//     ( o.o )
//      > ^ <
//
// Author: Johan Hanekom
// Date: December 2024
//
// ~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~

// ========================= // NEW // =================================

package main

import (
	"context"
	"crypto/sha256"
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
	redisUrl string
)

const (
	redisURLEnv = "REDIS"
)

// =========== // DEFINE STRUCTS // ===========

type Response struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

type Input struct {
	Link string `json:"link"`
}

// =========== // DEFINE LAMBDA INIT // ===========

func init() {
	log.Println("Initializing")
	redisUrl = os.Getenv(redisURLEnv)
	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		log.Println(err)
	}

	client = redis.NewClient(opt)
}

// =========== // DEFINE UTIL FUNCS // ===========

func hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// =========== // DEFINE LAMBDA HANDLER // ===========

func handleRequest(ctx context.Context, event json.RawMessage) (Response, error) {
	log.Println("Starting new lambda")

	// =========== // PARSE INCOMING REQUEST EVENT // ===========
	// The gateway has to be set up to receive the event as a json only!
	var I Input
	err := json.Unmarshal([]byte(event), &I)
	if err != nil {
		return Response{}, err
	}

	// =========== // CREATE A NEW HAS FOR THE LINK // ===========
	h := hash(I.Link)[:5]
	log.Printf("Link shortened: %q -> %q", I.Link, h)

	// =========== // UPLOAD THE LINK TO REDIS // ===========
	err = client.Get(ctx, h).Err()
	if err == redis.Nil {
		err = client.Set(ctx, h, I.Link, 0).Err()
		if err != nil {
			log.Printf("Failed to add %q to Redis", h)
			return Response{}, err
		}
		log.Printf("Successfully added %q to Redis! Letz go!", h)
		return Response{
			StatusCode: 200,
			Body:       h,
		}, nil
	} else if err != nil {
		log.Println(err)
		return Response{}, err
	}

	log.Printf("Would you believe it? The same hash is already there!")
	return Response{
		StatusCode: 200,
		Body:       h,
	}, nil
}

// =========== // GO ENTRYPOINT // ===========

func main() {
	lambda.Start(handleRequest)
}
