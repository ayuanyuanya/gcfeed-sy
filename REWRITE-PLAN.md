# GCFeed 重写计划（17 天）

> **我的项目**：`E:\aworkpath\go_work\GCFeed-sy`
> **参考答案**：`E:\aworkpath\go_work\GCFeed`（只读，别抄）
> 版本 v2 · 2026-11-17（改为前后端平铺结构）

---

## 0. 这份计划怎么用

- 每天只看当天那一节，别提前翻。
- **每天收工必须能 `go build ./...` 通过、服务能起来**，功能可以不全。
- 完成后把 `- [ ]` 改成 `- [x]`，并写一句话差异笔记（见第 7 节）。
- 卡住超过 30 分钟就停下来问，不要硬耗。

## 1. 目录约定（Day 1 第一件事）

你选的前后端平铺比原项目的 `apps/api` + `apps/web` 更直观，就按这个来。**下文中所有 `backend/...` 和 `frontend/...` 都指这两个目录。**

```text
GCFeed-sy/
  backend/                       # ← Go module 根，module GCFeed-sy
    go.mod
    go.sum
    cmd/
      feed/main.go               # HTTP 服务入口
      worker/main.go             # 异步消费入口（Day 10 才建）
    configs/
      config.yaml
    internal/
      domain/                    # 实体、错误、仓储接口
      application/               # 用例编排
      infra/                     # MySQL / Redis / RabbitMQ / JWT / 配置
      interfaces/http/           # 路由、Handler、DTO、中间件
    test/                        # 接口流程测试（Day 3 起陆续补）
  frontend/                      # React + Vite（Day 14 才建）
    package.json
    vite.config.js
    index.html
    nginx.conf                   # 生产镜像用（Day 16）
    Dockerfile                   # 生产镜像用（Day 16）
    src/
  docker-compose.yml             # 编排，放仓库根（Day 16）
  deploy.yaml                    # K8s 清单，放仓库根（Day 16）
  README.md
  REWRITE-PLAN.md
  NOTES.md                       # 差异笔记（第 7 节）
```

五条规则：

1. **`go.mod` 放在 `backend/`，module 名保持 `GCFeed-sy`。** import 前缀是 `GCFeed-sy/internal/...`，跟参考代码 `GCFeed/internal/...` 只差一个后缀，对照阅读最省力。
2. **整个仓库只能有一个 `go.mod`。** 现在的 `backstage/go.mod` 就是它，**把 `backstage/` 改名为 `backend/` 即可，go.mod 不用动**；`backstage/main.go` 是 GoLand 的 Hello World 模板（`fmt.Println("Hello and welcome, %s!", s)` 还是个错格式化，会打出 `%!s(MISSING)`），直接删。
3. **命名一次定死。** 如果你最终决定后端目录叫 `backstage` 而不是 `backend`，就把下文中所有 `backend/` 全局替换成 `backstage/`，不要两套混用。
4. **`frontend/` 先用 Vite 初始化就好**，不要现在就抄 `App.jsx`（2810 行，见 Day 14）。
5. 如果你本机 Go 版本低于 `go.mod` 里写的 `go 1.26`，把那一行改成你本地版本（`go version` 查），否则会报 `go.mod requires go >= 1.26`。

```powershell
# Day 1 起步命令（在 GCFeed-sy 下执行）
Rename-Item backstage backend
Remove-Item backend\main.go
New-Item -ItemType Directory -Force backend\cmd\feed, backend\configs, backend\internal | Out-Null

cd backend
go get github.com/gin-gonic/gin gorm.io/gorm gorm.io/driver/mysql
go get github.com/golang-jwt/jwt/v5 golang.org/x/crypto gopkg.in/yaml.v3
go get github.com/redis/go-redis/v9 github.com/rabbitmq/amqp091-go
```

> 依赖可以不一次性装全，`go get` 装当天要用的那个即可；`go mod tidy` 会帮你清理没用到的。

## 2. 规模与预算

以下路径均在 `backend/` 下。

