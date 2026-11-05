# esports 进度档 — 认领状态与验收记录

> 归属:esports 前沿进度档,三人共写。**剪枝律**(全文=README §5):片收口即剪枝为一行收口注(状态+日期+hash+关键判定);节奏行=索引形 ≤150 字;活档区(快照/认领表/遗留清单)=唯一可变区;历史压实迁 progress-archive.md;里程碑收口做活档漂移对账。
> **2026-11-05 瘦身役**:10-31..11-05 已收池(M2 核心组/E9/R 线/Q 线/MIG1 五阶段/MSG-1 首刀/PAY3 四片)全量压实迁 archive;需求分母与页面处置=verify/demand-ledger.md。

## 现状快照(2026-11-05,W3 中)

| 人 | 一行定位(详单=archive) | 当前 → 下一片 |
|---|---|---|
| lzx | M2 核心组+E9 全收(10-31..11-01);MIG1 五阶段毕(冻结待 P26/P26.1);Q 队列①-⑩清(11-03..04);PAY3 四片全收双推(11-04..05,末笔 2c3ecb7c);MSG-1 ctr 腿+G-403/404 先行件收+admin 三页 V10 毕(11-05,提交 5d3bded5/7b267589);PAY2 真通道双腿毕(11-05..06:首探+NO_STATEMENT_EXIST 修 3fe2cc3f+0.01 闭环+退款腿活证);MIG1 备料毕(11-06:生产副本直载+谱系三源互证+census 三绿,预演前两决策 D-MIG-1/2 待裁) | MIG1 预演 11-15(先过 D-MIG-1/2)→CP3 核验 11-28→cutover |
| hzy | R 线五片+PC/PM 批+消费片 ✅(11-01..02);Q1-Q6 全毕(11-03..04,MR FF 合入);W6 值守备料 ✅(11-04 cutover-watch 草案) | 待分配:W4-W5 上线走查执行稿修订+值守演练 |
| ljr | P7 定稿(11-02);P21/P19/P22/PM-12/同意口径已裁(D-BC);P6 已代裁定档(D-BE 预演 11-15);H5 递延+Native 已开通(D-BF);经营类目降级(11-05,只读,结算时再改) | 行政件清空→P22 否决窗 11-14→P26/P26.1(理想 11-12/硬限 11-24)→appid+绑域 11-26→P23/P20 11-28;P4 11 月底 |

## 活片认领表(定义=plan.md;已收片迁 archive)

