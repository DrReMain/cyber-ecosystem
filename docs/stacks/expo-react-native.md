# Expo + React Native Stack Guide

## Overview

This guide covers the mobile client stack for the cyber-ecosystem monorepo, used for C-end (consumer-facing) applications built with Expo and React Native.

---

## Core Stack

| Category | Technology | Version |
|----------|-----------|---------|
| Framework | Expo SDK | 53+ |
| Runtime | React Native | (managed by Expo) |
| Language | TypeScript | strict mode |
| Package Manager | Bun | (monorepo-wide) |

---

## Routing

**expo-router** — File-based routing, same paradigm as TanStack Router used in web clients.

- Convention: `app/` directory with file-based routes
- Supports layouts, tabs, stack navigation via file conventions
- Deep linking support built-in

---

## Styling

**NativeWind v5** — Tailwind CSS v4 for React Native.

- Syntax identical to web Tailwind: `className="flex p-4 bg-white dark:bg-zinc-900"`
- Responsive breakpoints align with web client (see breakpoint convention below)
- CSS-first configuration (Tailwind v4 style)
- Does NOT handle animations — use reanimated for that

### Breakpoint Convention

Align with the web client's Tailwind breakpoints:

| Name | Value | Description |
|------|-------|-------------|
| base | 0 | Mobile (no prefix) |
| sm | 640 | Large phone / small tablet |
| md | 768 | Tablet portrait |
| lg | 1024 | Tablet landscape |
| xl | 1280 | Desktop |
| 2xl | 1536 | Large desktop |

Mobile-first: default styles target mobile, use breakpoints for larger screens.

---

## Animations

**react-native-reanimated** — UI-thread animations at 60fps.

- Declarative `useAnimatedStyle` and `withSpring`/`withTiming` APIs
- Pair with `react-native-gesture-handler` for gesture-driven interactions
- This is the standard animation layer; do NOT rely on CSS/Tailwind animations

---

## UI Components

No full UI component library. Build custom components with NativeWind.

- **Simple components** (Button, Card, Badge, etc.) — write with NativeWind directly
- **Complex components** (DatePicker, Select, BottomSheet) — import primitives from **Gluestack UI v3** on demand
- **Icons** — `lucide-react-native` (consistent with web client)

### Why No UI Library

C-end apps require high visual freedom and custom designs. Heavy UI libraries impose design constraints that don't serve consumer-facing products. The component composition approach maximizes flexibility while keeping bundle size minimal.

---

## State Management

**Jotai** — Atomic state management, same as web client.

- Primitive atoms for simple state
- `atomWithStorage` for persisted state (via AsyncStorage or MMKV)
- Consistent pattern with web client's Jotai usage

---

## Data Fetching

**TanStack Query** — Server state management.

- Same library as web client, React Native compatible
- Use `@tanstack/react-query` directly
- Query keys follow the same convention as web client

---

## API Protocol

**Connect RPC** — Type-safe API client generated from Protobuf definitions.

- `@connectrpc/connect-web` works with React Native's native `fetch`
- No custom transport adapter needed (unlike mini-programs)
- Proto definitions are the SSOT — generated clients are derived output
- For non-unary/non-JSON endpoints, use `fetch` directly

### Client Generation

Same proto files as the backend, generate RN-compatible clients via the existing `template_api` Nx target.

---

## Internationalization

**expo-localization + i18next**

- `expo-localization` — reads device locale automatically
- `i18next` + `react-i18next` — runtime translation
- Different from web client's Paraglide (which is Vite-dependent and supports language routing)
- No language routing needed — app uses device locale

### Translation Files

```
src/i18n/
  locales/
    en.json
    zh.json
    ...
  index.ts          # i18next config
```

---

## Error Reporting

**Sentry SDK → GlitchTip**

- `@sentry/react-native` sends to self-hosted GlitchTip
- DSN configured per environment via environment variables
- Separate GlitchTip project from web client for issue isolation

---

## Observability

**OpenTelemetry**

- OTLP HTTP exporter to SigNoz (same endpoint as web: `localhost:4318` in dev)
- `@opentelemetry/sdk-trace-web` compatible with RN
- Service name convention: `template_client_<name>`

---

## Monorepo Integration

### Nx Project Structure

```
apps/template/clients/
  admin/          # Web client (TanStack)
  mobile/         # Expo client
```

### Nx Targets (to be configured)

| Target | Command |
|--------|---------|
| `dev` | `npx expo start` |
| `build:ios` | `eas build --platform ios` |
| `build:android` | `eas build --platform android` |
| `lint` | `eslint` |
| `typecheck` | `tsc --noEmit` |

### Shared Code with Web Client

Code that can be shared (place in `shared-ts/`):
- Connect RPC generated clients
- API type definitions
- Jotai atom patterns (reference only, not direct import)
- Utility functions

Code that cannot be shared:
- UI components (different rendering)
- Styling (NativeWind vs Tailwind CSS)
- Routing (expo-router vs TanStack Router)
- i18n (different systems)
