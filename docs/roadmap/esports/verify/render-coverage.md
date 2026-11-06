# 渲染可达性台账 — U1 渲染可达性审计(UI 校准役第一段)

> 建立:2026-11-06(用户令,插队于旅程走查前)。目的:两客户端**所有 jsx 文件级可达+分支级可活证**,防死码;用户令原文=「确保客户端所有的 jsx 在满足条件时都会渲染」。
> 法条:死码判定双档律——**永不可达才算死**(含 remote 场景推演),零引用≠死;预留门(登记在案的 parked/gated)登记不删。射面=ctr 全客户端+admin esports 模块;骨架 IAM 面排除(Q 线法条)。
> 顺序(用户 11-06 三裁):U1 ctr→U1 admin→U2 ctr 原型对齐(harness 重激活)→U3 admin 对齐基座范式(IAM/文件/平台/智能体=参考实现)→回旅程矩阵(X 停于 X1-1)。

## 1) 文件级 · ctr(2026-11-06 毕)

**盘点(机器)**:routes 50 tsx(47 产品路由+sitemap/robots/api/root)+features 115 tsx+domains 1 tsx=166;文件数=本节核验对象数。

**方法与命令形状(复跑照抄)**:
- 路由注册:`grep -cF '<route file basename>' routeTree.gen.ts`(文件约定加载,无需 import;`$` 参数路由须 `-F` 字面量——ERE 下 `$`=锚符,首扫 5 条假孤儿即此伪影)。
- 孤儿扫描:`grep -rFl -- '<basename>' .` 自排除后计数=0 报嫌疑;自检锚(member-card,已知活)=2 引用过。陷阱:ugrep 的 `--include` 置于 `--` 后被当文件名,include 失效=扫描面变宽但方向安全(更少假孤儿)。
- 重名互掩:重复 basename 逐份以 import 行实证(`grep -rn -E "from ['\"].*<base>"`);相对路径 import 不含家族前缀,路径启发式不可用(实证翻车),以 import 行为准。

**结果**:
| 层 | 数 | 判定 |
|---|---|---|
| routes 50 | 50/50 注册于 routeTree.gen.ts(含 5 条 `$` 参数路由逐条核) | 零死路由 |
| features 115 | 孤儿扫描 0 嫌疑;重名 5 组(detail/filing-footer/member-card/page/wechat-callback)逐份 import 行实证全活(page×12=每模块 1 importer,detail 1/1,filing-footer 2 份 3 importers,member-card 2 份 2 importers,wechat-callback 1/2) | 零死文件 |
| domains 1 | 未被扫描标记 | 底座(骨架车道)不计 |

**ctr 文件级收口:166 文件,零死文件,零未注册路由。**

## 2) 分支级 · ctr(进行中;记账单位=**渲染门**,非路由页)

> 模型(2026-11-06 用户校准:交互门一等公民——弹窗/toast/错误提示等**默认不渲染、事件后才渲染**的面是「写完从未走到」的高危区;MSG-1 V10 五伤全在此层,页面级走查全漏):

| 门类 | 渲染时机 | 活证方式 |
|---|---|---|
| **交互门** | 事件触发:Modal/Sheet/Toast/确认框/banner/notfound | 走查中真开一次(点触发钮/输入触发校验/造真错误) |
| **状态门** | 数据态:空/加载/错误/会员/未成年/锁/未读/过期 | 造态后读面 |
| **路由门** | 守卫/重定向(beforeLoad/profileFlowGuard) | 直达 URL 验落点 |

> 列:门实例 | 触发条件 | 可达判定 | 活证/处置。判定词:**已活证**(引证据)/**待造态**/**数据门**(数据在场才渲染,合法)/**预留门**(登记 parked)/**观测受限**(条件不可拟真,须注替代证据如 e2e)/**死码嫌疑**。
> **交互门普查(机器,2026-11-06)**:命名门组件 6 文件(app/dialog 两件+match/switch-confirm+pay/state-banner+profile/logout-sheet+test/invalid-notice);调用点级=toast/sonner 31 处+BottomConfirmSheet 12+Dialog 8+ExitConfirm 5+错误面族 7(ErrorBoundary/errorFallback/errorPage/notfound)+Modal 3;Popconfirm/Drawer=0(RAC 栈)。门实例逐调用点记账,toast 各点须注其触发事件(错误路径 toast 的活证=真造该错误)。
> 提取批次:壳(错误面族+states+dialog)→login→向导→home→match→test→report→member/pay→messages/feedback→security→agreement/age;每族含路由门+状态门+交互门三类行。

### 2.0 应用壳(app/*,跨面承重)——提取毕 2026-11-06

| 门实例 | 触发条件 | 可达判定 | 活证/处置 |
|---|---|---|---|
| ErrorPage·network 分支 | root loader 抛 network 错(服务不可达) | 待造态 | 活证=停 esports 刷页;暖卡+三钮=Q1-③ 裁已落(桌面),live 未走 |
| ErrorPage·非 network 非 silent | loader 抛 4xx/5xx | 待造态 | ERR chip+大 HTTP 码+重试/回家;活证=造真 5xx(临时 SQL/拦截) |
| ErrorPage·silent 支 | 静默错(abort/cancel) | 观测受限 | 竞态难拟真;view() 词表桌面承载 |
| ErrorFallback·network→NetworkErrorFace | 组件级抛 network | 待造态 | 双宿主之一(Q1-③);随上同刀活证 |
| ErrorFallback·silent/非 silent 两支 | 组件级抛错 | 待造态/观测受限 | 同上 |
| BaseErrorBoundary 宿主 | 组件树抛错 | 待造态 | resetKeys=leafRouteId;活证=任一面造组件错 |
| NotFoundPage | 未匹配路由 | 待造态(易) | 直访 /nope;404 品牌=ctr(G-009) |
| StateScreen·empty(Inbox) | 内容区数据空 | 待造态 | 空态面多处(C5-2 消息空/报告空) |
| StateScreen·loading | 页内加载 | 观测受限 | 时序门;dev throttle 可试 |
| LoadingStrip·inherit/ink 双 tone | PendingFallback(inherit)/StateScreen(ink) | 已活证 | 每路由 pending 常见态 |
| **Skeleton(states/skeleton.tsx)** | — | **已判(11-06):删** | 唯一引用=barrel 再导出(features/app/index.ts:11),零消费方零 e2e;现行 loading idiom=LoadingStrip(活,12+ 点)+StateScreen,Skeleton 注释自述的 host 组合用法与之重叠=同位置两 idiom 竞争;Q1-③ 裁弃残骸且无预留登记(「预留能力不删」前提不成立);判据=孤儿组件是 idiom 分裂种子,后来者 import 即分裂第二套 loading 观感。删组件+barrel 行,git 可寻回;归开发会话执行 |
| BottomConfirmSheet(12 调用点) | 各触发钮(退出/登出/切换/合并/支付 sheet…) | 部分活证 | **2/12**:merge 双闸全脸(11-06)+**logout sheet 确认钮实证**(11-06 晚:入口+sheet 内第二「退出登录」钮双击链;sheet 面文未捕,复走补);余点逐族真开;visual 槽=QR 码图消费点(G-403) |
| ExitConfirmDialog(挂载条件化) | open 受控(RAC exit 死锁律,注释在案) | 已活证 | test/attempt.tsx 单测引用;running 退出活证随 C1-5 |
| FeedbackSonner·Toaster host | 常驻非门 | 已活证 | toast 实例 31 点逐点随族记账 |
| PendingFallback·route pending | 路由 loader 进行中 | 观测受限 | throttle 活证;mono 条=维持 idiom(Q1-③) |
| sessionGuard·unreachable 支(补遗 11-06) | session 探测 unreachable | 待造态 | 停 esports 后直访受护页→throw network envelope→接 ErrorPage·network(§2.0 首行同刀活证);removeQueries 防僵尸 |
| sessionGuard·anonymous 支(补遗 11-06) | 匿名直访受护页 | **已活证** | 11-06:/messages→`/login?next=%2Fmessages%3Ft%3Dmsg`(**next 带全 href 含 search 参数**);登出踢出标签页同形(/login?next=%2F);routeAfterLogin·next 回环支仍待随链 |
| sessionGuard·expired 支(补遗 11-06) | 死 session 刷受护页 | 待造态 | SQL 删 session 行;→login?expired=true&next=,接 login toast·session_expired |
| profileFlowGuard·未完成支(补遗 11-06) | authed+profile 不完整 | 待造态 | →向导 nickname;守卫面=match+test.$code+report.* 共 7 路由(grep 实证);新鲜匿名访不带 expired(仅死 session 才带) |

守卫补遗注:守卫挂点=member/pay/age-required/profile(含 setup 子树)/messages/feedback 走 sessionGuard,match/test.$code/report.{center,career,cognitive,trend,$recordId,session.$sessionId} 走 profileFlowGuard;__root beforeLoad 仅预热 session+lang/dir 非门。completeness 法=nickname+生日对,**sex 不在门内**(P8 前),但 age 步本地强制选性——契约法条非 gap。

壳批次注:codegraph blast radius 报 ctr Skeleton 有 13 admin caller=**同名合并伪影**(admin 用 antd Skeleton,import 行实证),查消费方一律以 import 行为准。

### 2.1 login 族(routes 4 + login 模块 11 tsx;提取毕 2026-11-06)

> 族文件:routes/login{,.index,.forgot,.wechat.callback}.tsx + features/modules/login/(page、wechat-callback、ui/×9:agree-box/-gate、brand-lockup、form-error-line、login-footer、mode-segment、password-fields、sms-fields、wechat-button)。状态机=sms.ts(useSmsLogin)、落地路由=after-login.ts。/login/forgot 的页面组件(ForgotPasswordPage)住 security 族,归 security 批提取。
> 族内零死码:ui×9 全被 page 消费;sms.ts 三导出(mainlandMobile/phoneLiveInvalid/smsFailureCopy)被 security 族 use-sms-code/use-bind-sms/bind-phone/password 复用(import 行实证)——跨族共享件非死码。routes/login.tsx 布局壳注释在案的陷阱(无 Outlet 子路由匹配不挂载=静默死渲染)已由 Outlet 承载,属已防护不属缺口。

| 门实例 | 触发条件 | 可达判定 | 活证/处置 |
|---|---|---|---|
| routeAfterLogin·next 支 | 登录带安全 ?next=(同源路径,协议相对 `//` 拒) | 待造态 | 活证=登出态直访受保护区→guard 弹回带 next→登录→回原页 |
| routeAfterLogin·未完成支 | profile 不完整 | e2e 在案 | fullchain 链已走(向导落地);live 复走随 U1 活证轮 |
| routeAfterLogin·bind 提示两支 | 无手机凭据→bind-phone?from=login;微信内无微信凭据→bind-wechat(localStorage `ctr.bindprompt.<uid>` 一次性旗) | 待造态/观测受限 | 无手机支=SQL 造仅 wechat_mp 凭据账户;微信内支须 MicroMessenger UA |
| /login validateSearch·expired 支 | 401 watcher 携 expired=true 落地 | **已活证** | 11-06:直访 ?expired=true→toast「登录已过期，请重新登录」+**param 自剥**(URL 归裸 /login);watcher 触发路径(真 401)待 SQL 删 session 刀 |
| /login/wechat/callback validateSearch | code/state/next 透传 | **已活证** | 11-06 codeless:直访无参→failed 面「微信登录失败,请重试」+返回登录钮(default 支) |
| ModeSegment·password 支 | 切 tab(默认 code) | 已活证 | X1-1 密码登录走过 |
| SmsFields·phoneErr/codeErr 红边 | 错运营商前缀/非法 11 位;空码 | 部分活证 | 11-06:phoneErr 活(「12345678901」→「手机号格式不正确」+input danger class);codeErr 支未单捕 |
| SmsFields·countdown 60s 支 | 发码被受理(state 落地) | **已活证** | 11-06:「60s 后重发」(mock 后端;同挂载两段式发码+redis 读码+提交=完整 sms 腿复走) |
| WechatButton·wechatEnv=false 支(桌面默认) | UA 非 MicroMessenger(post-mount 判,SSR 同暗态防水合裂) | **已活证** | 11-06:env note 行全文在场(「当前非微信环境——点『微信授权登录』会提示『请在微信中打开』」)+点击 toast「请在微信中打开」;半透明视觉未截(小项) |
| WechatCallback·pending 面 | effect 进行中 | 观测受限 | 瞬态;真 OAuth 才驻留 |
| WechatCallback·error 面 3 支+回登录链 | failed/serverReason(USER_DISABLED/TENANT_DISABLED/default) | 待造态(易) | 直访无参=default 支;DISABLED 两支须 SQL 停用账户/租户 |
| busy 态钮(pending/sending/sending) | 三腿提交进行中 | 观测受限 | 瞬态;dev throttle 可视 |
| FormErrorLine 渲染门(8 支 copy) | copy!==null 才渲染 | 待造态(易) | 错账密=LOGIN_FAILED;短密<6=short;非 System reason=generic |
| FormErrorLine·换账号钮 | **双条件**:password 模 ∧ LOGIN_SEAT_REQUIRED | 待造态 | 需 R-multi 多座位号;点击=表单重置+聚焦账名。注:code 模同 reason 只出文案不出钮(原型换号跳转=密码腿专属),观察非 gap |
| AgreeGate·missing 红字+AgreeBox 三态 | 缺选提交;单勾 | 部分活证 | 11-06:missing toast 活;单勾 indeterminate(Minus)态未造——**编舞法定案:checkbox 本体按行文本关联 .click()**(行 label 点击不可靠) |
| toast·agree_missing(error) | 三腿任一在缺选时触发 | **已活证** | 11-06:「请先阅读并同意《用户协议》 和 《隐私政策》」(**locale joiner「和」实证**) |
| toast·session_expired | expired 支落地(同上路由门) | **已活证** | 11-06:「登录已过期，请重新登录」;id 去重防复 toast 机制面未见复触 |
| toast·code_send_first | 未发码直接提交 | 待造态(易) | 编舞窗失误未捕(缺协议时 agree toast 优先级遮蔽) |
| toast·code_sent(success) | 发码受理 | **已活证** | 11-06:「验证码已发送」 |
| toast·login_success(password/sms 腿) | 登录成功 | **已活证** | X1-1 密码腿+**11-06 sms 腿「登录成功」**(落地=/ 默认首页,与向导毕落 /match 分野实证) |
| toast·wechat_env(error) | 桌面点微信钮 | **已活证** | 11-06:「请在微信中打开」 |
| WechatButton·onError→generic 错行 | begin authorize 抛错 | 观测受限 | 停 system 可造 |

族小计:22 门=路由门 5/状态门 7/交互门 10;**11-06 登出态批收割后**:已活证 12(expired 自剥/session_expired/callback codeless/wechatEnv 支+toast/phoneErr/countdown/agree_missing(joiner)/code_sent/login_success sms 腿)+部分活证 3(FormErrorLine 待/AgreeBox indeterminate 待/code_send_first 待);守卫匿名支(§2.0)同批活证。toast 位 8/31 live。**走查法沉淀:checkbox=本体 .click() 按行文本关联;登录协议行 label 点击不可靠**。

### 2.2 向导族 /profile/setup(routes 4 + setup 5 tsx;提取毕 2026-11-06)

> 族文件:routes/profile.setup.{nickname,age,games,wechat-callback}.tsx + features/modules/profile/setup/ 同名 4 件+step-frame。StepFrame(进度轨三态+步数+back fab+每步 orb 美术)=三页共享壳,随页活。族内零死码(step-frame 3 消费方;文件级已核)。
> 关键法条:守卫=父层 /profile 布局 sessionGuard(setup 三页自身无 beforeLoad,吃父层);completeness=nickname+生日对,sex 不在门(**11-06 勘误:age 步并不强制选性——空性别过步实测定性,「*」虚诺立 G-412**);向导毕落 **/match** 非首页(原型法,**live×2 已证**)。

