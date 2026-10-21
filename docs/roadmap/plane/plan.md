# plane 开发计划 — 切片池

> 归属:plane 前沿开发档。总则/裁决台账见 README;状态记 progress.md。
> 规模:S(≤1 天)/ M(2-4 天)。依赖:PG-1 → {PG-2, PG-3}(可并行) → PG-4 → PG-5。

---

| 片 | 内容与验收要点 | 依赖 | 规模 |
|---|---|---|---|
| PG-1 平面门服务端片 | shared-go `plane_guard.go`(BackendPlaneGuard/MemberPlaneGuard 两导出构造+私有共实现,镜像 PlatformGuard;PR-7:拒绝码 rebind 复用 GeneralErrorPermissionDenied,无 proto 增量);三服务 ladder:ADMIN 梯挂 backend 门(门在 authz 前)、esports BUILTIN 梯挂 member 门(PR-6:agent BUILTIN 移 PG-3,system BUILTIN=双面基础设施结构性免挂);esports client 补 Plane 映射(缺口③,esports 独有——agent client 已映射);kratos CONVENTIONS §6 平面门法条+K-LAD-1;三服务机器守卫 `TestAdminRungPlaneGate`(esports 加 `TestBuiltinRungPlaneGate`)。验收:反例探针(member 牌打 system ADMIN 面拒/backend 牌打 esports BUILTIN 面拒)+双绿路径+三服务 build/test+守卫绿+docs:check | — | M |
| PG-2 双客户端接线片 | admin/ctr 登录传 population(`env.VITE_POPULATION`,admin 默认 user/ctr 默认 member;缺口①);登录裁决为数据跨界(`{status:"ok"|"rejected",reason}`,判别式一律字符串标签——布尔过 server-fn 序列化边界变 truthy 字符串;reason 提取走错误域 normalize——transport 抛信封不抛 ConnectError);错误分流:SEAT_REQUIRED→"该账号无此端访问资格"+换账号动作、USER_DISABLED/LOGIN_FAILED/TENANT_DISABLED 各有文案,禁吞(PR-3;admin 五语言/ctr zh+en);~~dev 双主机名~~(PR-4 被 PR-9 撤销)。验收:浏览器矩阵——会员凭据登 admin=登录步拒+文案/动作正确;admin 凭据登 ctr 同构;双端正路绿 | PG-1 | S |
| PG-6 会话凭据命名空间片(已落地) | 单 host N 客户端凭据**并存**(PR-9/PR-10):Go `krauth.SessionNamespaceHeader`+`SessionCookieNameFor`+SessionAuth 按声明读槽;shared-transport `sessionCookieNameFor`+`SESSION_NAMESPACE_HEADER` 导出(每语言单源)+命名空间拦截器;custody 接线(槽名=`sessionCookieNameFor(VITE_POPULATION)`),ctr logout 手拼 cookie 收编 custody;`WithSessionToken` 服务间路径留基名(无声明=基槽,天然一致)。验收:服务端三探针(错门 401/对门 200 双槽并存/遗留基名 200)+浏览器双世界并存(会员+admin 同 host 同时在线,三槽同头,admin 登录后会员会话存活)+全链绿 | PG-2 | S-M |
| PG-3 agent 面重定级片(**撤案**) | 查证后撤案(PR-11):chat×6/agent_config×3 全为 `/me/` 第一人称面(运营者内嵌聊天+个人中心"我的连接"消费,跨用户监控本就在 AgentConfigAdminService/ADMIN 档)——agent BUILTIN=双面 me 基础设施,按平面门法第二款保持只验会话,不重定级不挂门;PR-6 的"监控面"前提为误标(勘误入台账)。零代码改动,K-LAD-1 守卫形状本就正确 | — | — |
| PG-4 真实场景验证片(矩阵四绿一待) | 封禁真路传导 ✅(UpdateUserStatus→下探即401);双世界并存 ✅(PG-6 三槽实证);跨面偷牌 ✅(PG-1 P2/P3+M7b member牌打ADMIN面403);kick 语义 ✅(F1-F4:封禁即踢/解禁复活新登录/旧牌≤标记永不复活);**微信会员链路回归待用户手机复测**(授权→落座 member→双世界并存下的槽位 cookie) | PG-2, PG-6 | S |
| PG-5 镜像迁移片 | 骨架件经 worktree 回 dev(shared-go PlaneGuard/system 接线/CONVENTIONS 法条/两 client 骨架面);esports 树只留业务增量;dev 侧全链绿;本档 destination 预演(法条回灌位确认) | PG-4 | S-M |
