# Use a minimal base image
FROM debian:bullseye-slim

# Install necessary runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create a non-root user for security
RUN useradd -m -u 1000 awuser

# Create directory for ActivityWatch data
RUN mkdir -p /data && chown awuser:awuser /data

# Copy the ActivityWatch server and its dependencies
COPY --chown=awuser:awuser activitywatch/aw-server /opt/aw-server

# Set the working directory
WORKDIR /opt/aw-server

# Switch to non-root user
USER awuser

# Expose the default ActivityWatch server port
EXPOSE 5600

# Set environment variables
ENV AW_DATA_DIR=/data
ENV PYTHONPATH=/opt/aw-server

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:5600/api/0/info || exit 1

# Run the ActivityWatch server
CMD ["./aw-server", "--host", "0.0.0.0", "--port", "5600"]