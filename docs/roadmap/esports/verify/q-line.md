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

## Q3 ctr 账号与合规(11-03 开工;射面=账号安全/密码/绑定/合并族;E6c+E9 在案判定吸收)

清单基线=原型 account-security/password(三态 ?m=)/bind-wechat/bind-phone/account-merge-confirm(两态)五页+profile.proto/system 契约。吸收在案:E6c 走查(建片自审)、G-005(微信 UA 态隔离)、G-203(多凭证掩码取序=D-BC 保持现状)、d8718fd3(bind-phone 冲突 spec)、R2 B批⑤(安全两行)、R3/P10(语言行)、E9(协议两页+版本真源)、Q1-A5(age-required)、G-007(备取向替代注记)。合并成功页=原型缺(built 自增,矩阵在册)。

### 逐项清单

| 段 | 面 | 原型→重建 | 校对点(三源断言) | 态 |
|---|---|---|---|---|
| G1 | 账号与安全 | account-security→/profile/security | 顶注逐字(微信与手机号可同时绑定到同一账号。绑定后可任选一种方式登录，换设备不丢数据。);三行=微信绑定/手机号绑定/登录密码(态 tag 已绑定·未绑定·已设置·未设置+desc 双态逐字:已绑=掩码+可用于一键登录/验证码登录,未绑=未绑定——绑定后可用…;去向:未绑微信→bind-wechat,未绑手机→bind-phone,密码→password?m=change/set);解绑说明卡逐字(解绑：…需另一种登录方式已绑…解绑入口在各自详情页。);账号信息卡(昵称/性别/出生年月(未填注记仅填了年龄段)/游戏偏好);返回个人中心钮 | 审毕 |
| G2 | 密码管理 | password→/profile/security/password(+/login/forgot) | 三态头注逐字(forget=用手机号+验证码重设+未绑手机的账号请先绑定手机号/set=微信登录的账号默认没有密码…/change=已设置过密码——需先验证旧密码…);栏位(当前密码/手机号+发送 60s 倒计时/6 位验证码/新密码（≥6 位，含字母与数字）/确认新密码);校验文案五条逐字(11 位手机号/6 位验证码/新密码至少 6 位/需同时包含字母与数字/两次输入不一致);成功 toast(密码已修改/密码已设置)→回账号与安全;演示注=设计注记不迁 | 审毕 |
| G3 | 绑定手机号 | bind-phone→/profile/security/bind-phone | 头卡逐字(微信登录的账号绑手机后，可用手机号验证码登录，换设备也不丢数据。);手机号+验证码+发送钮 60s;冲突卡逐字(该手机号已绑定其他账号/本次绑定已拦截、未做任何改动。/建议改用「验证码登录」…或在那边解绑后再来绑定。);入口注记行(入口：X · 当前账号：昵称（微信登录，未绑手机）);确认绑定/跳过/跳过 toast;成功 toast 已绑定手机号 | 审毕 |
| G4 | 绑定微信 | bind-wechat→/profile/security/bind-wechat | 头卡逐字(💬+授权并绑定微信+绑定后可用微信一键登录，换手机也不丢数据。/我们只会获取昵称与头像，不会读取聊天记录。);冲突卡逐字(该微信已绑定其他账号/已拦截、未做任何改动。/建议改用「微信登录」…先解绑后再来绑定。)+按钮变「重新授权」;授权并绑定/跳过+跳过说明逐字(跳过不影响使用，可稍后在「我的 → 账号与安全」再绑。);非微信环境 toast 请在微信中打开(G-005) | 审毕 |
| G5 | 账号合并 | account-merge-confirm→/profile/security/merge(+success) | 可合并态:确认合并头+双账号卡(将被注销（空号）+掩码+无记录 · 无会员 · 无解锁→保留（富号）+事实行)+凭证控制校验卡逐字(短信验证码校验空号手机号+openid 一致+两项都过才允许合并)+黄警示卡逐字(合并后空号立即注销…此操作不可撤销，系统会留存操作记录。)+确认合并/暂不合并;确认 sheet 逐字(确认合并这两个账号？/空号将被注销，其手机号与微信释放。此操作不可撤销。);不可合并态:不可自动合并卡(两个账号都已有数据（测试记录 / 会员 / 报告解锁 / 推荐关系至少其一）…不开放自动合并)+人工引导+判定依据卡逐字(空号 = 没做过测评、没有会员、没有已解锁报告、没有推荐关系)+去意见反馈钮(G-202 占位族);seg 演示切换器=设计件不迁,态由 PreviewMerge 事实定 | 审毕 |

