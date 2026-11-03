# Q 线·业务码校对精修役台账

> 2026-11-02 D-BB 立役。法条=plan.md Q 线节:三源对照(原型 D-AQ 渲染态读+prd/demand-ledger+proto 契约),作者盲独立复核;发现项**全部上线前修完**——面级即时走链落码,契约/schema 级入 lzx 队列同窗修完;e2e 全套=回归门。射面=esports 业务面/模块(IAM 底座六面=骨架车道不在射面)。
> 每片开工细化逐项验收清单(本档);收口行迁 progress 认领表。

## Q1 ctr 测评主链(11-03 开工;射面=登录/注册/向导→match→games→成绩+pending 明细)

清单基线=ctr-prototype-matrix.md 44 行映射(状态列滞后以 demand-ledger §一为准)。校对序=先 ◐ 待面检(未被任何走查覆盖),后近期已走查面(独立复扫)。已毕走查的消费面不重复登记存量判定,只记新发现。

### 逐项清单

| 段 | 面 | 原型→重建 | 校对点(三源断言) | 态 |
|---|---|---|---|---|
| A1 | 登录 | login.html→/login | 全字面逐字过(含页脚公司/ICP 带「备」/客服邮箱);UA 态=G-005;门槛/全选=E9 复走查毕 | 毕 |
| A2 | 昵称步 | profile-nickname→/profile/setup/nickname | 逐字过;观察=规则块蒸馏(原型含设计注记文字非用户文案,重建两行用户面)+max 措辞同口径简化 | 毕 |
| A3 | 年龄步 | profile-age→/profile/setup/age | 逐字过;性别 hint 后半句缺=10-30 裁剔除(代码注在案);**年龄段双轨缺块=P11 在册票**(契约单轨 vs 原型双轨,ljr 待裁)交叉引用不另立 | 毕(挂 P11) |
| A4 | 偏好步 | profile-setup→/profile/setup/games | 逐字过;④半角问号(已修);⑤权重×1.5 子句缺=服务端无实现 | 毕(④⑤) |
| A5 | 年龄门槛 | age-required→/age-required | 全落含退出说明逐字 | 毕 |
| A6 | 协议壳 | agreement→/agreement/{user,privacy} | E9 走查毕+版本行真源+演示文本注携带;一壳 vs 两路由=矩阵在册 | 毕 |
| B1 | 首页 | home.html→/ | 续测卡(✅R2);最近成绩/数据统计/报告中心入口逐字全过;会员 greet=Q1-① | 毕(①) |
| B2 | 编排列表 | match-guide→/match | 形态+字面全过(PM-9/12 裁已吸收;剩N次收据=E7 增强) | 毕 |
| B3 | 类型选择 | match-genre-pick→/match/pick | tie/normal/empty 三态+toast 逐字全过;推荐角标=登记增强 | 毕 |
| B4 | 切换确认 | match-switch-confirm→/match/switch | 逐字全过(rule2 分歧=G-208 N1 已注册) | 毕 |
| B5 | 前须知 | test-guide→/match/guide | 16/17 逐字全过;rule5 过期语义=Q1-② | 毕(②) |
| C1 | 单项介绍 | test-intro→/test/$code intro | 面结构+收据逐字过;**⑥introduction 种子 12/12 NULL→how-to 空盒**(客户端半边已修,种子半边入 lzx 队列) | 毕(⑥) |
| C2 | 测试中 | test-running→running 态 | 题头=PM-9 裁毕;退出=G-208 毕;e2e 全链日跑 | 毕 |
| C3 | 退出确认 | exit-confirm→BottomConfirmSheet | G-208 首刀闭(逐字+几何实测) | 毕 |
| C4 | 通用态 | loading/empty/error→全局 | idiom 族分歧=Q1-③(P 票) | 裁决中 |
| C5 | 记录页 | test-records | 矩阵行 17+block4 三态在册 | 核登记毕 |
| D1 | 成绩面 | test-result→end/review | 低优先复扫过(E5-f2+双轨律+ca729f5f 三重验证在案) | 毕 |
| D2 | 完成页 | test-done→/report/session/$id | E5-f2 按原型落地+走查在案 | 毕 |
| D3 | 续测提醒 | resume-remind→/ 续测卡 | ✅R2 毕(去向差=G-208 页面片);Q1 只核在册 | 核在册 |

### 发现与处置(面级即时修;契约级入 lzx 队列)

