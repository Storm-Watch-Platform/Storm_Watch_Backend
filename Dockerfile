# FROM golang:1.24-alpine

# RUN mkdir /app

# ADD . /app

# WORKDIR /app

# RUN go build -o main cmd/main.go

# CMD ["/app/main"]

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

# Cài Python dependencies
RUN pip install --no-cache-dir fastapi uvicorn scikit-learn joblib

# --- Final image ---
FROM alpine:3.18

# Copy Go binary và app
COPY --from=builder /app /app
WORKDIR /app

# Expose port Gin (Railway PORT)
EXPOSE 8080

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Start both services
CMD ["/entrypoint.sh"]
