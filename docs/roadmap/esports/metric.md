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

归属地(E4-a 定案 2026-10-23):模式轴挂在**匹配会话**——场次出生即冻结模式,会话内所有尝试按该模式计分(提交缝消费 difficultyFactor);M1 场次仅出生 experience,standard 的出生入口随正式测评权益面(E8/PAY)再开。

## 四、单位分表(线性链 · 📐 全表待实测校准)

计数类设计逻辑:**操作越慢的游戏单位分越高**(决策 25 > 识别 20 > 纯反应 10-15),让不同节奏游戏典型分数落在同一爽感区间。本表值=服务端默认;引擎参数键 `unit_score`(test_config KV 下发,§五一致性契约)按款覆写,C1 重校走配置不改码——键与本表同为本档单源。

| 游戏 | 能力 | 单位分 | 30s 典型成绩推断 | 典型分数区间 | M 槽位模式 |
|---|---|---|---|---|---|
| red_dot | 反应 | 10 分/命中 | 20-40 个 | 200-400 | 反应类 |
| flash_aiming | 瞄准 | 10 分/命中 | 15-35 个 | 150-350 | 反应类 |
| multi_target | 瞄准 | 10 分/命中 | 20-40 个 | 200-400 | 反应类 |
| target_recognition | 反应 | 20 分/正确识别 | 10-20 个 | 200-400 | 反应类 |
| number_stars | 反应 | 15 分/正确数 | 15-30 个 | 225-450 | 反应类 |
| dual_zone_dodge | 反应 | 15 分/成功闪避 | 15-30 次 | 225-450 | 反应类 |
| character_counter | 资源 | 15 分/克制命中 | 15-30 对 | 225-450 | 资源类 |
| offense_defense | 决策 | 25 分/正确决策 | 8-15 个 | 200-375 | 决策类 |
| resource_match | 资源 | 20 分/有效收集 | 15-30 个 | 300-600 | 资源类 |
| moving_target | 追踪 | 10 分/秒完美追踪 | 完美追踪 20-30s | 200-300 | 追踪类 |
| tracking_moving | 追踪 | 10 分/有效命中 | 20-30 发 | 200-300 | 追踪类 |
| skill_trajectory | 技能释放 | 8 分/贴合度百分点 | 贴合度 30-80% | 240-640 | 技能类 |
| economic_sim | 资源 | 15 分/有效正补 | 15-30 补 | 225-450 | 资源类 |

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

**Tier A 十三款=全员**(事实流足以全量重算):red_dot · number_stars · dual_zone_dodge · flash_aiming · skill_trajectory · offense_defense · resource_match · character_counter · multi_target · target_recognition · moving_target(E3-j 升格,采样流全量重算,§6.3) · tracking_moving(E3-k 升格,契约增量6 补 hit+dot_lifetime 后帧流全量重算,§6.13) · economic_sim(E3-l 升格,契约增量7 补窗口事实后事件流全量重算,§6.14)。
**Tier B 已消亡**(E3-l:economic_sim 经契约增量7 升 A 后,13 款全量重算——「采信客户端分」在新契约本无落点,E0 facts-only 设计使每款终须重算,三个 B→A 升格=该终局的兑现)。
新遥测契约(E7)逐款补字段后 B 可升 A(§0.2 领域机制重设计);multi_target 的 Tier A 已按代码现实裁决(README §0.5,覆写 MSB 的 Tier B 标注);target_recognition 同理——E0 增量4 将 spawned_enemy 落进明细契约后缺口闭合,E3-d 升 A 定稿(§6.6)。

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
- 引擎 boot 键(2026-10-23 E7-a 物化入 test_config KV,值=编译默认,键与本卡同源):target_size_px=20/target_lifetime_ms=1500/spawn_padding_px=50;判分键 unit_score=10。

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
重算(公式已于 E3-b 2026-10-21 誊录入 `module/test/flash_aiming.go`,锚=旧前端 FlashAimingTestPage.tsx:134-178):`composite = clamp(hitRate + accuracy + deviation + reaction, 0, 10)`,其中 `hitRate = 命中/(命中+失误) × 3`、`accuracy = 中心命中/命中 × 4`(中心命中 = deviation ≤ center_hit_deviation_px,引擎参数键与客户端 boot pack 同源,默认 20px)、`deviation = max(0, (30−平均偏差)/5)`、`reaction = max(0, (800−平均瞄准时)/100)`;空 shots 退化为 平均偏差 30/平均瞄准时 800(两分项归零);平均瞄准时 <80ms 判数据异常拒收。composite 为未来校准曲线的输入;权威分曲线未定稿前 authoritative 留空。
引擎 boot 键(2026-10-23 E7-a 物化入 test_config KV,值=编译默认):spawn_interval_ms=800/target_size_min_px=30/target_size_span_px=50/target_lifetime_min_ms=500/target_lifetime_span_ms=1000/center_hit_deviation_px=20;判分键 unit_score=10。

### 6.3 moving_target 定稿卡(E3-j;**Tier B→A 升格**,合成式誊录)

