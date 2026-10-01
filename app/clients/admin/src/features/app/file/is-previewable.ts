export function isPreviewable(contentType: string): boolean {
  const type = contentType.toLowerCase();
  if (
    type.startsWith("image/") ||
    type.startsWith("video/") ||
    type.startsWith("audio/") ||
    type.startsWith("text/")
  ) {
    return true;
  }
  return type === "application/pdf" || type === "application/json";
}
