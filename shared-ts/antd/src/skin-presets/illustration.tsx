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
  lineWidth: 2,
  lineWidthBold: 2,
  borderRadius: 0,
  borderRadiusSM: 0,
  borderRadiusLG: 0,
  borderRadiusXS: 0,
  fontWeightStrong: 600,
  motionDurationFast: "0.1s",
  motionDurationMid: "0.15s",
  motionDurationSlow: "0.25s",
  fontFamily:
    "'Space Grotesk', 'Inter', -apple-system, BlinkMacSystemFont, 'SF Pro Text', 'Helvetica Neue', 'PingFang SC', 'Microsoft YaHei', sans-serif",
}

const lightTokens: Partial<GlobalToken> = {
  ...shared,
  colorPrimary: "#7C5CFC",
  colorInfo: "#7C5CFC",
  colorSuccess: "#5EEAD4",
  colorWarning: "#F0A500",
  colorError: "#FB7185",
  colorBgBase: "#FFFAEB",
  colorBgContainer: "#FFFFFF",
  colorBgElevated: "#FFFFFF",
  colorTextBase: "#001858",
  colorBorder: "#001858",
  colorBorderSecondary: "#001858",
}

const darkTokens: Partial<GlobalToken> = {
  ...shared,
  colorPrimary: "#9E97FE",
  colorInfo: "#9E97FE",
  colorSuccess: "#80F0DC",
  colorWarning: "#FFC040",
  colorError: "#FD9DA8",
  colorBgBase: "#1A1825",
  colorBgContainer: "#252238",
  colorBgElevated: "#2E2B42",
  colorTextBase: "#FFFAEB",
  colorBorder: "#FFFAEB",
  colorBorderSecondary: "#FFFAEB",
}

const lightBodyBg = "#FFFAEB"
const darkBodyBg = "#1A1825"

const illustrationShadow = "2px 2px 0 var(--ant-color-border)"

const panelRoot: CSSProperties = {
  border: "2px solid var(--ant-color-border)",
  boxShadow: illustrationShadow,
}

const tagRootStyle: CSSProperties = {
  border: "2px solid var(--ant-color-border)",
  fontWeight: 600,
  margin: 0,
  boxShadow: illustrationShadow,
}

const progressRailStyle: CSSProperties = {
  border: "none",
  borderRadius: 0,
  boxShadow: `inset 0 0 0 var(--ant-line-width) var(--ant-color-border), ${illustrationShadow}`,
}

const progressTrackStyle: CSSProperties = {
  border: "none",
  borderRadius: 0,
}

