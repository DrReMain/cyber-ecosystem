# esports 进度档 — 认领状态与验收记录

> 归属:esports 前沿进度档,三人共写。**剪枝律**(全文=README §5):片收口即剪枝为一行收口注(状态+日期+hash+关键判定);节奏行=索引形 ≤150 字;活档区(快照/认领表/遗留清单)=唯一可变区,收口时**重写而非追加**;历史压实迁 progress-archive.md;里程碑收口做活档漂移对账。
> 2026-11-07 重构役三:本档按剪枝律整体重写(旧快照巨行拆解,叙事全文=git proven-by 尾);走查态清零=verify/integration-matrix 重立,覆盖映射=verify/coverage-map.md。

## 现状快照(2026-11-07,W3 末)

| 人 | 一行定位(详单=archive) | 当前 → 下一片 |
|---|---|---|
| lzx | M2 核心组+E9/Q 线 Q1-Q6/PAY3 四片/MSG-1 全链/PAY2 Native+0.01 真钱+退款腿/MIG1 五阶段+备料+彩排全收(10-31..11-06);走查 R1 轮(X1-1/2、X2、X3、A 套、C1)毕+G 伤单 14 绿;11-07 起=roadmap 重构役三(demand-ledger 重核+八域覆盖核验) | MIG1 预演 11-15(先过 P26/P26.1+D-MIG-1/2)→PAY2 JSAPI 活证→CP3 核验 11-28→cutover |
| hzy | R 线五片+PC/PM 批+消费片毕(11-01..02);Q1-Q6 全毕(11-03..04,MR FF 合入);W6 备料+执行稿修订+九环本地预演毕(11-04..05;预演=T1 门禁四套绿+13 截图+带出 ⑪⑫→B-11/遗留行,详=cutover-watch §六);**U3 admin 基座范式对齐毕(11-07:四列表页 Filter 化+drawer 降级+href locale 锚点族+全 17 面容器审计);U2 片A B-6 done 页三轮毕+片B 保真 harness 重激活毕(11-07,SOP 照抄+常红读数 4 组全在案/等价形)** | 随实况随修 cutover-watch+11-24 复核终版;U2 余=铺开(home/profile/report fidelity spec);W5 值守演练配合(预演已提前);cutover 值守 12-01~03 |
| ljr | P7 定稿(11-02);P21/P19/P22/同意口径已裁(D-BC);P6 已代裁定档(D-BE 预演 11-15);H5 递延+Native 已开通(D-BF);经营类目降级(11-05) | P22 否决窗 11-14→P26/P26.1(理想 11-12/硬限 11-24)→appid+绑域 11-26→P23/P20 11-28;P4 11 月底 |

## 活片认领表(定义=plan.md;已收片迁 archive)

| 片 | 认领 | 状态 | 一行注 |
|---|---|---|---|
| MSG-1 余面 | lzx(D-BB 全片) | 活 | 已毕=契约+ctr 腿+admin 三页+V10 闭环(11-05);余=窄化裁决落地(PM-13..20=prd 待裁区,全维持=零改动,裁翻契约面小改随改) |
| PAY2 通道联调 | lzx | 活 | 已毕=行政件全清+Native 腿全生命周期+0.01 真钱闭环+全额退款腿+NO_STATEMENT_EXIST 修(11-05..06);余=JSAPI 真实 appid+绑域/微信侧活证(G-402/403/404 尾)/微信外降级面/XFF 拨盘收口/两裁点(query 无触发器、boot 放弃日 ERROR) |
| MIG1 冻结件+预演 | lzx(专职 D-BB) | 活 | 已毕=五阶段+生产副本直载+谱系三源互证+census 三绿+回滚腿彩排+密码面实数定谳(11-06);余=snapshot/recompute 冻结待 P26/P26.1;预演 11-15(D-BE 绿门含回滚腿,先过 D-MIG-1/2);cutover 执行同专职 |
| U2 ctr 原型对齐 | hzy | 活 | 片A B-6 三轮+片B harness 重激活毕(11-07);余=铺开(home/profile/report 逐页逐预设 fidelity spec,序=ctr-prototype-matrix)+N-1 趋势 CTA(prd 待裁) |
| M2-4 trace Tab | 递延 | 暂缓 | 11-03 用户裁,P22 否决窗 11-14 后复评(messages Tab 已随 G-413 闭合并销账) |