| 门实例 | 触发条件 | 可达判定 | 活证/处置 |
|---|---|---|---|
| 父层 sessionGuard(setup 全子树) | 匿名直访任一 setup 页 | 待造态(易) | →login?next=(§2.0 守卫补遗同刀) |
| games·onDone 成功落 /match | 完成向导 | **已活证** | 11-06 live×2(完成→/match;含编辑路径重提交同链) |
| setup-callback·codeless 支 | 直访无 code/state | **已活证** | 11-06:toast「微信授权失败，请手动填写昵称」+弹回 /profile/setup/nickname(迷路书签诚实路) |
| nickname·微信 pill | wechatEnv ∧ 无 parked 档案 | 待造态 | chrome mcp emulate UA=MicroMessenger 即渲染;非微信浏览器永不渲染(设计如此) |
| nickname·prefill「来自微信」行 | parked 档案 sanitize 后非空 | 待造态 | sessionStorage `ctr:wechat-profile` 直注+刷新即造(evaluate_script) |
| nickname·note 两支 | 全 emoji sanitize 空/截断改写 | 待造态 | 同上直注(昵称全 emoji→empty 支;带 emoji→trim 支) |
| nickname·计数器三色 | len>12 红 / 2-12 绿 / <2 灰 | 部分活证 | 11-06:数值态三档活(1/12·绿区/13/12)+**超帽红实证**(13/12 计数器本体转 danger 元素);灰/绿色 class 未逐帧捕(视觉小项);maxLength=24 超帽法 live |
| nickname·error 4 支 | 首尾空白/含 emoji/>12/<2 | **已活证** | 11-06:**四支全词面**(「昵称至少 2 个字符」/「昵称暂不支持 emoji 表情」/「昵称首尾不能有空格」/「13 / 12」计数器红)=**提交时校验非输入时**(emoji 输入态零红行,提交才出);**编辑路径预填支顺收**(阿包在场→2/12) |
| nickname·pending 钮 | isPending | 观测受限 | 瞬态 |
| age·band 派生行 | 年月齐→ageOf→bandOfAge | **已活证** | 11-06 三值:「已按出生年月自动带出年龄段:30-40岁/12岁以下/15-17岁」 |
| age·u13 黄条+监护人 checkbox | age<13 | **已活证** | 11-06 全脸:黄条「未满 13 周岁——需监护人陪同并在监护人同意下使用;实名信息不入测评库。」+监护人 checkbox 全文+**合规留痕注**(「勾选记录(时间戳+协议版本号)单独留存用于合规举证,与《用户协议》勾选分开记录」) |
| age·u18 黄条支 | 13≤age<18 | **已活证** | 11-06:无 checkbox 版「未满 18 周岁——可以测评,但无法购买付费报告/会员(付费时会再次拦截)。」(另一 JSX 分支实证) |
| age·error 3 支 | 未选性/生日不全/未勾监护 | 部分活证 | 11-06:监护支「未满 13 周岁须先取得监护人同意」活;**未选性支实测不拦→G-412**;生日不全支未单独造 |
| age·sex 预填(typeof 窄化) | wire 名串→枚举映射 | 部分活证 | 11-06:**nickname/生日编辑路径预填实测**(阿包 2/12;生日预填=间接证:空性别提交推进到 games 步,生日必在场);**待复核观察:空性别提交未拦即过步**——「age 步本地强制选性」法条疑涉,profile 终态性别仍未填(服务器未受污),需干净复走定性(法条 gap vs 复走操作伪影) |
| age·pending 钮 | isPending | 观测受限 | 瞬态 |
| games·分组渲染 | 目录 category:fps/moba 有头,余类裸网格 | 已活证(面) | 11-06:FPS 组头(CS2/无畏契约/守望先锋/APEX/和平精英)+MOBA 组头(王者/LOL/DOTA2/平安京)+「其他游戏」裸组三形齐;StepFrame 3/3 轨第三态同捕 |
| games·cap 满行+添加钮 disabled | customs≥3 | **已活证** | 11-06:「已添加 3 / 3 · 已达上限，可先删除」+添加钮 disabled 实证;**观察:超帽 toast 经 disabled 按钮不可达**(防患设计;Enter 路径存疑) |
| games·chips 区+chip 两态+× 删 | customs>0 | **已活证** | 11-06:chips 区渲染+chip 本体=aria-pressed toggle 钮(切投票)+**× =chip 行内 svg 兄弟件**(SVG 无 .click(),MouseEvent dispatch 法;三删实证 3/3→0/3) |
| games·票数行 | totalPicked>0 | 部分活证 | 11-06:**平票支 live**(「已选 2 款 · 类型投票 FPS 1 / MOBA 1 · 默认类型:平票 → 需手动选类型」);un 版文案/单数默认组别支待造(切换目录选择即可) |
| games·error 支 | 0 选提交 | **已活证** | 11-06 干净复走:「至少选择 1 款游戏」拦截+停留本页(首次「通过」假象=走查 finder 失配非校验缺;完成钮保持 enabled=提交时校验,与 nickname 同法) |
| games·pending 钮 | isPending | 观测受限 | 瞬态 |
| toast·games_dup/cap_toast/added_toast(3 点) | 重复加/超帽加/添加成 | 部分活证 | 11-06:dup「已经添加过了」+added「已添加并选中:…(未归类,不计类型票)」活;cap_toast 经 disabled 按钮不可达(见 cap 行观察) |
| toast·nickname pill bind_failed | begin authorize 抛错 | 观测受限 | 停 system 可造 |
| toast·setup-callback bound/bind_failed(3 点) | OAuth 回三支 | 观测受限 | 真微信链;codeless 支已易造(上表) |
| setup-callback·pending 面 | exchange 进行中 | 观测受限 | 瞬态 |

