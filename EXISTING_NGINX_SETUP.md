# Setup dengan Nginx yang Sudah Ada

Panduan deploy aplikasi WhatsApp ke server yang **sudah punya nginx/service lain** tanpa ganggu service existing.

## Scenario

Server Anda sudah ada:
- ✅ Nginx running (port 80/443 sudah dipakai)
- ✅ Service lain (website, API, dll)
- ❌ Tidak mau ganggu service yang sudah ada

## Solution: Use Existing Nginx as Reverse Proxy

```
Internet
    ↓
Existing Nginx (Port 80/443)
    ├─> Existing Services (tetap jalan)
    └─> WhatsApp App (via subdomain)
            ├─> Docker Frontend (localhost:8080)
            └─> Docker Backend (localhost:3000)
```

**Key Points:**
- Docker containers exposed ke **localhost only** (127.0.0.1)
- Nginx existing proxy ke Docker containers
- No port conflict!
- Existing services tidak terganggu

---

## Step-by-Step Setup

### 1. Deploy Docker Containers (Without Nginx)

```bash
# Use special compose file (no nginx container)
docker-compose -f docker-compose.existing-nginx.yml build
docker-compose -f docker-compose.existing-nginx.yml up -d

# Check containers
docker-compose -f docker-compose.existing-nginx.yml ps
```

**Ports exposed:**
- Backend: `127.0.0.1:3000` (localhost only)
- Frontend: `127.0.0.1:8080` (localhost only)
- MongoDB: `127.0.0.1:27017` (localhost only)

### 2. Test Docker Containers

```bash
# Test backend
curl http://localhost:3000/health

# Test frontend
curl http://localhost:8080/

# Should work from localhost!
```

### 3. Configure Existing Nginx

#### A. Copy Config File

```bash
# Copy config template
sudo cp nginx/existing-nginx-site.conf /etc/nginx/sites-available/whatsapp

# Edit dengan domain Anda
sudo nano /etc/nginx/sites-available/whatsapp
```

#### B. Update Domain

Edit file, ganti `wa.yourdomain.com` dengan subdomain Anda:

```nginx
server_name wa.yourdomain.com;  # Ganti ini!
```

#### C. Update SSL Paths (if needed)

```nginx
ssl_certificate /etc/letsencrypt/live/wa.yourdomain.com/fullchain.pem;
ssl_certificate_key /etc/letsencrypt/live/wa.yourdomain.com/privkey.pem;
```

#### D. Enable Site

```bash
# Create symlink
sudo ln -s /etc/nginx/sites-available/whatsapp /etc/nginx/sites-enabled/

# Test config
sudo nginx -t

# Reload nginx (no downtime!)
sudo systemctl reload nginx
```

### 4. Setup SSL Certificate

#### Option A: Let's Encrypt (Recommended)

```bash
# Install certbot if not installed
sudo apt install certbot python3-certbot-nginx

# Get certificate
sudo certbot certonly --nginx -d wa.yourdomain.com

# Certificates will be at:
# /etc/letsencrypt/live/wa.yourdomain.com/fullchain.pem
# /etc/letsencrypt/live/wa.yourdomain.com/privkey.pem
```

#### Option B: Existing Wildcard Cert

If you have `*.yourdomain.com` cert:

```nginx
ssl_certificate /path/to/wildcard-cert.pem;
ssl_certificate_key /path/to/wildcard-key.pem;
```

### 5. DNS Configuration

Add A record for subdomain:

```
Type: A
Name: wa
Value: YOUR_SERVER_IP
TTL: 3600
```

Wait for DNS propagation (5-30 minutes).

### 6. Test Deployment

```bash
# Test from server
curl https://wa.yourdomain.com/health

# Test from browser
open https://wa.yourdomain.com
```

---

## Architecture

### Full Stack

```
Internet
    ↓
Existing Nginx (:80, :443)
    ├─> example.com → Existing Website
    ├─> api.example.com → Existing API
    └─> wa.yourdomain.com → WhatsApp App
            ↓
        Proxy to localhost
            ├─> Frontend (127.0.0.1:8080)
            └─> Backend (127.0.0.1:3000)
                    ↓
                MongoDB (127.0.0.1:27017)
```

### Port Bindings

| Service | Container Port | Host Binding | Public Access |
|---------|---------------|--------------|---------------|
| Frontend | 8080 | 127.0.0.1:8080 | Via nginx only |
| Backend | 3000 | 127.0.0.1:3000 | Via nginx only |
| MongoDB | 27017 | 127.0.0.1:27017 | Localhost only |
| Nginx (Docker) | N/A | Not deployed | N/A |

**Security:**
- ✅ Containers only accessible from localhost
- ✅ No direct internet access to containers
- ✅ All traffic goes through nginx (with SSL, rate limiting, etc)

---

## Configuration Files

### docker-compose.existing-nginx.yml

Located at project root. Key features:
- No nginx container
- Backend exposed to `127.0.0.1:3000`
- Frontend exposed to `127.0.0.1:8080`
- MongoDB exposed to `127.0.0.1:27017`

### nginx/existing-nginx-site.conf

