package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/snaply/relation-service/internal/repository"
	"go.uber.org/zap"
)

var ErrCannotFollowSelf = errors.New("cannot follow yourself")

const topicUserFollowed = "user.followed"

type Page struct {
	UserIDs    []uuid.UUID
	NextCursor string
}

type Counts struct {
	Followers int `json:"followers"`
	Following int `json:"following"`
}

type FollowService interface {
	Follow(ctx context.Context, followerID, followeeID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error
	Status(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error)
	Counts(ctx context.Context, userID uuid.UUID) (*Counts, error)
	Followers(ctx context.Context, userID uuid.UUID, cursor string, limit int) (*Page, error)
	Following(ctx context.Context, userID uuid.UUID, cursor string, limit int) (*Page, error)
}

type followService struct {
	follows repository.FollowRepository
	kafka   *kafka.Writer
	log     *zap.Logger
}

func NewFollowService(follows repository.FollowRepository, kafkaWriter *kafka.Writer, log *zap.Logger) FollowService {
	return &followService{follows: follows, kafka: kafkaWriter, log: log}
}

func (s *followService) Follow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	if followerID == followeeID {
		return ErrCannotFollowSelf
	}
	if err := s.follows.Create(ctx, followerID, followeeID); err != nil {
		return fmt.Errorf("creating follow: %w", err)
	}

	// Publish async — a slow Kafka broker must never delay the follow request itself.
	go s.publishFollowed(followerID, followeeID)

	return nil
}

func (s *followService) publishFollowed(followerID, followeeID uuid.UUID) {
	if s.kafka == nil {
		return
	}
	event := map[string]any{
		"follower_id": followerID,
		"followed_id": followeeID,
		"timestamp":   time.Now().UTC(),
	}
	data, err := json.Marshal(event)
	if err != nil {
		s.log.Warn("failed to marshal user.followed event", zap.Error(err))
		return
	}
	if err := s.kafka.WriteMessages(context.Background(), kafka.Message{
		Topic: topicUserFollowed,
		Key:   []byte(followerID.String() + ":" + followeeID.String()),
		Value: data,
	}); err != nil {
		s.log.Warn("failed to publish user.followed event", zap.Error(err))
	}
}

func (s *followService) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	if err := s.follows.Delete(ctx, followerID, followeeID); err != nil {
		return fmt.Errorf("deleting follow: %w", err)
	}
	return nil
}

func (s *followService) Status(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error) {
	exists, err := s.follows.Exists(ctx, followerID, followeeID)
	if err != nil {
		return false, fmt.Errorf("checking follow status: %w", err)
	}
	return exists, nil
}

func (s *followService) Counts(ctx context.Context, userID uuid.UUID) (*Counts, error) {
	followers, err := s.follows.CountFollowers(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("counting followers: %w", err)
	}
	following, err := s.follows.CountFollowing(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("counting following: %w", err)
	}
	return &Counts{Followers: followers, Following: following}, nil
}

func (s *followService) Followers(ctx context.Context, userID uuid.UUID, cursorStr string, limit int) (*Page, error) {
	return s.list(ctx, userID, cursorStr, limit, s.follows.ListFollowers)
}

func (s *followService) Following(ctx context.Context, userID uuid.UUID, cursorStr string, limit int) (*Page, error) {
	return s.list(ctx, userID, cursorStr, limit, s.follows.ListFollowing)
}

func (s *followService) list(
	ctx context.Context,
	userID uuid.UUID,
	cursorStr string,
	limit int,
	fn func(context.Context, uuid.UUID, *repository.Cursor, int) ([]uuid.UUID, *repository.Cursor, error),
) (*Page, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	var cur *repository.Cursor
	if cursorStr != "" {
		decoded, err := decodeCursor(cursorStr)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		cur = decoded
	}

	ids, nextCur, err := fn(ctx, userID, cur, limit)
	if err != nil {
		return nil, fmt.Errorf("listing relations: %w", err)
	}

	page := &Page{UserIDs: ids}
	if page.UserIDs == nil {
		page.UserIDs = []uuid.UUID{}
	}
	if nextCur != nil {
		page.NextCursor, err = encodeCursor(nextCur)
		if err != nil {
			s.log.Warn("failed to encode cursor", zap.Error(err))
		}
	}
	return page, nil
}

type cursorPayload struct {
	CreatedAt time.Time `json:"ca"`
	ID        string    `json:"id"`
}

func encodeCursor(c *repository.Cursor) (string, error) {
	b, err := json.Marshal(cursorPayload{CreatedAt: c.CreatedAt, ID: c.ID.String()})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func decodeCursor(s string) (*repository.Cursor, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var p cursorPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, err
	}
	return &repository.Cursor{CreatedAt: p.CreatedAt, ID: id}, nil
}
