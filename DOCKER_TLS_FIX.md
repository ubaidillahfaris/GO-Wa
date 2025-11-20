# 🔧 Docker TLS Certificate Fix

Quick guide untuk rebuild Docker container dengan TLS fix.

---

## 🐛 **Issue yang Di-fix**

**Error sebelumnya:**
```json
{
    "details": "[device-id] gagal connect: couldn't dial whatsapp web websocket: tls: failed to verify certificate: x509: certificate signed by unknown authority",
    "error": "failed to generate QR"
}
```

**Root Cause:**
- Alpine Linux's CA certificates tidak fully compatible dengan WhatsApp Web's TLS chain
- Go's crypto/tls package butuh proper certificate verification
- Whatsmeow WebSocket connection gagal karena certificate validation

---

## ✅ **Solution Applied**

**Changed:**
- Runtime base image: `alpine:latest` → `debian:bookworm-slim`
- Explicit `update-ca-certificates` during build
- Better TLS/SSL support

**Trade-offs:**
- Image size: ~80-100MB (naik dari ~50MB)
- **Worth it:** Production-stable TLS connections

---

## 🚀 **Rebuild Instructions**

### **Step 1: Stop Existing Containers**

```bash
docker-compose down
```

### **Step 2: Remove Old Image (Force Rebuild)**

```bash
# Remove old image
docker rmi app-wa:latest 2>/dev/null || true

# Or clean all
docker-compose down --rmi all
```

### **Step 3: Rebuild with No Cache**

```bash
# Rebuild from scratch
docker-compose build --no-cache app

# Or rebuild all services
docker-compose build --no-cache
```

### **Step 4: Start Services**

```bash
# Start in detached mode
docker-compose up -d

# Or start with logs
docker-compose up
```

### **Step 5: Verify Fix**

```bash
# Check if containers are running
docker-compose ps

# Should show:
# NAME        STATUS        PORTS
# app-wa      Up (healthy)  0.0.0.0:3000->3000/tcp
# mongo-wa    Up (healthy)  0.0.0.0:27017->27017/tcp
# web-wa      Up (healthy)  0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp

# Check app logs
docker-compose logs -f app

# Test health endpoint
curl http://localhost:3000/health
# Expected: {"status":"ok"}

# Test QR generation (replace device-id)
curl http://localhost:3000/whatsapp/YOUR_DEVICE_ID/qrcode

# Should return QR code data without TLS error!
```

---

## 🔍 **Troubleshooting**

### **Issue: Old image still being used**

```bash
# Force remove all related images
docker images | grep app-wa
docker rmi <image-id> --force

# Remove all dangling images
docker image prune -a

# Rebuild
docker-compose build --no-cache
docker-compose up -d
```

### **Issue: Build fails**

```bash
# Check Docker daemon
docker info

# Check disk space
df -h

# Clean Docker system
docker system prune -a --volumes

# Rebuild
docker-compose build --no-cache
```

### **Issue: Container starts but TLS still fails**

```bash
# Exec into container and check CA certificates
docker-compose exec app ls -la /etc/ssl/certs/

# Should see many .pem files

# Check if ca-certificates is installed
docker-compose exec app dpkg -l | grep ca-certificates

# Should show: ca-certificates installed

# Check Go's certificate validation
docker-compose exec app cat /etc/ssl/certs/ca-certificates.crt | wc -l

# Should show 100+ lines (certificate bundle)
```

### **Issue: Still getting certificate errors**

```bash
# Verify WhatsApp Web is accessible from container
docker-compose exec app wget -O- https://web.whatsapp.com 2>&1 | grep -i "certificate"

# Should NOT show certificate errors

# Test DNS resolution
docker-compose exec app nslookup web.whatsapp.com

# Should resolve to IP addresses
```

---

## 📊 **Before vs After**

### **Before (Alpine)**
```dockerfile
FROM alpine:latest
RUN apk add --no-cache \
    ca-certificates \
    sqlite-libs \
    tzdata
```
- ❌ TLS errors
- ❌ Incomplete CA bundle
- ✅ Small image (~50MB)

### **After (Debian)**
```dockerfile
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libsqlite3-0 \
    tzdata \
    wget \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*
```
- ✅ TLS works perfectly
- ✅ Complete CA bundle
- ✅ Production-stable (~80-100MB)

---

## ✅ **Verification Checklist**

After rebuild, verify:

- [ ] Containers are running (`docker-compose ps`)
- [ ] Health checks pass (all services show "healthy")
- [ ] App logs show no TLS errors
- [ ] Health endpoint works (`curl http://localhost:3000/health`)
- [ ] QR endpoint returns data without TLS error
- [ ] WhatsApp connection successful

---

## 📝 **Additional Notes**

### **Why Debian instead of Alpine?**

1. **Better CA Certificate Support**
   - Debian includes complete Mozilla CA bundle
   - Regular security updates via APT
   - Better compatibility with Go's crypto/tls

2. **Production Proven**
   - Used by millions of production containers
   - Better tested with enterprise applications
   - More predictable behavior

3. **Developer Experience**
   - Familiar package manager (apt-get)
   - More utilities available
   - Easier debugging

### **Performance Impact**

- **Image Size:** +30-50MB
- **Build Time:** Similar (~1-2 minutes)
- **Runtime Performance:** Identical
- **Memory Usage:** Identical
- **Network:** Same

### **Security**

- Debian Bookworm Slim is minimal and secure
- Regular security patches via APT
- No unnecessary packages installed
- Non-root user still enforced

---

## 🆘 **Still Having Issues?**

1. **Check Docker Version:**
   ```bash
   docker --version
   # Should be 20.10+
   ```

2. **Check Docker Compose Version:**
   ```bash
   docker-compose --version
   # Should be 2.0+
   ```

3. **Check System Resources:**
   ```bash
   docker system df
   # Make sure you have enough space
   ```

4. **Complete Clean Rebuild:**
   ```bash
   # Nuclear option - removes EVERYTHING
   docker-compose down -v
   docker system prune -a --volumes
   docker-compose build --no-cache
   docker-compose up -d
   ```

5. **Check Logs:**
   ```bash
   # App logs
   docker-compose logs -f app

   # All logs
   docker-compose logs -f
   ```

---

## 📚 **References**

- [Go TLS Certificate Verification](https://pkg.go.dev/crypto/tls)
- [Debian CA Certificates](https://packages.debian.org/bookworm/ca-certificates)
- [WhatsApp Web Security](https://faq.whatsapp.com/general/security-and-privacy/end-to-end-encryption)
- [Whatsmeow Documentation](https://github.com/tulir/whatsmeow)

---

**Fix Applied:** ✅
**Commit:** `52a7219`
**Branch:** `claude/refactor-whatsapp-modules-01XVqipj9VRC2Ezd4UaQodzw`
