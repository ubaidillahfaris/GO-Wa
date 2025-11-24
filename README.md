# WhatsApp API - Production Ready

Full-stack WhatsApp API application dengan frontend Vue.js dan backend Go, siap deploy ke production dengan Docker.

## Features

- **Backend (Go)**
  - WhatsApp integration via whatsmeow
  - RESTful API
  - JWT & API Key authentication
  - MongoDB database
  - Rate limiting
  - File upload support

- **Frontend (Vue.js)**
  - Modern UI dengan Tailwind CSS
  - Client management
  - QR code scanning
  - Message sending
  - Real-time updates

- **Production Ready**
  - Docker multi-container setup
  - Nginx reverse proxy
  - SSL/TLS support
  - Health checks
  - Auto-restart
  - Resource limits
  - Backup/restore scripts

## Quick Start

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- 4GB RAM minimum (8GB recommended)

### Installation

```bash
# 1. Clone repository
git clone <your-repo-url>
cd whatsapp

# 2. Setup environment and SSL
make install

# 3. Edit configuration
nano .env.production

# 4. Deploy!
make deploy

# 5. Check status
make status
make health
```

That's it! Your application is now running at:
- **Frontend**: `https://localhost` or `https://your-domain.com`
- **Backend API**: `https://localhost/api`

## Project Structure

```
whatsapp/
├── be/                          # Backend (Go)
│   ├── handlers/               # API handlers
│   ├── internal/               # Internal packages
│   ├── middlewares/            # Auth, CORS, etc
│   ├── models/                 # Data models
│   ├── routes/                 # API routes
│   ├── services/               # Business logic
│   ├── Dockerfile              # Backend Docker image
│   ├── go.mod                  # Go dependencies
│   └── main.go                 # Entry point
│
├── fe/                          # Frontend (Vue.js)
│   ├── src/                    # Source code
│   │   ├── components/        # Vue components
│   │   ├── pages/             # Page components
│   │   ├── router/            # Vue Router
│   │   ├── stores/            # Pinia stores
│   │   └── services/          # API services
│   ├── Dockerfile              # Frontend Docker image
│   ├── nginx.conf              # Frontend nginx config
│   └── package.json            # NPM dependencies
│
├── nginx/                       # Nginx configs
│   ├── nginx.prod.conf         # Production config
│   └── ssl/                    # SSL certificates
│
├── docker-compose.prod.yml      # Production compose file
├── Makefile                     # Deployment commands
├── .env.production.example      # Environment template
│
└── docs/                        # Documentation
    ├── QUICKSTART.md           # Quick start guide
    ├── DOCKER_DEPLOYMENT.md    # Full deployment guide
    └── ARCHITECTURE_PRODUCTION.md  # Architecture docs
```

## Documentation

### Quick Start
- [QUICKSTART.md](QUICKSTART.md) - Get started in 5 minutes

### Deployment
- [DOCKER_DEPLOYMENT.md](DOCKER_DEPLOYMENT.md) - Complete deployment guide
- [ARCHITECTURE_PRODUCTION.md](ARCHITECTURE_PRODUCTION.md) - Architecture overview

### API Documentation
- [be/docs/API_DOCUMENTATION.md](be/docs/API_DOCUMENTATION.md) - API endpoints
- [be/docs/API_KEY_AUTHENTICATION.md](be/docs/API_KEY_AUTHENTICATION.md) - API key auth

### Backend
- [be/DEPLOYMENT.md](be/DEPLOYMENT.md) - Backend deployment
- [be/ARCHITECTURE.md](be/ARCHITECTURE.md) - Backend architecture

## Architecture

### Production Architecture

```
Internet
   │
   ▼
Nginx (Port 80/443) ─── SSL/TLS Termination
   │                    Rate Limiting
   │                    Compression
   ├─────────────┬────────────────┐
   │             │                │
Frontend     Backend          MongoDB
(Port 8080)  (Port 3000)   (Port 27017)
Internal     Internal        Internal
   │             │                │
   └─────────────┴────────────────┘
         Docker Network
```

### Key Features

- **Single Entry Point**: All traffic goes through Nginx (port 80/443)
- **Security**: Internal ports not exposed, non-root users, rate limiting
- **Scalability**: Easy horizontal scaling with Docker Swarm/Kubernetes
- **Reliability**: Health checks, auto-restart, volume persistence
- **Performance**: Multi-stage builds, UPX compression, nginx caching

## Configuration

### Environment Variables

Key variables to configure in `.env.production`:

```bash
# MongoDB
MONGO_USER=admin_whatsapp
MONGO_PASS=your_strong_password_here

# JWT
JWT_SECRET=your_long_random_secret_here

# CORS
CORS_ALLOWED_ORIGIN=https://yourdomain.com
```

See [.env.production.example](.env.production.example) for all options.

### SSL Configuration

#### Development (Self-Signed)
```bash
make setup-ssl
```

#### Production (Let's Encrypt)
```bash
sudo certbot certonly --standalone -d yourdomain.com
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/key.pem
```

## Deployment

