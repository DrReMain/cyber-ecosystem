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
  borderRadius: 12,
  borderRadiusSM: 8,
  borderRadiusLG: 16,
  borderRadiusXS: 4,
  fontWeightStrong: 600,
  motionDurationFast: "0.15s",
  motionDurationMid: "0.25s",
  motionDurationSlow: "0.4s",
  fontFamily:
    "'Inter', -apple-system, BlinkMacSystemFont, 'SF Pro Text', 'Helvetica Neue', 'PingFang SC', 'Microsoft YaHei', sans-serif",
  fontFamilyCode:
    "'JetBrains Mono', 'SF Mono', SFMono-Regular, ui-monospace, Menlo, monospace",
}

const lightTokens: Partial<GlobalToken> = {
  ...shared,
  colorPrimary: "#006FEE",
  colorInfo: "#006FEE",
  colorSuccess: "#17C964",
  colorWarning: "#F5A524",
  colorError: "#F31260",
  colorBgBase: "#FFFFFF",
  colorBgContainer: "#FFFFFF",
  colorBgElevated: "#FAFAFA",
  colorBgLayout: "#F4F4F5",
  colorTextBase: "#11181C",
  colorBorder: "#E4E4E7",
  colorBorderSecondary: "#E4E4E7",
}

const darkTokens: Partial<GlobalToken> = {
  ...shared,
  colorPrimary: "#338EF7",
  colorInfo: "#338EF7",
  colorSuccess: "#45D483",
  colorWarning: "#F7B750",
  colorError: "#F54180",
  colorBgBase: "#111111",
  colorBgContainer: "#18181B",
  colorBgElevated: "#27272A",
  colorBgLayout: "#09090B",
  colorTextBase: "#ECEDEE",
  colorBorder: "#3F3F46",
  colorBorderSecondary: "#27272A",
}

const lightBodyBg = "#F4F4F5"
const darkBodyBg = "#09090B"

const heroAlertStyle: CSSProperties = {
  borderRadius: "var(--ant-border-radius)",
}

const tagRootStyle: CSSProperties = {
  borderRadius: "var(--ant-border-radius)",
  fontWeight: 500,
  margin: 0,
}

const progressTrackStyle: CSSProperties = {
  borderRadius: "var(--ant-border-radius-sm)",
}

const useStyles = createStyles(({ cssVar }, isDark: boolean) => {
  const shadowSmall = isDark
    ? "0px 0px 5px 0px rgb(0 0 0 / 0.05), 0px 2px 10px 0px rgb(0 0 0 / 0.2), inset 0px 0px 1px 0px rgb(255 255 255 / 0.15)"
    : "0px 0px 5px 0px rgb(0 0 0 / 0.02), 0px 2px 10px 0px rgb(0 0 0 / 0.06), 0px 0px 1px 0px rgb(0 0 0 / 0.3)"
  const shadowMedium = isDark
    ? "0px 0px 15px 0px rgb(0 0 0 / 0.06), 0px 2px 30px 0px rgb(0 0 0 / 0.22), inset 0px 0px 1px 0px rgb(255 255 255 / 0.15)"
    : "0px 0px 15px 0px rgb(0 0 0 / 0.03), 0px 2px 30px 0px rgb(0 0 0 / 0.08), 0px 0px 1px 0px rgb(0 0 0 / 0.3)"
  const shadowLarge = isDark
    ? "0px 0px 30px 0px rgb(0 0 0 / 0.07), 0px 30px 60px 0px rgb(0 0 0 / 0.26), inset 0px 0px 1px 0px rgb(255 255 255 / 0.15)"
    : "0px 0px 30px 0px rgb(0 0 0 / 0.04), 0px 30px 60px 0px rgb(0 0 0 / 0.12), 0px 0px 1px 0px rgb(0 0 0 / 0.3)"

  const heroFlat = {
    "&:not(.ant-input-compact-item):not(.ant-btn-compact-item)": {
      borderRadius: cssVar.borderRadius,
    },
    '&:not(.ant-input-group-wrapper)[class*="-outlined"]': {
      background: cssVar.colorBgContainer,
    },
  }

  const heroPanel = {
    background: cssVar.colorBgContainer,
    border: isDark ? `1px solid ${cssVar.colorBorder}` : "none",
    borderRadius: cssVar.borderRadiusLG,
    boxShadow: shadowSmall,
  }

  const heroPanelMedium = {
    background: cssVar.colorBgContainer,
    border: isDark ? `1px solid ${cssVar.colorBorder}` : "none",
    borderRadius: cssVar.borderRadiusLG,
    boxShadow: shadowMedium,
  }

  const heroPanelLarge = {
    background: cssVar.colorBgContainer,
    border: isDark ? `1px solid ${cssVar.colorBorder}` : "none",
    borderRadius: cssVar.borderRadiusLG,
    boxShadow: shadowLarge,
  }

  const heroDrawer = {
    background: cssVar.colorBgContainer,
    border: isDark ? `1px solid ${cssVar.colorBorder}` : "none",
    boxShadow: shadowLarge,
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
    fontWeight: 500,
    boxShadow: `0 6px 20px color-mix(in srgb, ${cssVar.colorPrimary} 35%, transparent)`,
    transition: `opacity ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
    "&:hover": {
      opacity: 0.92,
    },
    "&:active": {
      transform: "scale(0.97)",
    },
  }

  const buttonSolidColoredRoot = {
    border: "none",
    fontWeight: 500,
    boxShadow: shadowSmall,
    transition: `opacity ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
    "&:hover": {
      opacity: 0.92,
    },
    "&:active": {
      transform: "scale(0.97)",
    },
  }

  const buttonDefaultRoot = {
    background: "transparent",
    border: `1px solid ${cssVar.colorBorder}`,
    color: cssVar.colorTextBase,
    fontWeight: 500,
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
    background: `color-mix(in srgb, ${cssVar.colorPrimary} 12%, transparent)`,
    color: cssVar.colorPrimary,
    border: "none",
    fontWeight: 500,
    transition: `background ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
    "&:hover": {
      background: `color-mix(in srgb, ${cssVar.colorPrimary} 20%, transparent)`,
    },
    "&:active": {
      transform: "scale(0.97)",
    },
  }

  const buttonFilledColoredRoot = {
    border: "none",
    fontWeight: 500,
    transition: `background ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
    "&:active": {
      transform: "scale(0.97)",
    },
  }

  const progressRail = {
    borderRadius: cssVar.borderRadiusSM,
    boxShadow: isDark ? "none" : shadowSmall,
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
    heroFlat,
    heroPanel,
    heroPanelMedium,
    heroPanelLarge,
    heroDrawer,
    buttonSolidPrimaryRoot,
    buttonSolidColoredRoot,
    buttonDefaultRoot,
    buttonDashedRoot,
    buttonFilledRoot,
    buttonFilledColoredRoot,
    progressRail,
    switchRoot,
    stepsSliderOverride,
  }
})

