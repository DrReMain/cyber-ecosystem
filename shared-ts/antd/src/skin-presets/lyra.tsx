import {
  theme as antdTheme,
  ConfigProvider,
  type ConfigProviderProps,
  type GlobalToken,
} from "antd"
import { createStyles } from "antd-style"
import type { CSSProperties, ReactNode } from "react"
import { useMemo } from "react"
import type { SkinPlugin } from "../skins/types"
import { type ButtonVariantKind, pickButtonVariant } from "../skins/utils"

const shared: Partial<GlobalToken> = {
  borderRadius: 0,
  borderRadiusSM: 0,
  borderRadiusLG: 0,
  borderRadiusXS: 0,
  fontWeightStrong: 600,
  motionDurationFast: "0.15s",
  motionDurationMid: "0.25s",
  motionDurationSlow: "0.4s",
  fontFamily:
    "'JetBrains Mono', 'SF Mono', SFMono-Regular, ui-monospace, Menlo, 'Cascadia Code', monospace",
  fontFamilyCode:
    "'JetBrains Mono', 'SF Mono', SFMono-Regular, ui-monospace, Menlo, 'Cascadia Code', monospace",
}

const lightTokens: Partial<GlobalToken> = {
  ...shared,
  colorPrimary: "#71717A",
  colorInfo: "#71717A",
  colorSuccess: "#65a30d",
  colorWarning: "#f59e0b",
  colorError: "#f87171",
  colorBgBase: "#FAFAFA",
  colorBgContainer: "#FFFFFF",
  colorBgElevated: "#FFFFFF",
  colorBgLayout: "#FAFAFA",
  colorTextBase: "#18181B",
  colorBorder: "#E4E4E7",
  colorBorderSecondary: "#F4F4F5",
}

const darkTokens: Partial<GlobalToken> = {
  ...shared,
  colorPrimary: "#A1A1AA",
  colorInfo: "#A1A1AA",
  colorSuccess: "#84cc16",
  colorWarning: "#fbbf24",
  colorError: "#fca5a5",
  colorBgBase: "#09090B",
  colorBgContainer: "#18181B",
  colorBgElevated: "#27272A",
  colorBgLayout: "#09090B",
  colorTextBase: "#E4E4E7",
  colorBorder: "#3F3F46",
  colorBorderSecondary: "#27272A",
}

const lightBodyBg = "#FAFAFA"
const darkBodyBg = "#09090B"

/* ------------------------------------------------------------------ */
/*  Static styles (module-scope CSSProperties — no cssVar needed)     */
/* ------------------------------------------------------------------ */

const lyraPanel: CSSProperties = {
  background: "var(--ant-color-bg-container)",
  border: "1px solid var(--ant-color-border)",
  borderRadius: 0,
  boxShadow: "none",
}

const lyraElevatedPanel: CSSProperties = {
  background: "var(--ant-color-bg-elevated)",
  border: "1px solid var(--ant-color-border)",
  borderRadius: 0,
  boxShadow: "none",
}

const lyraAlert: CSSProperties = {
  borderRadius: 0,
}

const lyraDrawer: CSSProperties = {
  background: "var(--ant-color-bg-elevated)",
  border: "1px solid var(--ant-color-border)",
  boxShadow: "none",
  borderRadius: 0,
}

const tagRoot: CSSProperties = {
  borderRadius: 0,
  fontWeight: 500,
  margin: 0,
}

const progressRail: CSSProperties = {
  borderRadius: 0,
  border: "1px solid var(--ant-color-border)",
}

const progressTrack: CSSProperties = {
  borderRadius: 0,
}

/* ------------------------------------------------------------------ */
/*  Interactive styles (createStyles — cssVar / pseudo-classes)       */
/* ------------------------------------------------------------------ */

const useStyles = createStyles(({ cssVar }, _isDark: boolean) => {
  return {
    lyraFlat: {
      "&:not(.ant-input-compact-item):not(.ant-btn-compact-item)": {
        borderRadius: 0,
      },
      '&:not(.ant-input-group-wrapper)[class*="-outlined"]': {
        background: cssVar.colorBgContainer,
      },
    },
    buttonSolidPrimaryRoot: {
      border: "none",
      transition: `opacity ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
      "&:hover": {
        opacity: 0.88,
      },
      "&:active": {
        transform: "scale(0.97)",
      },
    },
    buttonSolidColoredRoot: {
      border: "none",
      transition: `opacity ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
      "&:hover": {
        opacity: 0.88,
      },
      "&:active": {
        transform: "scale(0.97)",
      },
    },
    buttonDefaultRoot: {
      background: "transparent",
      border: `1px solid ${cssVar.colorBorder}`,
      color: cssVar.colorTextBase,
      transition: `border-color ${cssVar.motionDurationFast}, color ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
      "&:hover": {
        borderColor: cssVar.colorPrimary,
        color: cssVar.colorPrimary,
      },
      "&:active": {
        transform: "scale(0.97)",
      },
    },
    buttonDashedRoot: {
      background: "transparent",
      border: `1px dashed ${cssVar.colorBorder}`,
      color: cssVar.colorTextBase,
      transition: `border-color ${cssVar.motionDurationFast}, color ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
      "&:hover": {
        borderColor: cssVar.colorPrimary,
        color: cssVar.colorPrimary,
      },
      "&:active": {
        transform: "scale(0.97)",
      },
    },
    buttonFilledRoot: {
      background: `color-mix(in srgb, ${cssVar.colorPrimary} 10%, transparent)`,
      color: cssVar.colorPrimary,
      border: "none",
      transition: `background ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
      "&:hover": {
        background: `color-mix(in srgb, ${cssVar.colorPrimary} 16%, transparent)`,
      },
      "&:active": {
        transform: "scale(0.97)",
      },
    },
    buttonFilledColoredRoot: {
      border: "none",
      transition: `background ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
      "&:active": {
        transform: "scale(0.97)",
      },
    },
    switchRoot: {
      border: "none",
      background: cssVar.colorBorder,
      "&.ant-switch-checked": {
        background: cssVar.colorPrimary,
      },
    },
    stepsSliderOverride: {
      "& .ant-steps-item-icon": {
        borderRadius: 0,
      },
      "& .ant-slider-handle": {
        borderRadius: 0,
      },
    },
  }
})

