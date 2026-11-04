# Gap 台账

> 每条含:期望(原型字面量)/实际(重建渲染)/处置。处置:红=按原型严格还原的工单;待面检=需按态比对定性。存疑不阻塞(原型固化律,见 README)——一律转 pm-review-items.md 留档。
> 机检来源:e2e 比对 spec 的 diff 输出(app `e2e/prototype/*.fidelity.spec.ts`,钉 zh-CN);手工来源:走查记录。基线注明工作树状态。

## 首批 · V-1 login 样板(2026-10-28)

出处:`prototype/login.html`(锚=app/clients/ctr/e2e/prototype/fixtures/login.json);实际=工作树 `/login` @ zh-CN(E5-f3 未提交态,但 login 面不属该切片)。

| # | 面 | 类型 | 期望(原型) | 实际(重建) | 处置 |
|---|---|---|---|---|---|
| G-001 | 品牌 icon | 缺行 | ⚡ | (无) | ✅闭(11-01 PM-4 现状核验):lucide Zap 全站随所有面走查落地,仓法胜出,emoji 不还原 |
| G-002 | 默认登录 tab | 状态 | 默认=验证码登录:手机号/6 位验证码/发送验证码 | 默认=验证码登录(E6c-a 后) | ✅修复(10-28 E6c-a bf0a99a0,双通道 live 走查随片 proven-by;10-30 验证会话复跑=代码级:useState 初值 "code"+segment DOM 序 code 先,page.tsx:66/239);PM-3 随之收口 |
| G-003 | 账号字段语义 | 文案 | placeholder=手机号 | 请输入账号 | ✅修复(10-28 E10-a;原型两 tab 皆「手机号」已渲染态实证;手机号登录功能随 E6c) |
| G-004 | 表单字段标签 | 多行 | 无标签(仅 placeholder) | 多出「账号」「密码」两行标签 | ✅修复(10-28 E10-a;sr-only 文本泄漏 innerText 实证→改 aria-label) |
| G-005 | 非微信提示 | 状态逻辑 | 微信内提示区为空;非微信态显示该文案 | 浏览器常显(与非微信态一致) | ✅闭(11-01 PM-5+态比对 live 实证):UA 嗅探已落(page.tsx:77);微信 UA 隔离页实测=提示区空+微信钮全亮,非微信态常显注记(常态走查既证)——两态与原型全同 |
| G-006 | 底部链接行 | 标点 | 忘记密码？ · 用户协议 · 隐私政策 · 返回入口 | 忘记密码？·《用户协议》·《用户协议》·返回入口 | ✅修复(10-28 E10-a;裸名键+「 · 」字面空格——原型空格=源码空白折叠,margin 不入 innerText 实证) |
| G-007 | 合规主体 | 标点+缺行 | 浙ICP2026028293号;另有独立行 客服邮箱 xgwledu@126.com | 浙ICP备2026028293号;客服邮箱行缺 | ✅闭(11-01 PM-1 知情复裁,**维持 D-AO① 带「备」**,928077f3 回滚):当日去「备」翻案查明 D-AO① 在案后复裁回滚;现行=浙ICP备2026028293号(login 页脚 live 实读);邮箱行既证;上线复核=D-AO① 原案 beian.miit.gov.cn 一次 |
| G-008 | 杂行 | 多行 | — | 「xl」 | dev-only 实证:TailwindIndicator 有 DEV 守卫,prod 恒 null;已标 data-dev-chrome,harness 剥离(devtools 的 Filter placeholder 同族=vite 插件 prod 剥离实证) |
| G-009 | 页面标题 | 文案 | ctr · 登录 | 登录 · CYBER ESPORTS | ✅闭(11-01 PM-2 用户裁,ab7ec52f):品牌词回原型「ctr」,全站「ctr · 页名」与原型字面量全同(标题+404/错误面品牌词同步,live 实读) |

计数:9 条 = 修复 4 + ✅闭 4(G-001/G-005/G-007/G-009,11-01 裁决+现状核验随行;G-007 终态=维持 D-AO① 带「备」,当日去备翻案经知情复裁回滚 928077f3) + dev-only 1;首批物料 PM-1..PM-5 **全收口**。
E10-a 复对账(10-28,live dev 栈,同款 innerText oracle):14 行中 12 行逐字全同,余 2 行=已裁(ICP 备)+态条件(非微信注记);placeholder 对(手机号/密码（≥6 位）)逐字同。