function HeroProvider({
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
          Layout: { bodyBg: isDark ? darkBodyBg : lightBodyBg },
          Button: {
            fontWeight: 500,
          },
          Select: {
            optionSelectedColor: isDark
              ? darkTokens.colorPrimary
              : lightTokens.colorPrimary,
            optionSelectedBg: isDark
              ? `color-mix(in srgb, ${darkTokens.colorPrimary} 20%, transparent)`
              : `color-mix(in srgb, ${lightTokens.colorPrimary} 20%, transparent)`,
            optionActiveBg: isDark
              ? `color-mix(in srgb, ${darkTokens.colorPrimary} 10%, transparent)`
              : `color-mix(in srgb, ${lightTokens.colorPrimary} 10%, transparent)`,
          },
          Switch: {
            handleBg: "#FFFFFF",
            handleShadow: "0 1px 3px rgba(0,0,0,0.15)",
          },
          Tooltip: { borderRadius: 8 },
        },
      },
      alert: { styles: { root: heroAlertStyle } },
      button: {
        classNames: ({ props }) => {
          const kind = pickButtonVariant(props)
          if (kind === "skip") return {}
          const isBluePrimary =
            props.color === "primary" || props.type === "primary"
          const isDefaultColor = !props.color || props.color === "default"
          if (
            (kind === "solid" || props.variant === "solid") &&
            !isBluePrimary
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
      card: { classNames: { root: styles.heroPanel } },
      colorPicker: {
        classNames: { root: styles.heroPanel },
        arrow: false,
      },
      datePicker: {
        classNames: {
          root: styles.heroFlat,
          popup: { container: styles.heroPanelMedium },
        },
      },
      drawer: { classNames: { section: styles.heroDrawer } },
      dropdown: { classNames: { root: styles.heroPanelMedium } },
      input: { classNames: { root: styles.heroFlat } },
      inputNumber: { classNames: { root: styles.heroFlat } },
      textArea: { classNames: { root: styles.heroFlat } },
      modal: { classNames: { container: styles.heroPanelLarge } },
      popover: { classNames: { container: styles.heroPanelMedium } },
      select: {
        classNames: {
          root: styles.heroFlat,
          popup: { root: styles.heroPanelMedium },
        },
      },
      timePicker: {
        classNames: { root: styles.heroFlat },
      },
      cascader: {
        classNames: { root: styles.heroFlat },
      },
      treeSelect: {
        classNames: { root: styles.heroFlat },
      },
      mentions: {
        classNames: { root: styles.heroFlat },
      },
      switch: { classNames: { root: styles.switchRoot } },
      tag: { styles: { root: tagRootStyle } },
      progress: {
        classNames: {
          rail: styles.progressRail,
        },
        styles: {
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

const heroSkin: SkinPlugin = {
  id: "hero",
  name: "Hero",
  bodyBg: { light: lightBodyBg, dark: darkBodyBg },
  Provider: HeroProvider,
}

export default heroSkin
