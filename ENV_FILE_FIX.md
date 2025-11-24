# Fix: Environment Variables Auto-Loading

## 🐛 Problem

Ketika menjalankan docker compose commands, environment variables tidak di-load:

```bash
make down
# Output:
WARN[0000] The "CORS_ALLOWED_ORIGIN" variable is not set. Defaulting to a blank string.
WARN[0000] The "MONGO_PASS" variable is not set. Defaulting to a blank string.
WARN[0000] The "JWT_SECRET" variable is not set. Defaulting to a blank string.
```

**Root Cause:**
- Docker Compose hanya membaca file `.env` (default)
- Kita pakai `.env.production` atau `.env.localhost`
- Tidak ada flag `--env-file` di Makefile commands

## ✅ Solution

**Updated Makefile** untuk auto-detect dan load environment file.

### Changes Made

1. **Auto-detect ENV_FILE based on COMPOSE_FILE:**
   ```makefile
   # Auto-detect environment file based on compose file
   ifeq ($(COMPOSE_FILE),docker-compose.localhost.yml)
       ENV_FILE ?= .env.localhost
   else ifeq ($(COMPOSE_FILE),docker-compose.nginx-docker.yml)
       ENV_FILE ?= .env.production
   else ifeq ($(COMPOSE_FILE),docker-compose.existing-nginx.yml)
       ENV_FILE ?= .env.production
   else
       ENV_FILE ?= .env.production
   endif
   ```

2. **Created DOCKER_COMPOSE variable with --env-file:**
   ```makefile
   # Docker compose command with env file
   DOCKER_COMPOSE = docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE)
   ```

3. **Replaced all `docker-compose` commands:**
   ```makefile
   # OLD:
   docker-compose -f $(COMPOSE_FILE) up -d

   # NEW:
   $(DOCKER_COMPOSE) up -d
   ```

### How It Works

**Scenario 1: Default (docker-compose.prod.yml)**
```bash
make up
# Uses: docker compose --env-file .env.production -f docker-compose.prod.yml up -d
```

**Scenario 2: Localhost**
```bash
make up COMPOSE_FILE=docker-compose.localhost.yml
# Uses: docker compose --env-file .env.localhost -f docker-compose.localhost.yml up -d
```

**Scenario 3: Existing Nginx (Host)**
```bash
make up COMPOSE_FILE=docker-compose.existing-nginx.yml
# Uses: docker compose --env-file .env.production -f docker-compose.existing-nginx.yml up -d
```

**Scenario 4: Existing Nginx (Docker)**
```bash
make up COMPOSE_FILE=docker-compose.nginx-docker.yml
# Uses: docker compose --env-file .env.production -f docker-compose.nginx-docker.yml up -d
```

## 📋 Testing

### Before Fix
```bash
make down COMPOSE_FILE=docker-compose.nginx-docker.yml
# Multiple WARN messages about unset variables
```

### After Fix
```bash
make down COMPOSE_FILE=docker-compose.nginx-docker.yml
# No warnings! Variables loaded from .env.production
```

### Verify Environment Loading
```bash
# Check what env file is used
make status COMPOSE_FILE=docker-compose.nginx-docker.yml
# Should see containers without warnings

# Check variables in container
docker compose --env-file .env.production -f docker-compose.nginx-docker.yml exec backend env | grep JWT
# Should show JWT_SECRET value
```

## 🎯 Usage

### Standard Commands (No Changes!)

Semua Makefile commands tetap sama, tapi sekarang **otomatis load env file**:

```bash
# Default (production)
make up
make down
make logs
make restart

# Localhost
make up COMPOSE_FILE=docker-compose.localhost.yml
make logs COMPOSE_FILE=docker-compose.localhost.yml

# Existing nginx (host)
make up COMPOSE_FILE=docker-compose.existing-nginx.yml

# Existing nginx (docker) - Your case!
make up COMPOSE_FILE=docker-compose.nginx-docker.yml
make down COMPOSE_FILE=docker-compose.nginx-docker.yml
make logs COMPOSE_FILE=docker-compose.nginx-docker.yml
```

### Override ENV_FILE (If Needed)

```bash
# Use custom env file
make up ENV_FILE=.env.custom

# Use different compose + env
make up COMPOSE_FILE=docker-compose.nginx-docker.yml ENV_FILE=.env.staging
```

## 💡 Benefits

1. **No More Warnings** ✅
   - Environment variables always loaded
   - No "variable is not set" messages

2. **Automatic Detection** ✅
   - Right env file for each compose file
   - No manual `--env-file` flag needed

3. **Backward Compatible** ✅
   - All existing Makefile commands work
   - Can still override ENV_FILE if needed

4. **Cleaner Commands** ✅
   - No need to type `--env-file` every time
   - Shorter command lines

## 🔍 Technical Details

### ENV_FILE Selection Logic

```
docker-compose.localhost.yml     → .env.localhost
docker-compose.prod.yml          → .env.production
docker-compose.existing-nginx.yml → .env.production
docker-compose.nginx-docker.yml  → .env.production
<any other>                      → .env.production
```

### Variable Precedence

Docker Compose loads variables in this order (highest priority first):
1. Shell environment variables
2. Variables from `--env-file` flag
3. Variables from `.env` file (default)
4. Default values in docker-compose.yml

Our fix adds `--env-file` flag, so variables are loaded with high priority.

## 📝 Notes

### For Existing Users

**No action needed!** All Makefile commands continue to work.

If you have custom scripts using docker-compose directly:
```bash
# OLD (manual flag needed):
docker compose --env-file .env.production -f docker-compose.nginx-docker.yml up -d

# NEW (Makefile handles it):
make up COMPOSE_FILE=docker-compose.nginx-docker.yml
```

### Environment File Requirements

Make sure your env file exists:
```bash
# Production
ls -la .env.production

# Localhost
ls -la .env.localhost

# If missing, copy from example:
cp .env.production.example .env.production
cp .env.localhost.example .env.localhost
```

### Verification

Test that env variables are loaded:
```bash
# Start services
make up COMPOSE_FILE=docker-compose.nginx-docker.yml

# Check no warnings during down
make down COMPOSE_FILE=docker-compose.nginx-docker.yml

# Should complete without WARN messages ✅
```

## 🚀 Deployment Impact

### Before (Manual):
```bash
# Every command needed --env-file
docker compose --env-file .env.production -f docker-compose.nginx-docker.yml up -d
docker compose --env-file .env.production -f docker-compose.nginx-docker.yml logs
docker compose --env-file .env.production -f docker-compose.nginx-docker.yml down
```

### After (Automatic):
```bash
# Makefile handles it automatically
make up COMPOSE_FILE=docker-compose.nginx-docker.yml
make logs COMPOSE_FILE=docker-compose.nginx-docker.yml
make down COMPOSE_FILE=docker-compose.nginx-docker.yml
```

## ✅ Status

**Fix Applied:** ✅ Complete
**Testing:** ✅ Verified
**Documentation:** ✅ Updated

All Makefile commands now automatically load the correct environment file! 🎉

---

**Date:** 2025-11-24
**Related Files:**
- [Makefile](Makefile) - Updated with auto env loading
- [SETUP_GUIDE.md](SETUP_GUIDE.md) - Usage guide

**Issue:** Environment variables not loaded
**Solution:** Auto-detect and load env file in Makefile
**Result:** No more WARN messages! ✅
