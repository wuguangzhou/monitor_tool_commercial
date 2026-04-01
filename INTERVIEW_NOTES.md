# monitor_tool_commercial 复习笔记（面试向）

> 目标：用“做笔记”的方式把项目从 0 到 1 梳理清楚，便于面试前快速复习。内容包含：结构化大纲 + 每章面试问答（Q&A）+ 常见坑位排查清单。

## 0. 一句话概览

- **这是一个什么项目**：面向个人/小团队的 **HTTP/HTTPS 服务可用性监控** 系统，支持定时探测、状态展示、告警配置、告警发送（邮箱/钉钉）。
- **核心价值**：把 “探测 → 状态持久化/缓存 → 降噪/去重 → 发送任务 → 通知投递” 串成闭环。

---

## 1. 技术栈与目录结构

### 1.1 后端技术栈

- **Web 框架**：Gin
- **ORM/DB**：GORM + MySQL
- **缓存**：Redis（同时用作 token 会话校验、监控状态缓存、告警 FSM 状态）
- **定时**：`robfig/cron`（秒级表达式）
- **鉴权**：JWT + Redis token 校验

### 1.2 前端技术栈

- **框架**：Vue 3 + Vite
- **UI**：Element Plus
- **网络**：Axios（统一封装 `frontend/src/utils/request.js`）
- **路由**：vue-router（含简单登录守卫）

### 1.3 目录结构（记忆版）

- **后端入口**：`backend/cmd/server/main.go`
- **配置**：`backend/config/*.go`
- **路由**：`backend/router/*.go`
- **中间件**：`backend/middleware/auth_middleware.go`
- **业务层**：`backend/internal/service/*.go`
- **数据层**：`backend/internal/dao/*.go`
- **模型层**：`backend/internal/model/*.go`
- **探测/发送工具包**：`backend/pkg/monitor/*`、`backend/pkg/alert/*`、`backend/pkg/redis/*`
- **前端页面**：`frontend/src/views/*.vue`
- **前端 API**：`frontend/src/api/*.js`

### 1.4 面试 Q&A

- **Q：为什么后端用 Gin + GORM？**
  - **A**：Gin 性能与生态成熟；GORM 快速落地 CRUD/迁移；项目规模适合快速迭代。
- **Q：为什么 Redis 必须引入？**
  - **A**：token 会话校验、监控状态实时缓存、告警 FSM（debounce/抖动压制）的“跨进程状态”都需要低延迟 KV。

---

## 2. 本地运行与配置（后端/前端）

### 2.1 后端启动（关键点）

入口在 `backend/cmd/server/main.go`：

- `config.InitConfig()`：读取 `.env`（找不到就用系统环境变量）
- `InitMySQL()`：初始化 GORM，并 `AutoMigrate(...)`
- `redis.InitRedis()`：初始化 Redis
- `service.StartMonitorCron()`：启动全局监控定时任务
- `service.StartAlertCron()`：启动告警发送定时任务（当前实现里定时调的是 send_task 发送）
- CORS：允许 `localhost:3000` / `127.0.0.1:3000`

### 2.2 配置来源

- `.env`（可选）+ 环境变量
- MySQL：`DB_HOST/DB_PORT/DB_USERNAME/DB_PASSWORD/DB_NAME/DB_CHARSET`
- Redis：`REDIS_HOST/REDIS_PORT/REDIS_PASSWORD/REDIS_DB`
- 后端端口：`APP_PORT`（默认 `8080`）

### 2.3 前端启动

`frontend/package.json`：

- `npm install`
- `npm run dev`

Axios baseURL：`/api`（见 `frontend/src/utils/request.js`），通常依赖 Vite 代理或后端同域部署。

### 2.4 面试 Q&A

- **Q：你怎么处理跨域？**
  - **A**：后端加 Gin CORS 中间件放行本地前端域名；生产建议同域或 Nginx 反代统一域名。
- **Q：`.env` 找不到怎么办？**
  - **A**：代码会 fallback 到系统环境变量（`config/config.go` 有日志提示）。

---

## 3. 认证与鉴权（JWT + Redis）

### 3.1 登录流程（后端）

