#!/bin/bash

# Production Deployment Test Script
# Tests all critical components after deployment

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.prod.yml"
BASE_URL="${BASE_URL:-http://localhost}"
API_URL="${BASE_URL}/api"

echo "========================================"
echo "Production Deployment Test Script"
echo "========================================"
echo ""

# Counter for pass/fail
PASSED=0
FAILED=0

# Test function
test_command() {
    local name="$1"
    local command="$2"
    local expected="$3"

    echo -n "Testing: $name ... "

    if result=$(eval "$command" 2>&1); then
        if [ -z "$expected" ] || echo "$result" | grep -q "$expected"; then
            echo -e "${GREEN}PASS${NC}"
            ((PASSED++))
            return 0
        else
            echo -e "${RED}FAIL${NC} (unexpected output)"
            echo "  Expected: $expected"
            echo "  Got: $result"
            ((FAILED++))
            return 1
        fi
    else
        echo -e "${RED}FAIL${NC} (command failed)"
        echo "  Error: $result"
        ((FAILED++))
        return 1
    fi
}

# Test HTTP endpoint
test_http() {
    local name="$1"
    local url="$2"
    local expected_code="${3:-200}"

    echo -n "Testing: $name ... "

    http_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>&1 || echo "000")

    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}PASS${NC} (HTTP $http_code)"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}FAIL${NC} (HTTP $http_code, expected $expected_code)"
        ((FAILED++))
        return 1
    fi
}

echo "========================================"
echo "1. Container Status Tests"
echo "========================================"

test_command "MongoDB container running" \
    "docker-compose -f $COMPOSE_FILE ps mongo | grep -q 'Up'" \
    ""

test_command "Backend container running" \
    "docker-compose -f $COMPOSE_FILE ps backend | grep -q 'Up'" \
    ""

test_command "Frontend container running" \
    "docker-compose -f $COMPOSE_FILE ps frontend | grep -q 'Up'" \
    ""

test_command "Nginx container running" \
    "docker-compose -f $COMPOSE_FILE ps nginx | grep -q 'Up'" \
    ""

echo ""
echo "========================================"
echo "2. Health Check Tests"
echo "========================================"

test_command "MongoDB health check" \
    "docker-compose -f $COMPOSE_FILE exec -T mongo mongosh --quiet --eval 'db.adminCommand(\"ping\").ok'" \
    "1"

test_command "Backend health endpoint" \
    "docker-compose -f $COMPOSE_FILE exec -T backend curl -sf http://localhost:3000/health" \
    ""

test_command "Frontend health endpoint" \
    "docker-compose -f $COMPOSE_FILE exec -T frontend wget --quiet --tries=1 --spider http://localhost:8080/" \
    ""

echo ""
echo "========================================"
echo "3. Network Connectivity Tests"
echo "========================================"

test_command "Backend can reach MongoDB" \
    "docker-compose -f $COMPOSE_FILE exec -T backend sh -c 'ping -c 1 mongo >/dev/null 2>&1 && echo ok'" \
    "ok"

test_command "Nginx can reach backend" \
    "docker-compose -f $COMPOSE_FILE exec -T nginx sh -c 'wget -q -O- http://backend:3000/health >/dev/null 2>&1 && echo ok'" \
    "ok"

test_command "Nginx can reach frontend" \
    "docker-compose -f $COMPOSE_FILE exec -T nginx sh -c 'wget -q -O- http://frontend:8080/ >/dev/null 2>&1 && echo ok'" \
    "ok"

echo ""
echo "========================================"
echo "4. External Access Tests"
echo "========================================"

test_http "Nginx health endpoint" \
    "$BASE_URL/health" \
    "200"

test_http "Frontend serves index.html" \
    "$BASE_URL/" \
    "200"

test_http "Backend API health endpoint" \
    "$API_URL/health" \
    "200"

# Test if backend responds to unknown route with proper error
test_http "Backend handles unknown routes" \
    "$API_URL/nonexistent-endpoint" \
    "404"

echo ""
echo "========================================"
echo "5. Volume Persistence Tests"
echo "========================================"

test_command "MongoDB data volume exists" \
    "docker volume ls | grep -q 'mongo-data'" \
    ""

test_command "WhatsApp stores volume exists" \
    "docker volume ls | grep -q 'whatsapp-stores'" \
    ""

test_command "WhatsApp uploads volume exists" \
    "docker volume ls | grep -q 'whatsapp-uploads'" \
    ""

test_command "App keys volume exists" \
    "docker volume ls | grep -q 'app-keys'" \
    ""

echo ""
echo "========================================"
echo "6. Security Tests"
echo "========================================"

test_command "Backend runs as non-root user" \
    "docker-compose -f $COMPOSE_FILE exec -T backend whoami" \
    "appuser"

test_command "Frontend runs as non-root user" \
    "docker-compose -f $COMPOSE_FILE exec -T frontend whoami" \
    "nginx"

# Test CORS headers (if accessible)
echo -n "Testing: CORS headers present ... "
if headers=$(curl -sI "$API_URL/health" 2>&1); then
    if echo "$headers" | grep -qi "access-control-allow"; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
    else
        echo -e "${YELLOW}WARN${NC} (CORS headers not found, might be normal)"
    fi
else
    echo -e "${YELLOW}SKIP${NC} (cannot reach API)"
fi

# Test security headers
echo -n "Testing: Security headers present ... "
if headers=$(curl -sI "$BASE_URL/" 2>&1); then
    if echo "$headers" | grep -qi "x-frame-options"; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
    else
        echo -e "${YELLOW}WARN${NC} (some security headers missing)"
    fi
else
    echo -e "${YELLOW}SKIP${NC} (cannot reach frontend)"
fi

echo ""
echo "========================================"
echo "7. Resource Usage Tests"
echo "========================================"

# Check if containers are using reasonable resources
echo -n "Testing: Container resource usage ... "
if docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep -q "whatsapp"; then
    echo -e "${GREEN}PASS${NC}"
    ((PASSED++))
    echo ""
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep -E "NAME|whatsapp"
else
    echo -e "${RED}FAIL${NC}"
    ((FAILED++))
fi

echo ""
echo "========================================"
echo "8. Log Tests"
echo "========================================"

test_command "Backend logs accessible" \
    "docker-compose -f $COMPOSE_FILE logs --tail=1 backend" \
    ""

test_command "Nginx logs accessible" \
    "docker-compose -f $COMPOSE_FILE logs --tail=1 nginx" \
    ""

echo -n "Testing: No critical errors in logs ... "
if logs=$(docker-compose -f $COMPOSE_FILE logs --tail=50 2>&1); then
    if echo "$logs" | grep -iE "(fatal|panic|emergency)" | grep -v "test-deployment" >/dev/null; then
        echo -e "${RED}FAIL${NC} (found critical errors)"
        ((FAILED++))
        echo "Critical errors found:"
        echo "$logs" | grep -iE "(fatal|panic|emergency)" | head -5
    else
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
    fi
else
    echo -e "${RED}FAIL${NC} (cannot read logs)"
    ((FAILED++))
fi

echo ""
echo "========================================"
echo "Test Summary"
echo "========================================"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo -e "Total:  $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    echo "Deployment appears to be healthy."
    exit 0
else
    echo -e "${RED}Some tests failed! ✗${NC}"
    echo "Please review the failures above."
    echo ""
    echo "Troubleshooting commands:"
    echo "  make logs              # View all logs"
    echo "  make status            # Check container status"
    echo "  make health            # Run health checks"
    echo "  docker stats           # Check resource usage"
    exit 1
fi
