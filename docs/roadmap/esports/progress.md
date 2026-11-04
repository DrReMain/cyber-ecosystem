# esports 进度档 — 认领状态与验收记录

> 归属:esports 前沿进度档,三人共写。**剪枝律**(全文=README §5):片收口即剪枝为一行收口注(状态+日期+hash+关键判定);节奏行=索引形 ≤150 字;活档区(快照/认领表/遗留清单)=唯一可变区;历史压实迁 progress-archive.md;里程碑收口做活档漂移对账。
> **2026-11-05 瘦身役**:10-31..11-05 已收池(M2 核心组/E9/R 线/Q 线/MIG1 五阶段/MSG-1 首刀/PAY3 四片)全量压实迁 archive;需求分母与页面处置=verify/demand-ledger.md。

## 现状快照(2026-11-05,W3 中)

| 人 | 一行定位(详单=archive) | 当前 → 下一片 |
|---|---|---|
| lzx | M2 核心组+E9 全收(10-31..11-01);MIG1 五阶段毕(冻结待 P26/P26.1);Q 队列①-⑩清(11-03..04);PAY3 四片全收双推(11-04..05,末笔 2c3ecb7c);MSG-1 ctr 腿收(11-05 未提交) | MSG-1 admin 三页(W4)→PAY2 联调/真通道刀先行件→MIG1 预演 11-15→CP3 核验 11-28→cutover |
| hzy | R 线五片+PC/PM 批+消费片 ✅(11-01..02);Q1-Q6 全毕(11-03..04,MR FF 合入);W6 值守备料 ✅(11-04 cutover-watch 草案) | 待分配:W4-W5 上线走查执行稿修订+值守演练 |
| ljr | P7 定稿(11-02);P21/P19/P22/PM-12/同意口径已裁(D-BC);P6 已代裁定档(D-BE 预演 11-15) | 行政件 11-10→P22 否决窗 11-14→P26/P26.1(理想 11-12/硬限 11-24)→appid+绑域 11-26→P23/P20 11-28;P4 11 月底 |

## 活片认领表(定义=plan.md;已收片迁 archive)

| 片 | 认领 | 状态 | 一行注 |
|---|---|---|---|
| MSG-1 余面 | lzx(D-BB 全片) | ctr 腿收口(11-05,未提交):四路由(messages 树 2+feedback 树 2)+未读双红点(tabbar/home bell)+G-202 真跳;全态 live 走查+en 面板毕;净清毕 | 余=admin 三页+V10 闭环走查(W4)+窄化裁决接洽(两裁决+R-08,W6 ljr) |
| PAY2 通道联调 | lzx | W4(行政件门 11-10) | deploy 切 wechat+0.01 真钱闭环+全额退款(CP3 门);H5 scene 随权限;APIv3 密钥已核对 |
| 真通道刀(G-402/403/404) | lzx | G-403/404 先行件可开工;G-402 建议裁②报备中 | 详见 gap-ledger 五批;真通道三红=JSAPI openid/QR 码图/H5 回程 |
| MIG1 迁移工具 | lzx(专职 D-BB) | 五阶段毕;**预演 11-15**(D-BE,绿门含回滚腿) | snapshot/recompute 冻结待 P26/P26.1;口径=migration.md;cutover 执行同专职 |
| M2-4 trace/messages Tab | 递延 | trace=暂缓后置(11-03 用户裁,P22 否决窗 11-14 后复评);messages=随 MSG-1 admin 面 | — |

## 开放遗留清单(完成即删行)

| 项 | 归属 | 触发/时点 |
|---|---|---|
| 微信回调 next 丢失(redirect_uri 不能带 query,须 state 中转) | 真通道刀 | 设计时定 |
| 上线前登录切真实公众号 appid(JSAPI 绑域;测试号绑不了商户) | PAY2 | 上线清单 |
| ICP 串上线前 beian.miit.gov.cn 复核一次(D-AO① 维持带「备」) | 上线清单 | cutover 前 |
| G-202 反馈 CTA 门控占位(soon toast) | MSG-1 余面 | 本片接真跳 |
| G-208 余族 share-invite 腿(pay-confirm 腿已随 PAY3 片2 落,oracle=P7 稿) | MSG-2 | 落地时先验 oracle 形状 |
| 报告页 ④⑤ CTA live 走查(裸记录过不了 getReport 判分面;导航形与 profile 卡同构已活证) | lzx | 随真通道刀/激活码走查复核 |
| age-required 页本体逃生口(「先看免费报告」;墙侧未成年红卡逃生已随片2 落=minor-card) | lzx | 随 MSG-1 余面或走查批 |
| G-203 多凭证掩码取序=已裁保持现状(最近绑定序) | — | 运营反馈再议 |
| G-304 开通 Modal 接用户搜索选择器(列表→开通需经详情转手) | lzx | 下轮会员面增量裁 |
| G-305 会员列表全表 fold 下推(万行/预演触发器) | lzx | MIG1 预演时评估 |
| G-306 订单/会员手机号搜索(需 system 名册反查投影) | lzx | 运营反馈高频时 system 车道立契约 |
| consent 渠道列缺契约字段(Q5 登记;可并 P23 归因记账一句裁) | lzx | P23 裁时 |
| 合规配置提示文案无「恢复默认」手势(契约 min_len1 挡空串) | lzx | ljr 提出时契约小补 |
| 出生采集开关消费面未接(ctr 向导仍恒问出生年月;付费年龄线/文案执法已随 PAY3 落) | lzx | 注册面向导片/消费端立契约时 |
| 同意记录同版重复登录追加行=已裁维持每次记(D-BC);admin 展示去重未排 | — | 运营反馈再议 |
| dev 常模池 fixture 机制未立(幻影池 boot 种子=生产污染硬否;percentile 列消费端落地片再议) | lzx | 消费端落地片 |
| mt-D5 unitScoreLabel 种子(介绍文半边已随 Q1-⑥ 落;单位标签仍空隐) | lzx | 随种子/国际化面 |
| skill_trajectory 30s 走查体感偏紧 | lzx | 评估加时 |
| 组内 spec doc 存量约 40 处行尾化 | lzx | 触碰 test 模块顺路或独立机械片 |
| conventions guard 正则精度修正回流 dev 兄弟服务(esports 已 AST 化) | lzx | dev 车道下次触碰同步 |
| dev 栈老库接 atlas 需基线(无 revisions 表先 `atlas migrate set` 打基线;专用 schema `atlas_schema_revisions`) | lzx | migrate:baseline 补 nx target 或本行即文档 |
| dev 栈状态目录 .run/ 迁仓外(D-AX,`$TMPDIR/cyber-dev/<project>.json`) | 各机 coding agent | 下次开栈前 |