### 发现与处置(11-03,桌面三源对照+服务端只读核)

| # | 面 | 断言与证据 | 级 | 处置 |
|---|---|---|---|---|
| Q3-① | G2 密码页漂移族 | 三头注重写(set=「当前为快捷登录…」vs 原型「微信登录的账号默认没有密码。设置后即可用「手机号 + 密码」登录。」等);占位符简化(至少 6 位 vs 新密码（≥6 位，含字母与数字）/再次输入新密码 vs 确认新密码);提交钮词漂(完成设置/确认修改/重置密码 vs 设置密码/修改密码/重设密码);**客户端预检三缺**(旧密码空/手机号格式/验证码格式)+**字母+数字组成规则整体缺**——客户端无检查,服务端 ResetPassword 亦无(biz.go:436 仅 bcrypt,无组成校验) | 字面+行为 | **客户端半边已修**(头注/占位/钮词/五条预检含组成规则回原型);**服务端执法缺=lzx 队列**(密码组成规则服务端无实现,同 Q1-⑤ 家族) |
| Q3-② | G1 已绑微信行缺掩码+出生年月缺注记 | 原型已绑微信 desc=openId 掩码+可用于一键登录;重建「已绑定 —— 可用于微信一键登录」无掩码——契约 MaskedIdentifier 承载(biz maskIdentifier openid 3****4 式,biz.go:655)。出生年月未填时原型注记「未填（仅填了年龄段 X）」;重建只显 未填——age_band 契约在(profile.proto:141) | 字面 | **已修**:微信行掩码+未填年龄段注记 |
| Q3-③ | G3/G4 入口注记行半句缺 | 原型「入口：X · 当前账号：昵称（微信登录，未绑手机）」;重建只剩「入口：{entry}」——当前账号半句(绑定安全语义:告诉用户绑进哪个号)整体缺 | 字面 | **已修**:半句回位(昵称+登录方式,bind-phone=微信登录，未绑手机/bind-wechat=掩码手机号或微信登录) |
| Q3-④ | G4 头卡题 | 原型头卡题=「授权并绑定微信」;重建头卡与页头同词「绑定微信」 | 字面 | **已修**:头卡题回原型,页头维持 |

登记不修(在案吸收):G1 解绑卡=服务台口径(E6c 裁「那些详情页不存在」,code 注在案);G1 未绑行=动作链接代状态 chip(呈现选择,desc 已含未绑定词);G4 冲突建议句=CTA 化重构(用该微信登录›+注,E6c 真流);G5 全族=真流程在案(键逐字毕+双账号卡/凭证卡/sheet/警示/阻断族全对+成功页自增=原型缺在册+merge_cta/blocked_cta 增强);G2 forget 拆独立路由 /login/forgot=登录前语境(变更页内链接指路);G3 预检后发码=原型「无需发码」法条落地+SELF 态+merge 升级路=E6c 在案。

**Q3 收口(11-03)**:五面审毕,四发现全处置(①②③④ 客户端半边即修;①服务端执法缺=lzx 队列,同 Q1-⑤ 家族)。回归门=ctr 门禁链+paraglide+单测 21 文件/129+门控 e2e 4 过 2 跳 0 败(修后);活探=hzypw 真登录五读:安全页三行 desc 逐字+解绑卡服务台口径在场;bind-phone 入口行全句(入口：登录后 · 当前账号：哈哈哈（微信登录，未绑手机）);密码页 change 模式头注/栏/钮逐字+校验梯五连拒实测(空旧密码→请输入当前密码;纯字母/纯数字→需同时包含字母与数字;5 位→至少 6 位;不一致→两次输入不一致,原型顺序原句);bind-wechat 头卡 授权并绑定微信+入口行(哈哈哈（微信登录）)。已绑微信掩码分支+出生年月年龄段注记分支=走查号态不达(无微信凭证/已填出生),桌面+类型承载。

