# Monitor Tool Commercial

一个 **前后端分离** 的网站可用性监控与告警系统：支持创建监控项、定时探测、记录历史、展示仪表盘，并通过 **告警分组/去重/降噪** 的新告警体系（`incident` + `alert_send_task` + Redis FSM）可靠地发送 **邮箱 / 钉钉** 通知。

---

## 功能特性

- **监控**
  - HTTP/HTTPS 探测（响应耗时、错误信息）
  - 定时检测（Cron）+ 手动触发检测
  - 监控项列表、详情、历史记录、暂停/恢复
- **告警（新体系）**
  - **降噪 FSM**（Redis）：宕机/恢复确认阈值 + 抖动压制窗口
  - **分组聚合**：以监控项为粒度的事件生命周期（`incident`）
  - **发送去重**：发送任务队列（`alert_send_task`）+ 原子 claim（`SELECT ... FOR UPDATE`）
  - 邮箱 / 钉钉通知（钉钉参数可在前端配置）
- **用户与权限**
  - 注册/登录
  - **JWT + Redis** 会话校验（支持“重新登录使旧 token 失效”）
  - 需要登录的接口通过 Gin 中间件统一鉴权
- **仪表盘**
  - 监控项统计、今日告警统计、最近告警记录
  - 告警配置展示（按当前登录用户实时读取）

---

## 技术栈

### Backend

- **Go** + **Gin**：REST API、路由分组、中间件
- **GORM** + **MySQL**：数据持久化、自动迁移
- **Redis**：JWT token 校验缓存、FSM 状态存储、监控实时状态缓存
- **robfig/cron**：定时检测与定时发送任务

### Frontend

- **Vue 3** + **Vite**
- **Element Plus**
- **Axios**（请求拦截器自动携带 token）

---

## 项目结构（简版）

```
monitor_tool_commercial/
  backend/
    cmd/server/                 # 程序入口
    config/                     # .env 配置读取
    internal/
      handler/                  # HTTP Handler（参数/响应）
      service/                  # 业务逻辑（监控、告警、用户）
      dao/                      # 数据访问（GORM/SQL）
      model/                    # 数据模型（MySQL 表映射）
    middleware/                 # Gin 中间件（JWT鉴权）
    pkg/                        # 可复用组件（jwt/redis/monitor/alert 等）
    router/                     # 路由注册
    .env                        # 本地开发配置（示例已提供）
  frontend/
    src/views/                  # 页面（Dashboard/Monitor/Alert 等）
    src/api/                    # API 封装
```

> 目录范式：后端为典型 **分层架构（Handler/Service/DAO）** + Go 常用 `cmd/internal/pkg` 布局；前端为常见 Vue SPA 工程结构。

---

## 快速开始（本地开发）

### 1) 准备依赖

- **MySQL**：创建数据库（默认名 `monitor_tool_commercial`）
- **Redis**
- **Node.js**（建议 LTS）
- **Go**（项目使用 Go Modules；请使用你本机可用的 Go 版本）

### 2) 配置后端环境变量

后端会读取 `backend/.env`（已提供示例）。常用配置：

```dotenv
# MySQL
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=123456
DB_NAME=monitor_tool_commercial
DB_CHARSET=utf8mb4

# Redis
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=123456
REDIS_DB=0

# App
APP_PORT=8080
APP_ENV=development
```

> 注意：后端启动时会执行 `AutoMigrate` 自动迁移表结构。

### 3) 启动后端

在仓库根目录执行：

```bash
cd backend
go run ./cmd/server
```

默认后端地址：`http://localhost:8080`

### 4) 启动前端

```bash
cd frontend
npm install
npm run dev -- --port 3000
```

默认前端地址：`http://localhost:3000`

> 后端 CORS 允许 `http://localhost:3000` / `http://127.0.0.1:3000`。

---

## 核心流程说明

### 1) 监控检测（Cron）

- 定时任务周期性扫描有效监控项并执行探测
- 每次探测写入 `monitor_history`，更新 `monitor` 当前状态与 `last_status`

### 2) 告警降噪（Redis FSM）

FSM 将“观测到宕机/恢复”转成事件：

- `down_confirmed`：确认宕机（创建/触发 incident 生命周期 + 生成宕机 send_task）
- `down_observed`：宕机持续（只更新聚合计数，不重复生成发送任务）
- `up_pending_observed`：恢复疑似（进入抖动压制窗口，不立刻发恢复）
- `up_final_confirmed`：最终确认恢复（生成恢复 send_task + 关闭 incident）

### 3) 发送去重（alert_send_task claim）

发送服务通过事务 + `SELECT ... FOR UPDATE` 原子领取任务：

- `pending` / “过期 processing” 任务可被 claim
- claim 成功后置为 `processing` 并写入 lock_token/locked_at
- 发送成功标记 `sent` 并写入 `send_time`；失败标记 `failed`

---

## 主要接口（概览）

> 路由前缀均为 `/api/...`，除注册/登录外其余接口需要 `Authorization: Bearer <token>`。

### 用户

- `POST /api/user/register`
- `POST /api/user/login`
- `GET /api/user/info`

### 监控

- `POST /api/monitor/create`
- `GET /api/monitor/list`
- `GET /api/monitor/detail/:id`
- `POST /api/monitor/run/:id`
- `POST /api/monitor/pause/:id`
- `POST /api/monitor/resume/:id`

### 告警

- `GET /api/alert/config`
- `POST /api/alert/config/update`
- `GET /api/alert/list`（告警记录列表，已兼容新告警体系）

---

## 文档

- `ALERT_SYSTEM_DESIGN.md`：告警分组/去重/降噪的设计与实现说明
- `INTERVIEW_NOTES.md`：面试复盘用的项目总结（技术栈、架构、Q&A）

---

## 常见问题（FAQ）

### 1) 为什么用了 JWT 还要加 Redis 校验？

JWT 用于携带声明与过期；Redis 用于 **服务端可控的登录态**：重新登录可立即让旧 token 失效，也便于后续实现“退出登录/踢下线”。

### 2) 告警为什么要分 `incident` 和 `alert_send_task`？

`incident` 表示一个监控项的一次“宕机→恢复”生命周期（聚合、去重、统计）；`alert_send_task` 表示一次具体的通知发送任务（队列化、并发安全、可重试）。

---

## License

暂未指定（如需开源协议可添加 MIT/Apache-2.0 等）。

