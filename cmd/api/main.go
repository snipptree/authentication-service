package main

import (
	"authentication/data"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	mongoURL = "mongodb+srv://REDACTED_USER:REDACTED_PASSWORD@snipptree-dev.cza8e.mongodb.net/?retryWrites=true&w=majority&appName=snipptree-dev"
	gRPCPort = "5001"
)

type Config struct {
	Repo data.Repository
}

func main() {
	log.Printf("starting authentication service on PORT %s", webPort)

	// connect to mongo
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{}
	app.setupRepo(mongoClient)

	app.Serve()
}

func (app *Config) Serve() {
	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func connectToMongo() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoURL)

	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("error connceting:", err)
		return nil, err
	}
	log.Println("connected to mongo db")
	return c, nil
}

func (app *Config) setupRepo(client *mongo.Client) {
	authRepo := data.NewAuthRepository(client)
	app.Repo = authRepo
}
