# 骨架风险修复（skeleton-risk）

2026-10-27 评审（system/agent/admin 资深架构师审查，守卫与测试清除后首轮）衍生的修复役。评审存档=会话记忆 `skeleton-arch-review-2026-10-27`；本档为唯一工作面，逐项处理。

## 法条与纪律

- 法条=根壳硬法 1（一切工作流走 Nx target；**测试一律 `./nx run <svc>:test`**，裸 `go test` 在 settings deny 里，不松动）与硬法 2（pnpm）。
- **测试纪律（2026-10-27 升格入 kratos CONVENTIONS §7，法源在那边）**：测试只有两个家——e2e 只在 `<svc>/internal/e2e/`（ENT_E2E 门）；unit 只针对 biz 层（`biz_test.go`）。每片 TDD——先写红测，最小改绿，再重构；测试随片留存为回归锚，项域收口、不铺开全量测试面（家法「业务模块默认不写测试」不因此推翻）。
- DB 面红绿走 `ENT_E2E=1` 门（dev 栈 postgres+redis 在跑才执行，否则 skip）——沿 80ba9a59^ 旧例惯例。
- 每片收口=计数声明行（根壳工作契约第四条）。
- 完成门禁=所属链（build/lint/format）+ 本片新测试计数 + 适用 live 探针。
- 2026-10-27 追记：同文件多 UC/RP 法条同步入 kratos CONVENTIONS §3（承重=esports entitle/pay/report 既有形状）。

## 复核修订（2026-10-27，对评审原稿）

- **R5 撤销**：chat content 上限存在——proto 注释声明 + `chat/service.go:165-176` handler 显式校验（流式旁路 unary protovalidate 是刻意设计）。归入「已注释确认」清单。
- **R4 加重**：双 dev 默认值并存且不一致——`agent/configs/config.yaml:34`（agent-dev-master-key）与 `deploy/docker/compose.app.yaml:112`（agent-insecure-dev-key）；prod 忘配环境变量即静默用公开钥。
- **BUILTIN 面清点（TR 升维为 SR-B 的证据底；2026-10-27 修正）**：初次清点 42 口、含「ListResource/PreviewGrants 两口提档候选」——**为注解 grep 假阳性**（reserved 注释文本误当注解；两口现状 ACCESS_ADMIN）。精确重清=全仓 43 口（system 16+agent 9+esports 18），全部=me-face/own-keyed 或 transfer 夹具。esports 18 与 agent 9=member 面设计意图。

## 切片池

顺序即处理序。2026-10-27 二次裁定入位。

### SR-0 测试目标回位（前置基建）✅

system 与 agent 的 `project.json` 各加 `test` target（`go test -race ./...`，cwd=projectRoot；ENT_E2E 门由测试自身 env-gate）。验证=target 空跑绿。

### SR-1 租户写保护 mixin 钩子（R1+D2）【裁定=A】✅

- **红**：复活 `system/internal/e2e/`（夹具形状取自 `80ba9a59^`），写 `tenant_write_guard_test.go`（ENT_E2E 门）：双租户夹具下，subject-A 的 ctx 直接 `Role.UpdateOneID(B行).SetName(...)` 与 `DeleteOneID(B行)` 断言**零行生效**；当前写穿→红。同测锁两个豁免位不回归（ExplicitTenant 优先、无 subject 跳过）。
- **绿**：`local_mixins/tenant.go` 增 `Hooks()`——OpUpdate|OpUpdateOne 挂 `tenant_id` 谓词，形状镜像同仓 `DatascopeMixin.Hooks()`（datascope.go:57-77 已证该形状在 SoftDelete 重写入链下成立）；优先级与查询拦截器一致（ExplicitTenant > subject > 无 subject 跳过）。
- **重构**：各 UC 前读保留（NotFound 语义与模糊反枚举仍需）；D2 注释随片补进 TenantMixin 头注释。
- **esports 解锁注**：SR-1 绿之前，esports 写路径暂守「前读收窄」现行约定。
- 开工前置：dev 栈 postgres+redis 在跑（否则红测 skip 不成红）。
- **实施事实（落地时钉出）**：① 谓词须挂**四 op**（+OpDelete|OpDeleteOne）——Datascope 模板的两 op 靠软删重写兜 delete，而 Role/permission 等 TenantMixin 表硬删，原生 delete op 直挂；② mixin Hooks 变更**必须重跑 `generate:ent`**——钩子注册编在生成物 `ent/runtime/runtime.go`，不重生成则执法静默缺席（绿阶段曾假绿一次，即此因）；③ `UpdateOneID`/`DeleteOneID` 零行返回 `ent.IsNotFound` 可断言；④ ~~残留拷贝位~~ **已清（同日裁定：三服务底座对齐）**：agent 与 esports 的 `TenantMixin` 已逐字节对齐 system 形状（路径改写拷贝，diff 验证 TRIPLE-ALIGNED），各自拦截器补齐 registry-read skip 与 ExplicitTenant 分支（两侧服务内零消费方，无行为回退面），各自 e2e 锚落位（agent=AgentConfig 软删表钉重写链形、esports=ActivationCode 硬删表钉原生 delete 形——delete 双形各有专锚），esports project.json 补回 test target。