登记不修:E4 三配置模块(学科优势匹配/职业发展潜能/提升计划)+免责声明后台可配置——契约无字段,D-AP 剪裁+P4 报告形态锚(ToB 交付物),**入 ljr/P4 登记**;E4 to-trend CTA 无条件(认知响应无进阶判定,会员见自家 upsell)+E1 digest 卡态语义(原型键=解锁态,重建键=已测态,真还原需 digest 携解锁判定)——契约级微欠,**lzx 队列注**;增强在案:进度条/较上局对比卡/提升建议板/SKU 真价口径/五区文案(E5 设计注)/较上局 delta(AnalysisCard contract)。

## Q4 ctr 会员订单面(11-04 开工;射面=会员态呈现/SKU 价格呈现/订单契约面;PAY1 已建面)

清单基线=原型三页(member-detail/pay-confirm/pay-result 渲染态)+prd P1 计费矩阵(已定稿 10-21)+P1.5 订单行形状+sku 种子(19.9/29.9 P1 冻结价)。**边界先立**:member-detail/pay-confirm/pay-result 三页本体=PAY3(W5 付费墙全链)未建面,Q4 不建只登记;Q4 射面=已建的会员态/SKU/判决消费面+支付族契约就绪度。已毕面不重复:home greet(Q1-①毕)/digest·cognitive 判决(Q2-⑥毕)/合并事实行 memberUntil(Q3 G5 毕)。

### 逐项清单

| 段 | 面 | 原型→重建 | 校对点(三源断言) | 态 |
|---|---|---|---|---|
| H1 | profile ③ 会员卡两态 | profile-overview.html ③→/profile MemberCard | 原型两态:已开通(green tag+有效期至 X ›去 member-detail+已解锁 N 份报告永久保留+续费会员)/未开通(月卡 tag+查看详情 ›+¥29.9/月+开通会员);desc 双态逐字(会员期内测试自动解锁完整报告 · 历史报告永久保留 · 训练趋势 / 职业差距对比等增值页 vs 测评自动解锁完整报告 · …);重建=借用报告 ④ 卡(MemberCard)且 **unlocked 硬编码 false**(profile/page.tsx:128),memberUntil 契约已在(f0e928da)未消费 | 审毕(①) |
| H2 | 报告 ④ 会员卡 | report-single.html ④→report/page.tsx MemberCard | 原型两态:会员态(会员期内测试自动全解锁 · 历史报告永久保留)/未开通(¥29.9/月 · 会员期内测试自动全解锁 · 历史报告永久保留+开通专属会员 ¥29.9/月);重建 unlocked=report.unlocked 契约真值;词面+parked CTA 口径核 | 审毕(②) |
| H3 | 单次解锁卡 ⑤ | report-single.html ⑤→UnlockCard | 价格=SKU 读面真价(priceYuanOf trial kind);P1 词=¥19.9 买该次完整报告永久有效;report_unlock_title/desc/forever 逐字(Q2 已审 desc 族,本段复认价格链与 once/forever 词) | 审毕(过) |
| H4 | SKU 种子与呈现链 | seed.go seedSkus→ListSkus→两卡 | 种子价 1990/2990=P1 冻结价;名称(单次能力报告/专属会员·月卡)vs 原型标题(认知综合报告·单次解锁/专属会员·月卡)——P1 粒度=trial 非报告类型,种子名合规性核;listPrice=price 无折扣;kind 枚举串消费(String(s.kind))口径 | 审毕(过) |
| H5 | 订单状态面 | pay-result.html 订单卡→**C 端零订单面** | 原型订单面=结果页订单卡(订单号/金额/支付渠道 wx_h5|wx_jsapi/时间)+fail 卡引「我的 → 订单」(原型自身未画列表页);契约 OrderService GetOrder/ListOrders 在(C 端可读性核:鉴权/owning);重建零面=PAY3 范围,登记不修 | 审毕(登记) |
| H6 | 支付族契约就绪度(PAY3 输入) | pay-confirm.html 决策④+guard adult+under18 拦截+四态结果 | 决策④ 会员期隐单次;guard=adult;under18 内联拦截(FCA5A5 红卡+两出口);四态(ok/member/fail/cancel)文案族;merchant_not_configured 演示注——逐条对契约/合规配置(compliance-config 出生采集开关/付费年龄线)就绪度,缺口入 lzx 队列供 PAY3 | 审毕(过) |
| H7 | memberUntil 消费一致性 | 契约 member_until(最远活性窗尾,null=无) | 三消费面口径一致:home greet(在场性)/merge 事实行(有效期至)/profile H1(待接);失效/terminated 态呈现词(已过期(X)· 期内解锁的报告仍可看=原型 member-detail expired 态,PAY3 前无面,登记) | 审毕(过) |

