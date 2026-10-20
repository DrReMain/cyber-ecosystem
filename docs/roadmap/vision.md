# 长远 Roadmap — 备忘录与愿景

> 状态：粗放型备忘（north-star），不承诺时间；活跃前沿的专项 roadmap 见本目录（当前活跃：Agent 能力 `agent.md`；已收口前沿按治理法删除，durable 规则在 `docs/conventions/`）。
> destination: 长期驻留演进道（north-star；各阶段收口时更新，不归档）。
> 本文档回答"这个骨架要长成什么样、按什么顺序、什么不做"，供决策时对照，不作为排期依据。
> 2026-10-04 重理：把 IAM、文件、Agent 与第二服务落地后暴露出的共性问题收束为"平台硬化主线"，并重新排列其与产品化、垂直验证、业务分叉的关系。
> 2026-10-04 补充：记录个人/小团队长期维护的创建初衷、AI-native 双层目标，以及私有演进 / 公开快照 / 业务分叉的仓库演化模型。
> 2026-10-17 中立重铸：中立骨架战役（身份/租户/人口/provider/存储路由）收口——§2④ 租户原则、§3 P1 模型描述、§5a 决策清单随终态更新。

---

## 1. 定位

**Cyber Ecosystem 是一个通用开发骨架，不是一个有业务指向的完整应用。**

目标是让以下多种产品形态能从同一套骨架分叉开发，共享底座、各自生长：

| 产品形态 | 骨架需要提供的关键底座 |
|---|---|
| ERP / OA / CRM | 多租户、IAM、组织架构、CRUD 范式、审批流扩展位 |
| AI Agent | 流式传输（SSE/WS）、事件（MQ）、LLM 网关、向量存储、工具调用 |
| IoT / 机器人 / 无人机集群 | 双向实时流、设备身份（API Key 扩展）、遥测管道、命令下发、fleet 管理 |
| 类 YouTube/TikTok 音视频 | 对象存储 + presign/multipart、转码管线、HLS/CDN、播放器 |
| 类 Zoom 在线会议 | LiveKit（已部署）、信令、房间/预约、录制 |
| 无纸化会议（局域网交付，License+订阅） | 审批流（会签）、文件预览管线（缝）、留档合规、AI 会议总结、内网整体交付 |

**骨架的成功标准**：新产品从分叉到第一个垂直切片的时间最短；骨架升级时业务代码受
影响面最小；范式复制不依赖人肉对照文档。

当前骨架的实际形状已经比"工程模板"更进一步：proto 契约同时驱动 RPC、Connect 注册、
资源目录、access/builtin/datascope 标注、前端 operation 与菜单授权；system 是身份、
授权、租户、审计与资源目录的底座；agent 是第二个服务和 me-face 业务聚合的参考实现。
后续演进的重点因此不是继续横向加功能，而是把这条契约驱动链路的语义、性能、可靠性
与复制成本压到可长期承诺的产品化水平。

### 1.1 创建初衷与维护约束

这个骨架源于一个具体的长期约束：维护者以个人/小团队形态承接多类真实产品——
从简单后台、CRM、无纸化会议、表决，到在线音视频、具身智能巡检、短视频/直播。
这些产品表层差异极大，但共同底层高度重复；同时，维护者无法承担一个独立 DevOps
平台团队的治理成本。

因此骨架要对抗的首要敌人不是"缺少某个功能"，而是长期自由迭代造成的漂移：

```text
契约漂移
权限漂移
模块分层漂移
依赖边界漂移
生成物手改
文档过期
测试缺口
业务底座重复实现
```

演进策略是**仓库内可执行治理**，而不是预建重平台：

```text
proto 契约 + 生成器
conventions + roadmap 三车道
Nx 目标
架构 / 权限 / 传输测试
可观测与失败语义
AI/human 共用的编辑规则
```

开源骨架的意义是让外部项目从一个受约束的底座开始；私有演进的意义是保证这套约束
不会为了短期业务交付被反复削弱。

### 1.2 AI-native 的两层目标

AI-native 不是"内置一个聊天页"，而是同时覆盖开发期和运行期。

**开发期 AI-native**：coding agent 与人共用同一套范式，并被仓库规则约束：

