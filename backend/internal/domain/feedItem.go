package domain

import "time"

type FeedItem struct {
	UserID    string    `json:"userId"`
	PostID    string    `json:"postId"`
	AuthorID  string    `json:"authorId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}
