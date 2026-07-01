package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/snaply/relation-service/internal/repository"
)

type followRepo struct {
	db *sqlx.DB
}

func NewFollowRepository(db *sqlx.DB) repository.FollowRepository {
	return &followRepo{db: db}
}

func (r *followRepo) Create(ctx context.Context, followerID, followeeID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO follows (follower_id, followee_id)
		VALUES ($1, $2)
		ON CONFLICT (follower_id, followee_id) DO NOTHING`,
		followerID, followeeID)
	return err
}

func (r *followRepo) Delete(ctx context.Context, followerID, followeeID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM follows WHERE follower_id = $1 AND followee_id = $2`,
		followerID, followeeID)
	return err
}

func (r *followRepo) Exists(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = $2)`,
		followerID, followeeID)
	return exists, err
}

func (r *followRepo) CountFollowers(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM follows WHERE followee_id = $1`, userID)
	return count, err
}

func (r *followRepo) CountFollowing(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM follows WHERE follower_id = $1`, userID)
	return count, err
}

func (r *followRepo) ListFollowers(ctx context.Context, userID uuid.UUID, cursor *repository.Cursor, limit int) ([]uuid.UUID, *repository.Cursor, error) {
	return r.list(ctx, "followee_id", "follower_id", userID, cursor, limit)
}

func (r *followRepo) ListFollowing(ctx context.Context, userID uuid.UUID, cursor *repository.Cursor, limit int) ([]uuid.UUID, *repository.Cursor, error) {
	return r.list(ctx, "follower_id", "followee_id", userID, cursor, limit)
}

type followRow struct {
	ID        uuid.UUID `db:"id"`
	CreatedAt time.Time `db:"created_at"`
}

// list is shared by ListFollowers/ListFollowing: fixedCol is the column pinned to userID,
// selectCol is the column whose values are returned (the "other side" of the relationship).
// Both column names are internal literals from the two call sites above, never user input.
func (r *followRepo) list(ctx context.Context, fixedCol, selectCol string, userID uuid.UUID, cursor *repository.Cursor, limit int) ([]uuid.UUID, *repository.Cursor, error) {
	base := `SELECT ` + selectCol + ` AS id, created_at FROM follows WHERE ` + fixedCol + ` = $1`

	var (
		query string
		args  []interface{}
	)
	if cursor == nil {
		query = base + ` ORDER BY created_at ASC, ` + selectCol + ` ASC LIMIT $2`
		args = []interface{}{userID, limit + 1}
	} else {
		query = base + ` AND (created_at, ` + selectCol + `) > ($2, $3) ORDER BY created_at ASC, ` + selectCol + ` ASC LIMIT $4`
		args = []interface{}{userID, cursor.CreatedAt, cursor.ID, limit + 1}
	}

	rows := []*followRow{}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, nil, err
	}

	var nextCursor *repository.Cursor
	if len(rows) > limit {
		last := rows[limit-1]
		nextCursor = &repository.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
		rows = rows[:limit]
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}

	return ids, nextCursor, nil
}
