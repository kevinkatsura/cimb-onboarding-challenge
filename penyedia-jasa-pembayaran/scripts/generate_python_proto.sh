#!/bin/bash

# This script generates Python gRPC stubs using Docker.

# Get the absolute path of the project root
PROJECT_ROOT=$(pwd)

echo "Generating Python gRPC stubs..."

docker run --rm \
    -v "$PROJECT_ROOT":/workspace \
    -w /workspace \
    rvolosatovs/protoc:4.0.0 \
    -I. \
    --python_out=services/fraud-detection-service \
    --grpc-python_out=services/fraud-detection-service \
    proto/fraud/v1/fraud.proto

echo "Done."
