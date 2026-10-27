package main

import (
	"log"
	"net/http"

	"github.com/MohsenDowlati/footbook/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	config "github.com/MohsenDowlati/footbook/internal/api"
)

func main() {
	env := config.LoadServer()

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", env.Port)
	log.Fatal(http.ListenAndServe(":"+env.Port, nil))
}
