# Production Architecture

## Overview

Arsitektur production menggunakan Docker multi-container dengan Nginx sebagai reverse proxy untuk routing frontend dan backend.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         INTERNET                                 │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                      ┌──────────▼──────────┐
                      │   NGINX (Port 80/443)│
                      │   Reverse Proxy      │
                      └──────────┬───────────┘
                                 │
                    ┌────────────┼────────────┐
                    │            │            │
          ┌─────────▼────┐  ┌───▼─────────┐  │
          │   Frontend   │  │   Backend   │  │
          │  (Port 8080) │  │ (Port 3000) │  │
          │  Internal    │  │  Internal   │  │
          └──────────────┘  └─────┬───────┘  │
                                  │          │
                            ┌─────▼──────┐   │
                            │  MongoDB   │   │
                            │ (Port 27017)│  │
                            │  Internal  │   │
                            └────────────┘   │
                                             │
┌────────────────────────────────────────────┘
│  Docker Network: whatsapp-network
│  Subnet: 172.20.0.0/16
└────────────────────────────────────────────
```

## Container Details

### 1. Nginx Container
- **Image**: `nginx:alpine`
- **Exposed Ports**: 80 (HTTP), 443 (HTTPS)
- **Function**: Reverse proxy & load balancer
- **Routes**:
  - `https://yourdomain.com/` → Frontend Container (port 8080)
  - `https://yourdomain.com/api/*` → Backend Container (port 3000)
- **Features**:
  - SSL/TLS termination
  - Rate limiting
  - Gzip compression
  - Static file caching
  - Security headers

### 2. Frontend Container
- **Base Image**: `node:20-alpine` (build) → `nginx:alpine` (runtime)
- **Internal Port**: 8080 (tidak exposed ke public)
- **Build Process**:
  1. Install dependencies dengan `npm ci`
  2. Build Vue.js app dengan `npm run build`
  3. Copy static files ke nginx container
  4. Serve static files via nginx
- **Environment Variables**:
  - `VITE_API_BASE_URL=/api` (build-time variable)

### 3. Backend Container
- **Base Image**: `golang:1.25.1-bookworm` (build) → `debian:bookworm-slim` (runtime)
- **Internal Port**: 3000 (tidak exposed ke public)
- **Build Process**:
  1. Compile Go binary dengan CGO enabled (untuk SQLite)
  2. Compress binary dengan UPX
  3. Copy ke minimal runtime image
  4. Run sebagai non-root user
- **Volumes**:
  - `whatsapp-stores:/app/stores` (WhatsApp sessions)
  - `whatsapp-uploads:/app/uploads` (uploaded files)
  - `app-keys:/app/keys` (RSA keys)
  - `app-logs:/app/logs` (application logs)

### 4. MongoDB Container
- **Image**: `mongo:7.0`
- **Internal Port**: 27017 (tidak exposed ke public)
- **Volumes**:
  - `mongo-data:/data/db` (database persistence)
  - `mongo-config:/data/configdb` (config persistence)
- **Authentication**: Username/password dari environment

## Port Mapping

### Development vs Production

#### Development Mode
```
User Browser → Frontend:5173 (Vite dev server)
               Backend:3000 (direct access)
               MongoDB:27017 (exposed for debugging)
```

#### Production Mode
```
User Browser → Nginx:80/443 (only exposed port)
               ├─> Frontend:8080 (internal only)
               └─> Backend:3000 (internal only)
                   └─> MongoDB:27017 (internal only)
```

### Why No Direct Port Exposure?

1. **Security**: Container ports tidak exposed ke public, hanya accessible via Nginx
2. **Single Entry Point**: Semua traffic masuk lewat Nginx (easier to monitor & secure)
3. **Flexibility**: Bisa change internal ports tanpa affect users
4. **SSL Termination**: SSL handling di Nginx, backend tidak perlu handle HTTPS

## Request Flow

### Frontend Request Flow
```
1. User → https://yourdomain.com/
2. Nginx receives request on port 443
3. Nginx forwards to frontend:8080
4. Frontend nginx serves index.html
5. Browser loads Vue.js app
6. Vue router handles client-side routing
```

### API Request Flow
```
1. User browser → https://yourdomain.com/api/clients
2. Nginx receives request on port 443
3. Nginx rewrites /api/clients → /clients
4. Nginx forwards to backend:3000/clients
5. Backend processes request
6. Backend queries MongoDB at mongo:27017
7. Backend sends response to Nginx
8. Nginx sends response to user browser
```

### WebSocket Flow (if applicable)
```
1. User → wss://yourdomain.com/api/ws
2. Nginx upgrades connection to WebSocket
3. Nginx forwards to backend:3000/ws
4. Backend handles WebSocket connection
5. Bidirectional communication established
```

## Environment Variables

### Build-Time vs Runtime

#### Frontend (Build-Time)
```dockerfile
ARG VITE_API_BASE_URL=/api
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
```
- Di-inject saat `docker build`
- Hardcoded ke dalam bundled JS files
- Tidak bisa diubah setelah build

**Set via docker-compose:**
```yaml
frontend:
  build:
    args:
      - VITE_API_BASE_URL=/api
```

#### Backend (Runtime)
```yaml
environment:
  PORT: 3000
  MONGO_HOST: mongo:27017
  JWT_SECRET: ${JWT_SECRET}
```
- Di-inject saat `docker run`
- Bisa diubah tanpa rebuild
- Read from .env.production file

## Scaling Considerations

### Horizontal Scaling

#### Single Server (Current Setup)
```
Nginx (1) → Frontend (1) + Backend (1) + MongoDB (1)
```

