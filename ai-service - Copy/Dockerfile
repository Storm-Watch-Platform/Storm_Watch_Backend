FROM python:3.11-slim

# 1. Chuẩn bị thư mục
WORKDIR /app

# 2. Cài thư viện
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 3. Copy code + model
COPY . .

# 4. Biến môi trường (nếu cần)
ENV MODEL_DIR=/app/models
ENV CONFIDENCE_MIN=0.65

# 5. Chạy server
CMD ["uvicorn", "ai_service:app", "--host", "0.0.0.0", "--port", "8001"]
