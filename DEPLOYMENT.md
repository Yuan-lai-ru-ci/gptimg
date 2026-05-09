# Deployment Guide

This document describes the current production-style deployment for `gptimg`.

## Current Server Layout

Current known deployment target:

```text
server: 47.109.108.57
app path: /home/ecs-user/gptimg
container: gptimg_backend_1
internal app port: 127.0.0.1:8080
public frontend path: http://47.109.108.57/gptimg
public API path: http://47.109.108.57/gptimg-api/api/v1
```

The app is deployed as one Docker image:

- Stage 1 builds the Next.js frontend.
- Stage 2 builds the Go backend.
- Final image serves the Go API and static frontend files from `/app/frontend`.

## Required Runtime Environment

The container expects these environment variables:

```bash
SERVER_PORT=8080
DATABASE_PATH=/data/gptimg.db
JWT_SECRET=<random-secret>
ENCRYPTION_KEY=<32-byte-key>
STORAGE_PATH=/storage/images
FRONTEND_PATH=/app/frontend
ALLOWED_ORIGINS=http://47.109.108.57,http://47.109.108.57/gptimg,https://your-domain
```

Important:

- `JWT_SECRET` should be a long random string.
- `ENCRYPTION_KEY` must be 32 bytes/chars because it is used to encrypt stored API keys.
- Keep these values out of GitHub.

## Docker Build And Restart

From the server:

```bash
cd /home/ecs-user/gptimg
docker build -t gptimg_backend:latest -f Dockerfile .
docker rm -f gptimg_backend_1 || true
docker run -d \
  --name gptimg_backend_1 \
  --restart unless-stopped \
  -p 127.0.0.1:8080:8080 \
  -e SERVER_PORT=8080 \
  -e DATABASE_PATH=/data/gptimg.db \
  -e JWT_SECRET="<random-secret>" \
  -e ENCRYPTION_KEY="<32-byte-key>" \
  -e STORAGE_PATH=/storage/images \
  -e FRONTEND_PATH=/app/frontend \
  -e ALLOWED_ORIGINS="http://47.109.108.57,http://47.109.108.57:80,http://47.109.108.57/gptimg,http://lqbz.sheyice.com,https://lqbz.sheyice.com" \
  -v gptimg_backend-data:/data \
  -v gptimg_backend-storage:/storage/images \
  gptimg_backend:latest
```

Health check:

```bash
curl -s http://127.0.0.1:8080/api/v1/health
```

Expected:

```json
{"status":"ok"}
```

## Docker Compose

The repository includes `docker-compose.yml`, which is the preferred template when secrets are provided through environment variables:

```bash
cd /home/ecs-user/gptimg
JWT_SECRET="<random-secret>" ENCRYPTION_KEY="<32-byte-key>" docker compose up -d --build
```

If the host only has the legacy Compose binary:

```bash
docker-compose up -d --build
```

## Nginx Reverse Proxy

The current public paths are path-based:

- Frontend: `/gptimg`
- API: `/gptimg-api/api/v1`
- Storage assets: served through the backend storage route

Recommended Nginx shape:

```nginx
client_max_body_size 20M;

location /gptimg-api/ {
    proxy_pass http://127.0.0.1:8080/;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 600s;
    proxy_send_timeout 600s;
}

location /gptimg/ {
    proxy_pass http://127.0.0.1:8080/;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

After editing Nginx:

```bash
sudo nginx -t
sudo systemctl reload nginx
```

## Persistent Data

Docker volumes:

```text
gptimg_backend-data      -> /data/gptimg.db
gptimg_backend-storage   -> /storage/images
```

Backup:

```bash
mkdir -p ~/gptimg-backups
docker run --rm -v gptimg_backend-data:/data -v "$HOME/gptimg-backups:/backup" alpine \
  sh -c 'cp /data/gptimg.db /backup/gptimg-$(date +%Y%m%d-%H%M%S).db'
docker run --rm -v gptimg_backend-storage:/storage -v "$HOME/gptimg-backups:/backup" alpine \
  sh -c 'tar czf /backup/storage-$(date +%Y%m%d-%H%M%S).tar.gz -C /storage .'
```

Restore should be done with the container stopped.

## Operational Commands

Logs:

```bash
docker logs -f --tail=200 gptimg_backend_1
```

Container status:

```bash
docker ps --filter name=gptimg_backend_1
```

Restart:

```bash
docker restart gptimg_backend_1
```

Check listening port:

```bash
curl -s http://127.0.0.1:8080/api/v1/health
```

## Admin Setup

The application supports admin-only screens for:

- user creation/deletion
- quota tracking
- image API pool configuration
- LLM API pool configuration

Initial admin users may need to be created through the existing application flow or directly in SQLite, depending on the current database state. Avoid committing database dumps containing user data.

## Common Issues

### Upload returns 413

Nginx request body limit is too small. Set:

```nginx
client_max_body_size 20M;
```

Then reload Nginx.

### Generation times out

Image generation can take several minutes. Keep proxy read/send timeouts high:

```nginx
proxy_read_timeout 600s;
proxy_send_timeout 600s;
```

### API keys cannot be decrypted

Check `ENCRYPTION_KEY`. Existing encrypted API keys depend on the same encryption key that was used when they were saved.

### Frontend cannot call API

Check:

- `ALLOWED_ORIGINS`
- Nginx `/gptimg-api/` proxy path
- frontend API base path handling
- container health endpoint

## Release Checklist

1. Pull or upload the latest code to `/home/ecs-user/gptimg`.
2. Build Docker image.
3. Restart `gptimg_backend_1`.
4. Verify `curl http://127.0.0.1:8080/api/v1/health`.
5. Verify public page `http://47.109.108.57/gptimg/chat`.
6. Verify public API `http://47.109.108.57/gptimg-api/api/v1/health`.
7. Test login and one small generation request.
