#!/bin/bash

# WhatsApp API - Interactive Setup Wizard
# This script helps configure deployment for different scenarios

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

generate_secret() {
    openssl rand -base64 32 | tr -d "=+/" | cut -c1-32
}

# Main wizard
clear
print_header "WhatsApp API - Interactive Setup Wizard"

echo "This wizard will help you configure your WhatsApp API deployment."
echo ""

# Step 1: Choose deployment scenario
print_header "Step 1: Choose Deployment Scenario"
echo "1) Localhost (Development - HTTP only)"
echo "2) Fresh Server (Production - Docker with built-in Nginx)"
echo "3) Existing Nginx on Host (Production - System nginx)"
echo "4) Existing Nginx in Docker (Production - Docker nginx)"
echo ""
read -p "Choose your scenario [1-4]: " SCENARIO

case $SCENARIO in
    1)
        COMPOSE_FILE="docker-compose.localhost.yml"
        NGINX_CONFIG="nginx.local.conf"
        ENV_FILE=".env.localhost"
        USE_SSL=false
        DOMAIN="localhost"
        print_info "Selected: Localhost Development"
        ;;
    2)
        COMPOSE_FILE="docker-compose.prod.yml"
        NGINX_CONFIG="nginx.production.conf"
        ENV_FILE=".env.production"
        USE_SSL=true
        print_info "Selected: Fresh Server with Docker"
        ;;
    3)
        COMPOSE_FILE="docker-compose.existing-nginx.yml"
        NGINX_CONFIG="existing-nginx-site.conf"
        ENV_FILE=".env.production"
        USE_SSL=true
        print_info "Selected: Existing Nginx on Host"
        ;;
    4)
        COMPOSE_FILE="docker-compose.nginx-docker.yml"
        NGINX_CONFIG="nginx-docker.conf"
        ENV_FILE=".env.production"
        USE_SSL=true
        print_info "Selected: Existing Nginx in Docker"
        ;;
    *)
        print_error "Invalid choice. Exiting."
        exit 1
        ;;
esac

# Step 2: Domain configuration
if [ "$USE_SSL" = true ]; then
    print_header "Step 2: Domain Configuration"
    read -p "Enter your domain or subdomain (e.g., wa.yourdomain.com): " DOMAIN

    if [ -z "$DOMAIN" ]; then
        print_error "Domain cannot be empty. Exiting."
        exit 1
    fi

    print_success "Domain: $DOMAIN"
fi

# Step 3: MongoDB configuration
print_header "Step 3: MongoDB Configuration"
read -p "MongoDB username [admin]: " MONGO_USER
MONGO_USER=${MONGO_USER:-admin}

read -sp "MongoDB password (leave empty to generate): " MONGO_PASS
echo ""
if [ -z "$MONGO_PASS" ]; then
    MONGO_PASS=$(generate_secret)
    print_success "Generated MongoDB password: $MONGO_PASS"
else
    print_success "Using provided MongoDB password"
fi

read -p "MongoDB database name [whatsapp_production]: " MONGO_DB
MONGO_DB=${MONGO_DB:-whatsapp_production}

# Step 4: JWT configuration
print_header "Step 4: JWT Secret Configuration"
read -sp "JWT secret (leave empty to generate): " JWT_SECRET
echo ""
if [ -z "$JWT_SECRET" ]; then
    JWT_SECRET=$(generate_secret)
    print_success "Generated JWT secret: $JWT_SECRET"
else
    print_success "Using provided JWT secret"
fi

read -p "JWT expiration in minutes [60]: " JWT_EXPIRES
JWT_EXPIRES=${JWT_EXPIRES:-60}

# Step 5: CORS configuration
print_header "Step 5: CORS Configuration"
if [ "$USE_SSL" = true ]; then
    CORS_ORIGIN="https://${DOMAIN}"
else
    CORS_ORIGIN="http://localhost"
fi
print_info "CORS allowed origin will be set to: $CORS_ORIGIN"