族小计:25 门=路由门 3/状态门 16/交互门 6;**11-06 直访群收割后**:已活证 11(nickname error 四支/band 三值/u13 全脸+合规注/u18 版/onDone×2/codeless/setup 分组面/0 选 error/cap 满/chips×删)+部分活证 6(计数器色票数行 un+平票+默认组别支/dup+added toast/sex 预填勘误);待造态 5(nickname pill/note 两支/pending 钮×3/toasts 真微信腿/生日不全支)、观测受限 7。toast 位新增 3 点 live(setup-callback/dup/added)。**法条勘误一条:age 步不强制选性(G-412)**。

### 2.3 home 族(routes/index 1 + home 模块 5 tsx;提取毕 2026-11-06)

> 族文件:routes/index.tsx + features/modules/home/(page、ui/×4:cta/resume/recent/stats-card)+ recent.ts/resume.ts 纯推导层。**横切法条(诚实零)**:读回卡缺席=诚实无数据面,guest/读失败/薄 cohort 永不渲染零冒充分数(composite 无→「—」,tier none→无 chip,abilities 无 percentile→strip 整条不渲染,absent→「—」非 0)。home **零 toast**,错误全部静默成卡缺席。
> 死码:零(4 ui 全被 page 消费;recent/resume helpers 有单测家园 resume.test.ts);TierChip(test/ui)跨族引用,归 test 族记。

| 门实例 | 触发条件 | 可达判定 | 活证/处置 |
|---|---|---|---|
| onStart 三支落点 | 未登录→/login?next=/;无 games→/profile/setup/games;齐备→/match | 待造态(易) | 三种号态各点一次:登出态/R 无游戏号/R-rich |
| greet 三态 | guest/logged/member(memberUntil 非 nil=presence) | 已活证 | X1-1 会员翻转走过 logged→member;guest=登出即见 |
| 昵称行两支 | authed+nickname→昵称;否则访客名 | 已活证 | X1-1 |
| ResumeCard 渲染门 | peek 活跃 session ∧ steps>0 ∧ done<total | 待造态 | SQL 造半完成活跃 session(X1-1 用过同法:update steps 留半);真路径=开测中途退 |
| RecentCard 渲染门 | authed ∧ digest 成功 ∧ latestScored 在 ∧ anchorRecordId 在 | 待造态 | fullchain fixture 已计分 session 即渲染;X1-1 wtrich1 零报告时=缺席(诚实零已活证) |
| RecentCard·composite「—」+无 chip 支 | 薄 cohort 无 verdict | 待造态 | 造 thin-cohort 数据(cohort 兄弟不足);CHIP 与数字分家的设计 |
| RecentCard·strip 整条缺席支 | abilities 全无 percentile | 待造态 | 同上薄 cohort |
| StatsCard·nudge 行 | covered<6 | 已活证 | X1-1 covered=0 带 note;**11-06 富态:covered=1/6+note「认知与潜力报告需测满 6 项才完整」+已完成场次 4(完成数口径)**;covered=6 支待造(需全六项) |
| StatsCard·reaction「—」支 | reactionBest undefined | 已活证 | X1-1 未测态(— 非 0) |
| BellEntry·红点门 | unread>0→dot+aria 带数;guest 跳读 | 待造态 | SQL 造未读消息行/真发站内信;已读回零走 messages 族 |
| CtaCard·desc 两支 | hasGames 单独定(desc 不看登录态,原型法) | 部分活证 | X1-1 见过 hasGames 态;need_games 支待造(易) |
| CtaCard·钮三态+hint 两支 | login/pick/start;hint 仅 !(authed∧hasGames) 渲染 | 部分活证 | start 态 X1-1;login/pick 态待造(易:登出/无游戏号) |
| 报告中心入口行 | 常驻 | 已活证 | X1-1 四面聚合已核 |

族小计:13 门=路由门 1/状态门 12;已活证 6、部分活证 2、待造态 5(易造 1)、零观测受限。零 toast、零交互门——本族全是数据门+导航。

### 2.4 match 族(routes 5 + match 模块 6 tsx;提取毕 2026-11-06)

> 族文件:routes/match{,.index,.pick,.switch,.guide}.tsx + features/modules/match/(page=guide 主面、genre-pick、switch-confirm、test-guide、ui/flow-header)+ session.ts(缓存键三形状法)+vocabulary.ts(开放词客户端词表)。守卫=/match 布局 profileFlowGuard+noindex 全子树承保。死码:零(flow-header 4 消费方;vocabulary 被 home/report 跨族引用)。
> 已在册法条:N1(switch rule2 文案**故意**偏离原型——服务器法=关旧场+强制新种子,同测试可再抽;注释明令禁回改)。

| 门实例 | 触发条件 | 可达判定 | 活证/处置 |
|---|---|---|---|
| /match 布局守卫(profileFlowGuard) | 匿名/未完成直访全子树 | 待造态(易) | 匿名→login?next;未完成→向导(§2.0 补遗) |
| undecided 平票重定向 | getMatchSession reason=GENRE_UNDECIDED→effect 弹 pick?why=tie | **已活证**(11-06 开发会话,G-408 修后) | 平票号(13977660011,CS2+王者荣耀 1:1)向导毕落 /match→自动收敛 /match/pick?why=tie;服务端互证 400 reason=ESPORTS_MATCH_GENRE_UNDECIDED;`.reason` wire 读法首次活跑成功 |
| pick·choose 种子创建 | 点 genre 卡→/match?genre=xx | 待造态(易) | resume 忽略 genre 参数(只种子创建),URL 参数=显式选择标记(switched 标) |
| test-guide·ready 双落点 | 有 undone 引擎步→test/$code;无→回 /match | 待造态 | 无步兜底=防呆 |
| guide·loading 面 | pending ∨ (error∧undecided 竞态窗) | 观测受限 | 瞬态;throttle 可视 |
| guide·error 面(题/述/重试) | sessionQuery 错误且非 undecided/emptyPool | 待造态(支已复活) | G-408 修毕(11-06):throwOnError:false 落,分支可达;GENRE_UNKNOWN/UNSPECIFIED 现导流 empty 面(G-406 接线毕)——此面仅余真 infra 类错误;专态造证(停 esports)归走查 |
| guide·genre null 支 | 服务端未知 genre 词→genre 面整块不渲染 | 数据门 | 防未来词;正常不可达 |
| GenreHeaderCard 2×2 | fps/moba × default/switched 标 | 部分活证 | X1-1 见过 fresh default 态;switched 标待造(URL 带 genre 即显) |
| StepRow 四态 | done(绿边✓+best 回执行)/ready undone left>0(绿字)/left=0(灰字)/!ready(dim+engine_soon) | 部分活证 | **done 支已活证**(11-06:MOBA 场 4/4 后行点击=review 重执行;fresh undone X1-1 在案);**left=0 已活证**(11-06 晚:MOBA 全耗尽场「已用完」灰字×4);!ready 待造;**per-test retake 额度差实捕**:FPS target_recognition=剩 1 次 vs MOBA 场 3 次(match_config 面;11-06 晚二抽 FPS 全 4 项皆剩 1=genre 级配置差再证) |
| 主钮双态+disabled | 有 next→go;全 done→go_result(→报告);next 空∧anchor 空→disabled | 部分活证 | go(第 1 项)已活证;**go_result 钮面已活证**(11-06 晚:MOBA 4/4 全耗尽场主钮「查看结果」——点击 1.1s 未导航,**单点存疑**:anchor 缺失下疑似 visible-but-dead 或慢导航,该场已关闭不可复核);**retake 耗尽面已活证**(11-06:重测 0/1+死灰钮「重新测试(机会已用完)」+「本款机会已用完,成绩已取最高——返回阵容继续下一款,或切换类型开新的一场」指引文=target_recognition 耗尽现场) |
| go_blocked 注释行 | next 空∧anchor 空 | **已活证** | 11-06 晚:MOBA 全耗尽场「可玩项目的重测次数已用完（或引擎即将上线）—— 切换类型可开始新的一场」两因+出路口径全字面 |
| **pick·empty 空池面** | 服务端池空两 reason 对值导流(生产者已接) | **已活证**(11-06 开发会话,G-406 修后) | 接线毕:session.ts 两 reason 常量+page error 分支对值 navigate why=empty;活证=直入 ?why=empty 面渲染毕(「没有可编排的测试…」+双逃生钮);机制腿由 G-408 平票号同分支族活证;服务器侧池空触发(删 config/抽干题)归走查真遇 |
| pick·tie 黄条/normal 灰条 | ?why=tie/无参 | **tie 双路活证**(11-06:手输 URL+**自动导流**——G-408 修后平票号 /match→pick?why=tie 收敛,黄条逐字+双卡全渲染);normal 灰条待造(直访即见) | 导流死循环已修(G-408 绿) |
| pick·推荐角标 | defaultGenre 偏好多数票;平票/无偏好→无角标 | 待造态 | 角标=原型外增强(注释在案);多数票号/平票号两态 |
| pick·两 genre 卡常驻 | GENRE_ORDER 冻结副本预览 | 常驻非门 | presentation-only,hydrate 无他处(注释在案) |
| switch·进度警告黄条 | doneCount>0 | 待造态 | SQL 造 done 数组 |
| switch·cur→tgt 卡 | session 读;null 兜底 fps | 数据门 | tgt=cur 反转 |
| switch·pending+成功链 | mutate 进行中;成功=invalidate 三键+peek+toast+回 /match | **已活证**(11-06 MOBA→FPS 整链:switch-confirm 页三规则+成绩保留说明+双钮→确认→回 /match FPS fresh 已测 0/4) | invalidate 键族法=三 session key+peek 全灭(§session.ts);X 线 X1 续费/切换例可并刀 |
| guide 页·first/second 双支 | firstSession 布尔;读未落兜底 first | 待造态 | 首场=新号即见;second 需一场完成史 |
| guide 页·count 兜底 | steps 0→显 4 | 数据门 | 防呆兜底 |
| toast·pick choose(success) | 选 genre | **已活证**(11-06「已选择 MOBA 测评」文案实捕) | |
| toast·switch 成功(success) | 切换落库 | **已活证**(11-06「已切换为 FPS(本次会话生效)」文案实捕) | |

