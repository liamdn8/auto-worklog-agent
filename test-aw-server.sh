#!/bin/bash

# Test script to run ActivityWatch server locally
echo "Starting ActivityWatch server..."

# Set environment variables
export AW_DATA_DIR="./aw-data"
export PYTHONPATH="./activitywatch/aw-server"

# Create data directory if it doesn't exist
mkdir -p "$AW_DATA_DIR"

# Make the executable runnable
chmod +x ./activitywatch/aw-server/aw-server

# Run the server
cd ./activitywatch/aw-server
./aw-server --host 0.0.0.0 --port 5600