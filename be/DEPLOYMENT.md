# 🚀 GO-Wa Deployment Guide

Comprehensive deployment guide for GO-Wa WhatsApp API.

---

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start (Docker)](#quick-start-docker)
- [Production Deployment](#production-deployment)
- [Environment Configuration](#environment-configuration)
- [SSL/HTTPS Setup](#sslhttps-setup)
- [Monitoring & Logging](#monitoring--logging)
- [Backup & Restore](#backup--restore)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required
- Docker Engine 20.10+
- Docker Compose 2.0+
- 2GB+ RAM
- 10GB+ Disk Space

### Optional (for manual deployment)
- Go 1.25.1+
- MongoDB 7.0+
- Nginx (for reverse proxy)

---

## Quick Start (Docker)

### 1. Clone Repository

```bash
git clone https://github.com/ubaidillahfaris/GO-Wa.git
cd GO-Wa
```

### 2. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit configuration (IMPORTANT!)
nano .env  # or vim, code, etc.
```

**⚠️ CRITICAL: Change these values in production:**
```env
JWT_SECRET=your-super-secret-jwt-key-here
MONGO_USER=your_mongo_user
MONGO_PASS=your_strong_password
ENVIRONMENT=production
```

### 3. Start Services

```bash
# Build and start all services
docker-compose up -d

# Check logs
docker-compose logs -f app

# Check status
docker-compose ps
```

### 4. Verify Deployment

```bash
# Check health endpoint
curl http://localhost:3000/health

# Expected response:
# {"status":"ok"}
```

### 5. Access Application

- **API**: http://localhost:3000
- **Nginx Proxy**: http://localhost (if enabled)
- **MongoDB**: localhost:27017

---

## Production Deployment

### 1. Server Preparation

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker-compose --version
```

### 2. Firewall Configuration

```bash
# Allow HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Optional: Allow direct app access
sudo ufw allow 3000/tcp

# Optional: Allow MongoDB (if external access needed)
# sudo ufw allow 27017/tcp

# Enable firewall
sudo ufw enable
```

### 3. Production Environment Setup

```bash
# Create production .env
cp .env.example .env

# Generate strong JWT secret
openssl rand -base64 64

# Edit .env with production values
vim .env
```

**Production `.env` example:**
```env
# Server
PORT=3000
ENVIRONMENT=production

# MongoDB (use strong passwords!)
MONGO_USER=admin_prod
MONGO_PASS=<generate-with: openssl rand -base64 32>
MONGO_HOST=mongo:27017
MONGO_DB=whatsapp_prod

# JWT (use strong secret!)
JWT_SECRET=<generate-with: openssl rand -base64 64>
JWT_EXPIRES_MIN=60

# WhatsApp
WHATSAPP_STORES_DIR=./stores
WHATSAPP_UPLOADS_DIR=./uploads/whatsapp
WHATSAPP_MAX_CONCURRENCY=20

# CORS (update to your frontend domain)
CORS_ALLOWED_ORIGIN=https://yourdomain.com
CORS_MAX_AGE=43200
```

### 4. Start Production Stack

```bash
# Build with no cache
docker-compose build --no-cache

# Start services
docker-compose up -d

# Check logs
docker-compose logs -f

# Verify all services are healthy
docker-compose ps
```

---

## Environment Configuration

### Server Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | Application port |
| `ENVIRONMENT` | `development` | Environment mode |

### MongoDB Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGO_USER` | `root` | MongoDB username |
| `MONGO_PASS` | `password` | MongoDB password |
| `MONGO_HOST` | `mongo:27017` | MongoDB host |
| `MONGO_DB` | `qr_db` | Database name |

### JWT Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | - | **REQUIRED** JWT signing secret |
| `JWT_EXPIRES_MIN` | `60` | Token expiration (minutes) |

### WhatsApp Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `WHATSAPP_STORES_DIR` | `./stores` | Session storage directory |
| `WHATSAPP_UPLOADS_DIR` | `./uploads/whatsapp` | File upload directory |
| `WHATSAPP_MAX_CONCURRENCY` | `10` | Max concurrent message processing |

### CORS Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `CORS_ALLOWED_ORIGIN` | `http://localhost:5173` | Allowed origin for CORS |
| `CORS_MAX_AGE` | `43200` | Preflight cache duration |

---

## SSL/HTTPS Setup

### Option 1: Let's Encrypt (Recommended)

```bash
# Install Certbot
sudo apt install certbot

# Generate certificate
sudo certbot certonly --standalone -d yourdomain.com

# Certificates will be in:
# /etc/letsencrypt/live/yourdomain.com/fullchain.pem
# /etc/letsencrypt/live/yourdomain.com/privkey.pem

# Copy to nginx/ssl/
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/key.pem
```

### Option 2: Self-Signed Certificate (Development)

```bash
# Generate self-signed certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/key.pem \
  -out nginx/ssl/cert.pem

# Fill in the prompts (use localhost for Common Name)
```

### Enable HTTPS in Nginx

Edit `nginx/nginx.conf` and uncomment the HTTPS server block (line ~96):

```nginx
server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    # ... rest of config
}
```

Restart nginx:
```bash
docker-compose restart web
```

---

## Monitoring & Logging

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f app
docker-compose logs -f mongo
docker-compose logs -f web

# Last 100 lines
docker-compose logs --tail=100 app

# Follow with timestamps
docker-compose logs -f -t app
```

### Health Checks

```bash
# Application health
curl http://localhost:3000/health

# MongoDB health
docker-compose exec mongo mongosh --eval "db.adminCommand('ping')"

# Nginx health
docker-compose exec web nginx -t
```

### Resource Usage

```bash
# Container stats
docker stats

# Disk usage
docker system df

# Volume sizes
docker volume ls
du -sh /var/lib/docker/volumes/*
```

---

## Backup & Restore

### Backup WhatsApp Sessions

```bash
# Create backup directory
mkdir -p backups/$(date +%Y%m%d)

# Backup stores volume
docker run --rm \
  -v go-wa_whatsapp-stores:/data \
  -v $(pwd)/backups/$(date +%Y%m%d):/backup \
  alpine tar czf /backup/stores.tar.gz -C /data .

# Backup uploads volume
docker run --rm \
  -v go-wa_whatsapp-uploads:/data \
  -v $(pwd)/backups/$(date +%Y%m%d):/backup \
  alpine tar czf /backup/uploads.tar.gz -C /data .
```

### Backup MongoDB

```bash
# Dump database
docker-compose exec mongo mongodump \
  --username root \
  --password password \
  --authenticationDatabase admin \
  --db qr_db \
  --out /data/backup

# Copy to host
docker cp mongo-wa:/data/backup ./backups/$(date +%Y%m%d)/mongodb
```

### Restore WhatsApp Sessions

```bash
# Restore stores
docker run --rm \
  -v go-wa_whatsapp-stores:/data \
  -v $(pwd)/backups/YYYYMMDD:/backup \
  alpine sh -c "cd /data && tar xzf /backup/stores.tar.gz"

# Restore uploads
docker run --rm \
  -v go-wa_whatsapp-uploads:/data \
  -v $(pwd)/backups/YYYYMMDD:/backup \
  alpine sh -c "cd /data && tar xzf /backup/uploads.tar.gz"
```

### Restore MongoDB

```bash
# Restore database
docker-compose exec mongo mongorestore \
  --username root \
  --password password \
  --authenticationDatabase admin \
  --db qr_db \
  /data/backup/qr_db
```

---

## Troubleshooting

### Common Issues

#### 1. Container Won't Start

```bash
# Check logs for errors
docker-compose logs app

# Check if port is already in use
sudo netstat -tulpn | grep :3000

# Remove and recreate
docker-compose down
docker-compose up -d --force-recreate
```

#### 2. MongoDB Connection Failed

```bash
# Check if mongo is running
docker-compose ps mongo

# Check mongo logs
docker-compose logs mongo

# Verify credentials in .env
cat .env | grep MONGO

# Test connection
docker-compose exec mongo mongosh \
  -u root -p password \
  --authenticationDatabase admin
```

#### 3. WhatsApp Session Lost

```bash
# Check if stores volume exists
docker volume ls | grep stores

# Check permissions
docker-compose exec app ls -la /app/stores

# Restore from backup (see Backup & Restore section)
```

#### 4. Out of Disk Space

```bash
# Check disk usage
df -h

# Clean Docker system
docker system prune -a --volumes

# Remove old images
docker image prune -a

# Remove unused volumes (CAUTION: may delete data!)
docker volume prune
```

#### 5. Build Fails

```bash
# Clean build cache
docker-compose build --no-cache

# Remove all containers and rebuild
docker-compose down -v
docker-compose up -d --build
```

### Performance Issues

#### High CPU Usage
```bash
# Check container stats
docker stats

# Reduce WHATSAPP_MAX_CONCURRENCY in .env
WHATSAPP_MAX_CONCURRENCY=5

# Restart app
docker-compose restart app
```

#### High Memory Usage
```bash
# Add memory limits to docker-compose.yaml
services:
  app:
    mem_limit: 1g
    mem_reservation: 512m
```

#### Slow Response Times
```bash
# Check nginx logs for bottlenecks
docker-compose logs web | grep "upstream"

# Increase nginx workers in nginx.conf
worker_processes 4;

# Enable nginx caching (add to nginx.conf)
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=my_cache:10m;
```

---

## Maintenance

### Update Application

```bash
# Pull latest code
git pull origin main

# Rebuild and restart
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# Verify
docker-compose logs -f app
```

### Update Docker Images

```bash
# Pull latest images
docker-compose pull

# Recreate containers
docker-compose up -d --force-recreate
```

### Cleanup

```bash
# Remove stopped containers
docker container prune

# Remove unused images
docker image prune -a

# Remove unused volumes (CAUTION!)
docker volume prune

# Complete cleanup (CAUTION: removes everything!)
docker system prune -a --volumes
```

---

## Production Checklist

- [ ] Strong JWT_SECRET generated
- [ ] Strong MongoDB credentials
- [ ] Environment set to `production`
- [ ] CORS_ALLOWED_ORIGIN set to actual domain
- [ ] SSL certificate configured
- [ ] Firewall rules configured
- [ ] Backup strategy implemented
- [ ] Monitoring setup
- [ ] Log rotation configured
- [ ] Rate limiting enabled in nginx
- [ ] Health checks verified
- [ ] Documentation updated

---

## Support

For issues or questions:
- GitHub Issues: https://github.com/ubaidillahfaris/GO-Wa/issues
- Documentation: [ARCHITECTURE.md](./ARCHITECTURE.md)

---

## License

[Your License Here]