族小计:22 门=路由门 4/状态门 16/交互门 2;**11-06 晚三连收**:go_result 钮面+go_blocked 注释行+StepRow left=0 灰字(MOBA 全耗尽场);**开发会话片A(11-06)再收两门**:undecided 导流已活证+pick·empty 接线毕已活证——已活证 8、部分活证 4、待造态 10(易造 3)、观测受限 1、数据门 3、死码嫌疑 0。toast 位 2/31 已全活(累计 15/31)。**genre 级 retake 预算差再证**:二抽 FPS 全 4 项剩 1 vs MOBA 3。

### 2.5 test 族(routes 1 + test 模块 14 tsx;提取毕 2026-11-06)

> 族文件:routes/test.$code.tsx(守卫=**自身** beforeLoad profileFlowGuard,顶层路由非父层承保)+ page(宿主)/bootstrap(绑定)/run(runId 重挂壳)/attempt(状态机)+ ui/×10(end-face 517 行、intro-face、pause-overlay、test-hud、prog-strip、retake-chip、submit-error、invalid-notice、valid-score、tier-chip)。**11 游戏引擎全 .ts(canvas 系零 JSX)不入分支账**,registry 消费已记 match 族 !ready 支。死码:零(ui 全消费;tier-chip 跨族 home 用;ROW_LABELS 41 键=results 派生键全覆盖,单测在案)。
> 族法条:①practice 降解——session 缺/undecided/读败→练习模式(unbound submit),**绝不阻塞游戏**;②one-submit-per-attempt(submittedRef 抗 StrictMode;非 completed 不提交不耗 retake);③重测预算耗尽防浪费(intro spent 死灰,免跑 30s 才撞 RETEST_EXHAUSTED);④review=存储判决即真相(无本地事实不造假,record null→fact 行缺席)。

| 门实例 | 触发条件 | 可达判定 | 活证/处置 |
|---|---|---|---|
| /test/$code 守卫(自身挂 profileFlowGuard) | 匿名/未完成直访 | 待造态(易) | 同 §2.0 补遗 |
| ?review=true 读回相 | done 步行入口(match 族 done 行 tap) | 待造态 | SQL 造 done 数组后从 guide 点 done 行 |
| page·loading/error/unknown 三面 | listTests pending/错/code 不在列 | 待造态 | error 面**无重试钮**(bootstrap 面有)——不对称,观察注记待 U2 对原型;unknown=手输错码即见(易) |
| bootstrap·loading/error(双 refetch 重试)/no-engine 三面 | 三 query 态/引擎缺席 | 待造态(易) | no-engine=code 无注册引擎;error=停服 |
| bootstrap·practice 降解 | session 缺/undecided/读败 | 待造态(路径已复活) | 法条①;G-408 修(11-06):bootstrap sessionQuery 豁免落,读败不再边界 throw,降解可达;活证=无场号直访引擎码 |
| bootstrap·duration 兜底链 | config 0→test→30s | 数据门 | proto3 零值注释在案(implicit presence) |
| attempt·phase 三相 | intro→running→ended(review 直落 ended) | **已活证** | 11-06 晚:target_recognition+flash_aiming 两场 live 全相走(intro 面/running HUD/ended end 面);review 直落 ended 早证 |
| RunTimeBar(running 专属,6px 最小宽) | running | **已活证** | 11-06 晚:flash_aiming 暂停帧截图=顶栏蓝色细进度条(≈90% 宽@26.92/30)视觉证;class 选择器两 hunt 皆空=**样式入 style 不入 class 面**,视觉截图为准 |
| PauseOverlay(running∧paused) | hud pause 钮 | **已活证** | 11-06 晚:「游戏暂停/继续游戏/退出测试」文本+截图双证;时钟冻结法条再证(26.92 三帧不动);**坑:暂停钮=aria 图标钮,文本正则永失配**(已回灌 conventions §1.2) |
| ExitConfirmDialog+exitDesc 双支 | pause→exit 钮开 | **已活证**(session 版) | 11-06 晚:**session 版真开全脸**「退出测试？/已完成 1 / 4 项。进度会保留，随时可续测；本次未完成的这一款不计成绩」+双钮(继续测试/退出并保留进度);**交互细节:取消退出=关对话框回暂停层,不自动续跑**(恢复须再点继续游戏);练习版 desc 待 practice 支 |
| verdict 链六面 | idle/submitting 脉冲/done✓ValidScore/done✗+reason→InvalidNotice/done✗无 reason→review 无分词/error→SubmitError | **五面已活证**(11-06):idle(intro);submitting=**「成绩判定中…」过渡文案实捕**(target_recognition 交卷瞬间,终态落 0 分·原始分);done✓(0 分 ValidScore×3 游戏+**tier 双态对照**:character_counter/skill_trajectory 未评级灰 vs red_dot 0 分 C·待提升);review 无分=**?review=true 面分数行「— 分」诚实隐去**(done 步行点击即入,phase 直进 ended);**InvalidNotice=BACKGROUND_OVER_BUDGET 词面实捕「本次成绩无效 · 检测到长时间切后台」**(无分数行/无 tier;patch 法=started 后 5s 派发+16s 持有>10s 阈值;pause 冻结实证:patch 前 hud 28.03→unpatch 后 27.23);待造=ABSURD_×2+DETAIL_MISMATCH 三词+SubmitError | **MSG-1 V10 伤痕同层高危区**;造态法沉淀:切后台法=mcp 工具层死路(tab 切换不产生 visibilitychange,visLog 空证);服务端真值=test_config.background_invalid_seconds 全表 10s(任一 span>10s 即无效);ABSURD_REACTION_TIME 阈=scoring.go reactionTimeMin 80ms/Max 60s(秒点法可行);**连坐发现**:review 场重测钮=visible-but-dead(end-face.tsx:481);retake 额 per-session binding=切换重新抽签后 counter 从 3 变 1 |
| one-submit 守卫+非 completed 不提交 | endRecord.exit≠completed | 法条在案 | 中止/失败场零服务器痕迹 |
| retry 重挂(runId key) | EndFace 重测钮 | 待造态(易) | 全新 virgin attempt(host/clock/overlays 全重置) |
| intro·HowToCard 空文→null | introduction="" | 数据门 | 诚实空面(无人撰写的行不加空框) |
| intro·prevBest 收据 | 该码有计分线;historyBest null→「—」 | 待造态 | 二场同码才见 history 支 |
| intro·双 mini 收据(进度/retake 余) | step≠null | **已活证** | 11-06 晚:「本场进度 0/4」+「剩余重测 1 次」(target_recognition)与「本场进度 1/4」(flash_aiming)两场;practice 无收据支待 practice |
| intro·spent 死灰+blocked 注 | step.retestLeft≤0 | 待造态 | 法条③;SQL 造 retest_left=0 |
| intro·countdown 覆盖层 | go 钮→3-2-1 | **已活证** | 11-06 晚:两场各捕中倒帧(「2 准备开始/取消开始」「1 准备开始」);cancel 退出链未点(保留) |
| intro·practice 支 | step null:无条/无收据/无编号,testName 当题 | 待造态 | 与 practice 降解同刀 |
| end·reviewVerdict 三支 | pending 行(raw 槽面)/无行(诚实无分)/有行(best+grade) | 待造态 | 曲线未校准游戏=新游戏;无行=未玩/作废两态同面(注释在案:仅此二态) |
| end·ValidScore uncalibrated 支 | authoritativeScore undefined→raw 占大槽+「原始分」标 | **已活证** | 11-06 晚:两场皆「0 分 · 原始分/未评级/该款常模积累中，暂无权威档位」(D-AO 不冒充校准分法条 live) |
| end·TierChip 五态 | S/A/B/C/灰 off(曲线 pending→未评级) | 部分活证 | **end 面灰 off 支已活证**(11-06 晚:「未评级」×两场);S/A/B/C 支待校准曲线数据;X1-1 报告面 off 态在案 |
| end·PrevTag 三支 | 首测词/delta≥0 绿/delta<0 红 | 部分活证 | **首测词「首局成绩」已活证**(11-06 晚×两场);delta 双支待二场同码。**数据观察:晨间 FPS 场(切换关闭,2/4)的 target_recognition 计分未成为新场 prevBest 基线**(切换关闭的不完场成绩不入 prevBest 线,语义注记) |
| end·BestTag 无史→null | historyBestScore undefined | **已活证** | 11-06 晚:两场 end 面均无 BestTag 块(诚实缺席) |
| end·AttemptsNote | 空线→null;pending 走 raw 行;best 高亮 taken 徽 | 部分活证 | **单 attempt 形已活证**(「本次共测 1 次：0本次取值——多次尝试取最高分，被选中的那次已标注…」);多 attempt taken 徽待 budget>0 场 |
| end·RetakeChip 灰→amber | left≤0 | **已活证** | 11-06 晚:「重测机会 0 / 1」灰态 |
| end·重测钮三态 | 活/死灰+blocked 行(spent)/死灰(review) | 部分活证 | **spent 死灰+blocked 行已活证**(11-06 晚×两场「重新测试（机会已用完）」);review 死灰 11-06 早证;活钮待 budget>0 场 |
| end·next/finish 双钮 | 有未完成引擎步→next(带能力名);无→finish(session→session-done 报告页;practice→/match) | **已活证** | **next 带能力名×2**(「继续 · 瞄准精准度」「继续 · 追踪稳定性」)+**finish 链**(11-06 晚:第 4 项 end 面「完成测评」→/report/session/$id 落地;注意 next 钮词=能力名非游戏名,「技能轨迹」正则失配教训);practice 支待 practice |
| end·fact 行 | record≠null(本地跑过);review 缺席 | **已活证** | 11-06 晚:**明细双 schema 实捕**(target_recognition 五行:击中/命中率/错失/误击友军/误击中立+无效点击;flash_aiming:命中数/失误数/中心命中/命中率/平均瞄准——per-game 明细面=法条④ record 在场) |
| hud·low 倒计时 danger 脉冲 | remainingMs<10s | 部分活证 | 11-06 晚:**<10s 文本态活**(「9.42 秒」采样命中窗口);danger 脉冲视觉未截(timer 元素在假设容器外,§3.4 坑二犯);legacy law 注释在案 |
| hud·step 行/practice 无 | step≠null | **已活证** | 11-06 晚:「反应速度 · 第 1 题 / 共 4 题」「瞄准精准度 · 第 2 题 / 共 4 题」两形;practice 无行支待 practice |
| SubmitError 重试链 | verdict error→resubmit | 待造态 | 停服在 submit 时窗可造 |