| 文件 | 行数 | 备注 |
| --- | ---: | --- |
| `internal/application/feed/service.go` | 883 | **最重**，拆 Day5+Day6 |
| `internal/infra/cache/feed_cache.go` | 856 | **次重**，拆 Day7+Day8 |
| `internal/application/interaction/service.go` | 520 | 拆 Day9+Day10 |
| `internal/interfaces/http/upload/handler.go` | 490 | 含 ffmpeg 抽帧、MIME 校验 |
| `internal/application/recommendation/service.go` | 448 | Day 13 |
| `internal/infra/mq/rabbitmq.go` | 427 | Day 10 |
| `internal/interfaces/http/router/router.go` | 249 | 全项目目录页，分多次长大 |
| `internal/application/message/service.go` | 198 | Day 12 |
| `internal/application/video/fanout_worker.go` | 168 | Day 11 |
| `internal/application/video/service.go` | 165 | Day 3 |
| `internal/application/playback/service.go` | 120 | Day 12 |
| `cmd/worker/main.go` | 110 | Day 10 |
| `internal/application/exposure/service.go` | 74 | Day 12 |
| `internal/infra/persistence/migration/migration.go` | 71 | Day 2 |
| `cmd/feed/main.go` | 51 | Day 1 |
| `frontend/src/App.jsx` | 2810 | **单文件怪物**，Day14/15 只挑三块 |

后端合计约 **9,000~10,000 行**，前端约 **3,500 行**（含 `styles.css`）。

结论：全量手抄不现实。本计划设计为 **手写核心约 4,000 行 + 复用样板**，每天 2~3 小时。

## 3. 六条纪律

1. **先读后写，读完关掉参考文件再写。** 边看边抄等于没写。
2. **每天必须能跑。** 宁可功能残缺，不留编译不过的烂尾。
3. **允许直接复制的**：DTO struct、错误码映射表、Grafana JSON、`package-lock.json`、`.gitignore`。
4. **不许复制的**：Feed 组装、Feed 缓存、互动同步路径、MQ 封装、Worker 消费循环——这五块就是全部学习价值。
5. **每个文件收工记一行差异**：我的写法 vs 原实现，谁更好、为什么。
6. **遇到"原项目为什么这么写"的问题，先自己给两个候选答案再问。**

## 4. 进度总览

- [ ] Day 01 骨架与 `/health`
- [ ] Day 02 account 竖切 + JWT
- [ ] Day 03 视频发布 + 幂等
- [ ] Day 04 上传与静态文件（Range）
- [ ] Day 05 Feed 读取主干（游标分页）
- [ ] Day 06 Feed 组装（ID-Detail 分离、批量取）
- [ ] Day 07 Feed 页缓存
- [ ] Day 08 热榜（ZSET 分钟桶）
- [ ] Day 09 互动同步路径（Redis 状态与计数）
- [ ] Day 10 RabbitMQ + Worker 异步落库
- [ ] Day 11 关注关系 + 推拉结合
- [ ] Day 12 曝光 / 播放 / 消息
- [ ] Day 13 推荐 + 向量（embedding）
- [ ] Day 14 前端三块核心（apiRequest / 无限滚动 / 播放器）
- [ ] Day 15 前端页面补齐 + 全链路冒烟
- [ ] Day 16 工程化：Docker / compose / nginx / K8s / 监控
- [ ] Day 17 自主设计一个缺口模块（建议：审核后台）

---

## 5. 每日任务

### Day 01 — 骨架与 `/health`

**目标**：`go run ./cmd/feed` 起在 8080，`/health` 返回 200，配置从 YAML 读，数据库连接池能 ping 通。

工作目录：`backend/`

**新建/修改**

| 文件 | 参考 | 要点 |
| --- | --- | --- |
| `go.mod` | `../../GCFeed/apps/api/go.mod` | module 保持 `GCFeed-sy`，只补 require |
| `configs/config.yaml` | `../../GCFeed/apps/api/configs/config.yaml` | 可整份复制；注意 MySQL 端口是 **3307** |
| `internal/infra/config/entity.go` | 同名文件 | 结构体 + yaml tag，先只写 port / jwt / database / redis 四段 |
| `internal/infra/config/config.go` | 同名文件 | 读文件、`yaml.Unmarshal`、默认值 |
| `internal/infra/database/database.go` | 同名文件 | `sql.Open("mysql", dsn)` + 连接池参数 + `PingContext` |
| `internal/infra/httpgin/gin.go` | 同名文件 | 创建 gin.Engine，`Run(addr)` |
| `cmd/feed/main.go` | 同名文件（51 行） | 加载配置 → 连库 → Register → Run |

**验收**

```bash
cd backend
go build ./...
go run ./cmd/feed
# 另一个终端
curl -i http://127.0.0.1:8080/health
# 期望：HTTP/1.1 200 + {"message":"All is well"}
```

**自测（写进 NOTES.md）**

1. `main.go` 里为什么先建 DB 再建 Router？反过来会怎样？
2. `config` 包为什么拆 `entity.go` + `config.go` 两个文件？合成一个有什么代价？
3. `/health` 不查数据库。这对 K8s 的 readinessProbe 意味着什么？（提示：`../../GCFeed/apps/deploy.yaml` 里 api 的探针就是打 `/health`）

---

### Day 02 — account 竖切 + JWT

