package domain

import "time"

type Post struct {
	ID             string    `json:"id"`
	UserID         string    `json:"userId"`
	Content        string    `json:"content"`
	MediaURL       *string   `json:"mediaUrl,omitempty"`
	OriginalPostID *string   `json:"originalPostId,omitempty"`
	LikesCount     int       `json:"likesCount"`
	RepostsCount   int       `json:"repostsCount"`
	CommentsCount  int       `json:"commentsCount"`
	CreatedAt      time.Time `json:"createdAt"`
}