族小计:32 门=路由门 3/状态门 26/交互门 3;**11-06 双批活证后**:已活证 11(attempt 三相/RunTimeBar 视觉/PauseOverlay/ExitConfirm session 版/intro 双收据/countdown/ValidScore uncalibrated/BestTag 缺席/RetakeChip 灰/fact 明细**四 schema**/hud step 行)+**next/finish 双钮**;部分活证 6(verdict 五面/PrevTag 首测词/AttemptsNote 单次/重测钮 spent/TierChip 灰/低危文本态);待造态 ~9(retry 重挂/practice 族三支/prevBest history 支/多 attempt 徽/delta 支/unknown 面/SubmitError);数据门 3、观测受限 2。**零 toast 维持**;MSG-1 V10 同层高危区基本收割完毕。**交互新知:取消退出=回暂停层不自动续跑;HUD 控件=aria 图标钮**。

### 2.6 report 族(routes 6 + report 模块 20 tsx;提取毕 2026-11-06)

> 族文件:routes/report.{center,cognitive,career,trend,$recordId,session.$sessionId}(**六路由全挂自身 profileFlowGuard**)+ 模块 6 页面件(page=单报详情/center/cognitive/trend/career/session-done)+ ui/×14 + 5 ts 纯推导层(career/cognitive/digest/session/trend/vocabulary)。死码:零(ui 全消费;filing-footer 双份之一,文件级已核)。X1-1 红利:center 锁态四卡+covered 0/6 已活证。
> 族法条:①**free seat**(E5-0 裁)——最优已测主能力三件套(档/析/谏)全员免费开,卫星不竞争座,无 cohort 座虚位;②价格永不硬编码(全从 SKU 读面);③诚实零贯穿(composite 无→「—」、雷达缺轴=缺非零、career 空板不发明赢家、trend 混史黄警=决策②)。
> 观察项已结:G-001 早已裁定**仓法胜出 emoji 不还原**——center 四卡 📊🧠🎯📈+「›」字形+career 关闭面 🎯=违反已裁法的漏网,**已登记 G-405 红条**(换 lucide,随 UI 校准役或开发会话修)。

| 门实例 | 触发条件 | 可达判定 | 活证/处置 |
|---|---|---|---|
| 六路由守卫(各自挂 profileFlowGuard) | 匿名/未完成直访 | 待造态(易) | 同 §2.0 补遗 |
| 单报·onBack 双支 | 有史→history.back;直入→/match | 待造态(易) | 两刀:链入访问/直输 URL |
| 单报·pay 跳转两链 | MemberCard→/pay?from;UnlockCard→/pay?ref=recordId&from(trial 范围目标) | 待造态 | X1-2 mock 购会员链复用 |
| 六页面·loading 面(各页 query pending) | query pending | 观测受限 | 瞬态;throttle |
| 六页面·error 面重试钮(各页) | query 错误 | 待造态 | 停 esports 统一活证(六页同刀) |
| center·四卡状态徽 | digest.reportAccess/career.access/sessionsTotal 三键组合 | 部分活证 | X1-1 已见全锁态(unlocked 假/0 场 untested);解锁态待造 R-rich |
| center·covered 双文案+进度条 | covered≥6 切 full 句 | 已活证 | X1-1 见 0/6;full 支待造(全六项) |
| center·goDigest 空兜底 | latestAnchor 无→toast | 待造态(易) | 0 报告号点首卡 |
| session-done·金卡+网格 vs 诚实空板 | 有计分能力/全无 | 部分活证 | 诚实空板已活证两形(11-06:挂机场「成绩未定」;**晚:四场全跑全未评级场「0 / 4 项能力已测·本场成绩未定」**);金卡(有计分)待真分场。**11-06 用户视觉裁定:本页低保真→立 G-410**(已测收据语义反转/CTA 族反转/blob 氛围缺失;hero+勾选空板态 DOM 实证在位) |
| session-done·beat 徽/报告入口钮 | percentile 有/anchor 有 | 待造态 | 随完成场 |
| 单报·CompositeCard 双态 | composite 有(top X% 反转+beat 面板)/无(「—」) | 待造态 | 薄 cohort=无支;有 cohort=有支 |
| 单报·AbilityRadar shape null 支 | 全 score-only 不成形+satellite 注两支 | 数据门 | 卫星=子能力场次 |
| 单报·AbilityRow 四座 | free&&!unlocked 黄顶标/isSub 灰标/unlocked TierChip/locked LockLine 模糊条 | 部分活证 | X1-1 未见详情页(0 报告);free seat 法条活证待 unlocked 数据 |
| 单报·comparison/improvements | unlocked-only(null 整卡不渲染) | 待造态 | R-rich 解锁号 |
| 单报·CmpBar 三支 | best 无→null;refreshed(≥best 绿✓)/gap | 待造态 | 二场同能力 |
| 单报·MemberCard 四支 | unlocked×memberActive 窗口/单购;locked pitch+CTA | 待造态 | 四态号族(R-rich 会员/单购号/锁号) |
| 单报·UnlockCard | 仅 !unlocked 渲染 | 待造态 | |
| cognitive·locked 面 vs 开放面 | access.allowed | 待造态 | 锁=未满覆盖或非会员(裁决键待核 demand-ledger) |
| cognitive·radarReady 双支 | ≥3 轴雷达/少轴解释卡 | 待造态 | 少轴=浅覆盖号 |
| cognitive·missing 行+go_test 钮 | 未覆盖能力名单非空 | 待造态(易) | 部分覆盖号 |
| cognitive·CognitiveHead ring 双态+chip 仅 has+full_note | composite 有/无;tested<total | 待造态 | |
| cognitive·trend upsell 橙 CTA | !trendAdvancedAccess 才渲染(会员不见自我推销) | 待造态 | R-rich 非高级会员态 |
| trend·空态 CTA | sessionsTotal 0 或无系列 | 待造态(易) | 新号直访 |
| trend·advanced 徽双态+mixedGenres 黄警 | allowed;fps+moba 混史 | 待造态 | 混史=双 genre 场次号;决策②措辞 |
| trend·TrendChart 三色 chip/单点无线/点形三态 | delta±0;1 点;fps 填充/moba 空心/未知灰 | 待造态 | 单点=一场;点形=genre 词表 |
| career·closed 面(原型 verbatim)vs open 面 | access.allowed | 待造态(易) | 普通号=closed(已闭态);open 需 seeded library 数据 |
| career·open 空态 | 推荐+改进全空 | 数据门 | seeded 行缺席 |
| career·改进板 | improvements>0 | 待造态 | |
| FilingFooter/minor note 常驻 | 常驻 | 已活证 | X1-1 未见详情…常驻件随页活;归待造(易) |