```text
AGENTS / conventions / active roadmap
Nx 工作流
生成物禁手改
测试与 docs check
可验证 definition of done
```

AI 提高生成速度后，漂移也会加速；因此骨架必须让 AI 在边界内工作，而不是自由改写
范式。

**运行期 AI-native**：真实应用的使用者不只是人类用户，还包括：

```text
用户分身 agent：代表某个用户执行任务
machine agent：通过 API key 等凭据无人值守执行任务
device agent：未来具身智能 / 巡检设备上的智能主体
internal automation：系统后台自动化
```

这些主体必须进入同一套身份、授权、审计与观测模型。业务应用可以通过骨架范式快速
内置纵向 agent 模块，但不能让 tool calling、RAG 或自动任务绕过 IAM。

### 1.3 仓库与产品演化模型

仓库演化采用"私有事实源 + 公开快照 + 业务分叉"：

```text
私有 GitLab 仓库
    └─ 骨架演进事实源；完整 roadmap / 前沿 / 商业化判断只留在这里

GitHub 公开仓库
    └─ MIT 发布快照；不承载私有演进过程，也不发布 working roadmap

真实业务项目
    └─ 从私有骨架 tag / release 或隔离分支开始
       业务聚合不回灌骨架主干，共性能力按第二消费者法则回收
```

业务分叉规则：

- 骨架主干只收多产品共性，不收单一业务观点；
- 产品侧变更如需回上游，先抽成能力或范式，并附带测试、观测与复制路径；
- 长期产品线优先独立仓库 + upstream tag 同步；独立分支只作为隔离/过渡手段，
  不成为业务代码反向污染骨架的通道；
- 公开快照不依赖私有文档；公开 README 只描述当前能力、使用方式与方向，不承载
  内部排期和未裁决设计。

---

## 2. 不可动摇的原则

1. **Proto 是唯一事实源** — gen/ 是产物；契约先行，三端（Go/TS/OpenAPI）同源生成。
2. **模块化单体优先，按抽取信号拆服务** — 不预先微服务；模块边界即未来服务边界。
3. **能力下沉 shared-go / shared-ts** — 跨服务复用的必须离开业务服务目录。
4. **单/多租户共用一套骨架** — 租户是注册表行（code join-key，不可变、无软删）；归属只住座位与路由轴（行 mixin 与存储桶读同一优先律解析）。退化态 = 单默认租户零感知运行；多租户不开新面，只加注册表行。
5. **拒绝默认放行** — access 注解 deny-by-default；漏注解大声失败。
6. **错误是一等公民** — reason 唯一、三语言 i18n、前后端同构闭环。
7. **可观测默认在** — trace/metrics/log/慢查询随骨架走，不是事后补。
8. **一切工作流走 Nx** — 有目标的操作必须声明在 project.json，禁止野命令。
9. **占位代码有生命周期** — 每个插桩注定走向"固化"（成为骨架能力）或"清除"
   （验证完成后删除）；不允许第三种"永远留着"。
10. **不投机脚手架** — 事件/outbox 等能力等第一个真实需求出现才实做，但留扩展点。
11. **能力语义先于能力数量** — 每个平台能力必须说明一致性、失败模式、幂等性、
    观测与验证切片；没有这些语义的能力不算完成。
12. **骨架复杂度需要偿还** — 每引入一个横切机制，必须同时给出复制路径、降级路径
    和删除/替换边界；否则复杂度会吞掉"通用骨架"的收益。

---

## 3. 阶段路线（P0 → P4+）

### P0 — 工程护栏（补欠账）【CI 项暂缓 2026-09-29】
> 2026-09-29 裁决：CI 与生成器同条件——骨架足够固化后再做（CI 固化的门禁规则
> 同样是范式的快照，范式未稳即固化 = 反复返工）。业务测试、已知问题修复、插桩
> 标记等其余项不受暂缓约束，随切片按需推进。

- system 服务业务测试（UC 层 port fake 优先）+ admin 冒烟测试；
- CI：`nx affected` lint/test/build + 生成幂等校验；
- 修存量已知问题（以当前代码与测试盘点为准，不沿用已失效的历史例子）；
- 临时代码显式标记（区分"范式代码"与"验证插桩"，为固化/清除与生成器做语义准备）。

