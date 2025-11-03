# ActivityWatch Docker Setup - README

## What we created:

1. **Dockerfile** - Custom Docker image that uses your downloaded ActivityWatch server
2. **docker-compose.yaml** - Updated to build and run our custom image
3. **test-docker.sh** - Comprehensive test script to verify the setup
4. **test-aw-server.sh** - Script to run ActivityWatch server directly (without Docker)

## How to use:

### Start the ActivityWatch server with Docker:
```bash
docker compose up -d
```

### Test the setup:
```bash
./test-docker.sh
```

### Stop the server:
```bash
docker compose down
```

### View logs:
```bash
docker compose logs aw-server
```

### Inspect captured work sessions:
ActivityWatch exposes the raw event stream via its HTTP API. To inspect the
events produced by `awagent`, query the bucket directly:

```bash
curl -s "http://localhost:5600/api/0/buckets/" | jq
curl -s "http://localhost:5600/api/0/buckets/<bucket-id>/events?limit=5" | jq
```

Replace `<bucket-id>` with the bucket name emitted by the agent (for example
`awagent.auto-worklog-agent`). The default server build served from port 5600
does not ship the full ActivityWatch web UI; the API calls above provide a
lightweight way to validate that work sessions are being recorded.

## API Endpoints:

- Server info: http://localhost:5600/api/0/info
- Buckets: http://localhost:5600/api/0/buckets/
- Full API docs: http://localhost:5600/api/

## File Structure:

- **ActivityWatch server**: `activitywatch/aw-server/aw-server` (executable)
- **Data persistence**: Docker volume `auto-worklog-agent_aw_data`
- **Port**: 5600 (exposed to host)

## Notes:

- The server runs as non-root user `awuser` for security
- Data is persisted in a Docker volume
- Health checks are configured
- The server binds to all interfaces (0.0.0.0) for Docker networking

## Current Status: ✅ WORKING

The ActivityWatch server is successfully running in Docker and responding to API requests.