# esports 进度档 — 认领状态与验收记录

> 归属:esports 前沿进度档,三人共写。**剪枝律(2026-10-23 重申,约束三个 agent)**:片收口即剪枝为一行收口注(状态+日期+提交 hash+关键判定);节奏行一行一事;验证证据一律归提交信息的 proven-by 尾,本档不存叙述——档是给 agent 读的,读档成本即团队成本。长历史随剪枝删除,git log 即存档。检查点每站在下表登记。

## 现状快照(2026-10-23,W1 末)

| 人 | 已交付(均已并 esports/main) | 当前 → 下一片 |
|---|---|---|
| lzx | E0 契约(种子+增量2/3/4)、E1 骨架、E2 schema+种子、E3 评分(框架+Tier A 六款+提交缝,ce763107)、E6a/E6b、会员读面+admin 用户管理(88bcd9ba/3f272b28)、system 组合缝 ListUsersByPopulation(a8b9f0a4) | **▶ 视觉对齐批 V1 ✅ V2 ✅ V3 ✅(首页+编排+个人中心,genre 120°/mcard 族/fab 全族;待提交)→ admin 交互/文案对齐批(D-AK)**;后排:Tier A 余款四款(W3,首片 number_stars 补 CP1「过半」)→ E4(W4)→ PAY1+E8(W5) |
| hzy | H1 系统态件、H2 协议页、H-A 引擎框架、H-B 反应瞄准批(四款引擎,mock 净拔除)、10-22 精修批(控制台三红根治/profile 写缝/D-AG 落地/C3) | **H-C 判定回路闭环(已分配 10-23,D-AJ 首例,冒烟唯一硬断点)** → E7 追踪技能批(W2-3 起);挂起:D-AG/C3 实机走查随下次登录 |
| ljr | P1 计费(D-AB/AC)、P3 完成页(D-AI)、P9 能力归属(D-AD)、D-AG/AH 同批 | 待裁:**P13 窗外豁免签认(CP2 前)、P15 报告簇窗口(E5b 前)**、P8 性别(E5 前,W4 末;含协议正文联动)、P11(B 批前)、P12(E4 前)、P4(11 月底;HL 问卷回复红线 11-15)、P7(12-02)、P10(随 P4)、P2/P5/P6 |

**M1 燃尽**:E0✅ E1✅ E2✅ E6a/E6b✅;E3 ✅(13 款评分面收满,报告分 e2e 挂 E5);E7 引擎 4/13(判定回路待 H-C);E4 ⏳(增量a 已落);E5/E8/E9/E10 未认领;M2 用户列表+档案首面已前拉落地;支付 PAY1-3 全未启动。**CP1(11-08)前瞻:达标面基本在手,风险面=支付沙箱零启动。**

## 切片认领表(窗口与定义见 plan.md)

