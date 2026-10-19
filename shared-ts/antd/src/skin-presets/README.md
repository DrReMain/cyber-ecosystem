# Skin presets — implementation guide & pitfalls

Audience: coding agents adding or modifying skins. Read fully before editing this directory.

## System map

- A skin is a `SkinPlugin` (`../skins/types.ts`): `{ id, name, bodyBg: { light, dark }, Provider }`.
  `bodyBg` feeds Storybook's canvas; the app reads surface colors from tokens, not from this field.
- One file per skin, default-exported: `skin-presets/<id>.tsx`. Registration is the ONLY wiring
  step — add the import + entry to `SKINS` in `../skins/registry.tsx`. Every consumer (admin
  settings drawer grid, Storybook toolbar, `SkinSwitcher`) derives from `getAllSkins()`; touch
  nothing else.
- Providers must stay pure and cheap: props (`isDark`, `compact`, `children`) only. The settings
  drawer renders every skin's thumbnail inside that skin's own Provider — thumbnails self-prove,
  and expensive providers would multiply across the grid.

## Three styling channels — escalate in this order

1. **Tokens** — `theme.token` (global) and `theme.components.<C>` (component tokens). Cheapest;
   cascades to every instance. Prefer for colors, radius, motion, fonts.
2. **Semantic styles slots** — top-level `<component>: { styles: { <slot>: CSSProperties } }` on
   ConfigProvider (v6 semantic DOM). For values antd hardcodes on real elements (Switch track
   `borderRadius: 100`, popup surfaces, panel paddings). Slot names are typed — `typecheck` is
   the drift alarm when antd upgrades rename slots.
3. **createStyles selectors** (`antd-style`, already a peer dep) — the only channel that reaches
   pseudo-elements and internal descendants (`.ant-switch-handle::before`, `.ant-steps-item-icon`,
   `.ant-slider-handle`). React style objects cannot target `::before`; static `CSSProperties`
   constants have no `cssVar` — use `var(--ant-*)` strings there, `cssVar.*` only inside
   `createStyles`.

## Token risk classes

Colors are layout-inert — swap them freely. Geometry and metric tokens change box sizes and
therefore layouts: `borderRadius*`, `controlHeight`, `fontSize`/`lineHeight`, paddings,
`fontFamily` (monospace stacks run ~10% wider than proportional ones), motion durations
(perceived timing only). Treat every geometry token as a high-risk change that demands a
Storyboard sweep, not just a spot look.

## Pitfalls (symptom → cause → fix)

- **The default skin's thumbnail ghosts the active skin.** Nested ConfigProviders MERGE:
  a Provider wins only the tokens it explicitly sets; everything else inherits from the
  enclosing provider — which, for a thumbnail, is the ACTIVE skin. A skin with an empty
  `token` renders as a tinted clone of whatever is active. → The default skin passes its full
  resolved map (`token: LIGHT_TOKENS | DARK_TOKENS`); lyra/luma set every seed the visible
  surfaces derive from. Do NOT use `theme.inherit: false` here — it would also drop the
  domain provider's breakpoint alignment tokens. When adding tokens to a skin, cover the
  complete visible vocabulary or the thumbnail will drift under the other skins.
