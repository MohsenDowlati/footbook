package repository

import (
	"context"
	"fmt"
	"github.com/MohsenDowlati/footbook/internal/api"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"log"
)

type Clients struct {
	Neo4jDriver neo4j.DriverWithContext
	DynamoDB    *dynamodb.Client
	Redis       *redis.Client
	RabbitMQ    *amqp.Connection
	RabbitChan  *amqp.Channel
}

func InitClients(ctx context.Context) (*Clients, error) {
	neo4jConfig := api.LoadNeo4j()
	neo4jDriver, err := neo4j.NewDriverWithContext(neo4jConfig.URI, neo4j.BasicAuth(neo4jConfig.Username, neo4jConfig.Password, neo4jConfig.Realm))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to neo4j: %w", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	dynamodbConfig := api.LoadDynamoDB()

	dynamoClient := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = awsPtr(dynamodbConfig.URI)
	})

	redisConfig := api.LoadRedis()

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisConfig.URI,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	ampqConfig := api.LoadRabbitMQ()

	amqpConn, err := amqp.Dial(ampqConfig.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	amqpChan, err := amqpConn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open rabbitmq channel: %w", err)
	}

	_, err = amqpChan.QueueDeclare(
		"feed_fanout_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare rabbitmq queue: %w", err)
	}

	log.Println("Successfully connected to all databases and message broker!")

	return &Clients{
		Neo4jDriver: neo4jDriver,
		DynamoDB:    dynamoClient,
		Redis:       rdb,
		RabbitMQ:    amqpConn,
		RabbitChan:  amqpChan,
	}, nil
}

func awsPtr(s string) *string {
	return &s
}
