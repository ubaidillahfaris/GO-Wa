# Quick Start Guide - Production Deployment

Panduan cepat untuk deploy aplikasi WhatsApp API ke production.

## TL;DR

```bash
# 1. Setup environment
make install

# 2. Edit environment file
nano .env.production

# 3. Deploy!
make deploy

# 4. Check status
make status
make health
```

## Prerequisites Checklist

- [ ] Server dengan Docker installed
- [ ] Domain sudah pointing ke server (optional, bisa pakai IP)
- [ ] SSL certificate (atau pakai self-signed untuk testing)

## Step-by-Step

### 1. Clone Repository

```bash
git clone <your-repo-url>
cd whatsapp
```

### 2. Setup Environment

```bash
# Automatic setup (creates .env.production and SSL)
make install

# Or manual
cp .env.production.example .env.production
```

### 3. Configure Environment

Edit `.env.production` dan update:

```bash
nano .env.production
```

**WAJIB diubah:**
```bash
MONGO_USER=admin_whatsapp
MONGO_PASS=YOUR_STRONG_PASSWORD_HERE      # CHANGE THIS!
JWT_SECRET=YOUR_LONG_RANDOM_SECRET_HERE   # Generate: openssl rand -base64 64
CORS_ALLOWED_ORIGIN=https://yourdomain.com  # Your domain
```

### 4. Setup SSL Certificate

#### Option A: Self-Signed (Development)
```bash
make setup-ssl
```

#### Option B: Let's Encrypt (Production)
```bash
sudo apt install certbot -y
sudo certbot certonly --standalone -d yourdomain.com
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/key.pem
sudo chmod 644 nginx/ssl/*.pem
```

### 5. Deploy

```bash
# Full deployment (build + start)
make deploy

# Or step by step
make build        # Build images
make up           # Start services
```

### 6. Verify

```bash
# Check container status
make status

# Check health
make health

# View logs
make logs

# Test endpoints
curl http://localhost/health
curl https://localhost/api/health
```

## Access Your Application

- **Frontend**: `https://yourdomain.com` or `https://YOUR_SERVER_IP`
- **Backend API**: `https://yourdomain.com/api`
- **Health Check**: `https://yourdomain.com/health`

## Common Commands

```bash
# View all available commands
make help

# View logs
make logs              # All services
make logs-backend      # Backend only
make logs-frontend     # Frontend only
make logs-nginx        # Nginx only

# Restart services
make restart           # All services
make update           # Update with zero downtime

# Backup data
make backup

# Stop services
make stop             # Stop but keep containers
make down             # Stop and remove containers

# Clean up
make clean            # Remove all containers and images
```

## Troubleshooting

### Services not starting?

```bash
# Check logs
make logs

# Check specific service
docker-compose -f docker-compose.prod.yml logs backend
```

### Port 80/443 already in use?

```bash
# Check what's using the port
sudo lsof -i :80
sudo lsof -i :443

# Stop conflicting service (example: apache)
sudo systemctl stop apache2
```

### MongoDB connection failed?

```bash
# Check MongoDB logs
make logs-mongo

# Verify credentials in .env.production
# Make sure MONGO_USER and MONGO_PASS match
```

### Can't access from browser?

```bash
# Check firewall
sudo ufw status
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Check if nginx is running
make status
```

### SSL certificate error?

```bash
# Check certificate
openssl x509 -in nginx/ssl/cert.pem -text -noout

# For self-signed cert, browser will show warning (click "Advanced" > "Proceed")
```

## Updating Application

### Method 1: Full Redeploy
```bash
git pull origin main
make deploy
```

### Method 2: Zero-Downtime Update
```bash
git pull origin main
make update
```

## Backup & Restore

### Backup
```bash
# Backup MongoDB and volumes
make backup

# Files saved to: backups/mongodb_backup_YYYYMMDD_HHMMSS.archive
```

### Restore
```bash
# Restore from backup
make restore FILE=backups/mongodb_backup_20241124_103000.archive
```

## Monitoring

### Resource Usage
```bash
make stats
```

### Service Health
```bash
make health
```

### View Logs
```bash
make logs
```

## Security Checklist

Before going to production:

- [ ] Changed `MONGO_PASS` to strong password
- [ ] Generated secure `JWT_SECRET` (min 64 chars)
- [ ] Using valid SSL certificates (not self-signed)
- [ ] Set correct `CORS_ALLOWED_ORIGIN`
- [ ] Firewall configured (only 80, 443, 22 open)
- [ ] SSH key-based authentication enabled
- [ ] Regular backups scheduled
- [ ] Monitoring setup (optional but recommended)

## Production Best Practices

1. **Use Strong Passwords**
   ```bash
   # Generate secure password
   openssl rand -base64 32
   ```

2. **Enable Firewall**
   ```bash
   sudo ufw enable
   sudo ufw allow 22/tcp   # SSH
   sudo ufw allow 80/tcp   # HTTP
   sudo ufw allow 443/tcp  # HTTPS
   ```

3. **Setup Auto-Backup**
   ```bash
   # Add to crontab (daily at 3 AM)
   crontab -e
   0 3 * * * cd /path/to/whatsapp && make backup
   ```

4. **Monitor Logs**
   ```bash
   # Setup log rotation
   sudo apt install logrotate
   ```

5. **Keep Docker Updated**
   ```bash
   sudo apt update && sudo apt upgrade docker-ce
   ```

## Getting Help

### View detailed documentation:
- [Full Deployment Guide](DOCKER_DEPLOYMENT.md)
- [Architecture Overview](ARCHITECTURE_PRODUCTION.md)
- [Backend Docs](be/DEPLOYMENT.md)

### Check logs for errors:
```bash
make logs
```

### Test each service:
```bash
# Test nginx
curl http://localhost/health

# Test backend
docker-compose -f docker-compose.prod.yml exec backend curl http://localhost:3000/health

# Test MongoDB
docker-compose -f docker-compose.prod.yml exec mongo mongosh --eval "db.adminCommand('ping')"
```

## Port Reference

### Exposed to Public
- Port 80: HTTP (redirects to HTTPS)
- Port 443: HTTPS (main access point)

### Internal Only (not accessible from internet)
- Frontend: 8080
- Backend: 3000
- MongoDB: 27017

All traffic goes through Nginx on port 80/443!

## Next Steps

After successful deployment:

1. Test API endpoints dengan Postman/curl
2. Setup monitoring (Portainer, Grafana, etc)
3. Configure domain DNS
4. Setup SSL auto-renewal
5. Configure backup automation
6. Review logs regularly

## Support

Issues? Check:
1. Logs: `make logs`
2. Status: `make status`
3. Health: `make health`
4. Documentation: [DOCKER_DEPLOYMENT.md](DOCKER_DEPLOYMENT.md)

---

**Happy Deploying!** 🚀

---

**Last Updated**: 2025-11-24