## 二批 · E5b-d 报告四页(2026-10-28)

出处:`prototype/report-{center,cognitive,trend,career}.html`(渲染态读:chrome file:// 活原型+逻辑校验面板,登录门经模拟验证码过);实际=工作树 `/report/{center,cognitive,trend,career}` @ zh-CN(E5b-d 未提交态)。

| # | 面 | 类型 | 期望(原型) | 实际(重建) | 处置 |
|---|---|---|---|---|---|
| G-101 | 职业卡/关闭态文案 | 文案 | 「职业推荐 + 能力差距 + 4 周处方」 | 「…+ 训练处方」 | 白名单形②:4 周制随 E5b-0 裁定②作废(新契约处方=维度差距行,无周结构);E5b-a 契约注释自证 |
| G-102 | 职业页「发展路径」 | 缺面 | 关闭态预告含「发展路径」 | 开放态无路径段 | 白名单形①:契约没有的面不造(CareerPathStep 无 E5b 载体,语料裁出);预告文案照录 |
| G-103 | 认知页「学科优势匹配」段 | 缺面 | 学科优势/职业潜能/提升计划三块嵌入认知页 | 学科优势不建;职业/处方在独立职业页 | 白名单形②+①:D-AP 四卡四页裁定(职业独立页);学科优势无契约不造 |
| G-104 | 「后台可配置」徽记 | 状态 | 三块带徽记 | 无标记 | 白名单形②:库文案服务端种子随行=配置下发的新栈实现形(careerAreas/careerPrescriptions),徽记语义由服务端承载 |
| G-105 | 免责声明配置化 | 文案 | 「后台可配置」徽记+固定文案 | 前端常量键 | 白名单形②:admin report-config(A3)未实现=原型同为前端硬编码(其逻辑校验面板自证);配置面若立再迁 |
| G-106 | 报告中心卡 1 维数词 | 文案 | 「最近一次测评的 5 维能力图」 | 照录 | 零偏离:照录(D-O2 文案一样);实际维数随 genre 4-5 维由单次报告面既有承载 |

计数:6 条 = 白名单形② 4 + 形① 2 + 照录 1(重叠计:G-106 零偏离不计处置)。
对齐判定:四页布局/文案/状态徽章/空态句=原型渲染态逐字誊录(覆盖条 N/6 动态、<3 维不画雷达句、缺维不画轴句、趋势 2 次门槛句、职业关闭态全版);live 验证=四路由注册+profileFlowGuard 链(curl 307→login?next 各自路径);登录态渲染递延片 e(策略翻转 e2e 带会话全链)。

## 三批 · E6c-c 账号安全面(2026-10-30)

出处:`prototype/{bind-phone,bind-wechat,account-merge-confirm}.html`+security/密码族 block3/4 文案块(渲染态读);实际=工作树 `/profile/security/*`+`/login/forgot` @ zh-CN(E6c-c 未提交态)。

| # | 面 | 类型 | 期望(原型) | 实际(重建) | 处置 |
|---|---|---|---|---|---|
| G-201 | onboarding 昵称步「获取微信昵称」pill | 缺面 | 昵称输入侧「获取微信昵称」一键带入 | 无 | ✅修复(10-30 微信回填片):snsapi_userinfo 全链——登录/绑定响应带 WeChatProfile,昵称步预填(截断/剔除法)+pill(微信环境+未带出时,绑定流复用,占用 openid 诚实拒绝)+头像同笔落 profile.avatar_url;emoji/✓/⚠ 字形按图标律换 lucide |
| G-206 | 年龄步性别区预告文案 | 文案 vs 缺面 | 「必选；微信授权登录时会自动填入」 | 文案已改「必选」,微信 sex 不回填 | ✅修复(10-30 真机回归用户裁定:改文案去承诺——原型文案对用户是不兑现的承诺,sex 回填被裁不做;性别三态 undetermined=诚实形状是 E6b 裁定) |
| G-207 | 微信昵称清洗规则基数 | 策略 | 加权计(CJK 2/上限 24=12 中文字或 24 字母),规则卡含 pill 行 | 码点计(上限 12,中英同权——E6b 冻结契约) | 白名单②:现仓契约(UpdateProfileRequest.nickname 2..12 码点)为正身,清洗按契约截断;原型加权制与规则卡 pill 行=文案差,随文案定稿批次(PM);另:头像显示面递延(现仓 profile 页无头像位,原型 profile-overview 有——存储+供给已闭,显示随 profile 页迭代;10-30 真机回归用户佐证「找不到填头像的地方」) |
| G-208 | 登出确认 sheet 形状 | 浮层 | 竖排按钮(取消上/退出下,btn-pri 蓝青渐变)+上圆角 18px+15px 标题+全角问号 | 横排按钮+退出=红 danger+四角圆角+14px+半角问号 | **首刀已收(10-30)**:BottomConfirmSheet 单源组件(CTS.sheet 正身:390px/上圆角 18px/竖排全宽 gap 9px/三色族 primary·hot·ghost——红只留真危险),exit-confirm 重写+logout-sheet 薄壳+merge sheet 三迁,原型逐字文案+chrome 双侧 7 项几何实测全同;**余族审计毕(11-01)**:match-switch-confirm/resume-remind 双腿到期核验=**原型 oracle 均为整页非浮层**(余族按浮层记系误挂)——前者重建逐元素逐字全同(唯 rule2=已注册走查 N1 诚实分歧:原型许「不再重复编排」而服务端重抽签法许同款再现,文案从行为),零欠账闭;后者重建=home 续测卡紧凑形+直跳 /match,形态/去向差系 R2 已注册项维持挂账(卡片 vs 整页取舍=P 票),无 sheet 对齐义务;pay-confirm/share-invite 未落地(gated,落地时先验各自 oracle 形状再谈对齐);取消 autoFocus 无障碍增强保留 |
| G-202 | BLOCKED pane「去意见反馈」CTA | 契约门控 | 跳 feedback-submit 页 | soon toast「即将开放」 | 红:feedback 契约未建(P13 消息中心+反馈召回在窗),门控占位;契约落地时接真跳 |
| G-203 | 合并确认页保留卡掩码 | 策略 | 原型单手机号场景无此歧义 | 多凭证账号显示第二号(7777)非登录主号(6666) | 面检毕(11-01 代码级全链追查):**现行法=最近绑定 sms 胜出**——esports 适配器覆写循环(client/system.go:153-161)×system 凭证列表序,DB 序=绑定序;与走查所见(7777)互证;主号语义候选修正为「保持现状(最新绑定)/改登录主号/改首凭证」——**已裁保持现状(11-02 D-BC,lzx 证据了结+用户授权)**;改动面=适配器一处(+roster 投影 data.go 同构处若求一致) |
| G-204 | 解绑/换绑 | 缺面 | 原型 security 页注记「解绑请联系客服」 | 同文案照录,无自助解绑 | 白名单形②:E6c-b 裁定重绑解绑零建设走客服;文案照原型 |
| G-205 | 设置密码页手机号输入 | 状态 | 原型设置态仅掩码+验证码 | 增 typed phone input(placeholder 含掩码) | 白名单形②:服务端 ResetPassword 证明=码到绑定手机,客户端只持掩码无法发码——设计缺口以客户端补全闭合,契约零改动;原型无此场景交集 |

计数:5 条 = 红 1(G-202 递延有主) + 待面检 1 + 白名单形② 2 + 已修收口 1(G-208 首刀 10-30);增量 G-206 已裁修复(10-30 真机回归裁定)+G-207 白名单②;G-208 余族审计毕(11-01):match-switch/resume-remind 双腿=原型整页非浮层,误挂销、零欠账;余 gated 两腿(pay-confirm/share-invite)=落地时先验各自 oracle 形状再谈对齐。
对齐判定:原型 n31/32/33/34 全文案逐字誊录(冲突卡/凭证校验卡/琥珀警告/双闸 sheet 文案/blocked pane 全段——「如需合请走人工」原型笔误修正为「如需合并」);success 页文案=block3 微信语境照录,短信腿场景共用(原型无短信腿 success 变体);四场景全态 live 走查(短信腿/微信真人腿/KEEP_OTHER/BLOCKED)含 DB 断言与净清。

## 四批 · admin 菜单面(2026-10-30,用户发现「菜单规划/菜单名不对」)

出处:真值=`_nav.js` CFG(九组 33 项,5 非本期);实际=路由树 staticData 机提(esports 四叶+messages zh-CN 解键)。详证=verify/admin-nav-anchor.md(N-1..N-5)。

| # | 面 | 类型 | 期望(原型) | 实际(重建) | 处置 |
|---|---|---|---|---|---|
| G-301 | 菜单组结构 | 功能规划 | 九组按业务域(用户与会员/题库与测试/定价与订单/渠道与推广/反馈与客服/系统与合规…) | esports 四叶全塞单组「电竞」(order 60) | ✅绿(10-31 M2-1 走查):esports 自有四组按原型落位(用户与会员/题库与测试/报告与内容/渠道与推广),骨架五组+IAM 六叶原名保留(D-AY);chrome 走查逐组展开核 |
| G-302 | 菜单命名族+概念错位 | 文案 | 用户列表/激活码管理/游戏与因子配置;01-user-detail=用户详情(用户域) | 会员/激活码/测试项;$id 名「会员详情」且消费 MemberService.getMember(用户/会员两域混置) | ✅绿(10-31 M2-1 走查):用户列表/用户详情/会员管理/会员详情/游戏与因子配置/激活码管理全正名,users 族 URL 迁 /dashboard/people/;走查连带修 users-table 详情链接旧路径残留 |
| G-303 | mutation 错误静默 | 机制 | 列表/详情操作失败给运营可见反馈 | Grant 404(名册外 user_id 被拒)后 Modal 无提示静默停留 | ✅已消解(11-01 live 复现反证):同场景重放=GrantMembership 404 后 sonner 吐司「请求未成功 未找到有效会员记录」实读,Modal 留置=可重试正形;链路=骨架 errorMutationCache(router.tsx:21,mutation 默认 feedback:true)→errorDomain→sonnerFeedback——「全域无 onError 展示」不成立,全局缓存级覆盖,勿逐点加 onError;10-30 所见疑为当帧未及新接线或吐司即逝 |
| G-304 | 开通 Modal 的 ID 可见性 | 可用性 | 运营从列表直达开通对象 | 手动开通 Modal 要 user_id(20 字符),但用户列表显示 xid=account;user_id 仅用户详情页「用户 ID」行可见 | 走查实测:列表→开通需经用户详情转一手;候选修法=Modal 接用户搜索选择器,随下轮会员面增量裁 |
| G-306 | 订单搜索手机号 | 契约缺口 | 订单列表搜索框「订单号 / 昵称 / 手机」(04-order-record) | keyword=订单号前缀+昵称(profile.ListUserIDs);手机反查需 system 名册投影,契约无 | M2-2 走查登记(10-31):system ListUsers 无手机模糊过滤投影,跨服务契约增量留 system 车道;触发器=运营反馈手机找单高频 |
| G-305 | 会员列表全表 fold | 效率 | 列表读面随分页/过滤下推,响应量级随页面不随台账 | ListMembershipRows(nil) 每请求载全量台账(每会员每期一行)入内存 fold/过滤/排序,续费堆叠+MIG1 旧码导入后线性恶化 | M2-1 审查登记(10-31):M1 千级规模内存 fold 为片内明选取舍;触发器=台账过万行或 MIG1 预演,届时 SQL 聚合/窗口下推 |

计数:6 条=绿 2(G-301/302,10-31 走查转)+✅已消解 1(G-303,11-01 live 复现反证:全局错误反馈链实存)+红 3(G-304/305 可用性+效率取舍,G-306 手机搜索契约缺口)。

## 五批 · PAY3 付费墙真通道腿(2026-11-05,片2 走查日登记)

出处:pay-confirm 三态(P7 稿)+旧原型「模拟收银台三态」;实际=工作树 `/pay` @ 片2 未提交态,mock+native 全链 live 走通(态②翻面/收银台三钮/MarkPaid RPC/轮询收敛/sessionQuery 失效),真通道三腿中唯 QR 值展示形在案。三红条皆=白名单形①(契约/服务端没有的面不造):

| # | 面 | 类型 | 期望(原型) | 实际(重建) | 处置 |
|---|---|---|---|---|---|
| G-401 | H5 通道 client_ip | 服务端缝 | 微信外 H5 可下单(pay-confirm 文案面) | **✅ 绿 11-05(片3)**:order/service.go clientIPOf 双形状(transport.Header 读 XFF/X-Real-IP,connect 包装器 *connect.Transport.RemoteAddr 与 kratos *http.Transport.Request 两型落 peer 地址)→CreateIn.ClientIP→Payer 组装第三肢;PayOrder 落库 channel=wechat_h5 credential=jump_url,手机仿真 mock 全链 live 走通,PAY_E2E 增 h5 无 IP 拒/有 IP 过双臂 | 绿:微信侧真通道联调随 PAY2/cutover 承接 |
| G-402 | JSAPI 通道 payer_openid | 契约缺口 | 微信内 JSAPI 可支付 | 会员面契约无处可读自身 openid(仅 OAuth 交换知悉);客户端报 wechat_jsapi 必被渠道拒(require openid) | 红:**lzx 建议裁②服务端按绑定富集**(D-BC 报备形,2026-11-05;证据=JSAPI v3 openid 仅服务端统一下单入参,前端 invoke 只消费 prepay 参数,客户端零消费方——契约暴露=凭据无收益扩散面);真通道刀执行时落,客户端 invoke 分派代码已在案 |
| G-403 | QR 腿渲染 | 依赖缺口 | 桌面 QR 扫码支付 | QR 凭据以值文本+「我已完成支付」轮询 sheet 呈现(无码图) | 红:qrcode 依赖引入+码图渲染随真通道联调刀;mock+native 走查不受影响(provider=mock 恒走收银台 sheet) |
| G-404 | H5 真通道跳转回程 | 服务端缝×客户端 | 渠道收银台跳回后落支付结果页 | 11-05 两轴审查发现:jump 通道 returnUrl=墙自身 URL,微信回跳后墙重挂载无 orderNo 上下文→轮询不续、「三钮全落收据」仅 mock/QR 出口成立 | 红:修形随真通道刀=returnUrl 改指 /pay/result?order=…&face=…(或墙持单入 search);另 clientIPOf 信任拨盘(XFF 优先 vs peer 地址优先)同刀裁——现注释口径=现行 edge 拓扑保证,直暴拓扑归部署面剥头 |

计数:1 绿 3 红(G-401 服务端缝已闭;余 G-402 契约裁决、G-403 依赖件、G-404 回程+信任拨盘=真通道联调刀);mock 全链零红条。

## 六批 · 已登记偏离登记簿(2026-11-05 校准役立)

> 与红条分工:红条=「按原型严格还原」的工单;本批=**契约/产品口径级已裁定或已登记的偏离**(白名单形②仓法/③真实生命周期),非工单——集中登记防散失,各随对应裁决窗复核。出处=切片收口登记(原散于 plan/progress 巨格,2026-11-05 瘦身时收拢)。

| # | 面 | 原型/旧码 | 现行实现 | 依据/复核窗 |
|---|---|---|---|---|
| R-01 | MSG-1 消息类型与范围 | 三型四范围(含触达已读统计) | 两型两范围,无触达已读统计 | 白名单①(契约窄化显式登记);W6 ljr |
| R-02 | MSG-1 contact 上限 | 输入 40 字 | 128 | 同上;W6 ljr |
| R-03 | pay-result cancel 文案 | 「未产生订单」 | 「未产生扣款」(真实订单生命周期 CLOSED 形,订单存在无扣款) | 白名单③(真实生命周期);真通道刀随检 |
| R-04 | terminated 会员 /member 呈现 | (原型无此态) | 折 never 形(解锁内容仍可读,灰条/计数行不显);e2e 钉死 | 白名单②(无 oracle 面);运营/真通道刀复核 |
| R-05 | member FAQ 硬价 | ¥19.9/五折逐字 | 文案硬编码键(messages/member),调价需改文案 | 白名单②;运营调价时 |
| R-06 | game-config 呈现 | 难度三档权重/上下架四列+因子抽屉 | 运行配置平键值表(更贴 test_config schema 真相) | 白名单②(schema 更真);M2 余组 test-type 档案面 |
| R-07 | consent 渠道列 | 列=小红书/抖音/归因 | 契约无 channel 字段,列未建 | 白名单①;P23 归因记账(遗留清单并) |

