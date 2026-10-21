# ctr-shell 前沿 — C 端壳机制(会话真值/双态面/首屏预取)

> 状态:活跃前沿,2026-10-21 立项并当日落片。源于用户报告的两个首屏症状(已登录刷新首页先匿名后登录;"我的"刷新出现 loading),深研原型/PRD/审计后定性为**壳移植缺脊柱**:admin 的守卫墙模型被搬来,但 ctr 的双态面世界需要根级 SSR 会话真值——这根脊柱整体缺位。
> destination: 收口时 durable 法条已回灌 tanstack CONVENTIONS §3(First-paint data + T-RTE-5);本档为前沿索引——机制定性与依据存档在此,切片与进度在子档。

---

## 0) 机制定性(2026-10-21 实证)

- **脱水契约**(读穿 @tanstack/react-router-ssr-query 源码):只有被 await 的查询进脱水载荷;`useQuery` 在 SSR 渲染中点火不等待,流关即丢,客户端必然补取。
- **实证**(持会员 cookie 抓 SSR HTML):修复前首页=匿名脸+会话不进载荷;`/profile`=LOADING+画像不进载荷。修复后双双翻转。
- **admin 无此问题的原因**=`/dashboard` 布局 `beforeLoad: sessionGuard` 在 SSR 期解析会话;ctr 移植时守卫只挂墙内页(/profile、/match 树),公共双态面无根级真值。
- **潜伏雷两枚**(本片顺带根除):①向导 updateProfile 成功后无任何失效→60s 内 profileFlowGuard 读陈旧"资料不完整"裁决弹回 nickname;②connect-query 缓存键对每个 transport 实例唯一(t1/t2…),跨端不稳定——loader 预取必须用静态键 `connectTransport` 单例,`createTransport()` 会静默分裂缓存。

## 1) 原型与需求的深度依据(逐点根据存档)

- **双态首页=产品拍板**,非实现自由发挥:LOGIC_AUDIT B-⑪(未登录允许浏览首页,点"开始测评"才拦,`/match/** /report/**` 才需认证)+A-⑤(拦下跳登录+next 回跳双写)+2.5-⑨(首页定稿=原型 home,含续测卡/最近成绩/统计/报告入口)。
- **守卫三档全谱**(44 页全量声明清点):guard:null×8/login×20/profile×12/adult×1;TabBar 挂 7 页且守卫异构(home 无墙/me=login/rec=login+profile)。
- **原型与现仓的完整性判据为契约驱动分叉,防误对齐**:原型 2.5b-⑭=birthDate∨ageBand 二选一+sex 必选;现仓=birth 对(esports v1 契约无年龄段字段,P8 前无 sex,guard.ts 注释在案)。
- **渠道版=第三准入世界**(scheme-b×3 guard:null;isLogged=登录∨渠道token;L2-A-⑦ 报告先挂租户/学员后认领;P4 ToB 裁决 11 月底):会话语义必须保持 status 词表(anonymous|authed|expired|unreachable,未来加渠道态),不加布尔不写死两态。
- **首页数据面会长大**(定稿 home 的已登录模块依赖 records/member):预取法=按 block3 PRD §7 页面×数据依赖矩阵逐页生长的通用法,本片只落现在唯一建成的内容页(我的 overview)。
- **优先级背书**:esports README 裁决 D-V——"接通层(骨架/cookie/会话/水合)才是长命资产,业务代码随需求反复重写"。

## 2) 壳机制五法(已落;法条本体=tanstack §3 First-paint data)

1. **会话真值法**:根路由 beforeLoad 解析探针(`query()`+吞非会话错),脱水携带;60s staleTime 内客户端零重取。
2. **双态面法**:首页以会话内嵌画像为唯一身份源(探针以 getProfile 证凭据,画像搭车),不另立查询。
3. **守卫分层法**:sessionGuard(/profile 树)/profileFlowGuard(/match 树)挂布局 beforeLoad 读同一缓存=零额外 RPC;adult 档挂位 PAY3(age-required 页已建)。
4. **内容页预取法**:route loader 预取页面数据(connectTransport 静态键单例;authed 门控+容错,数据失败回落页面自有客户端取数)。
5. **写路径失效律**:useProfileUpdate 单缝(updateProfile 成功即 invalidate 会话查询)。

## 3) 残余边界(设计如此或触发制)

- 渠道版准入态:P4 裁决后以 status 词表加态,非重构。
- 版本更新全局检查(原型 guard:null 页):未来 root 挂点(root loader 已有 getStoreCookies 先例)。
- 报告/记录/消息/付费页的首屏预取:随 E5/E2/E8/PAY 片按 §7 依赖矩阵逐页接。
- admin 不回改(无公共面,守卫墙模型即其正确形状);plane PG-5 镜像时评估五法是否泛化进 dev 骨架(第二消费者=未来任何 C 端客户端)。

## 子档地图

| 档 | 内容 | Owner |
|---|---|---|
| progress.md | 节奏记录 + 验证存档 + 测试工具教训 | 开发共同维护 |
