# Setup dengan Nginx Docker yang Sudah Ada

Panduan deploy aplikasi WhatsApp ke server yang **sudah punya nginx di Docker** tanpa ganggu setup existing.

## Scenario

Anda punya:
```bash
docker ps | grep nginx
000ec4f9e630   system-nginx   ...   0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp   nginx-server
```

- ✅ Nginx running di Docker (bukan host)
- ✅ Port 80/443 sudah dipakai nginx docker
- ✅ Ada service/website lain via nginx docker
- ❌ Tidak mau ganggu setup existing

## Solution: Join Docker Network

**Key Concept:**
- WhatsApp containers join **same network** dengan nginx docker
- Nginx docker bisa akses WhatsApp containers via **container name** (Docker DNS)
- No port binding ke host needed!

---

## Architecture

```
Internet
    ↓
Nginx Docker (Port 80/443)
    ├─> Existing Services (tidak terganggu!)
    └─> wa.yourdomain.com
            ↓
        Docker Network
            ├─> whatsapp-frontend-prod:8080
            └─> whatsapp-backend-prod:3000
                    └─> mongo:27017 (internal network)
```

**Network Topology:**
```
nginx-network (shared)
    ├─> nginx-server (existing)
    ├─> whatsapp-frontend-prod (new)
    └─> whatsapp-backend-prod (new)

whatsapp-internal (isolated)
    ├─> whatsapp-backend-prod
    └─> mongo
```

---

## Step-by-Step Setup

### 1. Find Nginx Docker Network

```bash
# Check nginx container networks
docker inspect nginx-server | grep -A 10 Networks

# Or list all networks
docker network ls

# Inspect specific network
docker network inspect <network_name>
```

**Common network names:**
- `bridge` (default)
- `nginx_default`
- `nginx-network`
- Custom name

**Example output:**
```json
"Networks": {
    "nginx-network": {
        "IPAddress": "172.18.0.2"
    }
}
```

### 2. Update docker-compose Configuration

Edit `docker-compose.nginx-docker.yml`:

```yaml
networks:
  nginx-network:
    external: true
    name: YOUR_NGINX_NETWORK_NAME  # ← Ganti ini!
```

**Example:**
```yaml
networks:
  nginx-network:
    external: true
    name: nginx-network  # Sesuaikan dengan network nginx Anda
```

### 3. Deploy WhatsApp Containers

```bash
# Build images
docker-compose -f docker-compose.nginx-docker.yml build

# Deploy
docker-compose -f docker-compose.nginx-docker.yml up -d

# Check containers
docker-compose -f docker-compose.nginx-docker.yml ps
```

**Verify containers in same network:**
```bash
# Check backend is in nginx network
docker inspect whatsapp-backend-prod | grep -A 5 Networks

# Should show both networks:
# - whatsapp-internal
# - nginx-network
```

### 4. Test Container Communication

From nginx container, test if can reach WhatsApp containers:

```bash
# Enter nginx container
docker exec -it nginx-server sh

# Test backend (use container name!)
wget -qO- http://whatsapp-backend-prod:3000/health

# Test frontend
wget -qO- http://whatsapp-frontend-prod:8080/

# Exit
exit
```

Should return success! Docker DNS resolves container names.

### 5. Add Config to Nginx Docker

#### Option A: Volume Mount (Recommended)

If your nginx has config volume:

```bash
# Copy config to nginx volume
docker cp nginx/nginx-docker.conf nginx-server:/etc/nginx/conf.d/whatsapp.conf

# Test config
docker exec nginx-server nginx -t

# Reload nginx
docker exec nginx-server nginx -s reload
```

#### Option B: Edit Main Config

If nginx config is in volume:

```bash
# Find nginx config volume
docker inspect nginx-server | grep -A 5 Mounts

# Edit config file in volume
# Add: include /etc/nginx/conf.d/*.conf;
```

#### Option C: Rebuild Nginx (if needed)

If you manage nginx via docker-compose:

1. Add volume mount in nginx compose:
```yaml
nginx:
  volumes:
    - ./nginx-configs:/etc/nginx/conf.d
```

2. Copy config:
```bash
cp nginx/nginx-docker.conf /path/to/nginx-configs/whatsapp.conf
```

3. Restart nginx:
```bash
docker-compose restart nginx
```

### 6. Update Domain in Config

Edit config file (wherever you placed it):

```nginx
server_name wa.yourdomain.com;  # ← Ganti ini!
```

### 7. Setup SSL Certificate

#### If Nginx Has Certbot

```bash
# Enter nginx container
docker exec -it nginx-server sh

# Run certbot (if installed)
certbot certonly --webroot -w /var/www/html -d wa.yourdomain.com

# Or if certbot not in container, run from host
docker run -it --rm \
  -v nginx_ssl:/etc/letsencrypt \
  -v nginx_www:/var/www/html \
  certbot/certbot certonly \
  --webroot -w /var/www/html \
  -d wa.yourdomain.com
```

#### Manual Certificate

Copy cert to nginx volume:

```bash
# Copy cert files
docker cp cert.pem nginx-server:/etc/nginx/ssl/wa.yourdomain.com/fullchain.pem
docker cp key.pem nginx-server:/etc/nginx/ssl/wa.yourdomain.com/privkey.pem

# Reload nginx
docker exec nginx-server nginx -s reload
```

### 8. DNS Configuration

Add A record:

```
Type: A
Name: wa
Value: YOUR_SERVER_IP
TTL: 3600
```

### 9. Test Deployment