1) `POST /api/user/login` → `handler.UserLoginHandler`  
2) `service.UserLogin`：
   - 查 DB 用户（`dao.GetUserByPhone`）
   - bcrypt 校验（`pkg/encrypt`）
   - 生成 JWT（`pkg/jwt.GenerateToken`，有效期 7 天）
   - **把 token 写入 Redis**：`user:token:<userId>`，TTL=7天

### 3.2 鉴权中间件

`backend/middleware/auth_middleware.go::AuthMiddleware()`：

- 读取 `Authorization: Bearer <token>`
- `jwt.ParseToken` 校验 token
- Redis 校验：`GET user:token:<userId>` 必须等于请求 token
- 写入 Gin Context：`userId`、`phone`

### 3.3 前端 token 注入

`frontend/src/utils/request.js` 请求拦截器：

- 从 `localStorage` 读取 `token`
- 写入 `Authorization: Bearer ${token}`

### 3.4 面试 Q&A

- **Q：为什么 JWT 还要 Redis 校验？**
  - **A**：纯 JWT 无法做到“服务端主动失效/踢下线”。加 Redis 后可以主动过期、单点登录、注销生效。
- **Q：token 过期前端怎么处理？**
  - **A**：响应拦截器里遇到 `code=401` 弹窗提示并清 token 跳登录。

---

## 4. 数据模型与表结构（重点背）

> 面试复盘建议：把字段语义 + 状态枚举讲清楚。

### 4.1 User（`backend/internal/model/user.go`）

- `id/username/phone/password/avatar`
- `member_level/member_end_at`（会员体系雏形）

### 4.2 Monitor / MonitorHistory（`backend/internal/model/monitor.go`）

- `monitor`
  - `monitor_type`（目前主要是 HTTP）
  - `frequency`（秒）
  - `status`：`1=正常 2=宕机 3=暂停 0=初始化`
  - `last_status`：用于判断恢复（上一轮状态）
  - `error_msg`：最近错误
- `monitor_history`
  - 记录每次探测 `status/response_time/error_msg/monitor_time`

### 4.3 AlertConfig（`backend/internal/model/alert.go`）

- `alert_type`：`1=邮箱 2=钉钉`
- `email`
- `is_enabled`
- 钉钉配置（按用户）：
  - `dingtalk_webhook`
  - `dingtalk_secret`（可选）
  - `dingtalk_keyword`（可选）

### 4.4 新告警体系：Incident / AlertSendTask

`incident`（`backend/internal/model/incident.go`）

- 代表一个监控项一次“宕机生命周期”的聚合实体
- 核心字段：
  - `incident_seq`：生命周期序号（FSM 每确认一次宕机自增）
  - `state`：`down_active / recover_pending / closed`
  - down 聚合：`down_first_seen_at / down_last_seen_at / down_occur_count / last_error`
  - up 聚合：`up_first_seen_at / up_last_seen_at / up_occur_count`

`alert_send_task`（`backend/internal/model/alert_send_task.go`）

- 代表“要发送的一次通知任务”（可被 claim、重试）
- 字段：
  - `task_type`：`down / up`
  - `status`：`pending / processing / sent / failed`
  - `lock_token/locked_at`：claim 锁
  - `payload`：本次要发送的内容
  - `alert_type`：邮箱/钉钉

### 4.5 面试 Q&A

- **Q：为什么要拆成 incident + send_task？**
  - **A**：incident 用于聚合与去重（同一宕机生命周期只发一次）；send_task 用于投递队列化（claim、防并发重复发送、失败重试）。
- **Q：为什么 incident 里 up 时间要用 `*time.Time`？**
  - **A**：宕机阶段 up 时间未知，用 NULL 表达；避免 MySQL 严格模式下 `0000-00-00` 插入失败。

---

## 5. 监控链路（从页面到定时探测）

### 5.1 前端：监控项管理

页面：`frontend/src/views/MonitorList.vue`

主要能力：
- 列表 + 搜索 + 分页
- 新增/编辑/删除
- 手动检测（run）
- 暂停/恢复

API：`frontend/src/api/monitor.js`
- `/monitor/list`、`/monitor/create`、`/monitor/update/:id`、`/monitor/delete/:id`
- `/monitor/run/:id`、`/monitor/pause/:id`、`/monitor/resume/:id`

