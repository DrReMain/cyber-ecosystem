# esports 进度档 — 认领状态与验收记录

> 归属:esports 前沿进度档,三人共写。认领即登记;里程碑收口时验收记录剪枝为一行收口注(提交 hash + 判定),长历史删除。检查点结果每站在下表登记。

## 切片认领表(目标窗口见 plan.md 时间轴)

| 片 | 认领人 | 分支 | 状态 | 备注 |
|---|---|---|---|---|
| E0 契约首刀 | lzx | esports/dev-lzx | 种子片完成(待评审);增量2 完成(待评审);**增量3 完成(待评审)=统一提交契约** | D-U 增量推进;2026-10-18 种子=画像域+游戏目录读面(proto 3 文件,生成/lint/breaking/ctr 链全绿,catalog manifest 已挂联邦树);性别不入选转 P8(ljr);错误码 61xx 段;2026-10-18 增量2=测试域读面冻结落盘:test.proto 三读(ListTests 水合源/GetTest 详情+introduction/GetTestConfig 引擎启动包 duration 镜像+background_invalid 机制阈值+engine_params KV);设计侧对比修正 5 处(ability 六词 camelCase 原形、档位百分位服务端判 GradeBand 撤回、unit_score_label 补、introduction per-test 超集、ListTests 角色重定位水合源);max_score 删除获设计侧先例(admin test-type 已删满分);ESPORTS_TEST_NOT_FOUND=6110(404 码系)+ctr 双语文案;全链绿(generate/breaking/esports build+test/ctr build→typecheck→check);target_recognition 能力冲突转 P9;题库两读出列随 E7。**2026-10-20 增量3=SubmitTestAttempt 统一提交**(设计稿四决策经用户裁决:汇总标量不上行/attempt_no 占键/attempt 详情回读归 E5/upload·rank·pk 不并入):单 RPC 替旧版 14 端点(新版 13 款=去 skill_prediction),「事实上行、判定下行」——明细 oneof 13 款逐款按旧版 SubmitData 语义翻译(proto 行尾 legacy 对照即评审锚),弱型两款(NumberStars roundHistory/SkillTrajectory rounds)经旧引擎 types 强型化;事件时间统一 elapsed_ms(相对 attempt 单时钟,防墙钟篡改面);MovingTarget 补 pattern 相位、MultiTarget 补 tier、TargetRecognition/OffenseDefense 去 is_correct(服务端判);BackgroundSpan 引擎标记+服务端判 valid(invalid_reason 回执);响应=双轨分(raw/authoritative+grade 四档)+ability 回执+retest_left 余额位(E4 执法);admin 记录读面(ACCESS_ADMIN)注记挂 E9 消费本契约存储形状;门禁 generate/format/breaking/esports build/ctr 三链全绿 |
| E1 服务骨架 | lzx | esports/dev-lzx | 完成(待评审) | 2026-10-18 kratos-new-service 八步全走:profile schema 从 E2 前拉(ent 零 schema 不可生成,键形按 AGENTS §8=user_id+租户列);init 迁移已 apply;catalog 经 manifest 自动挂树;部署面全接(compose/traefik/02-databases/deploy 目标/admin vite 代理);shared-go 注解测试纳入 esports;会员链探针 e2e 绿(登录→自省座位→GET profile 200 全 null 惰性态,无 cookie 401)。GameService 注册随 E2(目录种子先行才有意义) |
| E2 schema+种子 | lzx | esports/dev-lzx | **已撤回(D-V 业务前置律)** | 2026-10-18 当日先落地后撤回:game 聚合+9 款种子+画像写面全规则曾全绿(探针 16 断言),用户裁决接通先行后整体撤回(备份 /tmp/e2-loop-backup/ 至重启,契约重建即据冻结 proto;探针脚本 /tmp/probe_loop.py);服务回到 E1 探针形态复验绿(login→401→惰性 GET;写面 501)。重启时机=接通闭环验证过 |
| E3 评分引擎 | — | — | 未认领 | XL,可分批(Tier A 前半/后半、Tier B、档位层) |
| E4 匹配引擎 | — | — | 未认领 | |
| E5 报告框架+能力报告 | — | — | 未认领 | 服务/客户端两半可并行 |
| E6a ctr 纯壳 | lzx | esports/dev-lzx | 完成(待评审) | 2026-10-18 先行开工(例会追认);技术栈 D-R;主题 dark-only + i18n zh/en(D-S);同日补入 SEO 全量面(public:robots/sitemap hreflang/noindex 流程页/逐页 title);对照 admin 全量审计:20/33 字节级一致,13 处偏离全部参数化有据(生成器前置标准),4 处漂移当日修正;收口:三链绿 + live 冒烟全绿(含 404 错误面 LocaleSelect);同日 D-T:收敛为仅根路由+空白落地页(能力配置为交付面),login/profile/test 路由/页面/消息目录随之移除(E6b 登录片按范式重建),home 消息留 title+SEO description |
| E6b 登录+画像流 | lzx 主刀 | esports/dev-lzx | **完整形态片完成(待评审)=登录/画像/首页照原型语义级成型** | 2026-10-19 接通片:微信真人链路全绿(静默授权→code 交换→自动注册→HttpOnly cookie→惰性画像;LoginByIdentity 200→0.3s 后 GetProfile 200 双证);关键修复=回调 Outlet 正典拆法/system config 门控自动注册/vite strictPort/shared-transport+Sentry 对齐;骨架演进件按分支纪律拆 dev 4 笔回流。**2026-10-20 完整形态片**(消费既有契约,零后端改动):login 完整形态=品牌头+双 Tab 段(验证码槽占位禁用,SMS 无契约)+协议双勾+全选联动+指名拦截 toast+微信环境门控(UA 后置判定防 SSR 水合漂移)+法务尾;画像完整性门(昵称+出生年月)接管登录后路由——next 回跳优先,不全落向导昵称步,全进首页(login/微信回调共用 routeAfterLogin);三步向导 /profile/setup/{nickname,age,games}(缺中间层路由容忍,profile.tsx 降布局+profile.index 总览):昵称步镜像契约 2-12 码点+规则卡,年龄步 6-60 岁+派生八档即时展示+未满 13 监护人硬闸(独立勾选,guardianAccepted 随 birth 同发)+未满 18 提示,游戏步按目录 category 分组+类型投票注(未归类不计票)+自定义≤3 capbar,每步独立 UpdateProfile 落库(向导/编辑双用途);个人中心九区=身份卡(档位·N 款偏好 meta)/资料三行(复用向导步)/安全三行未开放/更多模块虚线占位/关于我们/退出底弹层(RAC ModalOverlay 底停,cancel autoFocus);首页=问候态机+双门控 CTA(未登录→login?next=/,无偏好→游戏步,齐备→E7 占位门);共享 TabBar 四槽两活(H1 ambient 血统,features/app 深导入避 hzy 合流点);H5 触屏适配(≥44px 热区/active 压反馈/无 hover 依赖/390x844 触摸仿真走查);图标全 lucide 零 emoji(用户裁决);ctr 三链绿+桌面/移动双走查全路径(完整画像直进首页/空画像三步向导闭环/退出弹层/TabBar 占位门);微信真人复验待用户参与。**2026-10-20 机制修复(用户上报未登录渲染错误)**:admin 会话守卫机制同形移植——features/auth session.fn(会员链探针三态吸收:anonymous/authed+profile/expired/unreachable,401 不入错误域)+guard(/profile 布局 beforeLoad,未登录 redirect login?next=);首页 getProfile 改按 session verdict enabled 门控(游客零 401),routeAfterLogin 先清 session 缓存防守卫回弹;SSR/客户端双路复验绿(未登录首页=200 游客面,未登录 /profile=307→login,真登录态 SSR 昵称首屏直出)。**同日退出弹层双修(用户上报:取消无反应+safe 遮挡)**:①根因=RAC 1.21.1 受控 ModalOverlay 退出状态机死锁——useExitAnimation 的 exiting 态卡死(ref 空致 onEnd 永不调,isOpen=false 永不卸载),cancel/遮罩/库内 dismiss 全失效,仅路由跳转可强拆;修法=**挂载条件化**(open 直接拥有挂载周期,关=卸载,无 exit 相),LogoutSheet+ExitConfirmDialog(H1 kit 零消费者同构雷同修,hzy 领地跨界注记);②Escape 链另证 filterDOMProps 白名单只放 pointer 类、onKeyDown* props 被丢弃,库内链又断在 onOpenChange 之前——按用户裁决 H5 无 ESC 不做键盘通道;③遮罩关闭走原生 onPointerDown(target=overlay 即关,pointer 类白名单放行);④safe-mb 升级 max(env,12px) 地板——原 m-3(margin 简写)与 safe-mb(margin-bottom)同类名冲突,产物顺序令无地板值胜出,桌面/无 env webview 贴底 0px 被_home indicator 压住。桌面+390x844 仿真回归绿(cancel/遮罩/重开无残留/confirm 登出全链)。**2026-10-20 全项目 review 修复片(R1+R2)**:机制=profile 死错误分支删(全局 throwOnError 下交边界)/三步 mutateAsync 补 try/catch(服务端校验失败不再落 window-capture 进 Sentry)/games key fallback+添加按钮真 disabled/next 开放重定向卫生(protocol-relative 双层拒绝:validateSearch+routeAfterLogin,实测 //example.com 被剥参)/回调失败面 a→Link;H5=base.css 全局 button cursor+tap-zone 伪元素热区(9 处小按钮 ≥44px);a11y=setup 输入补 name+aria-label(DevTools form-field issue 消失)/行内错误 role=alert/StateScreen loading 补 sr-only;结构=profile 四卡+login AgreeBox 拆 ui/(一文件一组件法)/7 处手写渐变收编 bg-cta-primary+新 token(accent-fps/moba/info、wechat、identity、line-accent-blue)/SoonCard 抽共享件(components/,双份 markup 收敛+badge 键升 common_soon_badge+孤儿键清理)/profile loading 换 StateScreen(kit 首消费者)/tab_* 改 common_tab_*(§6 前缀法);三链绿+桌面复验(role=alert 播报/弹层回归/卫生实证/games a11y 清);转交项见「Review 遗留转交」节 |
| E7 游戏框架+13 款 | — | — | 未认领 | XL 三批,批内可并行 |
| E8 权益最小面 | — | — | 未认领 | |
| E9 最小 admin 面 | — | — | 未认领 | |
| E10 全链 e2e | — | — | 未认领 | |

## Review 遗留转交(2026-10-20 ctr 全项目 review;修复已落本分支,转交项拉取后自取)

| 去向 | 事项 | 出处 |
|---|---|---|
| ljr | 首页未登录文案:问候行与标题均渲染「未登录」叠字(home_greet_guest/home_guest_name),请定文案方向(如问候行改「先登录再开始」/标题改「玩家」) | review D6 |
| ljr/部署确认 | CSP `frame-ancestors 'none'`(start.ts)与「租户 app iframe 内嵌」场景冲突待确认;微信/原生 webview 顶层浏览不受影响,若需 iframe 内嵌须改租户域白名单 | review B3 |
| hzy | StateScreen loading 的 LOADING 扫光条与 PendingFallback markup 双份复制,lzx 已跨界最小修 sr-only 播报(features/app/states/state-screen.tsx);视觉收敛(抽共享扫光条)归 features/app 领地自决 | review C3 |
| 后续片 | 微信回调 next 丢失:redirect_uri 不能带 query(微信 authorize 校验拒),微信登录后回跳意图丢失,固定走向导/首页;如需保 next 须 state 中转,微信链路设计时定 | review A6 |

接受不改注记:B4 `user-scalable=no`(游戏 H5 防双击缩放,WCAG 张力由产品定位接受);D8 stores 空壳(D-T 能力配置面,机制待命)。
| H1 ctr 系统态件 | hzy | esports/dev-hzy | 完成(待评审) | 2026-10-20 交付 features/app 五件:AmbientBackdrop/AmbientLayout(scroll page\|inside+dense 三态,maxWidth 四档)、StateScreen(empty\|loading,loading 承 PendingFallback 血统 mono+扫光)、Skeleton(原子块,宿主组 a11y)、ExitConfirmDialog(受控,首个 react-aria-components 消费者;仓内 1.21.1 无 AlertDialog/OverlayTrigger,实配 ModalOverlay+Dialog role=alertdialog;ESC/遮罩/取消键统一走 onCancel,confirm 不自动关,cancel=autoFocus 防误触);error 件沿用 E6a 既有族不重复造;tokens 增 ambient 三色+漂移/扫光 keyframes(@theme 单源,PendingFallback 内联 keyframes 收敛入 tokens);common 增 6 键双语;三链绿+SSR 走查断言 10/10(临时挂 home 后还原);命中 §6 paraglide 陷阱一次(build 跳过再生掏空 _index,rm -rf 强制再生修复) |
| H2 协议页+版本化承载 | hzy | esports/dev-hzy | 未开工 | agreement/age-required 页面+路由+版本参数;文案占位等 ljr 定稿;consent 裁决面归 E6b(lzx)。**URL 形状已定约(2026-10-20 lzx)**:/agreement?p=user|privacy(p 缺省=user;版本参数随 consent 契约后加),登录页协议链接届时从占位 toast 切真链 |
| H-A 引擎框架片 | hzy | esports/dev-hzy | 未开工 | E7 框架前拉:GameClock 单实现(墙钟+暂停顺延,fake timers 可测;旧版 13 处手搓计时收敛为 1)+参数化路由 /test/:code+GameModule 插件契约(运行时禁 import auth/services)+overlay 族;config 消费用 gen 真类型(GetTest/GetTestConfig 已冻结,mock 只落 transport 缝);验收=mock config 下 red_dot 可玩+时钟正确性测试 |
| H-B 反应瞄准批 | hzy | esports/dev-hzy | 未开工 | E7 首批前拉:red_dot/flash_aiming/multi_target/target_recognition;harness 免登录;每款=运行时+rawDetail 采集+本地结果页;不烤权威指标(评分归服务端,旧架构同构);判定阈值来自 config;rawDetail shape 反哺 E0 统一提交契约;玩法语义级对照旧版(D-O);提交适配器等 E0 冻结后薄接,游戏零改动 |
| PAY1 订单域 | — | — | 未认领 | 依赖 P1(10-28) |
| PAY2 微信支付接入 | — | — | 未认领 | 商户号 W1 确认 |
| PAY3 合规执法+付费墙 | — | — | 未认领 | 依赖 P7(12-02) |

## 检查点登记(D-M)

| 检查点 | 日期 | 判定 | 拉闸动作 |
|---|---|---|---|
| CP1 | 11-08 | 待检 | — |
| CP2 | 11-22 | 待检 | — |
| CP3 | 12-02 | 待检 | — |

## 里程碑验收记录

(空;收口时一行一注)

## 待裁项状态

见 README §2.2;分析进展记 prd.md 各条目。

## 节奏记录

- 2026-10-18 立项+目录化;M1 切片池就绪。
- 2026-10-18 grill 定稿(D-K..D-Q):60 天作战计划落档,承诺日 12-17(含支付);大框架冻结,变更走 README 台账。
- 2026-10-18 D-R:ctr UI 基座 = react-aria+tailwind(无 antd);E6 拆 E6a 纯壳/E6b 登录+画像流,lzx 先行建壳。
- 2026-10-19 hzy 外围启动线定形(H1/H2/H-A/H-B):登录重构与 E0 余面为承重面留 lzx,hzy 全部切片圈定 ctr 非登录区(文件面与 lzx 线不相交,合流点仅 routeTree.gen+未来一次提交适配器);E7 框架与首批经 H-A/H-B 前拉,rawDetail 采集面反哺 E0 统一提交契约。admin 只读面前拉提案否决:BUILTIN 会员面(ListTests 等)不得作运营消费面,一面一消费等级,运营读面随 E0 余面/E9 契约设计走。
- 2026-10-20 H1 完成(待评审):系统态五件落 features/app(ambient/states/dialog 三族),首个 RAC 消费者随 D-T 落地;走 tanstack 三链+切片节奏法 onboarding 全程走通。
- 2026-10-20 精修基线盘点:微信真人复验过、最小闭环收口(c9f7924c 已合 main);coverage.md 建立(44 页交互基线全量×新版状态矩阵,语义锚=prototype/_check.js rules 逐条搬进承接片评审锚);抓到缺口=version-update 无承接片,例会定归属。同日增补(用户三指示):D-O2 验收四要素落台账(布局一样/细节接近/文案一样/视觉接近——原型系 AI 产出+组件库约束,不追像素级);admin 端 41 页×9 组矩阵并入(双端覆盖,系统与合规组=IAM 底座面复用);语料版本律(引用先过 DESIGN_DOC_MAP 权威表,快照 09-18 后演进文档经 ljr 定级,评分数值档 E3 单源化后才可依赖)。
- 2026-10-20 E6b 完整形态片契约缺口登记(M1 语义降级已落,契约到位后增量激活):SMS 验证码登录/忘记密码/手机绑定=system 无契约(dysmsapi provider+SendCode+phone credential,建议独立骨架片);协议 consent 版本记录=openspec agreement-consent 未落(H2 页面先行,checkbox 门已客户端执法);性别采集=P8(ljr);年龄段次级口径=UpdateProfile 无 band 入参(仅出生年月单口径);昵称加权 24=契约已冻 2-12 码点(原型 09-16 裁决冲突,待 ljr 追认);自定义游戏上限=服务端常量 20/产品口径 3(客户端按 3 展示,config 读面缺)。
