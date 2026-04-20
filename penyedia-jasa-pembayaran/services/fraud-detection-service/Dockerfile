FROM python:3.12-slim

WORKDIR /app

# Install system deps for asyncpg
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libpq-dev && \
    rm -rf /var/lib/apt/lists/*

# Install Python deps
COPY services/fraud-detection-service/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy app code
COPY services/fraud-detection-service/ .

# Copy proto for reference (not used at runtime for Go stubs)
COPY proto/fraud/v1/fraud.proto ./proto/fraud/v1/

COPY services/fraud-detection-service/scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 8085

ENTRYPOINT ["entrypoint.sh"]
