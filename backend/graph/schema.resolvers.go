package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/MohsenDowlati/footbook/graph/model"
	"github.com/MohsenDowlati/footbook/internal/service"
	"github.com/google/uuid"
)

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

func (r *mutationResolver) FollowUser(ctx context.Context, followerID string, targetID string) (bool, error) {
	err := r.SocialService.FollowUser(ctx, followerID, targetID)
	if err != nil {
		return false, err
	}
	return true, nil
}

// UnfollowUser لغو دنبال‌کردن در Neo4j و پاکسازی فید
func (r *mutationResolver) UnfollowUser(ctx context.Context, followerID string, targetID string) (bool, error) {
	// ۱. حذف رابطه از گراف Neo4j
	if err := r.SocialService.UnfollowUser(ctx, followerID, targetID); err != nil {
		return false, err
	}

	// ۲. پاکسازی پست‌های کاربر آنفالو شده از فید دنبال‌کننده در DynamoDB و کش Redis
	go func() {
		// برای پاسخ‌دهی بدون معطلی به کلاینت، پاکسازی در بک‌گراند اجرا می‌شود
		_ = r.FeedService.RemovePostsFromUserFeed(context.Background(), followerID, targetID)
	}()

	return true, nil
}