### 发现与处置(11-04,桌面三源对照+admin 授会员活探)

| # | 面 | 断言与证据 | 级 | 处置 |
|---|---|---|---|---|
| Q4-① | H1 profile ③ 会员卡 | 原型两态(已开通=green tag+有效期至 X+已解锁 N 份报告 永久保留+续费会员/未开通=月卡 tag+查看详情+¥29.9/月+开通会员,desc 双态各有本句);重建借报告 ④ 卡且 **unlocked 硬编码 false**(profile/page.tsx:128)——memberUntil 契约在(f0e928da)页内 getProfile 本就携带,零加跳可接 | 字面+状态 | **已修**:新建 profile/ui/member-card.tsx ③ 本卡两态(键=memberUntil),desc/tag/CTA 全 ③ 本句双 locale;已开通态有效期至 X+parked 续费,pitch 态 ¥{price}/月+parked 开通 |
| Q4-② | H2 报告 ④ 卡 unlocked 分叉 | 原型 unlocked 态 desc 二分:memberActive?会员期内测试自动全解锁 · 历史报告永久保留:单次解锁 · 永久有效;重建恒显会员句——单次购买解锁的报告会自标会员权益 | 字面 | **已修**:MemberCard 增 memberActive(session 探针缓存零加跳),非会员解锁落新键 report_member_unlocked_single(单次解锁 · 永久有效/Single unlock · valid forever)双 locale |

过面孔:H3 ⑤ 解锁卡逐字(价格=SKU 读面真价,Q2 desc 族复认);H4 SKU 种子=P1 冻结价 1990/2990+名称合 P1 trial 粒度+listPrice=price 无折扣;H7 memberUntil 三消费面(home 在场性/merge 事实行/profile ③ 卡)口径一致。

登记不修(PAY3 输入固化):member-detail/pay-confirm/pay-result 三页本体未建=W5 PAY3 范围; parked CTA 三处(单次解锁/开通/续费)+③ 查看详情 ›+member-detail expired 态词(已过期(X)· 期内解锁的报告仍可看)随 pay client/页面落地;决策④(会员期隐单次)+guard adult+under18 拦截+四态结果文案族已固化在案;契约就绪**零缺口**(OrderService 四面+PayCredential 三通道+compliance pay_age_line 配置在,under18/merchant 执法=PAY3 合规执法计划内);原型「我的 → 订单」列表页原型自身未画——PAY3 需裁入口形态。

**lzx 队列注(微)**:profile ③ 已开通态「已解锁 N 份报告 永久保留」行无 wire 计数(digest 无 trial 解锁计数;admin MembershipSummary.unlockCount 有先例)——PAY3 前随 digest 或订单面携带,本片以 parked 注让位不硬编码。

**Q4 收口(11-04)**:七面审毕,两发现全落(①profile ③ 两态本卡回位 ②④ 卡解锁分叉);其余五面过/登记。**活探抓真雷一枚:运行时 memberUntil=JSON 平铺 ISO 标量非水合 Timestamp,seconds 读法 NaN→toISOString 抛→整面 ERR——修为 agreement 面双形状读法(string|{seconds}),不可读落「—」不抛面**。活探=hzypw 真登录双向:非会员态 ③ 卡逐字→admin GrantMembership 授 31 天(至 12-05)→首页 greet=专属会员·有效期内(**Q1-① 三态会员支首度活证**)+profile 已开通/有效期至 2026-12-05/③ 会员句逐字+中心四卡判决全翻(单报告/认知=已解锁,趋势=**进阶已解锁三态键首度活证**)→Terminate 回 pitch 态(键控双向);④ 卡 unlocked 支=走查号无锚记录不可达(Q2 同款),桌面+e2e 承载。回归门=ctr 链 --skip-nx-cache+paraglide+单测 22 文件/134+门控 e2e 4 过 2 跳 0 败。Q 线 Q1-Q4 毕,Q5 admin 面(W5)待派。

