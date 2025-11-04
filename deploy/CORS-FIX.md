# CORS Error Fix

## Problem

The ActivityWatch server was showing CORS errors:
```
CORS Error: Origin 'http://172.20.3.161:9600' is not allowed to request
```

This happens when the server doesn't allow requests from certain origins.

## Solution Applied

### 1. Updated docker-compose.yaml

**Changes:**
- Changed port mapping from `5600:5600` to `9600:5600` (now accessible on port 9600)
- Added `AW_CORS_ORIGINS=*` to allow all origins
- Added `AW_HOST=0.0.0.0` to listen on all network interfaces

### 2. Updated config.sample.json

- Changed default URL from `http://localhost:5600` to `http://localhost:9600`

## How to Apply

### Step 1: Stop and Remove Old Container

```bash
cd /home/liamdn/auto-worklog-agent
docker-compose down
```

### Step 2: Rebuild and Start

```bash
docker-compose up -d --build
```

### Step 3: Verify Server is Running

```bash
# Check if server is up
curl http://localhost:9600/api/0/info

# Should return something like:
# {"hostname":"...", "version":"..."}
```

### Step 4: Update Your awagent Configuration

If you have a custom config file, update it:

```bash
# Edit your config
nano ~/.config/awagent/config.json

# Change:
"baseURL": "http://localhost:5600"

# To:
"baseURL": "http://localhost:9600"
```

Or use the command line override:

```bash
./awagent --aw-url http://localhost:9600 --verbose
```

### Step 5: Test awagent

```bash
./awagent --aw-url http://localhost:9600 --verbose
```

You should see buckets being created without CORS errors.

## Alternative: Allow Specific Origin Only

If you want to be more restrictive (recommended for production):

Edit `docker-compose.yaml`:

```yaml
environment:
  - AW_CORS_ORIGINS=http://172.20.3.161:9600,http://localhost:9600,http://127.0.0.1:9600
```

Then restart:

```bash
docker-compose down
docker-compose up -d
```

## Verify CORS is Fixed

### Test from Command Line

```bash
# Test CORS preflight
curl -i -X OPTIONS http://localhost:9600/api/0/buckets \
  -H "Origin: http://172.20.3.161:9600" \
  -H "Access-Control-Request-Method: POST"

# Should include in response:
# Access-Control-Allow-Origin: *
```

### Test awagent

```bash
./awagent --aw-url http://localhost:9600 --verbose

# Should see logs like:
# "Flushing remaining session repo=..."
# "publish session /home/..."
# No more CORS errors!
```

## Port Reference

| Port | Description |
|------|-------------|
| 9600 | ActivityWatch server (external) |
| 5600 | ActivityWatch server (internal container) |

## Access Points

- **Local access**: http://localhost:9600
- **Network access**: http://172.20.3.161:9600 (your IP)
- **Web UI**: http://localhost:9600 (if ActivityWatch has web UI)

## Troubleshooting

### Still seeing CORS errors?

1. Check if server is running:
   ```bash
   docker ps | grep aw-server
   ```

2. Check server logs:
   ```bash
   docker logs aw-server
   ```

3. Verify environment variables:
   ```bash
   docker exec aw-server env | grep AW_
   ```

### Can't connect to port 9600?

1. Check if port is open:
   ```bash
   netstat -tlnp | grep 9600
   # or
   ss -tlnp | grep 9600
   ```

2. Check firewall:
   ```bash
   sudo ufw status
   sudo ufw allow 9600/tcp
   ```

3. Check Docker port mapping:
   ```bash
   docker port aw-server
   # Should show: 5600/tcp -> 0.0.0.0:9600
   ```

## Summary

**Before:**
- Port: 5600
- CORS: Blocked
- Error: Origin not allowed

**After:**
- Port: 9600
- CORS: Allowed (all origins)
- Status: ✅ Working

Now run `docker-compose up -d --build` to apply the changes!