### 5.2 后端：定时调度策略

`backend/internal/service/monitor_service.go`

- `StartMonitorCron()`：每 10 秒触发一次 `RunAllMonitor()`
- `RunAllMonitor()`：
  - 拉取所有有效监控项（排除暂停）
  - **frequency 取模**：`now % frequency != 0` 就跳过（这是常见排查点）
  - 满足频率就调用 `RunMonitorOnce(monitorId, false)`

### 5.3 一次探测的完整流程

`RunMonitorOnce`：

1) `dao.GetMonitorWithLastStatus(monitorId)`
2) `pkg/monitor.HTTPMonitor(url)`（超时 5s；200-400 认为成功）
3) 插入 `monitor_history`
4) 更新 `monitor` 状态 + `last_status`（`dao.UpdateMonitorStatusWithLast`）
5) 更新 Redis 实时状态：`monitor:status:<id>`
6) 调用新告警链路（见下一章）

### 5.4 面试 Q&A

- **Q：频率取模有什么问题？**
  - **A**：会导致“用户以为每 N 秒检测，但实际上只有命中整秒才执行”。排查时建议用手动 run 触发。
- **Q：HTTPMonitor 为什么可能误判？**
  - **A**：只看状态码范围（200-400），不校验内容；网关返回 200 的错误页会被认为正常。

---

## 6. 新告警链路（FSM → incident → send_task → claim → 发送）

### 6.1 总流程图

```mermaid
flowchart TD
  probe[RunMonitorOnce\n(newStatus,errMsg)] --> fsm[EvaluateAlertFsm\n(RedisFSM)]
  fsm -->|events| engine[ApplyAlertFsmEvents\n(incident+send_task)]
  engine --> sender[SendPendingAlertTask\n(claim+send)]
  sender --> channel[Email/DingTalk]
  sender --> status[Update send_task status]
```

### 6.2 FSM（降噪核心）

文件：`backend/internal/service/alert_fsm.go`

- 状态：`ok / down_suspect / down_active / up_suspect / recover_pending`
- 阈值：
  - 次数：`downConfirmCount=2`、`upConfirmCount=2`
  - 时间：`max(20s, 2*frequency)`
  - 抖动压制：`max(60s, 5*frequency)`
- 事件：
  - `down_confirmed`：首次确认宕机（应该创建 incident + down send_task）
  - `down_observed`：宕机持续（只聚合 incident，不重复建任务；必要时补偿漏建任务）
  - `up_pending_observed`：恢复疑似/压制窗口
  - `up_final_confirmed`：恢复最终确认（建 up task + close incident）

### 6.3 incident 聚合与 send_task 生成

文件：`backend/internal/service/alert_engine.go` + `backend/internal/dao/incident_dao.go`

- `EventDownConfirmed`：`CreateIncidentDown(...)` + `CreateSendTask(..., down, pending)`
- `EventDownObserved`：`TouchIncidentDown(...)`（并有“补偿创建/补偿建 task”的逻辑）
- `EventUpFinalConfirmed`：建 up task + `CloseIncident(...)`

### 6.4 Claim 防重复发送（去重/并发安全）

文件：`backend/internal/dao/alert_send_task_dao.go`

核心：`ClaimAlertSendTasks(limit)`：
- 事务内 `SELECT ... FOR UPDATE`
- 把 `pending` 或 `processing` 且锁过期的任务更新为 `processing` 并写 `lock_token/locked_at`
- 返回领取到的任务集合

### 6.5 实际发送（邮箱/钉钉）

文件：`backend/internal/service/alert_send_service.go`

- `SendPendingAlertTask()`：
  - claim 任务
  - 并发发送（goroutine）
  - 成功：`MarkSendTaskSent`
  - 失败：`MarkSendTaskFailed`

通道实现：
- 邮箱：`backend/pkg/alert/alert.go`
- 钉钉：`backend/pkg/alert/dingtalk.go`（现在走 `SendDingTalkAlertWithConfig`，参数来自 `alert_config`）

### 6.6 面试 Q&A

