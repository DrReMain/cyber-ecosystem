# Hero Skin 优化设计文档

## 1. 设计目标

Hero 皮肤定位：现代 SaaS 管理后台的默认推荐皮肤。设计语言强调专业、清晰、有层次感，通过克制的阴影层级和精确的圆角系统建立视觉秩序。

优化核心目标：
1. 补齐当前缺失的组件覆盖，确保所有 storybook 场景下视觉一致
2. 修复 dark 模式下的对比度和可读性问题
3. 强化 compact 模式下的视觉辨识度（不碰尺寸，通过色彩和边框）
4. 统一阴影/圆角/边框的语言体系，消除"某些组件像没换皮肤"的断裂感

## 2. 设计禁区

以下属性在本次优化中**绝不调整**：
- `padding`、`margin`、`gap`、`width`、`height`、`minWidth`、`minHeight`
- `fontSize`（token 中的 `fontSize*` 系列除外，已设定的不动）
- `lineHeight`
- 任何影响布局盒模型的尺寸属性

可安全调整的属性：
- `color*` 系列 token
- `borderRadius*` 系列 token
- `border`、`borderColor`、`borderWidth`、`boxShadow`
- `background`、`backgroundColor`
- `opacity`
- `fontWeight`
- `motionDuration*`

## 3. 色彩体系调整

### 3.1 Light 模式（当前良好，微调）

当前 light token 已成熟，仅需一处微调：

```ts
// 当前
// colorBorderSecondary: "#E4E4E7"

// 优化后
// colorBorderSecondary 保持 #E4E4E7，但增加 Table/Tree 等组件的分隔线显式覆盖
```

### 3.2 Dark 模式（重点修复）

当前 dark 模式存在严重的对比度问题：

| 问题 | 当前值 | 问题描述 | 优化值 |
|---|---|---|---|
| Text/Link 按钮不可见 | `colorTextBase: #ECEDEE` | 在暗色背景上 Text 按钮因为没有背景，文字色和背景色对比度足够，但视觉权重过轻 | 不修改 token，通过组件覆盖给 Text/Link 按钮增加 hover 时的背景色 |
| 输入框 placeholder | 继承 `colorTextPlaceholder` | 在 `#18181B` 背景上偏暗 | 不修改全局 token，通过 Input 组件的 `colorTextPlaceholder` 覆盖为 `#9CA3AF` |
| 默认按钮文字 | `colorTextBase: #ECEDEE` | 在 `#3F3F46` 边框内对比度 OK，但默认按钮整体太融入背景 | 给 default/outlined 按钮增加 `borderColor` 在 dark 下的显式亮度提升 |

**Dark Token 实际调整：**

```ts
const darkTokens: Partial<GlobalToken> = {
  // ... 现有值保留 ...
  // 唯一 token 级调整：
  colorBorder: "#52525B",        // 从 #3F3F46 提升，让边框在 dark 下更可见
  colorBorderSecondary: "#3F3F46", // 原 colorBorder 降级为 secondary
}
```

> 为什么只调 border？dark 模式下最突出的问题是"组件边界消失"，导致界面像一坨色块。提升 border 亮度是最安全、影响面最广的修复。

### 3.3 阴影体系（Dark 模式修正）

当前 hero 的阴影在 dark 模式下使用了 `inset 0px 0px 1px 0px rgb(255 255 255 / 0.15)` 的高光边，这是正确的做法。但需要确保所有面板组件都使用这套阴影体系。

当前阴影定义：
```ts
const shadowSmall = isDark
  ? "0px 0px 5px 0px rgb(0 0 0 / 0.05), 0px 2px 10px 0px rgb(0 0 0 / 0.2), inset 0px 0px 1px 0px rgb(255 255 255 / 0.15)"
  : "0px 0px 5px 0px rgb(0 0 0 / 0.02), 0px 2px 10px 0px rgb(0 0 0 / 0.06), 0px 0px 1px 0px rgb(0 0 0 / 0.3)"
```