**第一腿 11-03(B 段全+C 段毕)**:B2/B3 逐字全过(增强项有据:/match 步行「剩 N 次」收据=E7 追踪批、pick「推荐」角标=代码注登记);B4 过(rule2 分歧=G-208 走查 N1 已注册,零欠账);C3 过(G-208 首刀闭,逐字+几何实测);C5 核登记毕(test-records ❌=矩阵行 17+block4 三态,在册)。

| # | 面 | 断言与证据 | 级 | 处置 |
|---|---|---|---|---|
| Q1-① | B1 home greet 会员态 | 原型 home.html:52-54 会员活性时 greet=「专属会员 · 有效期内」;重建仅 欢迎/未登录 两态。C 端无会员活性读:ProfileService 仅 Get/Update/PreviewMerge(profile.proto 三 RPC),唯一 member_until 在合并预览事实(profile.proto:114);MemberService=admin 面(member.proto:20「this is the admin plane」);GetAbilityDigest 无会员字段(report.proto:231) | **契约级** | 入 lzx 队列。建议形状:GetProfileResponse 增会员活性(如 member_until)——home 会话探针本就调 GetProfile,零加跳;客户端三态 greet 随契约落 |
| Q1-② | B5 须知 rule5 过期语义 | 原型/LOGIC_AUDIT(D-AL 吸收锚「续测 7 天反推」):「超 24h 缓存失效时会按已测结果反推剩余项,不重排」;重建文案「过期后下次进入会重新编排」=现行服务端行为(match/biz.go:195-203 expireIfStale 过窗懒闭+下场新抽) | 口径 | **已裁(用户 11-03)=服务端改反推续测**:入 lzx 队列(match 域改过期续用=按已测结果反推剩余项不重排,对齐 D-AL 锚);rule5 文案随服务端落地回反推语义(「24h 缓存失效」系旧仓 localStorage 细节,届时按服务端真行为表述) |
| Q1-③ | C4 通用态族 idiom | error:原型暖卡(📡+「网络连接失败」+「本次进度已保留」安抚行+重试/返回上一页/返回首页三钮,error.html:9-18)vs 重建 ERR 终端美学;loading:原型骨架屏 vs 重建 mono 扫描条;empty:原型三域变体 vs 重建通用+域内空态 | 作者级 | **已裁(用户 11-03)=网络错面还原暖卡,已落**:network kind 双宿主渲染暖卡 network-face.tsx(root ErrorPage=loader 级+ErrorFallback=组件级,📡→RadioTower 按 PM-4 lucide 仓法/进度保留安抚行/502·NETWORK 诊断行/三钮);其余 kind 维持 ERR idiom(无原型面);loading/empty 维持骨架 idiom,分歧在册 |

Q1 余量:A 段(login 11-01 E9 复走查毕/向导三步/age-required/agreement 壳)+C1/C2/D1/D2(近走查面,低优先独立复扫)——第二腿。


**第二腿 11-03(A 段全+C1/C2/D1/D2)+ 用户复测轮**:Q1 清单 21 面全毕。A1/A2/A5/A6/C2/D1/D2 零新发现(E9/G-005/G-208/PM-9/E5-f2/ca729f5f 在案判定吸收);A3 双轨缺块挂 P11 在册票;④⑤⑥ 三发现+复测轮活探翻案⑥/新增⑦:

| # | 面 | 断言与证据 | 级 | 处置 |
|---|---|---|---|---|
| Q1-④ | A4 偏好步标题 | 原型「你常玩哪些游戏？」全角;重建半角 ? | 字面 | **已修**(本轮,zh 键改全角) |
| Q1-⑤ | A4 偏好步权重子句 | 原型 desc 含「命中偏好的能力池内权重 ×1.5」;重建缺;match 域 grep 无权重实现(抽签=类型票定池后随机) | 行为级 | 入 lzx 队列待一句裁:服务端落偏好加权(抽签权重) or 子句弃(登记分歧) |
| Q1-⑥ | C1 介绍文种子 | test.introduction 12/12 NULL(活库实证);**首判「空盒」系误判——活探翻案:bootstrap.tsx:114 本有通用兜底句(在限定时间内,尽可能快且准确地完成每一项操作),盒从未空过**;真差=通用兜底 vs 原型 6 段能力 how-to 逐字文 | 跨端 | 客户端防御守卫已加(空文隐盒,兜底在时永不触发);**服务端半边入 lzx 队列:seed 增 introduction(文=原型 6 段逐字),兜底随之退役** |
| Q1-⑦ | C1 身份卡类词裸出 | 活探捉获:类目词(universal/fps/moba=内部池词)直出中文面身份卡(原型身份卡无此行);unit_score_label 同未种(仅空显) | 呈现级 | **已修**(category 全链拔除:bootstrap→run→attempt→intro-face 四文件死链清;unitLabel 在场才显,原型态) |

