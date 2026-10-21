# plane 前沿 — 平面准入骨架演进(人口/平面两轴补全)

> 状态:活跃前沿,2026-10-21 立项。源于 esports 真实场景暴露(后台封禁连坐/会员凭据可入后台壳),查证后定性为**骨架级机制未接线**,非路线错误。
> destination: 收口时 durable 法条回灌 kratos/tanstack CONVENTIONS(平面门/词表封闭/错误分流),切片记录剪枝为一行收口注(记入 progress.md);本档为前沿索引——总则/裁决台账在此,切片与进度在子档。
> 输入:2026-10-21 全链代码查证(shared-go security + system auth/authz/introspect + 双客户端)+ 豆包两轴分析(经逐条核实采纳其半)+ Google/OIDC 受众范式对照。

---

## 子档地图

| 档 | 内容 | Owner |
|---|---|---|
| plan.md | 切片池 PG-1..PG-5(依赖/规模/验收) | 开发共同维护 |
| progress.md | 节奏记录 + 查证结论存档 | 三人共写 |

---

## 0) 机制存量清单(2026-10-21 逐文件查证定案,防止重新发明)

| 层 | 已存在 | 证据位 |
|---|---|---|
| 数据 | population 行自带 plane 列;座位(seat)行 | `system/ent/population`、`seed.go` |
| 契约 | `LoginRequest.population` 字段(optional)已在 | `auth.proto:91` |
| 前门 | `resolveSeat(u, hint)`:带人口 hint 登录,无该座位即 `SEAT_REQUIRED` 拒绝;多座无 hint 也拒(强制选择) | `auth/biz.go:402` |
| 推导 | 验证时 `PlaneOf(population)`(authz 快照)写入 `Subject.Plane` | `auth/biz.go:305`、`security/identity.go` |
| 名望三层 | Enabled=活封禁(每请求查)/撤销标记=kick(只杀≤标记的会话)/租户禁用=租户级 | `auth/biz.go:244-309` |
| authz 分支 | `MemberSubject` → member 平面不读 grants(结构性零权限) | `authz/data.go:93` |
| PLATFORM 档 | `PlatformGuard` 默认租户密封(第三档已建) | `security/platform_guard.go` |
| 登录方式落座 | provider 行 → `landingPopulation`(微信→member,密码缺省→后台人口) | `auth/biz.go:454` |
| 会话隔离 | `session_token` host-only cookie → 分域名即双世界并存 | `custody.server.ts` |

**三缺口(全部是接线,不是机制)**:① admin/ctr 登录调用均不传 population——门在,没人报名字;② ADMIN 链无平面门,靠 grants 兜底——member 凭据可认证入后台壳(零权限泄露面);③ esports 的 `AuthenticateSession` 组装 Subject 时丢弃 wire 上已有的 Plane 字段(`esports/client/system.go:66`)。**勘误(2026-10-21 复查)**:缺口③为 esports 独有——agent 的 client 已映射 Plane(`agent/client/system.go:68`),两 client 非同代拷贝。

## 1) 总则(每片设计前对照)

1. **两轴正交**(PR-1):population=内聚轴(分组/数据视图,业务数据行,自由加);plane=安全轴(门的种类,编译期两词 `backend|member`)。谁也别替谁干活——内聚轴不得做安全裁决,安全轴不得做业务分组。
2. **信任模型**:客户端只报门牌(population hint,服务端座位表校验,报无座门牌即拒);平面永远由服务端从已证明座位推导,不来自任何请求输入。三道门全在服务端:发牌门(resolveSeat)/每请求门(ladder 验 Plane)/权限门(authz)。
3. **词表封闭**:平面词表=编译期两词,第三词是骨架级决策(触发制:真实出现新门种类);人口=数据行,业务自由增加,system 零代码改动。判据:**这段代码在零业务的纯骨架里是否一模一样存在**——是则本档,否则退回业务前沿。
4. **拒绝深度律**(UX 法):拒绝发生越早,UI 面积越小——登录步拒绝(一句文案+换号按钮)优先于空壳后台(整棵零权限页面树);能力级封禁(抖音式封直播权)归业务面完整表达,是业务主权,与本档正交。
5. **错误诚实分层**:密码错≠无座≠停用≠踢线≠能力封禁,各有独立文案与下一步动作;客户端禁吞 reason 一律映射"账号或密码不正确"。
6. **交付模式**(PR-5):esports 分支组先行落地,esports 真实场景验证,镜像迁移回 dev 骨架侧是**显式切片**(PG-5),不是口头承诺。
7. **原生形态律**(2026-10-21,用户裁定):存量机制的演进按"从零编写即此形状"实施——新代码是机制本来的样子,不是补丁;禁过渡注释、禁兼容垫片、禁残留开关。判据:重写该文件的人不会发现这里发生过演进。**射场与时间贯穿全程**:不止直接改动的代码,含爆炸半径与软关联(语义重叠/命名/组织/文件归属);PG-1..PG-5 每片收口前过一遍此律。

