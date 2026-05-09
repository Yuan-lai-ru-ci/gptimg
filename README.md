# GPT Image PPT Generator

一个面向图片生成和 PPT 图片页生成的 Web 应用。前端使用 Next.js，后端使用 Go/Gin + SQLite，支持用户账号、额度、API 池配置、LLM 配置、图片参考图上传、剪切板/拖拽上传，以及基于参考图的 PPT 页面生成。

## Features

- 图片生成：文本生成图片、参考图生成/编辑图片。
- PPT Mode：上传参考图，按用户提示词生成 16:9 PPT 页面。
- 多图上传：支持文件选择、拖拽、剪切板粘贴，多次粘贴会追加图片。
- 三种 PPT 参考图模式：
  - `模式① 风格图 + 文字`：第一张图作为所有页面的风格参考，用户文字作为每页内容。
  - `模式② 多图内容 + 风格`：第 N 张图作为第 N 页核心内容参考，用户文字作为统一风格描述。
  - `模式③ 自由美化`：图片和文字一起作为参考，交给图片模型自由整理。
- 管理员后台：账号发放、额度查看、API 池状态、图片 API 配置、LLM 配置。
- 额度系统：普通用户按生成扣额度；管理员默认不扣额度。

## Tech Stack

- Frontend: Next.js 14, React 18, TypeScript, Tailwind CSS, Zustand
- Backend: Go, Gin, SQLite
- Deployment: Docker multi-stage build, single container serving backend and static frontend
- Storage: SQLite database volume + generated image storage volume

## Repository Layout

```text
.
├── backend/                 # Go API server
│   ├── cmd/server/          # server entrypoint
│   ├── internal/api/        # routes, middleware, handlers
│   ├── internal/services/   # image/LLM service logic
│   ├── internal/repository/ # SQLite repositories
│   └── internal/database/   # database initialization
├── frontend/                # Next.js frontend
│   ├── src/app/             # app routes
│   ├── src/components/      # chat/settings UI
│   └── src/lib/             # API clients and Zustand stores
├── Dockerfile               # production single-container build
├── docker-compose.yml       # deployment compose file
└── DEPLOYMENT.md            # deployment and ops guide
```

## Local Development

### Backend

```bash
cd backend
go mod download
go run cmd/server/main.go
```

Required environment variables:

```bash
SERVER_PORT=8080
DATABASE_PATH=./data/gptimg.db
JWT_SECRET=replace-with-a-random-secret
ENCRYPTION_KEY=replace-with-32-byte-key
STORAGE_PATH=./storage/images
FRONTEND_PATH=../frontend/out
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

For production, the frontend is statically built and copied into the backend container.

## Production Build

```bash
docker build -t gptimg_backend:latest -f Dockerfile .
```

Run with Docker:

```bash
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
  -e ALLOWED_ORIGINS="http://your-host,https://your-domain" \
  -v gptimg_backend-data:/data \
  -v gptimg_backend-storage:/storage/images \
  gptimg_backend:latest
```

See [DEPLOYMENT.md](./DEPLOYMENT.md) for Nginx path deployment and operational commands.

## API Overview

All API routes are under `/api/v1` inside the container. In the current Nginx deployment they are exposed as `/gptimg-api/api/v1`.

- `POST /auth/login`
- `GET /auth/me`
- `GET /sessions`
- `POST /sessions`
- `GET /sessions/:id/messages`
- `POST /images/generate`
- `POST /ppt/generate`
- `GET /admin/users`
- `POST /admin/users`
- `GET /admin/api-pool`
- `GET /admin/llm-pool`

## Notes

- Do not commit production secrets, API keys, database files, or generated images.
- `JWT_SECRET` and `ENCRYPTION_KEY` must be changed before production use.
- The image API can be slow; backend HTTP clients are configured with long timeouts.
- Nginx must allow sufficiently large request bodies for image upload.