Q1 收口:审计毕,余项全在队列(①⑤⑥=lzx;②=lzx 已裁待落;P11=ljr)。回归门=ctr 门禁链+单测 129+门控 e2e 4/4(三轮全绿);**复测轮=hzypw 真登录活探**(④全角？实读/⑦universal 消失/⑥兜底句在场),教训=**判空盒前先追消费链兜底**。

## Q2 ctr 报告与成长(11-03 开工;射面=report 族五面+解锁呈现;E5/E5b+D-AB 计费口径)

清单基线=原型 report-center/trend/career/cognitive/single 五页(_state.js 渲染态读)+demand-ledger §一 18-22 行+report.proto。profile 内嵌报告中心入口=R2 已核,test-done/成绩面=Q1-D1/D2 已毕,均不重扫只引在案。解锁/会员 CTA 族在案 parked→PAY3(demand-ledger 18 行),本线核**在场与文案**,不核支付链。

### 逐项清单

| 段 | 面 | 原型→重建 | 校对点(三源断言) | 态 |
|---|---|---|---|---|
| E1 | 报告中心 | report-center→/report/center | 覆盖横幅逐字(已覆盖 N/6 核心能力——测满 6 项才能出完整 8 维认知报告);四卡 title/desc 逐字;tag 态族=已解锁/需解锁/暂未开放/进阶已解锁/基础免费/需先测评(重建现为 查看/已锁/开放/未测 系,待逐字比);digest 卡去向=最近记录锚(records[0] 语义);重建进度条=增强登记 | 审毕 |
| E2 | 成长趋势 | report-trend→/report/trend | 空态逐字(还没有可比较的测评/至少 2 次测评(或 FPS + MOBA 各一次)才能画出/去测评);共 N 次横幅;**混类型警示(原型决策②:含 FPS 与 MOBA 两类→综合分不可直接横向比较)**;逐能力行(Δ chip ↑+/↓/—+N 次·首 X→末 Y);基础趋势 tag 免费·已测即可用;进阶分析模块双态+会员 CTA(parked→PAY3,核在场与文案) | 审毕 |
| E3 | 职业指导 | report-career→/report/career | 闭态逐字(暂未开放·敬请期待/后台开关控制+开启后出现清单/不单独售卖);注意卡(不参与「测满 8 维自动解锁」+旧稿两表述不成立);返回报告中心钮;开态=旧仓语义级(E5b-c 在案:D-AB 默认 open,种子面核 areas/gap/处方) | 审毕 |
| E4 | 认知报告 | report-cognitive→/report/cognitive | 报告头(环+档位+已测 N/8 维度+M/6 核心能力);认知雷达缺维不画轴+<3 维退化文案;能力详情行(tier+值+条);缺维说明(还有 N 个维度没有数据…缺的核心能力:X、Y+测满句)+去测 CTA;**学科优势匹配/职业发展潜能/提升计划三配置模块在场否(原型=后台可配置+空态不造假)**;免责声明默认文;去趋势 CTA;锁态(ToC 默认不开放=D-AB) | 审毕 |
| E5 | 单次能力报告 | report-single→/report/$recordId | 页头档位徽章(免费可见);综合分卡(前 X%+档位 tag+超过同龄玩家 X%);雷达(N 维认知能力图+M 主+N 副+副能力聚合注/无副能力注);能力行族(**免费席=最高分·已解锁 tag+黄边**/锁行 blur+解锁查看详细分析与职业差距/分析三档措辞「中心命中率高,接近专业水准,是本测评的最强能力。」等/改进建议三档/对比条=当前 N 分 vs 历史最佳+✓ 刷新历史最佳);会员卡双态;单次解锁卡(¥19.9=SKU 真价·含 N 项能力的详细分析·解锁后永久有效);未成年拦截提示;出口族(返回+分享(parked MSG-2)+备案 footer);back 去向(原型=test-done,重建=/match,查在案裁决) | 审毕 |

### 发现与处置(第一波 11-03,桌面三源对照)

