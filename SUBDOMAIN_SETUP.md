# Setup dengan Subdomain

Panduan deploy WhatsApp API menggunakan subdomain (contoh: `wa.yourdomain.com`).

## Scenario

Anda ingin deploy aplikasi dengan subdomain:
- ✅ Main domain untuk website utama: `yourdomain.com`
- ✅ Subdomain untuk WhatsApp: `wa.yourdomain.com`
- ✅ 1 pintu akses (frontend + API di subdomain sama)

---

## Architecture

```
yourdomain.com           → Website/Service lain
wa.yourdomain.com/       → WhatsApp Frontend
wa.yourdomain.com/api/   → WhatsApp API
```

**Benefits:**
- ✅ 1 SSL certificate (untuk subdomain)
- ✅ No CORS issues
- ✅ Clean separation dari main domain
- ✅ Easy to manage

---

## Option A: Standalone Nginx (Port 80/443 Available)

Jika server dedicated atau port 80/443 tidak dipakai.

### 1. Setup DNS

Add A record untuk subdomain:

```
Type: A
Name: wa
Value: YOUR_SERVER_IP
TTL: 3600
```

Result: `wa.yourdomain.com` → `YOUR_SERVER_IP`

**Test DNS:**
```bash
dig wa.yourdomain.com
# Should return your server IP
```

### 2. Update Environment

Edit `.env.production`:

```bash
# CORS - use subdomain
CORS_ALLOWED_ORIGIN=https://wa.yourdomain.com

# Other settings
MONGO_USER=admin
MONGO_PASS=your_strong_password
JWT_SECRET=your_secret_key
```

### 3. Update Nginx Config

Edit [nginx/nginx.prod.conf](nginx/nginx.prod.conf):

Find lines with `server_name` and update:

```nginx
server {
    listen 80;
    server_name wa.yourdomain.com;  # ← Update this
    # ...
}

server {
    listen 443 ssl http2;
    server_name wa.yourdomain.com;  # ← Update this
    # ...
}
```

### 4. Setup SSL Certificate

#### Using Let's Encrypt (Recommended)

```bash
# Install certbot
sudo apt install certbot

# Stop nginx container temporarily
docker compose -f docker-compose.prod.yml stop nginx

# Get certificate
sudo certbot certonly --standalone -d wa.yourdomain.com

# Copy certificates to nginx volume
sudo cp /etc/letsencrypt/live/wa.yourdomain.com/fullchain.pem nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/wa.yourdomain.com/privkey.pem nginx/ssl/key.pem
sudo chmod 644 nginx/ssl/*.pem

# Start nginx
docker compose -f docker-compose.prod.yml start nginx
```

#### Or Self-Signed (Development)

```bash
mkdir -p nginx/ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/key.pem \
  -out nginx/ssl/cert.pem \
  -subj "/C=ID/ST=Jakarta/L=Jakarta/O=Company/CN=wa.yourdomain.com"
```

### 5. Deploy

```bash
# Build images
docker compose -f docker-compose.prod.yml build

# Deploy
docker compose -f docker-compose.prod.yml up -d

# Check status
docker compose -f docker-compose.prod.yml ps
```

### 6. Test

```bash
# Health check
curl https://wa.yourdomain.com/health

# Frontend
curl https://wa.yourdomain.com/

# API
curl https://wa.yourdomain.com/api/health
```

### 7. Auto-Renew SSL (Let's Encrypt)

```bash
# Test renewal
sudo certbot renew --dry-run

# Add cron job for auto-renewal
sudo crontab -e

# Add this line (renew daily at 2 AM)
0 2 * * * certbot renew --quiet --deploy-hook "docker compose -f /path/to/whatsapp/docker-compose.prod.yml restart nginx"
```

---

## Option B: With Existing Nginx (Port 80/443 Already Used)

Jika server sudah ada nginx/service di port 80/443.

### For Nginx in Host

Use [docker-compose.existing-nginx.yml](docker-compose.existing-nginx.yml)

**Quick Steps:**

1. **Deploy Docker (no nginx container)**
```bash
docker compose -f docker-compose.existing-nginx.yml up -d
```

2. **Add Nginx Config**
```bash
sudo cp nginx/existing-nginx-site.conf /etc/nginx/sites-available/whatsapp
sudo nano /etc/nginx/sites-available/whatsapp
```

