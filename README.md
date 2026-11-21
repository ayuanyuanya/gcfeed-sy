# GCFeed

GCFeed 是一个面向短视频场景的 Feed 系统工程。项目用 Go API 单体、React Web 客户端、MySQL、Redis 和 RabbitMQ 承载内容供给、分发、消费、互动和治理链路。

## 当前状态

已实现能力：

- 后端分层结构：Domain、Application、Infrastructure、Interfaces。
- Gin HTTP 服务入口和 REST API 路由。
- MySQL + GORM 持久化。
- JWT 登录态。
- Redis Feed 缓存、热榜和互动计数。
- RabbitMQ 异步互动落库、视频发布事件和向量任务。
- React + Vite Web 客户端。
- 消息中心和播放优化接入。
- API 流程测试和 Web 生产构建。
- Prometheus 指标和 Grafana 监控面板。

重点待补能力：

- 审核后台。
- 后台运营。
- 系统治理。



