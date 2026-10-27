package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamoTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type PostCreatedEvent struct {
	PostID    string    `json:"postId"`
	AuthorID  string    `json:"authorId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type FeedService struct {
	dynamo   *dynamodb.Client
	redis    *redis.Client
	rabbitCh *amqp.Channel
	social   *SocialService
}

func NewFeedService(
	dynamo *dynamodb.Client,
	redis *redis.Client,
	rabbitCh *amqp.Channel,
	social *SocialService,
) *FeedService {
	return &FeedService{
		dynamo:   dynamo,
		redis:    redis,
		rabbitCh: rabbitCh,
		social:   social,
	}
}

func (f *FeedService) PublishPostCreatedEvent(ctx context.Context, evt PostCreatedEvent) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	return f.rabbitCh.PublishWithContext(ctx,
		"",
		"feed_fanout_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func (f *FeedService) StartFanoutWorker(ctx context.Context) {
	msgs, err := f.rabbitCh.Consume(
		"feed_fanout_queue",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Worker failed to register consumer: %v", err)
	}

	go func() {
		log.Println("Fan-out Worker is listening for new posts...")
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}

				var evt PostCreatedEvent
				if err := json.Unmarshal(d.Body, &evt); err != nil {
					log.Printf("Failed to decode event: %v", err)
					_ = d.Nack(false, false)
					continue
				}

				followers, err := f.social.GetFollowerIDs(ctx, evt.AuthorID)
				if err != nil {
					log.Printf("Error fetching followers from Neo4j: %v", err)
					_ = d.Nack(false, true) // retry
					continue
				}

				for _, followerID := range followers {
					sortKey := fmt.Sprintf("POST#%d#%s", evt.CreatedAt.UnixNano(), evt.PostID)

					_, err := f.dynamo.PutItem(ctx, &dynamodb.PutItemInput{
						TableName: aws.String("FeedsTable"),
						Item: map[string]dynamoTypes.AttributeValue{
							"PK":        &dynamoTypes.AttributeValueMemberS{Value: "USER#" + followerID},
							"SK":        &dynamoTypes.AttributeValueMemberS{Value: sortKey},
							"postId":    &dynamoTypes.AttributeValueMemberS{Value: evt.PostID},
							"authorId":  &dynamoTypes.AttributeValueMemberS{Value: evt.AuthorID},
							"content":   &dynamoTypes.AttributeValueMemberS{Value: evt.Content},
							"createdAt": &dynamoTypes.AttributeValueMemberS{Value: evt.CreatedAt.Format(time.RFC3339)},
						},
					})
					if err != nil {
						log.Printf("Failed to write feed item for follower %s: %v", followerID, err)
						continue
					}

					f.redis.Del(ctx, fmt.Sprintf("feed:%s", followerID))
				}

				_ = d.Ack(false)
			}
		}
	}()
}

func (f *FeedService) RemovePostsFromUserFeed(ctx context.Context, followerID, authorID string) error {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String("FeedsTable"),
		KeyConditionExpression: aws.String("PK = :pk"),
		FilterExpression:       aws.String("authorId = :authorId"),
		ExpressionAttributeValues: map[string]dynamoTypes.AttributeValue{
			":pk":       &dynamoTypes.AttributeValueMemberS{Value: "USER#" + followerID},
			":authorId": &dynamoTypes.AttributeValueMemberS{Value: authorID},
		},
	}

	result, err := f.dynamo.Query(ctx, queryInput)
	if err != nil {
		return fmt.Errorf("failed to query feed items for cleanup: %w", err)
	}

	for _, item := range result.Items {
		sk := item["SK"].(*dynamoTypes.AttributeValueMemberS).Value
		_, err := f.dynamo.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String("FeedsTable"),
			Key: map[string]dynamoTypes.AttributeValue{
				"PK": &dynamoTypes.AttributeValueMemberS{Value: "USER#" + followerID},
				"SK": &dynamoTypes.AttributeValueMemberS{Value: sk},
			},
		})
		if err != nil {
			log.Printf("failed to delete feed item %s: %v", sk, err)
		}
	}

	f.redis.Del(ctx, fmt.Sprintf("feed:%s", followerID))

	return nil
}
