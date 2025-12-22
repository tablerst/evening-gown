## 依赖选用

gin + gorm(postgres) + go-redis/v9 + minio-go/v7 + jwt/v5

## 快速开始

1) 复制环境变量文件：

- 参考 `./.env.example`
- 本项目会尝试读取 `./.env`（不会覆盖系统环境变量）

2) 运行后端：

- `go test ./...`（用于验证编译/依赖）
- `go run .`（启动服务，默认监听 `0.0.0.0:8080`）

3) （可选）插入演示数据（seed）：

- ⚠️ 确保 `POSTGRES_DSN` 指向**本地/测试**数据库（seed 会写入数据；请勿对生产库执行）
	- 建议先从 `./.env.example` 复制出 `./.env` 再修改
- （可选）设置 `ADMIN_EMAIL` / `ADMIN_PASSWORD`，seed 会在“尚无管理员”时自动创建单一超管
- 运行：`go run ./cmd/seed`

seed 会写入：

- Products：3 条（其中 2 条已发布、1 条草稿）
- Updates：3 条（其中 2 条 published、1 条 draft）
- Contacts：2 条
- Events：1 条

说明：前台公开接口只会展示已发布内容（草稿仅在后台可见）。

## 环境变量

应用：

- `APP_HOST`：默认 `0.0.0.0`
- `APP_PORT`：默认 `8080`

Postgres（空则禁用）：

- `POSTGRES_DSN`
- `POSTGRES_MAX_CONNS`
- `POSTGRES_MIN_CONNS`
- `POSTGRES_MAX_CONN_LIFETIME`

Redis（空则禁用）：

- `REDIS_ADDR`
- `REDIS_PASSWORD`
- `REDIS_DB`
- `REDIS_POOL_SIZE`
- `REDIS_DIAL_TIMEOUT`

MinIO（空则禁用）：

- `MINIO_ENDPOINT`
- `MINIO_ACCESS_KEY`
- `MINIO_SECRET_KEY`
- `MINIO_USE_SSL`
- `MINIO_REGION`
- `MINIO_BUCKET`

JWT（空则禁用）：

- `JWT_SECRET`
- `JWT_ISSUER`（默认 `evening-gown`）
- `JWT_AUDIENCE`
- `JWT_EXPIRES_IN`（默认 `24h`）

## 接口

基础：

- `GET /ping`：存活探针
- `GET /healthz`：依赖探针（postgres / redis / minio）。未配置的依赖会显示为 `disabled`。

JWT（仅在配置了 `JWT_SECRET` 时启用）：

- `POST /auth/token`：签发 HS256 JWT
	- Body：`{"sub":"your-subject"}`
	- 返回：`token`、`expires_at`
- `GET /auth/verify`：校验 JWT
	- 支持 `?token=...` 或 `Authorization: Bearer <token>`

## 中间件 / 调试

- Request ID：默认启用 `github.com/gin-contrib/requestid`
	- 响应头会包含 `X-Request-Id`
- CORS：默认启用 `github.com/gin-contrib/cors` 的 `cors.Default()`（开发环境友好）
	- 如需限制来源：设置 `CORS_ALLOW_ORIGINS` 为逗号分隔白名单
- pprof：默认关闭（避免暴露调试端点）
	- 设置 `ENABLE_PPROF=true` 后启用（注册在默认路径下，例如 `/debug/pprof/`）