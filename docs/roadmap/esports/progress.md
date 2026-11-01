# esports 进度档 — 认领状态与验收记录

> 归属:esports 前沿进度档,三人共写。**剪枝律**(2026-10-23 立,10-28 补细则,约束三个 agent;全文见 README §5):片收口即剪枝为一行收口注(状态+日期+hash+关键判定);节奏行=索引形 ≤150 字;活档区(快照/认领表/遗留清单)=唯一可变区;历史随里程碑压实迁 progress-archive.md;里程碑收口做活档漂移对账。
> **2026-10-30 重构役**:已收池(M1 全量+H 线+PAY-C0/C1/PAY1+质量批)压实迁 archive;认领表只载活片;需求分母与页面处置=verify/demand-ledger.md(coverage/矩阵状态列退役)。

## 现状快照(2026-10-30,W2 尾)

| 人 | 一行定位(详单=archive 10-30 压实块) | 当前 → 下一片 |
|---|---|---|
| lzx | M1 全量收满+质量批三件收口(99c5bb06);CP1 提前 14 天;M2-1 收口 347a9ce7+PAY0 收口 4a590622+M2-2 收口 9860eb34+M2-3 收口 45c276d4+M2-4 首刀收口 197fdbc6(11-01) | **M2-4 余半边+E9 增量**(trace 待 P22/messages 待 MSG-1,盯例会);W7 PAY2/PAY3/MSG-1 |
| hzy | E-line-b ✅ 10-30 收口+走查修批五件(=archive 10-30 块) | **R3 余 P10 语言行(mt-F1 已落)**→R1/R2(分配制 D-AV) |
| ljr | P7 视觉稿已落库(1c30b767,精修窗至 12-02);P1/P3/P9/P15/P10/P7 四参数已裁 | **P19/P21/P22(W5)→P14/P23(W6)→P7 定稿/P6(W7-W8)**;P4 11 月底 |

**M1 燃尽**:E0-E10+E5b+E6c 全簇+H 线+质量批 全✅;PAY-C0/C1+PAY1 ✅;PAY2-3 未启动(行政件=用户即办);CP1 ✅ 提前 14 天;CP2 原项全实质达成,**新增门构成(M2 核心组)自 10-30 重构起算**。

## 活片认领表(定义=plan.md;已收片迁 archive)

| 片 | 认领 | 状态 | 一行注 |
|---|---|---|---|
| M2-1 会员终态+运营三操作 | lzx | ✅ 收口 347a9ce7(10-31 走查+双轴审查毕) | 四刀+验收刀全毕:契约/schema(MIG1 口径冻结)/服务端(system 投影+roster 存在性守卫)/客户端(菜单九组 G-301/302 绿+两页+三 Modal 逐字)/e2e 顺延-翻转-解锁窗口-终止回落;走查修七处(isLive 枚举比较/跨小月天数/Modal 日期 NaN/详情变体文案/ageBand 本地化/旧链接残留/e2e 基线断言);G-303/304/305 登记;双轴审查(正确性+家法)修 13 处:unlockedFor 接 HasActiveMembership(终止仍解锁+续费多行 Only 炸两缝同修)/fold 当期=最长活窗(与访问判定同族)/sku_name 死契约删除/客户端复用四件/e2e 补 settle 形守卫+清理对称/规范四搬移(SkuFacts↔DO+adapter 归 data.go/WindowUnlockRow/hydrateIdentity/const/ProviderSet 序) |
| M2-2 SKU/订单面 | lzx | ✅ 收口 9860eb34(10-31 双轴审查+走查毕) | 双 admin 面:skuadmin 纯 face+OrderAdminUC 归 module/order 兄弟位(§1 UC→UC 禁后重构,retirePending 共享仲裁)/下架即时同步 C 端/关闭 operator 审计行/走查捉四 bug(enum 全名×数字恒假族+hydrate 去重+optional 零值族+零值窄化)/G-306 手机搜索登记;CSV+审计管线递延;双轴审查修:close 入口+user_id 链接/unlock_count 包装/键序组序+跨面键回收/sonner 统一/PUT 映射/枚举映射归属主 |
| M2-3 激活码全面 | lzx | ✅ 收口 45c276d4(11-01 回退重审+走查毕) | 批次模型(activation_batch 行+码行挂 batch_id,计数聚合免漂移)/码表页(逐码核销人/时间+进度≥95% 补码+码值精确搜索补原型搜索框缺失半边)/停用-复启(未核销即失效,幂等答案压过停用不压过期,新错误码 6155)/导出 CSV(ACTION_WRITE 审计法条,客户端拼 BOM CSV)/e2e 捉模型洞(家族 grant 单行 ratchet 丢核销去向→trail 落码行 TakeUse 同语句)+trail 逐码对核销人+幂等答案相等性断言;擅自提交三笔回退重审(授权逐片裁定入档)修 12:分区两处(hydrateRedeems 入 Private/batchAgg 归 imports 后)+方法序三层(proto Export 序纠上轮误驳,§2 Other 垫底)+迁移去切片标(DB 版本列时间戳无感)+vocab 三份合一+CodesTable 命名+明细 tab 并入列表+Generate 死赋值+copyUnused 守卫+Disabled 直传;type 双用途+school_id=原型旧库形态不采,核销人搜索递延会员面,RESTOCK_AT 升 libs 触发=第三模块消费者 |
| M2-4 详情余 Tab / E9 增量 | lzx | ✅ 首刀两 Tab 收口 197fdbc6(11-01 走查+双轴审查毕) | assessments(场次=attempts 按 session 分组派生,原型 done/total 无 admin 数据源砍为项数·试次;解锁区=码/会员/策略三事实并陈不重 fold)+membership(GetMembership 整份载荷,6154 空态+头部开通 GrantMembership 真链);契约增量一处=ListActivationCodes+user_id(承接 M2-3 核销人搜索递延+解锁事实源);顺手修 O-1 五语言「返回用户列表」;双轴审查修 7:明细 lift 方向反转(降序 fold 把前次算成未来次,走查数据未覆盖同测两次故漏网)+route read-set 补全 6 op(§3 whole-set 法条,M2-1 摘要「Tab 级靠兜底」认知作废)+文件名对齐导出名(attempts→assessments)+ja 三键简体残留+错误行 i18n 键+开通按钮静默死区(summary==null 一律轻确认形)+act_renew 跨面键回收;trace 待 P22/messages 待 MSG-1 |
| MSG-1 消息反馈公告 / MSG-2 分享裂变 | — | 未认领(W7-W8) | P21/P23 前置;D-AS①② |
| PAY0 模块拆分 | lzx | ✅ 收口 4a590622(10-31) | 双聚合根拆毕(module/sku+module/order,order 跨模块持 sku.SkuRP),guard 豁免清零;顺手捉 M2-1 声明注释残留(归位体内)+go test 缓存×文件树守卫假绿洞(esports test target 已 -count=1;system 六守卫同洞留骨架车道) |
| PAY2 通道联调 | — | 等行政件 | 用户即办:类目+H5 权限+APIv3 密钥 32 位 |
| PAY3 合规执法+付费墙 | — | 未认领 | P7 稿就绪;PAY2 后 |
| MIG1 迁移工具 | — | W6-W7 建 | P14 开工前必裁;W8 预演(P6);**权益翻译口径已冻结(10-31,M2-1)**:Entitlement 终态=scope_membership 行 status(active/terminated)+expires_at+terminated_at,续期=append-only 新行,旧码 activation 翻译按此落 |
| R1..R5 精修带 | hzy(分配制,lzx 可代领) | R3 半收(mt-F1 ✅ 64ac733c),余 P10;R1/R2 次之 | 定义=plan hzy 精修带;R4 内含 ljr 确认点 |

