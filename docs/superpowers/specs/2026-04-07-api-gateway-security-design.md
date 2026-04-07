# API Gateway Security Design Spec

**Goal:** Implement strict CORS policies and rate limiting on the API Gateway to protect against unauthorized cross-origin requests and brute-force/DoS attacks.

## Architecture
The security features will be implemented as middleware within the `api-gateway` service, which uses the `go-chi/chi/v5` router.

### 1. Strict CORS Policy
- **Library:** `github.com/go-chi/cors` (already in use).
- **Configuration:**
  - `AllowedOrigins`: Restricted to `CORS_ALLOWED_ORIGINS` environment variable. Default: `http://localhost:5173`.
  - `AllowedMethods`: `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`.
  - `AllowedHeaders`: `Accept`, `Authorization`, `Content-Type`, `X-CSRF-Token`.
  - `ExposedHeaders`: `Link`.
  - `AllowCredentials`: `true` (required for cookie-based authentication).
  - `MaxAge`: `300` seconds.

### 2. Rate Limiting
- **Library:** `github.com/go-chi/httprate`.
- **Global Rate Limit:**
  - **Limit:** 100 requests per minute per IP address.
  - **Applied to:** All routes under `/api` and `/auth`.
  - **Configurable via:** `RATE_LIMIT_GLOBAL_REQUESTS` (default: 100) and `RATE_LIMIT_GLOBAL_WINDOW` (default: 1m).
- **Auth Rate Limit (Stricter):**
  - **Limit:** 5 requests per minute per IP address.
  - **Applied to:** `/auth/login`, `/auth/register`.
  - **Configurable via:** `RATE_LIMIT_AUTH_REQUESTS` (default: 5) and `RATE_LIMIT_AUTH_WINDOW` (default: 1m).

## Configuration (Environment Variables)
Add the following to `.env.example` and the service configuration:
- `CORS_ALLOWED_ORIGINS`: Comma-separated list of allowed origins.
- `RATE_LIMIT_GLOBAL_REQUESTS`: Number of requests for the global limit.
- `RATE_LIMIT_GLOBAL_WINDOW`: Time window for the global limit (e.g., "1m").
- `RATE_LIMIT_AUTH_REQUESTS`: Number of requests for the auth limit.
- `RATE_LIMIT_AUTH_WINDOW`: Time window for the auth limit (e.g., "1m").

## Testing Strategy
- **Unit Tests:** Verify middleware configuration in `main.go`.
- **Integration Tests:** Use `curl` or a script to verify:
  - CORS headers are present and correct for allowed/disallowed origins.
  - Rate limiting triggers `429 Too Many Requests` after exceeding limits.
