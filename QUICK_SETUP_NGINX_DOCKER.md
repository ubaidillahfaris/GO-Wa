# Quick Setup: WhatsApp API dengan Nginx Docker Existing

Setup cepat untuk deploy WhatsApp API di server yang sudah punya nginx di Docker.

## 📋 Prerequisites

- Server dengan nginx running di Docker
- Port 80/443 sudah di-handle nginx Docker
- Domain/subdomain DNS pointing ke server

---

## 🚀 Quick Setup (5 Minutes)

### 1. Run Setup Wizard

```bash
cd /root/app/gowa  # atau path project Anda
make setup
```

**Wizard akan tanya:**
1. **Scenario:** Pilih `4` (Existing Nginx in Docker)
2. **Domain:** Input `whatsapp.agsatu.id`
3. **MongoDB credentials:** Generate otomatis atau input manual
4. **JWT secret:** Generate otomatis atau input manual
5. **Nginx network:** Input nama network nginx Docker Anda
6. **SSL:** Pilih Let's Encrypt (webroot challenge)
7. **Webroot path:** Input path webroot nginx (untuk ACME challenge)
8. **Deploy now:** Pilih `Y` untuk auto-deploy

**Done!** Wizard akan:
- ✅ Create `.env.production` dengan credentials
- ✅ Build Docker images
- ✅ Start containers (backend, frontend, mongo)
- ✅ Setup SSL certificate dengan Let's Encrypt
- ✅ Configure auto-renewal
- ✅ Show next steps

---

### 2. Manual Steps (Required)

Setelah wizard selesai, ada 3 langkah manual:

#### A. Copy Nginx Config

```bash
# Find your nginx container name
docker ps | grep nginx

# Copy config
docker cp nginx/nginx-docker.conf <nginx-container>:/etc/nginx/conf.d/whatsapp.conf
```

#### B. Update Domain

```bash
# Replace domain placeholder
docker exec <nginx-container> sed -i 's/wa.yourdomain.com/whatsapp.agsatu.id/g' /etc/nginx/conf.d/whatsapp.conf
```

#### C. Reload Nginx

```bash
# Test config
docker exec <nginx-container> nginx -t

# Reload
docker exec <nginx-container> nginx -s reload
```

---

### 3. Test Access

```bash
# Test HTTPS
curl -I https://whatsapp.agsatu.id

# Test API
curl https://whatsapp.agsatu.id/api/health

# Open in browser
open https://whatsapp.agsatu.id
```

---

## 📝 What the Wizard Does

### Environment Setup
Creates `.env.production` with:
- MongoDB credentials (auto-generated)
- JWT secret (auto-generated)
- CORS settings for your domain
- Database name

### Docker Network
Connects WhatsApp containers to your existing nginx network:
```
nginx-network (your existing)
  ├─> nginx (existing)
  ├─> whatsapp-backend-prod (new)
  └─> whatsapp-frontend-prod (new)
```

### SSL Certificate
- Detects port 80 in use
- Uses **webroot challenge** instead of standalone
- Stores cert in `/etc/letsencrypt/live/<domain>/`
- Copies to `nginx/ssl/` for Docker volume

### Auto-Renewal
Creates renewal hook at:
```
/etc/letsencrypt/renewal-hooks/deploy/whatsapp-<domain>.sh
```

Cron job runs daily at 2 AM:
```
0 2 * * * certbot renew --quiet
```

### Deployment
Builds and starts:
- `whatsapp-backend-prod` (Go API)
- `whatsapp-frontend-prod` (Vue.js)
- `whatsapp-mongo-prod` (MongoDB)

---

## 🔍 Troubleshooting

### Containers Not Accessible from Nginx

**Check network:**
```bash
# List networks
docker network ls

# Inspect your nginx network
docker network inspect <nginx-network>

# Should see WhatsApp containers in the list
```

**Fix:**
```bash
# Reconnect containers
docker network connect <nginx-network> whatsapp-backend-prod
docker network connect <nginx-network> whatsapp-frontend-prod
```

---

### 502 Bad Gateway

**Check containers:**
```bash
# Are containers running?
docker ps | grep whatsapp

# Check logs
docker logs whatsapp-backend-prod
docker logs whatsapp-frontend-prod
```

**Test from nginx container:**
```bash
# Enter nginx container
docker exec -it <nginx-container> sh

# Test backend
wget -qO- http://whatsapp-backend-prod:3000/health

# Test frontend
wget -qO- http://whatsapp-frontend-prod:8080/

# Exit
exit
```

**Fix:**
```bash
# Restart WhatsApp containers
docker restart whatsapp-backend-prod whatsapp-frontend-prod
```

---

### SSL Certificate Error

**Check certificate:**
```bash
# List certificates
sudo ls -la /etc/letsencrypt/live/whatsapp.agsatu.id/

# Should have:
# - fullchain.pem
# - privkey.pem
```

