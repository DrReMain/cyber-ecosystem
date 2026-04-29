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
  lineWidth: 1.5,
  lineWidthBold: 1.5,
  borderRadius: 8,
  borderRadiusSM: 6,
  borderRadiusLG: 12,
  borderRadiusXS: 4,
  fontWeightStrong: 600,
  motionDurationFast: "0.12s",
  motionDurationMid: "0.2s",
  motionDurationSlow: "0.3s",
  fontFamily:
    "'Caveat', 'Marker Felt', 'Comic Sans MS', 'Bradley Hand', cursive",
  fontFamilyCode: "'Cascadia Code', 'Comic Mono', 'Courier New', monospace",
}

const lightTokens: Partial<GlobalToken> = {
  ...shared,
  colorPrimary: "#5B8DEF",
  colorInfo: "#5B8DEF",
  colorSuccess: "#7ED957",
  colorWarning: "#FFD93D",
  colorError: "#FF6B6B",
  colorBgBase: "#FFFEF7",
  colorBgContainer: "#FFFFFF",
  colorBgElevated: "#FFFFFF",
  colorBgLayout: "#FFFEF7",
  colorTextBase: "#2B2620",
  colorBorder: "#2B2620",
  colorBorderSecondary: "rgba(43,38,32,0.35)",
}

const darkTokens: Partial<GlobalToken> = {
  ...shared,
  colorPrimary: "#8DB4F7",
  colorInfo: "#8DB4F7",
  colorSuccess: "#A3E87D",
  colorWarning: "#FFE066",
  colorError: "#FF8E8E",
  colorBgBase: "#2A2620",
  colorBgContainer: "#3A332B",
  colorBgElevated: "#4A433B",
  colorBgLayout: "#2A2620",
  colorTextBase: "#FFFEF7",
  colorBorder: "#FFFEF7",
  colorBorderSecondary: "rgba(255,254,247,0.35)",
}

const dotGridLight = `radial-gradient(circle, rgba(43,38,32,0.12) 1px, transparent 1px) 0 0 / 24px 24px`
const dotGridDark = `radial-gradient(circle, rgba(255,254,247,0.08) 1px, transparent 1px) 0 0 / 24px 24px`

const lightBodyBg = `${dotGridLight} #FFFEF7`
const darkBodyBg = `${dotGridDark} #2A2620`

const wobbleLG = "16px 10px 18px 8px / 10px 16px 8px 14px"
const wobbleMD = "10px 6px 12px 8px / 6px 10px 8px 10px"
const wobbleSM = "8px 5px 9px 7px / 5px 8px 7px 9px"

const sketchShadow =
  "1px 1px 0 var(--ant-color-border), 2px 0 0 color-mix(in srgb, var(--ant-color-border) 25%, transparent), 0 2px 0 color-mix(in srgb, var(--ant-color-border) 25%, transparent)"

const panelRoot: CSSProperties = {
  border: "1.5px solid var(--ant-color-border)",
  borderRadius: wobbleLG,
  boxShadow: sketchShadow,
}

const tagRootStyle: CSSProperties = {
  border: "1.5px solid var(--ant-color-border)",
  borderRadius: wobbleSM,
  fontWeight: 600,
  margin: 0,
}

const progressRailStyle: CSSProperties = {
  border: "1.5px solid var(--ant-color-border)",
  borderRadius: wobbleSM,
  boxShadow: sketchShadow,
}

const progressTrackStyle: CSSProperties = {
  borderRadius: wobbleSM,
}