## Q5 admin 已建面(11-04 开工;射面=esports 自有 14 面;骨架面不在射面)

清单基线=原型 admin outputs/prototype 14 页(D-AQ:功能规划+文案维度逐字,antd 视觉自由)+admin-nav-anchor.md(组/名真值)+proto 契约+demand-ledger;O-1(返回带目标名)/O-2(操作列详情打头+终态降级文本)/O-3(敏感操作「敏感·二次确认·入审计」+执法语义)三断言形状+空态逐面核。已毕走查(会员 M2-1/SKU·订单 M2-2/激活码 M2-3/用户详情首刀 M2-4/E9 合规族/R5 复扫)不重复存量判定,只记独立复核新发现;G-301/302 菜单裁定在案吸收。

### 逐项清单

| 段 | 面 | 原型→重建 | 校对点(三源断言) | 态 |
|---|---|---|---|---|
| I1 | 用户列表 | 01-user-list→people/users | 组名「用户与会员」+页名+列集+筛集+行去向;空态;O-1/2 | 审毕(①) |
| I2 | 用户详情 | 02-user-detail→people/users/$id | Tab 集(资料/测评记录/会员/trace 暂缓/messages 随 MSG-1);返回「返回用户列表」;O-1 | 审毕(过) |
| I3 | 会员管理 | 01-member-list→people/members | 四 Tab(全部/有效/≤7 天到期/已过期)+来源筛;列+状态词;oracle=admin-member-anchor.md | 审毕(过) |
| I4 | 会员详情 | 02-member-detail→people/members/$id | 有效期/来源/续订史 append-only 展示;开通-续费-终止 Modal+五语言;O-3 执法语义 | 审毕(过) |
| I5 | SKU 管理 | sku-manage→pricing/skus | 列集+改价敏感入审计+上下架即时同步;O-2/O-3 | 审毕(过) |
| I6 | 订单与解锁记录 | 04-order-record→pricing/orders | 列集+状态词+筛;解锁记录视图(去向/来源)在场 | 审毕(②) |
| I7 | 订单详情 | 05-order-detail→pricing/orders/$id | 事实行集(订单号/金额/渠道/状态/时间线);关闭操作+O-3;O-1 返回 | 审毕(过) |
| I8 | 激活码管理 | activation-code→channels/activation-codes | 批次列表+创建流+停用(未核销即失效)+导出 CSV;O-3 | 审毕(过) |
| I9 | 码表明细 | activation-code-detail→channels/activation-codes/$batchNo | 逐码核销人/时间+进度条+≥95% 补码提示 | 审毕(过) |
| I10 | 协议版本 | agreement-manage→compliance/agreement-versions | 版本列表+发布/下架流+真源 versions.ts 口径 | 审毕(过) |
| I11 | 同意记录 | agreement-consent→compliance/agreement-consents | 记录列集+筛(版本/用户/时间) | 审毕(登记) |
| I12 | 合规配置 | compliance-config→compliance/compliance-config | 出生采集开关+付费年龄线+提示文案;敏感二次确认+入审计 | 审毕(过) |
| I13 | 报告访问策略 | report-config→reports/report-policies | 策略面(open/membership/tenant/closed 四值×报告族);D-AB 计费口径 | 审毕(过) |
| I14 | 游戏与因子配置 | game-config→assessment/game-config | 阵容/能力/渲染序呈现+编辑执法 | 审毕(登记) |

### 发现与处置(11-04,桌面三源对照+admin 活探)

