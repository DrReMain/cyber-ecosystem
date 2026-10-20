# 评分与指标规格 · 单源档(METRIC)

> 归属:esports 前沿档,2026-10-20 由 E3 片建立(README §3 落位;coverage §六.3 数值档单源化的落点)。
> **单源律**:本档是评分规格(13 款指标定义/单位分/难度系数/锚点常量/8 维权重矩阵/档位定义)的唯一事实源。消费方:esports 评分模块(E3-b/c 代码)、黄金用例(E3-b 测试)、admin 配置面(E9)、报告板块(E5)。旧仓 UNIT_SCORE_DESIGN / METRIC_SPEC 族 / CALIBRATION_PLAN / FLASH_AIMING_TEST_STANDARD 自本档建立起**废止引用**——内容已清洗入档,冲突已按「代码现实 > 规格草案」裁决并留痕。
> 数值状态图例:🧊 冻结(代码现实/已裁,不随人群漂移) · 📐 草案(推断值,待实测校准) · 🔗 外部锚点(公开基准,上线后内部百分位替换)。
> 事实来源缩写:USD=UNIT_SCORE_DESIGN · MST=METRIC_SPEC_TEMPLATE · MSF/MSM=METRIC_SPEC_FLASH_AIMING/MOVING_TARGET · MSB=METRIC_SPEC_BATCH · RFS=REPORT_FRONTEND_SPEC §2/§3 · JAVA=esport-server Scorer/报告代码 · API=OpenAPI_用户端接口文档 · ST=ctr 原型 _state.js。

---

## 一、评分总纲(双轨,已裁 README §0.5)

1. **轨一 · 无上限分(单游戏展示)**:`分数 = 原始成绩 × 单位分 × 难度系数`(§四/§三),无上限;仅单局报告与分享卡的「原始战绩」处展示,不进跨游戏比较。爽感区间:典型一局 200-1000 分 📐。
2. **轨二 · 权威分/档位(跨游戏比较)**:百分位统一尺度;档位 S/A/B/C 按百分位判定(§二)。「超越本款游戏 X% 玩家」「XX 能力超过 X% 玩家」。
3. **red_dot 校准分(0-100,锚点冻结 🧊)**继续做权威分输入(§六专卡)——唯一已实装校准曲线的款;其余各款的校准曲线随 E3-b/c 逐款定稿。
4. **综合分** = 各能力百分位等权均值,缺维不参与、不计 0(ST compScore)。
5. 旧系统「12 款 0-10 分 ×10 过渡」**不迁移**:新系统全量双轨一步到位(JAVA ScoringEngine 刻度过渡的遗留废止)。

## 二、档位与评级套(门槛层)

**C 端四档**(百分位判定,全局一致;🧊 已裁):

| 百分位 | 字母 | 中文 | en key(展示词表候选) | 主色 |
|---|---|---|---|---|
| 前 5% | S | 专业 | professional | 金 |
| 前 15% | A | 优秀 | excellent | 绿 |
| 前 40% | B | 良好 | good | 蓝 |
| 其余 | C | 待提升 | needs_improvement | 橙 |

**系统套六档**(评级管理面口径,不进 C 端):顶尖/职业/准职业/高水平/普通/入门,百分位目标 ≥99/95/85/60/30/<30。C 端四档 ↔ 系统六档映射 = 待裁 P2(ljr,M2 评级片前 3 天)。

## 三、难度系数(后台可配 📐)

| 档 | 系数 | 定位 |
|---|---|---|
| 体验 | ×1.0 | 入门/训练,基准分 |
| 正式 | ×1.5 | 正式测评 |
| 高难挑战 | ×2.0 | future 档 |

校准验证 C2:同水平用户正式/体验成绩中位数比 ≈1.5;偏离(>2 或 <1.2)先调难度因子再调系数(§八)。

## 四、单位分表(线性链 · 📐 全表待实测校准)

计数类设计逻辑:**操作越慢的游戏单位分越高**(决策 25 > 识别 20 > 纯反应 10-15),让不同节奏游戏典型分数落在同一爽感区间。

