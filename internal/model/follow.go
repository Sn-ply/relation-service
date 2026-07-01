package model

import (
	"time"

	"github.com/google/uuid"
)

type Follow struct {
	FollowerID uuid.UUID `db:"follower_id" json:"follower_id"`
	FolloweeID uuid.UUID `db:"followee_id" json:"followee_id"`
	CreatedAt  time.Time `db:"created_at"  json:"created_at"`
}