#### Multi-Container Scaling
```yaml
# docker-compose.prod.yml
backend:
  deploy:
    replicas: 3
```

Result:
```
Nginx (1) → Frontend (1) + Backend (3) + MongoDB (1)
```

#### Multi-Server Scaling
Use Docker Swarm or Kubernetes:
```
        ┌─> Server 1 (Backend + Frontend)
Nginx → ├─> Server 2 (Backend + Frontend)
        └─> Server 3 (Backend + Frontend)
                ↓
        MongoDB Cluster (Replica Set)
```

## Security Features

### Network Isolation
```yaml
networks:
  whatsapp-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```
- All containers in isolated network
- No direct access from internet
- Only Nginx exposes ports

### Non-Root Users
- **Frontend**: Runs as `nginx` user
- **Backend**: Runs as `appuser` (UID 1000)
- **MongoDB**: Runs as `mongodb` user

### Rate Limiting (Nginx)
```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=3r/m;
limit_req_zone $binary_remote_addr zone=upload_limit:10m rate=5r/s;
```

### Security Headers (Nginx)
```nginx
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
```

## Resource Management

### Resource Limits
```yaml
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 2G
    reservations:
      cpus: '1'
      memory: 1G
```

### Recommended Allocation

| Service   | CPU Limit | Memory Limit | CPU Reserved | Memory Reserved |
|-----------|-----------|--------------|--------------|-----------------|
| MongoDB   | 2 cores   | 2 GB         | 1 core       | 1 GB            |
| Backend   | 2 cores   | 2 GB         | 0.5 core     | 512 MB          |
| Frontend  | 0.5 core  | 512 MB       | 0.25 core    | 256 MB          |
| Nginx     | 1 core    | 512 MB       | 0.25 core    | 256 MB          |
| **Total** | **5.5**   | **5 GB**     | **2 cores**  | **2 GB**        |

**Minimum Server**: 4 vCPU, 8 GB RAM
**Recommended Server**: 8 vCPU, 16 GB RAM

## Volumes & Data Persistence

### Volume Mapping
```yaml
volumes:
  mongo-data:           # MongoDB database files
  mongo-config:         # MongoDB configuration
  whatsapp-stores:      # WhatsApp session data
  whatsapp-uploads:     # Uploaded media files
  app-keys:             # RSA keys for JWT
  app-logs:             # Application logs
  nginx-logs:           # Nginx access/error logs
  nginx-cache:          # Nginx cache
```

### Backup Strategy
```bash
# Backup volumes
docker run --rm \
  -v whatsapp_whatsapp-stores:/data \
  -v $(pwd)/backups:/backup \
  alpine tar czf /backup/stores_$(date +%Y%m%d).tar.gz /data
```

## Health Checks

### Nginx
```yaml
healthcheck:
  test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:80/health"]
  interval: 30s
  timeout: 10s
```

### Backend
```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:3000/health"]
  interval: 30s
  timeout: 10s
  start_period: 40s
```

### MongoDB
```yaml
healthcheck:
  test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
  interval: 10s
  timeout: 5s
```

## Deployment Workflow

### 1. Initial Setup
```bash
# Setup environment
cp .env.production.example .env.production
nano .env.production

# Setup SSL
mkdir -p nginx/ssl
# Copy SSL certificates

# Build & deploy
make build
make up
```

### 2. Updates (Zero Downtime)
```bash
# Pull latest code
git pull origin main

# Build new images
docker-compose -f docker-compose.prod.yml build

# Rolling update
docker-compose -f docker-compose.prod.yml up -d --no-deps --build backend
docker-compose -f docker-compose.prod.yml up -d --no-deps --build frontend
docker-compose -f docker-compose.prod.yml restart nginx
```

### 3. Rollback
```bash
# Revert to previous image
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml up -d
```

## Monitoring

### Container Stats
```bash
docker stats
```

### Logs
```bash
# All services
docker-compose -f docker-compose.prod.yml logs -f

# Specific service
docker-compose -f docker-compose.prod.yml logs -f backend

# Export logs
docker-compose -f docker-compose.prod.yml logs --no-color > logs.txt
```

### Health Check
```bash
# Overall health
curl http://localhost/health

# Backend health
curl http://localhost/api/health
```

## Troubleshooting

### Common Issues

#### 1. Frontend shows 404 on refresh
**Cause**: Nginx not configured for SPA routing
**Solution**: Check `try_files` directive in nginx.conf

#### 2. API calls fail with CORS error
**Cause**: Wrong CORS_ALLOWED_ORIGIN
**Solution**: Set correct frontend URL in .env.production

#### 3. Backend can't connect to MongoDB
**Cause**: MongoDB not ready or wrong credentials
**Solution**: Check MongoDB health and credentials

#### 4. SSL certificate error
**Cause**: Invalid or expired certificate
**Solution**: Renew certificate with Let's Encrypt

## Performance Optimization

### 1. Enable BuildKit
```bash
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
```

### 2. Multi-stage Build Caching
Already implemented in Dockerfiles

### 3. Nginx Caching
```nginx
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=api_cache:10m;
```

### 4. Binary Compression
Backend binary compressed with UPX (30-50% size reduction)

## Conclusion

Production architecture dirancang dengan prinsip:
- **Security First**: Isolated network, non-root users, SSL
- **Scalability**: Easy to scale horizontally
- **Maintainability**: Clear separation of concerns
- **Performance**: Optimized builds, caching, compression
- **Reliability**: Health checks, auto-restart, backups

---

**Last Updated**: 2025-11-24
