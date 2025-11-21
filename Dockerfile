# --- Base image ---
FROM golang:1.24-alpine AS builder

# Cài các tool cơ bản
RUN apk add --no-cache bash git python3 py3-pip build-base

# Tạo thư mục app
RUN mkdir /app
WORKDIR /app

# Copy toàn bộ project
ADD . /app

# Build Go Gin server
RUN go build -o main cmd/main.go

# --- Cài Python dependencies trong virtualenv ---
RUN python3 -m venv /opt/venv \
    && /opt/venv/bin/pip install --upgrade pip \
    && /opt/venv/bin/pip install fastapi uvicorn scikit-learn joblib

# --- Final image ---
FROM alpine:3.18

# Copy toàn bộ app + Go binary + Python venv
COPY --from=builder /app /app
COPY --from=builder /opt/venv /opt/venv

WORKDIR /app

# Expose port Gin server (public)
EXPOSE 8080

# Copy entrypoint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Start cả 2 service
CMD ["/entrypoint.sh"]