### SR-2 master_key 收敛（R4）【裁定=默认值归层】✅

- **红**：单测 `chat.NewChatUC` 以 <16B key 构造应失败（当前仅非空→红）；`agentconfig.NewAgentConfigUC` 已 ≥16B（对照绿）。
- **绿**：阈值统一为一处；`config.yaml` 去默认值改必填（`${AGENT_MASTER_KEY:?...}`）；compose 保留**单一** dev 默认值（归层惯例=默认值住部署面，沿 Windows POSIX 役先例）；dev-up 裸跑路径由 devserver 注入 dev 值（实施时核 devserver env 通道）。
- **实施事实**：① kratos config 的 `${VAR:default}` 是 **config 树内解析非 shell 展开**——`APP_` 前缀 env source 把 `APP_AGENT_MASTER_KEY` 以 key `AGENT_MASTER_KEY` 进树、mapper 命中即覆写、未命中走 default、**无 default 返回空串**；kratos 不支持 `:?` 语法，必填语义=去默认后空串→UC 构造失败→启动 panic（fail-loud 链已活体实证）；② 阈值常量 `shared.MinMasterKeyBytes=16` 单源，chat/agentconfig 两构造统一引用；③ devserver 原只读 dev target 的 command 不读 env——**补齐 `options.env` 合入**（通用能力：target env 在 nx 直跑与 devserver 直启两路同语义）；④ dev 默认值单一落 `agent:dev` target env（devserver 经③消费），compose dev 栈保同值单条（弃 agent-insecure-dev-key）；⑤ **双面镜像核查（deploy §6 律，事后补核）**：`container/config.yaml` 自始即 `${AGENT_MASTER_KEY:}` 空默认形状（无需改），`configs/config.yaml` 统一为同形 `${AGENT_MASTER_KEY:}`（空默认与无默认在 mapper 里等价=空串，字面同形便于 field-aligned 扫描）；dev-up 新形状复证 panic 0。教训：改 `configs/` 触发 deploy §6 镜像义务，读图当跨法域（当时只读 kratos 侧）。

### SR-3 审计 CSV 消毒（R3）

- **红**：表驱动单测 `sanitizeAuditCell`（`=`,`+`,`-`,`@`,`\t` 开头前置 `'`；正常字段原样；空串）——函数不存在，编译红即红。
- **绿**：`audit/biz.go` emit 接线（UserAgent/Operation/HTTPPath/DenyReason/Actor/IP 全过）。纯函数片。

### SR-4 grant meta 缺失防御（R8）

- **红**：单测手工构造 snapshot（enf 含行、grants 图缺该键）→ `ImplicitGrants` 断言该 pattern 不出现；当前零值 ScopeKind="" 出现→红。
- **绿**：`authz/data.go` ImplicitGrants/ExplainGrants 两处 `if _, ok := snap.grants[...]; !ok { continue }`。

### SR-5 reconcile 互斥（R9）

小改：reconcile 入口加 mutex，头注释一句；验证=`-race` 单测双 goroutine 触发。TDD 形状弱，按小改+竞态检测收口。