- **Menu shows a blue-black patch on the dark-mode sider.** antd's generic dark-menu background
  paints over the skin's sider. → Set `Menu: { darkItemBg: "transparent", darkSubMenuItemBg:
  "transparent" }`. `SiderMenu` switches `theme="dark"` with the app mode and must let the sider
  container color show through.
- **Topbar is 64px under one skin, 56px under another.** `Layout.Header` height comes from the
  `headerHeight` component token, antd default 64. → Pin `headerHeight: 56` in every skin.
- **Floating sider painted with generic antd colors.** `Layout.Sider` reads `lightSiderBg` when
  `theme="light"` but `siderBg` when `theme="dark"` — two separate tokens. → Pin BOTH to the
  mode's container color; pinning one leaves the other mode drifting.
- **Switch stays round in a square skin.** Switch (and Steps icons, Slider handles) has
  geometry-intrinsic shapes that ignore the global `borderRadius` token. → Zero/reshape via all
  three channels at once (lyra): `styles: { root, content }` slots for the track and inner
  clip, plus a createStyles selector for `.ant-switch-handle::before` — the semantic `indicator`
  slot maps to the invisible positioning wrapper, the visible dot is the pseudo-element.
  Round skins (luma) may leave the pill as-is.
- **Blue hover/selection in Select options.** Option tokens don't inherit `colorPrimary` for
  backgrounds. → Pin `optionSelectedColor/optionSelectedBg/optionActiveBg` to the skin primary
  (hoist one `primary` const inside the config memo).
- **Button overrides misfire on colored/danger buttons.** Button variant + color + type interact;
  the dispatch in `utils.pickButtonVariant` plus the isPrimary/isDefaultColor branching encodes
  tested per-variant decisions. → Keep the dispatch structure verbatim from lyra/luma; only
  restyle the style objects it maps to. Dedupe style keys only when the CSS is byte-identical.
- **Text truncates, rows wrap, or scrollbars appear only under one skin.** Geometry tokens
  moved a box: mono/wide font stacks overflow labels that fit before, large paddings crush
  toolbars, radius changes alter effective inner space. Guardrails: keep `borderRadiusXS`
  under half the smallest chip height (bigger radii clamp and visually eat padding); expect
  the compact algorithm to compound tightness — check normal × compact on every geometry
  change. → Diagnose on the Storybook density surfaces (see Tuning loop), not in the app.
- **Layout regressions under a skin.** Providers render a DOM wrapper (`<div
  className={stepsSliderOverride}>`) for the Steps/Slider patch — sanctioned for exactly that.
  Adding more wrappers breaks `h-svh`/flex chains in the app layout.
- **Skin looks right alone, wrong next to chrome.** The app chrome consumes a fixed token
  contract every skin must fill in its own palette:
  `Layout: { bodyBg, headerBg: <mode container>, headerHeight: 56, siderBg + lightSiderBg:
  <mode container> }`, `Menu: { itemHeight: 40, darkItemBg/darkSubMenuItemBg: "transparent" }`.
  Treat `skin-presets/default.tsx` + lyra/luma as the checklist; a skin missing any slot falls
  back to antd defaults (wrong hue or wrong geometry) somewhere in the layout.

## Recipe (new skin)

1. Copy the closest existing preset (square → lyra, round/soft → luma).
2. Replace the token palettes (light/dark objects), `bodyBg`, font stacks, shadow constants.
3. Keep: Layout/Menu chrome contract, button dispatch, flat-input patch, steps/slider patch,
   Select option tokens, `useMemo` config with deps `[styles, base, isDark, compact]`.
4. Register in `../skins/registry.tsx`.
5. Validate: see Tuning loop.

## Tuning loop

Skins are continuously iterated — first landings ship small geometric flaws by nature
(colors are free, metrics are not). Storybook exists to make that iteration cheap; use it
as the regression harness, the app only as final confirmation.

1. `./nx run @cyber-ecosystem/shared-storybook:dev` — toolbar drives Skin × Light/Dark ×
   Compact.
2. Sweep order per change: Foundation/Tokens (scale sanity) → Combinations/Login Card,
   Form Page, Table Page (the density stress surfaces: toolbars, filter rows, pagers,
   popups) → admin settings drawer (thumbnail + live switch) → one dense admin page
   (users table, diag).
3. All four mode combos; with locale-sensitive surfaces cover the five locales including
   ar (RTL) — popups and menu items are where width changes bite.
4. Per-skin fixes stay in the skin's own file; re-sweep after every geometry token change
   — a radius tweak that fixes a card can break a tag.
5. Static pass each round: `./nx run @cyber-ecosystem/shared-antd:typecheck` + `:check`
   (scoped `-- --write --unsafe <files>` on new or class-shuffled files).
