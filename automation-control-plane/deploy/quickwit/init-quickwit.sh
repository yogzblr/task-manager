#!/bin/bash
# Quickwit initialization script
# Creates the automation-logs index and S3 bucket

set -e

echo "Waiting for MinIO to be ready..."
sleep 5

echo "Creating S3 bucket in MinIO..."
mc alias set myminio http://minio:9000 minioadmin minioadmin || true
mc mb myminio/quickwit-indexes --ignore-existing || true
mc mb myminio/quickwit-indexes/indexes --ignore-existing || true
mc mb myminio/quickwit-indexes/metastore --ignore-existing || true

echo "S3 bucket created successfully"

echo "Creating automation-logs index..."
quickwit index create --index-config /quickwit/config/automation-logs-index.yaml || echo "Index may already exist"

echo "Quickwit initialization complete!"
