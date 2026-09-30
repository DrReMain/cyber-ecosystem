export function noindexHead(): { meta: Array<{ name: "robots"; content: "noindex" }> } {
  return { meta: [{ name: "robots", content: "noindex" }] };
}