const useStyles = createStyles(({ cssVar }) => {
  const btnInteractive = {
    transition: `box-shadow ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
    "&:hover, &:active": { boxShadow: "none" },
    "&:active": { transform: "translate(2px, 2px)" },
  }

  return {
    inputRoot: {
      "&:not(.ant-input-compact-item):not(.ant-btn-compact-item)": {
        borderRadius: wobbleMD,
      },
      '&:not(.ant-input-group-wrapper)[class*="-outlined"]': {
        "&:not([class*='-status-error']):not([class*='-status-warning']):focus, &:not([class*='-status-error']):not([class*='-status-warning']):focus-within":
          { borderColor: cssVar.colorPrimary },
      },
    },
    buttonDefaultRoot: {
      border: `${cssVar.lineWidth} solid ${cssVar.colorBorder}`,
      borderRadius: wobbleMD,
      ...btnInteractive,
    },
    buttonOutlinedRoot: {
      border: `${cssVar.lineWidth} solid ${cssVar.colorBorder}`,
      borderRadius: wobbleMD,
      ...btnInteractive,
    },
    buttonDashedRoot: {
      border: `${cssVar.lineWidth} dashed ${cssVar.colorBorder}`,
      borderRadius: wobbleMD,
      ...btnInteractive,
    },
    buttonFilledRoot: {
      border: `${cssVar.lineWidth} solid ${cssVar.colorBorder}`,
      borderRadius: wobbleMD,
      background: `color-mix(in srgb, ${cssVar.colorPrimary} 18%, ${cssVar.colorBgContainer})`,
      color: cssVar.colorPrimary,
      fontWeight: 600,
      transition: `background ${cssVar.motionDurationFast}, box-shadow ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
      "&:hover": {
        background: `color-mix(in srgb, ${cssVar.colorPrimary} 28%, ${cssVar.colorBgContainer})`,
        boxShadow: "none",
      },
      "&:active": { boxShadow: "none", transform: "translate(2px, 2px)" },
    },
    buttonSolidPrimaryRoot: {
      border: `${cssVar.lineWidth} solid ${cssVar.colorBorder}`,
      borderRadius: wobbleMD,
      ...btnInteractive,
    },
    buttonSolidColoredRoot: {
      border: `${cssVar.lineWidth} solid ${cssVar.colorBorder}`,
      borderRadius: wobbleMD,
      ...btnInteractive,
    },
    buttonColoredDefaultRoot: {
      borderRadius: wobbleMD,
      ...btnInteractive,
    },
    buttonColoredFilledRoot: {
      border: `${cssVar.lineWidth} solid ${cssVar.colorBorder}`,
      borderRadius: wobbleMD,
      fontWeight: 600,
      transition: `background ${cssVar.motionDurationFast}, box-shadow ${cssVar.motionDurationFast}, transform ${cssVar.motionDurationFast}`,
      "&:hover, &:active": { boxShadow: "none" },
      "&:active": { transform: "translate(2px, 2px)" },
    },
    switchRoot: {
      border: "none",
      boxShadow: `inset 0 0 0 ${cssVar.lineWidth} ${cssVar.colorBorder}`,
      background: cssVar.colorBgContainer,
      "&.ant-switch-checked": { background: cssVar.colorSuccess },
      "& .ant-switch-inner .ant-switch-inner-unchecked": {
        color: cssVar.colorText,
      },
    },
    stepsSliderOverride: {
      "& .ant-steps-item-icon": { borderRadius: cssVar.borderRadius },
      "& .ant-slider-handle": { borderRadius: cssVar.borderRadiusSM },
    },
  }
})

function DoodleProvider({
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
          Layout: { bodyBg: isDark ? darkBodyBg : lightBodyBg },
          Button: {
            defaultShadow: sketchShadow,
            primaryShadow: sketchShadow,
            dangerShadow: "none",
            fontWeight: 600,
          },
          Modal: { boxShadow: "none" },
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
            handleShadow: "0 0 0 1.5px var(--ant-color-border)",
          },
          Tooltip: { borderRadius: 0 }, // fallback; overridden by inline styles
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
          if (!isDefaultColor && kind === "filled") {
            return { root: styles.buttonColoredFilledRoot }
          }
          if (isDefaultColor && kind === "filled") return {}
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
          rail: progressRailStyle,
          track: progressTrackStyle,
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
      tooltip: { styles: { container: panelRoot }, arrow: false },
    }),
    [styles, base, isDark, compact],
  )

  return (
    <ConfigProvider {...config}>
      <div className={styles.stepsSliderOverride}>{children}</div>
    </ConfigProvider>
  )
}

const doodleSkin: SkinPlugin = {
  id: "doodle",
  name: "Doodle",
  bodyBg: { light: lightBodyBg, dark: darkBodyBg },
  Provider: DoodleProvider,
}

export default doodleSkin