### Development
```bash
# Backend
cd be
go run main.go

# Frontend
cd fe
npm install
npm run dev
```

### Production

```bash
# Full deployment
make deploy

# Or step by step
make build       # Build images
make up          # Start services
make status      # Check status
make logs        # View logs
```

### Available Commands

```bash
make help              # Show all commands
make build             # Build Docker images
make up                # Start services
make down              # Stop services
make restart           # Restart services
make logs              # View logs
make logs-backend      # Backend logs only
make logs-frontend     # Frontend logs only
make logs-nginx        # Nginx logs only
make health            # Check service health
make backup            # Backup data
make restore FILE=...  # Restore data
make clean             # Clean up Docker
make deploy            # Full deployment
make update            # Zero-downtime update
```

## Monitoring

### Container Status
```bash
make status
docker-compose -f docker-compose.prod.yml ps
```

### Resource Usage
```bash
make stats
docker stats
```

### Logs
```bash
make logs              # All services
make logs-backend      # Backend only
make logs-nginx        # Nginx only
```

### Health Checks
```bash
make health
curl http://localhost/health
```

## Backup & Restore

### Backup
```bash
# Manual backup
make backup

# Automated (cron)
0 3 * * * cd /path/to/whatsapp && make backup
```

Backups stored in `backups/` directory.

### Restore
```bash
make restore FILE=backups/mongodb_backup_20241124_103000.archive
```

## Troubleshooting

### Common Issues

#### Port 80/443 already in use
```bash
sudo lsof -i :80
sudo systemctl stop apache2  # or nginx
```

#### MongoDB connection failed
```bash
make logs-mongo
# Check credentials in .env.production
make restart
```

#### Frontend shows 404 on refresh
Check [fe/nginx.conf](fe/nginx.conf:58) has proper SPA routing:
```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

#### SSL certificate error
```bash
# Check certificate
openssl x509 -in nginx/ssl/cert.pem -text -noout

# Recreate self-signed cert
make setup-ssl
```

### Debug Commands

```bash
# Shell into container
docker-compose -f docker-compose.prod.yml exec backend sh

# Check logs
docker-compose -f docker-compose.prod.yml logs -f

# Inspect container
docker inspect whatsapp-backend-prod

# Network debug
docker network inspect whatsapp_whatsapp-network
```

## Security

### Production Checklist

- [ ] Use strong MongoDB password
- [ ] Generate secure JWT_SECRET (64+ chars)
- [ ] Use valid SSL certificates (not self-signed)
- [ ] Set correct CORS_ALLOWED_ORIGIN
- [ ] Enable firewall (UFW/iptables)
- [ ] Keep .env.production out of git
- [ ] Regular security updates
- [ ] Monitor logs for suspicious activity
- [ ] Setup automated backups
- [ ] Use non-root users (already configured)

### Firewall Setup

```bash
sudo ufw enable
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
```

## Performance

### Optimization Features

- **Multi-stage Docker builds** - Smaller images
- **UPX compression** - 30-50% binary size reduction
- **Nginx caching** - Static asset caching
- **Gzip compression** - Response compression
- **Resource limits** - Prevent resource exhaustion
- **Health checks** - Auto-restart unhealthy containers

### Resource Allocation

| Service   | CPU Limit | Memory Limit | Minimum Requirements |
|-----------|-----------|--------------|----------------------|
| MongoDB   | 2 cores   | 2 GB         | 1 core, 1 GB         |
| Backend   | 2 cores   | 2 GB         | 0.5 core, 512 MB     |
| Frontend  | 0.5 core  | 512 MB       | 0.25 core, 256 MB    |
| Nginx     | 1 core    | 512 MB       | 0.25 core, 256 MB    |

**Server Requirements**: 4 vCPU, 8 GB RAM (minimum)

## Scaling

### Horizontal Scaling

#### Scale Backend
```yaml
backend:
  deploy:
    replicas: 3
```

#### Multi-Server (Docker Swarm)
```bash
docker swarm init
docker stack deploy -c docker-compose.prod.yml whatsapp
```

#### Kubernetes
See [k8s/](k8s/) directory for Kubernetes manifests (if available).

## API Usage

### Authentication

#### JWT Authentication
```bash
# Login
curl -X POST https://yourdomain.com/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# Use token
curl https://yourdomain.com/api/clients \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### API Key Authentication
```bash
curl https://yourdomain.com/api/clients \
  -H "X-API-Key: YOUR_API_KEY"
```

See [API Documentation](be/docs/API_DOCUMENTATION.md) for all endpoints.

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## License

[Your License Here]

## Support

- **Documentation**: See [docs/](docs/) directory
- **Issues**: [GitHub Issues](https://github.com/yourusername/whatsapp/issues)
- **Email**: your-email@example.com

## Credits

- [whatsmeow](https://github.com/tulir/whatsmeow) - WhatsApp library
- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [Vue.js](https://vuejs.org/) - Frontend framework
- [MongoDB](https://www.mongodb.com/) - Database

---

**Built with ❤️ using Go, Vue.js, and Docker**

**Last Updated**: 2025-11-24