| 片 | 认领人 | 状态 | 收口注 |
|---|---|---|---|
| E0 契约首刀 | lzx | ✅ 10-21 | 种子(画像+目录)+增量2(测试读面)+增量3(统一提交,4 决策经裁)+增量4(三计数反哺,193a9b7c);余面已裁去向:报告读面随 E5/E9、entitlement 随 E8(D-AA) |
| E1 服务骨架 | lzx | ✅ 10-18 | kratos-new-service 八步;双面 accessLadder+catalog+部署面;会员链探针 e2e 绿 |
| E2 schema+种子 | lzx | ✅ 10-20 | D-V 撤回后按冻结契约 clean-room 重建;五表+三读+13 款种子骨架(校准值不杜撰,随 E3 单源);multi_target 补行=D-Z |
| E3 评分引擎 | lzx | ⏳ | 增量a metric.md 单源档、增量b Registry+纯函数+20 黄金 ✅;增量c 提交缝(4c5f0617)、增量d MT/TR 定稿(5b0160fb)✅,双轴评审两实错已吸收;增量e DZ/ST 定稿 ✅(净钟+哨兵双射闸,metric §6.7/6.8);增量g offense_defense 定稿 ✅(判键式正确性+题库入通用银行,metric §6.10,Tier A 8/10);增量h resource_match 定稿 ✅(碰撞事件流=契约增量5 补位+出生经济闸,metric §6.11,Tier A 9/10);增量i character_counter 定稿 ✅(rework 规格落地=c_max 分母+速度乘数+150 封顶,题库 12 关系行入通用银行,metric §6.12,**Tier A 10/10 收满**);增量j moving_target 定稿 ✅(Tier B→A 升格:采样流全量重算+阈值 20/30/80 实码解除冲突,metric §6.3);增量k tracking_moving 定稿 ✅(Tier B→A 升格:契约增量6=optional hit+dot_lifetime_ms,hit 存在性兼帧判别器+断言服务端交叉验证,metric §6.13);增量l economic_sim 定稿 ✅(Tier B→A 升格:契约增量7=窗口事实上事件+漏刀/终局血上顶层,判定与金额双交叉+时间线金币守恒,metric §6.14)——**Tier B 消亡,13 款全量重算,评分面收满**(报告分 e2e 验收随 E5);**余:无 scorer 余款;E3 收口面=报告分 e2e(挂 E5)+guard 正则回流**;「报告分 e2e」验收面挂 E5,H-C 前移回执层;guard 正则修正随下次镜像批回流 dev 兄弟服务 |
| E4 匹配引擎 | lzx | ⏳ | 增量a 聚合出生 ✅(match_config+match_session 两表+迁移+8 行种子;match.proto 双 RPC+错误 6130/6131;编排引擎=render_order 池法+FNVA/SplitMix64 确定性抽款+7 天 resume 窗+切换计数盐;重测 3/1 出生冻结;难度模式轴归属地=会话,metric §三已记;×1.5 偏好权重裁定不落——原型从未实现且新目录无 game→test 亲和轴,偏好影响=类型轴投票);增量b 提交缝接线 ✅(契约增量8=optional match_session_id,存在=会话局/缺省=练习局;序数真律=会话内计数+1,客户端 attempt_no 忽略化待 c 清号;预算执法=RETEST_EXHAUSTED,无效局同扣;阵容闸=TEST_NOT_IN_SESSION;会话失效=6132(归属折叠进 miss);difficultyFactor 消费=experience1.0/standard1.5,练习局恒体验基线;done/retest_left 推导经写入器激活;跨模块执法=test 声明 MatchSessionLaw 端口+match 侧 law.go 实现,wire 咬合零环);双轴评审三实错吸收=test_record (session_id,test_code,attempt_no) 部分唯一索引(count-then-insert 原子化,碰撞重数一次)+Switch 先验目标类型再关旧局+Get 创建竞态读回幸存者;余 c ctr 切换服务端编排(+attempt_no 清号)+d e2e 收口 |
| E5 报告框架+能力报告 | — | 未认领 | 服务/客户端两半可并行 |
| E6a ctr 纯壳 | lzx | ✅ 10-18 | D-R/S/T 全落;SEO 全量面同日收口 |
| E6b 登录+画像流 | lzx | ✅ 10-20 | 接通片(微信真人链路绿)+完整形态(登录/三步向导/个人中心/首页)+机制修复+R1R2+流程前段四页;遗留契约缺口随对应片:SMS 骨架(system provider)、consent 版本记录、自定义游戏上限 3v20 口径 |
| E7 游戏框架+13 款 | — | ⏳(引擎先行) | 反应瞄准批四款已由 H-A/H-B 前拉交付,批片余面=H-C 后 live 冒烟;追踪技能批/决策资源批未动 |
| E8 权益最小面 | — | 未认领 | D-W population 余件触发器=E8 |
| E9 最小 admin 面 | — | 未认领 | 用户管理子面已另落(下行);本片=激活码+测试项/配置管理+报告策略 |
| E10 全链 e2e | — | 未认领 | 全部依赖收口 |
| 会员读面+admin 用户管理 | lzx | ✅ 10-20 | MemberService 读面(88bcd9ba)+admin esports-users 列表/详情 Tab 骨架(3f272b28)+ListUsersByPopulation 组合缝(a8b9f0a4,D-X 落位);裁:一面一消费等级,BUILTIN 会员面不作运营消费面;详情余 Tab 随 E2/E5/E8 挂内容 |
| H1-H2(系统态件/协议页) | hzy | ✅ 10-20 | MR!2 双轴评审合并;EN 协议译文 ljr 定稿后整目录换值 |
| H-A 引擎框架片 | hzy | ✅ 10-20 | GameClock+GameModule 契约+TestRunHost+/test/$code+red_dot;biome 插件契约执法 |
| H-B 反应瞄准批 | hzy | ✅ 10-21 | 三款引擎语义锚钉死+本地结果面;mock 插桩净拔除零残留;E0 增量4 反哺闭环 |
| H-C 判定回路闭环 | hzy | ▶ 已分配 10-23 | D-AJ 首例;提交缝适配器+服务端判分回执;验收=真人全链(登录→/match→red_dot→分数回执) |
| PAY1-3 支付片 | — | 未认领 | P1 已定(D-AB/AC);商户资质就绪;遗留=经营类目补设+H5 支付权限申请(建议即办);**CP1 门含支付沙箱通=当前风险项** |
| MIG1 迁移工具片 | — | 未认领 | 2026-10-23 终局对账立片;W6-W7 建/W8 预演;school 步时序=P14;依赖 E8/PAY1/E5 |
| E5b 报告簇扩片 | — | 未认领 | 2026-10-23 终局对账立片;P15 裁后开工;8 维公式源已锚 plan 行 |
| E6c 账号域收口 | — | 未认领 | 2026-10-23 终局对账立片;SMS+密码+绑定合并;收编 SMS 散遗留行 |

