# WhatsApp API - Quick Setup Guide

Production-ready Docker setup untuk WhatsApp API dengan Go backend dan Vue.js frontend.

## 🚀 Quick Start

### Cara Termudah: Setup Wizard

Jalankan interactive setup wizard:

```bash
make setup
```

Wizard akan membantu Anda:
- ✅ Pilih deployment scenario (localhost/production)
- ✅ Konfigurasi domain/subdomain
- ✅ Generate MongoDB credentials
- ✅ Generate JWT secret
- ✅ Setup SSL certificate
- ✅ Create environment file

---

## 📋 Deployment Scenarios

### 1. Localhost (Development)

**Use Case:** Development lokal tanpa SSL

**Files:**
- `docker-compose.localhost.yml`
- `nginx/nginx.local.conf`
- `.env.localhost`

**Deploy:**
```bash
# Dengan wizard
make setup
# Pilih option 1 (Localhost)

# Atau manual
docker compose -f docker-compose.localhost.yml up -d
```

**Access:** http://localhost

---

### 2. Fresh Server (Production)

**Use Case:** Server baru dengan port 80/443 available

**Files:**
- `docker-compose.prod.yml`
- `nginx/nginx.production.conf`
- `.env.production`

**Deploy:**
```bash
# Dengan wizard
make setup
# Pilih option 2 (Fresh Server)

# Atau manual
# 1. Setup environment
cp .env.production.example .env.production
nano .env.production

# 2. Configure domain
make configure-domain DOMAIN=wa.yourdomain.com

# 3. Setup SSL
make ssl-letsencrypt DOMAIN=wa.yourdomain.com

# 4. Deploy
make deploy
```

**Access:** https://wa.yourdomain.com

---

### 3. Existing Nginx on Host

**Use Case:** Server sudah ada nginx di system (bukan Docker)

**Files:**
- `docker-compose.existing-nginx.yml`
- `nginx/existing-nginx-site.conf`
- `.env.production`

**Deploy:**
```bash
# 1. Deploy Docker (tanpa nginx container)
docker compose -f docker-compose.existing-nginx.yml up -d

# 2. Copy nginx config
sudo cp nginx/existing-nginx-site.conf /etc/nginx/sites-available/whatsapp
sudo ln -s /etc/nginx/sites-available/whatsapp /etc/nginx/sites-enabled/

# 3. Update domain in config
sudo nano /etc/nginx/sites-available/whatsapp
# Change: server_name wa.yourdomain.com;

# 4. SSL certificate
sudo certbot --nginx -d wa.yourdomain.com

# 5. Reload nginx
sudo nginx -t
sudo systemctl reload nginx
```

**Catatan:**
- Backend di: 127.0.0.1:3000
- Frontend di: 127.0.0.1:8080

---

### 4. Existing Nginx in Docker

**Use Case:** Server sudah ada nginx di Docker

**Files:**
- `docker-compose.nginx-docker.yml`
- `nginx/nginx-docker.conf`
- `.env.production`

**Deploy:**
```bash
# 1. Find nginx network
docker network ls
docker inspect <nginx-container> | grep Networks

# 2. Update network in compose file
nano docker-compose.nginx-docker.yml
# Update: name: YOUR_NGINX_NETWORK

# 3. Deploy containers
docker compose --env-file .env.production -f docker-compose.nginx-docker.yml up -d

# 4. Add config to nginx
docker cp nginx/nginx-docker.conf <nginx-container>:/etc/nginx/conf.d/whatsapp.conf

# 5. Update domain in config dan reload
docker exec <nginx-container> nano /etc/nginx/conf.d/whatsapp.conf
docker exec <nginx-container> nginx -t
docker exec <nginx-container> nginx -s reload
```

**Key:** Containers menggunakan Docker DNS (container names sebagai hostname)

---

## 📁 File Structure

### Nginx Configs (2 files only!)

