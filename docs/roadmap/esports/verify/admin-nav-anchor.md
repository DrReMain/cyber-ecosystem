# admin 菜单锚 — 原型九组真值 × 现行树落差(单一真值)

> 建立:2026-10-30(用户令:admin 菜单规划/菜单名须高度还原原型——属 D-AQ admin 标准的「功能规划+文案」维度;视觉仍走 antd 自由)。
> 真值源:`lab/esports/manage/design-systems/admin/outputs/prototype/_nav.js` CFG 块(机器读取,九组 33 项含 5 非本期);现行=路由树 staticData 提取(`app/clients/admin/src/routes/dashboard/**`+messages zh-CN 解键,2026-10-30 树)。
> 消费:M2-1 起所有 admin 切片的菜单落位与命名以此为准;骨架自有面(非原型账)保留;**IAM 五面=骨架直接承接不改名(D-AY,2026-10-30 用户裁定)**。

## §一 原型九组(真值,_nav.js CFG 逐字)

| # | 组 | 项(名 → 页键) | 非本期 |
|---|---|---|---|
| 1 | 仪表盘 | 运营总览→dashboard;用户画像→user-portrait | |
| 2 | 用户与会员 | 用户列表→user-list;会员管理→member-list;公告管理→announcement | |
| 3 | 题库与测试 | 题库管理→question-bank;测试类型→test-type;测试记录→test-record;游戏与因子配置→game-config;游戏归类→game-category | |
| 4 | 定价与订单 | SKU 管理→sku-manage;订单与解锁记录→order-record;支付渠道管理→payment-channel | |
| 5 | ToB 租户 | 租户与机构→tob-tenant-list | |
| 6 | 报告与内容 | 报告配置→report-config | ✔ |
| 7 | 渠道与推广 | 激活码管理→activation-code;渠道管理→channel-manage;落地页与素材→landing-material;归因分析→attribution | ✔×3 |
| 8 | 反馈与客服 | 反馈管理→feedback-manage | |
| 9 | 系统与合规 | **操作审计▮、角色管理▮、权限管理▮、管理员账号▮、菜单管理▮**(▮=D-AY 骨架直接承接,esports 不动);合规配置→compliance-config;协议版本→agreement-manage;同意记录→agreement-consent;评级管理→rating-manage;系统配置→system-config;App 版本→app-version | ▮×5 |

## §二 现行树(2026-10-30 机提)

| 组(order) | 叶(order) | 页面键 |
|---|---|---|
| 工作台(0) | — | agents-chat 骨架面 |
| 平台管理(300) | 租户管理(10);用户类型(20) | platform 骨架面 |
| 智能体(90,group) | 连接(10) | agents 骨架面 |
| 文件管理(200) | 文件列表(10) | files 骨架面 |
| 系统管理(100,divider) | 用户管理;角色管理;部门管理;策略管理;权限诊断;审计日志 | system 骨架面 |
| **电竞(60,divider)** | 激活码(10);会员(10,=users);报告访问策略(20);测试项(30) | esports 面;(users/$id「会员详情」hide) |

## §三 差表(处置)

| # | 落差 | 原型真值 | 现行 | 处置 |
|---|---|---|---|---|
| N-1 | **组结构整体错位** | 九组按业务域(用户与会员/题库与测试/定价与订单/渠道与推广/反馈与客服/系统与合规…) | esports 四叶全塞单组「电竞」;原型无此组 | **随 M2-1 重构 esports 自有组**(用户与会员/题库与测试/定价与订单/渠道与推广/反馈与客服+报告与内容/仪表盘落位);**系统与合规组=骨架 system 组直接承接(D-AY),esports 不重构不改名**;骨架自有面(工作台/平台/智能体/文件)保留原位 |
| N-2 | 命名族错(限 esports 自有面) | 用户列表/激活码管理/游戏与因子配置 | 会员/激活码/测试项 | 随 M2-1 改键(G-302);操作审计/管理员账号等 IAM 名**不在列**(D-AY 骨架原名自治) |
| N-3 | **概念错位** | 01-user-list=用户列表+01-user-detail=用户详情(用户档案);02-member-list=会员管理+02-member-detail=会员详情(会员域)两域四页 | 现行 users 页名「会员」、$id 名「会员详情」且消费 MemberService.getMember | M2-1 落会员两页时正名:users 族=用户列表/用户详情;member 族=会员管理/会员详情;MemberService 消费归属随 M2-1 契约面核 |
| N-4 | 组名 | 系统与合规 | 系统管理 | 随 N-1 同批 |
| N-5 | 未建组项 | 公告/反馈/SKU/订单/支付渠道/题库管理/测试类型/测试记录/游戏归类/合规/协议/同意/评级/系统配置/App 版本 | — | 各归 M2-2/M2-3/M2-4/E9/MSG/余组/P 票,建时按 §一 落位落名;菜单管理=骨架已有(D-AY) |

计数:原型 9 组 33 项(5 非本期);现行 esports 相关落差=N-1..N-5 五族;骨架自有面 5 组不计原型账(保留)。
