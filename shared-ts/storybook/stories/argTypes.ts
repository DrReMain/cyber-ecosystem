export const sizeArg = (
  options: readonly string[] = ["large", "medium", "small"],
) => ({
  control: "radio" as const,
  options,
})

export const variantArg = (
  options: readonly string[] = [
    "outlined",
    "filled",
    "borderless",
    "underlined",
  ],
) => ({
  control: "select" as const,
  options,
})

export const statusArg = {
  control: "radio" as const,
  options: ["warning", "error"],
}

export const disabledArg = { control: "boolean" as const }
export const loadingArg = { control: "boolean" as const }
export const borderedArg = { control: "boolean" as const }

export const callbackArgs = {
  onChange: { action: "changed" },
  onClick: { action: "clicked" },
  onFocus: { action: "focused" },
  onBlur: { action: "blurred" },
  onSelect: { action: "selected" },
  onOpenChange: { action: "openChanged" },
}
