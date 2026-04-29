export type ButtonVariantKind =
  | "skip"
  | "solid"
  | "dashed"
  | "filled"
  | "outlined"
  | "default"

export function pickButtonVariant(props: {
  variant?: string
  type?: string
  color?: string
}): ButtonVariantKind {
  const { variant, type, color } = props
  const isColored = !!color && color !== "default"
  if (
    variant === "text" ||
    variant === "link" ||
    type === "text" ||
    type === "link"
  )
    return "skip"
  if (type === "primary") return "solid"
  if (variant === "solid") return "solid"
  if (variant === "dashed" || type === "dashed") return "dashed"
  if (variant === "filled") return "filled"
  if (variant === "outlined" && isColored) return "outlined"
  return "default"
}
