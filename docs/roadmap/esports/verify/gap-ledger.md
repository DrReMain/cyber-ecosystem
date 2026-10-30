# Gap 台账

> 每条含:期望(原型字面量)/实际(重建渲染)/处置。处置:红=按原型严格还原的工单;待面检=需按态比对定性。存疑不阻塞(原型固化律,见 README)——一律转 pm-review-items.md 留档。
> 机检来源:e2e 比对 spec 的 diff 输出(app `e2e/prototype/*.fidelity.spec.ts`,钉 zh-CN);手工来源:走查记录。基线注明工作树状态。

## 首批 · V-1 login 样板(2026-10-28)

出处:`prototype/login.html`(锚=app/clients/ctr/e2e/prototype/fixtures/login.json);实际=工作树 `/login` @ zh-CN(E5-f3 未提交态,但 login 面不属该切片)。

| # | 面 | 类型 | 期望(原型) | 实际(重建) | 处置 |
|---|---|---|---|---|---|
| G-001 | 品牌 icon | 缺行 | ⚡ | (无) | E10-a 对账:视觉在(lucide Zap),文本 oracle 不可见=emoji↔SVG 通道差;随 PM-4 裁 |
| G-002 | 默认登录 tab | 状态 | 默认=验证码登录:手机号/6 位验证码/发送验证码 | 默认=验证码登录(E6c-a 后) | ✅修复(10-28 E6c-a bf0a99a0,双通道 live 走查随片 proven-by;10-30 验证会话复跑=代码级:useState 初值 "code"+segment DOM 序 code 先,page.tsx:66/239);PM-3 随之收口 |
| G-003 | 账号字段语义 | 文案 | placeholder=手机号 | 请输入账号 | ✅修复(10-28 E10-a;原型两 tab 皆「手机号」已渲染态实证;手机号登录功能随 E6c) |
| G-004 | 表单字段标签 | 多行 | 无标签(仅 placeholder) | 多出「账号」「密码」两行标签 | ✅修复(10-28 E10-a;sr-only 文本泄漏 innerText 实证→改 aria-label) |
| G-005 | 非微信提示 | 状态逻辑 | 微信内提示区为空;非微信态显示该文案 | 浏览器常显(与非微信态一致) | 文案逐字同+UA 条件已对(渲染态探针);态比对=微信 UA 下应空,harness 执行;PM-5 在档 |
| G-006 | 底部链接行 | 标点 | 忘记密码？ · 用户协议 · 隐私政策 · 返回入口 | 忘记密码？·《用户协议》·《用户协议》·返回入口 | ✅修复(10-28 E10-a;裸名键+「 · 」字面空格——原型空格=源码空白折叠,margin 不入 innerText 实证) |
| G-007 | 合规主体 | 标点+缺行 | 浙ICP2026028293号;另有独立行 客服邮箱 xgwledu@126.com | 浙ICP备2026028293号;客服邮箱行缺 | 邮箱行在(live 实测);ICP 带「备」=D-AO 压倒原型笔误,PM-1 在档 |
| G-008 | 杂行 | 多行 | — | 「xl」 | dev-only 实证:TailwindIndicator 有 DEV 守卫,prod 恒 null;已标 data-dev-chrome,harness 剥离(devtools 的 Filter placeholder 同族=vite 插件 prod 剥离实证) |
| G-009 | 页面标题 | 文案 | ctr · 登录 | 登录 · CYBER ESPORTS | 格式已还原(品牌·页名,44/44 页标题惯例实证);品牌词=PM-2 待裁 |

