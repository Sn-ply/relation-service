# relation-service

Follow graph for Snaply — who follows whom. Trusts the `X-User-ID` header set by `api-gateway`.

## Environment Variables

| Variable       | Default                                                                       | Description                    |
|----------------|---------------------------------------------------------------------------------|---------------------------------|
| `DATABASE_URL` | `postgres://snaply:snaply_secret@localhost:5432/relations?sslmode=disable`    | PostgreSQL connection string    |
| `SERVER_PORT`  | `8083`                                                                        | HTTP listen port                |

## Endpoints

| Method | Path                                | Description                                    |
|--------|-------------------------------------|--------------------------------------------------|
| POST   | /api/v1/relations/{user_id}/follow   | Caller follows `user_id`                         |
| DELETE | /api/v1/relations/{user_id}/follow   | Caller unfollows `user_id`                       |
| GET    | /api/v1/relations/{user_id}/status   | `{following: bool}` — does caller follow `user_id`|
| GET    | /api/v1/relations/{user_id}/counts   | `{followers, following}` counts for `user_id`     |
| GET    | /api/v1/relations/{user_id}/followers| Cursor-paginated list of follower user IDs        |
| GET    | /api/v1/relations/{user_id}/following| Cursor-paginated list of user IDs `user_id` follows|
| GET    | /health                              | Health check                                      |

Followers/following/counts return raw user IDs only — resolving them to usernames/avatars is done
by calling `user-service`'s `POST /api/v1/users/batch`.

## Running Locally

```bash
cd ../infra && make up
make migrate-up
make run
```