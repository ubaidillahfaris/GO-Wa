# Localhost Setup Guide

Panduan cepat untuk deploy ke localhost (HTTP only, tanpa SSL/domain).

## Quick Start

```bash
# 1. Copy environment file
cp .env.localhost.example .env.localhost

# 2. Build & deploy
COMPOSE_FILE=docker-compose.localhost.yml make build
COMPOSE_FILE=docker-compose.localhost.yml make up

# 3. Check status
COMPOSE_FILE=docker-compose.localhost.yml make status
```

Atau lebih simple:

```bash
# Set default to localhost
export COMPOSE_FILE=docker-compose.localhost.yml

# Then use normal commands
make build
make up
make status
```

---

## Access Points

### Local Development
- **Frontend**: http://localhost
- **API**: http://localhost/api
- **Health Check**: http://localhost/health
- **MongoDB**: localhost:27017 (exposed for debugging)

---

## Configuration

### Environment (.env.localhost)

```bash
# MongoDB
MONGO_USER=admin
MONGO_PASS=password

# CORS - allow localhost
CORS_ALLOWED_ORIGIN=http://localhost

# JWT (use simple key for dev)
JWT_SECRET=dev-secret-key-change-in-production

# Port
HTTP_PORT=80
```

### Key Differences from Production

| Feature | Production | Localhost |
|---------|-----------|-----------|
| SSL/HTTPS | Required | Disabled (HTTP only) |
| Domain | Required | localhost |
| MongoDB Port | Internal only | Exposed (27017) |
| CORS | Strict domain | http://localhost |
| SSL Cert | Let's Encrypt | Not needed |
| Port 443 | HTTPS | Not used |

---

## Architecture

```
Browser (localhost)
    ↓
Nginx:80 (HTTP only)
    ├─> Frontend:8080 (internal)
    └─> Backend:3000 (internal)
        └─> MongoDB:27017 (exposed for debugging)
```

**No SSL redirect, no certificate needed!**

---

## Docker Compose Files

### Production
```bash
docker-compose -f docker-compose.prod.yml up -d
# Uses: nginx.prod.conf (with SSL)
```

### Localhost
```bash
docker-compose -f docker-compose.localhost.yml up -d
# Uses: nginx.localhost.conf (HTTP only)
```

---

## Commands

### Using environment variable
```bash
export COMPOSE_FILE=docker-compose.localhost.yml
make build
make up
make logs
make down
```

### Using inline variable
```bash
COMPOSE_FILE=docker-compose.localhost.yml make build
COMPOSE_FILE=docker-compose.localhost.yml make up
COMPOSE_FILE=docker-compose.localhost.yml make status
```

### Or direct docker-compose
```bash
docker-compose -f docker-compose.localhost.yml build
docker-compose -f docker-compose.localhost.yml up -d
docker-compose -f docker-compose.localhost.yml ps
docker-compose -f docker-compose.localhost.yml logs -f
```

---

## Testing

### Health Checks
```bash
# Nginx health
curl http://localhost/health

# Backend health
curl http://localhost/api/health

# Frontend
curl http://localhost/
```

### API Test
```bash
# Test API endpoint (should return 401 without auth)
curl http://localhost/api/clients

# Login (if you have user)
curl -X POST http://localhost/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### MongoDB Test
```bash
# Connect to MongoDB
mongosh mongodb://admin:password@localhost:27017/qr_db --authenticationDatabase admin

# Or via Docker
docker-compose -f docker-compose.localhost.yml exec mongo mongosh
```

---

## Browser Access

### Chrome/Firefox
Just open: http://localhost

### Test with curl
```bash
# Homepage
curl -I http://localhost/

# API
curl http://localhost/api/health

# Health check
curl http://localhost/health
```

---

## Troubleshooting

### Port 80 already in use?

```bash
# Check what's using port 80
sudo lsof -i :80

# Option 1: Stop the service
sudo systemctl stop apache2  # or nginx

# Option 2: Change port in .env.localhost
HTTP_PORT=8080

# Then access via http://localhost:8080
```

### Can't access MongoDB from host?

MongoDB is exposed on localhost:27017 for debugging:

```bash
# Test connection
mongosh mongodb://admin:password@localhost:27017/qr_db --authenticationDatabase admin

# Or use MongoDB Compass
# Connection string: mongodb://admin:password@localhost:27017/qr_db?authSource=admin
```

### CORS error in browser?

Check backend logs:
```bash
COMPOSE_FILE=docker-compose.localhost.yml make logs-backend
```

Make sure `CORS_ALLOWED_ORIGIN=http://localhost` in `.env.localhost`

### Frontend shows 404 on refresh?

Check nginx logs:
```bash
COMPOSE_FILE=docker-compose.localhost.yml make logs-nginx
```

Should see SPA routing working (try_files directive).

---

## Development Workflow

### 1. Initial Setup
```bash
export COMPOSE_FILE=docker-compose.localhost.yml
make build
make up
```

### 2. Make Changes
Edit code in `be/` or `fe/` directories

### 3. Rebuild & Restart
```bash
# Rebuild specific service
docker-compose -f docker-compose.localhost.yml build backend
docker-compose -f docker-compose.localhost.yml up -d backend

# Or rebuild all
make build
make restart
```

### 4. View Logs
```bash
make logs              # All services
make logs-backend      # Backend only
make logs-frontend     # Frontend only
```

### 5. Clean Up
```bash
make down              # Stop containers
make clean             # Remove everything
```

---

## Switching Between Localhost and Production

### Use Localhost (HTTP)
```bash
export COMPOSE_FILE=docker-compose.localhost.yml
cp .env.localhost.example .env.localhost
make build
make up
```

### Use Production (HTTPS)
```bash
export COMPOSE_FILE=docker-compose.prod.yml
cp .env.production.example .env.production
# Edit .env.production with real values
make setup-ssl
make build
make up
```

---

## Advantages of Localhost Setup

✅ **No SSL needed** - Skip certificate setup
✅ **Simple configuration** - Default values work
✅ **Fast iteration** - No certificate validation
✅ **MongoDB access** - Debug database directly
✅ **Easy testing** - Just http://localhost

---

## When to Use

### Use Localhost Mode When:
- ✅ Local development on your machine
- ✅ Testing features quickly
- ✅ No domain available
- ✅ Learning/experimenting
- ✅ Internal network only

### Use Production Mode When:
- ✅ Deploying to server
- ✅ Need HTTPS/SSL
- ✅ Have a domain
- ✅ Public internet access
- ✅ Security matters

---

## Next Steps

After testing on localhost, deploy to production:

1. Review [QUICKSTART.md](QUICKSTART.md)
2. Setup domain and SSL
3. Update `.env.production`
4. Use `docker-compose.prod.yml`
5. Deploy with `make deploy`

---

## Quick Reference

```bash
# Set to localhost mode
export COMPOSE_FILE=docker-compose.localhost.yml

# Essential commands
make build             # Build images
make up                # Start services
make down              # Stop services
make logs              # View logs
make status            # Check status

# Access
Frontend:  http://localhost
API:       http://localhost/api
Health:    http://localhost/health
MongoDB:   localhost:27017
```

---

**Perfect for local development! No SSL, no hassle.** 🚀

**Last Updated**: 2025-11-24
