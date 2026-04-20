# LiftIQ — Production Deployment Guide

This guide covers deploying the full LiftIQ backend stack to a single production server.

**Services deployed:**

| Service | Port | Description |
|---|---|---|
| TimescaleDB | internal only | PostgreSQL 16 + TimescaleDB extension |
| elevator-simulator | 8000 | Simulated elevator telemetry source |
| telemetry-ingestor | — | Background poller, no public port |
| compliance-engine | 8080 | ASME A17.1 compliance HTTP API |

---

## 1. Server requirements

| Resource | Minimum | Recommended |
|---|---|---|
| CPU | 2 vCPU | 4 vCPU |
| RAM | 4 GB | 8 GB |
| Disk | 20 GB SSD | 50 GB SSD |
| OS | Ubuntu 22.04 LTS | Ubuntu 24.04 LTS |

TimescaleDB grows at ~1 M rows/day at default settings (3 elevators, 5 s poll). At 20 bytes/row that's ~20 MB/day — 50 GB gives you ~7 years of headroom.

---

## 2. Install Docker on the server

```bash
# Add Docker's official GPG key and repo
sudo apt-get update
sudo apt-get install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

echo "deb [arch=$(dpkg --print-architecture) \
  signed-by=/etc/apt/keyrings/docker.asc] \
  https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Allow current user to run docker without sudo
sudo usermod -aG docker $USER
newgrp docker

# Verify
docker compose version
```

---

## 3. Clone the repository

```bash
# On the production server
git clone <your-repo-url> /opt/liftiq
cd /opt/liftiq
```

If the repo is private, use a deploy key or a personal access token:

```bash
git clone https://<token>@github.com/<org>/liftiq.git /opt/liftiq
```

---

## 4. Configure production secrets

```bash
cd /opt/liftiq
cp deploy/env/.env.prod.example deploy/env/.env.prod
```

Edit `.env.prod`:

```bash
nano deploy/env/.env.prod
```

Required fields:

```env
DB_USER=liftiq
DB_PASSWORD=<strong-random-password>   # change this — minimum 32 chars
DB_NAME=liftiq

TICK_INTERVAL=1.0
POLL_INTERVAL=5
STALE_WINDOW=5
LOG_LEVEL=warn
```

Generate a strong password:

```bash
openssl rand -base64 32
```

> **Never commit `.env.prod` to git.** It is listed in `.gitignore`.

---

## 5. First-time deployment

```bash
cd /opt/liftiq
make compose-prod
```

This builds all Docker images and starts all four services in detached mode.

### Verify the stack is healthy

```bash
# Check all containers are running
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml ps

# Health endpoints
curl http://localhost:8000/health   # simulator
curl http://localhost:8080/health   # compliance engine

# Confirm units appear after ~10 seconds
curl http://localhost:8080/units
```

Expected:

```json
{"units":["ELV-001","ELV-002","ELV-003"]}
```

### Check logs

```bash
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml \
  logs --tail=50 ingestor

docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml \
  logs --tail=50 compliance
```

---

## 6. Expose the compliance engine via a reverse proxy (recommended)

In production, put the compliance engine behind nginx with TLS instead of exposing port 8080 directly.

### Install nginx

```bash
sudo apt-get install -y nginx certbot python3-certbot-nginx
```

### nginx config

```nginx
# /etc/nginx/sites-available/liftiq
server {
    server_name api.your-domain.com;

    location / {
        proxy_pass         http://127.0.0.1:8080;
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_read_timeout 30s;
    }
}
```

```bash
sudo ln -s /etc/nginx/sites-available/liftiq /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# Issue TLS certificate (requires DNS pointed at this server)
sudo certbot --nginx -d api.your-domain.com
```

Update the mobile app `.env` to use the HTTPS URL:

```
EXPO_PUBLIC_COMPLIANCE_API_URL=https://api.your-domain.com
```

---

## 7. Firewall

Only expose what is needed:

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp    # nginx HTTP (redirects to HTTPS)
sudo ufw allow 443/tcp   # nginx HTTPS
# Port 8000 and 8080 are NOT opened — accessed only via nginx or internal network
sudo ufw enable
```

---

## 8. Updating to a new release

```bash
cd /opt/liftiq

