package service

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type SocialService struct {
	driver neo4j.DriverWithContext
}

func NewSocialService(driver neo4j.DriverWithContext) *SocialService {
	return &SocialService{driver: driver}
}

func (s *SocialService) FollowUser(ctx context.Context, followerID, targetID string) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MERGE (a:User {id: $followerId})
			MERGE (b:User {id: $targetId})
			MERGE (a)-[r:FOLLOWS]->(b)
			ON CREATE SET r.createdAt = timestamp()
			RETURN r
		`
		params := map[string]any{
			"followerId": followerID,
			"targetId":   targetID,
		}
		_, runErr := tx.Run(ctx, query, params)
		return nil, runErr
	})
	return err
}

func (s *SocialService) UnfollowUser(ctx context.Context, followerID, targetID string) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (a:User {id: $followerId})-[r:FOLLOWS]->(b:User {id: $targetId})
			DELETE r
		`
		params := map[string]any{
			"followerId": followerID,
			"targetId":   targetID,
		}
		_, runErr := tx.Run(ctx, query, params)
		return nil, runErr
	})
	return err
}

func (s *SocialService) GetFollowerIDs(ctx context.Context, userID string) ([]string, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (follower:User)-[:FOLLOWS]->(target:User {id: $userId})
			RETURN follower.id AS followerId
		`
		res, runErr := tx.Run(ctx, query, map[string]any{"userId": userID})
		if runErr != nil {
			return nil, runErr
		}

		var ids []string
		for res.Next(ctx) {
			record := res.Record()
			if val, ok := record.Get("followerId"); ok {
				ids = append(ids, val.(string))
			}
		}
		return ids, res.Err()
	})

	if err != nil {
		return nil, err
	}
	return result.([]string), nil
}
