# security 开发计划 — 切片池

> 归属：security 前沿开发档。总则/裁决台账见 README；状态记 progress.md。
> 顺序律：简单 → 爆炸面小 → 风险低（S-1）；方案先行律：每片落码前先与用户讨论方案（S-1）。规模：XS（≤半天）/ S（≤1 天）/ M（2-4 天）/ L（>4 天）。
> 已裁不处理项已删（S-2：SH-1 刻意保留 / SH-3·4·5 不处理；S-3：SH-7 登录限流不处理——现阶段徒增复杂度），ID 不复用。证据均为 2026-10-26 于 esports/dev-lzx 树内核实所记（file:line）。

---

| 片 | 内容与验收要点 | 依赖 | 规模 | 爆炸面 | 建议 | 裁决 |
|---|---|---|---|---|---|---|
| SH-2 README 双语补漏 | 能力族清单补 `payment` 行（cache/storage/mq/payment 四族，v0.8.0 已回收但漏列）；审计「队列满按设计丢弃、非保证落盘」一句入性能形态节。验收：docs:check 绿，双语同步 | — | XS | docs only | **处理** | ✅ 2026-10-26 |
| SH-6 租户跨租拒绝守卫 | executable 守卫：租户 A 不能读写租户 B 的角色/用户/文件/审计（读面=Traverse 拦截器含「无 ctx 即跳过」的有意豁免；写面=biz 先读后写约定）。test-only 零生产代码。验收：ENT_E2E 跨租四臂拒绝断言 | — | S | test only | **处理**（静默回归高危面，姿态表义务） | ✅ 2026-10-26 |
| SH-8 agent 出网守卫（SSRF） | `chat/biz.go:253-271`：裸 `http.Client{Timeout:5s}` 默认跟重定向 + `io.ReadAll` 无上限；base_url 仅 `^https?://` 契约约束（`agent_config.proto:56`）；agent ladder 的 BUILTIN 档仅 SessionAuth（`server/server.go`）→ member 人群可探 compose 内网。修法（终版，S-4「拓扑不进代码」裁后）：拨号期仅拒协议保留类（回环/链路本地/组播/未指定——RFC 常量零部署知识，私网单播=一等场景放行）+ 禁重定向 + 响应体上限；单守卫件双注入（ListModels 裸 client + eino `ChatModelConfig.HTTPClient` 官方口）。残余（成员探 compose 兄弟）归部署面执法（网络分段/防火墙），法条注记随收口。先落 agent（重复驱动律，第二消费者再升 shared-go）。验收：保留类拒绝反例 ×N + 私网正例 + 正常拨号零回归 + Stream 路径同守卫 | — | M | agent 服务出站 | **处理**（高危项里最便宜；参考服务应自带安全出网范式） | ✅ 2026-10-27 |
| SH-9 authz 版本收敛修复 | `authz/data.go:276-278` 注释与代码相反（">= so a stamp that ever moves backwards never pins an old snapshot" vs `if r.version() >= ver { return }`）：Redis 丢数据后计数器重置 → 副本钉死旧快照直到计数器爬过旧水位；`policy_notify.go:16` Incr 失败仅告警，变更等下一次成功变更才自愈。修法：版本真值入库（与变更同事务）+ 比较改 `!=` + 有上限强制重建间隔。验收：计数器重置/Incr 失败注入后副本收敛断言 + 守卫测试 + 全量 e2e | — | L | **授权引擎（高承重）** | **处理**（正确性 bug 必修；排最后单独切片） | ✅ 2026-10-27 |
| SH-10 OAuth state 发起方绑定 | state 只存 providerKey（`auth/biz.go:163`），`BindCredential` 不校验发起方（`:196`）→ 绑定 CSRF 契约缺口；需客户端实现「GET 触发 bind」才成活（现状无消费者：admin 不用、ctr 只走 login exchange）。修法=state 绑定发起方标识（proto 契约改动）。**建议挂账**：选项 a 纯挂账（触发=首个 bind 消费者落地）；选项 b 现在加 proto 注释标缺口（防新项目照抄继承，不动契约字段） | D-2 | M | contract | **挂账**（b 形性价比更高） | ✅ 2026-10-27（D-2 裁 b：`BindCredentialRequest.state` 字段注释标缺口+触发条件+勿照抄警示；激活=首个 bind 消费者落地时补发起方绑定） |