## 开放遗留清单(完成即删行;切片级任务=plan.md,缺陷=plan 缺陷工单)

| 项 | 归属 | 触发/时点 |
|---|---|---|
| 微信回调 next 丢失(redirect_uri 不能带 query,须 state 中转) | 真通道刀 | 设计时定 |
| 上线前登录切真实公众号 appid(JSAPI 绑域;测试号绑不了商户) | PAY2 | 上线清单 |
| H5 支付重审(D-BF 递延:新版部署上线、有真实付费界面后再提;过审前微信外=降级引导) | PAY2 | 新版部署上线后即提 |
| ICP 串上线前 beian.miit.gov.cn 复核一次(D-AO① 维持带「备」) | 上线清单 | cutover 前 |
| G-208 余族 share-invite 腿(pay-confirm 腿已随 PAY3 片2 落;oracle=P7 稿) | MSG-2 | 落地时先验 oracle 形状 |
| 报告页 ④⑤ CTA live 走查(裸记录过不了 getReport 判分面;导航形已活证) | lzx | 随真通道刀/激活码走查复核 |
| age-required 页本体逃生口(「先看免费报告」;墙侧未成年红卡逃生已随片2 落) | lzx | 随 MSG-1 余面或走查批 |
| G-203 多凭证掩码取序=已裁保持现状(最近绑定序) | — | 运营反馈再议 |
| G-304 开通 Modal 接用户搜索选择器(列表→开通需经详情转手) | lzx | 下轮会员面增量裁 |
| G-305 会员列表全表 fold 下推(万行/预演触发器) | lzx | MIG1 预演时评估 |
| G-306 订单/会员手机号搜索(需 system 名册反查投影) | lzx | 运营反馈高频时 system 车道立契约 |
| consent 渠道列缺契约字段(Q5 登记;可并 P23 归因记账一句裁) | lzx | P23 裁时 |
| 合规配置提示文案无「恢复默认」手势(A5-1 报告策略覆盖行同族) | lzx | ljr 提出时契约小补 |
| 出生采集开关消费面未接(ctr 向导仍恒问出生年月) | lzx | 注册面向导片/消费端立契约时 |
| banner 型公告 ctr 首页消费面未建(ListBanners 契约在、无消费方) | lzx | MSG-2/精修带评估 |
| 同意记录同版重复登录追加行=已裁维持每次记;admin 展示去重未排 | — | 运营反馈再议 |
| dev 常模池 fixture 机制未立(幻影池 boot 种子=生产污染硬否) | lzx | 消费端落地片 |
| mt-D5 unitScoreLabel 种子(单位标签仍空隐) | lzx | 随种子/国际化面 |
| skill_trajectory 30s 走查体感偏紧 | lzx | 评估加时 |
| 组内 spec doc 存量约 40 处行尾化 | lzx | 触碰 test 模块顺路或独立机械片 |
| conventions guard 正则精度修正回流 dev 兄弟服务(esports 已 AST 化) | lzx | dev 车道下次触碰同步 |
| dev 栈老库接 atlas 需基线(无 revisions 表先 `atlas migrate set`) | lzx | migrate:baseline 补 nx target 或本行即文档 |
| dev 栈状态目录 .run/ 迁仓外(D-AX,`$TMPDIR/cyber-dev/<project>.json`) | 各机 coding agent | 下次开栈前 |
| trial 单退款后解锁回收无操作面工具(会员侧回收=entitleadmin 终止组合即得) | lzx | 运营需要时裁 |
| ctr 输码面未建(redeem 无消费方;入口形态属 P4;四拒绝词条已备 locale) | lzx | P4 裁时(11-30)定形随建 |