### P1 — IAM 闭环（✅ 已收口 2026-09-29；身份面中立重铸 2026-10）
五个产品形态的共同前置。模型为 **RBAC / ABAC / Datascope 三维可组合授权**
（可单独或叠加使用，全部构建在既有认证机制之上）。M0–M7 全部落地并 e2e 验收：
权限目录零登记、三维独立可启用、组合语义测试钉死、单/多租户双模式、前端真授权、
审计（MQ 第一个真实消费者）、会话治理（epoch 印记强制下线）；API Key 预埋已按
占位生命周期清除（恢复 = git 历史 + additive 迁移）。引擎机制与装配规则见
`docs/conventions/kratos/CONVENTIONS.md` §5a；选型决策见本文 §5a。

中立重铸后的身份面（2026-10 战役收口）：**身份 = 用户行（无租户列）+ 座位
（(租户, 人口) 唯一——登录推导与会话人口的唯一来源：单座即铸、多座拒以
SEAT_REQUIRED、缺租户回落平台座位）+ 凭证**（密码面与交换面同表：kind 词表是
代码、instance 是 conf 行，wechat 为首例 provider；(kind, identifier)→user 全局
唯一，未绑定交换以失败同形拒绝反枚举）。**人口 = 原型注册表**（engine/patron
词表；patron 零授权面、引擎不被咨询）。**授权 = 编译快照引擎**（grants 与人口
原型同快照，Decide 零库；缺码 = 注册表不变量破坏，fail-closed 为 infra 非策略
判决）。**存储 = 按注册表域路由桶**（行过滤与桶路由一根轴，平台租户走配置默认
桶，域桶惰性供给）。

政企场景增量（MFA / 审批流 / 临时授权 / 账号生命周期 / SSO / SoD / 设备主体等）
全部能归约到三个原语，**按项目需求拉取，不早期全量实现**：

| 原语 | 归约到它的场景 |
|---|---|
| ① 写绑定行/改参数 + NotifyChanged | 审批流生效、权限复核、临时授权授予 |
| ② compile() 加一行过滤（与 Enabled 同形状） | 临时授权到期、休眠禁用、一切生命周期状态位 |
| ③ AttributeSource + ConstraintPolicy 插件 | step-up MFA、会话新鲜度、未来 ip/geo |

切片要点：临时授权 = principal_role 加 `expires_at` + compile 绑定查询一行过滤
（到期=放宽，与 policy 停用同族、与 role 停用相反）；MFA 基础层在 auth 域改 TOTP，
step-up 走插件家族（auth.strength 属性源）；SoD = 互斥角色对表 + 绑定入口校验
（引擎零改动）；ApiKey = 临时授权的机器版原型。骨架义务 = 保证归约成立，而非提前实现。

**审批流边界契约（确定会做，方向 Temporal；IAM 只守边界不实现）**：
① IAM 的 grant/revoke API 幂等、可编程、无副作用顺序依赖——审批引擎
（Temporal worker）作为普通调用方执行授权变更，IAM 不感知审批上下文；
② 审批动作本身是 operation（受 IAM 管：谁可发起/审批什么单据）；
③ "双人复核"类跨角色交集归审批流建模，不进授权模型。

### P2 — 平台硬化与骨架产品化（分水岭）【生成器族暂缓 2026-09-29】
> 2026-09-29 裁决：`new-service`/`new-module`/`new-client` 生成器暂缓——骨架尚未
> 特别完善，先走 P3 垂直切片把骨架压测补全，范式稳定后再固化工具（生成器固化
> 的是范式，范式未稳即固化 = 返工）。本节的硬化项、README、ADR 体系、固化/清除
> 盘点不受暂缓约束，可按需先行。

P2 的目标不是把所有增强一次性做完，而是把已经落地的能力从"链路可用"推进到
"语义可承诺、失败可解释、性能有边界、复制成本低"。硬化项按四条收敛：

#### P2-A 契约与授权语义闭环

- **授权面语义统一**：角色保存侧与前端资源目录必须共用同一可授权判定——
  ADMIN 且非 builtin；datascope 仅在目录声明可表达时可选。禁止 UI 与后端各自
  解释 catalog。
