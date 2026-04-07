# API Gateway Security Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement strict CORS policies and rate limiting on the API Gateway.

**Architecture:** Add `github.com/go-chi/httprate` for rate limiting and refine the existing `github.com/go-chi/cors` configuration in `services/api-gateway/main.go`. Use environment variables for configurable limits and origins.

**Tech Stack:** Go, Chi v5, httprate, CORS middleware.

---

### Task 1: Setup Dependencies and Environment

**Files:**
- Modify: `go.mod`
- Modify: `.env.example`

- [ ] **Step 1: Add httprate dependency**
Run: `go get github.com/go-chi/httprate`

- [ ] **Step 2: Update .env.example with security configurations**
```env
# CORS
CORS_ALLOWED_ORIGINS=http://localhost:5173

# Rate Limiting
RATE_LIMIT_GLOBAL_REQUESTS=100
RATE_LIMIT_GLOBAL_WINDOW=1m
RATE_LIMIT_AUTH_REQUESTS=5
RATE_LIMIT_AUTH_WINDOW=1m
```

- [ ] **Step 3: Commit**
```bash
git add go.mod go.sum .env.example
git commit -m "chore: add httprate dependency and security env variables"
```

---

### Task 2: Implement Strict CORS Policy

**Files:**
- Modify: `services/api-gateway/main.go`

- [ ] **Step 1: Update CORS configuration**
Refine `cors.Options` in `main.go` to be stricter.

```go
// Replace existing CORS setup
r.Use(cors.Handler(cors.Options{
    AllowedOrigins:   corsAllowedOrigins,
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
    ExposedHeaders:   []string{"Link"},
    AllowCredentials: true,
    MaxAge:           300,
}))
```

- [ ] **Step 2: Verify CORS configuration compiles**
Run: `go build -o /dev/null ./services/api-gateway/main.go`

- [ ] **Step 3: Commit**
```bash
git add services/api-gateway/main.go
git commit -m "feat: implement strict CORS policy"
```

---

### Task 3: Implement Global Rate Limiting

**Files:**
- Modify: `services/api-gateway/main.go`

- [ ] **Step 1: Read Rate Limit environment variables**
```go
// In main()
globalLimitReqs := getEnvAsInt("RATE_LIMIT_GLOBAL_REQUESTS", 100)
globalLimitWindow, _ := time.ParseDuration(getEnv("RATE_LIMIT_GLOBAL_WINDOW", "1m"))
```
*Note: I will need to add helper functions `getEnv` and `getEnvAsInt` or use existing ones if available.*

- [ ] **Step 2: Apply global rate limiter middleware**
```go
r.Use(httprate.LimitByIP(globalLimitReqs, globalLimitWindow))
```

- [ ] **Step 3: Commit**
```bash
git add services/api-gateway/main.go
git commit -m "feat: implement global rate limiting"
```

---

### Task 4: Implement Auth-specific Rate Limiting

**Files:**
- Modify: `services/api-gateway/main.go`

- [ ] **Step 1: Configure auth rate limit**
```go
authLimitReqs := getEnvAsInt("RATE_LIMIT_AUTH_REQUESTS", 5)
authLimitWindow, _ := time.ParseDuration(getEnv("RATE_LIMIT_AUTH_WINDOW", "1m"))
```

- [ ] **Step 2: Apply to auth routes**
```go
authRouter := chi.NewRouter()
authRouter.Use(httprate.LimitByIP(authLimitReqs, authLimitWindow))
routes.SetupAuthRoutes(authRouter, authHandler)
r.Mount("/auth", authRouter)
```

- [ ] **Step 3: Verify all changes compile**
Run: `go build -o /dev/null ./services/api-gateway/main.go`

- [ ] **Step 4: Commit**
```bash
git add services/api-gateway/main.go
git commit -m "feat: implement auth-specific rate limiting"
```

---

### Task 5: Final Verification

- [ ] **Step 1: Create a verification script `scripts/verify_security.sh`**
```bash
#!/bin/bash
# Test CORS
curl -v -X OPTIONS http://localhost:8080/health \
  -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: GET"

# Test Rate Limiting (this might be slow, so maybe just a few requests)
for i in {1..10}; do
  curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/health
done
```

- [ ] **Step 2: Run verification (requires running gateway)**
*Manual verification recommended if automated environment is not available.*

- [ ] **Step 3: Commit verification script**
```bash
git add scripts/verify_security.sh
git commit -m "test: add security verification script"
```
