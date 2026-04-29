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
  borderRadius: 24,
  borderRadiusSM: 16,
  borderRadiusLG: 32,
  borderRadiusXS: 12,
  fontWeightStrong: 600,
  motionDurationFast: "0.15s",
  motionDurationMid: "0.25s",
  motionDurationSlow: "0.4s",
  fontFamily:
    "'Inter', 'Inter var', -apple-system, BlinkMacSystemFont, 'SF Pro Text', 'Helvetica Neue', 'PingFang SC', 'Microsoft YaHei', sans-serif",
  fontFamilyCode:
    "'JetBrains Mono', 'SF Mono', SFMono-Regular, ui-monospace, Menlo, 'Cascadia Code', monospace",
}

const lightTokens: Partial<GlobalToken> = {
  ...shared,
  colorPrimary: "#84cc16",
  colorInfo: "#84cc16",
  colorSuccess: "#65a30d",
  colorWarning: "#f59e0b",
  colorError: "#f87171",
  colorBgBase: "#FAFAF7",
  colorBgContainer: "#FFFFFF",
  colorBgElevated: "#FFFFFF",
  colorBgLayout: "#FAFAF7",
  colorTextBase: "#1C1C1A",
  colorBorder: "#E6E6E2",
  colorBorderSecondary: "#F0F0EC",
}

const darkTokens: Partial<GlobalToken> = {
  ...shared,
  colorPrimary: "#a3e635",
  colorInfo: "#a3e635",
  colorSuccess: "#84cc16",
  colorWarning: "#fbbf24",
  colorError: "#fca5a5",
  colorBgBase: "#0D0D0C",
  colorBgContainer: "#1A1A18",
  colorBgElevated: "#222220",
  colorBgLayout: "#0D0D0C",
  colorTextBase: "#F2F2EE",
  colorBorder: "#2A2A28",
  colorBorderSecondary: "#1E1E1C",
}

const lightBodyBg = "#FAFAF7"
const darkBodyBg = "#0D0D0C"

const shadowSmall =
  "0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)"
const shadowMedium =
  "0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)"
const shadowLarge =
  "0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)"
const ringBorder =
  "0 0 0 1px color-mix(in srgb, var(--ant-color-text-base) 5%, transparent)"

const panelRoot: CSSProperties = {
  background: "var(--ant-color-bg-container)",
  border: "none",
  borderRadius: "32px",
  boxShadow: `${ringBorder}, ${shadowSmall}`,
}

const panelRootMedium: CSSProperties = {
  background: "var(--ant-color-bg-elevated)",
  border: "none",
  borderRadius: "32px",
  boxShadow: `${ringBorder}, ${shadowMedium}`,
}

const panelRootLarge: CSSProperties = {
  background: "var(--ant-color-bg-elevated)",
  border: "none",
  borderRadius: "32px",
  boxShadow: `${ringBorder}, ${shadowLarge}`,
}

const alertRoot: CSSProperties = {
  borderRadius: "var(--ant-border-radius-sm)",
}

const tagRootStyle: CSSProperties = {
  borderRadius: "var(--ant-border-radius-xs)",
  fontWeight: 500,
  margin: 0,
}

const progressRailStyle: CSSProperties = {
  borderRadius: "var(--ant-border-radius-sm)",
}

const progressTrackStyle: CSSProperties = {
  borderRadius: "var(--ant-border-radius-sm)",
}