**目标**：注册 → 登录拿 token → 带 token 访问 `/api/users/me`。这是全项目标准分层的模板，后面每个模块都复制这套形状。

**新建/修改**（顺序就是依赖顺序，别跳）

1. `internal/domain/account/entity.go` — User 实体 + 规则；**密码只存 bcrypt 哈希**（参考：原文件用 `bcrypt.GenerateFromPassword` / `CompareHashAndPassword`）
2. `internal/domain/account/errors.go` — 领域错误（账号已存在、密码错误…）
3. `internal/domain/account/repository.go` — 只定义接口，不写实现
4. `internal/infra/persistence/account/model.go` + `gorm.go` — GORM 模型与实现，注意 model ↔ entity 的转换函数
5. `internal/infra/persistence/migration/migration.go`（71 行）— AutoMigrate 清单，**每加一个模块就来这里登记一次**
6. `internal/infra/jwt/jwt.go` — 签发 + 解析
7. `internal/interfaces/http/middleware/jwt_auth.go` — 从 `Authorization: Bearer` 取 token，写进 gin.Context 的 user_id
8. `internal/application/account/service.go` — Register / Login / Me / UpdateMe
9. `internal/interfaces/http/account/dto.go` + `handler.go` — DTO 可以整份复制；handler 里做**领域错误 → HTTP 状态码**的映射
10. `internal/interfaces/http/router/router.go` — 第一版：`Register(g, cfg, db)`，先装 `/health`、`/api/users`、`/api/sessions`

**验收**

```bash
# 注册
curl -X POST http://127.0.0.1:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"account":"alice","password":"pass1234","nickname":"Alice"}'

# 登录，取出 access_token
curl -X POST http://127.0.0.1:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"account":"alice","password":"pass1234"}'

# 带 token
curl http://127.0.0.1:8080/api/users/me -H "Authorization: Bearer <access_token>"
```

**自测**

1. 为什么 `jwt.Manager` 既给 Service（签发）又给 Middleware（校验）？放两个实例行吗？
2. 领域错误到 HTTP 状态码的映射，你放在 handler 还是 service？原项目放哪？（看 `handler.go` 与 `service.go` 分别 import 了什么）
3. `UpdateProfileRequest` 的字段为什么用 `*string` 而不是 `string`？

---

### Day 03 — 视频发布 + 幂等

**目标**：`POST /api/videos` 能发视频，同一 `Idempotency-Key` 重放不产生第二条；`video` 与 `video_stat` 两张表同时写入。

**新建**

- `internal/domain/video/{entity,errors,repository}.go`
- `internal/infra/persistence/video/{model,gorm}.go` — **主表 + 统计表**两张
- `internal/application/video/service.go`（参考 165 行）
- `internal/interfaces/http/video/{dto,handler}.go` — handler 里读 `c.GetHeader("Idempotency-Key")`
- `migration.go` 登记两张表
- `router.go` 加 `videos` 路由组：POST / GET /:videoId / DELETE

**验收**

```bash
# 第一次
curl -X POST http://127.0.0.1:8080/api/videos \
  -H "Authorization: Bearer <token>" -H "Idempotency-Key: k1" \
  -H "Content-Type: application/json" \
  -d '{"title":"first","description":"d","media_url":"/uploads/video/a.mp4","cover_url":"/uploads/cover/a.jpg"}'

# 同 key 再来一次 → 返回同一个 id
curl -X POST http://127.0.0.1:8080/api/videos \
  -H "Authorization: Bearer <token>" -H "Idempotency-Key: k1" \
  -H "Content-Type: application/json" \
  -d '{"title":"first","description":"d","media_url":"/uploads/video/a.mp4","cover_url":"/uploads/cover/a.jpg"}'
```

SQL 验证：`select count(*) from videos;` 应该是 1。

**自测**

1. 幂等键存在哪张表/哪个唯一索引上？为什么不能只靠内存 map？（提示：多实例部署）
2. 为什么发布时要同时写 `video_stat`？只在点赞时才建这行行不行？
3. 删除视频用软删除还是硬删除？软删除后 Feed 查询要怎么过滤？

---

### Day 04 — 上传与静态文件（Range）

**目标**：`POST /api/uploads` 收 multipart，落盘到 `uploads/{kind}/`，返回 `/uploads/...` URL；浏览器能播放且能拖动进度条。

**新建**

- `internal/interfaces/http/upload/handler.go`（参考 490 行）
- `router.go` 里加 `g.Static("/uploads", "./uploads")` 和 `POST /api/uploads`

**必须自己实现的部分**

