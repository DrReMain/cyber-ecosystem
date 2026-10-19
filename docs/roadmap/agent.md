# Agent 能力前沿（P3 第一刀）

> 状态：活跃（2026-10-01 开启；用户裁决 Agent 切片先于 IoT，vision §3 P3 顺序已同步翻转）。
> 归宿：收口时 durable 规则回灌对应 conventions（kratos/proto/tanstack/deploy），本档按治理法删除。
> 范围对齐 vision §3 P3：会话 → SSE 流式回答 → 工具调用一次；压测面 = 流式 + 事件 + storage + LLM 网关扩展位 + 向量存储。

## 0) 开局裁决（2026-10-01）

- **框架 = eino**（cloudwego/eino，v0.9.21，Apache-2.0）。对比 Blades（go-kratos/blades，2026-03 发布）：成熟度差一个量级（242+ releases / 380 使用者 vs ~816 stars）；Blades 的 Kratos 亲和非真实杠杆——编排框架按能力纪律只住 biz 层缝里，与自家 transport 装配之间有防腐层。Blades 留作观察对象。
- **eino-ext 无 pgvector indexer**（清单：es/milvus/qdrant/redis/vikingdb 等）→ 向量数据面自建（pgvector-go + ent/pgx），eino 只占编排层（model/embedding/agent loop）；`retriever.Retriever` 接口小，自建 repo 适配。框架在缝上，数据面归自己。
- **agent = 第二个真实服务**（现仅 system）→ 落地即触发 vision P4+ 预埋的"第二个服务出现时定义拆分剧本"；多服务约定（远程 RP 可见性 / proto 域演化 / 新服务引导）第一次真实受力。
- **扩展归属切分**（已回灌 deploy CONVENTIONS §10）：initdb 管 postgres 库 + atlas_dev scratch（容错）；服务库的扩展 = 各自首张 atlas 迁移。

## S0) pgvector 部署面 —— ✅ 2026-10-01 收口

- 镜像层早已就位：`deploy/images/pg-extended`（postgres:18.3-trixie + PGDG postgresql-18-pgvector/postgis）= dev/pre 主 pg（compose.dev.yaml service `postgresql`）。
- 实测：**vector 0.8.6**（PG18 硬约束 ≥0.8.1；0.8 线 iterative scans / halfvec / sparsevec 在场）；atlas_dev 冒烟 = `vector(3)` 列 / cosine `<=>` 排序 / HNSW(vector_cosine_ops) 索引全绿。
- 脚本：`02-databases.sql` atlas_dev 建库后容错装 vector（新卷路径）；现存卷一次性补装已执行；一次性容器重放 initdb 验证新卷全路径绿。
- 后续（S1 时兑现）：agent 服务库壳 + 首张迁移 `CREATE EXTENSION IF NOT EXISTS vector`。

## S1) agent 第二服务骨架 —— ✅ 2026-10-03 收口

- 派生走"文档驱动复制"全程（拷贝集 / ent 两遍 stub dance / wire 冷启动 / 守卫与 mixin 测试改写 / redis db 隔离 / 首张迁移 + vector 扩展前置）——清单已回灌 kratos CONVENTIONS §7 "Deriving a new service"，skill `kratos-new-service` 入口就位。
- 业务载体 = per-user agent 连接配置（base_url + api_key）薄 CRUD：`agent_config`（builtin me 面）+ `agent_config_admin`（可剥离运营面：独立 proto + module，DAG 守卫执法"业务模块禁反向依赖"）；api_key AES-GCM 落库（`utils.Seal`，与密码单向哈希语义隔离），任何响应不回显。S0 遗留的库壳 + 首迁移 vector 前置已兑现。
- IAM 消费范式落地：system `IntrospectService`（VerifySession / CheckOperation，身份从凭证来、不信任下游声明）；agent 两段 Decide（builtin 短路 → 远程 grants，fail-closed）；错误链序法则（logging 内 / sanitize 中 / 熔断外）回灌 kratos CONVENTIONS §5a "Non-system services"。
- **超出 S1 预划**：角色授权树的目录联邦——`proto:generate:catalog` 从 proto image 派生 per-package manifest，system 读时热加载（regen 即生效、零重启、零 import）；细粒度授权对 agent 域闭环（普通用户 403→200 实证）。管线工具 Go 化（`proto/cmd/{catalog,clean-gen}`，os.Root 沙箱，cjs 退役）。
- 部署面全绿：compose 迁移 job + 服务块（`agent.api` 别名 / 无 host 端口）、traefik 包名前缀分流（admin strip-connect / api 裸过程路径）、edge 冒烟（401 / 403 / 200 + 跨服务水合经容器网络）。
- 刻意未做（S2+ 前置）：harness / 会话 / SSE；embedding 表（S3 随 RAG 落）；服务凭证缝（§5a 已记接入点：interceptor 并列 + conf.Remote credentials 块）；conf.Remote timeout 字段（请求上下文由 server 1s ctx 兜底，非请求上下文调用是 harness 前置条件）。

## S2) 会话 + SSE 流式 —— ✅ 2026-10-04 收口（流式框架先行）

- 纵贯管道一次成形：`chat.proto`（Chat server-streaming：content/reasoning 双通道增量 + 末帧 finish_reason；ListModels 透传 GET /v1/models）+ eino-ext openai seam（per-request 构造、ReasoningContent 双通道实测通）+ agentconfig 解密封印开启（`FindSealedByUserID` → `utils.Open`，apiKey 全链路零日志）+ admin agents-chat（@ant-design/x 视图层 + XMarkdown 流式管线，mermaid/LaTeX 在动态边界）。浏览器 bidi = connect 生态死路结论在案；流式消费范式回灌 tanstack CONVENTIONS §1，eino seam 注记回灌 kratos CONVENTIONS §5。
- 顺带修复：shared-go connect adapter 流式 ctx 富集（middleware 注入值现在进入 `stream.Context()`；adapter_ctx_test 回归锁定）。
- **会话为标题债**：conversations 内存态、历史客户端全量回传（`messages max_items=200` 悬崖）→ 递延独立切片：历史上移服务端 + 服务端裁剪 + tool-ready 消息形状（role=tool / content parts JSONB）预留。
- 刻意未做：harness（并发闸/重试）、audit（流面无审计中间键）、多模态、工具调用（S4）、RAG（S3）。
- 实录：qwen3-27b 本机载入失败（宿主 45.4GB 内存墙），e2e 全程 gemma-4-12b-it-qat；模型裸输出 `\begin{align}` 不带 `$$` 定界时 LaTeX 按原文透传——输出约定（定界符风格等）归 S3 服务端 system prompt 面。

## S3) RAG 闭环

- embed → index → retrieve → 流式回答；eino 编排进场；provider seam = eino-ext model/openai + OpenAI-compatible base_url（可换 GLM 等）。
- **开放：provider 终裁**（GLM 候选：环境已有 key，embedding-3 支持 2048 维可降维；压测维度参数化，不锁 provider）。

## S4) 工具调用一次

## 附) pgvector 深水区压测清单（S0 之后按需排期）

- ent `Modify` 表达 `<=> ORDER BY ... LIMIT k` + 过滤条件；pgvector-go 注册进 pgx 连接。
- 规模阶梯 10k / 100k / 1M（边界）× 维度 1024 / 2048 × recall@k / p99 / QPS。
- 过滤比例 × `hnsw.iterative_scan`（strict / relaxed）防 overfiltering；halfvec 半精度对比；并发写时索引构建代价。
