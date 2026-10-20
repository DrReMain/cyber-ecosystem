# plane 开发计划 — 切片池

> 归属:plane 前沿开发档。总则/裁决台账见 README;状态记 progress.md。
> 规模:S(≤1 天)/ M(2-4 天)。依赖:PG-1 → {PG-2, PG-3}(可并行) → PG-4 → PG-5。

---

| 片 | 内容与验收要点 | 依赖 | 规模 |
|---|---|---|---|
| PG-1 平面门服务端片 | shared-go `PlaneGuard`(镜像 PlatformGuard 形状:ADMIN 面验 `Subject.Plane == backend`;PR-2:BUILTIN 面验 member)挂入 system/agent/esports 三服务对应链;esports client 补 Plane 映射(缺口③);kratos CONVENTIONS 增补法条"ADMIN 链必含 PlaneGuard"+复制 ladder 的机器守卫(防回归);如需新错误码走 proto 增量(五语言)。验收:反例探针——member 牌打 ADMIN 面拒、backend 牌打 BUILTIN 面拒、正常双面绿;三服务链全绿 | — | M |
| PG-2 双客户端接线片 | admin/ctr 登录传 population(admin→其后台人口码;ctr→member;缺口①);登录错误分流:SEAT_REQUIRED→"该账号无此端访问资格"+换账号动作、USER_DISABLED→"账号已被停用"+申诉路径,客户端按 reason 分支禁吞(PR-3;admin 五语言/ctr zh+en);dev 双主机名说明落地(PR-4)。验收:会员凭据登 admin=登录步拒且文案正确;后台无座账号登 ctr 同构;语言链绿 | PG-1 | S |
| PG-3 双服务消费片 | agent 服务同构接线(client Plane 映射+两面校验,esports 已随 PG-1 落)。验收:agent 链反例探针双拒;esports 既有面回归绿 | PG-1 | S |
| PG-4 真实场景验证片 | esports 实测矩阵全绿:微信会员链路回归(登录/落座 member);封禁真路传导(UpdateUserStatus→ctr 下次探针即踢,双端文案正确);双世界并存(dev 分主机名,admin+ctr 同时在线互不登出);跨面偷牌(反例);kick 语义(撤销标记只杀旧会话)。可脚本化的落 e2e | PG-2, PG-3 | S |
| PG-5 镜像迁移片 | 骨架件经 worktree 回 dev(shared-go PlaneGuard/system 接线/CONVENTIONS 法条/两 client 骨架面);esports 树只留业务增量;dev 侧全链绿;本档 destination 预演(法条回灌位确认) | PG-4 | S-M |