# Step 6: Additional configuration for nginx in docker
if [ "$SCENARIO" = "4" ]; then
    print_header "Step 6: Docker Network Configuration"
    echo "Please enter the name of your existing nginx Docker network."
    echo "To find it, run: docker network ls"
    echo ""
    read -p "Nginx Docker network name: " NGINX_NETWORK

    if [ -z "$NGINX_NETWORK" ]; then
        print_error "Network name cannot be empty. Exiting."
        exit 1
    fi

    print_success "Nginx network: $NGINX_NETWORK"
fi

# Create environment file
print_header "Creating Environment File"
cat > "$ENV_FILE" <<EOF
# MongoDB Configuration
MONGO_USER=$MONGO_USER
MONGO_PASS=$MONGO_PASS
MONGO_DB=$MONGO_DB

# JWT Configuration
JWT_SECRET=$JWT_SECRET
JWT_EXPIRES_MIN=$JWT_EXPIRES

# CORS Configuration
CORS_ALLOWED_ORIGIN=$CORS_ORIGIN
CORS_MAX_AGE=43200

# WhatsApp Configuration
WHATSAPP_MAX_CONCURRENCY=10
EOF

print_success "Environment file created: $ENV_FILE"

# Update nginx config with domain
if [ "$USE_SSL" = true ] && [ "$SCENARIO" = "2" ]; then
    print_header "Updating Nginx Configuration"

    # Create backup
    cp "nginx/$NGINX_CONFIG" "nginx/${NGINX_CONFIG}.bak"

    # Replace domain placeholder
    sed -i.tmp "s/__DOMAIN__/$DOMAIN/g" "nginx/$NGINX_CONFIG"
    rm -f "nginx/${NGINX_CONFIG}.tmp"

    print_success "Nginx config updated with domain: $DOMAIN"
fi

# Update docker-compose for nginx docker scenario
if [ "$SCENARIO" = "4" ]; then
    print_header "Updating Docker Compose Configuration"

    # Create backup
    cp "$COMPOSE_FILE" "${COMPOSE_FILE}.bak"

    # Replace network name
    sed -i.tmp "s/YOUR_NGINX_NETWORK_NAME/$NGINX_NETWORK/g" "$COMPOSE_FILE"
    rm -f "${COMPOSE_FILE}.tmp"

    print_success "Docker Compose updated with network: $NGINX_NETWORK"
fi