Update `server_name`:
```nginx
server_name wa.yourdomain.com;  # ← Your subdomain
```

3. **Enable Site**
```bash
sudo ln -s /etc/nginx/sites-available/whatsapp /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

4. **SSL Certificate**
```bash
sudo certbot --nginx -d wa.yourdomain.com
```

### For Nginx in Docker

Use [docker-compose.nginx-docker.yml](docker-compose.nginx-docker.yml)

**Quick Steps:**

1. **Find nginx network**
```bash
docker inspect nginx-server | grep -A 5 Networks
```

2. **Update compose**
```yaml
networks:
  nginx-network:
    external: true
    name: YOUR_NGINX_NETWORK  # ← Update this
```

3. **Deploy**
```bash
docker compose --env-file .env.production -f docker-compose.nginx-docker.yml up -d
```

4. **Add config to nginx**
```bash
# Edit nginx/nginx-docker.conf
# Update server_name to wa.yourdomain.com

# Copy to nginx container
docker cp nginx/nginx-docker.conf nginx-server:/etc/nginx/conf.d/whatsapp.conf
docker exec nginx-server nginx -t
docker exec nginx-server nginx -s reload
```

---

## Option C: Separate Subdomains (Advanced)

Frontend dan API di subdomain berbeda:
- `app.yourdomain.com` → Frontend
- `api.yourdomain.com` → Backend API

### Setup

1. **DNS Records**
```
Type: A, Name: app, Value: YOUR_SERVER_IP
Type: A, Name: api, Value: YOUR_SERVER_IP
```

2. **Use nginx.subdomain.conf**

Edit [nginx/nginx.subdomain.conf](nginx/nginx.subdomain.conf):

```nginx
# Frontend
server_name app.yourdomain.com;
ssl_certificate /etc/nginx/ssl/app-cert.pem;
ssl_certificate_key /etc/nginx/ssl/app-key.pem;

# API
server_name api.yourdomain.com;
ssl_certificate /etc/nginx/ssl/api-cert.pem;
ssl_certificate_key /etc/nginx/ssl/api-key.pem;
```

3. **Update Frontend Build**

Edit `fe/.env.production`:
```bash
# Point to API subdomain
VITE_API_BASE_URL=https://api.yourdomain.com
```

**Note:** Need rebuild frontend after changing this!

4. **Update Backend CORS**

Edit `.env.production`:
```bash
# Allow frontend subdomain
CORS_ALLOWED_ORIGIN=https://app.yourdomain.com
```

5. **SSL Certificates**

Get certificates for both subdomains:
```bash
sudo certbot certonly --standalone -d app.yourdomain.com
sudo certbot certonly --standalone -d api.yourdomain.com

# Copy certificates
cp /etc/letsencrypt/live/app.yourdomain.com/fullchain.pem nginx/ssl/app-cert.pem
cp /etc/letsencrypt/live/app.yourdomain.com/privkey.pem nginx/ssl/app-key.pem
cp /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem nginx/ssl/api-cert.pem
cp /etc/letsencrypt/live/api.yourdomain.com/privkey.pem nginx/ssl/api-key.pem
```

6. **Deploy**

Use custom compose:
```bash
# Update docker-compose to use nginx.subdomain.conf
docker compose -f docker-compose.prod.yml up -d
```

---

## Comparison

| Setup | Domains | SSL Certs | CORS | Complexity |
|-------|---------|-----------|------|------------|
| Single Subdomain | wa.yourdomain.com | 1 | No issues | ⭐ Simple |
| Separate Subdomains | app + api | 2 | Need config | ⭐⭐ Medium |
| With Wildcard | *.yourdomain.com | 1 wildcard | No issues | ⭐⭐⭐ Advanced |

**Recommendation:** Use **Single Subdomain** (Option A or B) untuk most cases.

---

## Wildcard SSL (Bonus)

Jika punya wildcard SSL (`*.yourdomain.com`):

### Benefits
- ✅ 1 certificate untuk semua subdomain
- ✅ wa.yourdomain.com works
- ✅ api.yourdomain.com works
- ✅ app.yourdomain.com works
- ✅ Any subdomain works!

### Setup

```bash
# Get wildcard cert (DNS challenge required)
sudo certbot certonly --manual --preferred-challenges dns -d "*.yourdomain.com"

