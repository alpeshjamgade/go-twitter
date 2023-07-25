package main

import (
	"authentication-service/data"
	"fmt"
	"log"
	"net/http"
)

const (
	redisPort = "6379"
	webPort   = "80"
)

type Config struct {
	Cache *data.Cache
}

func main() {
	log.Println("Starting authentication service ...")

	// connect to redis
	cache := data.NewCache(fmt.Sprintf("redis:%s", redisPort))
	err := cache.Connect()
	if err != nil {
		log.Fatalf("Error while connecting to redis, %s", err)
	}

	// Set up config
	app := Config{
		Cache: cache,
	}

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start the server
	err = srv.ListenAndServe()
	if err != nil {
		log.Panicln(err)
	}
}