# Pull latest code
git pull origin main

# Rebuild and restart with zero-downtime rolling update
make compose-prod
```

Docker Compose will rebuild only changed services. Containers are replaced one at a time.

To force a full rebuild (e.g. after a base image update):

```bash
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml \
  --env-file deploy/env/.env.prod \
  up --build --force-recreate -d
```

---

## 9. Database backup

### Manual backup

```bash
docker exec liftiq-timescaledb \
  pg_dump -U liftiq liftiq | gzip > /opt/backups/liftiq-$(date +%Y%m%d-%H%M).sql.gz
```

### Automated daily backup (cron)

```bash
sudo mkdir -p /opt/backups
sudo crontab -e
```

Add:

```cron
0 3 * * * docker exec liftiq-timescaledb pg_dump -U liftiq liftiq \
  | gzip > /opt/backups/liftiq-$(date +\%Y\%m\%d).sql.gz \
  && find /opt/backups -name "*.sql.gz" -mtime +30 -delete
```

This backs up at 3 AM daily and retains 30 days.

### Restore from backup

```bash
gunzip -c /opt/backups/liftiq-20260420.sql.gz | \
  docker exec -i liftiq-timescaledb psql -U liftiq liftiq
```

---

## 10. Reset data

```bash
# Truncate all rows, keep schema (safe — ingestor resumes immediately)
make db-reset

# Full schema reset (ingestor re-migrates on next start)
make db-reset-hard
```

---

## 11. Stopping and removing the stack

```bash
# Stop containers (data volume preserved)
make compose-down

# Stop and remove the data volume (DESTRUCTIVE — all DB data lost)
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml \
  --env-file deploy/env/.env.prod \
  down -v
```

---

## 12. Monitoring

### Container status

```bash
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml ps
```

### Resource usage

```bash
docker stats liftiq-simulator liftiq-ingestor liftiq-compliance liftiq-timescaledb
```

### Row ingestion rate (last 10 minutes)

```bash
docker exec liftiq-timescaledb psql -U liftiq -d liftiq -c "
SELECT
    u.unit_tag,
    time_bucket('1 minute', t.time) AS bucket,
    count(*) AS rows
FROM telemetry t
JOIN elevator_units u ON u.id = t.unit_id
WHERE t.time > NOW() - INTERVAL '10 minutes'
GROUP BY 1, 2
ORDER BY 2 DESC, 1;
"
```

### ASME threshold check

```bash
docker exec liftiq-timescaledb psql -U liftiq -d liftiq -c "
SELECT u.unit_tag, t.metric, last(t.value, t.time) AS latest_value
FROM telemetry t
JOIN elevator_units u ON u.id = t.unit_id
WHERE t.metric IN ('door_close_force_n', 'brake_response_ms', 'leveling_accuracy_mm')
  AND t.time > NOW() - INTERVAL '10 minutes'
GROUP BY 1, 2
ORDER BY 1, 2;
"
```

---

## 13. Troubleshooting

**Containers keep restarting**

```bash
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml \
  logs --tail=100 <service-name>
```

Common causes:
- `ingestor` or `compliance` crashing → usually a `DATABASE_URL` misconfiguration. Check `.env.prod`.
- `timescaledb` unhealthy → disk full or OOM. Check `df -h` and `free -m`.

**`make compose-prod` says `.env.prod not found`**

```bash
cp deploy/env/.env.prod.example deploy/env/.env.prod
# edit DB_PASSWORD
```

**Compliance engine returns `{"units":[]}`**

The ingestor hasn't written data yet. Check ingestor logs:

```bash
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml \
  logs --tail=30 ingestor
```

Look for `"msg":"poll complete"`. If absent, the ingestor can't reach the simulator — confirm the simulator container is healthy:

```bash
docker inspect liftiq-simulator | grep Status
```

**Database migration failure on first start**

Migrations are idempotent. If a migration fails, check the ingestor log for the error, fix the underlying issue (usually a TimescaleDB version mismatch), then restart the ingestor:

```bash
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml \
  --env-file deploy/env/.env.prod \
  restart ingestor
```
