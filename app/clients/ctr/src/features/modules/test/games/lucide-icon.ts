// lucide geometry rendered to real SVG DOM. Game modules render raw DOM
// (the GameModule contract — no React inside), so lucide-react's
// components cannot serve them; this helper takes lucide's icon node
// trees (the same [tag, attrs] shape the package ships) and builds the
// identical SVG, keeping every in-game icon lucide-shaped per the
// repo's zero-emoji ruling.
export type IconNode = readonly (readonly [tag: string, attrs: Readonly<Record<string, string>>])[];

export function createLucideIcon(node: IconNode, size: number): SVGSVGElement {
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("xmlns", "http://www.w3.org/2000/svg");
  svg.setAttribute("width", String(size));
  svg.setAttribute("height", String(size));
  svg.setAttribute("viewBox", "0 0 24 24");
  svg.setAttribute("fill", "none");
  svg.setAttribute("stroke", "currentColor");
  svg.setAttribute("stroke-width", "2");
  svg.setAttribute("stroke-linecap", "round");
  svg.setAttribute("stroke-linejoin", "round");
  for (const [tag, attrs] of node) {
    const el = document.createElementNS("http://www.w3.org/2000/svg", tag);
    for (const [name, value] of Object.entries(attrs)) {
      // lucide's "key" attr is React bookkeeping, not geometry.
      if (name !== "key") {
        el.setAttribute(name, value);
      }
    }
    svg.appendChild(el);
  }
  return svg;
}