| 游戏 | 能力 | 单位分 | 30s 典型成绩推断 | 典型分数区间 | M 槽位模式 |
|---|---|---|---|---|---|
| red_dot | 反应 | 10 分/命中 | 20-40 个 | 200-400 | 反应类 |
| flash_aiming | 瞄准 | 10 分/命中 | 15-35 个 | 150-350 | 反应类 |
| multi_target | 瞄准 | 10 分/命中 | 20-40 个 | 200-400 | 反应类 |
| target_recognition | 反应/追踪(P9) | 20 分/正确识别 | 10-20 个 | 200-400 | 反应类 |
| number_stars | 反应 | 15 分/正确数 | 15-30 个 | 225-450 | 反应类 |
| dual_zone_dodge | 反应 | 15 分/成功闪避 | 15-30 次 | 225-450 | 反应类 |
| character_counter | 资源 | 15 分/正确计数 | 15-30 个 | 225-450 | 资源类 |
| offense_defense | 决策 | 25 分/正确决策 | 8-15 个 | 200-375 | 决策类 |
| resource_match | 资源 | 20 分/正确配对 | 10-20 个 | 200-400 | 资源类 |
| moving_target | 追踪 | 10 分/秒完美追踪 | 完美追踪 20-30s | 200-300 | 追踪类 |
| tracking_moving | 追踪 | 10 分/秒完美追踪 | 完美追踪 20-30s | 200-300 | 追踪类 |
| skill_trajectory | 技能释放 | 8 分/贴合度百分点 | 贴合度 30-80% | 240-640 | 技能类 |
| economic_sim | 资源 | **待实测定**(资源量级未知) | — | — | 资源类 |

> red_dot 双轨注:展示分走本表线性链;**权威分输入**走 §六校准曲线——两条链并存不互替(总纲 1/3)。

## 五、M1-M10 槽位体系(MST,含 MSF 定稿)

槽位名固定、语义按游戏适配;某游戏无对应语义留空 null,不得自定义槽位名。专有指标在 `metrics` 外扩展(`ability_metrics`)。

| # | 槽位 | 反应类语义 | 追踪类语义 | 因子依赖 |
|---|---|---|---|---|
| M1 | 平均反应时 | 出现→点击反应时 | 平均重锁定时间 | ✅ 稳定锚点 |
| M2 | 反应时标准差 | 命中反应时离散 | 重锁定标准差 | ✅ |
| M3 | 命中率 | 命中/目标数 | 完美追踪占比 | ✅ |
| M4 | 命中精度 | 命中/总点击 | 轨迹贴合度(平均偏差) | ✅ |
| M5 | 漏触率 | 漏触/目标数 | 丢失次数/率 | ✅ |
| M6 | 误触率 | 误触/总点击 | 留空 null | ✅ |
| M7 | 疲劳指数 | 后段−前段反应时 | 后段−前段偏差 | ⚠️ 半依赖(60s 可靠) |
| M8 | 全程稳定性 | 变异系数 stdRT/avgRT | 偏差变异系数 | ✅ |
| M9 | 最快单次 | 最快反应时 | 最小偏差 | ✅ |
| M10 | 最佳窗口 | 连续命中最长序列 | 最长连续完美追踪时长 | ⚠️ 全依赖(只进评分层不进评级) |

决策/资源类适配:决策类 M1=决策时、M3=配对正确率;资源类 M1=单回合耗时、M3=资源转化率(M6 留空)。

**summary_data 标准字段**(开发契约,字段名固定):`metrics{avg_rt_ms, rt_std_ms, hit_rate, accuracy, miss_rate, false_click_rate, fatigue_ms, cv, min_rt_ms, best_streak}` + `ability_metrics{}` + `difficulty(experience|standard)` + `factor_snapshot{}`(必填,评分换算与审计回溯)。

**一致性契约**(新系统落位):判定全在服务端(E0 契约「facts up, verdicts down」已承重);判定阈值/因子/名称服务端单源经 test_config 下发,C 端零写死数值;config 带 version + 短 TTL(MSF §5.2)。验收红线:后台改门槛/因子/名称 → C 端同一数据源展示。