机制(代码确认):单目标持续追踪(非多目标 MOT);30s;6 图案各 5s 顺序轮换;20Hz 采样流。**阈值冲突解除**:旧注「前端 30/80 vs 后端 5/15」两注皆陈旧注释——前端 metrics 循环实码=perfect ≤20/good ≤30/图案可接受 ≤80(字面值);新系统单源引擎参数键 perfect_deviation_px/good_deviation_px/pattern_deviation_px(默认 20/30/80=实码 parity,E7 联调可调)。
事实流:shots[](契约名;实为 20Hz 采样点:elapsed_ms+deviation+pattern;target/crosshair/hit 留引擎审计面不入本公式)——E0 契约本就按服务端重算设计,B→A 升格零契约改动(multi_target 经 E0 增量4 同路)。
重算(对齐旧前端 MovingTargetTestPage 评分区与 Java Tier B 闸门):perfect/good/lost 三态按阈值逐点分类;**时间=计数×采样间隔**(旧版固定步长 0.05s 累积的精确对偶,引擎键 sample_interval_ms 默认 50);lost=转移计数(状态机初值 lost,首样本 lost 不计转移——旧版同臂);patternStats[p]=图案 p 内 dev≤图案阈值的采样秒数(六图案恒全初始化,avg 恒除 6——旧版同臂);coverage=goodTime/total×3(total=config 时长镜像);perfectScore=perfectTime/total×4;stability=max(0,(50−stddev)/10)(总体方差);adaptation=avg(patternStats)×2(秒均 [0,5]→分 [0,10],旧公式原样——该项独大易饱和,曲线吸收);lossPenalty=lossCount×0.3;composite=clamp(Σ,0,10)。
闸门:词表外 pattern=DETAIL_MISMATCH;deviation<0=DETAIL_MISMATCH;乱序戳=DETAIL_MISMATCH;span>config+1500ms=ABSURD_TIME_SPAN;**密度闸**=count×sample_interval>span+1500ms=ABSURD_TIME_SPAN(计数×间隔不得超流跨度——密集伪造凭空铸造追踪时间的面关闭,§6.5 族);空流→诚实零(旧空流 stability 白送 5 分,E3-f 空局瑕疵同族修掉)。
轨一:raw=perfectTime(秒)×unit_score(§四:10/秒完美追踪)×难度(完美追踪 20-30s→200-300);权威轨:曲线未定稿留空,composite 入 extras。
引擎 boot 键(2026-10-23 E7-b 物化入 test_config KV,值=编译默认=旧仓 parity;采样流节拍=sample_interval_ms 与重算的时间对偶同键):sample_interval_ms=50/pattern_duration_ms=5000/target_size_px=60/margin_px=100/perfect_deviation_px=20/good_deviation_px=30;判分键 pattern_deviation_px=80、unit_score=10。
方案因子 📐:体验 200px/s/40px/简单图案/宽松/30s → 正式 400px/s/28px/全 6 图案/严格/60s。
评级映射 🔗(主锚点 M3 追踪占比+辅助 M4 偏差):顶尖 ≥90% ≤15px / 职业 ≥80% ≤25px / 准职业 ≥70% ≤35px / 高水平 ≥55% ≤45px / 普通 ≥40% / 入门 <40%。追踪类无公开基准,推断成分更高,上线后内部百分位替换优先。
专有指标:轨迹适应性(6 图案各自 0-100)/最大偏差(需逐帧)/丢失次数。多目标「100 追 10」不属本游戏(归 future)。

### 6.4 批量卡(MSB,机制推断待代码确认;公式誊录随 E3 逐款定稿;十一款已升卡出表 → §6.5-§6.14,批量表现已清空)

| 游戏 | 双难度因子(体验/正式) | 专有指标 | 重算要点(JAVA 指针) |
|---|---|---|---|
| number_stars | 展示 1200ms / 800ms | 无 | 已定稿升卡 → §6.9 |
| dual_zone_dodge | 预警 800ms / 400ms | 无 | 已定稿升卡 → §6.7 |
| target_recognition | 干扰物 3 / 6 个 | 干扰抑制率(选择性注意范式) | 已定稿升 A → §6.6 |
| multi_target | 3 / 6 目标 | 目标切换耗时(注意转移范式) | 已定稿升 A → §6.5 |
| offense_defense | 决策窗口 2000ms / 1200ms | 压力决策正确率(执行控制范式) | 已定稿升卡 → §6.10 |
| resource_match | 旧代码无难度差分(时长可配;方案因子待 E7) | 收集速率(collected/span) | 已定稿升卡 → §6.11 |
| character_counter | 回合数 3(旧硬码)/克制轮盘密度 | 克制命中率 Q+速度乘数 | 已定稿升卡 → §6.12(rework 规格落地) |
| skill_trajectory | 轨迹速度慢 / 快 | 时机精度(提前/滞后量,能力核心) | 已定稿升卡 → §6.8 |
| tracking_moving | 速度 200 / 400 px/s(同 moving_target) | 同 moving_target | 已定稿升卡 → §6.13(契约增量6 补 hit+dot_lifetime) |
| economic_sim | 待确认(回合数/资源复杂度) | 最终资产、资源转化率 | 已定稿升卡 → §6.14(契约增量7 补窗口事实) |

### 6.5 multi_target 定稿卡(E3-d;Tier A,合成式誊录)

