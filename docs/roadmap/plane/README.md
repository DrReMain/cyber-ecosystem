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
| authz 分支 | `IsMemberPlane` → member 平面不读 grants(结构性零权限) | `authz/data.go:105` |
| PLATFORM 档 | `PlatformGuard` 默认租户密封(第三档已建) | `security/platform_guard.go` |
| 登录方式落座 | provider 行 → `landingPopulation`(微信→member,密码缺省→后台人口) | `auth/biz.go:454` |
| 会话隔离 | `session_token` host-only cookie → 分域名即双世界并存 | `custody.server.ts` |

**三缺口(全部是接线,不是机制)**:① admin/ctr 登录调用均不传 population——门在,没人报名字;② ADMIN 链无平面门,靠 grants 兜底——member 凭据可认证入后台壳(零权限泄露面);③ esports 的 `AuthenticateSession` 组装 Subject 时丢弃 wire 上已有的 Plane 字段(`esports/client/system.go:66`)。

## 1) 总则(每片设计前对照)

1. **两轴正交**(PR-1):population=内聚轴(分组/数据视图,业务数据行,自由加);plane=安全轴(门的种类,编译期两词 `backend|member`)。谁也别替谁干活——内聚轴不得做安全裁决,安全轴不得做业务分组。
2. **信任模型**:客户端只报门牌(population hint,服务端座位表校验,报无座门牌即拒);平面永远由服务端从已证明座位推导,不来自任何请求输入。三道门全在服务端:发牌门(resolveSeat)/每请求门(ladder 验 Plane)/权限门(authz)。
3. **词表封闭**:平面词表=编译期两词,第三词是骨架级决策(触发制:真实出现新门种类);人口=数据行,业务自由增加,system 零代码改动。判据:**这段代码在零业务的纯骨架里是否一模一样存在**——是则本档,否则退回业务前沿。
4. **拒绝深度律**(UX 法):拒绝发生越早,UI 面积越小——登录步拒绝(一句文案+换号按钮)优先于空壳后台(整棵零权限页面树);能力级封禁(抖音式封直播权)归业务面完整表达,是业务主权,与本档正交。
5. **错误诚实分层**:密码错≠无座≠停用≠踢线≠能力封禁,各有独立文案与下一步动作;客户端禁吞 reason 一律映射"账号或密码不正确"。
6. **交付模式**(PR-5):esports 分支组先行落地,esports 真实场景验证,镜像迁移回 dev 骨架侧是**显式切片**(PG-5),不是口头承诺。

## 2) 裁决台账

- **PR-1 两轴正交**(2026-10-21):population=内聚、plane=安全,两轴分离执法;豆包分析经代码核实采纳,其"Plane 已在 subject/PlatformGuard 已存在"断言逐条属实。
- **PR-2 BUILTIN 面验 member 平面**(2026-10-21):采纳——纵深防御+对称隔离(后台账号无 member 座即进不了 C 端);运营测 C 端=显式授 member 座,可审计可撤销。
- **PR-3 门与门牌同片**(2026-10-21):population 传参与登录错误分流必须同片落地——门建好挂错牌等于没建。
- **PR-4 dev 双主机名**(2026-10-21):cookie 不按端口隔离,localhost:5173/5174 共享 `session_token` 互踩会话(2026-10-20 实测踩过);dev 以 localhost(ctr)+127.0.0.1(admin)分离。
- **PR-5 先行后镜像**(2026-10-21,用户裁定):本前沿在 esports 分支组处理,esports 真实场景验证后镜像迁移回骨架侧;与既有 worktree 纪律的方向相反,是本前沿的显式选择(验证优先)。
- **勘误存档**(2026-10-21):撤回 2026-10-20"esports introspection 不查 enabled"结论——验证链每请求查 Enabled,当日 curl 200 为账号已手动恢复的时间线伪影。

## 3) 边界(明确不做)

单页面多身份账号选择器(单 host 单 cookie 已覆盖);第三平面词(触发制);esports 本地封禁/能力级资格位(esports 前沿独立线程,与本档正交可并行);重写 population 数据模型或拆双表(单核+座位是已验证路线,修法仅接线——修 bug 的代价即路线正确性的度量)。