- **catalog 信任边界**：为 manifest 建立生成/加载校验与指纹策略，明确热加载的
  误报/漏报条件；manifest 不再只是展示数据，而是授权管理输入。
- **ACCESS_APP 裁决**：要么为 APP audience 定义并接入真实中间件链，要么在当前
  骨架阶段明确禁用/reserved。enum 存在而运行链未定义的状态不能带入产品化。
- **授权撤销一致性等级**：保留版本戳 + PubSub + 定期兜底模型，但必须给出可观测
  的收敛指标与告警；若未来需要强撤销，定义同步确认或阻断策略，而不是模糊依赖
  定时 reconcile。
- **跨服务 Decision 契约**：Introspect 是否返回 scopes 必须显式定案。若业务服务
  要复用 system datascope，则远程 Decision 需携带 scope；若不返回，则非 system
  服务只能做 operation-level 授权 + 自有 owner 过滤，并在 ADR 中固化边界。

#### P2-B 可靠性与一致性硬化

- **审计交付等级**：区分 operational audit 与 compliance-grade audit。现有
  队列满丢弃、MQ 订阅失败降级属于前者；若产品需要合规审计，必须引入 outbox /
  durable carrier / 阻断策略之一，并把丢失率、延迟、恢复语义变成可观测指标。
- **流式交互幂等**：Chat 需要钉死"服务端已持久化但最终帧丢失"的 retry 语义，
  必要时引入请求/交换 id 或服务端恢复查询，避免客户端重发导致重复 exchange。
- **跨服务身份链路韧性**：为 system introspection 定义超时、熔断、重试、失败
  分类与降级边界；服务凭证/mTLS 或可信网络策略必须成为部署可选件，而非隐含
  假设。
- **文件与长任务矩阵**：把 presign 过期、multipart resync、对象先写/行后写、
  generation 中断、S3不可用、metadata 与对象不一致等场景固化为回归测试。
- **远程身份水合批量化**：业务列表页不得依赖逐用户远程 GetUser；定义 batch
  hydration 或受控缓存契约，并把缺 profile 降级语义保留。

#### P2-C 性能与容量边界

- **authz 热点折叠**：dept_tree 从每次 Decision 全量扫描迁入版本化 snapshot；
  部门/角色变更统一走 policy version bump。
- **跨服务延迟预算**：为 VerifySession / CheckOperation / hydration 建立预算与
  压测口径；若无法达标，再评估受控缓存，不提前引入决策缓存。
- **全量 snapshot 容量**：为 role/permission/policy/principal 数量建立阶梯数据
  与编译耗时、Decision p99 指标，验证当前内存编译模型的边界。
- **多副本变更风暴**：定义 policy version 抖动、重复 rebuild、PubSub 断连恢复
  的测试与限流策略。

#### P2-D 开发者体验与类型边界

- **前端 wire shape 类型化**：将 protojson/plain-object 边界从散落 `as` 断言
  收敛为显式 wire DTO / adapter 层；generated message 类型不得继续对业务代码
  "说谎"。
- **Connect/WS 错误契约**：WebSocket 方言需要保留或恢复 Connect code/reason，
  保证前端 error domain 分类不降级为 unknown。
- **平台能力装配档案**：新服务 Platform 目前要求 cache/db/storage/mq 全量配置；
  产品化前提供能力需求矩阵或装配 profile，避免不用存储/MQ的业务服务被迫携带
  全部部署依赖。
- **骨架复制清单收缩**：重复的 service-local kernel/mixin/server wiring 在不引入
  cross-service dependency 的前提下寻找生成或下沉时机；复制是边界策略，不是最终
  维护形态。

#### P2-E AI-native 主体与工具边界

- **Agent Principal 契约**：将 `user_agent`、`machine_agent`、未来 device agent
  定义为一等主体，与 user 一起进入认证、授权、审计与观测；先定契约，不先堆功能。
- **Agent Credential 模型**：API key 是真实需求触发的能力，但不是恢复旧预埋。
  需要定义 principal 归属、tenant、scope、过期、旋转、撤销、last-used、速率与
  hash 存储；明文只在创建时返回一次。