1. **nginx/nginx.local.conf** - Untuk localhost/development
   - HTTP only (no SSL)
   - server_name: localhost
   - Routing: frontend + backend via proxy

2. **nginx/nginx.production.conf** - Untuk production (all scenarios)
   - HTTP to HTTPS redirect
   - SSL/TLS configuration
   - server_name: `__DOMAIN__` (placeholder, replaced by wizard)
   - Same routing as local

### Docker Compose Files

1. **docker-compose.localhost.yml** - Development lokal
2. **docker-compose.prod.yml** - Fresh server production
3. **docker-compose.existing-nginx.yml** - Nginx on host
4. **docker-compose.nginx-docker.yml** - Nginx in Docker

### Environment Files

- `.env.localhost.example` - Template untuk localhost
- `.env.production.example` - Template untuk production

---

## 🔧 Makefile Commands

### Setup & Configuration
```bash
make setup                    # Interactive setup wizard
make configure-domain         # Update domain (interactive)
make configure-domain DOMAIN=wa.yourdomain.com  # Update domain (specify)
make ssl-letsencrypt          # Setup Let's Encrypt SSL (interactive)
make ssl-letsencrypt DOMAIN=wa.yourdomain.com   # Setup SSL (specify)
```

### Deployment
```bash
make build                    # Build Docker images
make deploy                   # Full deployment (build + restart)
make up                       # Start services
make down                     # Stop services
make restart                  # Restart services
```

### Monitoring
```bash
make status                   # Show services status
make logs                     # Show all logs
make logs-backend             # Show backend logs
make logs-frontend            # Show frontend logs
make health                   # Check health of all services
```

### Management
```bash
make backup                   # Backup MongoDB
make restore FILE=backup.tar  # Restore MongoDB
make clean                    # Clean Docker resources
```

---

## 🌐 Domain/Subdomain Setup

### Single Subdomain (Recommended)

**Architecture:**
```
wa.yourdomain.com/       → Frontend
wa.yourdomain.com/api/   → Backend API
```

**Benefits:**
- ✅ 1 SSL certificate
- ✅ No CORS issues
- ✅ Simple management

**Setup:**
```bash
# 1. Add DNS A record
# Name: wa
# Value: YOUR_SERVER_IP

# 2. Run wizard
make setup

# 3. Or manual
make configure-domain DOMAIN=wa.yourdomain.com
make ssl-letsencrypt DOMAIN=wa.yourdomain.com
make deploy
```

### Separate Subdomains (Advanced)

**Architecture:**
```
app.yourdomain.com → Frontend
api.yourdomain.com → Backend API
```

**Setup:**
```bash
# 1. Add 2 DNS A records
# app → YOUR_SERVER_IP
# api → YOUR_SERVER_IP

# 2. Update nginx config (nginx.subdomain.conf)
# 3. Update frontend build (.env.production)
VITE_API_BASE_URL=https://api.yourdomain.com

# 4. Update backend CORS (.env.production)
CORS_ALLOWED_ORIGIN=https://app.yourdomain.com

# 5. Get SSL for both
sudo certbot certonly --standalone -d app.yourdomain.com -d api.yourdomain.com
```

**Note:** Perlu rebuild frontend setelah ubah VITE_API_BASE_URL!

---

## 🔒 SSL Certificate Options

### Option 1: Self-Signed (Development)
```bash
make setup-ssl
# Or manual:
mkdir -p nginx/ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/key.pem \
  -out nginx/ssl/cert.pem \
  -subj "/C=ID/ST=Jakarta/L=Jakarta/O=Dev/CN=localhost"
```

### Option 2: Let's Encrypt (Production)
```bash
make ssl-letsencrypt DOMAIN=wa.yourdomain.com
```

**Auto-renewal:**
```bash
# Test renewal
sudo certbot renew --dry-run

# Add to crontab (already done by wizard)
0 2 * * * certbot renew --quiet --deploy-hook "docker compose -f /path/to/docker-compose.prod.yml restart nginx"
```