事实流:picks(target_id/tier/is_hit/elapsed_ms)+ missed_high(E0 增量4)。
重算:hits=is_hit 拾取数(陷阱拾取计入,旧仓 targetsHit 口径);rawTotal=Σ 档位值(high 100/medium 50/low 25/trap −50;档位值=引擎参数键 tier_value_{high|medium|low|trap} 可调,缺省=本表);决策时间=相邻拾取 elapsed 差分(首枚=距开始)。⚠️ 合成式改编披露:旧仓 avgDecisionTime=点击−该靶出生时刻,冻结契约的 pick 无出生戳,差分重建即「目标切换耗时」本义(本卡范式名同源)。闸门:差分<80ms=捏造(AbsurdReactionTime 拒收;较旧仓均值闸更严,red_dot 逐点法同款);档位词表外值=DetailMismatch(辖全部拾取,is_hit=false 同拒)。差分链跨全部拾取入链入闸。
公式(对齐旧前端 MultiTargetTestPage.tsx:177-210 与 Java MultiTargetScorer 逐行):value=rawTotal/(max(durationSec,1)×100)×5;priority=highHits/hits×3;speed=max(0,(1500−avgDecision)/150);composite=clamp(value+priority+speed−0.3×trapHits−0.2×missedHigh,0,10);hits=0→诚实零。⚠️ duration 取服务端 test_config(旧仓取客户端起止差,缩 Span 可抬 value 的 fabrication 面就此关闭);durationSec≤0 回退 30s(Java 同款冷臂,种子默认下不触)。
轨一:raw=hits×unit_score×难度(§四:10/命中);权威轨:曲线未定稿,authoritative 留空,composite 入 extras 待曲线(§6.2 flash_aiming 同款)。
引擎 boot 键(2026-10-23 E7-a 物化入 test_config KV,值=编译默认):spawn_interval_min_ms=500/spawn_interval_span_ms=400/target_lifetime_ms=2500/max_on_screen=8;判分键 tier_value_high|medium|low|trap=100/50/25/−50、unit_score=10。

### 6.6 target_recognition 定稿卡(E3-d;Tier A 升格,合成式誊录)

事实流:clicks(target_id/target_type/reaction_time_ms 可缺/elapsed_ms)+ spawned_enemy(E0 增量4——旧 Tier B 缺口即此,升 A 依据)。
重算:correct=target_type=enemy 的点击数(正确性=服务端按靶型判定,契约明文;旧 isCorrect 客户端断言废止);falsePositives=friendly|neutral 点击;avgRt=敌靶 rt 均值;敌击缺 rt=DetailMismatch(引擎对靶点击恒打 rt,缺=事实被裁剪——杜绝无 rt 拿满速度分的膨胀面);friendly/neutral 的 rt 不入均值。
公式(对齐旧仓 targetRecognitionScoring.ts PRD §2.13 与 Java 闸门):accuracy=correct/spawned_enemy×5;speed=max(0,(600−avgRt)/60);visual=min(1,max(0,1−(avgRt−200)/800))×2;composite=clamp(accuracy+speed+visual−0.4×falsePositives−0.2×max(0,spawned−correct),0,10);correct=0 或 spawned=0→诚实零。
闸门:逐点击 rt∈[80,60000](present 才判);correct>spawned_enemy=DetailMismatch(计数偏差,旧仓客户端钳 [0,100] 的服务端化);靶型词表外值=DetailMismatch。
轨一:raw=correct×unit_score×难度(§四:20/正确识别);权威轨:曲线未定稿留空,composite 入 extras。
引擎 boot 键(2026-10-23 E7-a 物化入 test_config KV,值=编译默认):spawn_interval_ms=700/spawn_padding_px=100/click_detection_range_px=25/invalid_click_margin_px=2;判分键 unit_score=20。

### 6.7 dual_zone 定稿卡(E3-e;Tier A,合成式誊录)

事实流:behaviors(elapsed_ms + zone∈{left,right} + event∈{collision,dodge} + obstacle_id 可选;旧仓大写 LEFT/RIGHT→契约小写)。**别名注**:种子 code=`dual_zone`(E2 统一;旧仓 `dual_zone_dodge` 即 skillPrediction 页面族,同一款)。
派生:碰撞即终局(legacy triggerCollision 同构)——碰撞后不得有事件且必居末位;dodged=dodge 事件计数;存活=**净墙钟**(span−Σbackgrounds 段,暂停顺延的服务端对偶),不再信客户端聚合值(旧 survivalTime/collisionCount/obstaclesDodged 直报废止,从流+顶层时间戳重建)。
重算(Java DualZoneScorer 逐行):serverScore=min((净秒/base_score_time)×100×difficulty_weight, 100);引擎参数键 base_score_time(默认 30)/difficulty_weight(默认 1.15;旧三键 simple/normal/hard 单键化——wire 无难度字段,模式轴归 E4 difficultyFactor);composite=serverScore/10。
闸门:净<500ms 或净超 config 时长+1500ms=ABSURD_TIME_SPAN(新词,词表第三员);乱序戳/词表外/碰撞后事件/碰撞与净钟不对齐(±1500)/流超净钟/obstacle_id 重复=DETAIL_MISMATCH。⚠️ 改编披露:旧 Java 双向 |span−survival|≤1500 在含暂停时误杀——新净钟化+**单向化**(流可早于钟=末尾空转合法,永不可超钟)+配置锚(净超 config 时长拒,§6.5 duration fabrication closure 同族)+终局对齐(碰撞局虚报满时长白拿分的面就此关闭)。
轨一:raw=dodged×unit_score(§四:15/成功闪避)×难度;权威轨:曲线未定稿留空,composite 入 extras。
引擎 boot 键(2026-10-23 E7-c 物化入 test_config KV,值=编译默认=旧仓 parity;旧 left/right_move_speed 两键→单键 obstacle_speed_px_s、旧 spawn 区间 min/max 双键→min+span、旧 graph_size→shape_size_px;杆锚定距离=BAR_INSET_PX 12px 不入键):shape_size_px=36/bar_thickness_px=24/bar_ratio=0.35/obstacle_speed_px_s=120/spawn_interval_min_ms=800/spawn_interval_span_ms=700;判分键 base_score_time=30/difficulty_weight=1.15/unit_score=15。E7 引擎改动:碰撞即 `end("completed")` 引擎自然终局(墙锚盖在致命帧上,净钟与碰撞对齐闸天然合法);闪避判定=过杆面(旧仓右区判定混轴的 bug 弃,轴去镜像);躲避杆随区内触点单轴跟随(旧仓须抓住杆本身,H5 热区放宽,判定面不变)。