族小计:29 门=路由门 3/状态门 25/交互门 1;已活证 1、部分活证 3、待造态 21(易造 6)、观测受限 1、数据门 3。toast 位 1/31(累计 16/31)。emoji 图标冲突观察项 1(已结转 G-405 红条)。

### 2.7 member/pay 族(routes 4 + member 1 tsx + pay 9 tsx;提取毕 2026-11-06)

> 族文件:routes/{member,pay,pay.index,pay.result}.tsx(守卫:member 自身挂 sessionGuard;/pay 布局父层挂全子树)+ features/modules/member/page.tsx + pay/(page 691 行=墙状态机、result-page 377 行=五面判决、ui/×7、order.ts/search.ts 纯推导)。age-required 路由壳已读(自身 sessionGuard),页面件归 agreement/age 批。死码:零(pay/ui 全消费;member 引 pay/order.ts 跨模块共享件)。
> 族法条:①**墙四态早退状态机**(birthMissing→回填卡/minor→拒卡双出口/**两态均零支付钮**/memberActive→续费面/其余→双选面);②face 点击时定格(open/renew),poll 中 session refetch 不改收据判词;③result=订单行自身真相五面,错读/无单号=诚实 error 面不永转;④mock 收银台=dev-only(CASHIER_DISABLED 拒非 mock provider)。
> 在册 R 条执法现场:R-03(cancel 文案=未产生扣款)/R-04(terminated 折 never 形,member 页窗口三态)/R-05(FAQ 硬价);G-403(QR UI 半绿,mock 恒走收银台)/G-404(H5 回程真通道刀)挂观测受限。

| 门实例 | 触发条件 | 可达判定 | 活证/处置 |
|---|---|---|---|
| /member+/pay 树守卫(member 自身挂/pay 父层挂) | 匿名直访 | 待造态(易) | §2.0 补遗 |
| 墙·showTrial 分支 | ref 在场(报告语境)=双卡;无 ref=member-only(原型 profile 入口同形) | 待造态(易) | 两入口各一刀:报告解锁钮/profile 卡 |
| 墙·birthMissing 支 | 生日缺(pay_gate fold) | 待造态 | SQL 清生日;回填链→/age-required |
| 墙·minor 支 | 未达龄线(ageLine 默认 18) | 待造态 | X1-5 未成年拦截例共刀;红卡+双出口零支付钮 |
| 墙·memberActive 态② | 活窗口 | 待造态 | R-rich;绿续费钮+benefit 链+紫 banner 带止日 |
| 墙·DualFace 态①/③ | 非会员/过期会员 | 待造态(易) | 过期支=SQL 造过期窗口(expired banner 灰);无 ref member-only |
| StateBanner 双形 | expired 灰/active 紫带 tag | 待造态 | 随上两行 |
| mock 收银台 sheet(三钮) | provider=mock 的 create 应答 | 待造态(易) | dev 默认 mock;成功=markPaid 桥→poll;失败=直落 result(fail);取消=closeOrder→result(cancel);X1-2 mock 购链主刀 |
| QR sheet(visual 槽+双钮) | provider 真+credential=QR | 观测受限 | G-403 半绿:mock 恒走收银台不可达;UI 活证随真通道联调刀 |
| QrTile 双态 | PNG 渲染/编码拒→回退值文本 | 观测受限 | 随 QR 腿;白垫静区(暗卡会吞模块) |
| invoke 腿(JSAPI) | 微信内+credential=invoke | 观测受限 | 真微信;G-402 裁②后服务端富集 openid |
| jump 腿(H5 跳走) | 微信外+credential=jump_url | 观测受限 | G-404 真回程刀;returnUrl 带点击时 face |
| settle 轮询链 | 700ms poll;PAID→invalidate session→result;CLOSED→result | 待造态 | 随 mock 三钮;桥同步落定=首次 poll 即答 |
| pending 钮(disabled) | create 进行中 | 观测受限 | 瞬态 |
| result·loading/error 面 | 错读/无单号直入 | 待造态(易) | 直访 /pay/result 无参=error 面即刻(不永转) |
| result·trial-ok 面 | PAID∧scope=trial | 待造态 | X1-2;绿 orb+收据四行(单号/额/渠道/时刻) |
| result·member 面 open/renew 双文案 | PAID∧scope=member;face 参数分叉 | 待造态 | 紫 orb+Crown;D-AU 第三参数 |
| result·fail 面 | 订单未决(poll 中逃逸/超时) | 待造态 | 黄警卡+同单重试链(ref+from 保往返) |
| result·cancel 面 | CLOSED | 待造态 | R-03 文案;灰 orb+Ban |
| result·memberSub 三支 | 天数+日期/NaN 裸句/裸句 | 待造态 | session 未落定时裸句(永不显陈旧日期) |
| result·viewHref/retry 链 | from 赢;trial→record/member→center | 待造态 | 四钮矩阵(view/retry/home/member) |
| channelLabel 三词+「—」 | h5/jsapi/native/未知 | 数据门 | |
| member·loading/error 面 | digest pending/null | 待造态 | error=回家钮单出口 |
| member·窗口三态 | active 紫条/expired 灰条带止日/never 无条 | 待造态 | **R-04:terminated 折 never 形**;三态号=R-rich/过期号/新号 |
| member·BenefitCard 第四行 | periodDays 缺→诚实缺席 | 数据门 | 目录降级读 |
| member·CTA 双文案+disabled | active=renew/open;目录缺=disabled | 待造态 | |
| member·FAQ 五对常驻 | 常驻 | 待造态(易) | R-05 硬价在案 |
| toast·create 拒因 | mutation cache 承载 | 待造态 | 全局 mutation toast 面(非本页字面 toast 调用) |

族小计:28 门=路由门 2/状态门 23/交互门 3;待造态 19(易造 5)、观测受限 6、数据门 2。toast 位 1/31(累计 17/31,mutation-cache 承载形)。R-03/04/05+G-403/404 五条在册项挂本族面。

### 2.8 messages/feedback 族(routes 5 + messages 2 tsx + feedback 4 tsx;提取毕 2026-11-06)

> 族文件:routes/{messages,messages.index,messages.$messageId,feedback,feedback.index,feedback.$feedbackId}.tsx(守卫=两布局父层 sessionGuard)+ messages/(page=双 tab 中心、detail)+ feedback/(submit、detail、ui/status-badge)+ wire×2。死码:零(status-badge 双消费;wire 消费齐)。**MSG-1 V10 五伤原发族**——上轮走查翻车现场,活证轮重点。
> R 条执法:R-08(关单=灰徽+新文案;列表行以内容首行为题——FeedbackRow 首行=content 截断,代码证实)/R-01(两型两范围)/R-02(contact 输入 40 帽,契约 128,原型紧形注释在案)。

| 门实例 | 触发条件 | 可达判定 | 活证/处置 |
|---|---|---|---|
| /messages+/feedback 树守卫(父层挂) | 匿名直访 | 待造态(易) | §2.0 补遗 |
| ?t= 双 tab(URL search 承载) | msg/fb;切换 replace;双腿常活(tab 切换显暖数据) | 待造态(易) | 直访 /messages?t=fb |
| 双腿三态(pending/error/ok) | 各 query 独立态 | 待造态 | error 面**无重试钮**(与 test page 层同形)——观察注记 |
| 未读计数条 | unread>0 | 待造态 | SQL 造未读行;读回零走 detail 读即标记 |
| msg/fb 空态卡(fb 带新建钮) | 列表空 | 待造态(易) | 新号直访 |
| MessageRow·未读红点+icon 四形+兜底 | isRead=false;type 三专+**未知兜底=喇叭非错图标**;日期 NaN→缺席 | 待造态 | 四类型各造一封信(SQL/真发);未知 type=SQL 造脏行 |
| FeedbackRow·meta 三词+徽+日期拼支 | status 三态;meta 空则无分隔符 | 待造态 | 三态工单(admin 关单/回复/新) |
| StatusBadge 三态 | pending/replied/closed(R-08 灰徽) | 待造态 | 随上 |
| msg·detail loading/双题 error 面 | pending;成功行不在=not_found/失败=error 双题 | 待造态(易) | 手输不存在 id=not_found 面 |
| msg·detail 读即标记 | 开信 effect markRead(幂等抗 StrictMode);invalidate 全灭→badge+点回零 | 待造态 | 未读信开一次,回列表看点消 |
| msg·detail 深链双 CTA | type=report/fb-reply ∧ refId 在 | 待造态 | 真发信走链;公告信无 CTA |
| fb·submit 类别 chip 六选横滚 | rail 滚不折行 | 待造态(易) | |
| fb·submit err 两支 | 未选类/内容<10 字 | 待造态(易) | 原型极小=10 字可载复现 |
| fb·submit 计数行满红 | ≥500 | 待造态(易) | |
| fb·submit 成功链 | toast+跳队列(处理中徽=确认) | 待造态(易) | 真提一单 |
| fb·detail 线程双边 | admin 蓝边/user 无边 | 待造态 | admin 回复走 admin 面(与 A 线共刀) |
| fb·detail closed 支 vs 跟进条 | closed→关单注(R-08 文案),跟进条整块不渲染;open→输入+发送(disabled=空/pending) | 待造态 | 关单=admin 关;followup 翻回处理中 |
| fb·detail followup 成功链 | toast+badge 重绘 | 待造态 | 真追问一句 |
| toast·submit_ok/followup_ok(2 位) | 两 mutation 成功 | 待造态(易) | 随上两刀 |
| mutation 拒 toast(2 位,cache 承载) | 提交/追问被拒 | 待造态 | 停服窗 |

