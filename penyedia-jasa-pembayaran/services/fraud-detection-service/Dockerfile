FROM python:3.12-slim

WORKDIR /app

# Install system deps for asyncpg
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libpq-dev && \
    rm -rf /var/lib/apt/lists/*

# Install Python deps
COPY services/fraud-detection-service/requirements.txt .
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

# Copy app code
COPY services/fraud-detection-service/ .

# Copy proto for generation
RUN mkdir -p app/proto/v1
COPY proto/fraud/v1/fraud.proto app/proto/v1/

# Generate Python gRPC stubs directly into the app structure
RUN python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. app/proto/v1/fraud.proto && \
    touch app/proto/__init__.py app/proto/v1/__init__.py

COPY services/fraud-detection-service/scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 8085

ENTRYPOINT ["entrypoint.sh"]