### 6.8 skill_trajectory 定稿卡(E3-e;Tier A,合成式誊录)

事实流:rounds(round_number/shape 词表六形{circle|square|triangle|pentagon|hexagon|star}/deviation/placement_time_ms/player_placement 可空)。
重算(旧前端 SkillTrajectoryTestPage.tsx:332-374 与 Java 逐行):**哨兵 999=未匹配轮**——进 precision 分母(counted)、计 0 贴合度,不进偏差/速度均值(一轮没匹配不拖死两项,「少打轮不稀释精度分」同律);precision=(perfect+good×0.5)/counted×6(perfect<5/good<15);deviationScore=max(0,(50−avgDev)/10)(仅匹配轮均值);speedScore=max(0,(12000−avgPlace)/1500);composite=clamp(Σ,0,10)。空轮=诚实零(挂机不得分)。
闸门:shape 词表外/round_number 非严格递增(裁剪流)/deviation<0/placement 与哨兵不对射(placed⇔dev<999 双射;旧匹配上限 40px 使然)=DETAIL_MISMATCH。
轨一(**唯一非计数式**,§四:8 分/贴合度百分点):贴合度=clamp(1−dev/50, 0, 1) 按全体轮均值(未匹配计 0),raw=贴合百分点×unit_score×难度——零点 50 与偏差分同源,复算 §四表典型区间(贴合 30-80%→240-640)自洽。⚠️ 改编披露:旧 originalPosition 恒记 {0,0}、trajectory 实记玩家放置点(旧客户端未记真原位),故 deviation 作为客户端实测**事实**上报消费(与反应时同族),服务端不做几何重算;E7 引擎落 random_seed 服务端复板后可升全量几何重验。权威轨:曲线未定稿留空,composite 入 extras。
引擎 boot 键(2026-10-23 E7-c 物化入 test_config KV,值=编译默认=旧仓 parity;时长=配置驱动现行 30s——旧仓 60s 局内预算不入键,total_rounds=5 为引擎上限,钟尽即终、未确认回合不铸事实):total_rounds=5/display_time_ms=3000/target_size_px=60/palette_size_px=40/match_distance_px=40;判分键 unit_score=8。E7 引擎改动:棋盘=mulberry32(random_seed) 单源决定(位置→形→色抽取序固定,同种子同盘);回合代表=最优匹配对(旧 first-greedy 亦为最优对),deviation 由「匹配对均值」改为代表对**精确距离**——original/placement/deviation 三元组几何自洽,random_seed 复板几何重验的升级路径就此铺好;color_name=稳定英文 token(事实不随 locale 漂移);trajectory=代表件拖拽折线(旧仓误记放置点数组的脏点弃,originalPosition 记真靶心)。

### 6.9 number_stars 定稿卡(E3-f;Tier A,合成式誊录)

事实流:rounds(round_number/numbers 目标序列(引擎每轮 5 个 1-20 唯一)/clicks[]{number, elapsed_ms 回合内偏移, position})。旧 isCorrect/expectedNumber 客户端断言废止,roundsCompleted/correctClicks/incorrectClicks/avgDecisionInterval 四聚合值废止——全部从流重建(数组即真值,灌水回合数无面)。
重算(Java NumberStarsScorer + 旧前端 NumberStarsTestPage.tsx:225-310 同构):逐回合 expected=升序序列,**进度指针**判定——点击正确 ⟺ number==expected[progress](=「点当前未点数字中的最小值」本义),正确才推进;correct=Σ正确点击;incorrect=Σ(len(clicks)−progress);roundsCompleted=完整回合数(progress==len(numbers));avgDecisionInterval=各回合**正确序列内相邻差**均值(跨回合不连首尾;无相邻对→intervalCount=0 且 speed=0,「按 0ms 代入误给满速度分」的旧前端防呆同律)。公式:base=min(5, rounds/8×5)(Java 上限,前端无上限以 Java 为权威);speed=max(0,(2000−avgMs)/200);accuracy=max(0,3−incorrect×0.5);complexity=max(0,(maxSeqLen−5)×0.3)(引擎恒 5→常态 0,>5 留旧系数);composite=clamp(Σ,0,10)。⚠️ **correct=0→诚实零**(全错局/空局 composite=0;旧 Java 空局白拿 accuracy 3 分的瑕疵就此修掉,§6.5/6.6/6.8 同律)。
闸门:round_number 非严格递增(裁剪流,ST 同律)/回合内 elapsed_ms **非严格递增**(等值=同帧双击,0 间隔刷满速度分的新闸;降序=编辑流,DZ 同律)/numbers 空回合(引擎恒 5,空序列灌 roundsCompleted)=DETAIL_MISMATCH。⚠️ 改编披露:该款事实流无墙钟字段,无配置锚——伪造上限=满分 10,与旧架构持平(旧版客户端直报聚合值更易假);回合流自证(数组=真值)已封灌回合数的面。展示 1200/800ms 为客户端参数不入分。
轨一:raw=correct×unit_score(§四:15/正确点击)×难度(15-30 击→225-450,§四表自洽);权威轨:曲线未定稿留空,composite 入 extras。

