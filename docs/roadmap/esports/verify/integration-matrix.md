# 集成走查矩阵 — 旅程级测试用例设计(2026-11-07 重构役三重立)

> 定位:已建面的**旅程级集成测试用例设计+执行真值**。R1 轮(2026-11-05..07)已收割归档(态记录+活证叙事=git 391ccc87);**本档态列全域清零**,正式测试阶段按本设计多轮重走。
> 设计校准依据:verify/coverage-map.md(三方覆盖核验,缺口清单=走查轮输入);预期锚=q-line 锚+demand-ledger 行+裁决族;覆盖账:57 例对已建 45 页/85 RPC 全映射(豁免账=coverage-map §一)。
> 用户既裁口径:数据=混合制;执行序=跨面闭环 X→C 端 C→admin A;校准点=现场发现即时落 prd 待裁/plan 工单。

## 1) 约定(每轮开工前置对照)

- **起栈**:后台任务四件——system:dev(11001)+esports:dev(11003)+admin:dev(5173)+ctr:dev(5174);就绪判据=grep `ready in`(vite)/进程日志(kratos);栈状态目录=$TMPDIR/cyber-dev(D-AX)。
- **登录法**:真实 UI 登录(admin=admin/admin123456 种子;member=§2 造号),password tab;禁 token/cookie 注入。
- **拟真优先**:chrome mcp 真实工具打头;兜底=遮罩吞点击→`evaluate_script` 内 `.click()`;RAC 受控输入→原生 setter+`input` 事件;时敏局=e2e 配方单脚本自包含(工具调用间延迟≠脚本内延迟)。
- **隔离上下文**:ctr 与 admin 共享 `session_token` cookie 名——两线独立 isolated context。
- **截图纪律**:每张前 `select_page` 带 `bringToFront`;结果目检;关键态存证命名 `<套>-<例>-<态>.png`。
- **locale**:钉 zh-CN。**净清**:每套毕两端净清(一次性号端到端删;fixture 复原;毕套 DB 断言零残留,必含 match_session)。
- **发现处置四路**:面级即修(当场落码走门禁)/缺陷→plan 缺陷工单/偏离→prd 待裁区/用户裁定→prd 落裁。
- **三律(11-07 立,溯及)**:①开工先 join coverage-map 缺口清单+plan B 工单,有在案项不得另下新判;②未对原型核视觉的页注「视觉轴未测」——字面锚≠质量证,布局判定须几何测定;③毕=步骤列动词全走完,欠动词=「半@R<n>+欠腿清单」。

## 2) 角色号族(混合制;建档脚本=walkthrough/ 目录工件)

| 号 | 态 | 造法 | 消费套 |
|---|---|---|---|
| `wt<base36>` 一次性号 | member 人口,零业务态 | SQL 直插三表(user+credential+membership;D-BG 后 CreateUser 无 population 入参) | C1/X2 提交腿 |
| R-rich | 有锚评估记录 | 造号+fixture SQL 直插(test_record+cohort 样本+match_session;样本行不共享) | X1/X3/C2/C1-7 |
| R-minor | 未成年(birth<付费年龄线) | 造号+直插 birth | X1-5/C1-2(u13 闸)/C3 |
| R-multi | 双凭证(phone+wechat) | 造号+双绑定 | C4(换账号钮) |
| R-member | 活窗会员(admin 授 31d) | R-rich 上现场授 | X1 各腿 |
| **R-empty**(11-07 增) | 空号(零记录/零会员/零解锁) | 造号即用,不跑向导业务态 | C4-2(KEEP_* 升级支)/C4-4 |
| hzypw | 既有活探累积态 | 已存在 | 只读备用 |

## 3) X 套 · 跨面闭环

### X1 会员授权闭环(admin 授/续/停 ↔ ctr 全判决面)