```bash
# From outside
curl https://wa.yourdomain.com/health

# From browser
open https://wa.yourdomain.com
```

---

## Configuration Examples

### Upstream Configuration

**Key Point:** Use **container name** as hostname!

```nginx
upstream whatsapp_backend {
    # Container name (not IP, not localhost!)
    server whatsapp-backend-prod:3000;
}

upstream whatsapp_frontend {
    # Container name (not IP, not localhost!)
    server whatsapp-frontend-prod:8080;
}
```

Docker DNS automatically resolves container names within the same network.

### Location Configuration

Same as regular nginx, but proxy to container:

```nginx
location /api/ {
    rewrite ^/api/(.*) /$1 break;
    proxy_pass http://whatsapp_backend;  # upstream name
    # ... other proxy settings
}

location / {
    proxy_pass http://whatsapp_frontend;  # upstream name
    # ... other proxy settings
}
```

---

## Troubleshooting

### 502 Bad Gateway

**Problem:** Nginx can't reach containers

**Check:**
```bash
# Are containers running?
docker-compose -f docker-compose.nginx-docker.yml ps

# Are they in same network?
docker network inspect nginx-network

# Test from nginx container
docker exec -it nginx-server sh
wget -qO- http://whatsapp-backend-prod:3000/health
```

**Fix:**
```bash
# Ensure correct network name in compose file
# Restart containers
docker-compose -f docker-compose.nginx-docker.yml restart
```

### Container Name Not Resolving

**Problem:** DNS not working

**Possible causes:**
1. Containers not in same network
2. Wrong network name in compose
3. Typo in container name

**Fix:**
```bash
# Check network
docker inspect whatsapp-backend-prod | grep -A 5 Networks

# Should show nginx-network

# If not, update compose and recreate
docker-compose -f docker-compose.nginx-docker.yml down
docker-compose -f docker-compose.nginx-docker.yml up -d
```

### SSL Certificate Issues

**Problem:** Certificate not found in nginx

**Check:**
```bash
# List certificates in nginx container
docker exec nginx-server ls -la /etc/nginx/ssl/

# Check nginx error log
docker logs nginx-server
```

**Fix:**
```bash
# Copy cert to nginx container
docker cp /path/to/cert.pem nginx-server:/etc/nginx/ssl/
docker cp /path/to/key.pem nginx-server:/etc/nginx/ssl/

# Or use volume mount
# Add to nginx docker-compose:
volumes:
  - ./ssl:/etc/nginx/ssl
```

### Network Already Exists Error

**Problem:** Network name conflict

**Fix:**
```bash
# Use existing network (don't create)
networks:
  nginx-network:
    external: true  # ← Important!
    name: actual-network-name
```

---

## Advantages

✅ **Clean Separation**
- All containers in Docker
- No host port conflicts
- Easy to manage

✅ **Docker Native**
- Uses Docker DNS
- Network isolation
- Container-to-container communication

✅ **Scalable**
- Can add more containers
- Load balancing ready
- Easy to update

✅ **Secure**
- Containers not exposed to host
- Only nginx has public access
- Internal communication via private network

---

## Network Diagram

```
┌─────────────────────────────────────────────┐
│              Internet                       │
└────────────────┬────────────────────────────┘
                 │
                 ▼
         ┌───────────────┐
         │  nginx-server │ (Port 80/443)
         │  (existing)   │
         └───────┬───────┘
                 │
         ┌───────┴──────────────────┐
         │   nginx-network          │
         │   (shared Docker network)│
         ├──────────────────────────┤
         │                          │
    ┌────▼─────┐            ┌──────▼──────┐
    │ frontend │            │  backend    │
    │  :8080   │            │   :3000     │
    └──────────┘            └──────┬──────┘
                                   │
                            ┌──────┴──────────┐
                            │ whatsapp-internal│
                            │  (isolated)      │
                            └──────┬───────────┘
                                   │
                            ┌──────▼──────┐
                            │   mongo     │
                            │  :27017     │
                            └─────────────┘
```

---

## Comparison: Docker Network vs Host Binding

| Aspect | Docker Network | Host Binding |
|--------|---------------|--------------|
| Container Access | Via container name | Via localhost:port |
| Port Binding | Not needed | Required (127.0.0.1:3000) |
| DNS | Docker DNS | localhost |
| Security | Better (isolated) | Good |
| Flexibility | Higher | Medium |
| Complexity | Medium | Low |

**Recommendation:** Use Docker Network for production!

---

## Summary Commands

```bash
# 1. Find nginx network
docker network ls
docker inspect nginx-server | grep -A 5 Networks

# 2. Update compose file
# Edit: networks.nginx-network.name

# 3. Deploy
docker-compose -f docker-compose.nginx-docker.yml up -d

# 4. Test connectivity
docker exec nginx-server wget -qO- http://whatsapp-backend-prod:3000/health

# 5. Add nginx config
docker cp nginx/nginx-docker.conf nginx-server:/etc/nginx/conf.d/whatsapp.conf
docker exec nginx-server nginx -t
docker exec nginx-server nginx -s reload

# 6. Test from outside
curl https://wa.yourdomain.com/health
```

---

## Files Created

- `docker-compose.nginx-docker.yml` - Compose with external network
- `nginx/nginx-docker.conf` - Config for nginx docker (use container names)
- `NGINX_DOCKER_SETUP.md` - This guide

---

**Perfect untuk nginx yang running di Docker!** 🐳

**Last Updated**: 2025-11-24
