# TanStack Client Runbooks

**Status:** ACTIVE

Procedures for the tanstack lane. Law lives in the CONVENTIONS; each step cites the law it executes.

## 1. New skin

1. Copy the closest existing preset (square → lyra, round/soft → luma).
2. Replace the token palettes (light/dark objects), `bodyBg`, font stacks, shadow constants.
3. Keep verbatim: the Layout/Menu chrome contract, the button dispatch, the flat-input patch, the steps/slider patch, the Select option tokens, the `useMemo` config with deps `[styles, base, isDark, compact]` — docs/conventions/tanstack/CONVENTIONS.md §5 "Skin presets".
4. Register in `../skins/registry.tsx` — the only wiring step.
5. Validate per the tuning loop below.

## 2. Skin tuning loop

First landings ship small geometric flaws by nature — colors are free, metrics are not. Storybook is the regression harness; the app is final confirmation only.

1. `./nx run @cyber-ecosystem/shared-storybook:dev` — the toolbar drives Skin × Light/Dark × Compact.
2. Sweep order: Foundation/Tokens (scale sanity) → Combinations/Login Card, Form Page, Table Page (the density stress surfaces: toolbars, filter rows, pagers, popups) → the admin settings drawer (thumbnail + live switch) → one dense admin page (users table, diag).
3. All four mode combos; on locale-sensitive surfaces cover every registered locale including ar (RTL) — popups and menu items are where width changes bite.
4. Per-skin fixes stay in the skin's own file; re-sweep after every geometry token change — a radius tweak that fixes a card can break a tag.
5. Static pass each round: `./nx run @cyber-ecosystem/shared-antd:typecheck` and `:check`.