- **On-behalf-of 审计语义**：用户分身必须同时记录 actor principal、被代表用户、
  credential、task/tool 与 operation；machine agent 不得伪装成任何用户。
- **Tool = IAM operation**：agent 工具注册表必须把 tool 调用映射到受 access/catalog
  管理的 operation。工具输入/输出有契约，副作用有幂等与风险等级，禁止后端直接
  执行 planner 生成的自由函数调用。
- **Agent task 范式**：长任务需要 task id、状态、进度、取消、恢复、超时、artifact
  与审计。先做最小范式，第二个真实业务出现后再决定是否下沉或接 Temporal。
- **RAG 权限边界**：embedding、检索与 citation 必须携带 principal / tenant /
  datascope；agent 检索不得绕过行级授权。
- **模型与预算治理**：区分 platform-managed provider 与 bring-your-own provider；
  为模型路由、费用、限流、内容安全与 eval 留缝。
- **Agent eval**：prompt、模型、工具与知识库版本必须能进入回归用例；权限、安全、
  成本与输出质量不能靠肉眼自由漂移。

以上硬化边界与 P3 垂直验证共同给出稳定范式后，把"文档驱动复制"升级为
"工具驱动复制"：

- `new-service` 生成器：从参考实现派生 `app/services/<name>` 全套
  （目录、wire、server 三件套、conf proto、project.json、Nx 目标）；
- `new-module` 生成器：聚合根四件套（biz/data/service/module）+ proto 草稿；
- `new-client` 生成器：派生前端应用（shared-ts + 错误域 + auth 闭环）；
- 固化/清除第一轮执行（演示 transfer 内容、TODO 插槽等按 §6 清单处置）；
- 根 README + ADR 体系（记录 WS 方言、Centrifugo/LiveKit 定位、MQ 语义、租户双模式、
  审计交付等级、跨服务授权边界）。

### P3 — 垂直验证切片（每个域一条最小闭环，目的是压测骨架而非做产品）
| 切片 | 验证的底座 |
|---|---|
| IoT 设备：注册 → WS 遥测上行 → 命令下行 → fleet 列表 | 双向流 + MQ + 多租户 + API Key + IAM |
| AI Agent：会话 → SSE 流式回答 → 工具调用一次 | 流式 + 事件 + storage + LLM 网关扩展位 |
| 媒体：上传（presign/multipart）→ 转码触发 → HLS 播放 | storage（通用上传/下载已兑现，聚焦管线）+ 任务队列 + 媒体管线 |
| 会议：建房间 → LiveKit token → 加入 | LiveKit + 信令 + auth |

顺序裁决（2026-10-01）：Agent 最先（用户拉取：向量存储 + 流式 + LLM 网关先行），
IoT 其次（同时压四块底座），媒体再次（管线最长），会议最后（LiveKit 已部署，
接入面最窄）。

Agent 当前进度：服务骨架、IAM 消费、目录联邦、per-user 连接配置与会话/SSE 流式
（S0–S2）已收口；RAG、工具调用与 harness 继续由 `agent.md` 承载。P2-E 落地后，
该切片还要验收 agent principal、API key、tool 授权、任务审计与 RAG 数据边界。

垂直切片同时是 P2 硬化项的验收场：每个切片收口时必须声明它验证了哪条性能边界、
失败语义和复制范式；不能只证明 happy path。

### P4 — 真实业务分叉与骨架回归
- 真实产品在骨架上分叉，非固化代码（主要是业务代码）按需重写或调整；
- 建立骨架 → 产品的升级同步机制（tag merge / cherry-pick / 模板刷新 / 生成器重跑）；
  产品分支只做隔离，不把业务观点反向合并进骨架主干；
- 骨架侧只回收"多产品共性"固化，拒绝单一业务的特化反向污染骨架；
- 每次回收必须附带测试、观测与复制路径，不接受"业务代码搬个目录即成平台能力"。