| 片 | 认领 | 状态 | 一行注 |
|---|---|---|---|
| MSG-1 余面 | lzx(D-BB 全片) | **admin 三页+V10 闭环毕(11-05)**:反馈管理(状态段+快捷回复+关闭)+反馈详情(KV/对话气泡/contact)+公告管理(两步发布/编辑草稿/下线);V10 全环 live 走查毕(ctr 提交→admin 双路回复→追问回待处理→关闭终态;全员信 11 封/会员型空投/banner 零投递 DB 侧证);走查捕五伤全修(Form.Item 双子断绑定/form store 残值/枚举名比对/两处分型 confirm 文案);窄化 R-09..R-13 已登记 | 余=窄化两裁决接洽(两裁决+R-08..R-13,W6 ljr);banner ctr 消费面见遗留清单 |
| PAY2 通道联调 | lzx | W4(**行政件门已清空 11-05**);材料全齐活证(APP_ 前缀引号形 boot 绿+对账 sweep 获微信真应答);**Native 腿全生命周期活证(11-05 晚,dev 栈真 UI)**:create(真 code_url 落库+码图 sheet 渲染)→query(商户直签 GET,微信答 NOTPAY,应答签名平台公钥验过)→close(259ms 真往返,双侧 CLOSED);同单复用回放(零通道往返)与 notify 拒绝路径(ctr server route→服务 400 INVALID_SIGNATURE)亦证;**0.01 真钱支付腿毕(11-05 深夜)**:member_probe 1 分 SKU 真扫码支付→直签查单亲证 SUCCESS(22:58:29,transaction_id 4500…1881,应答验签过)→收敛环 23:07:44 翻 paid(paid_at=通道 success_time 分秒差零)→membership 权益 1 天窗(paid_at+1d)→UI 会员面「已开通」翻转;notify 不可达场景=收敛环兜底实证(10 分钟 stale 线后 26 秒拾起);探针 SKU 用后即下架;deploy 事实=localhost notify_url create 侧照收/v3 token=mchid 非 m | **退款腿毕(11-06 深夜,W5 件拉前)**:RefundPaymentOrder 出口+pay_order 派生事实六列(paid 不翻状态,RefundPaid/SettleRefund 双仲裁写)+admin 详情退款块/列表双标签/原因必填弹窗;0.01 标本真退款六段活证=微信答 PROCESSING(refund_id 503…8181,channel=ORIGINAL 原路)→刷新读穿收敛 succeeded(refunded_at=通道 success_time 分秒差零)→支付方零钱通到账亲证→直签退款单 SUCCESS+支付单 trade_state=REFUND 双验签过;e2e 退款/重退拒绝/错态拒绝/读穿收敛四臂绿;随片修 reconcile 单账单对全量→按 provider 分组对账(dev 库进多通道时代实弹暴露);余=JSAPI 腿随 G-402 裁②+真实 appid(上线清单)+微信外降级「引导微信内打开」(D-BF) |
| 真通道刀(G-402/403/404) | lzx | G-403/404 先行件收(11-05):qrcode 码图 tile+sheet visual 槽;returnUrl 三层修形(客户端收据 URL/biz 缝 order/网关 redirect_url 镜像);G-402 建议裁②报备中;**NO_STATEMENT_EXIST 误分类已修(11-05 晚,真通道首探发现:活通道「无账单」拼写落 ErrNotFound,boot sweep INFO 命中;shared-go/payment/wechat 双拼写修随本片落)** | 唯微信侧活证+G-402+信任拨盘=真通道联调刀(W5);query 无面层触发器(convergeStaleAfter=10m 静默;11-05 深夜 0.01 闭环实证=收敛环 10 分钟 stale 线后 26 秒拾起,notify 不可达场景兜底成立)与 boot 放弃日 ERROR 噪声两裁点随刀;详单=gap-ledger 五批 |
| MIG1 迁移工具 | lzx(专职 D-BB) | 五阶段毕;**备料毕 11-06**:生产副本直载(lab mysql8:13306,106MB)+谱系三源互证(master=生产 V15 世代/block34 未发布/转储=唯一真相,init.sql=CI 残缺基线不可信)+census 真数据首跑(guard 必停 school 4/463 含 7 真实用户 22 行+账号冲突 6 组 23 行=**预演前两决策 D-MIG-1/2**,详 migration.md 预演备料节);随片修 Nickname NullString+test_record 宽行排序(MySQL 1038);migrator 临时 nx target run/check 落地(go 直执禁) | snapshot/recompute 冻结待 P26/P26.1;**预演 11-15**(D-BE,绿门含回滚腿,先过 D-MIG-1/2);口径=migration.md;cutover 执行同专职 |
| M2-4 trace/messages Tab | 递延 | trace=暂缓后置(11-03 用户裁,P22 否决窗 11-14 后复评);messages=随 MSG-1 admin 面 | — |

## 开放遗留清单(完成即删行)

| 项 | 归属 | 触发/时点 |
|---|---|---|
| 微信回调 next 丢失(redirect_uri 不能带 query,须 state 中转) | 真通道刀 | 设计时定 |
| 上线前登录切真实公众号 appid(JSAPI 绑域;测试号绑不了商户) | PAY2 | 上线清单 |
| H5 支付重审(D-BF 递延:新版部署上线、有真实付费界面后再提;过审前微信外=降级「引导微信内打开完成支付」prd P7.2) | PAY2 | 新版部署上线后即提 |
| ICP 串上线前 beian.miit.gov.cn 复核一次(D-AO① 维持带「备」) | 上线清单 | cutover 前 |
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
| banner 型公告 ctr 首页消费面未建(ListBanners 契约在、无消费方;admin 发布+DB 空投递已活证 11-05) | lzx | MSG-2/精修带评估 |
| 同意记录同版重复登录追加行=已裁维持每次记(D-BC);admin 展示去重未排 | — | 运营反馈再议 |
| dev 常模池 fixture 机制未立(幻影池 boot 种子=生产污染硬否;percentile 列消费端落地片再议) | lzx | 消费端落地片 |
| mt-D5 unitScoreLabel 种子(介绍文半边已随 Q1-⑥ 落;单位标签仍空隐) | lzx | 随种子/国际化面 |
| skill_trajectory 30s 走查体感偏紧 | lzx | 评估加时 |
| 组内 spec doc 存量约 40 处行尾化 | lzx | 触碰 test 模块顺路或独立机械片 |
| conventions guard 正则精度修正回流 dev 兄弟服务(esports 已 AST 化) | lzx | dev 车道下次触碰同步 |
| dev 栈老库接 atlas 需基线(无 revisions 表先 `atlas migrate set` 打基线;专用 schema `atlas_schema_revisions`) | lzx | migrate:baseline 补 nx target 或本行即文档 |
| dev 栈状态目录 .run/ 迁仓外(D-AX,`$TMPDIR/cyber-dev/<project>.json`) | 各机 coding agent | 下次开栈前 |
| trial 单退款后解锁回收无操作面工具(会员侧回收=entitleadmin 终止组合即得;FAQ 口径言「退款单回收对应解锁」) | lzx | 运营需要时裁 |