### 6.10 offense_defense 定稿卡(E3-g;Tier A,合成式誊录)

事实流:decisions[]{scenario_id, choice?{attack|defense|retreat}(缺省=未作答), decision_time_ms, timeout}。旧 isCorrect 客户端断言废止——正确性=choice 对**题库答案键**(question_bank.answer.optimal_choice,EqualFold 等值,垃圾词表自然错);**查无 scenario_id=DETAIL_MISMATCH 硬拒**(旧版降级信客户端 isCorrect 是前端兜底库时代产物,判键不留降级面)。题库=E2 通用银行(test_code+scenario_code+payload 题面+answer 键;旧 offense_scenario+offense_choice_explanation 两表收敛为 payload/answer 两列,内容=旧 init.sql 经验数据逐字:20 场景×60 解析);答案键经提交缝入参(scoringParams.answerKey,TestRP.ListAnswers),scorer 保持纯函数。
重算(Java OffenseDefenseScorer 同构):correct=Σ等值;timeouts=Σtimeout;quality=correct/len(decisions)×6(**分母=决策数组自身**——旧版信客户端 totalRounds 属缩水分母伪造面,废止);speed=clamp((8000−avgMs)/2000×2, 0, 2)(avg≤6000 饱和);consistency=0.5(旧公式平项常数,曲线 parity 承袭);timeoutPenalty=min(2, timeouts×0.5);composite=clamp(Σ,0,10)。⚠️ 改编披露:旧 complexAccuracy 算而不入公式=死码弃;per-decision <80ms=ABSURD_REACTION_TIME **新闸**(旧版只闸均值,单个瞬时伪造可藏于慢兄弟之下);空 decisions→诚实零(旧 ZERO 同律)。伪造上限=全对+avg≤6000+无超时=8.5,与旧架构持平。
轨一:raw=correct×unit_score(§四:25/正确决策)×难度(8-15 对→200-375,§四表自洽);权威轨:曲线未定稿留空,composite 入 extras。决策窗口 2000/1200ms 为客户端参数不入分。
引擎 boot 键(2026-10-25 E7-d 物化入 test_config KV,值=编译默认):total_rounds=10/time_per_round_ms=2000/feedback_ms=800;判分键 unit_score=25。E7 引擎改动:题库=客户端镜像副本(M1 无题库读 RPC,scenarios.ts 逐字镜像种子 20 行题面+答案;内容变更须双改,题库读面缺口随 E9 题库面定形);客户端 optimal 只喂逐回合反馈面,判键永远在服务端(旧 isCorrect 断言废止律不变);决策窗单值 2000ms(卡面 2000/1200 双窗系难度轴,难度面未达引擎,standard 窄窗待难度面接入);超时回合无反馈窗(10×(2000+800)=28s≤30s 配置钟,钟尽截断=已铸决策照判);每局从 20 题银行随机抽 10(Fisher-Yates,无需复板——判键按 scenario_id,不依赖盘序)。

### 6.11 resource_match 定稿卡(E3-h;Tier A,合成式誊录)

