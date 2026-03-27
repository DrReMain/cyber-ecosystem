#!/bin/sh
# MinIO startup script with bucket initialization

# Start MinIO server in background
echo "Starting MinIO server..."
minio server /data --console-address ":9001" "$@" &
MINIO_PID=$!

# Wait for MinIO to be ready
echo "Waiting for MinIO to be ready..."
sleep 5
until mc alias set myminio http://localhost:9000 $MINIO_ROOT_USER $MINIO_ROOT_PASSWORD 2>/dev/null; do
    sleep 1
done

# Create bucket if it doesn't exist
echo "Creating bucket: uploads"
mc mb myminio/uploads --ignore-existing 2>/dev/null || true

# Set bucket public access
echo "Setting bucket public access..."
mc anonymous set download myminio/uploads 2>/dev/null || true

echo "Bucket initialization complete."

# Wait for MinIO server to exit
wait $MINIO_PID