计数:9 条 = 修复 4(G-002 10-28 E6c-a 落,10-30 复跑收口) + 已裁压倒 1(ICP) + 通道差随 PM 裁 2(⚡/品牌词) + 按态比对 1 + dev-only 1;物料挂接 PM-1..PM-5(PM-3 已收口)。
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
| G-208 | 登出确认 sheet 形状 | 浮层 | 竖排按钮(取消上/退出下,btn-pri 蓝青渐变)+上圆角 18px+15px 标题+全角问号 | 横排按钮+退出=红 danger+四角圆角+14px+半角问号 | **首刀已收(10-30)**:BottomConfirmSheet 单源组件(CTS.sheet 正身:390px/上圆角 18px/竖排全宽 gap 9px/三色族 primary·hot·ghost——红只留真危险),exit-confirm 重写+logout-sheet 薄壳+merge sheet 三迁,原型逐字文案+chrome 双侧 7 项几何实测全同;**余族待扫**(pay-confirm/match-switch-confirm/resume-remind/share-invite)=各对应页面片落地时逐一对齐;取消 autoFocus 无障碍增强保留 |
| G-202 | BLOCKED pane「去意见反馈」CTA | 契约门控 | 跳 feedback-submit 页 | soon toast「即将开放」 | 红:feedback 契约未建(P13 消息中心+反馈召回在窗),门控占位;契约落地时接真跳 |
| G-203 | 合并确认页保留卡掩码 | 策略 | 原型单手机号场景无此歧义 | 多凭证账号显示第二号(7777)非登录主号(6666) | 待面检:微信腿走查实证;掩码取 credential 列表序,主号语义待 PM 定(候选=最近登录凭证/首凭证) |
| G-204 | 解绑/换绑 | 缺面 | 原型 security 页注记「解绑请联系客服」 | 同文案照录,无自助解绑 | 白名单形②:E6c-b 裁定重绑解绑零建设走客服;文案照原型 |
| G-205 | 设置密码页手机号输入 | 状态 | 原型设置态仅掩码+验证码 | 增 typed phone input(placeholder 含掩码) | 白名单形②:服务端 ResetPassword 证明=码到绑定手机,客户端只持掩码无法发码——设计缺口以客户端补全闭合,契约零改动;原型无此场景交集 |

计数:5 条 = 红 1(G-202 递延有主) + 待面检 1 + 白名单形② 2 + 已修收口 1(G-208 首刀 10-30);增量 G-206 已裁修复(10-30 真机回归裁定)+G-207 白名单②;G-208 余族(pay-confirm/match-switch-confirm/resume-remind/share-invite)=各页面片落地时对齐。
对齐判定:原型 n31/32/33/34 全文案逐字誊录(冲突卡/凭证校验卡/琥珀警告/双闸 sheet 文案/blocked pane 全段——「如需合请走人工」原型笔误修正为「如需合并」);success 页文案=block3 微信语境照录,短信腿场景共用(原型无短信腿 success 变体);四场景全态 live 走查(短信腿/微信真人腿/KEEP_OTHER/BLOCKED)含 DB 断言与净清。

## 四批 · admin 菜单面(2026-10-30,用户发现「菜单规划/菜单名不对」)

出处:真值=`_nav.js` CFG(九组 33 项,5 非本期);实际=路由树 staticData 机提(esports 四叶+messages zh-CN 解键)。详证=verify/admin-nav-anchor.md(N-1..N-5)。

| # | 面 | 类型 | 期望(原型) | 实际(重建) | 处置 |
|---|---|---|---|---|---|
| G-301 | 菜单组结构 | 功能规划 | 九组按业务域(用户与会员/题库与测试/定价与订单/渠道与推广/反馈与客服/系统与合规…);原型无「电竞」组 | esports 四叶全塞单组「电竞」(order 60) | 红:随 M2-1 重构 esports 自有组;系统与合规组=骨架直接承接(D-AY)不动 |
| G-302 | 菜单命名族+概念错位 | 文案 | 用户列表/激活码管理/游戏与因子配置;01-user-detail=用户详情(用户域) | 会员/激活码/测试项;$id 名「会员详情」且消费 MemberService.getMember(用户/会员两域混置) | 红:随 M2-1 改键正名(两域四页分立);IAM 面名不涉(D-AY) |

计数:2 条=红 2(均随 M2-1;首批 admin 红条)。