## 开放遗留清单(完成即删行)

| 项 | 归属 | 触发/时点 |
|---|---|---|
| APIv3 密钥 wiki 记录 33 字符(应 32)——回调解密与 CP3 前必须核对 | 用户侧 | 即办 |
| 走查墙 4:运营测试账号口径(pg2-probe 无座位;种子超管=admin/admin123456) | 用户侧 | 即办(SQL 补座位或文档化) |
| 微信回调 next 丢失(redirect_uri 不能带 query,须 state 中转) | 后续微信链路片 | 设计时定 |
| 上线前登录切真实公众号 appid(JSAPI 绑域约束;测试号绑不了商户) | PAY2 | 上线清单 |
| G-202 反馈 CTA 门控占位(soon toast) | MSG-1 | 契约落地接真跳 |
| G-208 浮层余族(pay-confirm/match-switch/resume-remind/share-invite) | 各对应页面片 | 落地时逐一对齐 |
| G-203 多凭证保留卡掩码取序(主号语义) | ljr | 并入 P21 |
| conventions guard 正则精度修正回流 dev 兄弟服务(esports 已 AST 化) | lzx | dev 车道下次触碰同步 |
| 组内 spec doc 存量约 40 处行尾化 | lzx | 触碰 test 模块顺路或独立机械片 |
| 协议 consent 版本记录契约 | E9 增量 | 本片闭(H2 挂账销账) |
| skill_trajectory 30s 走查体感偏紧 | lzx | 评估加时 |
| mt-D5 intro 按款文案+unitScoreLabel 种子(10-30 走查补:后者现裸英文 universal,原型=中文单位如「15 分/成功闪避」) | lzx | 随种子/国际化面 |
| age-required 逃生按钮(先看免费报告;原型有面现仓无) | PAY3 | 付费墙落地时 |
| 双 N/4 撞义:match 头=已完成数 vs 测试面头=位置号,原型两处均裸 N/4(10-30 走查) | ljr(例会) | 措辞一句话裁,裁后 5 分钟落 |
| 待校准成绩报告不可见:report 契约缺行——SessionAttempts 滤 AuthoritativeScoreNotNil,curve-pending 行(有 raw 无权威)无处可见(10-30 DB 实证);建议 GetSessionReport 增 pending 行,ctr 消费端已备(6b2db084) | lzx | 契约+聚合小片+dev 常模种子+e2e pending 锚 |
| 回读面「本次成绩明细」行组缺(原型=mock,契约无逐次明细) | 例会 | 豁免候选或契约带明细行 |
| 未校准结算措辞「N 分·原始分」(原型无此态,10-30 新措辞已裁落) | ljr(例会追认) | 改词则换键 5 分钟 |
| home 续测卡 vs resume-remind 形态面检 | R2 | 精修带 |
| dev 栈状态目录 .run/ 迁仓外(D-AX):各开发会话代理的后台栈状态文件(pid/ports json)改写到仓外——建议 `$TMPDIR/cyber-dev/<project>.json`,pid 失活即视为陈旧可覆盖;本机存在 .run/ 者整目录删除(hzy 机实存自 10-26,lzx/用户机自查),.gitignore 不再兜底 | 各机 coding agent 自处 | 下次开栈前 |