**保持不变**，但新增覆盖的组件必须复用这套阴影变量。

## 4. 组件覆盖矩阵

### 4.1 已有覆盖的问题修复

| 组件 | 当前问题 | 修复方案 |
|---|---|---|
| **Button (Text/Link)** | Dark 模式下几乎不可见 | 增加 `textBtnRoot` / `linkBtnRoot` 样式：dark 模式下 hover 时背景为 `rgba(255,255,255,0.06)`，light 模式下为 `rgba(0,0,0,0.04)` |
| **Button (Link)** | Light 模式下缺少圆角 | Link 按钮添加 `borderRadius: cssVar.borderRadiusSM` |
| **Card** | 当前未覆盖，使用 antd 默认 | 增加 `heroPanel` 覆盖，使用 `shadowSmall` |
| **Input** | Dark placeholder 对比度低 | 在 theme.components.Input 中增加 `colorTextPlaceholder: isDark ? "#9CA3AF" : undefined` |
| **Select** | 当前只覆盖了 option 样式 | 增加 popup 的 `boxShadow: shadowMedium`，确保下拉面板与 hero 阴影体系一致 |
| **Progress** | 当前 rail 有 shadow，但 track 没有统一 | 确认 track 样式是否已覆盖，如未覆盖则添加 |

### 4.2 新增组件覆盖

以下组件当前**完全没有** hero 风格的覆盖，在 storybook 中显得像没换皮肤：

#### Tier 1 - 高优先级（组合页面中高频出现）

| 组件 | 覆盖策略 | 具体属性 |
|---|---|---|
| **Table** | 表头背景 + 行 hover + 边框 | `headerBg: colorBgElevated`, `headerColor: colorTextBase`, `rowHoverBg: color-mix(in srgb, colorPrimary 4%, transparent)`, `borderColor: colorBorderSecondary`, `headerBorderRadius: borderRadiusLG` |
| **Tabs** | 卡片式标签页 + 内容区 | `cardBg: colorBgContainer`, `cardGutter: 4`, `cardPadding: "8px 16px"`, `itemColor: colorTextSecondary` |
| **Timeline** | 节点圆点 + 连线 | `itemPaddingBottom: 24`, `dotBorderWidth: 2`, `dotBg: colorBgContainer` |
| **Menu** | 选中项 + hover | `itemSelectedBg: colorPrimary`, `itemSelectedColor: "#FFFFFF"`, `itemHoverBg: color-mix(in srgb, colorPrimary 8%, transparent)`, `darkItemSelectedBg: colorPrimary`, `darkItemHoverBg: color-mix(in srgb, colorPrimary 12%, transparent)` |
| **Steps** | 已完成步骤图标背景 | `iconBg: colorBgContainer`, `iconBorderColor: colorBorder`, `finishIconBorderColor: colorPrimary`, `finishIconBg: colorBgContainer` |
| **Pagination** | 页码按钮 | `itemBg: colorBgContainer`, `itemActiveBg: colorPrimary`, `itemActiveColor: "#FFFFFF"`, `itemBorderRadius: borderRadius` |

#### Tier 2 - 中优先级（独立 stories 中可见）

| 组件 | 覆盖策略 | 具体属性 |
|---|---|---|
| **Tree** | 选中节点 + hover | `nodeSelectedBg: color-mix(in srgb, colorPrimary 8%, transparent)`, `nodeHoverBg: color-mix(in srgb, colorPrimary 4%, transparent)` |
| **Upload** | Dragger 区域 | `draggerBg: colorBgContainer`, `draggerBorder: \`1px dashed \${colorBorder}\``, `draggerBorderRadius: borderRadiusLG`, hover 时 `borderColor: colorPrimary` |
| **Collapse** | 面板头 + 内容区 | `headerBg: colorBgContainer`, `contentBg: colorBgContainer`, `headerPadding: "12px 16px"` |
| **Descriptions** | 标签 + 内容 | `labelColor: colorTextSecondary`, `contentColor: colorTextBase`, `borderColor: colorBorderSecondary` |
| **Badge** | 状态点 | `statusProcessingBg: colorPrimary`, `colorError: colorError`（确认一致性） |
| **Avatar** | 组叠层边框 | `groupBorderColor: colorBgContainer` |
| **Divider** | 文字分割线 | `colorTextHeading: colorTextSecondary`, `textPaddingInline: 16` |
| **Slider** | 轨道 + 滑块 | `railBg: colorBorderSecondary`, `trackBg: colorPrimary`, `handleBg: colorBgContainer`, `handleBorderWidth: 2` |