- `kind` 白名单：`video` / `cover` / `avatar` / `file`（原项目 `normalizeKind`）
- 大小限制分别卡（视频和图片阈值不同）
- 扩展名 + MIME 双重校验（只看扩展名会被伪造）
- 视频落盘后用 **ffmpeg** 抽封面（原项目 Dockerfile 里 `apk add ffmpeg` 就是为这个）
- 文件名生成规则（防冲突、防路径穿越）

**验收**

```bash
curl -X POST http://127.0.0.1:8080/api/uploads \
  -H "Authorization: Bearer <token>" \
  -F "file=@a.mp4" -F "kind=video"
# 返回 {"url":"/uploads/video/xxx.mp4", ...}

# Range 请求必须返回 206
curl -i -H "Range: bytes=0-1023" http://127.0.0.1:8080/uploads/video/xxx.mp4
```

**自测**

1. 为什么 `Range` 能不能用，决定了视频能不能"拖动"？
2. 上传目录为什么在容器里必须挂卷？（提示：`../../GCFeed/apps/api/Dockerfile` 与 compose 的 `api_uploads`）
3. 文件名如果直接用用户传的原始名，会有哪两个安全问题？

---

### Day 05 — Feed 读取主干（游标分页，先不做缓存）

**目标**：`GET /api/feed-items?scene=timeline&limit=5&cursor=...` 能翻页，**不重不漏**。

**新建**

- `internal/domain/feed/{entity,errors,repository}.go` — 定义 `Scene`、`FeedPageItem`、`FeedQuery`
- `internal/infra/persistence/feed/{model,gorm}.go` — 排序 `published_at DESC, id DESC`，按游标切割
- `internal/application/feed/service.go` — **只写 timeline 分支**，先不做缓存、不做组装（只返回 video_id 列表）
- `internal/interfaces/http/feed/{dto,handler}.go`
- `router.go` 加 `GET /api/feed-items`、`POST /api/feed-queries`

**验收**：插 100 条视频，每页 10 条翻到底：

```bash
curl "http://127.0.0.1:8080/api/feed-items?scene=timeline&limit=10"
# 用返回的 next_cursor 继续
curl "http://127.0.0.1:8080/api/feed-items?scene=timeline&limit=10&cursor=<next_cursor>"
```

写个 `test/feed_api_test.go` 断言：翻完 10 页拿到 100 个不重复 id。

**自测**

1. 为什么不用 `LIMIT/OFFSET`？在"有人同时在发视频"的场景下 offset 会出什么错？
2. 游标里为什么要同时带 `published_at` 和 `id`？只带 `id` 行不行？
3. 为什么这天的 Service 只返回 id 列表、不返回完整卡片？（这是 Day 6 的伏笔）

---

### Day 06 — Feed 组装（ID-Detail 分离 + 批量取）

**目标**：把 Day 5 的 id 列表变成完整卡片，**一次翻页的 SQL 条数是常数**，不是 N+1。

**改写** `internal/application/feed/service.go`

三件事：

1. **ID-Detail 分离**：Feed 页只存 `video_id + 排序字段`，详情另取。
2. **批量取卡片**：一页 10 条 → 1 次 SQL 拿 10 个 video + 1 次 SQL 拿 10 行 stat + 1 次 SQL 拿作者，而不是 30 次。
3. **补齐互动状态**：当前用户是否 liked / favorited（可选登录，未登录时跳过）。

**验收**：开 GORM 日志，一次翻页的 SQL 条数应该 ≤ 5 且与 limit 无关（limit=10 和 limit=50 条数一样）。

```go
// main 或测试里临时打开
db = db.Debug()
```

**自测**

1. 为什么"批量取"比"循环查"在 10 条时差别不大、在 100 条时是灾难？
2. 作者昵称/头像每次都 join 查，还是缓存在别处？两种做法的失效难度对比？
3. 如果某条 video 已被删除，但 id 还在页里，你的组装会返回什么？原项目怎么处理空位？

---

### Day 07 — Feed 页缓存

**目标**：连续两次同样请求，第二次不查 MySQL（日志可证）。

**新增** `internal/infra/cache/feed_cache.go` 的 Feed 页部分（原文件 856 行，这节只写约一半）

- Redis 客户端封装（`infra/cache/redis.go`）
- 缓存 key 设计：**带版本号和场景**，如 `feed:page:v1:{scene}:{cursor}:{limit}`
- 序列化方式（JSON 就够）、TTL、空值缓存（防穿透）
- 回源逻辑：miss → 查库 → 写缓存
- 在 `router.go` 里按 `cfg.Redis.Addr != ""` **条件装配**（参考 `router.go:85`，这是本项目一个很重要的可降级设计）

**验收**

