# Docker Production Deployment Guide

Panduan lengkap untuk deploy aplikasi WhatsApp API dengan Docker di production.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [SSL/TLS Setup](#ssltls-setup)
- [Deployment](#deployment)
- [Monitoring](#monitoring)
- [Backup & Recovery](#backup--recovery)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### System Requirements
- Docker Engine 20.10+
- Docker Compose 2.0+
- Minimum 4GB RAM (8GB recommended)
- 20GB free disk space
- Linux server (Ubuntu 20.04+ / Debian 11+ recommended)

### Install Docker
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

---

## Quick Start

### 1. Clone Repository
```bash
git clone <your-repo-url>
cd whatsapp
```

### 2. Setup Environment
```bash
# Copy and edit production environment file
cp .env.production.example .env.production

# Edit dengan editor favorit
nano .env.production
```

**IMPORTANT:** Update nilai berikut:
- `MONGO_USER` - MongoDB username
- `MONGO_PASS` - Strong password untuk MongoDB
- `JWT_SECRET` - Generate dengan: `openssl rand -base64 64`
- `CORS_ALLOWED_ORIGIN` - Domain frontend Anda

### 3. Setup SSL Certificates
```bash
# Create SSL directory
mkdir -p nginx/ssl

# Option 1: Self-signed (development/testing only)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/key.pem \
  -out nginx/ssl/cert.pem

# Option 2: Let's Encrypt (recommended for production)
# See SSL/TLS Setup section below
```

### 4. Build and Run
```bash
# Build images
docker-compose -f docker-compose.prod.yml build

# Start services
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f
```

### 5. Verify Deployment
```bash
# Check health
curl http://localhost/health

# Check API (via nginx)
curl https://localhost/api/health
```

---

## Configuration

### Environment Variables

#### Backend (.env.production)
```bash
# Server
PORT=3000                    # Internal backend port
ENVIRONMENT=production

# MongoDB
MONGO_USER=admin_whatsapp    # Change this
MONGO_PASS=secure_password   # CHANGE THIS!
MONGO_HOST=mongo:27017
MONGO_DB=whatsapp_production

# JWT
JWT_SECRET=your-secret-key   # Generate: openssl rand -base64 64
JWT_EXPIRES_MIN=60

# WhatsApp
WHATSAPP_STORES_DIR=./stores
WHATSAPP_UPLOADS_DIR=./uploads/whatsapp
WHATSAPP_MAX_CONCURRENCY=20  # Adjust based on server capacity

# CORS
CORS_ALLOWED_ORIGIN=https://yourdomain.com  # Your frontend URL
CORS_MAX_AGE=43200
```

#### Frontend
```bash
# fe/.env.production
VITE_API_BASE_URL=/api
```

### Resource Limits

Edit `docker-compose.prod.yml` untuk adjust resource limits:

```yaml
deploy:
  resources:
    limits:
      cpus: '2'      # Maximum CPU cores
      memory: 2G     # Maximum memory
    reservations:
      cpus: '1'      # Reserved CPU cores
      memory: 1G     # Reserved memory
```

---

## SSL/TLS Setup

### Option 1: Let's Encrypt (Recommended)

#### Install Certbot
```bash
sudo apt install certbot python3-certbot-nginx -y
```

#### Generate Certificates
```bash
# Stop nginx if running
docker-compose -f docker-compose.prod.yml stop nginx

# Generate certificate
sudo certbot certonly --standalone -d yourdomain.com -d www.yourdomain.com

# Copy certificates
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/key.pem
sudo chmod 644 nginx/ssl/*.pem

# Restart nginx
docker-compose -f docker-compose.prod.yml up -d nginx
```

#### Auto-renewal
```bash
# Add cron job for auto-renewal
sudo crontab -e

# Add this line (runs every day at 2 AM)
0 2 * * * certbot renew --quiet --deploy-hook "docker-compose -f /path/to/whatsapp/docker-compose.prod.yml restart nginx"
```

### Option 2: Self-Signed Certificate (Development Only)

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/key.pem \
  -out nginx/ssl/cert.pem \
  -subj "/C=ID/ST=Jakarta/L=Jakarta/O=YourCompany/CN=yourdomain.com"
```

### Update nginx Configuration

Edit [nginx/nginx.prod.conf](nginx/nginx.prod.conf:88) untuk enable HTTPS redirect:

```nginx
# Uncomment this in HTTP server block
location / {
    return 301 https://$host$request_uri;
}
```

---

## Deployment

### Production Deployment

```bash
# 1. Pull latest changes
git pull origin main

# 2. Build images (no cache for clean build)
docker-compose -f docker-compose.prod.yml build --no-cache

# 3. Stop current containers
docker-compose -f docker-compose.prod.yml down

# 4. Start new containers
docker-compose -f docker-compose.prod.yml up -d

# 5. Verify deployment
docker-compose -f docker-compose.prod.yml ps
docker-compose -f docker-compose.prod.yml logs -f --tail=100
```

### Zero-Downtime Deployment

```bash
# Build new images
docker-compose -f docker-compose.prod.yml build

# Rolling update (one service at a time)
docker-compose -f docker-compose.prod.yml up -d --no-deps --build backend
docker-compose -f docker-compose.prod.yml up -d --no-deps --build frontend
docker-compose -f docker-compose.prod.yml up -d --no-deps --build nginx
```

### Rollback

```bash
# View images
docker images

# Tag previous version as rollback
docker tag whatsapp-backend:previous whatsapp-backend:latest

# Restart with previous image
docker-compose -f docker-compose.prod.yml up -d backend
```

---

## Monitoring

### Container Status
```bash
# List running containers
docker-compose -f docker-compose.prod.yml ps

# Resource usage
docker stats

# Service logs
docker-compose -f docker-compose.prod.yml logs -f [service_name]
```

### Health Checks
```bash
# Overall health
curl http://localhost/health

# Backend health
docker-compose -f docker-compose.prod.yml exec backend curl http://localhost:3000/health

# MongoDB health
docker-compose -f docker-compose.prod.yml exec mongo mongosh --eval "db.adminCommand('ping')"
```

### Log Management
```bash
# View logs
docker-compose -f docker-compose.prod.yml logs -f --tail=100

# Export logs
docker-compose -f docker-compose.prod.yml logs --no-color > logs_$(date +%Y%m%d).txt

# Clean old logs
docker-compose -f docker-compose.prod.yml down
docker system prune -af --volumes
```

### Monitoring Tools (Optional)

#### Portainer (Web UI)
```bash
docker volume create portainer_data
docker run -d -p 9000:9000 --name=portainer --restart=always \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v portainer_data:/data \
  portainer/portainer-ce:latest
```

Access: `http://your-server:9000`

---

## Backup & Recovery

### Backup

#### MongoDB Backup
```bash
# Create backup directory
mkdir -p backups

# Backup MongoDB
docker-compose -f docker-compose.prod.yml exec -T mongo \
  mongodump --username=$MONGO_USER --password=$MONGO_PASS \
  --authenticationDatabase=admin --db=whatsapp_production \
  --archive > backups/mongodb_backup_$(date +%Y%m%d_%H%M%S).archive

# Backup volumes
docker run --rm \
  -v whatsapp_whatsapp-stores:/data \
  -v $(pwd)/backups:/backup \
  alpine tar czf /backup/whatsapp-stores_$(date +%Y%m%d).tar.gz /data
```

#### Automated Backup Script
```bash
# Create backup script
cat > backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# Backup MongoDB
docker-compose -f docker-compose.prod.yml exec -T mongo \
  mongodump --username=$MONGO_USER --password=$MONGO_PASS \
  --authenticationDatabase=admin --db=whatsapp_production \
  --archive > $BACKUP_DIR/mongodb_$DATE.archive

# Backup volumes
docker run --rm \
  -v whatsapp_whatsapp-stores:/stores \
  -v whatsapp_whatsapp-uploads:/uploads \
  -v $(pwd)/backups:/backup \
  alpine sh -c "cd / && tar czf /backup/volumes_$DATE.tar.gz stores uploads"

# Keep only last 7 days
find $BACKUP_DIR -type f -mtime +7 -delete

echo "Backup completed: $DATE"
EOF

chmod +x backup.sh

# Add to crontab (daily at 3 AM)
echo "0 3 * * * cd /path/to/whatsapp && ./backup.sh" | crontab -
```

### Recovery

#### Restore MongoDB
```bash
# Stop services
docker-compose -f docker-compose.prod.yml stop backend

# Restore database
docker-compose -f docker-compose.prod.yml exec -T mongo \
  mongorestore --username=$MONGO_USER --password=$MONGO_PASS \
  --authenticationDatabase=admin --drop \
  --archive < backups/mongodb_backup_YYYYMMDD_HHMMSS.archive

# Restart services
docker-compose -f docker-compose.prod.yml start backend
```

#### Restore Volumes
```bash
# Stop services
docker-compose -f docker-compose.prod.yml down

# Restore volumes
docker run --rm \
  -v whatsapp_whatsapp-stores:/data \
  -v $(pwd)/backups:/backup \
  alpine sh -c "cd / && tar xzf /backup/volumes_YYYYMMDD_HHMMSS.tar.gz"

# Start services
docker-compose -f docker-compose.prod.yml up -d
```

---

## Troubleshooting

### Common Issues

#### 1. Port Already in Use
```bash
# Check what's using the port
sudo lsof -i :80
sudo lsof -i :443

# Kill the process or change port in docker-compose.prod.yml
```

#### 2. Permission Denied
```bash
# Fix volume permissions
docker-compose -f docker-compose.prod.yml down
sudo chown -R 1000:1000 be/stores be/uploads be/keys
docker-compose -f docker-compose.prod.yml up -d
```

#### 3. MongoDB Connection Failed
```bash
# Check MongoDB logs
docker-compose -f docker-compose.prod.yml logs mongo

# Check credentials in .env.production
# Restart MongoDB
docker-compose -f docker-compose.prod.yml restart mongo
```

#### 4. Backend Not Responding
```bash
# Check backend logs
docker-compose -f docker-compose.prod.yml logs backend

# Check if backend is healthy
docker-compose -f docker-compose.prod.yml exec backend curl http://localhost:3000/health

# Restart backend
docker-compose -f docker-compose.prod.yml restart backend
```

#### 5. Frontend 502 Bad Gateway
```bash
# Check frontend logs
docker-compose -f docker-compose.prod.yml logs frontend

# Check nginx logs
docker-compose -f docker-compose.prod.yml logs nginx

# Verify frontend is running
docker-compose -f docker-compose.prod.yml exec frontend ps aux

# Restart services
docker-compose -f docker-compose.prod.yml restart frontend nginx
```

#### 6. SSL Certificate Error
```bash
# Verify certificate files exist
ls -la nginx/ssl/

# Check certificate validity
openssl x509 -in nginx/ssl/cert.pem -text -noout

# Restart nginx
docker-compose -f docker-compose.prod.yml restart nginx
```

### Debug Commands

```bash
# Shell into container
docker-compose -f docker-compose.prod.yml exec backend sh
docker-compose -f docker-compose.prod.yml exec frontend sh

# View real-time logs
docker-compose -f docker-compose.prod.yml logs -f --tail=100

# Check container resource usage
docker stats

# Inspect container
docker inspect whatsapp-backend-prod

# Network inspection
docker network inspect whatsapp_whatsapp-network

# Volume inspection
docker volume inspect whatsapp_whatsapp-stores
```

### Performance Optimization

#### 1. Increase Resource Limits
Edit [docker-compose.prod.yml](docker-compose.prod.yml) and increase CPU/memory limits.

#### 2. Enable Docker BuildKit
```bash
# Add to ~/.bashrc or ~/.zshrc
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
```

#### 3. Clean Up Docker
```bash
# Remove unused images, containers, networks
docker system prune -a

# Remove unused volumes
docker volume prune
```

---

## Security Checklist

- [ ] Use strong passwords for MongoDB
- [ ] Generate secure JWT_SECRET (min 64 characters)
- [ ] Use valid SSL certificates (Let's Encrypt)
- [ ] Enable HTTPS redirect in nginx
- [ ] Set proper CORS_ALLOWED_ORIGIN
- [ ] Regularly update Docker images
- [ ] Enable firewall (UFW)
- [ ] Use non-root users in containers
- [ ] Regularly backup data
- [ ] Monitor logs for suspicious activity
- [ ] Keep Docker and Docker Compose updated

---

## Production Best Practices

1. **Environment Separation**
   - Never use development credentials in production
   - Keep `.env.production` secure and out of version control

2. **SSL/TLS**
   - Always use valid SSL certificates in production
   - Enable HSTS header in nginx

3. **Monitoring**
   - Set up monitoring and alerting
   - Use logging aggregation (ELK stack, Grafana)

4. **Backup**
   - Automated daily backups
   - Test restore procedures regularly
   - Store backups off-site

5. **Updates**
   - Keep Docker images updated
   - Apply security patches promptly
   - Test updates in staging first

6. **Scaling**
   - Use Docker Swarm or Kubernetes for multi-server deployment
   - Implement load balancing
   - Use managed database services

---

## Support

For issues or questions:
- GitHub Issues: [Your Repo Issues](https://github.com/yourusername/whatsapp/issues)
- Documentation: [Your Docs](https://docs.yourdomain.com)

---

## License

[Your License]

---

**Last Updated:** 2025-11-24
