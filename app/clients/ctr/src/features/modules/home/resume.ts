import { type Timestamp, timestampMs } from "@bufbuild/protobuf/wkt";

/* The wire stamp a resume card consumes: the generated type claims the WKT
 * message while the JSON transport flattens it to its RFC 3339 string, so
 * the declared union is narrowed by typeof — reading .seconds off a wire
 * string yields a NaN stamp, and a malformed one degrades to "started
 * today". */
export type WireStamp = Timestamp | string | undefined;

/* The resume card's desc vars: the progress counts plus the age face —
 * the prototype's inclusive day count (started today reads 已过 1 天) and
 * a local stamp to the minute (the prototype names the start moment, not
 * just its date). */
export function resumeDescVars(createdAt: WireStamp, done: number, total: number) {
  let created: number;
  if (typeof createdAt === "string") {
    created = Date.parse(createdAt);
  } else if (createdAt !== undefined) {
    created = timestampMs(createdAt);
  } else {
    created = Date.now();
  }
  if (Number.isNaN(created)) {
    created = Date.now();
  }
  const days = Math.max(1, Math.floor((Date.now() - created) / 86_400_000) + 1);
  const d = new Date(created);
  const pad = (n: number) => String(n).padStart(2, "0");
  const stamp = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
  return { done, total, date: stamp, days };
}