### P4+ — 远期备忘（出现需求再细化）
- 服务拆分触发条件与拆分剧本（从 system/agent 双服务走向更多服务时重新定义）。授权侧剧本已预设计
  （IAM 收口存档）：**内嵌 PDP + 集中数据**——代码（中间件/接口/compile 纯函数）
  住 shared-go 链进每个二进制，策略由 system 编译发布 `GetPolicyBundle(version)`
  + 既有 PubSub 通知，各服务同构喂入 compile()；写路径只在 system，永不外放。
  拆分前只守三条缝：① AuthenticateSession 零业务依赖（验票禁查 user 表）；
  ② compile() 纯函数、数据源可替换；③ Session 结构/key 前缀、版本戳等共享契约
  住中立层（现在服务内，下沉时随包上移）；
- SSO / SCIM / MFA / 临时授权 / 账号生命周期 / 审批流接入（IAM v2 方向；
  datascope/ABAC 已随 M4/M5 落地，按需拉取，三原语与切片要点见 §3 P1）；
- 多区域/边缘部署、CDN 策略；
- 计量计费、配额体系（多产品共性，信号出现再立项）；
- proto v2 与废弃策略（docs/conventions/proto 已标注"第二个版本落地前必须定义"）。

---

## 4. 基础设施演进备忘

当前状态（已部署 / 未接入的，都是"基础设施领先应用层"的既定布局）：

| 组件 | 状态 | 激活计划 |
|---|---|---|
| Centrifugo | 已部署，零代码引用 | P3 期间决策：业务通知通道 vs 自研 WS 的分工边界，写 ADR |
| LiveKit | 已部署，零代码引用 | 会议切片（P3）接入：token 签发 + 房间服务 |
| NATS / PG-MQ | 能力完整，第一消费者在位（IAM 审计） | IoT 遥测（P3 第一刀）为第二消费者；P2 先明确审计交付等级 |
| SeaweedFS | 已接入（file 模块 + admin 文件页/头像/审计导出；按注册表租户域路由桶） | 通用上传/下载已兑现（文件能力前沿 2026-10-01 收口；租户域路由 2026-10 收口）；媒体管线（转码/HLS）留 P3 媒体切片 |
| SigNoz 栈 | 完整 | 保持；P2 把授权收敛、审计丢失、跨服务延迟纳入指标口径；上线前确认 trace/metrics 不再默认 false |
| Traefik edge | 就绪 | 多服务/多前端的路由矩阵在 P2 定义；服务凭证/mTLS 可信边界同场收口 |

直连类设施（LiveKit/SeaweedFS/Centrifugo）的授权模型已定案（2026-09-24）：
**控制面/数据面分离 + 能力令牌**——铸造 RPC 进 IAM 操作目录
（三维全适用），后端只铸造窄于 Decision 的短 TTL 能力，设施侧密码学自证；
TTL 是管控粒度与旁路成本之间的旋钮，敏感交互流仍走自家 connect WS 四原语。

**客户端/服务端栈扩展方向**（2026-09-29 补充）：后续按产品形态引入 RN、
iOS/Android native、Rust 等栈。proto 三面（gRPC/Connect/HTTP）+ S3 标准面是
栈无关可达性的保证（connect-swift/kotlin 官方、Rust 走 gRPC 面）；约束是
跨栈协议编排（如文件上传链）不得锁进单栈库，须保持栈无关的协议描述。

---

## 5. 已定架构决策（ADR 暂存；P2 建 ADR 体系后迁出）

### 5a. IAM 收口决策

机制与 MUST 在 `docs/conventions/kratos/CONVENTIONS.md` §5a，这里只存决策、
理由与被否方案（防"为什么当初这么定"考古）：

1. **三维挂同一 permission**：RBAC/ABAC/Datascope 都挂在 permission 行上，天然
   可独立或叠加使用。被否：三套独立权限系统（组合语义无法形式化）；casbin 包办
   一切（ABAC 策略与行级 scope 不是它擅长的形态）。
2. **casbin 混合定位**：casbin 只承担 RBAC 求值 + operation 模式匹配 + domain；
   事实源是 ent 表，启动/变更时编译成内存策略（casbin 是可替换实现细节，被
   `Authorizer` 接口隔离）。被否：完全自研匹配（约 50 行可满足 v1，但丢角色
   继承/domain/成熟匹配语义）；casbin adapter（无官方 ent 支持，丢失迁移体系）。