### SR-6 注释补丁（D1/D3/D4）

无红绿。三处：`user/biz.go` ChangePassword 补「不撤存量会话」why（10-17 裁定入载体；admin Update password 路径同句）；`auth/biz.go` AuthenticateSession 补「新鲜度>缓存」取舍；`auth/biz.go` autoRegister 补 `hash=""`=密码面永久关。

### SR-7 会话验证热路径（R2）【裁定=C+量化触发】

接受现状，D3 注释随 SR-6 落。重开触发条件（其一即启）：① esports H5 放量门禁前评审；② introspect/验证路径 P99 或 system DB QPS 出现放量信号；③ 出现 A 案依赖补齐契机（tenant.enabled 入 stamp relevant 集）。届时按 A（载荷内嵌快照+stamp 比对）先行评估。

### SR-8 登录定向限速（R6）【裁定=递延】

不急于处理：登录后续可能引入生物验证等多重防御，触发风险后再处理可接受。重开触发=登录滥用事件 / 生物验证方案启动时并入设计。

### SR-9 dept reparent TOCTOU（R7）【裁定=注释封存】

保留 check-then-act，头注释记：并发对调可成环、环自愈=一次 reparent、IsAncestor visited 终止保障读不死循环。不写码。

### SR-B BUILTIN 面审计与处置（TR 升维）✅（改形两次收口）

- **原提档件撤销（误报二例，与 R5 同族）**：`ListResource`/`PreviewGrants` 的 proto 注解**现状已是 ACCESS_ADMIN**——评审清单把 reserved 注释行（`// builtin flag retired…ACCESS_BUILTIN…`）的文本误当注解（grep/awk 未锚定 `option (cyber.ext.v1.access)` 行形状）。教训入档：扫注解面必须精确匹配 option 行，proto 内注释文本可含同名词。
- **transfer 保留（2026-10-27 用户改裁）**：不删。四论据=数据面零访问（Echo/Pipe/Subscribe/Raw 只回环或发硬编码假事件）；流均有界（Subscribe 5×500ms 自终）；内容注入窗口被 SameSite=Lax cookie 模型关死（跨站 POST 不带 cookie→401）；且它是 kratos CONVENTIONS §3「Transport-fixture modules」法条的活例子，删之失载体。
- **守则件落位**：BUILTIN 档语义+选档守则入 proto CONVENTIONS §5 Access policy 段（member-plane trust ring：任何活会话可调/无平面卫/无 op 鉴权/无 datascope；只有 me-face/own-keyed 合格；信息型/资源型面 ADMIN 起步）。顺手偿还一处文档漂移：§5 豁免句仍写「ListResource 故意无注解练 503」——现状=proto 带 ADMIN、`accessExemptions` 空清单、503 覆盖移至 `TestDefaultGuard`，已按现状改写。
- **验证**：无 proto 改动（纯文档片），proto:check-drift exit 0；精确清点全仓 BUILTIN=43 口（system 16+agent 9+esports 18），system 侧全部=me-face/own-keyed/transfer 夹具，零信息泄露候选残留。live 探针免跑理由=ADMIN 档语义由注解+TestDefaultGuard 既有覆盖承载，无新行为面。

## 台账