#### Tier 3 - 低优先级（当前 stories 中未展示或展示较少）

| 组件 | 覆盖策略 | 备注 |
|---|---|---|
| **Checkbox** | 选中框颜色 | 确认 antd 默认是否已跟随 colorPrimary，如未跟随则覆盖 |
| **Radio** | 选中状态 | 同上 |
| **Rate** | 星星颜色 | `starColor: colorPrimary`, `starBg: colorBorder` |
| **Transfer** | 面板列表 | `listWidth: 200`（不碰尺寸，仅确认 token 跟随） |
| **Alert** | 当前已覆盖 root borderRadius | 确认各 type 的 `colorErrorBg` / `colorWarningBg` 等是否协调 |
| **List** | 边框/阴影 | `itemPadding: 16`（不碰），只确认 `borderColor` 跟随 token |
| **Tag** | 当前已覆盖 | 确认 dark 模式下各 status color 是否可读 |

### 4.3 Compact 模式差异化

Compact 模式下**不调整任何尺寸**，仅做以下视觉微调：

| 组件 | Compact 差异化 |
|---|---|
| **Button** | 保持现有，`fontWeight` 不变，不调整 padding |
| **Input/Select** | 保持现有，不调整 height |
| **Table** | `headerBg` 使用略深于默认行的颜色，增加表头与数据行的区分度（因为 compact 行高更小，容易糊在一起） |
| **Card** | 保持 `shadowSmall`，不调整 |
| **Tag** | 保持现有，不调整 |
| **Progress** | 已存在差异化（rail height: compact ? 8 : 12），保留 |

## 5. 具体代码实现规划

### 5.1 Token 调整

```ts
// darkTokens 中修改：
colorBorder: "#52525B",         // ↑ 从 #3F3F46
colorBorderSecondary: "#3F3F46", // 原 colorBorder 降级
```

### 5.2 新增/修改的 createStyles 项

```ts
const useStyles = createStyles(({ cssVar }, isDark: boolean) => {
  // ... 现有样式保留 ...

  // 新增：
  const textBtnRoot = {
    borderRadius: cssVar.borderRadiusSM,
    transition: `background ${cssVar.motionDurationFast}`,
    "&:hover": {
      background: isDark
        ? "rgba(255,255,255,0.06)"
        : "rgba(0,0,0,0.04)",
    },
  }

  const linkBtnRoot = {
    borderRadius: cssVar.borderRadiusSM,
    transition: `background ${cssVar.motionDurationFast}`,
    "&:hover": {
      background: isDark
        ? "rgba(255,255,255,0.06)"
        : "rgba(0,0,0,0.04)",
    },
  }

  const tableRoot = {
    // 通过 ConfigProvider components.Table 传入，非 classNames
  }

  const tabsCardRoot = {
    // 通过 ConfigProvider components.Tabs 传入
  }

  const menuRoot = {
    // 通过 ConfigProvider components.Menu 传入
  }

  // ... 其他新增样式

  return {
    // ... 现有返回 ...
    textBtnRoot,
    linkBtnRoot,
  }
})
```

### 5.3 ConfigProvider components 扩展

