#!/bin/bash

# Configuration
API_URL=${API_URL:-"http://localhost:8080"}
ALLOWED_ORIGIN=${ALLOWED_ORIGIN:-"http://localhost:5173"}
DISALLOWED_ORIGIN="http://malicious-site.com"

echo "=== API Gateway Security Verification ==="
echo "Target URL: $API_URL"
echo ""

# 1. Test CORS
echo "--- Testing CORS ---"

echo "Testing Allowed Origin: $ALLOWED_ORIGIN"
curl -s -v -X OPTIONS "$API_URL/health" \
  -H "Origin: $ALLOWED_ORIGIN" \
  -H "Access-Control-Request-Method: GET" 2>&1 | grep -E "Access-Control-Allow-Origin|HTTP/"

echo ""
echo "Testing Disallowed Origin: $DISALLOWED_ORIGIN"
curl -s -v -X OPTIONS "$API_URL/health" \
  -H "Origin: $DISALLOWED_ORIGIN" \
  -H "Access-Control-Request-Method: GET" 2>&1 | grep -E "Access-Control-Allow-Origin|HTTP/"

echo ""

# 2. Test Global Rate Limiting
echo "--- Testing Global Rate Limiting (Target: /health) ---"
echo "Making 10 requests to /health..."
for i in {1..10}; do
  CODE=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/health")
  echo "Request $i: $CODE"
  if [ "$CODE" == "429" ]; then
    echo "Rate limit exceeded at request $i"
    break
  fi
done

echo ""

# 3. Test Auth Rate Limiting (Stricter: 5 req/min)
echo "--- Testing Auth Rate Limiting (Target: /auth/login) ---"
echo "Making 10 requests to /auth/login..."
for i in {1..10}; do
  # We use POST for login, even if it fails due to no body, it should hit the rate limiter
  CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/auth/login")
  echo "Request $i: $CODE"
  if [ "$CODE" == "429" ]; then
    echo "Auth rate limit exceeded at request $i"
    break
  fi
done

echo ""
echo "Verification complete."
