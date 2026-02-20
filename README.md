# Football Manager API

Go + Gin API with PostgreSQL.

## Features

- User auth endpoints (`/api/v1/auth`)
- User endpoint (`/api/v1/users/:id`)
- Registration with email verification code
- Login with bearer token
- Reset password by code
- Auth check endpoint for logged-in user

## Requirements

- Go 1.25+
- PostgreSQL

## Environment

Create `.env` (example):

```env
DB_DSN="host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
LOG_LEVEL=debug
SERVER_PORT=8080
SERVER_CORS_ALLOWED_ORIGINS=http://dashboard.manager.localhost,http://localhost:5173

# Legacy API token auth for protected group/scores endpoints
AUTH_API_TOKENS=token1,token2

# User login token config
AUTH_JWT_SECRET=change-me-in-prod
AUTH_JWT_TTL_MINUTES=60

# Mailer (log | smtp)
MAILER_PROVIDER=smtp
MAILER_FROM=no-reply@manager.localhost
MAILER_HOST=localhost
MAILER_PORT=1025
MAILER_USERNAME=
MAILER_PASSWORD=
MAILER_LOGO_PATH=logo.png
MAILER_ADMIN_EMAILS=admin@manager.localhost
```

## Quick Start

```bash
make migrate
make run
```

Health check:

```bash
curl http://localhost:8080/
```

## Main Endpoints

Base prefix: `/api/v1`

### Auth

- `POST /auth/registration`
- `POST /auth/verify-email`
- `POST /auth/login`
- `POST /auth/reset-password/request`
- `POST /auth/reset-password/confirm`
- `GET /auth/check` (requires `Authorization: Bearer <token>`)

### Users

- `GET /users/me` (requires `Authorization: Bearer <token>`)
- `GET /users/:id`

## Auth Flow Example

### 1. Registration

```bash
curl -X POST http://localhost:8080/api/v1/auth/registration \
  -H 'Content-Type: application/json' \
  -d '{
    "username":"john",
    "full_name":"John Doe",
    "email":"john@example.com",
    "password":"secret123"
  }'
```

After registration verification code is sent through configured sender:
- `MAILER_PROVIDER=log`: code appears in API logs
- `MAILER_PROVIDER=smtp`: email is sent via SMTP (for local dev use MailHog on `localhost:1025`)
- `MAILER_LOGO_PATH`: path to logo file embedded in email template (default: `logo.png`)
- `MAILER_ADMIN_EMAILS`: comma-separated list of admin emails for new registration notifications
- Email templates are stored in `/Users/resoul/Projects/manager/projects/api/internal/usecase/templates/email/{en,ru}`
- Localization is selected from `Accept-Language` (`ru` -> Russian, otherwise English)

### 2. Verify Email

```bash
curl -X POST http://localhost:8080/api/v1/auth/verify-email \
  -H 'Content-Type: application/json' \
  -d '{
    "email":"john@example.com",
    "code":"123456"
  }'
```

### 3. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "email":"john@example.com",
    "password":"secret123"
  }'
```

Response contains `token`.

### 4. Check Auth

```bash
curl http://localhost:8080/api/v1/auth/check \
  -H 'Authorization: Bearer <token>'
```

Expected response:

```json
{"status":"ok"}
```

## Development

```bash
make help
make fmt
make test
```