## 六、13 款指标卡

### 6.0 总表(Tier 名册按 JAVA 代码现实)

**Tier A 九款**(事实流足以全量重算):red_dot · number_stars · dual_zone_dodge · flash_aiming · skill_trajectory · offense_defense · resource_match · character_counter · multi_target。
**Tier B 四款**(仅聚合数据 → 强一致性闸门通过后采信客户端分):target_recognition(缺 enemiesSpawned,闸门=反应时合理性) · moving_target(逐帧未上报,闸门=自洽/区间) · tracking_moving(缺 hit 标志) · economic_sim(缺 perfectCS/earlyHits/missedMinions,闸门=装备阶梯/金币守恒/HP 合法性)。
新遥测契约(E7)逐款补字段后 B 可升 A(§0.2 领域机制重设计);multi_target 的 Tier A 已按代码现实裁决(README §0.5,覆写 MSB 的 Tier B 标注)。

### 6.1 red_dot 定稿卡 🧊(E3 首个权威分输入)

- 重算公式(JAVA RedDotScorer):`raw = 命中数 × (RT_REF / avgRt) × (0.3 + 0.7 × precision)`
  - RT_REF=500(avgRt=500 → 速度因子 1.0);avgRt clamp 到 [200, 1500](剔作弊/离群);precision = 精准点击数(deviation < 15px)/命中数;精度是加成因子非乘 0 因子(最低保底 30%)。
  - 反应时口径:含移动到目标的时间(反应+瞄准),非纯神经反应。
- 校准曲线(raw → 0-100,锚点冻结自 **217 局干净数据**分位,绝对标准不随人群漂移):

| raw | 0 | 8.8 (p50) | 14.7 (p75) | 28 (p90) | 59.6 (p99) | 120 |
|---|---|---|---|---|---|---|
| 分 | 0 | 50 | 70 | 90 | 99 | 100(cap) |

  分段线性;raw > 59.6 段斜率 = (100−99)/(120−59.6)。离群防护:荒谬反应时直接判数据异常(不记 0 分)。
- 评级映射:主锚点=平均反应时,辅助=命中率,初始门槛沿用 §6.2 flash_aiming 反应时锚点族(🔗 Human Benchmark:普通人简单反应时 200-250ms,职业 170-200,顶尖 150-170;选择反应时 350-400)。

### 6.2 flash_aiming 定稿卡(MSF)

双难度因子 📐:

| 因子 | 体验(Sixshot 式) | 正式(随机闪现 Spidershot 式) |
|---|---|---|
| 目标位置 | 6 固定位置轮换 | 随机位置闪现 |
| 停留时长 | 1000ms | 700ms |
| 目标半径 | 40px(1080p 逻辑) | 28px |
| 总时长 | 30s | 60s |
| 刷新 | 点掉立即刷新 | 消失后 200ms 再出现 |

评级映射 🔗(主锚点 M1+辅助 M3):顶尖 ≤170ms ≥90% / 职业 ≤200 ≥85% / 准职业 ≤220 ≥80% / 高水平 ≤250 ≥75% / 普通 ≤280 / 入门 >280(百分位 ≥99/95/85/60/30/<30)。
重算:JAVA FlashAimingScorer(公式对齐旧前端 FlashAimingTestPage.tsx:134-178;公式誊录随 E3-b)。

### 6.3 moving_target 定稿卡(MSM)

机制 🧊 代码确认:单目标持续追踪(非多目标 MOT);30s(GAME_DURATION=30);6 图案各 5s 顺序轮换 straight/curve/zigzag/circle/random/variable;三态机 perfect/good/lost。
⚠️ **未裁决冲突(旧仓遗留)**:完美/良好偏差阈值前端注 30/80px vs 后端注 5/15px——新系统单源:服务端 test_config 下发,初值 E7 引擎片联调定。
方案因子 📐:体验 200px/s/40px/简单图案/宽松/30s → 正式 400px/s/28px/全 6 图案/严格/60s。
评级映射 🔗(主锚点 M3 追踪占比+辅助 M4 偏差):顶尖 ≥90% ≤15px / 职业 ≥80% ≤25px / 准职业 ≥70% ≤35px / 高水平 ≥55% ≤45px / 普通 ≥40% / 入门 <40%。追踪类无公开基准,推断成分更高,上线后内部百分位替换优先。
专有指标:轨迹适应性(6 图案各自 0-100)/最大偏差(需逐帧)/丢失次数。多目标「100 追 10」不属本游戏(归 future)。

