package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type Cursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type FollowRepository interface {
	Create(ctx context.Context, followerID, followeeID uuid.UUID) error
	Delete(ctx context.Context, followerID, followeeID uuid.UUID) error
	Exists(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error)
	CountFollowers(ctx context.Context, userID uuid.UUID) (int, error)
	CountFollowing(ctx context.Context, userID uuid.UUID) (int, error)
	ListFollowers(ctx context.Context, userID uuid.UUID, cursor *Cursor, limit int) ([]uuid.UUID, *Cursor, error)
	ListFollowing(ctx context.Context, userID uuid.UUID, cursor *Cursor, limit int) ([]uuid.UUID, *Cursor, error)
}