| 例 | 前置 | 步骤(行为级) | 预期锚 | 校准点 |
|---|---|---|---|---|
| X1-1 授会员五面翻面 | R-rich 非会员基线存证 | ctr 读 home/profile/report-center 基线→admin 会员管理开通(经用户详情,Modal 授 31d)→ctr 刷新五面 | home greet 三态(Q1-①);profile ③ 已开通态(Q4-①);member-detail 活窗紫条+权益 4+FAQ5;报告中心四卡判决翻 | 开通动线(G-304 在册);**毕@R2(11-07,沛然号,经 X1-2 造态链)**:admin 授 31d(`granted by operator period_days=31` 日志互证)→ctr 四面翻(home「专属会员 · 有效期内」/profile ③ 已开通·2027-01-08/member-detail 活窗紫条「已开通 · 有效期至」/report-center 判决翻「已解锁+进阶已解锁」) |
| X1-2 mock 购会员全链 | 一次性号走完向导 | ctr member-detail 开通 CTA→pay-confirm→mock 收银台三钮→支付→pay-result ok→回读 member-detail;**admin 读回腿:订单列表新增 paid 单+会员列表新行(来源=支付)** | pay-confirm 三态/决策④/mock MarkPaid/四态 GetOrder 真值;订单来源词面 | 收银台拟真度;R1 已活走一次=参照 git,正式轮重走;**毕@R2(11-07,沛然号重走)**:never 基线(无条+开通 CTA)→pay-confirm(P7 形+决策④)→mock 收银台三钮→成功→result ok(订单 Pdbn425tthm89nc9n3q2g,收据四行+有效期 2026-12-08)→member-detail 活窗;admin 读回腿=订单列表 paid 单(沛然·wechat_native)+membership tab 来源=**支付**·共 1 期;日志 `order settled`+`MarkPaid code=200` |
| X1-3 续费砌缝 | X1-1 号活窗内 | admin 再授(续费)→读新行有效期 | 续费=活窗尾+31d 砌瓦片(Q6-①;append-only 续订史) | 续费入口分工;**毕@R2(11-07,顺带收)**:GrantModal 续费形(「从 12-08 无缝顺延 31 天 → 01-08?」+敏感·二次确认·入审计徽)→共 2 期 append-only(手动开通 12-08→2027-01-08 砌瓦+支付 11-07→12-08);ctr profile/member 同步翻 2027-01-08 |
| X1-4 终止回落 | X1-1 号 | admin **真执行终止**(O-3 敏感,非捕后即撤)→ctr 回读 | ctr 回 pitch 态;terminated 折 never 形、解锁内容仍可读(R-04);报告判决回落 | 终止后文案语气;**毕@R2(11-07,真执行)**:TerminateModal detail 形(「终止后立即失效;期内已解锁报告永久保留(只改状态不删行)」)→头卡翻已终止+期历史行留档;日志 `membership terminated by operator`+CheckOperation 鉴权先决+code=200;ctr 回读=member pitch 回归(CTA 回「开通会员」)+report-center 判决回落「需解锁」×2+home 回非会员 greet(期内解锁 0 份,R-04 仍可读腿=零解锁无涉) |
| X1-5 未成年拦截 | R-minor | ctr 报告锁态 CTA→pay→年龄线拦 | pay-confirm 未成年红卡+逃生双钮;不产生订单(admin 侧空查证) | 红卡文案/双钮去向;**毕@R3(11-08,wt-minor-301 11 岁)**:member 开通 CTA→/pay 红卡「你未满 18 周岁（已按出生年月精确计算：11 岁），无法购买付费报告或会员。免费层（5 维能力图 + 综合分 + 最高 1 项的档位与分析）仍可正常查看」,双卡对未成年不渲染;逃生双钮「返回看免费报告」(回 from 源)/「返回首页」;pay_order 零行 |

### X2 消息反馈全环(V10 法+余腿)

| 例 | 前置 | 步骤 | 预期锚 | 校准点 |
|---|---|---|---|---|
| X2-1 反馈六腿闭环 | 一次性号已登录 | ctr 提交反馈(六类选一+contact)→admin 队列 pending→快捷回复→ctr 追问翻 reopened→admin 详情线程回复→admin 关单(幂等)→ctr 读关单态 | 六腿全序(V10 走查法);关单态=R-08 逐字;contact 自留行 admin 详情在场 | 关单幂等 UI 形态 |
| X2-2 公告扇出+**余腿三件** | admin 已登录 | 公告控制台:存草稿→**编辑草稿(改题/改正文)**→行内发布(letter 全员)→ctr 收信+未读三处(tabbar/home/行)+开信即读→admin **查看 modal**→**下线**→ctr 侧存量信不回收确认 | 全员 letter 名册全投(DB 计数对账);开信即读红点消;两步发布=PM-19;编辑/下线=PM-20 复核面 | **余腿三件毕@R2(11-07)**:编辑=UpdateAnnouncement 首证(「草稿已更新」+预填+改题落行);查看 modal=详情+类型/范围/状态 chips;下线=OfflineAnnouncement 首证(Popconfirm「已送达的站内信全部保留·公告归档为已下线」+行态翻);发布主链=R1 毕 |
| X2-3 notice 深链+用户信箱读 | X2-2 已发 | ctr 开信→类型 CTA 深链;admin 用户详情 messages Tab 读该用户信箱 | notice 形开信即读+CTA 键控(report→/report/$id;feedback_reply→/feedback/$id;纯公告无 CTA);ListUserMessages 五列表 | 深链目标页选择 |

### X3 激活码生命周期

