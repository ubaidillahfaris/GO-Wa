# Production Deployment Checklist

Gunakan checklist ini untuk memastikan deployment production berjalan lancar.

## Pre-Deployment

### Server Setup
- [ ] Server ready (minimum 4 vCPU, 8GB RAM)
- [ ] Docker installed (version 20.10+)
- [ ] Docker Compose installed (version 2.0+)
- [ ] Domain DNS pointing to server IP
- [ ] Firewall configured (ports 22, 80, 443 open)
- [ ] SSH key-based authentication enabled
- [ ] Non-root user with sudo access created

### Repository & Code
- [ ] Code cloned to server
- [ ] Latest stable branch checked out
- [ ] No sensitive data in git history
- [ ] .gitignore properly configured

### Environment Configuration
- [ ] `.env.production` created from example
- [ ] `MONGO_USER` set to unique username
- [ ] `MONGO_PASS` set to strong password (16+ chars)
- [ ] `JWT_SECRET` generated (64+ chars): `openssl rand -base64 64`
- [ ] `CORS_ALLOWED_ORIGIN` set to correct domain
- [ ] `WHATSAPP_MAX_CONCURRENCY` adjusted for server capacity
- [ ] All required environment variables set
- [ ] `.env.production` NOT committed to git

### SSL/TLS Certificates
- [ ] SSL certificates obtained (Let's Encrypt recommended)
- [ ] Certificates copied to `nginx/ssl/` directory
- [ ] Certificate files readable (chmod 644)
- [ ] Certificate expiry date noted
- [ ] Auto-renewal configured (for Let's Encrypt)

### Security
- [ ] Strong passwords used everywhere
- [ ] Secrets not in code or git
- [ ] Firewall enabled and configured
- [ ] SSH port changed from default (optional but recommended)
- [ ] Fail2ban installed (optional but recommended)
- [ ] Security updates enabled

## Deployment

### Build Phase
- [ ] Run `make build` successfully
- [ ] No build errors in logs
- [ ] Docker images created (`docker images` shows them)
- [ ] Image sizes reasonable (<500MB for backend, <50MB for frontend)

### Initial Deployment
- [ ] Run `make up` or `make deploy`
- [ ] All containers started: `make status`
- [ ] No restart loops: `docker ps` (check UPTIME)
- [ ] Health checks passing: `make health`
- [ ] Logs show no errors: `make logs`

### Service Verification

#### Nginx
- [ ] Nginx container running
- [ ] HTTP (80) redirects to HTTPS (443)
- [ ] HTTPS responds correctly
- [ ] SSL certificate valid (no browser warning)
- [ ] Health endpoint responds: `curl http://localhost/health`

#### Backend
- [ ] Backend container running
- [ ] Backend health check passes: `curl http://localhost/api/health`
- [ ] Can connect to MongoDB
- [ ] JWT authentication works
- [ ] API endpoints respond correctly
- [ ] File upload works (if applicable)

#### Frontend
- [ ] Frontend container running
- [ ] Static files served correctly
- [ ] SPA routing works (refresh on any route)
- [ ] API calls to backend work
- [ ] No console errors in browser
- [ ] Assets loaded correctly

#### MongoDB
- [ ] MongoDB container running
- [ ] MongoDB health check passes
- [ ] Can authenticate with credentials
- [ ] Database created automatically
- [ ] Data persists after restart

### Network & Connectivity
- [ ] All containers in same network
- [ ] Internal DNS resolution works
- [ ] Backend can reach MongoDB
- [ ] Nginx can reach backend and frontend
- [ ] External access works (from internet)

## Post-Deployment

### Testing

#### Functional Testing
- [ ] Homepage loads
- [ ] User can login/register
- [ ] Dashboard accessible
- [ ] Can add WhatsApp client
- [ ] QR code generation works
- [ ] Can send messages
- [ ] File uploads work
- [ ] All critical features working

#### Performance Testing
- [ ] Response times acceptable (<500ms for API)
- [ ] No memory leaks (check `docker stats`)
- [ ] CPU usage normal (<50% idle)
- [ ] Database queries fast
- [ ] Frontend loads quickly (<3s)

#### Security Testing
- [ ] HTTPS enforced (HTTP redirects)
- [ ] CORS configured correctly
- [ ] Rate limiting works
- [ ] Authentication required for protected routes
- [ ] No sensitive data exposed in responses
- [ ] Security headers present: `curl -I https://yourdomain.com`

### Monitoring Setup
- [ ] Log rotation configured
- [ ] Monitoring solution installed (Portainer/Grafana)
- [ ] Alerts configured for critical issues
- [ ] Uptime monitoring enabled
- [ ] Resource usage tracked
- [ ] Disk space monitored

### Backup Configuration
- [ ] Backup script tested: `make backup`
- [ ] Backup cron job configured
- [ ] Backup storage location set
- [ ] Backup retention policy defined
- [ ] Restore procedure tested: `make restore FILE=...`
- [ ] Off-site backup configured (optional)

### Documentation
- [ ] Deployment notes documented
- [ ] Credentials stored securely (password manager)
- [ ] Team notified of deployment
- [ ] Runbook created for common tasks
- [ ] Incident response plan defined

## Maintenance

### Daily Tasks
- [ ] Check service status: `make status`
- [ ] Check logs for errors: `make logs`
- [ ] Monitor resource usage: `make stats`

### Weekly Tasks
- [ ] Review logs for issues
- [ ] Check disk space usage
- [ ] Verify backups running
- [ ] Test restore procedure (once a month)

### Monthly Tasks
- [ ] Update Docker images: `docker pull`
- [ ] Apply system updates: `apt update && apt upgrade`
- [ ] Review security advisories
- [ ] Rotate logs if needed
- [ ] Review resource usage trends

### As Needed
- [ ] Scale resources if needed
- [ ] Optimize database queries
- [ ] Update dependencies
- [ ] Apply security patches

## Rollback Plan

If deployment fails:

1. **Stop new containers**
   ```bash
   make down
   ```

2. **Check logs for errors**
   ```bash
   make logs > error_logs.txt
   ```

3. **Restore from backup**
   ```bash
   make restore FILE=backups/latest_backup.archive
   ```

4. **Revert to previous version**
   ```bash
   git checkout previous-stable-tag
   make deploy
   ```

5. **Notify team**
   - Document the issue
   - Create incident report
   - Plan fix

## Troubleshooting Reference

### Common Issues

| Issue | Check | Fix |
|-------|-------|-----|
| Container won't start | `make logs-[service]` | Check env vars, ports |
| MongoDB connection failed | `make logs-mongo` | Verify credentials |
| SSL error | Check cert files | Regenerate or update certs |
| Port already in use | `sudo lsof -i :80` | Stop conflicting service |
| High memory usage | `docker stats` | Increase limits or scale |
| Slow response | `make logs-nginx` | Check rate limits, cache |

### Emergency Contacts

- **DevOps Lead**: [Name, Phone, Email]
- **Backend Team**: [Contact Info]
- **Frontend Team**: [Contact Info]
- **Infrastructure**: [Contact Info]
- **Security**: [Contact Info]

## Success Criteria

Deployment is successful when:

- [ ] All containers running and healthy
- [ ] All health checks passing
- [ ] Application accessible from internet
- [ ] All critical features working
- [ ] No errors in logs
- [ ] Performance acceptable
- [ ] Security measures active
- [ ] Backups configured
- [ ] Monitoring active
- [ ] Team notified and trained

## Sign-Off

### Deployment Information

- **Deployment Date**: _______________
- **Version/Tag**: _______________
- **Deployed By**: _______________
- **Server**: _______________
- **Domain**: _______________

### Verification

- **Tested By**: _______________
- **Approved By**: _______________
- **Date**: _______________

### Notes

```
[Add any deployment-specific notes here]
```

---

## Quick Commands Reference

```bash
# Status
make status          # Container status
make health          # Health checks
make stats           # Resource usage

# Logs
make logs            # All logs
make logs-backend    # Backend only
make logs-nginx      # Nginx only

# Control
make restart         # Restart all
make stop            # Stop all
make up              # Start all

# Maintenance
make backup          # Backup data
make clean           # Clean up

# Emergency
make down            # Stop everything
make logs > debug.txt  # Save logs
```

---

**Keep this checklist updated as your deployment process evolves.**

**Last Updated**: 2025-11-24