3. **组合语义 = 并集、无 deny**：单 permission 内策略 AND，跨 permission/角色
   UNION（权限只增不减）；v1 表达力 = DNF，缺的 NOT 与跨角色交集分别用角色组织
   与审批流解决。deny 的真实代价不是实现而是全员可解释性（引擎优先级、审计解释、
   前端消解、诊断链路全部复杂化；AWS IAM 前车之鉴）。启用条件：真实"例外管理"
   需求且角色重构无法覆盖。
4. **缓存分层 = 内存编译即缓存**：Decide 静态部分是纯内存遍历（微秒级、零 DB），
   Redis 取一次比算一次还慢——不加应用层决策缓存；ABAC 约束实时求值不可缓存
   （快照只存参数无 verdict，数据形状即免疫缓存污染）；多副本一致性 = 版本戳 +
   PubSub 重建（订阅断线重连校对 + 低频定期兜底）。重开条件：策略量或 principal
   数大到内存遍历成为压测热点。

中立重铸增补（2026-10 战役，身份/租户/人口/provider/存储路由）：

5. **座位承载租户与人口**：用户行无租户列；座位 (tenant, population) 唯一且是
   登录推导与会话人口的唯一来源；行过滤（tenant mixin）与桶路由（TenantStorage）
   读同一优先律（显式域开关 > subject > 平台常量）——一根轴，域趟不跨租户。
   被否：用户行租户列（多租户身份无处安放）；provider 端推导"家租户"（外部输入
   倒灌身份平面）。
6. **凭证是身份平面注册数据**：(kind, identifier)→user 全局无条件唯一、无租户
   无软删；kind 词表是代码（`provider_<kind>.go` init 注册、重复 panic、删文件
   即删注册），instance 是 conf 行（secret 走 APP_ 前缀 env 配置键）；未绑定交换
   读作失败交换（同形反枚举）。被否：membercredential 租户化聚合（外部标识符的
   全局唯一与租户语义冲突）。
7. **快照引擎零库判定 + 原型门**：人口原型与 grants 同快照；patron 主体零授权面
   （引擎不被咨询），快照缺码 = 注册表不变量破坏，fail-closed 为 infra。被否：
   每请求读 population 行（判定门与渲染门双读、双失败面）。
8. **注册表驱动而非面驱动**：多租户不开新面——清扫按注册表域逐趟、桶按域惰性
   供给、建租户语义归产品开通流（控制台无 Create）。被否：多租户管理面先于
   消费者（预测性抽象）。

### 5b. 待裁决决策（P2 ADR 候选）

1. **跨服务 Decision 是否携带 scopes**：携带则业务服务可复用 system datascope，
   但远程契约、缓存与泄漏面扩大；不携带则边界简单，业务服务必须自管行级 owner
   语义。
2. **审计交付等级**：operational audit 保持可用性优先；compliance audit 需要
   outbox/durable carrier/请求阻断中的一项，不能靠同一个 best-effort 管道承担
   两种承诺。
3. **APP audience 语义**：是保留 reserved、接入独立认证链，还是移除 enum 值；
   必须在第二个真实 APP 消费者出现前定案。
4. **前端 protojson wire shape**：是建立显式 DTO 层、生成 plain shape 类型，
   还是保留 message 类型并集中转换；目标都是消除业务代码里的类型欺骗。
5. **Agent principal 与 credential 边界**：API key 挂用户、租户还是独立 agent
   principal；用户分身是否必须降权；machine agent 是否可拥有独立角色。默认方向：
   credential 挂 principal，不直接挂 user；agent 不天然继承用户全部权限。
6. **Agent tool 授权形状**：tool 独立建模，还是映射到 proto operation。默认方向：
   tool 调用必须是 IAM operation，不能形成第二套权限系统。

---

## 6. 占位代码生命周期备忘（固化 / 清除）

