// Shared text normalization for prototype-fidelity checks: the prototype
// corpus and the running app must go through the exact same shape here, so
// both sides import this module instead of re-implementing the rules.
export function normalizeLines(raw) {
  return String(raw ?? "")
    .split("\n")
    .map((l) => l.replace(/\s+/g, " ").trim())
    .filter((l) => l.length > 0);
}