| # | 面 | 断言与证据 | 级 | 处置 |
|---|---|---|---|---|
| Q5-① | I1/I2 users 两页「会员」词面残留族 | G-302 正名只改了页名(用户列表/用户详情),体内词族仍说会员:列头 col_user=会员/空态 暂无会员/计数 共 N 名会员/封禁确认 确认封禁该会员？+该会员的存量会话…/详情错误态 会员不存在或无法访问——N-3 概念错位的残留半边;另 副标题缺(原型 通用档案入口：点「详情」进用户档案…本页零 desc,orders/skus/members/compliance 均有)+出生行缺(原型 用户与画像 cell 出生 2008-06✓/出生未采集⚠,Member wire 有 birth_year/month,纯呈现) | 字面+呈现 | **已修**(五语言):词族 会员→用户(列头/空态/计数/封禁·解禁确认/详情错误态);补 desc=通用档案入口句(去「本原型仅开放」原型自注);年龄段列下补出生行(yyyy-mm/出生未采集) |
| Q5-② | I6 orders 列头/筛选项词漂 | 原型列=订单号/用户/商品/金额/**解锁方式**/渠道/状态/时间/操作+筛 商品：全部/单次 · 报告解锁/**会员 · 月卡**;重建列头=类型(scope 列)、筛选项=会员 | 字面 | **已修**(五语言):col_scope 类型→解锁方式;scope_period 会员→会员 · 月卡 |

过面孔(11/14):I2 用户详情六 Tab 齐+O-1 返回用户列表;I3/I4 会员两页=admin-member-anchor §三白名单(扫描/提醒/批量续费/导出/订单记录按钮)+§四六钉(续费顺延 append-only/终止只改状态/状态四值/来源三值/统计三卡/掩码形)全过;I5 SKU 七列逐字+改价/上下架 O-3;I7 订单详情八事实行+#单号题头+退款口径+解锁三态;I8/I9 激活码两页(停用=未核销即失效+已核销不变/码表逐码+导出+补码);I10 协议版本七列逐字+设为当前/删除审计注;I12 合规配置三卡逐字(出生采集/年龄线敏感 Modal/提示文案默认句)+callout 逐字;I13 报告访问策略 tenant 刻意不提供(服务端拒绝在先,注释在案)。

登记不修:I11 同意记录 渠道 列+desc 半句——契约 AgreementConsent 无 channel 字段(仅 login_method),**lzx 队列注**(渠道轴语义 vs 并入登录方式,一句裁);I14 game-config 呈现塌缩——原型 难度预设/三档权重/简·标·难/单位分 四列+因子配置抽屉(config_type 五组+单调性校验) vs 重建 运行配置调参 平表(键值+整体替换),原型页自注=真实 schema=test_config 平键值,重建更贴 schema 真相;**登记分歧不建列**(具名列=特定 config 键的派生呈现,抽屉已可编辑同物),状态/上下架=M2 余组 test-type 档案面在案;users 批量导入/导出 CSV=P24 族、最近登录/测评/报告解锁/会员列+来源/年龄段/设备筛=契约无字段(P23 login_record 冷存/trace 暂缓在案)。

**Q5 收口(11-04)**:14 面审毕,两发现全落(①users 词族+副标题+出生行 ②orders 列头/筛词,皆五语言);11 面过,2 面登记,1 微注入 lzx 队列(consent 渠道)。活探=admin 真登录读 users 页(题头+desc 逐字/列头 用户/出生 2000-06 行在/残留「会员」仅余菜单组合法名)+orders 页(解锁方式列头/desc 逐字/诚实空态)。回归门=admin 链 --skip-nx-cache+单测 3 文件/12+paraglide 五语言平价。admin:dev 临时起毕即停。**Q 线 Q1-Q5 毕;余 Q6 服务端仲裁(lzx 陪跑,未排窗)**。

## Q6 服务端仲裁(11-04 读腿开工;射面=order/sku/entitlement/activation 四域仲裁与审计行;修腿=lzx 同窗)

法条=D-BB 承重面 lzx 陪跑——**陪跑绑修不绑读**(用户 11-04「继续」令):本片=作者盲独立读第一腿,服务端业务码只读不碰(lzx 车道),发现项全数入 lzx 队列、PAY3(W5)开工前同窗修完——发现赶在 PAY3 建购买面前=本片排窗意图兑现。三源=proto 契约(entitlement/order/sku/activation_admin)+裁定族(D-AB 计费矩阵/MEMBERSHIP_DESIGN R1-R7/原型评审④)+种子(seedSkus)+已建三套门控 e2e 断言(ENT_E2E/PAY_E2E)。读面=entitle/entitleadmin/order/orderadmin/sku 模块 biz+data+schema+shared/kernel(判决单序)。

### 逐项清单

| 段 | 域 | 校对点(三源断言) | 态 |
|---|---|---|---|
| J1 | order 创建仲裁 | pending-unique 局部索引=trial 绑定/membership NULLS DISTINCT 故意不绑(schema 注在案:并发会员单可双活双付各叠各行);同单凭证重放零渠道往返;创建竞标输家孤儿渠道单 best-effort 关+自过期兜底;渠道集校验;scope 执法=trial 须携本人未授记录/period 拒 ref | 审毕(过) |
| J2 | order 结算仲裁 | 单条件 UPDATE 单仲裁者(pending\|closed→paid;closed→paid=渠道真相纠正,钱已收必授予=schema 法条兑现);at-least-once 双回调幂等(settle+grant 同事务,e2e 断言);金额/提供方错配=ack 止重试+Error 告警+对账承接(注释自述人工路径,不静默);trial 结算 Source=单号/period=PaidAt+PeriodDays | 审毕(过) |
| J3 | order 生命周期外沿 | converger(boot sweep 先折停机间隙+到期 fold+渠道拒关→查询定真相);日对账(billZone=北京时间墙钟/未发布日单重试/单日弃守护栏不楔死);admin close 无属主检查=访问梯在前(注释法条);foldUnlock=trial 单号自带 ref/period 按 Source 对账+期内解锁计数(计数骑手见①) | 审毕(过) |
| J4 | trial 购买防线 | resolveScope=记录属主(GetAttempt)+未授(Exists)+pending-unique 兜底→AlreadyGranted;e2e 断言外记录拒/重复授拒/金额 1990;账本唯一键=一记录一授(谁付都一样) | 审毕(过) |
| J5 | entitlement 账本 | trial 无到期=R3 永久;membership 续费叠行;terminate=单条件 UPDATE 翻**全部**活窗(续费叠行全翻,注释自述);家族 ratchet 只向前(等值不动作,不可变列走特许 raw SQL);Grant OnConflict no-op 不毒化事务(ErrNoRows=吞并答) | 审毕(过) |
| J6 | **会员续费窗缝** | settle expiry=PaidAt+31d **不读窗缝** vs Grant 砌缝(活窗尾起算,注释自述「tile without gaps」);会员活窗内续买=新窗重叠站窗,余期被吞 | 审毕(**①**) |
| J7 | activation 兑换仲裁 | TakeUse 单条件 UPDATE(used<quantity+活批次子查询,使用与痕同语句);同员幂等**胜过**批次停用(停用只 void 未核销,e2e 双向断言);过期家族授予不困死(fall-through+ratchet 前推);短码在长行上不烧用(e2e);过期码查先于幂等支=②微注 | 审毕(过,②) |
| J8 | activation 铸码 | 三族枚举双端执法(proto defined_only+ent enum cognitive/career/school);quantity 恒 1(无单次能力码=D-AB);批次+码单事务;批次号唯一索引碰撞重试×3(身份走生成 reason 谓词);导出=自身 trace 行 | 审毕(过) |
| J9 | policy 判定序 | ReportAccessVerdict 单序=码>政策行>默认折叠;membership 门;explicitClosed 区分 POLICY_CLOSED/NO_ENTITLEMENT;tenant 写面拒 NotReady/读面答闭;entitle 面与 digest 三键共引=漂移结构不可能;默认三分 career open/cognitive membership/school closed=D-AB 逐字 | 审毕(过) |
| J10 | 审计行 | pay_order 不可变钱行+transaction_id 回填=对账 join 键;entitlement append-only+terminate 痕;report_policy/sku updated_by;code redeemed_by/at 在码行自持(授予可被 ratchet 改写而痕不丢——注释自述);操作面 Log 带 by 九处(Grant/Terminate/SetPolicy/会员·订单·SKU Close/Generate/Export/SetBatchStatus) | 审毕(过) |
| J11 | e2e 断言承载 | 三套门控(ENT_E2E/PAY_E2E):十人竞一码=1 胜 9 exhausted 锚;同员并发双兑仍单用;ratchet/短码/停用-幂等三矩阵;pay 全链双回调+converger 到期 fold+对账 mismatch=0+admin fold 懒水合;membership 砌缝/终止后准入回落/期内解锁计数/settle 短行取最远活窗 | 审毕(过) |
| J12 | P1 矩阵对表 | seed trial_report 1990/member_month 2990+31d 逐字;付费只发 trial+窗口(无付费 report_type 授予);R2 完成即授=事件在 storeAttempt 后 best-effort(存储真相不受授予失败损);R4 到期新报告恢复部分可见;R5 聚合随会员(digest 三键复用判决);R6 退款只记不回滚(M1 注释);R7 月卡 31 天第 32 天过期 | 审毕(过) |

### 发现与处置(11-04,桌面三源对照;服务端码=只读,全数入 lzx 队列)

| # | 域 | 断言与证据 | 级 | 处置 |
|---|---|---|---|---|
| Q6-① | J6 会员续费窗缝:同一 SKU 两写入方两缝法 | entitleadmin.Grant 砌缝(活窗尾起算,entitleadmin/biz.go 注释自述 renewal law「consecutive periods tile without gaps」);order.SettlePaid **不读缝**——expiry=PaidAt+PeriodDays(order/biz.go:399),会员活窗内续买→新窗与站窗重叠,余期被吞(例:剩 25 天时付 ¥29.9,实得 +6 天非 +31)。resolveScope 对 membership 无活窗检查=期内续买可下单可支付。产品暴露面=原型评审④「会员期内隐藏单次,只留续费」+member-detail 活窗 CTA「续费会员 ¥29.9」——PAY3(W5)恰要建此购买。MEMBERSHIP_DESIGN R7 只钉 31 天/第 32 天过期未钉中窗缝,Grant 路径注释钉了。**随附骑手**:foldAll 期内解锁计数按砌缝行假设(periodStart=前行窗尾),重叠行下重叠区内 trial 解锁漏出当期计数(授予本体无损永久,仅运营计数视图漏)——缝修好计数自愈。修形已在:Grant 的缝读(MembershipWindow in-tx)搬进 settle 事务即成 | **契约级** | **已修(11-04 ⑨片,settle 事务内 MembershipWindow 缝读搬入:活窗续费=站窗尾+PeriodDays 砌瓦片,首购起算 PaidAt 不变;并发续费 settle 同缝重叠=操作面既受回放形,fold 取最远窗;骑手 foldAll 期内解锁计数随缝自愈零改;resolveScope 无活窗检查=保留(期内可续买即 PAY3 所建购买);pay e2e 续费臂钉判别断言=第二行到期==前行到期+31d 精确相等,修前必红)** |
| Q6-② | J7 Redeem 幂等注 vs 过期码重呈 | entitlement.proto:13-15「idempotent per (member, code): handing the same code to the same member again **returns the same entitlement**」无限定;实现先查码过期(entitle/biz.go:411)再查幂等支(:423)——已用码过期后重呈答 ActivationCodeExpired 非站窗授予(e2e 只断言未过期路径)。两侧皆可辩(过期=对码的诚实答),一句裁:注释加过期限定或幂等支上移 | 契约注释 | **已修(11-04 ⑩片,取幂等支上移形:活授权在手过期码重呈=答站窗授予且不烧使用次数,无授权过期码仍答 Expired,失效授权落穿棘轮不变;entitle e2e 过期臂换无授权净用户保语义+新臂钉「过期码+活授权→幂等」优先序;proto 注原文成立不动=码向契约收敛而非契约降格)** |

**Q6 读腿收口(11-04)**:十二项清单审毕——10 项过,发现 ①(契约级续费窗缝)+②(微注),全数入 lzx 队列 PAY3 前同窗;验证=纯台账零码改(服务端=lzx 车道只读),三套门控 e2e 在案=修腿验收基线。**修腿毕(11-04 同日,lzx ⑨⑩ 片)**:①缝读搬入 settle 事务+pay e2e 判别断言、②幂等支上移+entitle e2e 过期臂换净用户新臂钉优先序;门禁 build/test/lint --skip-nx-cache 绿,门控 e2e 三套 3/3 绿,残渣 0;Q 线 Q1-Q6 读写两腿全毕。