| # | 面 | 断言与证据 | 级 | 处置 |
|---|---|---|---|---|
| Q2-① | E5 分析句百分位反演 | zoneCopyOf 把 wire 百分位(超过份额)直接喂「处于同龄前 {pct}%」(vocabulary.ts:179);原型=100−pct(report-single.html:75);同行右列同值又渲染「超过 {pct}%」→ 同页自相矛盾(超过 85%+前 85%) | 数学 | **已修**:vars.pct=100−pct |
| Q2-② | 逐字漂移族(E1/E3/E5) | 雷达题缺「认知」(N 维能力图 vs 原型 N 维认知能力图);analysis_high 缺「中心命中率高，」;unlock_desc 缺「认知能力」「专项」;career desc 两处缺「4 周」(4 周处方/4 周训练处方);career note 缺「8 维」;lock blur 原型含能力名+短句、重建改写加句;lock cta 原型「职业差距」重建「同龄对比」 | 字面 | **已修**:zh+en 逐字回原型 |
| Q2-③ | E1 覆盖横幅分支缺 | 原型 covered≥6 切「已测满，认知报告完整（8 维）」/认知卡 desc 切「8 维完整呈现」;重建恒显未测满句 | 字面 | **已修**:条件分支双语言 |
| Q2-④ | E2 横幅半句缺 | 原型「共 N 次测评 · M 项能力有趋势数据」;重建只剩前半 | 字面 | **已修** |
| Q2-⑤ | E4 认知头方言错位 | 重建借单报告 CompositeCard(前 {100−composite}%+超过同龄玩家);契约 composite=已渲染维度等权均值 0-100 分值刻度(report.proto CognitiveDimension 0-100 weighted sum);原型=环形+裸值+「综合能力分」标签,无前 X% 方言 | 作者级 | **已裁(用户 11-03)=还原原型环形,已落**:新 cognitive-head.tsx(环形+综合能力分裸值+档位+双计数,无前 X% 反演);CompositeCard 退回单报告专用 | | E2 趋势行收据缺 | 原型行=Δ chip(↑ +N/↓ N/— N)+末值大数+「N 次 · 首 X → 末 Y」+「基础趋势」模块框(免费·已测即可用 tag)+混类条件警示 banner(含 FPS 与 MOBA 两类→综合分不可直接横向比较);重建=裸曲线+「N 点」 | 作者级 | **已裁(用户 11-03)=还原原型收据,已落**:TrendChart 行重构(Δ 徽章三色+末值大数+「N 次·首 X→末 Y」)+基础趋势模块框(免费·已测即可用)+混类条件警示 banner(决策②语义,单类无 banner);恒显图例键退役 | | E5 出口族 | 返回钮:原型「返回个人中心」→profile-overview,重建「返回测试列表」→/match;页头 back:原型→test-done,重建→/match;分享测试结果钮缺(=D-AI②/MSG-2 族) | 口径 | **已裁(用户 11-03)=还原原型出口,已落**:返回钮=「返回个人中心」→/profile;页头返回=跟随来源(历史退回,直入无历史落 /match);分享钮维持缺位(parked→MSG-2 随 D-AI②) |

**Q2 收口(11-03)**:五面审毕,七发现全处置——①②③④ 面级即修,⑤⑥⑦ 用户三裁同日全落(皆=还原原型)。回归门=ctr 门禁链+paraglide+单测 21 文件/129+门控 e2e 4 过 2 跳 0 败(修后);活探=hzypw 真登录:报告中心三处实读(4 周处方/横幅分支/认知卡分支)、认知面锁态(D-AB 默认不开放口径生效,旧「前 X%」方言确认离场)、趋势面诚实空态逐字(走查号 cohort 样本薄=真无点,服务端 trend 点=cohort 答案在场才生,biz.go:814-868 只读核毕;与 digest 覆盖计数=best 分在场口径不矛盾);单报告面活证受限=走查号无锚记录(digest 无 anchor→诚实 toast),E5 以 e2e 报告面断言(修后绿)+桌面三源承载。

登记不修:E4 三配置模块(学科优势匹配/职业发展潜能/提升计划)+免责声明后台可配置——契约无字段,D-AP 剪裁+P4 报告形态锚(ToB 交付物),**入 ljr/P4 登记**;E4 to-trend CTA 无条件(认知响应无进阶判定,会员见自家 upsell)+E1 digest 卡态语义(原型键=解锁态,重建键=已测态,真还原需 digest 携解锁判定)——契约级微欠,**lzx 队列注**;增强在案:进度条/较上局对比卡/提升建议板/SKU 真价口径/五区文案(E5 设计注)/较上局 delta(AnalysisCard contract)。
