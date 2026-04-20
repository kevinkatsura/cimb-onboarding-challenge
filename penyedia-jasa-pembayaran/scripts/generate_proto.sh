#!/bin/bash

# This script generates Go gRPC stubs using Docker to avoid local protoc installation requirements.

# Get the absolute path of the project root
PROJECT_ROOT=$(pwd)

# Define the proto files to generate
PROTO_FILES=(
    "proto/account/v1/account.proto"
    "proto/ledger/v1/ledger.proto"
    "proto/account_information/v1/account_information.proto"
)

echo "Generating gRPC stubs..."

for FILE in "${PROTO_FILES[@]}"; do
    echo "Processing $FILE..."
    docker run --rm \
        -v "$PROJECT_ROOT":/workspace \
        -w /workspace \
        rvolosatovs/protoc:4.0.0 \
        -I. \
        --plugin=protoc-gen-go=/usr/bin/protoc-gen-go \
        --plugin=protoc-gen-go-grpc=/usr/bin/protoc-gen-go-grpc \
        --go_out=. --go_opt=module=proto \
        --go-grpc_out=. --go-grpc_opt=module=proto \
        "$FILE"
done

echo "Done."