## 2) 裁决台账

- **PR-1 两轴正交**(2026-10-21):population=内聚、plane=安全,两轴分离执法;豆包分析经代码核实采纳,其"Plane 已在 subject/PlatformGuard 已存在"断言逐条属实。
- **PR-2 BUILTIN 面验 member 平面**(2026-10-21):采纳——纵深防御+对称隔离(后台账号无 member 座即进不了 C 端);运营测 C 端=显式授 member 座,可审计可撤销。**精化(同日,逐面盘点后)**:"BUILTIN 面验 member"精确为"纯 C 端 BUILTIN 梯挂 member 门"——system 的 BUILTIN 梯是双面基础设施(introspect 验证器/presign/transfer/auth 自务面,会员世界自己也走),结构性免挂;agent 的 BUILTIN 梯当前唯一消费者是 admin 监控页(chat×6/agent_config×3),挂门前必须先重定级(见 PR-6)。
- **PR-3 门与门牌同片**(2026-10-21):population 传参与登录错误分流必须同片落地——门建好挂错牌等于没建。
- **PR-4 dev 双主机名**(2026-10-21,**同日被 PR-9 撤销**):cookie 不按端口隔离,localhost:5173/5174 共享 `session_token` 互踩会话(2026-10-20 实测;2026-10-21 再次实证——会员会话直接登入 admin 工作台,截图在案)。原裁决 dev 以 localhost(ctr)+127.0.0.1(admin)分离;撤销理由:host 写死把"两世界"假设漏进机制层,第三 web 客户端即破产。过渡代价:PR-9 落地前 dev 双世界在 localhost 互踩(已知状态)。
- **PR-5 先行后镜像**(2026-10-21,用户裁定):本前沿在 esports 分支组处理,esports 真实场景验证后镜像迁移回骨架侧;与既有 worktree 纪律的方向相反,是本前沿的显式选择(验证优先)。
- **勘误存档**(2026-10-21):撤回 2026-10-20"esports introspection 不查 enabled"结论——验证链每请求查 Enabled,当日 curl 200 为账号已手动恢复的时间线伪影。
- **PR-6 agent BUILTIN 挂门移 PG-3**(2026-10-21,用户裁定):agent 的 chat/agent_config 监控面重定级为 ADMIN(proto 契约变更)是挂 member 门前置——admin 客户端是这些面当前唯一消费者,先挂门会当场打断监控;PG-1 只落 esports 的 BUILTIN member 门。
- **PR-7 平面门拒绝码复用**(2026-10-21,用户裁定):平面门拒绝复用 `GeneralErrorPermissionDenied`+内部 reason(`NOT_BACKEND_PLANE`/`NOT_MEMBER_PLANE`,镜像 PlatformGuard 的 rebind 先例),不走 proto 新码——纵深防御层拒因,客户端语义=403;登录步的诚实分流由既有 SEAT_REQUIRED/USER_DISABLED 承担(PG-2)。
- **PR-8 BanGuard 搁置令**(2026-10-21,用户裁定):esports 会员封禁(曾起草的 member/BanGuard+accessLadder 第三中间件参数形状作废)待本前沿收口后重新设计;演进期间只登记考量不实现——①业务名望执法不得以附加参数穿 accessLadder 通用签名(ladder 是骨架件,业务执法归业务服务自己的面与模块,这正是"业务无关但可拓展"的接缝形状);②座位级名望(R-1)是其骨架级触发器;③能力级封禁归业务主权(总则 4)。**勘误(同日 PG-3 撤案)**:PR-6 所记"agent BUILTIN 唯一消费者是 admin 监控页"为误标——chat×6/agent_config×3 全部是 `/me/` 第一人称面,admin 侧消费是其内嵌聊天工作区与个人中心"我的连接"(运营者自己的使用);跨用户监控本来就在 `AgentConfigAdminService`(ADMIN 档)。
- **PR-9 cookie 键命名空间**(2026-10-21,用户裁定):单 host 多客户端的真问题是凭据**并存**而非隔离(平面门已在服务端执法)——机制形:cookie 键带客户端命名空间(`session_token_<ns>`,ns 与客户端门牌 VITE_POPULATION 同源部署接线),transport 随请求携带命名空间声明;声明=**不受信的凭据选择,非授权**(平面永远由会话内已证明座位推导,选错命名空间只是拿错自己的牌)。分域部署仍是强默认(host-only cookie 天然隔离),命名空间使单域产品形态(example.com/admin+主站)成为一等公民,且解除 R-3。落片=PG-6(同日落地)。
- **PR-10 cookie 名单源形状**(2026-10-21,用户质询后定案):cookie 名是**线契约**,按语言单源——TS 侧组合规则只写一次在 shared-transport(`sessionCookieNameFor`+`SESSION_NAMESPACE_HEADER` 导出,custody 只剩接线,第三客户端零拷贝);Go 侧单源在 krauth。**否决全后端单源**(接口下发名字/Set-Cookie 转发):①cookie 属性是边缘所有(Secure 由浏览器侧 `x-forwarded-proto` 判定,system 在内网不可见);②后端 Set-Cookie 需边缘改写属性→被迫解析名字,知识回流;③登录路径新增配置 RPC+启动失败模式,只省一行组合。跨语言一致性=每语言锚点测试+live 链路执法(登录→槽位 cookie 出现→认证)。
- **PR-11 agent BUILTIN=双面 me 面**(2026-10-21,PG-3 撤案时定案):平面门法第二款的实例——一个梯的服务对象是"两世界各自的第一人称面"时保持只验会话:agent 的 chat×6/agent_config×3 全部 `/me/` 主体作用域(运营者内嵌聊天+未来 C 端会员聊天共享),挂 member 门会杀 admin 自己的聊天组件;跨用户监控已住 ADMIN 档(`AgentConfigAdminService`)。判别式:**面的路径语义是 `/me/`(第一人称)还是产品本体**——前者双面,后者归其世界(esports BUILTIN=会员产品本体→挂门;agent chat=两世界共用能力→不挂)。

