package graph

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"time"

	"github.com/MohsenDowlati/footbook/graph/model"
	"github.com/MohsenDowlati/footbook/internal/service"
)

type Resolver struct {
	SocialService *service.SocialService
	FeedService   *service.FeedService
}

// CreatePost is the resolver for the createPost field.
func (r *mutationResolver) CreatePost(ctx context.Context, userID string, content string, mediaURL *string) (*model.Post, error) {
	postID := uuid.New().String()
	now := time.Now().UTC()

	err := r.FeedService.PublishPostCreatedEvent(ctx, service.PostCreatedEvent{
		PostID:    postID,
		AuthorID:  userID,
		Content:   content,
		CreatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to publish post event: %w", err)
	}

	return &model.Post{
		ID:            postID,
		UserID:        userID,
		Content:       content,
		MediaURL:      mediaURL,
		LikesCount:    0,
		RepostsCount:  0,
		CommentsCount: 0,
		CreatedAt:     now,
	}, nil
}

// LikePost is the resolver for the likePost field.
func (r *mutationResolver) LikePost(ctx context.Context, userID string, postID string) (bool, error) {
	panic("not implemented")
}

// AddComment is the resolver for the addComment field.
func (r *mutationResolver) AddComment(ctx context.Context, userID string, postID string, content string) (*model.Comment, error) {
	panic("not implemented")
}

// FollowUser is the resolver for the followUser field.
func (r *mutationResolver) FollowUser(ctx context.Context, followerID string, targetID string) (bool, error) {
	err := r.SocialService.FollowUser(ctx, followerID, targetID)
	if err != nil {
		return false, err
	}
	return true, nil
}

// UnfollowUser is the resolver for the unfollowUser field.
func (r *mutationResolver) UnfollowUser(ctx context.Context, followerID string, targetID string) (bool, error) {
	if err := r.SocialService.UnfollowUser(ctx, followerID, targetID); err != nil {
		return false, err
	}

	go func() {
		_ = r.FeedService.RemovePostsFromUserFeed(context.Background(), followerID, targetID)
	}()

	return true, nil
}

// GetFeed is the resolver for the getFeed field.
func (r *queryResolver) GetFeed(ctx context.Context, userID string) ([]*model.Post, error) {
	panic("not implemented")
}

// GetPostComments is the resolver for the getPostComments field.
func (r *queryResolver) GetPostComments(ctx context.Context, postID string) ([]*model.Comment, error) {
	panic("not implemented")
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type (
	mutationResolver struct{ *Resolver }
	queryResolver    struct{ *Resolver }
)

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
/*
	type Resolver struct {
	SocialService *service.SocialService
	FeedService   *service.FeedService
}
*/
