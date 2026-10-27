# Skin Conventions

**Scope:** antd skin presets under `shared-ts/antd/src/skin-presets/` — authoring, tuning, and registering skins. Business-agnostic by design: a skin is a palette-and-geometry fill-in, never a business surface. Prescriptive — `MUST` / `SHOULD` / `MAY`.

## Read map

| About to touch | Read first |
|---|---|
| authoring or modifying a skin file | §1, §4, §5 |
| tuning tokens on an existing skin | §2, §3, §5 |
| the registry or a skin consumer | §1 |

## 1. System map

- A skin is a `SkinPlugin` (`../skins/types.ts`): `{ id, name, bodyBg: { light, dark }, Provider }`. `bodyBg` feeds Storybook's canvas; the app reads surface colors from tokens, not from this field.
- One file per skin, default-exported (`skin-presets/<id>.tsx`). Registration is the only wiring step — the import + entry in `SKINS` (`../skins/registry.tsx`); every consumer (settings drawer grid, Storybook toolbar, `SkinSwitcher`) derives from `getAllSkins()`.
- Providers stay pure and cheap: props (`isDark`, `compact`, `children`) only — the settings drawer renders every thumbnail inside that skin's own Provider, so expensive providers multiply across the grid.

## 2. Three styling channels — escalate in order

1. `theme.token` (global) and `theme.components.<C>` — cheapest, cascades to every instance; prefer for colors, radius, motion, fonts.
2. Semantic style slots on ConfigProvider (`<component>: { styles: { <slot>: CSSProperties } }`, v6 semantic DOM) — for values antd hardcodes on real elements (Switch track radius, popup surfaces, panel paddings); slot names are typed — typecheck is the drift alarm when antd upgrades rename slots.
3. `createStyles` selectors (`antd-style`) — the only channel reaching pseudo-elements and internal descendants (`.ant-switch-handle::before`, `.ant-steps-item-icon`); React style objects cannot target `::before`; `cssVar.*` works only inside createStyles, static `CSSProperties` use `var(--ant-*)` strings.

## 3. Token risk classes

Colors are layout-inert; geometry and metric tokens change box sizes and therefore layouts: `borderRadius*`, `controlHeight`, `fontSize`/`lineHeight`, paddings, `fontFamily` (mono stacks run ~10% wider), motion durations. Every geometry token change demands a Storybook sweep, not a spot look.

## 4. Chrome token contract

The app chrome consumes a fixed contract each skin fills in its own palette: `Layout: { bodyBg, headerBg: <mode container>, headerHeight: 56, siderBg + lightSiderBg: <mode container> }`, `Menu: { itemHeight: 40, darkItemBg/darkSubMenuItemBg: "transparent" }`. `skin-presets/default.tsx` plus lyra/luma are the checklist; a skin missing any slot falls back to antd defaults (wrong hue or geometry) somewhere in the layout.

## 5. Anti-patterns (symptom → cause → fix)

- Default skin's thumbnail ghosts the active skin — nested ConfigProviders merge; a skin with an empty `token` renders as a tinted clone of whatever is active → the default skin passes its full resolved map (`LIGHT_TOKENS | DARK_TOKENS`), the others set every seed the visible surfaces derive from; never `theme.inherit: false` here (it drops the domain provider's breakpoint alignment tokens).
- Blue-black patch on the dark-mode sider — antd's generic dark-menu background paints over the skin → `Menu: { darkItemBg/darkSubMenuItemBg: "transparent" }`.
- Topbar 64px under one skin, 56px under another — `Layout.Header` height is the `headerHeight` token (antd default 64) → pin 56 in every skin.
- Floating sider painted with generic antd colors in one mode — `Layout.Sider` reads `lightSiderBg` under `theme="light"` but `siderBg` under `theme="dark"` → pin both.
- Switch stays round in a square skin — geometry-intrinsic shapes (Switch, Steps icons, Slider handles) ignore the global `borderRadius` → reshape via all three channels at once (root/content slots for track and clip, a createStyles selector for `.ant-switch-handle::before` — the semantic `indicator` slot is the invisible positioning wrapper).
- Blue hover/selection in Select options — option tokens don't inherit `colorPrimary` for backgrounds → pin `optionSelectedColor/optionSelectedBg/optionActiveBg` to the skin primary.
- Button overrides misfire on colored/danger buttons — variant + color + type interact; the `utils.pickButtonVariant` dispatch encodes tested per-variant decisions → keep the dispatch structure verbatim from lyra/luma, restyle only the mapped style objects; dedupe keys only on byte-identical CSS.
- Text truncates, rows wrap, scrollbars appear under one skin — a geometry token moved a box; keep `borderRadiusXS` under half the smallest chip height; the compact algorithm compounds tightness — check normal × compact; diagnose on the Storybook density surfaces, not in the app.
- Layout regressions under a skin — providers render exactly one sanctioned DOM wrapper (the Steps/Slider patch); more wrappers break `h-svh`/flex chains.

Procedure (new-skin recipe, tuning loop): `RUNBOOKS.md` beside this file.