### Option 3: Custom Certificate
```bash
# Copy your certificates
cp your-cert.pem nginx/ssl/cert.pem
cp your-key.pem nginx/ssl/key.pem
chmod 644 nginx/ssl/*.pem
```

---

## 🔍 Troubleshooting

### 502 Bad Gateway
```bash
# Check containers
make status

# Check logs
make logs-nginx
make logs-backend

# Restart
make restart
```

### CORS Error
```bash
# Check CORS setting
docker compose exec backend env | grep CORS

# Update .env.production
CORS_ALLOWED_ORIGIN=https://wa.yourdomain.com

# Restart backend
docker compose restart backend
```

### SSL Certificate Error
```bash
# Check certificate
openssl x509 -in nginx/ssl/cert.pem -text -noout | grep Subject

# Make sure matches domain in nginx config
grep server_name nginx/nginx.production.conf
```

### DNS Not Resolving
```bash
# Check DNS
dig wa.yourdomain.com
nslookup wa.yourdomain.com

# Wait for propagation (5-30 minutes)
```

---

## 📊 Architecture

### Network Topology

```
┌─────────────────────────────────────────────┐
│              Internet                       │
└────────────────┬────────────────────────────┘
                 │
         ┌───────▼───────┐
         │  Nginx (80/443)│
         │  SSL/TLS       │
         └───────┬────────┘
                 │
         Docker Network
         ┌───────┴────────────┐
         │                    │
    ┌────▼─────┐         ┌───▼────┐
    │ Frontend │         │Backend │
    │  :8080   │         │ :3000  │
    └──────────┘         └────┬───┘
                              │
                         ┌────▼─────┐
                         │ MongoDB  │
                         │  :27017  │
                         └──────────┘
```

### Service Communication

**Fresh Server (Scenario 2):**
- Internet → Nginx (port 80/443)
- Nginx → Frontend (Docker network)
- Nginx → Backend (Docker network)
- Backend → MongoDB (Docker network)

**Existing Nginx on Host (Scenario 3):**
- Internet → System Nginx (port 80/443)
- System Nginx → Backend (127.0.0.1:3000)
- System Nginx → Frontend (127.0.0.1:8080)
- Backend → MongoDB (Docker network)

**Existing Nginx in Docker (Scenario 4):**
- Internet → Nginx Docker (port 80/443)
- Nginx → Frontend (shared Docker network, container name)
- Nginx → Backend (shared Docker network, container name)
- Backend → MongoDB (internal Docker network)

---

## 📦 What's Included

### Backend (Go)
- WhatsApp API dengan whatsmeow
- MongoDB database
- JWT authentication
- API key management
- File upload support
- Health checks

### Frontend (Vue.js)
- Modern UI dengan Tailwind CSS
- QR code scanner
- Message management
- Device management
- Responsive design

### Infrastructure
- Multi-stage Docker builds
- Nginx reverse proxy
- SSL/TLS support
- Rate limiting
- Gzip compression
- Security headers
- Health checks
- Auto-restart
- Resource limits
- Volume persistence

---

## 🎯 Quick Reference

### Most Common Setup

**Fresh Server with Subdomain:**
```bash
# 1. Add DNS: wa → YOUR_SERVER_IP
# 2. Run wizard:
make setup
# Choose option 2, enter: wa.yourdomain.com
# 3. Done! Access: https://wa.yourdomain.com
```

**Localhost Development:**
```bash
# 1. Run wizard:
make setup
# Choose option 1
# 2. Done! Access: http://localhost
```

---

## 📚 Complete Documentation

Untuk dokumentasi lengkap, lihat [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md)

---

## 🆘 Need Help?

1. Check logs: `make logs`
2. Check health: `make health`
3. Check status: `make status`
4. See troubleshooting section above
5. Review DEPLOYMENT_GUIDE.md

---

**Last Updated:** 2025-11-24

**Perfect untuk production deployment!** 🚀