## 开放遗留清单(防丢网;完成即删行)

| 项 | 归属 | 触发/时点 |
|---|---|---|
| 微信回调 next 丢失(redirect_uri 不能带 query,保 next 须 state 中转) | 后续微信链路片 | 设计时定(review A6) |
| favicon 404(ctr 无 public/ 目录)资产口径 | hzy | 待裁,随手片 |
| birth.ts 上收共享层(原拟随 H-B,未随落) | 触文件者 | 顺手清 |
| conventions guard 正则精度修正回流 dev 兄弟服务 | lzx | 下次定期镜像批 |
| D-AG/C3 实机视觉走查 | hzy+用户 | 下次微信登录同场 |
| SMS 骨架片(system provider+capability) | lzx | E6c(2026-10-23 立片收口) |
| 协议 consent 版本记录契约 | lzx | H2 versions.ts 真源 GET /agreements/current |
| 自定义游戏偏好上限:服务端 20/产品口径 3 | ljr | config 读面缺口,随 E9 或裁 |
| D-W population 余件(provider 行 member_population/座位迁移/guard 注释清理) | lzx | E8 权益面触发 |
| 走查墙 3:match 编排含无引擎款(tracking/skill 排进四项,流程 4/4 走不完) | hzy(H-C 顺带)或 lzx 小修 | W2 |
| 走查墙 4:运营测试账号口径——pg2-probe 无 user 座位登不进 admin 端;种子超管=admin/admin123456 | 用户侧 | 即办(SQL 补座位或文档化) |
| B 批修正清单(login toast 指名/备案串/微信回填昵称;nickname emoji 显式;setup 后端取核验;profile 模块集对核;横屏残留核) | lzx | E3-f 收口后一片收(清单=coverage §七.5,D-AL) |
| P11 年龄采集单/双轨待裁(契约单轨 vs 原型 band-only 合法) | ljr | B 批修正片前 |
| ICP 备案号权威串(LOGIC_AUDIT 2.5b「浙ICP2026028293号」vs MEMBERSHIP §九「浙ICP备2026028293号」) | ljr | 随 B 批 |
| 微信登录回填昵称/头像(2.5b-⑫;需 system identity 持久化+向导去向联动) | lzx | E6b 余面小片(S1 登记) |
| B 批修正清单已修订(两撤四增一 V4,coverage §七.5/§九) | lzx | 校准战役 S2/S3 后随 V4 一并排 |
| 版本检查页+app-version/内容块管理(coverage 缺口①;S4 终局对账低优) | 待排 | P13 签认后随 M2 余组/E9 增量 |
| 走查注:admin 工作台 ERR·NETWORK(dev 栈缺数据源,非 esports 面) | lzx | pre 部署前查 |
| §指针存量清扫:已提交代码内 §/文档名+章节引用全量清洗(E3 scorer 族/test.proto/seed.go/test_record schema 等;注释改自含 why) | lzx | 独立机械片(新法 10-23,见节奏行);随下次镜像批或单独排 |
| hzy 台账评审标记销项(H1/H-B 行曾滞「待评审」,本表已并) | — | 已随本档剪枝闭 |

## Review 遗留转交(2026-10-20 ctr 全项目 review;开放项已并入上表,此处仅存注记)

接受不改注记:B4 `user-scalable=no`(游戏 H5 防双击缩放,产品定位接受);D8 stores 空壳(D-T 能力配置面,机制待命)。

