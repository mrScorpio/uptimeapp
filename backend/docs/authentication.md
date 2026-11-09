# Authentication API

## Local setup

Requires Go 1.26, Docker and Docker Compose. From `backend/`:

```sh
docker-compose up -d postgres
export DATABASE_URL='postgres://uptimeapp:local-dev-password@localhost:5432/uptimeapp?sslmode=disable'
export JWT_SECRET='replace-this-with-a-random-secret-of-at-least-32-characters'
go run ./cmd/api
```

To start the API and frontend together after starting PostgreSQL and installing frontend dependencies, export `DATABASE_URL` and `JWT_SECRET`, then run `./scripts/dev.sh` from the repository root. The script stops both servers when either exits or when interrupted.

The Compose password is for local development only. Set `POSTGRES_PASSWORD` and a matching `DATABASE_URL` for other environments. The API creates its schema on startup.

Optional settings: `API_ADDR` (default `:8080`), `ALLOWED_ORIGIN` (default `http://localhost:3000`), `COOKIE_SECURE` (default `false`; set `true` with HTTPS), `ACCESS_TTL` (default `15m`), and `REFRESH_TTL` (default `720h`). `DATABASE_URL` and `JWT_SECRET` are required; the secret must contain at least 32 characters. Browser requests from `ALLOWED_ORIGIN` may send credentials.

## Endpoints

All bodies and successful nonempty responses use JSON. Errors use `{"error":"message"}`. The access token is sent as `Authorization: Bearer <token>`.

| Method | Path | Request | Success |
| --- | --- | --- | --- |
| POST | `/api/v1/auth/register` | `{"email":"a@example.com","login":"alice","password":"password123"}` | `201`, access token and user; refresh cookie |
| POST | `/api/v1/auth/login` | `{"identifier":"alice","password":"password123"}`; email is also accepted | `200`, access token and user; refresh cookie |
| POST | `/api/v1/auth/refresh` | Refresh cookie | `200`, new access token and refresh cookie |
| POST | `/api/v1/auth/logout` | Refresh cookie | `204`, cookie removed |
| GET | `/api/v1/users/me` | Bearer access token | `200`, current user |
| PATCH | `/api/v1/users/me` | Bearer access token and `{"name":"Alice"}` | `200`, updated user profile |

Registration requires a valid email, a login of 3–32 lowercase Latin letters, digits, `_` or `-`, and an 8–72 byte password. Email and login are trimmed and stored lowercase. Login errors do not disclose whether an identifier exists. A duplicate email or login yields `409`; malformed input yields `400`; invalid or expired tokens yield `401`. Refresh tokens rotate on each successful refresh, and the previous token cannot be reused. The `HttpOnly` refresh cookie uses `SameSite=Lax` and path `/api/v1/auth`.

The profile name can be updated to a trimmed value containing 1–100 Unicode characters. Registration initializes the name from the login.