| 例 | 前置 | 步骤 | 预期锚 | 校准点 |
|---|---|---|---|---|
| X3-1 铸码导出 | admin | 创建批次(类型三族/数量/有效期)→导出 CSV→码表明细 | 批次列表/创建流/导出=自身 trace 行;quantity 恒 1;码表逐码+进度条 | CSV 双口径(列表页全库未核销 4 列/批次页按筛选 6 列) |
| X3-2 输码解锁 | R-rich+X3-1 码 | ctr 认知报告锁态→服务端腿兑码(输码面未建=P4)→判决翻面 | 解锁=码>政策>默认折叠(Q6 J9);DB kind=report_type;核销 trail 落码行 | 入口形态属 P4(11-30 裁) |
| X3-3 码表回读 | X3-2 毕 | admin 码表刷新 | redeemed_by/at 落码行;进度更新;≥95% 补码黄条 | 补码提示动作形态 |
| X3-4 停用+幂等 | 新批次 | admin 停用批次→未用码兑拒;已核销不受影响;同码同员重呈=幂等答原授权;族幂等宽形(活授权在手,停用批未兑异码不烧) | 停用语义(Q5 I8+Q6 J7);幂等胜过停用(Q6-②) | 停用确认措辞(O-3) |
| X3-5 过期码拒 | 过期批次(SQL 造短有效期) | 无族授权成员兑过期码;活授权者重呈过期码 | 无授权=Expired 诚实文案;活授权=幂等站窗(Q6-② 修形;拒绝序 INVALID→族幂等→EXPIRED→DISABLED) | 拒绝文案语气 |

## 4) C 套 · C 端单面族

### C1 冷启动主链(一次性号全流程)

