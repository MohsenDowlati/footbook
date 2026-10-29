package main

import (
	"context"
	"github.com/MohsenDowlati/footbook/internal/repository"
	"github.com/MohsenDowlati/footbook/internal/service"
	"log"
	"net/http"

	"github.com/MohsenDowlati/footbook/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	config "github.com/MohsenDowlati/footbook/internal/api"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clients, err := repository.InitClients(ctx)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer clients.Neo4jDriver.Close(ctx)
	defer clients.RabbitMQ.Close()
	defer clients.RabbitChan.Close()

	socialSvc := service.NewSocialService(clients.Neo4jDriver)
	feedSvc := service.NewFeedService(clients.DynamoDB, clients.Redis, clients.RabbitChan, socialSvc)

	feedSvc.StartFanoutWorker(ctx)

	resolver := &graph.Resolver{
		SocialService: socialSvc,
		FeedService:   feedSvc,
	}
	env := config.LoadServer()

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", env.Port)
	log.Fatal(http.ListenAndServe(":"+env.Port, nil))
}