## 3) 边界(明确不做)

单页面多身份账号选择器(单 host 单 cookie 已覆盖);第三平面词(触发制);esports 本地封禁/能力级资格位(PR-8 搁置:待本前沿收口后重新设计,演进期间只登记考量不实现);重写 population 数据模型或拆双表(单核+座位是已验证路线,修法仅接线——修 bug 的代价即路线正确性的度量)。

## 4) 演进后残余边界(本前沿全部落地后仍不支持的情景)

2026-10-21 应用户要求推演定案。此表防"演进完成=机制无所不能"的过度自信;每条都是**设计如此或触发制**,不是缺陷清单——出现真实触发时按新前沿立项,不回灌本档。

| # | 情景 | 演进后行为 | 缺什么 | 处置 |
|---|---|---|---|---|
| R-1 座位级停用 | 封运营身份但保留会员身份(或反之) | 名望三层全部作用在 user 账号级,UpdateUserStatus 一封双面连坐——本演进的原点场景在骨架级依然只能账号级 | seat 行的名望位(禁用/撤销) | 触发制;esports 本地资格位是业务侧近似解,双身份封禁的骨架级需求出现时立项 |
| R-2 后台多座选择 | 运营属于两个后台人口/租户(平台+MCN) | PG-2 后 admin 登录带部署配置的后台人口码,双后台座用户登录即 SEAT_REQUIRED 拒 | 同平面内的多座选择 UI(比账号选择器轻:同是 backend 面,选人口=选数据视图) | 触发制(esports 单后台人口不触发) |
| R-3 同 host 双平面并发 | 后台内嵌 C 端预览等产品形态 | 单 host 单 cookie,两平面会话互踢;分域名天然并存(已验证) | 会话键分平面或子域名 cookie 作用域 | **PG-6 解除**(PR-9 cookie 键命名空间;2026-10-21 已实证互踩:会员会话直入 admin 工作台) |
| R-4 迁座即时降权 | 人口 A→B 的立即生效 | 会话 Population 在 mint 时定格,迁座后旧会话持旧平面直到自然过期或手动 kick | 迁座触发按 user 撤销存量会话(revoke-all 语义已有,缺触发器) | 小增量,真实需求出现时挂入验证矩阵 |
| R-5 第三平面 | 设备面(IoT/机器狗)、服务身份面 | 词表封闭=编译期两词,第三词是骨架级决策 | PlaneGuard 家族+人口 plane 列词表同步扩词 | 触发制(kratos §5a 已记 service-principal 预留位) |
| R-6 业务资格位原语 | 封直播权/禁赛/禁止被关注 | 骨架不提供,esports 自建本地资格位 | 各业务重复实现 | 设计如此(正交性代价),不修 |
| R-7 服务间内网信任 | 内网内服务互调冒用 | 平面门管"用户牌走错门",不管服务层 | 服务身份(service token/mTLS) | 部署面前沿,R-5 的姊妹触发器 |
