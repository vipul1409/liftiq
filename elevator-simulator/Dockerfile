FROM python:3.11-slim

WORKDIR /app

# libpcap is required by BAC0 even when running --no-bacnet
RUN apt-get update \
    && apt-get install -y --no-install-recommends libpcap-dev curl \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8000

HEALTHCHECK --interval=10s --timeout=5s --retries=5 \
    CMD curl -f http://localhost:8000/health || exit 1

CMD ["python", "main.py", "--no-bacnet"]