```ts
const config = useMemo<Partial<ConfigProviderProps>>(
  () => ({
    theme: {
      // ... 现有 theme ...
      components: {
        // ... 现有组件覆盖 ...
        // 新增：
        Table: {
          headerBg: isDark ? "#27272A" : "#FAFAFA",
          headerColor: isDark ? darkTokens.colorTextBase : lightTokens.colorTextBase,
          rowHoverBg: `color-mix(in srgb, ${isDark ? darkTokens.colorPrimary : lightTokens.colorPrimary} 4%, transparent)`,
          borderColor: isDark ? darkTokens.colorBorderSecondary : lightTokens.colorBorderSecondary,
          headerBorderRadius: shared.borderRadiusLG,
        },
        Tabs: {
          cardBg: isDark ? darkTokens.colorBgContainer : lightTokens.colorBgContainer,
          cardGutter: 4,
          itemColor: isDark ? darkTokens.colorTextBase : lightTokens.colorTextBase,
        },
        Timeline: {
          dotBg: isDark ? darkTokens.colorBgContainer : lightTokens.colorBgContainer,
          dotBorderWidth: 2,
          itemPaddingBottom: 24,
        },
        Menu: {
          itemSelectedBg: isDark ? darkTokens.colorPrimary : lightTokens.colorPrimary,
          itemSelectedColor: "#FFFFFF",
          itemHoverBg: `color-mix(in srgb, ${isDark ? darkTokens.colorPrimary : lightTokens.colorPrimary} 8%, transparent)`,
          darkItemSelectedBg: isDark ? darkTokens.colorPrimary : lightTokens.colorPrimary,
          darkItemHoverBg: `color-mix(in srgb, ${isDark ? darkTokens.colorPrimary : lightTokens.colorPrimary} 12%, transparent)`,
        },
        Steps: {
          iconBg: isDark ? darkTokens.colorBgContainer : lightTokens.colorBgContainer,
          iconBorderColor: isDark ? darkTokens.colorBorder : lightTokens.colorBorder,
          finishIconBorderColor: isDark ? darkTokens.colorPrimary : lightTokens.colorPrimary,
          finishIconBg: isDark ? darkTokens.colorBgContainer : lightTokens.colorBgContainer,
        },
        Pagination: {
          itemBg: isDark ? darkTokens.colorBgContainer : lightTokens.colorBgContainer,
          itemActiveBg: isDark ? darkTokens.colorPrimary : lightTokens.colorPrimary,
          itemActiveColor: "#FFFFFF",
          itemBorderRadius: shared.borderRadius,
        },
        Tree: {
          nodeSelectedBg: `color-mix(in srgb, ${isDark ? darkTokens.colorPrimary : lightTokens.colorPrimary} 8%, transparent)`,
          nodeHoverBg: `color-mix(in srgb, ${isDark ? darkTokens.colorPrimary : lightTokens.colorPrimary} 4%, transparent)`,
        },
        Upload: {
          draggerBg: isDark ? darkTokens.colorBgContainer : lightTokens.colorBgContainer,
          draggerBorder: `1px dashed ${isDark ? darkTokens.colorBorder : lightTokens.colorBorder}`,
          draggerBorderRadius: shared.borderRadiusLG,
        },
        Collapse: {
          headerBg: isDark ? darkTokens.colorBgContainer : lightTokens.colorBgContainer,
          contentBg: isDark ? darkTokens.colorBgContainer : lightTokens.colorBgContainer,
        },
        Descriptions: {
          labelColor: isDark ? darkTokens.colorTextBase : lightTokens.colorTextBase,
          contentColor: isDark ? darkTokens.colorTextBase : lightTokens.colorTextBase,
          borderColor: isDark ? darkTokens.colorBorderSecondary : lightTokens.colorBorderSecondary,
        },
        Slider: {
          railBg: isDark ? darkTokens.colorBorderSecondary : lightTokens.colorBorderSecondary,
          trackBg: isDark ? darkTokens.colorPrimary : lightTokens.colorPrimary,
          handleBg: isDark ? darkTokens.colorBgContainer : lightTokens.colorBgContainer,
        },
        Input: {
          // ... 现有 ...
          colorTextPlaceholder: isDark ? "#9CA3AF" : undefined,
        },
      },
    },
    // ... 现有组件覆盖 ...
    button: {
      classNames: ({ props }) => {
        // ... 现有逻辑 ...
        // 新增 Text/Link 处理：
        if (props.type === "text") return { root: styles.textBtnRoot }
        if (props.type === "link") return { root: styles.linkBtnRoot }
        // ... 现有逻辑继续 ...
      },
    },
    card: {
      classNames: { root: styles.heroPanel },  // 从 styles 改为 classNames 以统一
      // 或保持现有，但确保使用 shadowSmall
    },
    // ... 其他新增覆盖 ...
  }),
  [styles, base, isDark, compact],
)
```