```bash
redis-cli -n 0 keys 'feed:*'
curl "http://127.0.0.1:8080/api/feed-items?scene=timeline&limit=10"   # 第一次：查库
curl "http://127.0.0.1:8080/api/feed-items?scene=timeline&limit=10"   # 第二次：应无 SQL
```

**自测**

1. key 里的 `v1` 是干嘛的？某次改了响应结构后怎么处理？
2. 空结果要不要缓存？缓存多久？
3. 用户点赞后，缓存的页里 `liked` 字段是旧值吗？原项目怎么解决（提示：页缓存只放轻量数据，状态另外取）

---

### Day 08 — 热榜（ZSET 分钟桶）

**目标**：`GET /api/feed-items?scene=hot` 返回最近 1 小时热度排名；点赞后 10 秒内名次能变。

**新增**（`feed_cache.go` 的另一半）

- 写入：`feed:hot:minute:v1:{yyyyMMddHHmm}` 的 ZSET，`ZINCRBY` 计分
- 读取：合并最近 60 个分钟桶（`ZUNIONSTORE` 到一个临时 key 再取 TopN，注意临时 key 的 TTL）
- 计分规则自己定（点赞 +1、收藏 +2、评论 +3 之类），并写进注释
- 在 `interaction/service.go` 写入侧调用（Day 9 会回头接上）

**验收**

```bash
redis-cli ZREVRANGE feed:hot:minute:v1:<当前分钟> 0 -1 WITHSCORES
curl "http://127.0.0.1:8080/api/feed-items?scene=hot&limit=10"
```

**自测**

1. 为什么按分钟分桶，而不是一个大的 ZSET 加时间戳？
2. 60 个桶合并每次查询都做，成本可接受吗？如果要优化，你会怎么做（提示：滑动窗口 + 增量合并）？
3. 桶的过期时间设多长合适？设太短/太长分别出什么问题？

---

### Day 09 — 互动同步路径（Redis 状态与计数）

**目标**：点赞/取消/收藏/取消，接口立即返回最新计数，且**重复点赞不涨计数**。

**新增** `internal/application/interaction/service.go` 的同步部分（原 520 行，估算 200 行上下）

- 状态 key：`interaction:action:v1:{user_id}:{video_id}:{action}`
- 计数 key：`video:stat:counter:v1:{video_id}`
- 快照 key：`video:stat:v1:{video_id}`
- `PUT` 幂等：已点赞再点，返回当前状态而不是报错
- 返回值设计：接口要返回"最新计数 + 当前状态"，让前端乐观更新能对齐
- 评论创建/删除（这一块直接写库，不走 MQ）
- `router.go` 加 `videos/:videoId/{like,favorite,comments}` 一组路由

**验收**

```bash
curl -X PUT http://127.0.0.1:8080/api/videos/1/like -H "Authorization: Bearer <token>"
curl -X PUT http://127.0.0.1:8080/api/videos/1/like -H "Authorization: Bearer <token>"  # 计数不应 +1
curl -X DELETE http://127.0.0.1:8080/api/videos/1/like -H "Authorization: Bearer <token>"
```

**自测**

1. 状态用什么 Redis 结构存？`STRING` / `HASH` / `SET` 各自适合什么查询？
2. 计数为什么放 Redis 而不是直接 update `video_stat`？
3. 用户连点 20 次，Redis 计数和 MySQL 计数会一致吗？什么时候会漂移？

---

### Day 10 — RabbitMQ + Worker 异步落库

**目标**：关掉 worker，点赞依然成功；开 worker，MySQL 里 `interaction_action` 和 `video_stat` 最终被写对。

**新增**

- `internal/infra/mq/rabbitmq.go`（参考 427 行）：连接、声明 exchange/queue/routing key、Publish、Consume
- `internal/application/interaction/worker.go`（参考 45 行）：消费循环 + **手动 ACK/NACK** + 失败重试策略
- `internal/application/interaction/event.go`：`ActionChangedEvent` 结构
- `cmd/worker/main.go`（参考 110 行）：独立进程装配，只连库/Redis/MQ，不启 HTTP
- `router.go` 里按 `cfg.RabbitMQ.URL != ""` 条件装配（参考 `router.go:101`，连不上只打日志不退出）

**验收**

```bash
# 1) 只起 api，不起 worker
go run ./cmd/feed
curl -X PUT .../videos/1/like     # 应成功
mysql> select count(*) from interaction_action;   # 应该是 0

# 2) 起 worker
go run ./cmd/worker
sleep 2
mysql> select count(*) from interaction_action;   # 变成 1
```

**自测**