## 检查点登记(D-M)

| 检查点 | 日期 | 判定 | 拉闸动作 |
|---|---|---|---|
| CP1 | 11-08 | 待检(前瞻:支付沙箱=风险项) | — |
| CP2 | 11-22 | 待检 | — |
| CP3 | 12-02 | 待检 | — |

## 里程碑验收记录

(空;收口时一行一注)

## 待裁项状态

见 README §2.2;分析进展记 prd.md 对应条目。

## 节奏记录(一行一事;证据与细节归 git log)

- 10-18 立项+目录化(D-J);五路旧仓分析+三路审计入档;grill 定稿(D-K..D-Q,12-17 承诺);D-R/S/T 落;E6a 壳开工同日收口;E0 种子片冻结;E1 八步落地;E2 首轮落地。
- 10-19 E6b 接通片:微信真人链路全绿(D-V 兑现);E2 按 D-V 撤回;D-W/X/Y 裁;组合缝 a8b9f0a4;admin 只读面前拉提案否决(一面一消费等级)。
- 10-20 E0 增量2;E2 增量a 重建;E3 增量a(metric.md 单源);E6b 完整形态+机制修复+R1R2+流程前段 UI;H1/H2 交付;MR!2(H2+H-A)评审合并;会员读面+admin 用户管理落地;D-Z/D-AA/D-O2;批量追审+修复片;mock 插桩 TODO 法裁。
- 10-21 E0 增量3/4;E3 增量b;H-A/H-B 交付+走查;P1(D-AB/AC)/P9(D-AD)定案;D-AE/D-AF;小程序栈初步讨论(plan.md M5)。
- 10-22 E3 增量c/d(提交缝+MT/TR 定稿,双轴评审吸收两实错);hzy 精修批 MR!4 回流(三红根治/profile 写缝/D-AG 落地/C3);D-AG/AH/AI 裁;让步节制+环境分裂律立(README §5);注释精简批;dev-hzy 全量并入 main。
- 10-23 D-AJ 任务分配判据+H-C 分配+hzy 线盘点(27f87fb8);本档整体剪枝重梳:状态对齐 git 事实(全部「待评审」滞标改已并 main)、补登会员读面/admin 用户管理行、修切片表结构断裂、开放遗留清单建网。
- 10-23 E3-e:dual_zone/skill_trajectory scorer 定稿(净钟单向化+配置锚+终局对齐;哨兵双射闸;无效词表增 ABSURD_TIME_SPAN),metric.md §6.7/6.8 入档;自起栈三探针全绿(dz 90/st 521.6/反例 DETAIL_MISMATCH)。
- 10-23 E3-f:number_stars 定稿(进度指针+三闸+诚实零),Tier A 7/10=CP1「过半」;live 探针 raw 330 黄金逐位吻合。
- 10-23 D-AL:ljr 材料全量吸收+roadmap 校准批(DESIGN_DOC_MAP/DOC_INVENTORY 为导览;LOGIC_AUDIT/GAME_LOGIC_SPEC/溯源 PRD/admin IA/BLOCK5/题库表/openspec 指针入档;E3 七款×单位分吻合零返工,metric.md 零改动;B 批清单落 coverage §七.5;P11/备案串/迁移基线注登记)。
- 10-23 校准战役 S1-S3:S1 ctr 语义 14 面全绿(B 批两撤四增一 V4 一 admin 文案);S2 admin 面合规(6-Tab=F2 逐一对上);S3 契约种子全对(game 九款逐字/13 款 D-Z 逐行)唯 moving_target 主能力两源冲突→P12 提裁;满岁月粒度口径锚 PAY3。台账=coverage §九。
- 10-23 S4 终局对账(四列融合:旧后端 33 表端点级/旧前端 admin20+ctr60 路由/新原型 ctr43+admin41/新 PRD 全量):总判定=承诺窗(M1+M2 核心+PAY+M4)基本可达,前提=支付即启+H-C 落地+未认领片有人接;全量原型需 P13 豁免签认。立片 MIG1/E5b/E6c+E9 增量锚;提裁 P13/P14/P15;档位四档疑点解除(=BLOCK5 §7.7 逐字,§7.3 六级=系统套称号并行词汇);旧库复核=无支付网关、基线 V15(b4「V16+wx_jsapi」=新版前向编号无存量面)。
- 10-23 E3-h:resource_match 定稿(碰撞事件流补位=契约增量5+出生经济三键配置化+死亡经济闸),Tier A 9/10;§6.4 批量表余 character_counter 一款。
- 10-23 E3-i:character_counter 定稿(rework 规格落地:96.3% 饱和治理/c_max 穷举分母/速度乘数/150 封顶;题库 12 关系行入通用银行),**Tier A 10/10 收满**;E3 余 Tier B 三款闸门。
- 10-23 E3-j:moving_target 定稿(Tier B→A 升格:采样流全量重算,阈值 20/30/80 实码解除旧注冲突,计数×间隔对偶+密度闸+空流诚实零),Tier A 十一款;余 Tier B 二款。
- 10-23 E3-k:tracking_moving 定稿(Tier B→A 升格:契约增量6 补 hit+dot_lifetime,hit 断言服务端交叉验证+lifetime 反应闸),Tier A 十二款;余 economic_sim 一款。
- 10-23 E3-l:economic_sim 定稿(Tier B→A 升格:契约增量7 补窗口事实,判定/金额双交叉+时间线守恒+装备阶梯配置化),**Tier B 消亡,13 款评分面收满**;E3 余面=报告分 e2e(随 E5)+guard 正则回流。
- 10-23 新法(用户裁定):代码内禁一切 § 文档指针与文档名+章节引用(注释/schema Comment/proto 注释同禁),注释自含 why;已提交存量清扫=独立机械片待排。同日 E4-a 已按此法清洗(含迁移 SQL 重放)。
- 10-23 E4-a:match 聚合出生(镜像 E3-b/c 切法):match_config(fps/moba 各 4 能力序,scenario=渠道轴后并入同列)+match_session(steps 出生锁死/7 天 resume 窗懒关/switched+计数盐强制换 seed/难度模式轴/重测预算 3|1 出生冻结)+部分唯一索引执法「每会员一场活跃会话」;GetMatchSession 缺 genre=偏好多数票,平票→6130 走手选页;错误文案 ctr 双语+admin 五语;P12(moving_target 主能力)不阻塞——翻案=一行种子改动。
- 10-23 E4-b:提交缝接线:契约增量8(match_session_id 存在性=会话局/练习局分流)+序数真律(会话内记录计数+1,服务端独占)+预算执法(超限 6134,无效局同扣=消耗即真)+阵容闸(6133)+会话失效(6132,所有权折叠进 miss 防泄露)+难度轴消费(standard×1.5 实测 20→30)+done/retest_left 真值化;练习局回执 retest_left 改 0(ctr 当前零消费者,改动零面);探针事故一则:连发 404 系本人探针脚本 URL 少段(MatchService/),非服务端缺陷,已当场更正。
- 10-23 E4-b 复审提交:两轴评审(Standards/Spec 双子代理)吸收十项:硬违规=engine.go 种子直转 int64 补 ConvNum;重复过期律收拢 expireIfStale;难度词目具名 difficultyStandard(端口携字符串,反引成环,注释自含);proto 块注释英文化(切换/重新匹配原型标签移除);schema Comment 去 M1/ToB/交付叙述——迁移两轮重放(match_aggregate 重生成+test_record_session_ordinal 新增部分唯一索引);Spec 三实错=序数索引原子化+storeAttempt 碰撞重数一次(单测×2)、Switch 先验再关(单测)、Get 竞态读回幸存者(单测);attempt_no 残留校验随 c 清号;探针 400 误读一则(connect 错误 reason 在 ErrorInfo details 内,顶层无键)。
- 10-23 E4-c:ctr 切服务端编排收口:提交适配器落生(引擎事件流→契约 oneof 详情包装,4 款映射表+墙钟锚点入 host record)+会话面真化(arranger.ts 退役词表迁 vocabulary;阵容/done/retest_left 全服务端派生,genre 参数仅种子创建;switch 真调 SwitchMatchGenre+进度警面;guide 页 first/count 真化)+EndFace 判决卡(双轨分/等级/剩余重测/无效 reason 词表双语)+attempt_no 清号(残校随灭);breaking 红处置=骨架期清号合法类按声明线上报,提交后自愈,不设 ignore 豁免(用户裁定);文案纠偏两则+match_flow_soon 占位键退役;live 闭环:resume 局+0 事件局(0/C 仍扣预算)+40/40 局(400/98.7/S)+切 MOBA 新轴预算 1,psql 双会话/序数行锤实。