function LyraProvider({
  isDark,
  compact,
  children,
}: {
  isDark: boolean
  compact: boolean
  children: ReactNode
}) {
  const { styles } = useStyles(isDark)
  const base = isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm

  const config = useMemo<Partial<ConfigProviderProps>>(
    () => ({
      theme: {
        algorithm: compact ? [base, antdTheme.compactAlgorithm] : base,
        token: isDark ? darkTokens : lightTokens,
        components: {
          Layout: {
            bodyBg: isDark ? darkBodyBg : lightBodyBg,
          },
          Button: {
            fontWeight: 500,
            defaultShadow: "none",
            primaryShadow: "none",
            dangerShadow: "none",
          },
          Input: {
            activeShadow: "none",
          },
          Select: {
            optionSelectedColor: isDark
              ? darkTokens.colorPrimary
              : lightTokens.colorPrimary,
            optionSelectedBg: isDark
              ? `color-mix(in srgb, ${darkTokens.colorPrimary} 15%, transparent)`
              : `color-mix(in srgb, ${lightTokens.colorPrimary} 15%, transparent)`,
            optionActiveBg: isDark
              ? `color-mix(in srgb, ${darkTokens.colorPrimary} 8%, transparent)`
              : `color-mix(in srgb, ${lightTokens.colorPrimary} 8%, transparent)`,
          },
          Switch: {
            handleBg: "#FFFFFF",
            handleShadow: "0 1px 2px 0 rgb(0 0 0 / 0.05)",
          },
          Tooltip: { borderRadius: 0 },
        },
      },
      alert: { styles: { root: lyraAlert } },
      button: {
        classNames: ({ props }) => {
          const kind = pickButtonVariant(props)
          if (kind === "skip") return {}
          const isGreyPrimary =
            props.color === "primary" || props.type === "primary"
          const isDefaultColor = !props.color || props.color === "default"
          if (
            (kind === "solid" || props.variant === "solid") &&
            !isGreyPrimary
          ) {
            return { root: styles.buttonSolidColoredRoot }
          }
          if (isDefaultColor && kind === "filled") return {}
          if (!isDefaultColor && kind === "filled") {
            return { root: styles.buttonFilledColoredRoot }
          }
          if (
            !isDefaultColor &&
            (kind === "outlined" || kind === "default" || kind === "dashed")
          ) {
            return {}
          }
          const map: Record<Exclude<ButtonVariantKind, "skip">, string> = {
            solid: styles.buttonSolidPrimaryRoot,
            dashed: styles.buttonDashedRoot,
            filled: styles.buttonFilledRoot,
            outlined: styles.buttonDefaultRoot,
            default: styles.buttonDefaultRoot,
          }
          return { root: map[kind] }
        },
      },
      card: { styles: { root: lyraPanel } },
      colorPicker: { styles: { root: lyraPanel }, arrow: false },
      datePicker: {
        classNames: { root: styles.lyraFlat },
        styles: { popup: { container: lyraElevatedPanel } },
      },
      drawer: { styles: { section: lyraDrawer } },
      dropdown: { styles: { root: lyraElevatedPanel } },
      input: { classNames: { root: styles.lyraFlat } },
      inputNumber: { classNames: { root: styles.lyraFlat } },
      textArea: { classNames: { root: styles.lyraFlat } },
      modal: { styles: { container: lyraElevatedPanel } },
      popover: { styles: { container: lyraElevatedPanel } },
      select: {
        classNames: { root: styles.lyraFlat },
        styles: { popup: { root: lyraElevatedPanel } },
      },
      timePicker: { classNames: { root: styles.lyraFlat } },
      cascader: { classNames: { root: styles.lyraFlat } },
      treeSelect: { classNames: { root: styles.lyraFlat } },
      mentions: { classNames: { root: styles.lyraFlat } },
      switch: { classNames: { root: styles.switchRoot } },
      tag: { styles: { root: tagRoot } },
      progress: {
        styles: {
          rail: progressRail,
          track: progressTrack,
        },
      },
      tooltip: { arrow: false },
    }),
    [styles, base, isDark, compact],
  )

  return (
    <ConfigProvider {...config}>
      <div className={styles.stepsSliderOverride}>{children}</div>
    </ConfigProvider>
  )
}

const lyraSkin: SkinPlugin = {
  id: "lyra",
  name: "Lyra",
  bodyBg: { light: lightBodyBg, dark: darkBodyBg },
  Provider: LyraProvider,
}

export default lyraSkin