1. 为什么消费要手动 ACK？自动 ACK 在 worker 崩溃时会丢多少消息？
2. 消息重复投递（at-least-once）时，你的落库怎么保证不双写？（提示：唯一索引）
3. worker 和 api 用同一个二进制镜像吗？`cmd/worker` 为什么单独一个 main？

---

### Day 11 — 关注关系 + 推拉结合

**目标**：关注一个人后，他的新视频出现在我的 following 流里；粉丝少的作者走"推"，粉丝多的走"拉"。

**新增**

- `internal/domain/relation/{entity,errors,repository}.go`
- `internal/infra/persistence/relation/{model,gorm}.go`
- `internal/application/relation/service.go` — 关注/取关/列表，**关注时回填作者最近作品**（参考 `router.go:213` 的 `FollowFeedBackfiller`）
- `internal/application/video/fanout_worker.go`（参考 168 行）— 视频发布后异步写入粉丝收件箱
- Feed 增加 `following` 场景：收件箱优先，不够再回源
- `router.go` 加 `/users/me/following/*` 路由

**验收**

```bash
# 用户 A 关注 B，B 发视频，A 的 following 流里应出现
curl "http://127.0.0.1:8080/api/feed-items?scene=following" -H "Authorization: Bearer <A的token>"
```

**自测**

1. 粉丝阈值定多少合适？阈值判断放在哪一层？
2. "推"模式下，一个百万粉账号发一条视频要写多少次 Redis？怎么控制？
3. 关注时为什么要"回填"最近作品，不填会怎样？

---

### Day 12 — 曝光 / 播放 / 消息

**目标**：三个较小的模块一次做完，都是 100~200 行的量级。

**新增**

- `exposure`（参考 74 行 + `event.go`）：`POST /api/video-view-events`，同人同视频**只记一次**（Redis Set），发布 MQ 事件
- `playback`（参考 120 行）：`GET /api/playback-config`、`GET /api/preload-videos`、`POST /api/playback-qos-reports`
- `message`（参考 198 行）：`GET /api/messages`、`PATCH /api/messages`、`GET /api/message-stats/unread`
- 把 Day 9/11 的事件接到 message writer 上（点赞别人视频 → 作者收到通知）

**验收**

```bash
curl -X POST http://127.0.0.1:8080/api/video-view-events -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" -d '{"video_id":1}'
curl http://127.0.0.1:8080/api/message-stats/unread -H "Authorization: Bearer <token>"
```

**自测**

1. 曝光去重为什么用 Redis 而不是 MySQL 唯一索引？
2. 播放预加载返回几条？依据什么（下一个视频的候选）？
3. 未读数怎么算的？每次都 count 一遍吗？有没有更好的做法？

---

### Day 13 — 推荐 + 向量

**目标**：`scene=recommend` 能出结果且不重复；embedding worker 能为视频生成向量。

**新增**

- `internal/domain/recommendation/{entity,errors,repository}.go`
- `internal/infra/persistence/recommendation/gorm.go`（只有 gorm，没 model——注意这个不寻常的信号）
- `internal/application/recommendation/service.go`（参考 448 行）：召回 → 粗排 → 去重 → 返回候选
- `internal/domain/embedding/{entity,errors,hash_ngram,repository}.go` + `application/embedding/worker.go`
- `internal/interfaces/http/recommendation/{dto,handler}.go` + `/internal/*` 路由

**验收**：`?scene=recommend` 连续刷三次，结果有差异但不重复；`video_embeddings` 表有数据。

**自测**

1. 召回和排序为什么必须分开？合成一个函数会怎样？
2. 哈希 n-gram 为什么不需要训练？它牺牲了什么？
3. 推荐结果要不要缓存？和 Day 7 的页缓存冲突吗？

---

### Day 14 — 前端三块核心

**目标**：别抄 2810 行。只手写三块，页面用最糙的样式先跑起来。

工作目录：`frontend/`（`npm create vite@latest . -- --template react` 起手）

**必写**

1. `src/api.js`：`apiRequest` 封装 —— 拼 base、带 `Authorization`、统一错误处理、`Idempotency-Key`（参考 `App.jsx:2645` 附近）
2. `src/useInfiniteFeed.js`：无限滚动 —— IntersectionObserver + cursor 分页 + 去重
3. `src/VideoPlayer.jsx`：播放器 —— 自动播放、暂停其他、上报 view event

**可抄**：`index.html`、`vite.config.js`、`main.jsx`、CSS 变量骨架。

**验收**：能刷 timeline、点赞、点进详情。

**自测**

1. 为什么请求不能写死 `http://127.0.0.1:8080`？（提示：`nginx.conf` 与 `vite.config.js` 的 proxy）
2. 无限滚动怎么防止同一个视频被插入两次？
3. 视频"播到 3 秒才上报曝光"这个逻辑放哪？