| 例 | 步骤 | 预期锚 | 校准点 |
|---|---|---|---|
| C1-1 登录族 | /login 双 tab+协议门槛+**手机号三时机校验(号段/满 11/失焦)+验证码 6 位与未发码先提交+协议半选态**+忘记密码入口+页脚合规行 | Q1-A1 全字面(含 ICP 带「备」);G-005 非微信注记;门槛未勾拒 | 真微信授权腿随真通道窗标注 |
| C1-2 向导三步 | 昵称(规则块/**emoji 拒/超长腿/计数红变**)→年龄(性别三态+出生+**u13 监护闸三腿:未勾拒/勾选落库/服务端 6100**,R-minor)→偏好(至少 1 款;**重复添加 toast/满 3 上限态/平票字面**) | Q1-A2/A3/A4 逐字;性别无默认强制选;helper ×1.5 句=已裁弃(D-AN⑤) | 服务端拒绝腿族(biz.go:200-215)顺走;**毕@R3(11-08,wt-minor-301/302)**:三腿全收——未勾拒=行内「未满 13 周岁须先取得监护人同意」滞留/勾选落库=DB 互证(guardian_consented_at 落+age_band under12+sex female)/6100=curl 直探(wt-minor-302 API 登录→UpdateProfile u13 无 consent)回 `ESPORTS_PROFILE_GUARDIAN_CONSENT_REQUIRED`;13-18 黄条(「未满 18 周岁——可以测评，但无法购买付费报告/会员（付费时会再次拦截）」无 checkbox)+平票(「类型投票 FPS 1 / MOBA 1 · 默认类型：平票 → 需手动选类型」)+重复 toast(「已经添加过了」)+满 3(「已添加 3 / 3」+添加钮 disabled)+band 派生行(2012-05→12-14 岁/2015-06→12 岁以下)顺路全收;**×1.5 句漏删=当场修**(D-AN⑤ 裁弃未落地,两 locale 删句,HMR 上页亲证);games 落库 DB 互证(2 官方+3 自定义) |
| C1-3 home 首观 | 向导毕落 /match(多数票)→tabbar 回 home | 续测卡=进行中才在场;空统计三零值;会员 greet 三态(X1 已证) | 首测引导 CTA |
| C1-4 编排族 | /match 列表→pick(**平票/正常/空池三变体;服务端拒答族 GENRE_UNKNOWN/no-drawable 对值导流**)→切换确认(**读后取消+真切换两腿**)→须知 | Q1-B2/B3/B4/B5 逐字;切换 rule2 分辙在册(N1) | 切换真腿+进度警告(done>0)变体=**R2 毕**(11-07,多次轮转);取消腿=R1 毕;空池活体变体仍待(P4) |
| C1-5 测评族 | intro→running(横屏+倒计时+暂停面+退出 sheet)→**切后台记无效面(patch 法造态)**→result(原始分收据+**较上局 delta 芯**) | Q1-C1/C2/C3;退出=BottomConfirmSheet 逐字;curve-pending 诚实 raw 收据 | R1 已证 5 款=red_dot/dual_zone/flash_aiming/tracking_moving/skill_trajectory(git 参照) |
| **C1-7 引擎余款扫尾(11-07 新)** | R-rich 号切换至 MOBA(**真执行 SwitchMatchGenre**)→逐款 intro→最低限完成(零输入可)→收据;覆盖 8 款:number_stars/target_recognition/multi_target/moving_target/offense_defense/resource_match/character_counter/economic_sim | 每款 detail 臂 schema=各自 logic 单测族+results.ts 行键;SubmitTestAttempt 臂形;未评级收据 | **毕@R2(11-07,沛然号)**:8 款全数首次活体执行,明细逐款收证(resource_match 六行含碰撞 2/offense_defense 四行含平均决策 2002ms+题库场景文案/multi_target 七行含错失高价值 11/number_stars 五行/target_recognition 六行含错失敌人 20/economic_sim 五行含终局血量 1500/moving_target 五行/character_counter 三行);抽签轮转=切换多次真执行;DB 双场互证(red_dot 0=C 判档,curvePending 全 raw-only);附带 Q1-⑥ 介绍文种子 live 首证(三款) |
| C1-6 完成与中心 | done 页→report-center 首观 | D-AI 分享主 CTA(parked MSG-2);四卡首观四态;0/6 横幅 | done 页=B-6 缺陷在案(三差),走查对照差距不另下判 |

### C2 报告族态集(R-rich)

| 例 | 步骤 | 预期锚 | 校准点 |
|---|---|---|---|
| C2-1 单报告态集 | 免费席/锁行 blur/解锁后/对比条历史最佳 | Q2-E5 全族;¥19.9=SKU 真价;**硬价在案=R-05**(R-04=terminated 呈现非硬价) | 锁行 blur 可读性边界;**毕@R2(11-07,沛然号 red_dot 真分记录直入)**:免费席(综合「—」同龄积累文案+反应 100 分+「同龄样本不足仅展示分数」+1/4 维)+锁行(「需解锁后查看完整内容」)+解锁后(墙翻「已解锁·单次解锁永久有效」;cohort 不足→「未评级」与分数并呈=Q2-⑤ 一致)+对比条(「当前 100|历史最佳 100|刷新历史最佳」持平支+「较上局对比|无上局记录|首测」delta 芯);center 卡 toast「还没有出分记录」=数据形非缺陷(LatestSessions 锚定每 genre 最新场,真分场被切换链老化,历史回看=test-records 未建面在案) |
| C2-2 认知报告 | 环形头+雷达+缺维说明+档位 | Q2-E4+Q2-⑤ 裁(环形+裸值无前 X%);<3 维退化文案;锁态=D-AB 默认 | 缺维说明引导性;**锁态毕@R2(11-07)**:「认知报告未解锁|会员或兑换码可开启」+去了解会员 CTA+免责声明全文+「查看能力成长趋势（进阶）」to-trend CTA;单次 trial 不开认知族(=record scope 非族)实证;**环形/雷达/缺维=解锁态,随会员窗造态批** |
| C2-3 成长趋势 | 空态→两点态(造 2 记录)→混类警示;**进阶分析模块+会员 CTA 偏离现场核(N-1 在 prd 待裁)** | Q2-E2 逐字;Δ chip 三色+首末收据;混类 banner 条件出 | N-1 裁后按裁走;**空态毕@R2(11-07)**:「还没有可比较的测评|趋势需要至少 2 次测评（或 FPS + MOBA 各一次）才能画出」+去测评 CTA;空态=诚实答非 bug(两点 scored 场 cohort 全 nil→`a point exists only where the cohort spoke`);**两点态/混类/N-1 现场=cohort 造态依赖(池=同龄+性别人群,dev 库无同龄样本),归 P3 造态批** |
| C2-4 职业指导 | 开态(D-AB 默认 open;策略面翻)+关闭态切换面 | Q2-E3 开态=旧仓语义级;关态预告逐字 | 处方呈现实用度;**毕@R2(11-07)翻面往返**:开态(gate open+1 维数据诚实空态「暂无职业方向数据」)+admin SetReportAccessPolicy 真翻 CAREER→CLOSED(首证:`policy:REPORT_ACCESS_POLICY_CLOSED` 日志+行留痕已覆盖/设置人/时间)+ctr 关态预告逐字(「暂未开放·敬请期待」+后台开关控制句+开启后预览清单+「不参与测满 8 维自动解锁」独立规则注+返回报告中心)+翻回开放+ctr 回开态;策略三选=开放/会员/关闭 |
| C2-5 出口族 | 返回钮/页头 back/深链直入 | **Q2-⑦ 裁**(返回个人中心→/profile;页头跟随来源,直入落 /match) | back 语义直觉;**毕@R2(11-07) 三支实证**:页脚「返回个人中心」→/profile;页头 icon back 跟随来源(SPA 来源=/pay/result 回 /pay/result);硬导航直入无来源 fallback 落 /match |

### C3 支付面(mock;真通道腿归 PAY2 联调刀)

| 例 | 步骤 | 预期锚 | 校准点 |
|---|---|---|---|
| C3-1 confirm 三态 | 非会员双卡/会员期(隐单次)/未成年红卡 | P7 三态+决策④+逃生双钮+缺出生回跳 | 三态切换入口清晰度;**非会员双卡毕@R2(11-07)**:报告锁面入口=?ref=record+双卡并列(单次 ¥19.9「解锁本次报告完整分析·永久有效」+月卡 ¥29.9「更划算」)双 CTA;月卡入口=X1-2;**未成年红卡毕@R3(11-08,随 X1-5 共刀)**:红卡替代双卡渲染;会员隐单次=会员窗造态批维持 |
| C3-2 收银台 | mock 三钮:支付/取消/失败 | MarkPaid 轮询收敛;cancel→cancel 态(R-03 在案) | 收银台拟真度;**三钮毕@R2(11-07)**:成功=X1-2+单次腿;取消=「已取消支付|未付款未解锁任何内容」+重新支付;失败=「支付失败|未完成付款未解锁未扣款」;**fail 单重新支付=同单复用首证**(订单号不变,「同一订单不会重复扣款」文案互证) |
| C3-3 result 面 | ok/member/fail/cancel 四腿+**直访无参 error 面(G-415 修后首帧)+fail 卡指路观察(B-8 在案)** | 四态 GetOrder 真值+收据四行;D-AU③ 双文案键;error 面不永转 | fail 态重试引导;B-8 裁后走;**四腿毕@R2(11-07)**:ok=X1-2+单次(「认知综合报告已解锁·永久有效」);cancel/fail 同批;error=直访无参「无法读取支付结果|订单不存在或网络异常」首帧即现不永转(G-415 修后 ✓);**B-8 现场再证**:fail 指路句「请稍后在「我的 → 订单」查看」——ctr 无该入口(订单读面=test-records 未建),在案缺陷维持 |
| C3-4 订单回读 | 毕单后 member-detail 权益行/报告判决/profile ③ 卡「已解锁 N 份」+admin 会员列表期内解锁列 | 期内解锁计数行(落点=profile ③+admin 列,matrix 校准 11-07) | **毕@R2(11-07) trial 形**:admin 订单详情「解锁去向=报告 <id> · 永久解锁」(foldUnlock ScopeTrial 分支实证)+ctr 报告页墙翻「已解锁·单次解锁永久有效」;月卡形回读=X1-2 admin 读回腿(期内解锁 0 份+profile ③)在案 |

### C4 账号族(R-multi/R-empty/R-rich/一次性)

| 例 | 步骤 | 预期锚 | 校准点 |
|---|---|---|---|
| C4-1 密码三态 | set(?m=set)/change/forgot 三腿+校验梯五连拒+**NoPhoneCard 死路卡(set∧无手机)+reason 词表(POLICY 造态:执法已落 88ada92a)** | Q3-① 全族(客户端五预检+**服务端组成执法=已落,按已落预期走**);forgot 三面 | **毕@R2(11-07,沛然号)**:校验梯五连拒逐字(11 位手机号/6 位码/至少 6 位/字母+数字组成/两次一致)+组成拒=客户端 zod 第六预检形;set=ResetPassword(SMS provider)code=200「密码已设置」+总面行翻;change=ChangePassword 错旧密码 400 SYSTEM_PASSWORD_MISMATCH 行内「旧密码不正确」(.text-danger-text 类名定位!通用关键词抓取曾误判静默)+正确 200「密码已修改」;forgot=发码→重设「密码已重置，请使用新密码登录」+自动切 /login;改密不撤会话(裁决兑现,SPA 续行);**POLICY 服务端执法=纵深防御不可达(客户端全拦,UI 路径无腿)**;**新缺陷 B-10:密码登录(手机号形)400 account not found——「手机号+密码」承诺假,SMS-born 死出口**(plan 工单);NoPhoneCard=真微信号依赖;**换账号钮毕@R3(11-08)**:admin 种子号密码登录 ctr 门→400 `SYSTEM_LOGIN_SEAT_REQUIRED error="identity holds no seat in population \"member\""`(服务端日志互证)→红行「该账号无此端访问资格，请更换账号」+「更换账号」钮(点击=表单清空+聚焦账号框+红行消);**P3-3「需 R-multi」断言修正=零双凭证依赖(admin 号即可造态)** |
| C4-2 绑手机 | SELF 态/冲突卡(占号)/**KEEP_* 空-占号升级 merge CTA(R-empty 新角色)**/跳过 toast | Q3 G3 全逐字+入口注记行;d8718fd3 冲突语义 | 冲突卡出路清晰度;**毕@R2(11-07)**:SELF 入口(「入口：个人中心 · 当前账号：沛然（手机号登录）」注记行)+跳过 toast「已跳过，可稍后在账号与安全再绑」+落 /;**占号冲突卡**(输 13977660011→ResolveCredential 预检 code=200→卡面不发码):「该手机号已绑定其他账号|本次绑定已拦截、未做任何改动」+双出路(验证码登录进彼号/彼侧解绑)+「为什么不能自动合并」链接;**KEEP_* 升级支毕@R3(11-08,wt-rich2 输空占号 13900000303)**:同一冲突卡+「检测到全新账号，去合并」升级 CTA(富对号无此 CTA=R2 对照互证;两空号对走 absorb-other 支方向不变) |
| C4-3 绑微信 | 非微信环境 toast/微信内授权(真微信腿随真通道窗) | Q3 G4;G-005 UA 隔离 | **非微信腿毕@R2(11-07)**:登录页「微信授权登录」→toast「请在微信中打开」(与页内预告行「当前非微信环境——点『微信授权登录』会提示『请在微信中打开』」互证);授权腿=真通道窗标注维持 |
| C4-4 合并四场景 | 可合并(R-empty 空号)/不可合并(双富号)/成功页/BLOCKED→反馈跳;**注:KEEP_OTHER 换边+切换尾页=真微信腿,随真通道窗** | Q3 G5 全族;凭证控制校验卡;merge_preview e2e 在案 | 富/空判定依据卡可理解度;**BLOCKED 面毕@R2(11-07,经 C4-2 冲突卡「查看说明」导航)**:/profile/security/merge「不可自动合并|两个账号都已有数据…不开放自动合并」+人工出路(意见反馈提交双号+说明,客服线下)+**判定依据卡**(「仅当一方为空号才开放——空号=无测评/会员/解锁/推荐」四要素)+「去意见反馈」CTA;**可合并+成功页毕@R3(11-08,keepCallerSms 全链,wt-rich2+wt-empty)**:确认页双卡(「注销（空号）139****0303」→「保留（富号）139****0305」+摘要行如实)+凭证控制校验卡(「已向空号手机发送验证码…两项校验通过才允许合并」)+「合并后空号立即注销，其手机号/微信释放；富号不变。此操作不可撤销，系统会留存操作记录」→收码→BottomConfirmSheet 双门→MergeAccounts→成功页「合并成功」;DB 三真值互证(account_merges poor/rich/kind=sms 行+credential 移籍+空号 enabled=f);**成功页文案错指微信=当场修**(body「微信已绑定到当前账号」在手机号吸收场景失真,四键两 locale 改中性形「全新账号已注销 · 其登录方式已并入当前账号」,HMR 上页亲证);**wt-rich2 的 done 场 match_session 未使 sessions>0=SessionCounts 按 test_record 口径**(走查观察在案);KEEP_OTHER 换边=真微信腿维持 |
| C4-5 安全总面 | 三行真状态+解绑说明卡+账号信息卡+**登出链(Logout+多标签踢出)** | Q3 G1 全逐字;登出=无 expired 通知刻意形 | **毕@R2(11-07)**:三行真状态(微信未绑「去绑定」/手机 139****0218「已绑定·管理」/密码 set 前「未设置·去设置」→set 后「已设置·修改」两态全见)+解绑说明卡(「需人工核验后处理，请联系客服」)+账号信息卡(昵称/性别/出生/偏好)+顶部同绑说明句;**登出链**:profile「退出登录」→alertdialog「确定退出登录？|退出后需要重新登录。你的测评记录与已解锁报告都会保留」→确认→/login;SMS 重登恢复「登录成功」落 /;多标签踢出=单会话 cookie 模型(底座裁决域) |
| C4-6 多凭证掩码 | R-multi 双凭证显示 | G-203 已裁保持现状——采集现形存证 | 现场目检表态;**毕@R2(11-07)现有形**:手机掩码 139****0218(安全面+profile ③+admin 三处一致)+密码行已设置态(「已设置——可用于账号密码登录」,文案入 B-10 证据);手机+微信双凭证并陈形=真微信腿依赖(P3-5) |

### C5 消息反馈 ctr 细态

| 例 | 步骤 | 预期锚 | 校准点 |
|---|---|---|---|
| C5-1 提交校验边界 | 六类枚举/10-500 字逐字/**contact=契约 128/输入 40(PM-14)** | feedback 六类+校验逐字(demand-ledger 27 行) | 字数计数呈现;**毕@R2(11-07)**:六类枚举逐字(功能异常/测评体验/报告内容/付费与会员/账号问题/其他建议)+计数「N / 500」+contact maxLength=40(PM-14 ✓)+「留空也可以——站内消息回复」注+状态流转说明(处理中→已回复→追问回处理中);15 字真提交二次活证(「已提交 · 状态：处理中」+列表行处理中 chip·日期);**10 下界拒绝形缺失=客户端拦截零反馈面**(5 字提交:toast/行内/计数/钮全静默,日志证 SubmitFeedback 未发)——轻缺记 P4(N-5 族) |
| C5-2 信箱双 Tab | 消息/我的反馈切换+空态;**顺带证 home 红点腿(X2-2 失真句销案)** | /messages 双 Tab ?t=;空态句 | 空态引导;**毕@R2(11-07)**:?t=msg↔?t=fb 双 Tab 切换+「1 条未读」横幅(在案公告信)+反馈空态「还没有提交过反馈」+「我要反馈」CTA→真提交后非空态(行=内容摘要+「处理中 · 11-07」chip);home 红点=tabbar「消息 1 条未读」徽在案(P0 批 snapshot+本批复核) |

### C6 变体扫尾

| 例 | 步骤 | 预期锚 | 校准点 |
|---|---|---|---|
| C6-1 续测恢复 | 进行中 session 中断(关页)→回 home 继续 | 续测卡已进行到 X/N+继续;进行中锁死随落库(README §0.5) | 卡片 vs 整页(G-208 P 票);**毕@R2(11-07)**:home 续测卡「有一场未完成的测评|进行到 1/4·开始于 2026-11-07（已过 1 天）」+「继续」→/match 编排页恢复(「已测 1/4」+逐项机会态「已用完/剩 1 次」=进行中锁死随落库);硬导航重开=多轮实证(全程导航链后卡仍在) |
| C6-2 重测语义 | 完成后再测同款/切换类型重排/机会用尽置灰 | 成绩追加不覆盖、展示取最高;切换按目标类型重排、同款可再现(README §0.5);rule2 分歧在案 | **毕@R2(11-07)**:①追加不覆盖=R1 dual_zone 双局在案 ②取最高=R1 并列双标注 ③切换重排+同款跨场再现=R2 轮转 live(red_dot/number_stars/skill_trajectory 多场再现);机会用尽置灰+mt-F1 双出口=R2 live;须知二次面=R2 live(「每款 1 次重测机会」变体+跨场次历史最佳口径卡) |
| C6-3 合规族 | age-required/协议壳 | Q1-A5/A6;版本行真源;逃生口=遗留清单在项 | **毕@R2(11-07)**:age-required=「补填出生年月|付费前必须补全|你之前只选了年龄段（次级口径），付费需要精确到月的出生年月」次级→精确口径闸+出生年滚选;协议壳=版本行「v1.0.0 · 2026-08-19 生效 · 演示文本，正式版以法务定稿为准」(真源+演示注双具)+13 周岁监护/18 付费拦截句;逃生口=遗留清单在项维持 |
| C6-4 通用态族 | 断网造 error/404 直达/loading | network 暖卡=Q1-③(三钮);404 品牌=ctr | **404 腿毕@R2(11-07)**:乱 URL→「PAGE NOT FOUND|404|页面不存在|您访问的页面不存在或已被移动」+返回首页/返回上页双钮,ctr 品牌形;loading=全程各页 loader 多轮见;断网暖卡=CDP offline 造态,顺下批(P4) |
| C6-5 主题与横竖屏 | 双主题抽检+running 横屏复核 | 双主题全站法;running=唯一横屏 | 暗色面对比度;**毕@R2(11-07) 标注**:ctr 无主题切换钮(仅语言钮)——暗色=全站单主题形,R2 全程暗色面走查覆盖人工轴;running 横屏=R1 已证(E7/H-A 在案)维持 |
| C6-6 服务端拒绝腿族 | UpdateProfile 拒绝面(昵称服务端校验/cap 超/unknown id)/GetMatchSession 拒答族 | biz.go:186-215 拒绝词表;6100/6131 族 | **毕@R2(11-07) 判层**:UpdateProfile 拒面=**客户端拦截纵深防御形**(emoji 拒「昵称暂不支持 emoji 表情」行内,UpdateProfile 零调用日志证;超长 maxLength=24 硬截同形)——服务端 6131 词表=不可达兜底(P3-7 服务端直探批);**直探毕@R3(11-08,curl 带 session token)**:UpdateProfile u13 无 consent→`ESPORTS_PROFILE_GUARDIAN_CONSENT_REQUIRED`(6100 族)+GetMatchSession 无偏好票→`ESPORTS_MATCH_GENRE_UNDECIDED`/genre 非法→`ESPORTS_MATCH_GENRE_UNKNOWN`;no-drawable 造态支维持 |

## 5) A 套 · admin 单面族

| 例 | 步骤 | 预期锚 | 校准点 |
|---|---|---|---|
| A1-1 SKU 改价 | 改价 Modal | O-3;listPrice=price 无折扣(H4) | 二次确认词面 |
| A1-2 上下架同步 | 下架→ctr 支付面不可达 | 上下架即时同步(Q5 I5) | ¥— 禁用形 |
| A1-3 新建 SKU | 建档表单(类型条件会员期字段;新建即上架) | sku-form-modal 字段集;原型 edit=改价非编辑 | — |
| A2-1 users 读面 | 列表+筛+出生行+详情六 Tab | Q5-① 词族;O-1/O-2 | — |
| A2-2 members 四 Tab | 全部/有效/≤7 天到期/已过期+来源筛+续费 | admin-member-anchor §四六钉;ratchet=新行 exp 旧到期+31d | 到期 Tab 运营衔接 |
| A2-3 会员详情 | 有效期/来源/续订史 append-only+**终止 Modal(X1-4 共刀真执行)** | 同上;O-3 终止五语言 | **毕@R2(11-07,X1-4 共刀)**:detail 页满态(剩余 62 天/至 2027-01-08/累计 2 期/当前会员三行=套餐·有效期·来源手动开通/期历史 append-only 两行+「终止只改状态不删行」注) |
| A3-1 orders 列表 | 筛(解锁方式列头+筛)+空态 | I6 列集+状态词 | — |
| A3-2 订单详情+关闭+**退款腿** | 八事实行+时间线+关闭(O-3);**退款:原因必填弹窗→PROCESSING→读穿收敛(0.01 标本流程;e2e 三臂在案)** | I7+退款口径;RefundPaymentOrder UI 面 | 退款权限面;**退款腿毕@R2(11-07)**:八事实行(含解锁去向「11-07~12-08 窗口·期内解锁 0 份(永久保留)」+I7 口径 Alert)+原因必填 Modal(空因禁 ok;「解锁与会员剩余有效期不随本操作自动回收」明示句)+受理 toast「退款已受理」→RefundFacts 收敛(已退款 ¥29.9+退款单号 RPdbn…·mock-rf id+审计原因落表);日志 `order refunded by operator state=succeeded`+CheckOperation 先决+reason 全文入日志;PAID+未退 gating 双侧一致(退款后全额退款钮消失) |
| A4-1 协议版本流 | 建即当前/切换/删除拒当前版 | E9 走查口径(I10) | — |
| A4-2 同意记录 | 预填/昵称搜索/CSV 掩码导出 | I11;渠道列=R-07 在案 | — |
| A4-3 合规配置 | 三卡+年龄线敏感 Modal+**到期提醒 SetRenewalReminder 写腿(调整→读回;全链零执行 RPC,本次承载)** | I12 逐字;callout 逐字 | **三写腿毕@R2(11-07)**:SetRenewalReminder 7→14→7 往返(args=days:14 日志互证)+SetPayAgeLine 18→19→18(敏感 Modal「需权限+二次确认+审计」徽)+SetCompliancePrompts minor_prompt 覆盖写(「文案已更新」;空串 SQL 净还原,行模型=空串默认);恢复默认手势=遗留维持 |
| A4-4 出生开关 | 开关翻转+ctr 侧现状互证 | 消费面未接=遗留清单 | — |
| A5-1 报告策略 | 四值×报告族读写 | I13;tenant 拒写在先 | 设置人水合小缺在案;**写腿往返毕@R2(11-07,随 C2-4 共刀)**:SetReportAccessPolicy CAREER 开→关→开(`REPORT_ACCESS_POLICY_CLOSED` 日志+CheckOperation 先决)+行留痕(已覆盖/设置人/时间)=覆盖旗设计形 |
| A5-2 game-config | 平键值表+题库抽屉+**调参 modal 开+读+写(GetTestConfig/UpdateTestConfig)** | I14;R-06 在案 | 具名列现场表态;**调参 modal 毕@R2(11-07)**:读=GetTestConfig 首证(modal 预填后台失效阈值 10s+引擎参数 4 对 key-value 表+「保存时整体替换」语义句);写=UpdateTestConfig 首证(target_size_px 20→24→20 往返,`value:"24"` 日志互证;静默无 toast 形);「确 定」全角空格钮 |
| A6-1 导航与底座在场 | 九组展开+IAM 六叶+死链检;**locale 前缀族观察(U3 承载)** | G-301/302 落位;IAM=骨架车道 | href 前缀四实例并 U3 |

## 6) 在场检查(未建面;只验诚实降级)

ctr:test-records(404 形)/share-invite(parked toast)/scheme-b×3(无路由)/version-update(未建);admin:余组 13+M3 2 页菜单不出现死链。出处=demand-ledger §一/§二。

## 7) 边界(本矩阵不模拟)

时钟触发(converger 到期 fold/日对账/订单自过期)=e2e 已钉,仅 DB 态观测;微信真通道三半绿条=PAY2 联调刀;IAM 底座六面=骨架车道;真微信腿(登录授权/绑定/合并回调族)=随真通道窗专项轮。

## 8) 计数收口

- 用例 **58** = X 13(X1 5+X2 3+X3 5)+ C 30(C1 7[含 C1-7 引擎扫尾]+C2 5+C3 4+C4 6+C5 2+C6 6[含 C6-6 服务端拒绝腿族])+ A 15。(2026-11-07 重立:56 例设计承袭 R1 轮,按覆盖核验缺口新增 C1-7/C6-6 两例、多处补步;锚失真 9 处已修=coverage-map §四)
- 覆盖账=coverage-map.md(58 例对 45 已建页/85 RPC;豁免与缺口同档)。
- 面覆盖:ctr 已建 34 面逐面入例;admin 已建 16 页逐面入例。

## 9) 对账环

- 轮次:R1 已收割(git);当前=**R2 起(正式测试阶段)**,态列全域未走;每例毕记「毕@R<n>+日期」;跨轮复走按面变更重置相关例。
- 每例开工先 join:coverage-map 缺口清单+plan 缺陷工单(B 系)+prd 待裁(偏离 N/PM 系)——三台账在案项不得另下新判。
- 每例毕:态列更新;发现按 §1 四路处置;用户偏好即时落 prd。
- 每套毕:净清声明(计数,必含 match_session);本档+coverage-map 同步。