事实流:trajectory[]{position, elapsed_ms?}(审计几何,不入公式)+collects[]{kind∈{minion|monster|dragon}, elapsed_ms}+**collisions[]{elapsed_ms}**(E0 增量5:冻结契约注释自称「重建全部效率聚合」却漏了碰撞事实——旧版靠客户端直报 collisionCount,恰是 facts-up 要废死的面;资源款零客户端消费者,加列即闭)。counts/encounters 全部从流重算,旧 minionsCollected/monstersCollected/dragonsCollected/collisionCount 直报聚合废止。
重算(对齐旧前端 ResourceMatchTestPage.tsx:174-202 与 Java ResourceMatchScorer 逐行):collected=Σcollects;resourceEfficiency=collected/optimal×4(optimal=Σ出生上限,引擎参数键 minion_count/monster_count/dragon_count 默认 25/12/5=42——旧 42 硬编码配置化);pathEfficiency=min(1, collected/span/2)×3;riskControl=max(0, 2−collisions×0.1);tactical=1(旧公式平项常数,曲线 parity 承袭,§6.10 consistency 同族);composite=clamp(Σ,0,10)。span=collects∪collisions 末事件 elapsed;无事件回退 config 时长(Java 30s 冷臂同款)。
闸门:词表外 kind=DETAIL_MISMATCH;出生经济越界(某 kind 计数>出生上限,或碰撞>collision_cap 默认 10=血量 100/每次 −10 的死亡经济)=DETAIL_MISMATCH(Java TEST_DATA_ABNORMAL 服务端化,上限随引擎参数走);乱序/重号戳=DETAIL_MISMATCH(§6.7 同族);相邻 collect 间隔<80ms=ABSURD_REACTION_TIME(§6.5 差分闸同族——出生重叠双拾的罕见真值会误伤,E7 引擎出生间距保证后消);span>config 时长+1500ms=ABSURD_TIME_SPAN(§6.5 duration fabrication closure 同族);有事件而 span<500ms=ABSURD_TIME_SPAN(10 次分离扣血物理上不可能短于半秒)。
⚠️ 改编披露:pathOptimization 名不副实——旧公式实为收集速率项非路径几何,按本义誊录,trajectory 留审计;span 分母=事件流末点(旧版=会话时长;契约无可信会话钟,末事件 span 是唯一服务端可重建的密度口径)——闲尾局(早拾后空转)可把 pathEff 抬至满格,较旧版至多 +2.25 composite,受完美清场 10.0 上界钳制,曲线以真实分布吸收;零收集活满场=3.0 保底(2 风控+1 战术,Java 同臂),composite 是曲线输入非常数分,曲线以真实分布校准吸收;伪造上限=完美清场 10.0,与旧架构持平。
轨一:raw=collected×unit_score(§四:20/有效收集)×难度(典型 15-30 个→300-600,完美清场上限 840);权威轨:曲线未定稿留空,composite 入 extras。双难度因子旧代码无差分(时长可配),方案因子待 E7 引擎片定。
引擎 boot 键(2026-10-25 E7-d 物化入 test_config KV,值=编译默认=旧仓 parity 换算,旧 px/帧@60fps→px_s):判分键 minion_count=25/monster_count=12/dragon_count=5/collision_cap=10、unit_score=20;引擎键 player_speed_px_s=180(旧 3px/帧)/enemy_speed_px_s=100(旧 1.5-2px/帧取中)/player_size_px=40/enemy_size_px=40/resource_size_px=30/collect_radius_px=25/enemy_count=5/on_field_count=6/spawn_min_gap_px=100/trajectory_interval_ms=200。E7 引擎改动:**分波出生**——旧仓全量静态撒 42 点系 800×600 桌面场;移动场宽下 100px 物理间距(拾取带 2×(25+15)px 与 80ms×180px/s 移动下限之和)容不下全量,改为场上常驻 6 件、拾取即从剩余预算补位,每 kind 总出生数仍=三键上限(optimal 分母不变);出生间距保证=本卡「相邻 collect<80ms 误伤」的 registered obligation 兑现(同时保证最早可拾时刻 ≥556ms>500ms span 下限);碰撞=沿接触去重(旧 wasHit 沿同臂),第 collision_cap 次碰撞血尽即引擎自然终局(墙锚盖在致命帧,§6.7 终局对齐同族);trajectory=200ms 时间采样(审计面,30s≈150 点);总血量=collision_cap×10(死亡经济与碰撞上限同源)。

### 6.12 character_counter 定稿卡(E3-i;Tier A,**rework 规格落地**——非旧仓已发布公式)

治理史:旧仓已发布公式=饱和公式(本地样本 26/27 满分=96.3%,全场最严重触顶,档位/能力聚合/报告区分度全失效),三根因=分母错配(每回合可达≈205/随机期望≈171,公式分母却按 100/回合)/服务端 3 分 fallback 虚高(查不到关系记 3 分,每回合白送~75 基线;前端同位记 0——移植「对齐前端口径」实际错齐)/timeBonus 常数(全局时限 30s<奖励阈值 30s 恒真)。admin 仓 openspec `scoring-character-counter-rework`(SCORING_TIGHTENING P0-1)立 rework 规格**未实施**——本卡=该规格在新系统落地(b3「denominator rework to land」兑现)。
事实流:rounds[]{round_number, enemy_lineup[], selected_lineup[], time_spent_ms};词表=克制关系题库的角色并集——种子 12 关系行(scenario=`<role>|<countersRole>`、answer=counter_score=20、payload 含 counter/countered 审计对),roster=六角色;题库走 E2 通用银行,E3-g 判键缝同款(ListAnswers→answerKey,scorer 纯函数)。旧 name→roleType 中文名解析废止——新词表=角色闭集,查无=DETAIL_MISMATCH 硬拒(旧版把裸字符串当 roleType 的降级面不留)。
重算(rework 规格):c_r=Σ命中对分值(命中=关系行 counter_score,未中=0——**3 分 fallback 删除**);**分母=c_max(E)**=roster 取 lineup 规模全部组合穷举取最优(6 选 5=6 组合,提交内联计算);Q=各回合 c_r/c_max 均值;SF=clamp(speed_ref_sec/avgTime, 0.5, 1.5)(REF 默认 8s,引擎参数键 speed_ref_sec——**常数 timeBonus 删除**);composite=Q×100×SF 封顶 **150**。
⚠️ 全族唯一 0-10 外曲线输入:150 封顶承重「完美克制×1.5 速度余量」的区分度(饱和治理的目的本身)——100=参考速完美,133.3=不完美但快,150=完美且两倍速。v0 机械锚点映射随曲线片,≥200 条真实数据后分位再校准(rework 原文承袭)。
闸门:空 rounds→诚实零(Java ZERO 同臂,回合数闸辖非空局);回合数≠round_count(默认 3,引擎参数键) =DETAIL_MISMATCH(旧版硬除 3、少报回合自降分,新律=回合数即契约);round_number 乱序=DETAIL_MISMATCH;阵容内重复/敌我不等长/空阵容=DETAIL_MISMATCH(**跨侧重叠=游戏本律**,6 角色双方各 5 必重叠,合法);词表外=DETAIL_MISMATCH;单回合 time<80ms=ABSURD_REACTION_TIME;Σ回合时长>config+1500ms=ABSURD_TIME_SPAN(回合共享会话钟,§6.5 同族);退化银行(c_max=0)护栏=c_max 置 1(回合无信号不炸除法)。
轨一:raw=克制命中对数×unit_score(§四:15/克制命中)×难度(典型 3 回合×8 对≈24→360);权威轨:曲线未定稿留空,composite(0-150)入 extras。双难度因子=旧无差分(回合数/时限可配),方案因子待 E7 引擎片定。

