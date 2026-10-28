# Gap 台账

> 每条含:期望(原型字面量)/实际(重建渲染)/处置。处置:红=按原型严格还原的工单;待面检=需按态比对定性。存疑不阻塞(原型固化律,见 README)——一律转 pm-review-items.md 留档。
> 机检来源:e2e 比对 spec 的 diff 输出(app `e2e/prototype/*.fidelity.spec.ts`,钉 zh-CN);手工来源:走查记录。基线注明工作树状态。

## 首批 · V-1 login 样板(2026-10-28)

出处:`prototype/login.html`(锚=app/clients/ctr/e2e/prototype/fixtures/login.json);实际=工作树 `/login` @ zh-CN(E5-f3 未提交态,但 login 面不属该切片)。

| # | 面 | 类型 | 期望(原型) | 实际(重建) | 处置 |
|---|---|---|---|---|---|
| G-001 | 品牌 icon | 缺行 | ⚡ | (无) | E10-a 对账:视觉在(lucide Zap),文本 oracle 不可见=emoji↔SVG 通道差;随 PM-4 裁 |
| G-002 | 默认登录 tab | 状态 | 默认=验证码登录:手机号/6 位验证码/发送验证码 | 默认=密码登录 | 依赖 E6c(SMS provider 落地时默认翻,segment 序已按原型排好零移节点);PM-3 在档 |
| G-003 | 账号字段语义 | 文案 | placeholder=手机号 | 请输入账号 | ✅修复(10-28 E10-a;原型两 tab 皆「手机号」已渲染态实证;手机号登录功能随 E6c) |
| G-004 | 表单字段标签 | 多行 | 无标签(仅 placeholder) | 多出「账号」「密码」两行标签 | ✅修复(10-28 E10-a;sr-only 文本泄漏 innerText 实证→改 aria-label) |
| G-005 | 非微信提示 | 状态逻辑 | 微信内提示区为空;非微信态显示该文案 | 浏览器常显(与非微信态一致) | 文案逐字同+UA 条件已对(渲染态探针);态比对=微信 UA 下应空,harness 执行;PM-5 在档 |
| G-006 | 底部链接行 | 标点 | 忘记密码？ · 用户协议 · 隐私政策 · 返回入口 | 忘记密码？·《用户协议》·《用户协议》·返回入口 | ✅修复(10-28 E10-a;裸名键+「 · 」字面空格——原型空格=源码空白折叠,margin 不入 innerText 实证) |
| G-007 | 合规主体 | 标点+缺行 | 浙ICP2026028293号;另有独立行 客服邮箱 xgwledu@126.com | 浙ICP备2026028293号;客服邮箱行缺 | 邮箱行在(live 实测);ICP 带「备」=D-AO 压倒原型笔误,PM-1 在档 |
| G-008 | 杂行 | 多行 | — | 「xl」 | dev-only 实证:TailwindIndicator 有 DEV 守卫,prod 恒 null;已标 data-dev-chrome,harness 剥离(devtools 的 Filter placeholder 同族=vite 插件 prod 剥离实证) |
| G-009 | 页面标题 | 文案 | ctr · 登录 | 登录 · CYBER ESPORTS | 格式已还原(品牌·页名,44/44 页标题惯例实证);品牌词=PM-2 待裁 |

计数:9 条 = 修复 3 + 依赖 E6c 1 + 已裁压倒 1(ICP) + 通道差随 PM 裁 2(⚡/品牌词) + 按态比对 1 + dev-only 1;物料挂接 PM-1..PM-5。
E10-a 复对账(10-28,live dev 栈,同款 innerText oracle):14 行中 12 行逐字全同,余 2 行=已裁(ICP 备)+态条件(非微信注记);placeholder 对(手机号/密码（≥6 位）)逐字同。
