package api

import (
	"os"
	"strconv"
	"time"
)

type ServerConfig struct {
	Port string
}

type Neo4jConfig struct {
	URI      string
	Username string
	Password string
	Realm    string
}

type DynamoDBConfig struct {
	URI string
}

type RedisConfig struct {
	URI      string
	Username string
	Password string
	DB       int
}

type RabbitMQConfig struct {
	URI string
}

func LoadServer() ServerConfig {

	return ServerConfig{
		Port: getEnv("PORT", "8080"),
	}

}

func LoadNeo4j() Neo4jConfig {

	return Neo4jConfig{
		URI:      getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Username: getEnv("NEO4J_USERNAME", "neo4j"),
		Password: getEnv("NEO4J_PASSWORD", "password123"),
		Realm:    getEnv("NEO4J_REALM", ""),
	}

}

func LoadDynamoDB() DynamoDBConfig {
	return DynamoDBConfig{
		URI: getEnv("DYNAMODB_URI", "http://localhost:8000"),
	}
}

func LoadRedis() RedisConfig {
	return RedisConfig{
		URI:      getEnv("REDIS_URI", "localhost:6379"),
		Username: getEnv("REDIS_USERNAME", ""),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getEnvAsInt("REDIS_DB", 0),
	}
}

func LoadRabbitMQ() RabbitMQConfig {
	return RabbitMQConfig{
		URI: getEnv("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvAsFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			return floatVal
		}
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