### 6.13 tracking_moving 定稿卡(E3-k;**Tier B→A 升格**,合成式誊录)

机制(代码确认):单 dot 存活制追踪射击;30s;点击即 shot(distance=准心到 dot 距离,**hit=distance≤45 实码** HIT_THRESHOLD);dotLifetime=dot 出生→击杀毫秒(玩家获取时间);movement=摇杆采样。**契约增量6**:旧提交=shot 帧与 movement 帧合并流(movement 帧以 target=(0,0)/deviation=0 混入,**无判别器**;且缺 hit/dotLifetime 两事实)——`TrackingMovingFrame` 增 `optional hit`+`optional dot_lifetime_ms`:**hit 存在性=帧判别器**(movement 帧两不设),lifetime 仅 shot 帧载。资源款零客户端消费者,加列即闭。
重算(对齐旧前端 TrackingMovingTestPage 评分区):accuracy=Σhit/len(shots);avgDev=shot 帧 deviation 均值;moveDist=movement 帧 player 折线累计;avgLifetime=命中发 dotLifetime 均值;multitasking=(accuracy×0.4 + avgLifetime<1500ms?0.3:0 + moveDist>500?0.3:0);composite=clamp(accuracy×4 + max(0,(40−avgDev)/8) + min(1,moveDist/1000)×2 + multitasking×2, 0, 10)。
闸门:**hit 断言交叉验证**=hit 值须与服务端按 deviation≤hit_threshold_px(引擎键,默认 45 实码)复算一致,不符=DETAIL_MISMATCH(facts-up:hit 是客户端断言,判定权在服务端——较旧版只闸区间前进);shot 帧缺 lifetime/movement 帧带 lifetime=DETAIL_MISMATCH(形状纪律);dotLifetime<80ms=ABSURD_REACTION_TIME(出生→击杀快于反应下限=伪造);乱序戳/负 deviation=DETAIL_MISMATCH;span>config+1500ms=ABSURD_TIME_SPAN;无 shot 帧→诚实零(客户端 totalShots=0 记 0 同臂——stability 项不得为「从不射击」付钱)。
⚠️ 改编披露:无命中时 avgLifetime 缺省 0<1500 → multitasking 0.3 白拿——旧版同臂原样承袭(曲线吸收);movement 帧的 target/deviation 旧线型恒 (0,0)/0,服务端不读(只读 player),新引擎可填真值不破坏。
轨一:raw=确认命中发数×unit_score(§四:10/有效命中)×难度(典型 20-30 发→200-300);权威轨:曲线未定稿留空,composite 入 extras。
引擎 boot 键(2026-10-23 E7-b 物化入 test_config KV,值=编译默认;crosshair_speed_px_s=400=方案因子正式档——旧仓速度为帧率耦合位移,新引擎以该键兑现 200/400 px/s 轴):hit_threshold_px=45/dot_size_px=40/spawn_padding_px=40/crosshair_speed_px_s=400/movement_sample_interval_ms=100/respawn_delay_max_ms=1000;判分键 unit_score=10。

### 6.14 economic_sim 定稿卡(E3-l;**Tier B→A 升格,E3 十三款收满**,合成式誊录)