# Follow instructions to add TXT record
# Then certificates at:
# /etc/letsencrypt/live/yourdomain.com/fullchain.pem
# /etc/letsencrypt/live/yourdomain.com/privkey.pem
```

### Nginx Config

```nginx
server {
    listen 443 ssl http2;
    server_name wa.yourdomain.com;  # Any subdomain works!

    ssl_certificate /etc/nginx/ssl/wildcard-cert.pem;
    ssl_certificate_key /etc/nginx/ssl/wildcard-key.pem;
    # ...
}
```

---

## Testing Checklist

After deployment:

- [ ] DNS resolves correctly: `dig wa.yourdomain.com`
- [ ] SSL certificate valid: `curl -I https://wa.yourdomain.com`
- [ ] Frontend loads: `https://wa.yourdomain.com`
- [ ] API responds: `https://wa.yourdomain.com/api/health`
- [ ] No CORS errors in browser console
- [ ] Can login and use features
- [ ] SSL auto-renewal configured

---

## Troubleshooting

### DNS Not Resolving

**Problem:** `wa.yourdomain.com` not found

**Check:**
```bash
dig wa.yourdomain.com
nslookup wa.yourdomain.com
```

**Fix:** Wait for DNS propagation (5-30 minutes) or check DNS records in domain provider.

### SSL Certificate Error

**Problem:** Browser shows certificate invalid

**Check:**
```bash
# Check certificate
openssl x509 -in nginx/ssl/cert.pem -text -noout | grep Subject

# Should match: CN=wa.yourdomain.com
```

**Fix:**
- Regenerate certificate with correct domain
- Make sure nginx config has correct `server_name`

### CORS Error

**Problem:** Browser console shows CORS error

**Check Backend:**
```bash
# Check CORS_ALLOWED_ORIGIN
docker compose -f docker-compose.prod.yml exec backend env | grep CORS
```

**Fix:**
```bash
# Update .env.production
CORS_ALLOWED_ORIGIN=https://wa.yourdomain.com

# Restart backend
docker compose -f docker-compose.prod.yml restart backend
```

### 502 Bad Gateway

**Problem:** Nginx can't reach containers

**Check:**
```bash
# Are containers running?
docker compose -f docker-compose.prod.yml ps

# Test backend health
docker compose -f docker-compose.prod.yml exec backend curl http://localhost:3000/health
```

**Fix:**
```bash
# Restart containers
docker compose -f docker-compose.prod.yml restart
```

---

## Quick Reference

### Single Subdomain (Recommended)

```bash
# 1. DNS
# Add A record: wa → YOUR_SERVER_IP

# 2. Update nginx config
# server_name wa.yourdomain.com;

# 3. SSL
sudo certbot certonly --standalone -d wa.yourdomain.com

# 4. Deploy
docker compose -f docker-compose.prod.yml up -d

# 5. Test
curl https://wa.yourdomain.com/health
```

### Separate Subdomains

```bash
# 1. DNS
# Add A record: app → YOUR_SERVER_IP
# Add A record: api → YOUR_SERVER_IP

# 2. Update configs
# nginx.subdomain.conf: server_name app/api
# fe/.env.production: VITE_API_BASE_URL=https://api.yourdomain.com
# .env.production: CORS_ALLOWED_ORIGIN=https://app.yourdomain.com

# 3. SSL
sudo certbot certonly --standalone -d app.yourdomain.com -d api.yourdomain.com

# 4. Rebuild frontend (env changed!)
docker compose -f docker-compose.prod.yml build frontend

# 5. Deploy
docker compose -f docker-compose.prod.yml up -d
```

---

## Summary

**Easiest Setup:** Single Subdomain
- DNS: `wa.yourdomain.com`
- Access: `https://wa.yourdomain.com/` (frontend), `https://wa.yourdomain.com/api/` (API)
- SSL: 1 certificate
- CORS: No issues

**Just:**
1. Add DNS record
2. Update `server_name` in nginx config
3. Get SSL certificate
4. Deploy!

---

**Perfect untuk production dengan subdomain!** 🚀

**Last Updated**: 2025-11-24
