package api

import "os"

type Env struct {
	Port string
}

func LoadEnv() *Env {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Env{Port: port}
}