---

### Day 15 — 前端页面补齐 + 全链路冒烟

**补**：登录/注册页、发布页（上传 + 进度）、个人主页、消息中心、搜索或热榜页。

**验收**：按附录 C 的 8 步冒烟流程完整走一遍。

**自测**

1. 上传大文件时怎么显示进度？（提示：`XMLHttpRequest.upload.onprogress`，`fetch` 没有）
2. token 过期了前端怎么处理？拦截 401 后跳登录还是静默刷新？

---

### Day 16 — 工程化

**目标**：一条 `docker compose up --build` 全起；K8s 清单能 apply。

**新增/配置**

- [ ] `backend/Dockerfile` — 两阶段，alpine 里装 `ca-certificates ffmpeg tzdata`
- [ ] `frontend/Dockerfile` — node 构建 → nginx 托管
- [ ] `frontend/nginx.conf` — **server 块**：`/` 静态 + SPA 回退、`/api` 与 `/uploads` 反代到 `api:8080`、`client_max_body_size`
- [ ] `docker-compose.yml`（仓库根）— mysql(3307) / redis / rabbitmq / api / worker / web(5173:80)，build context 分别指向 `./backend` 和 `./frontend`
- [ ] `deploy.yaml`（仓库根）— Namespace / Secret / PVC / Deployment / Service（web 用 NodePort）
- [ ] `/metrics`（promhttp）+ 一个业务 Counter
- [ ] `monitoring/prometheus.yml` + Grafana 面板
- [ ] k6 脚本

**验收**

```bash
docker compose up -d --build && docker compose ps
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:5173/            # nginx 托管的页面
```

**自测**

1. 为什么 web 镜像最终 base 是 nginx，而不是 node？
2. `nginx.conf` 里 `proxy_pass http://api:8080/api/;` 末尾那个斜杠去掉会怎样？
3. `/health` 不查库的话，MySQL 挂了 K8s 还会认为 api 健康吗？要不要加 `/readyz`？

---

### Day 17 — 自主设计一个缺口模块（建议：审核后台）

原项目 `docs/modules/review.md`、`admin.md`、`governance.md` 都还是"规划中"。这是检验你 16 天成果的最好方式——**没有参考答案**。

**必须产出**

- 表设计：审核任务、审核记录、处置动作
- 状态机：待审 → 通过 / 拒绝 / 转人工；视频状态与审核状态怎么联动
- 接口：查询队列、执行审核、批量处置
- 一条端到端测试
- 一份 `backend/docs/review.md`，按项目的模块文档模板写（职责/接口/表/规则/测试/前端接入）

**验收**：新视频进入待审队列 → 审核拒绝 → Feed 里查不到它。

**自测**：你怎么证明"拒绝了就一定不会出现在 Feed 里"？缓存里那份旧页怎么办？

---

## 6. 每 3 天回顾

Day 3 / 6 / 9 / 12 / 15 收工时做一次：

- [ ] `cd backend && go build ./... && go vet ./...` 干净
- [ ] `go test ./...` 通过（Day 3 起应该有测试了）
- [ ] 目录结构与第 1 节一致（没有乱塞的 `tmp.go`、`test2.go`）
- [ ] `router.go` 里每个路由都能说出属于哪条链路
- [ ] 差异笔记写了至少 3 条
- [ ] 能把今天做的东西用 3 分钟讲清楚（不能，就是没懂）

## 7. 差异笔记模板

每写完一个文件，往 `NOTES.md` 追加：

```markdown
## Day 06 · backend/internal/application/feed/service.go

**原实现做法**：先取 id 列表，再分三次批量查询，最后对齐空位。
**我的做法**：一次 join 把 video 和 stat 取出来。

**差异原因**：join 更少往返，但作者信息要 join 三张表。
**结论/取舍**：10 条内 join 更快；原项目选批量是为了让"取详情"这部分能单独换缓存。
**遗留问题**：如果 stat 行缺失，join 会丢数据，得用 LEFT JOIN。
```

## 8. 附录 A：关键文件地图

参考项目路径都从 `GCFeed-sy` 出发，用 `../GCFeed/...` 指过去。