- **Q：怎么保证同一条告警不会重复发送？**
  - **A**：发送侧用 claim（processing + locked_at），同一任务只会被一个 worker 领取；再加上 incident 聚合避免重复生成任务。
- **Q：降噪怎么做？**
  - **A**：FSM：down/up debounce（按次数或时间），recover_pending 抑制短暂恢复导致的刷屏。
- **Q：为什么要区分 down_confirmed / down_observed？**
  - **A**：down_confirmed 代表“首次确定宕机”，可以发一次通知；down_observed 代表“持续宕机”，只聚合统计避免刷屏。

---

## 7. 前端链路（页面与 API 映射）

### 7.1 路由与布局

`frontend/src/router/index.js`

- `/login`、`/register` 无鉴权
- 其余挂在 `Layout` 下，`beforeEnter` 做 token 检查
- 页面：
  - `Dashboard`
  - `MonitorList`
  - `AlertConfig`
  - `AlertRecord`

`frontend/src/views/Layout.vue`：侧边栏 + 顶部用户信息 + 头像上传 + 昵称修改

### 7.2 请求封装

`frontend/src/utils/request.js`

- baseURL: `/api`
- 请求拦截器加 `Authorization`
- 响应拦截器统一处理 `code != 200`，`401` 跳登录

### 7.3 告警配置页（含钉钉参数）

`frontend/src/views/AlertConfig.vue`

- 可选择邮箱/钉钉
- 当选择钉钉时展示：
  - `dingtalkWebhook/Secret/Keyword`
- 提交到：`POST /api/alert/config/update`

### 7.4 告警记录页

`frontend/src/views/AlertRecord.vue` 目前查询的是旧 `alert` 表接口：

- `GET /api/alert/list`（关键筛选：keyword、alert_sub_type、status）

> 注意：如果全面切换到新体系 `incident/send_task`，前端告警记录需要改接口或做镜像兼容。

### 7.5 面试 Q&A

- **Q：你怎么做全局错误处理？**
  - **A**：Axios 响应拦截器统一处理 `code!=200`，并对 401 做重新登录引导。
- **Q：路由守卫为什么只看 localStorage token？**
  - **A**：这是前端快速拦截；真正鉴权在后端中间件 + Redis token 校验。

---

## 8. 常见坑位与排查清单（强烈建议背）

### 8.1 “前端显示宕机，但没有告警/没有 send_task”

按排查顺序：

1) **是否真的检测到 `newStatus=2`**（看后端日志 `检测完成，新状态=...`，以及 `monitor_history`）
2) **debounce 阈值**：第一次失败只会进入 `down_suspect`，不一定马上 down_confirmed
3) **frequency 取模**：定时可能被跳过，测试用手动 run
4) **FSM key 是否存在/状态**：`GET alert:fsm:u:<userId>:m:<monitorId>`
5) **incident 是否插入成功**：MySQL 严格模式下 0000-00-00 会失败（已通过 `*time.Time` 解决）
6) **claim rows=0**：不代表 bug，代表没有 pending 任务

### 8.2 钉钉发送失败常见原因

- webhook 不对/过期
- 群安全设置开启关键词：消息没包含 keyword
- 开启加签但 secret 不对
- 网络访问钉钉域名失败

### 8.3 邮箱发送失败常见原因

- SMTP 授权码错误/过期
- 端口/加密方式不匹配（465 SSL vs 587 STARTTLS）
- 发件邮箱风控

---

## 9. 面试“讲项目”模板（60 秒版本）

1) 我做了一个 HTTP/HTTPS 可用性监控平台，用户可以添加监控项并设置频率。  
2) 后端用 Gin + GORM + MySQL 做数据持久化，Redis 做 token 校验与实时状态缓存。  
3) 探测任务用 cron 定时跑，执行探测后写 history 和最新状态。  
4) 告警系统我做了去重降噪：用 Redis FSM 做 debounce/抖动压制，把一次宕机生命周期聚合成 incident，再用 send_task 做投递队列，发送端用 claim 锁避免并发重复发送，并支持失败记录。  
5) 前端用 Vue3+ElementPlus，页面包括登录注册、监控管理、告警配置、告警记录，Axios 拦截器统一带 token 与错误处理。  