const useStyles = createStyles(({ cssVar }, isDark: boolean) => {
  const btnInteractive = {
    transition: `box-shadow ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
    "&:hover, &:active": { boxShadow: "none" },
    "&:active": { transform: "translate(2px, 2px)" },
  }

  return {
    inputRoot: {
      '&:not(.ant-input-group-wrapper)[class*="-outlined"]': {
        boxShadow: `2px 2px 0 ${cssVar.colorBorder}`,
        transition: `box-shadow ${cssVar.motionDurationFast}`,
        "&:not([class*='-status-error']):not([class*='-status-warning']):focus, &:not([class*='-status-error']):not([class*='-status-warning']):focus-within":
          {
            boxShadow: `2px 2px 0 ${cssVar.colorPrimary}`,
          },
      },
    },
    buttonDefaultRoot: {
      border: `${cssVar.lineWidth} solid ${cssVar.colorBorder}`,
      ...btnInteractive,
    },
    buttonOutlinedRoot: {
      border: `${cssVar.lineWidth} solid ${cssVar.colorBorder}`,
      ...btnInteractive,
    },
    buttonDashedRoot: {
      border: `${cssVar.lineWidth} dashed ${cssVar.colorBorder}`,
      ...btnInteractive,
    },
    buttonFilledRoot: {
      border: `${cssVar.lineWidth} solid ${cssVar.colorBorder}`,
      background: `color-mix(in srgb, ${cssVar.colorPrimary} 14%, ${cssVar.colorBgContainer})`,
      color: cssVar.colorPrimary,
      fontWeight: 600,
      transition: `background ${cssVar.motionDurationFast}, box-shadow ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
      "&:hover": {
        background: `color-mix(in srgb, ${cssVar.colorPrimary} 22%, ${cssVar.colorBgContainer})`,
        boxShadow: "none",
      },
      "&:active": { boxShadow: "none", transform: "translate(2px, 2px)" },
    },
    buttonSolidPrimaryRoot: {
      border: `${cssVar.lineWidth} solid ${cssVar.colorBorder}`,
      ...btnInteractive,
    },
    buttonSolidColoredRoot: {
      border: `${cssVar.lineWidth} solid ${cssVar.colorBorder}`,
      ...btnInteractive,
    },
    buttonColoredDefaultRoot: {
      ...btnInteractive,
    },
    buttonColoredFilledRoot: {
      border: `${cssVar.lineWidth} solid ${cssVar.colorBorder}`,
      fontWeight: 600,
      transition: `background ${cssVar.motionDurationFast}, box-shadow ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
      "&:hover, &:active": { boxShadow: "none" },
      "&:active": { transform: "translate(2px, 2px)" },
    },
    switchRoot: {
      border: "none",
      boxShadow: `inset 0 0 0 ${cssVar.lineWidth} ${cssVar.colorBorder}`,
      background: isDark ? cssVar.colorBgContainer : cssVar.colorBgBase,
      "&.ant-switch-checked": { background: cssVar.colorSuccess },
      "& .ant-switch-inner .ant-switch-inner-unchecked": {
        color: cssVar.colorText,
      },
    },
    stepsSliderOverride: {
      "& .ant-steps-item-icon": { borderRadius: 0 },
      "& .ant-slider-handle": { borderRadius: 0 },
    },
  }
})

function IllustrationProvider({
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
            defaultShadow: illustrationShadow,
            primaryShadow: "none",
            dangerShadow: "none",
            fontWeight: 600,
          },
          Input: {
            activeShadow: "none",
          },
          Card: {
            boxShadow: illustrationShadow,
            colorBgContainer:
              "color-mix(in srgb, var(--ant-color-primary) 6%, var(--ant-color-bg-container))",
          },
          Select: {
            optionSelectedColor: isDark
              ? darkTokens.colorPrimary
              : lightTokens.colorPrimary,
            optionSelectedBg: isDark
              ? `color-mix(in srgb, ${darkTokens.colorPrimary} 15%, transparent)`
              : `color-mix(in srgb, ${lightTokens.colorPrimary} 12%, transparent)`,
            optionActiveBg: isDark
              ? `color-mix(in srgb, ${darkTokens.colorPrimary} 8%, transparent)`
              : `color-mix(in srgb, ${lightTokens.colorPrimary} 6%, transparent)`,
          },
          Switch: {
            handleBg: "#FFFFFF",
            handleShadow: "2px 2px 0 rgba(0,0,0,0.25)",
          },
        },
      },
      alert: { styles: { root: panelRoot } },
      button: {
        classNames: ({ props }) => {
          const kind = pickButtonVariant(props)
          if (kind === "skip") return {}
          const isPrimary =
            props.color === "primary" || props.type === "primary"
          const isDefaultColor = !props.color || props.color === "default"
          if (kind === "solid" && !isPrimary) {
            return { root: styles.buttonSolidColoredRoot }
          }
          if (isDefaultColor && kind === "filled") return {}
          if (!isDefaultColor && kind === "filled") {
            return { root: styles.buttonColoredFilledRoot }
          }
          if (
            !isDefaultColor &&
            (kind === "outlined" || kind === "default" || kind === "dashed")
          ) {
            return { root: styles.buttonColoredDefaultRoot }
          }
          const map: Record<Exclude<ButtonVariantKind, "skip">, string> = {
            solid: styles.buttonSolidPrimaryRoot,
            dashed: styles.buttonDashedRoot,
            filled: styles.buttonFilledRoot,
            outlined: styles.buttonOutlinedRoot,
            default: styles.buttonDefaultRoot,
          }
          return { root: map[kind] }
        },
      },
      card: { styles: { root: panelRoot } },
      colorPicker: { styles: { root: panelRoot }, arrow: false },
      datePicker: {
        classNames: { root: styles.inputRoot },
        styles: { popup: { container: panelRoot } },
      },
      drawer: { styles: { section: panelRoot } },
      dropdown: { styles: { root: panelRoot } },
      input: { classNames: { root: styles.inputRoot } },
      inputNumber: { classNames: { root: styles.inputRoot } },
      textArea: { classNames: { root: styles.inputRoot } },
      modal: { styles: { container: panelRoot } },
      popover: { styles: { container: panelRoot } },
      progress: {
        styles: {
          rail: { ...progressRailStyle, height: compact ? 12 : 16 },
          track: { ...progressTrackStyle, height: compact ? 8 : 10 },
        },
      },
      select: {
        classNames: { root: styles.inputRoot },
        styles: { popup: { root: panelRoot } },
      },
      timePicker: { classNames: { root: styles.inputRoot } },
      cascader: { classNames: { root: styles.inputRoot } },
      treeSelect: { classNames: { root: styles.inputRoot } },
      mentions: { classNames: { root: styles.inputRoot } },
      switch: { classNames: { root: styles.switchRoot } },
      tag: { styles: { root: tagRootStyle } },
      tooltip: { styles: { root: panelRoot }, arrow: false },
    }),
    [styles, base, isDark, compact],
  )

  return (
    <ConfigProvider {...config}>
      <div className={styles.stepsSliderOverride}>{children}</div>
    </ConfigProvider>
  )
}

const illustrationSkin: SkinPlugin = {
  id: "illustration",
  name: "Illustration",
  bodyBg: { light: lightBodyBg, dark: darkBodyBg },
  Provider: IllustrationProvider,
}

export default illustrationSkin