| 想理解什么 | 看这个文件 |
| --- | --- |
| 服务怎么启动 | `../GCFeed/apps/api/cmd/feed/main.go`（51 行） |
| 依赖怎么装配 | `../GCFeed/apps/api/internal/interfaces/http/router/router.go`（249 行） |
| 配置结构 | `../GCFeed/apps/api/internal/infra/config/entity.go` |
| 数据库连接 | `../GCFeed/apps/api/internal/infra/database/database.go` |
| Gin 引擎 | `../GCFeed/apps/api/internal/infra/httpgin/gin.go` |
| JWT | `../GCFeed/apps/api/internal/infra/jwt/jwt.go` |
| Redis 缓存 | `../GCFeed/apps/api/internal/infra/cache/feed_cache.go`（856 行） |
| RabbitMQ | `../GCFeed/apps/api/internal/infra/mq/rabbitmq.go`（427 行） |
| 异步消费 | `../GCFeed/apps/api/internal/application/interaction/worker.go` |
| 数据模型 | `../GCFeed/apps/api/internal/infra/persistence/*/model.go` |
| 接口行为规格 | `../GCFeed/apps/api/test/*_api_test.go`（13 个文件） |
| 前端 | `../GCFeed/apps/web/src/App.jsx`（2810 行） |
| 部署对照 | `../GCFeed/apps/docker-compose.yml`、`../GCFeed/apps/web/nginx.conf` |

## 9. 附录 B：常见坑

1. **module 名和后端目录名是两件事**：module 叫 `GCFeed-sy`，目录叫 `backend`，import 是 `GCFeed-sy/internal/...` 不是 `backend/internal/...`。
2. **module 改名后 import 全红**：`GCFeed/internal/...` 要全换成 `GCFeed-sy/internal/...`。建议一开始就在自己的代码里从零写 import，不要粘贴。
3. **MySQL 端口**：compose 里映射到宿主机 **3307**，不是 3306。
4. **Redis / RabbitMQ 配空就降级**：原项目在 `router.go` 里用 `if cfg.Redis.Addr != ""`、`if cfg.RabbitMQ.URL != ""` 做条件装配，MQ 连不上只 `log.Printf` 不 panic。这个"可降级"设计建议保留。
5. **`AutoMigrate` 忘了登记**：新增表但没加进 `migration.go`，运行时报 "table doesn't exist"，排查半天。
6. **Windows 下同时跑 api 和 worker**：两个终端窗口，别用 `&`。
7. **`App.jsx` 不要整抄**：2810 行里大量是页面布局，抄了只会拖慢进度。
8. **`uploads` 目录**：容器里必须挂卷，否则重建镜像后视频全丢。
9. **幂等键是全项目的横切能力**：视频、关注、评论、点赞、播放上报都用到，Day 3 就要把它抽象对，后面才不用返工。

## 10. 附录 C：8 步冒烟流程

```bash
BASE=http://127.0.0.1:8080

# 1. 注册
curl -s -X POST $BASE/api/users -H 'Content-Type: application/json' \
  -d '{"account":"alice","password":"pass1234","nickname":"Alice"}'

# 2. 登录
TOKEN=$(curl -s -X POST $BASE/api/sessions -H 'Content-Type: application/json' \
  -d '{"account":"alice","password":"pass1234"}' | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
echo $TOKEN

# 3. 上传
curl -s -X POST $BASE/api/uploads -H "Authorization: Bearer $TOKEN" \
  -F "file=@a.mp4" -F "kind=video"

# 4. 发布
curl -s -X POST $BASE/api/videos -H "Authorization: Bearer $TOKEN" \
  -H 'Idempotency-Key: smoke-1' -H 'Content-Type: application/json' \
  -d '{"title":"smoke","description":"d","media_url":"/uploads/video/a.mp4","cover_url":""}'

# 5. 刷 Feed
curl -s "$BASE/api/feed-items?scene=timeline&limit=10"

# 6. 点赞
curl -s -X PUT $BASE/api/videos/1/like -H "Authorization: Bearer $TOKEN"

# 7. 看热榜
curl -s "$BASE/api/feed-items?scene=hot&limit=10"

# 8. 看消息未读
curl -s $BASE/api/message-stats/unread -H "Authorization: Bearer $TOKEN"
```

然后去 MySQL 验证 `interaction_action`、`video_stat`，去 RabbitMQ 管理台（`:15672`）确认队列已清空。

PowerShell 版等价写法（不依赖 bash）：

```powershell
$base = "http://127.0.0.1:8080"
$body = @{ account="alice"; password="pass1234"; nickname="Alice" } | ConvertTo-Json
Invoke-RestMethod -Uri "$base/api/users" -Method Post -ContentType "application/json" -Body $body
$login = @{ account="alice"; password="pass1234" } | ConvertTo-Json
$token = (Invoke-RestMethod -Uri "$base/api/sessions" -Method Post -ContentType "application/json" -Body $login).access_token
Invoke-RestMethod -Uri "$base/api/users/me" -Headers @{ Authorization = "Bearer $token" }
```
