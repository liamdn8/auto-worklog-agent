#!/bin/bash

echo "=== ActivityWatch Docker Test ==="
echo

# Test if the container is running
echo "1. Checking container status..."
docker compose ps

echo
echo "2. Testing API endpoint..."
response=$(curl -s http://localhost:5600/api/0/info)
if [ $? -eq 0 ]; then
    echo "✅ API is responding: $response"
else
    echo "❌ API test failed"
    exit 1
fi

echo
echo "3. Testing buckets endpoint..."
buckets=$(curl -s http://localhost:5600/api/0/buckets/)
echo "✅ Buckets endpoint working: $buckets"

echo
echo "4. Showing recent logs..."
docker compose logs --tail=5 aw-server

echo
echo "🎉 ActivityWatch server is running successfully!"
echo "📝 Access the API at: http://localhost:5600/api/0/"
echo "🔍 View server info at: http://localhost:5600/api/0/info"