Nginx config for existing nginx. Features:
- Upstream definitions for Docker containers
- SSL configuration
- Rate limiting
- Logging
- Frontend & API routing

---

## Maintenance

### View Logs

```bash
# Docker container logs
docker-compose -f docker-compose.existing-nginx.yml logs -f

# Nginx logs
sudo tail -f /var/log/nginx/whatsapp-access.log
sudo tail -f /var/log/nginx/whatsapp-error.log
```

### Update Application

```bash
# Pull latest code
git pull

# Rebuild and restart
docker-compose -f docker-compose.existing-nginx.yml build
docker-compose -f docker-compose.existing-nginx.yml up -d

# No need to reload nginx (containers auto-reconnect)
```

### Restart Services

```bash
# Restart Docker containers
docker-compose -f docker-compose.existing-nginx.yml restart

# Reload nginx (if config changed)
sudo nginx -t && sudo systemctl reload nginx
```

### Stop Application

```bash
# Stop Docker containers
docker-compose -f docker-compose.existing-nginx.yml down

# Disable nginx site
sudo rm /etc/nginx/sites-enabled/whatsapp
sudo systemctl reload nginx
```

---

## Troubleshooting

### 502 Bad Gateway

**Problem:** Nginx can't reach Docker containers

**Check:**
```bash
# Are containers running?
docker-compose -f docker-compose.existing-nginx.yml ps

# Can localhost access backend?
curl http://localhost:3000/health

# Can localhost access frontend?
curl http://localhost:8080/

# Check nginx error log
sudo tail -f /var/log/nginx/whatsapp-error.log
```

**Fix:**
```bash
# Restart containers
docker-compose -f docker-compose.existing-nginx.yml restart

# Or restart specific service
docker-compose -f docker-compose.existing-nginx.yml restart backend
```

### Port Already in Use

**Problem:** Port 3000 or 8080 already used

**Check:**
```bash
sudo lsof -i :3000
sudo lsof -i :8080
```

**Fix:** Change ports in `docker-compose.existing-nginx.yml`:
```yaml
backend:
  ports:
    - "127.0.0.1:3001:3000"  # Change 3000 to 3001

frontend:
  ports:
    - "127.0.0.1:8081:8080"  # Change 8080 to 8081
```

Then update nginx config upstream:
```nginx
upstream whatsapp_backend {
    server 127.0.0.1:3001;  # Update port
}
```

### SSL Certificate Issues

**Problem:** SSL certificate not found

**Check:**
```bash
# List certificates
sudo certbot certificates

# Check file exists
ls -la /etc/letsencrypt/live/wa.yourdomain.com/
```

**Fix:**
```bash
# Request new certificate
sudo certbot certonly --nginx -d wa.yourdomain.com

# Or use certbot auto-config
sudo certbot --nginx -d wa.yourdomain.com
```

### DNS Not Resolving

**Check:**
```bash
# Check DNS
dig wa.yourdomain.com

# Check from server
curl -I https://wa.yourdomain.com
```

**Wait:** DNS propagation takes 5-30 minutes

---

## Advantages

✅ **No Port Conflicts**
- Docker containers use localhost only
- Existing nginx handles all public traffic

✅ **Zero Downtime**
- Existing services not affected
- Can update Docker containers without nginx reload

✅ **Centralized SSL**
- One nginx manages all SSL certificates
- Easy to renew with certbot

✅ **Better Security**
- Containers not directly exposed
- All traffic filtered through nginx
- Rate limiting, logging centralized

✅ **Easy Scaling**
- Can add more upstream servers
- Load balancing ready

---

## Alternative: Different Port

If you don't want to use subdomain:

### Option 1: Custom Port
```yaml
# docker-compose.prod.yml
nginx:
  ports:
    - "8443:443"
```

Access: `https://yourdomain.com:8443`

### Option 2: Path-based Routing

Use existing nginx with path:
```nginx
location /whatsapp/ {
    proxy_pass http://127.0.0.1:8080/;
}

location /whatsapp/api/ {
    proxy_pass http://127.0.0.1:3000/;
}
```

Access:
- Frontend: `https://yourdomain.com/whatsapp/`
- API: `https://yourdomain.com/whatsapp/api/`

---

## Summary Commands

```bash
# Deploy
docker-compose -f docker-compose.existing-nginx.yml up -d

# Configure nginx
sudo cp nginx/existing-nginx-site.conf /etc/nginx/sites-available/whatsapp
sudo ln -s /etc/nginx/sites-available/whatsapp /etc/nginx/sites-enabled/
sudo nano /etc/nginx/sites-available/whatsapp  # Edit domain
sudo nginx -t
sudo systemctl reload nginx

# SSL
sudo certbot certonly --nginx -d wa.yourdomain.com

# Test
curl https://wa.yourdomain.com/health
```

---

## Need Help?

- Check logs: `docker-compose -f docker-compose.existing-nginx.yml logs -f`
- Test nginx: `sudo nginx -t`
- Check containers: `docker ps`
- Test locally: `curl http://localhost:3000/health`

---

**Perfect untuk server yang sudah ada service lain!** 🚀

**Last Updated**: 2025-11-24