### 6.4 批量卡十款(MSB,机制推断待代码确认;公式誊录随 E3-b/c 逐款定稿)

| 游戏 | 双难度因子(体验/正式) | 专有指标 | 重算要点(JAVA 指针) |
|---|---|---|---|
| number_stars | 展示 1200ms / 800ms | 无 | avgDecisionInterval 入参秒须 ×1000 |
| dual_zone_dodge | 预警 800ms / 400ms | 无 | baseTime/difficultyWeight/最终分全由服务端 config 决定,忽略客户端值;墙钟一致性校验 |
| target_recognition | 干扰物 3 / 6 个 | 干扰抑制率(选择性注意范式) | Tier B:闸门=反应时合理性 |
| multi_target | 3 / 6 目标 | 目标切换耗时(注意转移范式) | 公式对齐旧前端 MultiTargetTestPage.tsx:177-210 |
| offense_defense | 决策窗口 2000ms / 1200ms | 压力决策正确率(执行控制范式) | 查题库 scenarioId 服务端判 isCorrect |
| resource_match | 4 对 / 8 对 | 配对效率(正确数/分钟) | 公式对齐旧前端 ResourceMatchTestPage.tsx:174-202 |
| character_counter | 字符数 15 / 30 | 计数偏差(回答−真实) | 查 moba_character + character_counter_relation 题库判分;已知「96.3% 满分」评分收紧 P0 治理对象 |
| skill_trajectory | 轨迹速度慢 / 快 | 时机精度(提前/滞后量,能力核心) | precision/deviation/speed 可从 rounds 重算,本质等同 Tier A(公式对齐旧前端 :332-374) |
| tracking_moving | 速度 200 / 400 px/s(同 moving_target) | 同 moving_target | Tier B:缺 hit 标志 |
| economic_sim | 待确认(回合数/资源复杂度) | 最终资产、资源转化率 | Tier B:强一致性闸门(装备阶梯/金币守恒/HP) |

## 七、8 维认知权重矩阵(三源合一 🧊:JAVA 常量 = API 公式表 = ST DIM8 逐值核对一致)

维度 key 以 RFS §3 为法定命名(README §0.5);`dim_v = Σ(能力百分位 × w)`,四舍五入取整;**任一来源能力缺测 → 整维 null 不渲染**(不画轴不留空)。权重输入=各能力历史最佳百分位。

| 维度(法定 key) | 中文 | RS | AA | TS | SRA | RA | DE |
|---|---|---|---|---|---|---|---|
| neuralResponse | 神经反应 | 0.6 | 0.3 | 0.1 | | | |
| sustainedAttention | 持续专注 | | | 0.5 | | 0.2 | 0.3 |
| motorCoordination | 手眼协调 | | 0.5 | 0.3 | 0.2 | | |
| stressResilience | 压力适应 | 0.2 | | | | 0.3 | 0.5 |
| resourceAllocation | 认知资源分配 | | | 0.2 | | 0.5 | 0.3 |
| systemCognition | 系统认知 | | | | 0.2 | 0.4 | 0.4 |
| logicDecision | 逻辑决策 | 0.1 | | | | 0.3 | 0.6 |
| creativeThinking | 创新思考 | | | | 0.3 | 0.4 | 0.3 |

**别名对照**(消费旧数据/接口时映射,新代码一律用法定 key):