族小计:20 门=路由门 4/状态门 12/交互门 4;**11-06 活证走查八面收割**(18606513270 号):①messages 列表(双 tab+未读条「1 条未读」+公告 MessageRow);②消息详情(系统通知 heading+系统公告徽+日期+正文);③**读即标记闭环**(未读条+tabbar 徽标+BellEntry 三面同步清零——顺带 §2.3 BellEntry 红点门活证);④feedback 空 tab「还没有提交过反馈」+我要反馈 CTA;⑤提交页满面(六 chip/描述 0/500 计数/联系可选+站内信回复说明/生命周期预告文);⑥提交闭环(toast「已提交 · 状态:处理中」→回列表 pending 行=内容首行为题 R-08 形);⑦详情 pending(状态徽/类型·提交时间/用户线程首帖/发送钮+追问预告);⑧**追问 append**(toast「已追问 · 状态回到『处理中』」+第二帖入线程)。**X2-1 六腿前段活证**;admin 腿(队列/快捷回复/关单/幂等)待 admin 起栈。error 面无重试钮观察 1 仍在。

### 2.9 security 族(routes 13 + security 模块 13 tsx;提取毕 2026-11-06)

> 族文件:routes/profile.security{,.index,.password,.bind-phone,.bind-wechat(×4),.merge(×5)}.tsx + /login/forgot(公共重置腿,**无守卫=设计**)+ 模块 13 tsx(page 总览/password/forgot/bind-phone/bind-wechat/merge/merge-success+四 callback+pending-landing)+5 ts(fn/hooks/flow)。守卫=/profile 布局 sessionGuard 全树承保。死码:零(merge-failure/merge-flow/pending-landing/use-sms-code/use-bind-sms 全有消费方;mainlandMobile/inputClass 跨族复用再证)。
> 族法条:①**方向=数据非客选**(merge 的 poor→rich 由 offer 状态翻转定,客户端永不选边);②**preview 先于发码**(占用号→冲突卡不耗 SMS,原型「无需验证码」法);③贫方手机号 typed never carried(客户端只持掩码);④身份切换后全灭+refetch session(同 post-login 法);⑤不可逆合并=**双闸**(橙 CTA→BottomConfirmSheet 热钮最后确认);⑥掩码服务端造,前端永不拼。
> 注:本族 toast 调用点约 18(全 ctr 最密),与 §2 普查 31 处口径有差(普查按文件/去重口径)——活证轮以逐调用点为准复核。

| 门实例 | 触发条件 | 可达判定 | 活证/处置 |
|---|---|---|---|
| 守卫(父层全树)+forgot 无守卫 | 匿名直访 security 树;forgot 公开 | 待造态(易) | §2.0 补遗;forgot 公开性=设计非漏洞 |
| merge·offer 空手弹回 | readMergeFlow 空(一次性读,URL 不可再推导) | **已活证** | 11-06:SPA 导航入 /merge→effect 弹回 /profile/security(URL 往返证)。**机制勘误:handoff=模块级 in-memory 非 sessionStorage**(merge-flow.ts 注释明言 in-memory on purpose;硬刷新即重置=天然空手) |
| bind-phone·from 三入口 | me/login/guide→doneTo 双落点 | 待造态(易) | ?from=login 直访 |
| bind-wechat·failed 支 | callback 弹回带 failed=true | 待造态(易) | 直访 ?failed=true:红冲突卡+CTA 换词(去登录该微信=login 腿) |
| password·mode 双面 | set(SMS 证明)/change(旧密证明) | **已活证** | 11-06:change=「已设置过密码——需先验证旧密码」+忘记旧密出路链;set=「验证码将发送至已绑定手机 186\*\*\*\*3270」+手机号/验证码栏双面 |
| 总览·三凭据行×双态 | bound(绿 chip+服务端掩码)/unbound(绑定链) | 部分活证 | **bound 三面已活证**(11-06:手机 186\*\*\*\*3270+微信 o5C\*\*\*\*KQAQ 双掩码+绿 chip「已绑定」×2;阿包号三凭据齐全实测);unbound 态待新号 |
| 总览·密码行双态+信息四行 | hasPassword;生日三支(authored/band note/未填) | **已活证** | 11-06:bound 支「已设置——可用于账号密码登录·修改」+信息四行(昵称阿包/性别未填/出生年月 1990-08=**authored 支**/游戏偏好 cs2、dota2);band/未填支待 R-minor 类号 |
| password·校验梯五支 | 旧密空/手机非法/码非 6/新密<6/组成/不一致(原型序) | 部分活证 | 11-06:空旧密「请输入当前密码」+确认不一致「两次输入的密码不一致」两 rung 活;余 rung 桌面输入即得未逐支 |
| password·NoPhoneCard | set∧无手机∧ready→死路卡+去绑定链 | 待造态 | 微信生无手机号 |
| password·reason 词表四+成功双 toast | MISMATCH/POLICY/SMS_INVALID/SMS_RATE;set_done/change_done | 部分活证 | **MISMATCH 支活证**(11-06:内联词「旧密码不正确」×服务端互证=system 日志 reason=SYSTEM_PASSWORD_MISMATCH code=400);POLICY/SMS_*/成功双 toast 未造 |
| forgot·校验四支+reason 三词+成功链 | SMS_INVALID/USER_NOT_FOUND/RATE;done→回 /login | 部分活证 | 11-06:发码+「60s 后重发」倒计面+**USER_NOT_FOUND 词「账号不存在」**(failureCopy 分支;发码对未知号不拒=不泄漏注册态,观察注记);SMS_INVALID/RATE/成功链未造。**走查陷阱沉淀:useSmsCode state=挂载级内存态,send+submit 须同挂载**(跨导航重挂丢 state→zod throw→generic 兜底词假象,首次误读为无专属词) |
| bind-phone·preview 先行四果 | FREE→才发码;SELF→info toast;KEEP_*/BLOCKED→冲突卡 | 部分活证 | 11-06:**SELF**=toast「该手机号已绑定当前账号」;**FREE**=发码+倒计(13900000001);**BLOCKED**=占用四探全 BLOCKED 味(13800001234/18657170082/13900001111——fixture 号全有数据权重);**KEEP_\* 味待真空号**(fixture 无空号) |
| bind-phone·冲突卡双 CTA | BLOCKED=查看说明;可合并=去合并(setMergeFlow 带 offer+phone) | 部分活证 | **BLOCKED 味真链已活证**(11-06:「该手机号已绑定其他账号/本次绑定已拦截…」卡+「为什么不能自动合并？查看说明 ›」→点击走通真 producer 链 preview→setMergeFlow→merge BlockedPane,**与注入面逐字一致=注入保真度互证**);KEEP_\* 味「去合并」CTA 待空号 |
| bind-phone·三错误行+发码钮四态 | rate/codeErr/previewFailed;countdown/checking/sending/preview 锁 | 部分活证 | 11-06:发码钮 countdown 支「60s 后重发」活;三错误行未造(rate 键族=redis send-window/send-daily 已定位) |
| bind-phone·entry line+skip/done toast | from 词+method 双支+昵称;skipped/done | 部分活证 | 11-06:from=login 支「入口：登录后 · 当前账号：阿包（手机号登录）」(**from 词+method 支双活**)+skip toast「已跳过，可稍后在账号与安全再绑」+落回 /;done toast 未造(不执行真绑定保号态) |
| bind-wechat·env 半透明+env 拒 toast | UA 非 MicroMessenger | **已活证** | 11-06:桌面钮「授权并绑定」+点击 toast「请在微信中打开」(error toast 持续存在=跨 SPA 导航残影,观察注记) |
| bind-wechat·failed 红卡+note+CTA 换词 | failed=true 三件套 | **已活证** | 11-06:「该微信已绑定其他账号，本次绑定已拦截、未做任何改动。」+入口/当前账号行+**CTA 换词「用该微信登录 ›」**+note+skip 双文 |
| bind-wechat·begin 拒 toast×2+skip toast | beginBind/beginLogin 抛 | 观测受限 | 停服窗 |
| bind-wechat-cb·6005 竞态再 preview | IDENTITY_ALREADY_BOUND→**再开 fresh OAuth preview 腿** | 观测受限 | 真微信;注释在案(preview 不能复用已烧 code) |
| bind-wechat-cb·三果 | ok→toast+回;rejected→failed 弹回;codeless→failed | 待造态(易:codeless) | 真腿=观测受限 |
| preview-cb·四果 | rejected→failed 弹回;KEEP_*/BLOCKED→setMergeFlow→merge;SELF→toast | 观测受限 | 真微信 |
| merge-bind-cb·三态 | codeless→词表行;拒→词表;成功→success 页 | 观测受限 | 真微信 |
| merge-switch-cb·三相+身份切换 | switching→proving→failed;切换后全灭+refetch session | 观测受限 | 真微信 |
| merge-switch·发码/合并链 | sent toast;rate 词;merge 拒词表行 | 待造态 | proving 态须真切换后 |
| merge·四 scenario 面 | KEEP_CALLER(sms/wechat 双 proof)/KEEP_OTHER(切换 CTA)/BLOCKED(独立面) | **已活证** | 11-06:**直注缝修正=dev 动态 import 同实例**(merge-flow=模块级 in-memory,`import("/src/features/modules/security/merge-flow.ts")`→setMergeFlow→SPA 导航;非 sessionStorage);四脸全收:keepCallerSms(证明区+橙 CTA)/keepCallerWechat(「合并账号并绑定微信」)/keepOther(「切换到微信账号并合并」蓝 CTA)/BlockedPane(判定依据框);真 producer 链交叉验证见冲突卡行 |
| merge·poor→rich 双卡 | summaryLine 三支(空/会员止日/已解锁);label 掩码/微信/裸 | **已活证** | 11-06:**summaryLine 三支全收**(「无记录 · 无会员 · 无解锁」/「3 场记录 · 会员有效至 2026-12-31」/「5 场记录 · 报告已解锁」)+label 三形(掩码 186\*\*\*\*3270/裸「账号」/「微信授权」);poor 虚线框×rich 实线框 |
| merge·keepCallerSms 证明区+橙 CTA→**双闸 sheet** | smsReady=6 码+大陆号;sheet 热钮=最后确认 | **已活证** | 11-06:flow.phone 预填输入框实测+**CTA 状态机**(disabled→6 位码后 enabled)+双闸 sheet 全脸(热钮「确认合并」+ghost「取消」关闭);未点热钮(mergeSmsFn 真执行须 state,注入态无) |
| merge·三拒 toast 词表+later 钮 | MERGE_NOT_MERGEABLE/NOT_BOUND/SMS_INVALID;later 清 flow | 待造态(易) | |
| success 页(静态) | ✓ hero+两 check 行+回 profile | **已活证** | 11-06 直访:「合并成功/微信授权已绑定（昵称 / 头像已同步）/历史测评、报告与会员有效期均已保留」+返回链+尾部提示 |