const useStyles = createStyles(({ cssVar }) => {
  const lumaFlat = {
    "&:not(.ant-input-compact-item):not(.ant-btn-compact-item)": {
      borderRadius: cssVar.borderRadius,
    },
    '&:not(.ant-input-group-wrapper)[class*="-outlined"]': {
      background: cssVar.colorBgContainer,
    },
  }

  const lumaDrawer = {
    background: cssVar.colorBgElevated,
    border: "none",
    boxShadow: `${ringBorder}, ${shadowLarge}`,
    ".ant-drawer-left &": {
      borderStartEndRadius: cssVar.borderRadiusLG,
      borderEndEndRadius: cssVar.borderRadiusLG,
    },
    ".ant-drawer-right &": {
      borderStartStartRadius: cssVar.borderRadiusLG,
      borderEndStartRadius: cssVar.borderRadiusLG,
    },
    ".ant-drawer-top &": {
      borderEndStartRadius: cssVar.borderRadiusLG,
      borderEndEndRadius: cssVar.borderRadiusLG,
    },
    ".ant-drawer-bottom &": {
      borderStartStartRadius: cssVar.borderRadiusLG,
      borderStartEndRadius: cssVar.borderRadiusLG,
    },
  }

  const buttonSolidPrimaryRoot = {
    border: "none",
    transition: `opacity ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
    boxShadow: shadowSmall,
    "&:hover": {
      opacity: 0.88,
    },
    "&:active": {
      transform: "scale(0.97)",
    },
  }

  const buttonSolidColoredRoot = buttonSolidPrimaryRoot

  const buttonDefaultRoot = {
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
  }

  const buttonDashedRoot = {
    ...buttonDefaultRoot,
    border: `1px dashed ${cssVar.colorBorder}`,
  }

  const buttonFilledRoot = {
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
  }

  const buttonFilledColoredRoot = {
    border: "none",
    transition: `background ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
    "&:active": {
      transform: "scale(0.97)",
    },
  }

  const switchRoot = {
    border: "none",
    background: cssVar.colorBorder,
    "&.ant-switch-checked": {
      background: cssVar.colorPrimary,
    },
  }

  const stepsSliderOverride = {
    "& .ant-steps-item-icon": {
      borderRadius: cssVar.borderRadius,
    },
    "& .ant-slider-handle": {
      borderRadius: cssVar.borderRadiusSM,
    },
  }

  return {
    lumaFlat,
    lumaDrawer,
    buttonSolidPrimaryRoot,
    buttonSolidColoredRoot,
    buttonDefaultRoot,
    buttonDashedRoot,
    buttonFilledRoot,
    buttonFilledColoredRoot,
    switchRoot,
    stepsSliderOverride,
  }
})

function LumaProvider({
  isDark,
  compact,
  children,
}: {
  isDark: boolean
  compact: boolean
  children: ReactNode
}) {
  const { styles } = useStyles()
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
          Tooltip: { borderRadius: 16 },
        },
      },
      alert: { styles: { root: alertRoot } },
      button: {
        classNames: ({ props }) => {
          const kind = pickButtonVariant(props)
          if (kind === "skip") return {}
          const isLimePrimary =
            props.color === "primary" || props.type === "primary"
          const isDefaultColor = !props.color || props.color === "default"
          if (
            (kind === "solid" || props.variant === "solid") &&
            !isLimePrimary
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
      card: {
        styles: {
          root: panelRoot,
          header: { padding: "8px 16px" },
          body: { padding: 16 },
        },
      },
      colorPicker: { styles: { root: panelRoot }, arrow: false },
      datePicker: {
        classNames: { root: styles.lumaFlat },
        styles: { popup: { container: panelRootMedium } },
      },
      drawer: {
        classNames: { section: styles.lumaDrawer },
        styles: { body: { padding: 16 } },
      },
      dropdown: { styles: { root: panelRootMedium } },
      input: { classNames: { root: styles.lumaFlat } },
      inputNumber: { classNames: { root: styles.lumaFlat } },
      textArea: { classNames: { root: styles.lumaFlat } },
      modal: {
        styles: {
          container: panelRootLarge,
          body: { padding: 16 },
        },
      },
      popover: {
        styles: {
          container: { ...panelRootMedium, padding: 16 },
        },
      },
      select: {
        classNames: {
          root: styles.lumaFlat,
        },
        styles: { popup: { root: panelRootMedium } },
      },
      timePicker: {
        classNames: { root: styles.lumaFlat },
      },
      cascader: {
        classNames: { root: styles.lumaFlat },
      },
      treeSelect: {
        classNames: { root: styles.lumaFlat },
      },
      mentions: {
        classNames: { root: styles.lumaFlat },
      },
      switch: { classNames: { root: styles.switchRoot } },
      tag: {
        styles: {
          root: { ...tagRootStyle, padding: "4px 12px" },
        },
      },
      progress: {
        styles: {
          rail: progressRailStyle,
          track: progressTrackStyle,
        },
      },
      tooltip: {
        arrow: false,
      },
    }),
    [styles, base, isDark, compact],
  )

  return (
    <ConfigProvider {...config}>
      <div className={styles.stepsSliderOverride}>{children}</div>
    </ConfigProvider>
  )
}

const lumaSkin: SkinPlugin = {
  id: "luma",
  name: "Luma",
  bodyBg: { light: lightBodyBg, dark: darkBodyBg },
  Provider: LumaProvider,
}

export default lumaSkin
