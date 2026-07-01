CREATE TABLE follows (
    follower_id UUID        NOT NULL,
    followee_id UUID        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, followee_id),
    CONSTRAINT no_self_follow CHECK (follower_id <> followee_id)
);

CREATE INDEX idx_follows_followee ON follows (followee_id, created_at, follower_id);
CREATE INDEX idx_follows_follower ON follows (follower_id, created_at, followee_id);
