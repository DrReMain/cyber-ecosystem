export function safeRedirect(
  target: string | undefined,
  area: Readonly<{ home: string; self: string }>,
): string {
  if (target === undefined) {
    return area.home;
  }
  if (!target.startsWith("/") || target.startsWith("//") || target.startsWith("/\\")) {
    return area.home;
  }
  const firstSegment = `/${target.slice(1).split(/[/?#]/)[0]}`;
  if (firstSegment === area.self) {
    return area.home;
  }
  return target;
}
