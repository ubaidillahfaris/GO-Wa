# WhatsApp API - Complete Deployment Guide

Comprehensive guide untuk deploy aplikasi WhatsApp API dengan Docker di semua scenario.

**Table of Contents:**
- [Quick Start](#quick-start)
- [Choose Your Scenario](#choose-your-scenario)
- [Localhost Development](#localhost-development)
- [Production Deployment](#production-deployment)
- [Existing Nginx Setup](#existing-nginx-setup)
- [Subdomain Configuration](#subdomain-configuration)
- [SSL/TLS Setup](#ssltls-setup)
- [Maintenance](#maintenance)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

### Using Interactive Setup (Recommended)

```bash
# Interactive setup - akan tanya domain, SSL, dll
make setup

# Atau manual:
make install       # Setup .env + SSL
make deploy        # Build & start
make status        # Check containers
```

### Manual Setup

```bash
# 1. Choose your scenario
cp .env.production.example .env.production
nano .env.production

# 2. Deploy
make deploy

# 3. Access
# Localhost: http://localhost
# Production: https://yourdomain.com
```

---

## Choose Your Scenario

### Scenario 1: Localhost Development 🏠

**Use Case:**
- Local development
- Testing
- No domain needed

**Quick Start:**
```bash
export COMPOSE_FILE=docker-compose.localhost.yml
make build && make up
```

**Access:** `http://localhost`

[Jump to Localhost Guide](#localhost-development)

---

### Scenario 2: Fresh Server (Port 80/443 Free) 🆕

**Use Case:**
- New VPS/server
- No existing nginx
- Port 80/443 available

**Quick Start:**
```bash
make setup    # Interactive setup
# Or manual:
make install
make deploy
```

**Access:** `https://yourdomain.com`

[Jump to Production Guide](#production-deployment)

---

### Scenario 3: Server with Nginx (Host) 🖥️

**Use Case:**
- Nginx installed di host (not Docker)
- Port 80/443 already used
- Other sites running

**Quick Start:**
```bash
# Deploy containers (no nginx)
docker compose -f docker-compose.existing-nginx.yml up -d

# Add nginx config
sudo cp nginx/nginx.host.conf /etc/nginx/sites-available/whatsapp
sudo ln -s /etc/nginx/sites-available/whatsapp /etc/nginx/sites-enabled/
sudo nano /etc/nginx/sites-available/whatsapp  # Update domain
sudo nginx -t && sudo systemctl reload nginx
```

**Access:** `https://yourdomain.com`

[Jump to Existing Nginx Guide](#existing-nginx-setup)

---

### Scenario 4: Server with Nginx Docker 🐳

**Use Case:**
- Nginx running in Docker
- Port 80/443 used by nginx container
- Need to join Docker network

**Quick Start:**
```bash
# Find nginx network
docker inspect nginx-server | grep Networks

# Update compose (edit network name)
nano docker-compose.nginx-docker.yml

# Deploy
docker compose --env-file .env.production -f docker-compose.nginx-docker.yml up -d

# Add config to nginx
docker cp nginx/nginx.docker.conf nginx-server:/etc/nginx/conf.d/whatsapp.conf
docker exec nginx-server nginx -s reload
```

**Access:** `https://yourdomain.com`

[Jump to Nginx Docker Guide](#nginx-docker-setup)

---

## Localhost Development

### Setup

```bash
# 1. Set to localhost mode
export COMPOSE_FILE=docker-compose.localhost.yml

# 2. Build & run
make build
make up

# 3. Check
make status
curl http://localhost/health
```

### Configuration

**Environment (.env.localhost):**
```bash
MONGO_USER=admin
MONGO_PASS=password
CORS_ALLOWED_ORIGIN=http://localhost
HTTP_PORT=80
```

### Architecture

```
Browser
    ↓
Nginx:80 (HTTP only, no SSL)
    ├─> Frontend:8080
    └─> Backend:3000
        └─> MongoDB:27017 (exposed for debugging)
```

### Access Points

- Frontend: http://localhost/
- API: http://localhost/api/
- Health: http://localhost/health
- MongoDB: localhost:27017

### Features

- ✅ No SSL needed
- ✅ MongoDB exposed untuk debugging
- ✅ Fast iteration
- ✅ Simple config

### Common Commands

```bash
# Start
make up

# Stop
make down

# View logs
make logs

# Restart
make restart

# Clean up
make clean
```

---

## Production Deployment

### Prerequisites

- Server dengan Docker installed
- Domain pointing ke server IP
- Port 80/443 available

### Step-by-Step

#### 1. DNS Configuration

Add A record:
```
Type: A
Name: @ (or subdomain)
Value: YOUR_SERVER_IP
```

Wait 5-30 minutes for DNS propagation.

**Test:**
```bash
dig yourdomain.com
# Should return your server IP
```

#### 2. Environment Setup

```bash
# Interactive setup (recommended)
make setup

# Or manual
cp .env.production.example .env.production
nano .env.production
```

**Required changes:**
```bash
MONGO_PASS=your_strong_password_here
JWT_SECRET=$(openssl rand -base64 64)
CORS_ALLOWED_ORIGIN=https://yourdomain.com
```

#### 3. SSL Certificate

**Option A: Let's Encrypt (Recommended)**
```bash
make ssl-letsencrypt DOMAIN=yourdomain.com

# Or manual:
sudo apt install certbot
docker compose -f docker-compose.prod.yml stop nginx
sudo certbot certonly --standalone -d yourdomain.com
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/key.pem
docker compose -f docker-compose.prod.yml start nginx
```

**Option B: Self-Signed (Development)**
```bash
make setup-ssl
```

#### 4. Update Nginx Config

Edit `nginx/nginx.prod.conf`:
```nginx
server_name yourdomain.com;  # Update this
```

Or use interactive setup:
```bash
make setup  # Will update automatically
```

#### 5. Deploy

```bash
# Full deployment
make deploy

# Or step by step
make build
make up
make status
```

#### 6. Verify

```bash
# Health check
make health

# Manual tests
curl https://yourdomain.com/health
curl https://yourdomain.com/api/health

# From browser
open https://yourdomain.com
```

#### 7. Auto-Renew SSL

```bash
# Add to crontab
sudo crontab -e

# Add line (renew daily at 2 AM):
0 2 * * * certbot renew --quiet --deploy-hook "docker compose -f /path/to/docker-compose.prod.yml restart nginx"
```

### Architecture

```
Internet
    ↓
Nginx:80/443 (SSL termination)
    ├─> Frontend:8080 (internal)
    └─> Backend:3000 (internal)
        └─> MongoDB:27017 (internal)
```

### Access Points

- Frontend: https://yourdomain.com/
- API: https://yourdomain.com/api/
- Health: https://yourdomain.com/health

All internal ports (8080, 3000, 27017) NOT exposed to internet!

---

## Existing Nginx Setup

### For Nginx Installed in Host

#### 1. Deploy Containers (No Nginx)

```bash
docker compose -f docker-compose.existing-nginx.yml up -d
```

Containers exposed to localhost only:
- Backend: 127.0.0.1:3000
- Frontend: 127.0.0.1:8080
- MongoDB: 127.0.0.1:27017

#### 2. Configure Existing Nginx

```bash
# Copy config
sudo cp nginx/nginx.host.conf /etc/nginx/sites-available/whatsapp

# Edit domain
sudo nano /etc/nginx/sites-available/whatsapp
# Update: server_name yourdomain.com;

# Enable site
sudo ln -s /etc/nginx/sites-available/whatsapp /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

#### 3. SSL Certificate

```bash
sudo certbot --nginx -d yourdomain.com
```

#### 4. Test

```bash
curl https://yourdomain.com/health
```

### Architecture

```
Existing Nginx (Host)
    ↓
Proxy to localhost
    ├─> 127.0.0.1:8080 (Frontend)
    └─> 127.0.0.1:3000 (Backend)
```

### Nginx Config Example

```nginx
upstream whatsapp_backend {
    server 127.0.0.1:3000;
}

upstream whatsapp_frontend {
    server 127.0.0.1:8080;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    location /api/ {
        rewrite ^/api/(.*) /$1 break;
        proxy_pass http://whatsapp_backend;
    }

    location / {
        proxy_pass http://whatsapp_frontend;
    }
}
```

---

## Nginx Docker Setup

### For Nginx Running in Docker

#### 1. Find Nginx Network

```bash
docker inspect nginx-server | grep -A 5 Networks
# Note the network name (e.g., nginx-network)
```

#### 2. Update Compose File

Edit `docker-compose.nginx-docker.yml`:
```yaml
networks:
  nginx-network:
    external: true
    name: YOUR_NGINX_NETWORK_NAME  # ← Update this
```

#### 3. Deploy WhatsApp Containers

```bash
docker compose --env-file .env.production -f docker-compose.nginx-docker.yml up -d
```

#### 4. Test Connectivity

```bash
# From nginx container
docker exec nginx-server wget -qO- http://whatsapp-backend-prod:3000/health
```

#### 5. Add Nginx Config

```bash
# Copy config
docker cp nginx/nginx.docker.conf nginx-server:/etc/nginx/conf.d/whatsapp.conf

# Edit domain (if needed)
docker exec nginx-server vi /etc/nginx/conf.d/whatsapp.conf

# Reload nginx
docker exec nginx-server nginx -t
docker exec nginx-server nginx -s reload
```

### Architecture

```
Nginx Docker
    ↓
Docker Network (shared)
    ├─> whatsapp-frontend-prod:8080
    └─> whatsapp-backend-prod:3000
```

### Key Point

Use **container names** as hostname:
```nginx
upstream whatsapp_backend {
    server whatsapp-backend-prod:3000;  # Container name!
}
```

Docker DNS automatically resolves container names.

---

## Subdomain Configuration

### Single Subdomain (Recommended)

```
wa.yourdomain.com/       → Frontend
wa.yourdomain.com/api/   → Backend API
```

**Setup:**

1. **DNS Record:**
```
Type: A
Name: wa
Value: YOUR_SERVER_IP
```

2. **Update Nginx Config:**
```nginx
server_name wa.yourdomain.com;
```

3. **Update Environment:**
```bash
CORS_ALLOWED_ORIGIN=https://wa.yourdomain.com
```

4. **SSL Certificate:**
```bash
sudo certbot certonly --standalone -d wa.yourdomain.com
```

5. **Deploy:**
```bash
make deploy
```

### Separate Subdomains (Advanced)

```
app.yourdomain.com/      → Frontend
api.yourdomain.com/      → Backend API
```

**Setup:**

1. **DNS Records:**
```
Type: A, Name: app, Value: YOUR_SERVER_IP
Type: A, Name: api, Value: YOUR_SERVER_IP
```

2. **Update Frontend Build:**
```bash
# fe/.env.production
VITE_API_BASE_URL=https://api.yourdomain.com
```

3. **Update Backend CORS:**
```bash
# .env.production
CORS_ALLOWED_ORIGIN=https://app.yourdomain.com
```

4. **Use Subdomain Config:**
```bash
# Copy nginx.subdomain.conf to nginx.prod.conf
cp nginx/nginx.subdomain.conf nginx/nginx.prod.conf
# Edit domains
```

5. **Rebuild Frontend:**
```bash
docker compose -f docker-compose.prod.yml build frontend
```

6. **Deploy:**
```bash
make deploy
```

---

## SSL/TLS Setup

### Let's Encrypt (Free & Recommended)

```bash
# Stop nginx
docker compose -f docker-compose.prod.yml stop nginx

# Get certificate
sudo certbot certonly --standalone -d yourdomain.com

# Copy to nginx volume
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/key.pem
sudo chmod 644 nginx/ssl/*.pem

# Start nginx
docker compose -f docker-compose.prod.yml start nginx
```

### Auto-Renewal

```bash
# Test renewal
sudo certbot renew --dry-run

# Add cron job
sudo crontab -e

# Add line:
0 2 * * * certbot renew --quiet --deploy-hook "docker compose -f /path/to/docker-compose.prod.yml restart nginx"
```

### Wildcard Certificate (Advanced)

```bash
# DNS challenge required
sudo certbot certonly --manual --preferred-challenges dns -d "*.yourdomain.com"

# Follow instructions to add TXT record
```

Works for all subdomains: wa, api, app, etc.

### Self-Signed (Development Only)

```bash
make setup-ssl

# Or manual:
mkdir -p nginx/ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/key.pem \
  -out nginx/ssl/cert.pem \
  -subj "/C=ID/ST=Jakarta/L=Jakarta/O=Dev/CN=localhost"
```

---

## Maintenance

### Daily Operations

```bash
# Check status
make status

# View logs
make logs
make logs-backend
make logs-nginx

# Resource usage
make stats

# Health check
make health
```

### Updates

```bash
# Pull latest code
git pull

# Zero-downtime update
make update

# Or full redeploy
make deploy
```

### Backup

```bash
# Backup MongoDB
make backup

# Stored in: backups/mongodb_backup_YYYYMMDD_HHMMSS.archive
```

### Restore

```bash
make restore FILE=backups/mongodb_backup_20241124_103000.archive
```

### Automated Backup

```bash
# Add to crontab (daily at 3 AM)
0 3 * * * cd /path/to/whatsapp && make backup
```

---

## Troubleshooting

### Common Issues

#### 1. Port Already in Use

**Problem:** Port 80/443 already used

**Check:**
```bash
sudo lsof -i :80
sudo lsof -i :443
```

**Fix:**
```bash
# Option 1: Stop conflicting service
sudo systemctl stop apache2

# Option 2: Use different scenario
# If nginx running: Use docker-compose.existing-nginx.yml
# If nginx docker: Use docker-compose.nginx-docker.yml
```

#### 2. MongoDB Connection Failed

**Problem:** Backend can't connect to MongoDB

**Check:**
```bash
make logs-mongo
make logs-backend
```

**Fix:**
```bash
# Check credentials in .env.production
# Restart MongoDB
docker compose -f docker-compose.prod.yml restart mongo
```

#### 3. 502 Bad Gateway

**Problem:** Nginx can't reach containers

**Check:**
```bash
make status
make logs-nginx
```

**Fix:**
```bash
# Restart services
make restart

# Or specific service
docker compose -f docker-compose.prod.yml restart backend
```

#### 4. Frontend 404 on Refresh

**Problem:** SPA routing not working

**Fix:**
Frontend nginx config should have:
```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

#### 5. CORS Error

**Problem:** Browser shows CORS error

**Check:**
```bash
docker compose -f docker-compose.prod.yml exec backend env | grep CORS
```

**Fix:**
```bash
# Update .env.production
CORS_ALLOWED_ORIGIN=https://yourdomain.com

# Restart backend
docker compose -f docker-compose.prod.yml restart backend
```

#### 6. SSL Certificate Error

**Problem:** Browser shows certificate invalid

**Check:**
```bash
openssl x509 -in nginx/ssl/cert.pem -text -noout | grep Subject
```

**Fix:**
```bash
# Regenerate with correct domain
sudo certbot certonly --standalone -d yourdomain.com
# Copy certs again
```

### Debug Commands

```bash
# Shell into container
docker compose -f docker-compose.prod.yml exec backend sh

# Check logs
make logs

# Inspect container
docker inspect whatsapp-backend-prod

# Network debug
docker network inspect whatsapp_whatsapp-network

# Test backend from nginx
docker compose -f docker-compose.prod.yml exec nginx wget -qO- http://backend:3000/health
```

---

## Configuration Files

### Docker Compose Files

| File | Use Case | Nginx | Ports |
|------|----------|-------|-------|
| docker-compose.localhost.yml | Local dev | Internal | 80 |
| docker-compose.prod.yml | Fresh server | Internal | 80, 443 |
| docker-compose.existing-nginx.yml | Nginx in host | None | 127.0.0.1:3000, 8080 |
| docker-compose.nginx-docker.yml | Nginx in Docker | None | Internal only |

### Nginx Configs

| File | Use Case | Domain |
|------|----------|--------|
| nginx.local.conf | Localhost | localhost |
| nginx.prod.conf | Production | Single domain/subdomain |
| nginx.host.conf | Existing nginx (host) | Any domain |
| nginx.docker.conf | Existing nginx (docker) | Any domain |

### Environment Files

| File | Use Case |
|------|----------|
| .env.localhost.example | Local development |
| .env.production.example | Production deployment |

---

## Security Checklist

Before production:

- [ ] Strong MongoDB password (16+ chars)
- [ ] Secure JWT_SECRET (64+ chars)
- [ ] Valid SSL certificate (not self-signed)
- [ ] Correct CORS_ALLOWED_ORIGIN
- [ ] Firewall configured (ports 22, 80, 443)
- [ ] SSH key-based auth enabled
- [ ] .env.production not in git
- [ ] Regular backups configured
- [ ] Monitoring setup
- [ ] Auto-renewal SSL configured

---

## Performance Tips

### Resource Limits

Edit docker-compose file:
```yaml
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 2G
```

### Nginx Caching

Already configured in production nginx configs.

### Database Optimization

Monitor with:
```bash
docker exec whatsapp-mongo-prod mongosh --eval "db.stats()"
```

---

## All Available Commands

```bash
make help              # Show all commands
make setup             # Interactive setup
make install           # Setup env + SSL
make build             # Build images
make up                # Start services
make down              # Stop services
make restart           # Restart all
make deploy            # Full deployment
make update            # Zero-downtime update
make logs              # View logs
make logs-backend      # Backend logs
make logs-frontend     # Frontend logs
make logs-nginx        # Nginx logs
make logs-mongo        # MongoDB logs
make status            # Container status
make health            # Health checks
make backup            # Backup MongoDB
make restore FILE=...  # Restore backup
make clean             # Clean Docker
make stats             # Resource usage
make setup-ssl         # Self-signed SSL
make ssl-letsencrypt   # Let's Encrypt SSL
```

---

## Summary

### Choose Your Path

**Local Development:**
```bash
export COMPOSE_FILE=docker-compose.localhost.yml
make build && make up
# Access: http://localhost
```

**Fresh Server:**
```bash
make setup  # Interactive
# Or: make install && make deploy
# Access: https://yourdomain.com
```

**Existing Nginx (Host):**
```bash
docker compose -f docker-compose.existing-nginx.yml up -d
sudo cp nginx/nginx.host.conf /etc/nginx/sites-available/whatsapp
# Configure nginx
```

**Existing Nginx (Docker):**
```bash
# Update network in docker-compose.nginx-docker.yml
docker compose --env-file .env.production -f docker-compose.nginx-docker.yml up -d
docker cp nginx/nginx.docker.conf nginx-server:/etc/nginx/conf.d/
```

---

**Production-ready Docker setup untuk semua scenario! 🚀**

**Last Updated**: 2025-11-24