**Check in nginx container:**
```bash
docker exec <nginx-container> ls -la /etc/letsencrypt/live/whatsapp.agsatu.id/
```

**If missing, re-run certbot:**
```bash
# With webroot
sudo certbot certonly --webroot \
    -w /root/nginx/html \
    -d whatsapp.agsatu.id
```

---

### CORS Error

**Check backend CORS:**
```bash
docker exec whatsapp-backend-prod env | grep CORS
```

**Fix:**
```bash
# Edit .env.production
nano .env.production

# Update:
CORS_ALLOWED_ORIGIN=https://whatsapp.agsatu.id

# Restart backend
docker restart whatsapp-backend-prod
```

---

## 📊 Architecture

```
Internet (Port 80/443)
         ↓
    Nginx Docker (existing)
         ↓
   Docker Network
    ├─> Frontend:8080 (Vue.js)
    ├─> Backend:3000 (Go API)
    └─> MongoDB:27017
```

**Key Points:**
- Nginx uses **container names** as hostnames
- Docker DNS resolves: `whatsapp-backend-prod` → container IP
- No port binding to host needed
- All communication via Docker network

---

## 🎯 Useful Commands

```bash
# Check status
make status COMPOSE_FILE=docker-compose.nginx-docker.yml

# View logs
make logs COMPOSE_FILE=docker-compose.nginx-docker.yml

# Backend logs only
docker logs -f whatsapp-backend-prod

# Frontend logs only
docker logs -f whatsapp-frontend-prod

# Restart services
make restart COMPOSE_FILE=docker-compose.nginx-docker.yml

# Stop services
make down COMPOSE_FILE=docker-compose.nginx-docker.yml

# Rebuild and restart
docker compose --env-file .env.production -f docker-compose.nginx-docker.yml up -d --build
```

---

## ✅ Verification Checklist

After setup, verify:

- [ ] Containers running: `docker ps | grep whatsapp`
- [ ] In nginx network: `docker network inspect <nginx-network>`
- [ ] Backend health: `curl http://127.0.0.1:3000/health` (if exposed)
- [ ] Nginx config added: `docker exec <nginx-container> cat /etc/nginx/conf.d/whatsapp.conf`
- [ ] HTTPS working: `curl -I https://whatsapp.agsatu.id`
- [ ] API responding: `curl https://whatsapp.agsatu.id/api/health`
- [ ] Frontend loads in browser
- [ ] No CORS errors in browser console
- [ ] SSL certificate valid (no browser warning)
- [ ] Auto-renewal configured: `sudo crontab -l | grep certbot`

---

## 📝 File Locations

```
Project:
  .env.production              # Environment variables
  docker-compose.nginx-docker.yml  # Docker compose file
  nginx/nginx-docker.conf      # Nginx config template

Nginx Container:
  /etc/nginx/conf.d/whatsapp.conf  # WhatsApp nginx config

SSL:
  /etc/letsencrypt/live/whatsapp.agsatu.id/  # Certificates
  /etc/letsencrypt/renewal-hooks/deploy/whatsapp-whatsapp.agsatu.id.sh  # Renewal hook
```

---

## 🔧 Advanced: Custom Nginx Container Name

If your nginx container has custom name:

```bash
# Find container name
NGINX_CONTAINER=$(docker ps --format '{{.Names}}' | grep nginx | head -1)

# Use in commands
docker cp nginx/nginx-docker.conf $NGINX_CONTAINER:/etc/nginx/conf.d/whatsapp.conf
docker exec $NGINX_CONTAINER sed -i 's/wa.yourdomain.com/whatsapp.agsatu.id/g' /etc/nginx/conf.d/whatsapp.conf
docker exec $NGINX_CONTAINER nginx -t
docker exec $NGINX_CONTAINER nginx -s reload
```

---

## 💡 Tips

1. **Save nginx container name** for future commands:
   ```bash
   echo "export NGINX_CONTAINER=<your-nginx-container>" >> ~/.bashrc
   source ~/.bashrc
   ```

2. **Create alias** untuk commands:
   ```bash
   echo "alias wa-logs='docker logs -f whatsapp-backend-prod'" >> ~/.bashrc
   echo "alias wa-status='docker ps | grep whatsapp'" >> ~/.bashrc
   source ~/.bashrc
   ```

3. **Backup before changes:**
   ```bash
   docker exec $NGINX_CONTAINER cp /etc/nginx/conf.d/whatsapp.conf /etc/nginx/conf.d/whatsapp.conf.bak
   ```

---

**Perfect untuk nginx Docker existing!** 🚀

**Setup Time:** ~5 minutes dengan wizard
**Maintenance:** Auto SSL renewal, minimal manual work

**Last Updated:** 2025-11-24