族小计:29 门=路由门 5/状态门 18/交互门 6;**11-06 活证走查大批收割**(阿包号):已活证 9(空手弹回/password 双面/总览密码行+四行/bind-wechat env+failed 双面/merge 四 scenario+双卡+双闸 sheet/success 页)+部分活证 8(总览三凭据 bound 态/校验梯两 rung/MISMATCH 词/forgot 三面/preview 三果/冲突卡 BLOCKED 味/发码钮 countdown/entry+skip);待造态 5(守卫群/NoPhoneCard/三拒 toast/KEEP_\* 味空号/POLICY 支)、观测受限 7(真微信腿)。toast 位实收 6(SELF info/env 拒/跳过/验证码已发送×2/账号不存在为内联非 toast)。**活证缝修正:merge=dev 动态 import 同实例注入+SPA 导航(§6.2 conventions);同挂载连续律适用于一切 sms state 链**。

### 2.10 末批:profile 总览+agreement+age-required(routes 5 + 14 tsx;提取毕 2026-11-06)

> 族文件:routes/profile.index+agreement.{user,privacy}+age-required(壳已读)+ profile/(page 259、logout-sheet、ui/×7、overview/bands)+agreement/page+age-required/page。守卫:/profile 布局 sessionGuard;/age-required 自身挂;**协议双页公开=设计**(登录前链接需预读)。死码:零(logout.fn/overview/bands helpers 全消费)。
> **新发现:占位 toast 过时嫌疑(11-06 判决定案=接线,立 G-407)**——profile 面报告中心卡「查看全部/查看报告」+ICP park 在 `toast.info(common_soon)`,注释自证「center 页是 later face」;但 /report/center 与 /report/trend **均已建成**(§2.6),home 入口行也已接线(X1-1 活证)。**原型实锤(profile-overview.html:289-291):三钮全真跳转——toCenter→report-center.html、toTrend→report-trend.html**,「Phase-2 故意留」不成立(押后无任何登记;训练趋势是另一张卡,原型自带 Phase-2 虚线占位,与此卡无关)。=漏拆占位,原型固化律下对原型明确偏离。ICP soon 合法不动。

| 门实例 | 触发条件 | 可达判定 | 活证/处置 |
|---|---|---|---|
| IdentityCard·meta 行/pill×2/四 stat 条 | meta 空整行不渲染;pill=soon 占位;prog 恒「—」(无源面,诚实 dash 注释) | 部分活证 | **meta 行 live**(11-06:「30-40岁 · 2 款游戏偏好」=band 派生+偏好数);四 stat 条 live:总测试 4/综合分 —(诚实 dash)/**最高分 NaN 分=G-411 抓伤现场**/进步次数 —(恒 dash 法条 live);pill soon 待造(易) |
| DataCard 三行×未设词+年单支 | 未设/年无月 | 部分活证 | X1-1 |
| SecurityCard 三行 bound/unbound | manage/bind 双链+掩码+chip 三态 | 待造态(易) | R-rich/新号两态 |
| ProfileMemberCard 双态 | active(止日+unlock 行,digest 缺行不渲染)/pitch(价+月标) | **已活证** | X1-1 翻转前后两态 |
| AbilityOverviewCard 双态行 | tested(chip+best+track)/untested(dim+hint) | 部分活证 | 11-06:**untested 行×6 live**(dim+「未测」chip+hint 两支:单 genre「测 FPS/MOBA 测评可解锁」/双 genre「完成一场测评即可解锁」);tested 行待真分号**G-411 连带:played-0 分行落未测分支** |
| HistoryCard 三态行×2 genre | 未玩→去测评/未完→继续/完→读回(anchor 分叉 report/session-done) | 部分活证 | X1-1 未玩态;**完态 live**(11-06:FPS/MOBA 双行「2026-11-06 · 已完成」+查看报告 CTA);未完态待半场数据 |
| GenreStatsCard 共享轨道 | max 定标,零显零 | **已活证** | 11-06:FPS 2 次/MOBA 2 次双行满格(max=2 定标);零显零支待单 genre 数据 |
| 训练趋势占位卡(虚线+soon 标) | 常驻 | **已活证** | 11-06:「训练趋势 敬请期待 自由练习与专项训练模块…」Phase-2 placeholder(原型自带) |
| **报告中心行占位 toast×2** | 查看全部/查看报告 | **已判(11-06):接线,立 G-407** | 见批注(原型 289-291 实锤);执行归开发会话 |
| 语言行(LocaleSelect 现装,D-AT) | 常驻 | 待造态(易) | |
| about 行(双协议链+ICP soon) | 常驻 | 待造态(易) | |
| LogoutSheet+onLogout 链 | open 受控;取消 autoFocus 领衔;确认=logout→clear→clearWechatProfile→broadcastSessionKick(**无 expired 通知=刻意**)→/login | **已活证** | 11-06 晚:确认→本页落 /login;**多标签踢出实证**(第二标签页 home→被踢 `/login?next=%2F`,kick 带原 href 播种);无 expired toast 刻意形同证;sheet 取消/autoFocus 支待复走 |
| agreement·双页公开+版本行+demo note | meta 有才渲染;wireDate 双形状归一 | 待造态(易) | 直访;admin 发版后版本行变 |
| age-required·回填预填+校验两支+minor toast | 岁线=catalog payAgeLine(次月起算同服务端法);saved toast+回 / | 待造态(易) | 未成年选年月=minor 拒(X1-5 共刀) |
| toast·soon×5 位+minor+saved | 占位族 | 待造态(易) | |

族小计:15 门=路由门 1/状态门 11/交互门 3;已活证 1、部分活证 3、待造态 10(易造 9)、占位过时嫌疑 1。

## 3) 文件级 · admin esports(已迁独立档 `render-coverage-admin.md`;2026-11-06 admin 起栈日:U1-admin 首轮 25 面=22 live+1 观测受限(agent 依赖)+2 parked)

## 4) 计数收口与对账环

- **U1-ctr 分支级提取全量毕(2026-11-06)**:11 批=壳 15+守卫补遗 4+login 22+向导 25+home 13+match 22+test 32+report 29+member/pay 28+messages/feedback 20+security 29+末批 15=**254 门**(交互门 43/状态门 174/路由门 37)。
- 判定分布:已活证 13(X1-1/e2e 红利)、部分活证 11、e2e 在案 3、待造态 189(易造约 80)、观测受限 28(真微信腿+瞬态各半)、数据门 14、法条在案 4。
- **悬置三件已判(11-06,「先由你来判断」授权)**:①Skeleton=删(孤儿资产,Q1-③ 残骸无预留登记,现行 loading idiom=LoadingStrip 已竞争上岗);②pick·empty=接线(错接非预留,立 **G-406**);③profile 报告中心=接线(原型 289-291 三钮真跳转实锤,立 **G-407**)。①为代码删除、②③为红条工单,执行均归开发会话。红条 G-405 已立(emoji 字形)。
- 下一步:起栈 U1-ctr 活证走查(chrome mcp 人在环;优先序=test verdict 六面→messages/feedback 全族→security merge 离线缝→易造 80 门铺开),后 U1-admin。**进度(11-06 全天):前四项毕+易造大丰收**——verdict 5/6 面/messages 八面/security 大批(merge 全族/双闸/bind-phone 真链/password/forgot/bind-wechat)/test 单场×4 全脸/向导族直访群/登出态批(Logout 链+双守卫匿名支+login 族 toast 群)/profile 总览连walk;**新伤五枚(G-408..412)**。ctr 侧小残=FormErrorLine 三支/AgreeBox indeterminate/home guest 面/logout sheet 面文(皆需再登出循环,边际低)+InvalidNotice 余三词+SubmitError;**下一大块=admin 起栈(X2-1 admin 腿+U1-admin 文件级)**——**起栈日毕(11-06 晚)**:X2-1 六腿全闭环/X2-2 全毕/X2-3 半毕(G-413 立)+U1-admin 首轮 25 面(22 live+1 观测受限+2 parked)+U3 三裁定入 calibration-log(tab 筛选族→Filter/Segmented、布局留白、drawer 重型 table);下一=X1-2 mock 购会员→X3 铸码导出→A-套余面→U2 harness 重激活;伤单累计 G-405..413(2 绿 11 红)。**开发会话片A清偿(11-06 晚)**:G-408/406 翻绿(伤单 4 绿 9 红)——六文件 throwOnError 豁免族(含 8 页体检:4 真全修/4 假维持边界)+空池两 reason 对值接线;平票号(13977660011)导流收敛+服务端 400 互证+empty 面渲染活证;序=片B NaN 零值族(G-409/411)→片C 机械批(G-407/405/412)→片D G-413 admin 信箱 Tab。
- 死码嫌疑处置:修条件 bug 或删(逐条用户过目);预留门登记编号引 gap-ledger/pm-review。