> **注意**：card 当前使用 `styles: { root: heroPanel }` 是 inline style，无法响应 CSS variable 变化。应改为 `classNames` 或确保 `heroPanel` 使用 `cssVar`。当前 `heroPanel` 使用的是 `cssVar` 方式，所以其实没问题。但需要确认所有面板都使用统一的 `heroPanel` / `heroPanelMedium` / `heroPanelLarge`。

### 5.4 现有 heroPanel 系列确认

当前定义：
```ts
const heroPanel = {
  background: cssVar.colorBgContainer,
  border: isDark ? `1px solid ${cssVar.colorBorder}` : "none",
  borderRadius: cssVar.borderRadiusLG,
  boxShadow: shadowSmall,
}
```

这套定义是正确的，但当前 Card 使用的是 inline `styles: { root: heroPanel }`，而其他组件（modal、dropdown 等）使用的是 `classNames`。两种方式效果一样，但为了统一，建议全部使用 `classNames` 方式。

## 6. 验证清单

优化完成后，需在 storybook 中按以下清单验证：

### 6.1 Light 模式

- [ ] AllComponents Gallery：所有按钮类型圆角一致，Card 有阴影
- [ ] DataTablePage：表头有明确背景区分，行 hover 有 subtle 高亮
- [ ] FormDetailPage：Steps 图标有边框，Tabs 切换自然
- [ ] DrawerDetailPage：Timeline 节点有边框，Tabs 有 card 样式
- [ ] Form Gallery：Validation Error/Warning/Success 状态边框颜色协调
- [ ] Steps Gallery：各类型（default/dot/inline/navigation）视觉一致
- [ ] Menu Gallery：选中项高亮明显，hover 有反馈
- [ ] Tree Gallery：选中节点有 subtle 背景高亮
- [ ] Upload Gallery：Dragger 区域有虚线边框，hover 变 primary 色

### 6.2 Dark 模式

- [ ] AllComponents Gallery：Text/Link 按钮 hover 时有背景反馈
- [ ] AllComponents Gallery：输入框 placeholder 清晰可读
- [ ] DataTablePage：表头与数据行有明确区分
- [ ] Menu Gallery（Dark Sidebar）：选中项蓝色背景 + 白色文字，对比度充分
- [ ] Modal Gallery：弹窗阴影有层次感，暗色下不突兀

### 6.3 Compact 模式

- [ ] DataTablePage：表头背景更深，防止 compact 行高缩小后糊成一片
- [ ] AllComponents Gallery：按钮、输入框尺寸未变（验证禁区）

## 7. 实施顺序

按以下顺序逐步实现，每步完成后在 storybook 中验证：

1. **Token 修复**：调整 darkTokens `colorBorder` / `colorBorderSecondary`
2. **Button 修复**：Text/Link 按钮增加 hover 背景 + Link 圆角
3. **核心组件覆盖**：Table、Tabs、Timeline、Menu、Steps、Pagination
4. **次级组件覆盖**：Tree、Upload、Collapse、Descriptions、Slider
5. **微调与验证**：跑完验证清单，修复边角问题

---

*文档版本：v1.0*
*日期：2026-05-19*
*范围：Hero Skin 专属优化*