| 插桩 | 走向 |
|---|---|
| Dashboard 壳（多标签/滚动记忆/面包屑/主题/i18n） | **固化** — 骨架自带管理端布局 |
| 登录/登出/session 闭环 | **固化** — 认证范式（HttpOnly cookie + SSR 转发） |
| user/dept 四层模块 | **固化** — 模块模板的参考实现（生成器的种子） |
| proto image → catalog manifest | **固化** — IAM 权限目录的数据源；P2 升级为带校验/指纹的授权输入 |
| transfer 四种流 | **固化**为传输层 e2e 回归基座；演示内容清除 |
| 53 卡片首页 / Home 空页 / TODO 插槽 | **已清除**（2026-09-24：demo 路由全拔，`/` 重定向工作台，工作台为唯一路由） |
| biz/ops/patrol 菜单结构 | **已清除**（2026-09-24，同上；导航派生自 routeTree 自动收敛） |
| permissions.ts 全放行 | **已完成真化**（2026-09-24：session 消费 GetCurrentUser 真实 permissions，mock 四人格删除；守卫真判定在位） |
| IAM API Key 预埋（schema/proto/错误码/文案） | **已清除**（2026-09-29：M6 搁置后无激活扳机；恢复 = git 历史 + additive 迁移，6050 号码 reserved；principal_type 字符串列保留）。2026-10-04 出现真实扳机：无人 agent 需 API key —— 以 P2-E 的 Agent Principal / Credential 重开，不恢复旧预埋 |

---

## 7. 质量与治理

- **测试金字塔**：底座（传输/安全）保持现有高强度；业务层补 UC 单测 + connect 集成；
  前端补冒烟。每个垂直切片交付时必须带"可抄的测试范式"。
- **失败语义测试**：授权收敛、审计丢弃、流式断连、multipart 过期、远端 introspection
  超时、catalog 热加载等横切场景必须有回归切片；happy path 不等于骨架可用。
- **变更纪律**：proto 走 breaking check（注意它看不见校验规则弱化，需 review 兜底）；
  骨架范式变更必须同步 conventions + 生成器。
- **ADR**：每个不可逆技术决策（WS 方言、MQ 选型语义、租户双模式、IAM 模型、
  Centrifugo/LiveKit 边界、审计等级、跨服务 Decision 形状）留档，防止后续
  "为什么当初这么定"考古。
- **文档即接口**：conventions 是范式合同；roadmap 是意图合同；两者过期即 bug。

---

## 8. 明确的非目标

- 不做"演示用完整产品"，不做业务观点（具体行业的流程/字段/规则属于分叉后的产品）；
- 不预先微服务化、不预先多租户管理界面（双模式中的 single 不需要）；
- 不为单一产品特化骨架（特化留在产品仓库）；
- 不追求功能数量，追求"每个能力都端到端真实可用 + 可测试 + 可复制"；
- 不把"性能优化"等同于提前缓存，不把"可靠性"等同于默认阻断；先定义等级、边界与
  观测，再选择机制；
- 不做万能超级 Agent 平台：agent runtime 是可复用范式，业务 agent 留在各自业务
  服务/模块中纵向生长；
- 不允许 agent、tool、RAG 或自动化任务绕过 IAM、租户、datascope 与审计。

---

## 9. 里程碑自检问句（每阶段收口时问）

1. 新服务/新模块/新前端，不读源码只读文档 + 生成器，多久能跑起来 hello world？
2. 骨架升级一个版本，产品侧 diff 有多大？
3. 每个能力能否一句话说清它的验证切片在哪？
4. 删掉任何一个占位页，有没有测试证明骨架能力仍在？
5. 下一个产品形态分叉时，哪些代码原样抄、哪些必然重写 — 是否显式可见？
6. 这个横切能力的失败模式是什么：丢数据、降级、延迟，还是阻断？谁观测、谁告警、
   谁恢复？
7. 权限、目录、前端 operation 与运行中间件是否仍来自同一契约语义，没有分叉？
8. 新服务的最小部署依赖是否清楚，能否只装配实际使用的平台能力？
9. 复杂度新增后，复制路径变短了还是变长了？如果变长，凭什么保留？
10. Agent 操作能否完整归因：actor 是谁、代表谁、用哪个 credential、调用哪个
    task/tool、命中哪条授权？
11. 产品同步一个骨架 release 的冲突面是否可见，业务特化是否仍留在产品侧？
12. 公开快照是否仍可独立理解和使用，未依赖私有 roadmap 或内部商业上下文？
