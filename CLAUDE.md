# relation-service

Follow graph for Snaply — who follows whom. Go 1.22, port 8083, own Postgres DB `relations`. Trusts the `X-User-ID` header set by `api-gateway`.

## Layout

`cmd/main.go` → `internal/handler` (`follow.go`, `routes.go`) → `internal/service` (`follow_service.go`) → `internal/repository` (`postgres/follow.go`) → `internal/model`

## Endpoints

All mounted under `/api/v1/relations/{user_id}`:

| Method | Path        | Description                                        |
|--------|-------------|-----------------------------------------------------|
| POST   | /follow     | Caller (X-User-ID) follows `{user_id}`               |
| DELETE | /follow     | Caller unfollows `{user_id}`                         |
| GET    | /status     | `{following: bool}` — does caller follow `{user_id}` |
| GET    | /counts     | `{followers, following}` counts for `{user_id}`      |
| GET    | /followers  | Cursor-paginated list of follower user IDs           |
| GET    | /following  | Cursor-paginated list of user IDs `{user_id}` follows|
| GET    | /health     | Health check                                         |

## Conventions

- **IDs only, by design.** This service never stores or returns usernames/avatars — only UUIDs. Resolving a list of IDs to display names is the caller's job, via `user-service`'s `POST /api/v1/users/batch`. Keeps the follow graph decoupled from identity data.
- Publishes `user.followed` to Kafka on a successful follow (consumed by `notification-service`), async in a goroutine so a slow broker never delays the follow request; publish failures are logged, not surfaced to the caller. No event is published for unfollow.
- `follows` table: composite PK `(follower_id, followee_id)`, a `CHECK` constraint blocking self-follows, and two covering indexes (`followee_id, created_at, follower_id` and the mirror) for the followers/following cursor queries.
- `Follow()` is idempotent (`ON CONFLICT DO NOTHING`); `Unfollow()` on a non-existent relation is a silent no-op, not an error.
- Cursor pagination matches the rest of the platform: base64(JSON{ca, id}), `{data: [...], next_cursor: "..."}` envelope — except `data` here is a flat list of UUID strings, not objects.

## Running

```bash
cd ../infra && make up
make migrate-up
make run
```

Default `DATABASE_URL`: `postgres://snaply:snaply_secret@localhost:5432/relations?sslmode=disable`