| 法定 key | Java/API 全称 | ST 缩写 |
|---|---|---|
| neuralResponse | neuralResponseSpeed | neuralResponse |
| sustainedAttention | sustainedAttention | sustainedAttention |
| motorCoordination | visualMotorCoordination | visualMotor |
| stressResilience | psychologicalResilience | resilience |
| resourceAllocation | cognitiveResourceAllocation | cognitiveResources |
| systemCognition | systematicThinking | systemsThinking |
| logicDecision | logicalReasoning | logicalReasoning |
| creativeThinking | divergentThinking | divergentThinking |

- ⚠️ 同名冲突:法定维度 key `resourceAllocation` 与能力 key `resourceAllocation` 撞名——E5 代码层须命名空间隔离(维度词表/能力词表分域),文档层维持法定命名不改。
- **4 维分组**(能力测评报告页文案,每维 2 能力合成):`瞄准精准度 | 预判能力` / `神经反应时间 | 快速认知` / `逻辑决策能力 | 分析思维` / `资源优化配置 | 统筹规划`。
- **副能力**(PRD §8·29③):FPS 副能力 = decisionEfficiency(来源 flash_aiming 的 DE 次贡献);MOBA 无副能力(4 维,AA/TS 标未测)。
- **置信分层**(同龄百分位):⚠️ 两源冲突——API 文档:<10 样本 low(percentile 置 null)/10-50 medium/>50 high;JAVA 常量:LOW=10/MEDIUM=30。E5 置信层设计时定,本档暂记两值。
- 同龄百分位方法:Z-score CDF + 置信分层(plan E5 行)。

## 八、校准计划(C1-C4,CALIBRATION_PLAN 摘要)

| # | 目标 | 验收 |
|---|---|---|
| C1 | 单位分校准 | 13 款**中位数成绩** × 单位分 ≈ 500(爽感中心);新单位分 = 500 ÷ 中位数,取整偏好 5/10/15/20/25/50 |
| C2 | 难度系数验证 | 正式/体验成绩中位数比 ≈ 1.5(偏离则先调因子) |
| C3 | 爽感区间 | 每款 P25-P75 ∈ 150-1500 分 |
| C4 | 百分位初始锚定 | 内部试测分布(20+ 人:A 电竞 5+/B 普通 10+/C 新手 5+)作初始门槛,**上线后真实数据替换(首月每周校准)** |

试测:每人 13 款 × 2 难度 × ~1min。红点锚点例外:已冻结(§6.1),不参与重校,按校准计划用新数据校准的是其余款与百分位门槛。

## 九、溯源与废止

| 节 | 事实来源(旧仓,只读) | 处置 |
|---|---|---|
| §一/§四 | USD(2026-08-21 草案 v1) | 清洗入档,原稿废止 |
| §二 | RFS §2 档位表 + ST tierOf | 原稿废止(RFS 整体随块4 报告规格归 E5 再单源化) |
| §三 | USD §三 | 并入 §四来源 |
| §五 | MST + MSF §5.1/§5.2 | 模板与样例废止 |
| §6.1 | JAVA RedDotScorer(代码现实) | 代码即权威,本档为誊录 |
| §6.2/§6.3 | MSF/MSM(草案 v1) | 清洗入档,原稿废止 |
| §6.4 | MSB(批量草案) | 清洗入档,原稿废止 |
| §七 | JAVA SchoolReportServiceImpl + API「8维认知维度计算公式」+ ST DIM8 | 三源合一,原稿废止 |
| §八 | CALIBRATION_PLAN(方案定稿) | 摘要入档,原稿废止;执行随 E3-c 后运营 |
| 不入本档 | TRACKING_DESIGN(C 端行为埋点) | 归 E7 遥测契约 / M2 行为分析片 |
| 不入本档 | FLASH_AIMING_TEST_STANDARD(L0-L4 测试标准) | 其 fixture 思想承接到 E3-b 黄金用例;文档废止 |

## 变更记录

- 2026-10-20 建立(E3 增量 a):总纲/档位/难度系数/单位分/M 槽位/13 款卡(Tier 名册按代码现实)/8 维权重三源合一/red_dot 锚点/校准计划;裁决两处遗留冲突(moving_target 阈值口径→E7 定;置信样本阈值→E5 定)。各款重算公式誊录与校准曲线定稿随 E3-b/c 逐款增量入档。