机制(代码确认):单兵补刀窗口制;小兵血量衰减,点击补刀按剩余血量四窗口分类——perfect(melee/ranged ≤5,cannon ≤8)/good(≤10/≤15)/early(≤30)/too_early(>30);金币基数 melee/ranged 21、cannon 90,窗口比例 100%/60%/30%/0(floor);perfect/good 计 CS,perfect 另 +1s 时长;too_early 扣血 5%;商店三档 400/800/1200 ≤3 件;终局血 hp 口径 0-1500(health%×15)。
**契约增量7**:last_hits 增 `minion_type{melee|ranged|cannon}`+`minion_health`(1-100)——**窗口判定与金额复算权收服务端**,is_last_hit/gold_earned 降为交叉验证对象(facts-up:旧版两断言全信客户端);顶层增 `missed_minions`(漏刀无自然事件,计数申报)+`final_health_hp`(required,0-1500)。资源款零客户端消费者,加列即闭。
重算(对齐旧前端 PRD 2.6 公式):csEfficiency=cs/durationSec×3(分母=config 镜像);accuracy=cs>0?perfectCS/cs×3:0;items=len(purchases)×1.5;health=final_hp/1500×1.5;penalty=earlyHits×0.1+missedMinions×0.05;composite=clamp(Σ,0,10)。
闸门:词表外兵种=DETAIL_MISMATCH;**判定交叉**=is_last_hit 须与按 minion_health 复算的窗口一致;**金额交叉**=gold_earned 须与窗口复算金额逐一相等(金币守恒的全量版——不只 Σ 守恒,每笔可验);**时间线守恒**=两流按 elapsed 归并走余额,任一 purchase 时点余额<0=DETAIL_MISMATCH(startGold=0);cost 档外(item_cost_1/2/3 默认 400/800/1200)/件数>max_items(3)=DETAIL_MISMATCH(装备阶梯,Java 硬码三档服务端化+配置化);final_hp 非 15 步进=DETAIL_MISMATCH(HP 合法性,Java %15 同臂);两流各自乱序=DETAIL_MISMATCH;span>(config+perfectCS×1s)+1500ms=ABSURD_TIME_SPAN(perfect +1s 加时入上限,§6.5 族)。
⚠️ 改编披露:纯挂机=1.5 保底(health 项白送,旧版同臂)——composite 是曲线输入非常数分,曲线吸收;missed_minions 为申报计数无事件佐证(漏刀在引擎面不可观测,E7 引擎可加漏刀事件后收紧)。
轨一:raw=cs×unit_score(§四:15/有效正补)×难度(典型 15-30 补→225-450);权威轨:曲线未定稿留空,composite 入 extras。

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

- 2026-10-22 E3-d:multi_target/target_recognition 定稿卡入档(§6.5/§6.6,合成式誊录;决策时间差分改编与 duration 服务端化均披露);TR 升 Tier A(E0 增量4 spawned_enemy 契约落位,缺口闭合);§6.4 批量表缩为八款。

- 2026-10-23 E3-e:dual_zone/skill_trajectory 定稿卡入档(§6.7/§6.8,合成式誊录);净钟单向化+配置锚+终局对齐与哨兵双射闸均披露;无效词表增 ABSURD_TIME_SPAN;§6.4 批量表缩为六款。

- 2026-10-23 E3-f:number_stars 定稿卡入档(§6.9,合成式誊录);进度指针重建正确性、同帧双击新闸、correct=0 诚实零(旧空局 3 分瑕疵修掉)与无配置锚披露;§6.4 批量表缩为五款。Tier A 达 7/10——CP1「E3 过半」门达成。

- 2026-10-23 E3-g:offense_defense 定稿卡入档(§6.10,合成式誊录);判键式正确性(question_bank 答案键经缝入参)、查无场景硬拒、分母=数组、per-decision 时间闸新立、complexAccuracy 死码弃、一致性 0.5 常数承袭披露;题库种子=旧 init.sql 20 场景×60 解析逐字入通用银行(两表收敛两列);§6.4 批量表余二款。Tier A 达 8/10。

- 2026-10-23 E3-h:resource_match 定稿卡入档(§6.11,合成式誊录);碰撞事实流补位(E0 增量5:契约注释与公式事实面错位,加 collisions 事件列)、出生经济配置化(42 硬编码→三键)、聚合从流重建、间隔闸/时长锚/死亡经济闸新立、3.0 保底与速率项本义披露;§6.4 批量表余一款。Tier A 达 9/10。

- 2026-10-23 E3-i:character_counter 定稿卡入档(§6.12,**rework 规格落地非旧仓松公式**);c_max 穷举分母/删 3 分 fallback/速度乘数替常数 timeBonus/150 封顶四改编全披露;题库=12 关系行入通用银行(scenario=role|role 对);§6.4 批量表仅余 Tier B 二款。**Tier A 10/10 收满**。

- 2026-10-23 E3-j:moving_target 定稿卡升格(§6.3,Tier B→A:契约采样流本就支持全量重算,零契约改动);阈值冲突代码级解除(20/30/80 实码 vs 两注陈旧注释)、计数×间隔对偶、密度闸、空流诚实零(E3-f 族瑕疵修法)均披露;§6.0 名册 Tier A 十一款/Tier B 二款(tracking_moving 契约增量待立+economic_sim 待侦察)。

- 2026-10-23 E3-k:tracking_moving 定稿卡升格(§6.13,Tier B→A:**契约增量6**=TrackingMovingFrame 增 optional hit+dot_lifetime_ms,hit 存在性兼合并流帧判别器);hit 断言服务端交叉验证(deviation≤45 复算)、lifetime 反应下限闸、无 shot 诚实零、无命中 0.3 白拿臂承袭披露;§6.0 名册 Tier A 十二款/Tier B 一款(仅 economic_sim)。

- 2026-10-23 E3-l:economic_sim 定稿卡升格(§6.14,Tier B→A:**契约增量7**=minion_type/minion_health 上事件+missed_minions/final_health_hp 上顶层);窗口判定与金额复算双交叉、时间线金币守恒、装备阶梯三档配置化、HP 15 步进闸、perfect 加时入 span 上限、挂机 1.5 保底披露;**Tier B 消亡,13 款全量重算,E3 评分面收满**(报告分 e2e 验收随 E5);§6.4 批量表清空。