# SSL setup
if [ "$USE_SSL" = true ]; then
    print_header "SSL Certificate Setup"
    echo "How do you want to setup SSL?"
    echo "1) Self-signed certificate (for testing)"
    echo "2) Let's Encrypt (for production)"
    echo "3) I'll provide my own certificates"
    echo "4) Skip for now"
    echo ""
    read -p "Choose SSL option [1-4]: " SSL_OPTION

    case $SSL_OPTION in
        1)
            print_info "Creating self-signed certificate..."
            mkdir -p nginx/ssl
            openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
                -keyout nginx/ssl/key.pem \
                -out nginx/ssl/cert.pem \
                -subj "/C=ID/ST=Jakarta/L=Jakarta/O=Development/CN=$DOMAIN" \
                2>/dev/null
            print_success "Self-signed certificate created in nginx/ssl/"
            ;;
        2)
            print_warning "Let's Encrypt setup requires:"
            print_info "  1. Domain DNS pointing to this server"
            print_info "  2. Port 80 accessible from internet"
            print_info "  3. Certbot installed (sudo apt install certbot)"
            echo ""
            read -p "Continue with Let's Encrypt? [y/N]: " CONTINUE_LE

            if [[ $CONTINUE_LE =~ ^[Yy]$ ]]; then
                if [ "$SCENARIO" = "2" ]; then
                    print_info "Stopping nginx container temporarily..."
                    docker compose -f "$COMPOSE_FILE" stop nginx 2>/dev/null || true
                fi

                print_info "Running certbot..."
                sudo certbot certonly --standalone -d "$DOMAIN"

                mkdir -p nginx/ssl
                sudo cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" nginx/ssl/cert.pem
                sudo cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" nginx/ssl/key.pem
                sudo chmod 644 nginx/ssl/*.pem

                print_success "Let's Encrypt certificate installed!"

                # Setup auto-renewal
                print_info "Setting up auto-renewal..."
                CRON_CMD="0 2 * * * certbot renew --quiet --deploy-hook \"docker compose -f $(pwd)/$COMPOSE_FILE restart nginx\" 2>&1 | logger -t certbot"
                (crontab -l 2>/dev/null | grep -v "certbot renew"; echo "$CRON_CMD") | crontab -
                print_success "Auto-renewal configured (daily at 2 AM)"
            else
                print_warning "Skipped Let's Encrypt setup. You can run it later with: make ssl-letsencrypt DOMAIN=$DOMAIN"
            fi
            ;;
        3)
            print_info "Place your certificates in nginx/ssl/ with these names:"
            print_info "  - cert.pem (certificate + chain)"
            print_info "  - key.pem (private key)"
            mkdir -p nginx/ssl
            ;;
        4)
            print_warning "SSL setup skipped. Configure certificates before production use."
            mkdir -p nginx/ssl
            ;;
    esac
fi

# Final summary
print_header "Setup Summary"
echo "Configuration completed successfully!"
echo ""
print_info "Deployment Scenario: $([ "$SCENARIO" = "1" ] && echo "Localhost" || echo "Production")"
print_info "Compose File: $COMPOSE_FILE"
print_info "Environment File: $ENV_FILE"
if [ "$USE_SSL" = true ]; then
    print_info "Domain: $DOMAIN"
fi
print_info "MongoDB User: $MONGO_USER"
print_info "MongoDB Database: $MONGO_DB"
echo ""

# Deployment instructions
print_header "Next Steps"
case $SCENARIO in
    1)
        echo "Start your application with:"
        echo "  ${GREEN}docker compose -f $COMPOSE_FILE up -d${NC}"
        echo ""
        echo "Access your application at:"
        echo "  ${GREEN}http://localhost${NC}"
        ;;
    2)
        echo "Deploy your application with:"
        echo "  ${GREEN}docker compose -f $COMPOSE_FILE build${NC}"
        echo "  ${GREEN}docker compose -f $COMPOSE_FILE up -d${NC}"
        echo ""
        if [ "$USE_SSL" = true ]; then
            echo "Access your application at:"
            echo "  ${GREEN}https://$DOMAIN${NC}"
        fi
        ;;
    3)
        echo "1. Deploy Docker containers (without nginx):"
        echo "  ${GREEN}docker compose -f $COMPOSE_FILE up -d${NC}"
        echo ""
        echo "2. Copy nginx config to your system nginx:"
        echo "  ${GREEN}sudo cp nginx/$NGINX_CONFIG /etc/nginx/sites-available/whatsapp${NC}"
        echo "  ${GREEN}sudo ln -s /etc/nginx/sites-available/whatsapp /etc/nginx/sites-enabled/${NC}"
        echo ""
        echo "3. Update domain in nginx config:"
        echo "  ${GREEN}sudo nano /etc/nginx/sites-available/whatsapp${NC}"
        echo "  Change 'server_name' to: ${GREEN}$DOMAIN${NC}"
        echo ""
        echo "4. Test and reload nginx:"
        echo "  ${GREEN}sudo nginx -t${NC}"
        echo "  ${GREEN}sudo systemctl reload nginx${NC}"
        ;;
    4)
        echo "1. Deploy Docker containers:"
        echo "  ${GREEN}docker compose --env-file $ENV_FILE -f $COMPOSE_FILE up -d${NC}"
        echo ""
        echo "2. Copy nginx config to your nginx container:"
        echo "  ${GREEN}docker cp nginx/$NGINX_CONFIG <nginx-container>:/etc/nginx/conf.d/whatsapp.conf${NC}"
        echo ""
        echo "3. Update domain in the config, then reload nginx:"
        echo "  ${GREEN}docker exec <nginx-container> nginx -t${NC}"
        echo "  ${GREEN}docker exec <nginx-container> nginx -s reload${NC}"
        ;;
esac

echo ""
print_header "Useful Commands"
echo "Check status:     ${GREEN}make status${NC}"
echo "View logs:        ${GREEN}make logs${NC}"
echo "Restart services: ${GREEN}make restart${NC}"
echo "Health check:     ${GREEN}make health${NC}"
echo ""
echo "For more information, see: ${BLUE}DEPLOYMENT_GUIDE.md${NC}"
echo ""

print_success "Setup wizard completed! 🚀"