| 片 | 状态 | 日期 | 提交 | 收口计数声明 |
|---|---|---|---|---|
| SR-0 | ✅ | 2026-10-27 | 未提交 | system:test/agent:test 双绿 exit 0（32.4s/7.2s），各 [no test files] |
| SR-1 | ✅ | 2026-10-27 | 见变更录 | 红 1（跨租户 update 得 nil=写穿）→绿 4 臂（e2e ok，-race，exit 0）；system:generate:ent 重生成 10 文件；build/lint/format 全 0；测试留存为回归锚。**底座对齐同日落**：agent/esports tenant mixin 对齐+各自 e2e 锚（红→绿同形）+esports test target 补回，三服务 build/test/lint/format 全绿 |
| SR-2 | ✅ | 2026-10-27 | 未提交 | 红 1（15B key 构造放行）→绿（module/chat ok）；yaml 去默认+compose 单值+dev target env+devserver env 通道四面落；活体探针=fail-loud panic 实录→env 通道修后 dev-up ready/三 listener/healthz 200/panic 0；build/lint/format 全 0 |
| SR-3 | ✅ | 2026-10-27 | 未提交（随下一批） | 编译红→绿（module/audit ok，表驱动 9 例）；消毒挂 7 自由文本列（tenant/actor/operation/http_path/ip/user_agent/deny_reason），数值与词表列（status/latency/method/principal_type）刻意不过——负数是数据不是公式；build/format/lint 全 0 |
| SR-4 | ✅ | 2026-10-27 | 未提交（随下一批） | 红（meta-less 行以 scopeKind="" 浮出=零值宽化实录）→绿；**测试锚后撤**（同日测试分层约定入法：unit 只钉 biz 层，data_test.go 违法被删；R8 防御+注释保留自证，私有图不对称态 e2e 包外不可模拟）；链复跑 test/lint/format 全 0 |
| SR-5 | ✅ | 2026-10-27 | 未提交（随下一批） | 小改收口（mutex+why 注释，无测试——reconcile 在 RP 层，两家园法下无单测面；验证=build/test/lint/format 全 0+走查：锁内无递归无外呼长阻塞，构造期初始 compile 不经锁） |
| SR-6 | ✅ | 2026-10-27 | 未提交（随下一批） | 四处 why 注释落（ChangePassword/admin Update 不撤会话、AuthenticateSession 新鲜度>缓存、autoRegister hash="" 密码面永久关）；自纠一处纪律违例（日期/裁定指认=过程性锚，删）；build/test/lint/format 全 0 |
| SR-7 | 封存（触发条件在档） | | | |
| SR-8 | 递延（触发条件在档） | | | |
| SR-9 | ✅（注释封存落） | 2026-10-27 | 未提交（随下一批） | dept Update reparent 头注释记 TOCTOU 取舍（并发对调成环可接受=读安全自愈+修复代价不成比例）；无行为变化 |
| SR-B | ✅（改形收口） | 2026-10-27 | 随本批未提交 | 提档件撤销（误报二例：reserved 注释行假阳性，两口现状 ADMIN）；transfer 保留（用户改裁，无害+§3 活例子）；守则入 proto CONVENTIONS §5+偿豁免句漂移；精确清点 43 口零候选；check-drift 0 |

## 变更录

- 2026-10-27：评审复核完成（R5 撤销/R4 加重/新增 SR-TR），本档建立，切片池定序。
- 2026-10-27（二）：用户四裁定入档——SR-1=A 方案；SR-2=默认值归层；SR-7=C+量化触发；SR-8=递延（生物验证时机）；SR-9=注释封存；SR-TR 升维 SR-B（BUILTIN 面审计，transfer 删除已裁）。SR-0 落地双绿。
- 2026-10-27（三）：SR-1 落地（四臂红绿全过，实施事实四条入节；agent/esports 拷贝位待裁）。同日用户裁定底座对齐：agent/esports mixin+e2e 锚+esports test target 一并落地，SR-0/SR-1 与本档随批提交。
- 2026-10-27（四）：SR-2 落地（构造阈值单源+部署四面+devserver env 通道补齐；kratos `${VAR:default}` 树内解析语义定案入实施事实——esports 新服务抄 yaml 形状时的语义底）。
- 2026-10-27（五）：测试两家园+同文件多 UC/RP 两法条入 kratos CONVENTIONS（§7/§3）；SR-4 测试锚按法后撤（data_test.go 删）；SR-3/4/5/6/9 连落（CSV 消毒/grant meta 防御/reconcile 互斥/注释四+一处/TOCTOU 封存）。役内只剩 SR-B 与两个封存递延位。
- 2026-10-27（六，役闭）：SR-B 改形收口——提档件撤销（误报二例）、transfer 保留（用户改裁）、守则入 proto CONVENTIONS §5+偿豁免句漂移。**skeleton-risk 役全闭**：SR-0..SR-6/SR-9/SR-B ✅；SR-7/SR-8 封存递延（触发条件在档）。esports 写路径解锁（SR-1 结构执法三服务同型）。
