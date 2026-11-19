// ledger-guard — docs/roadmap/ holds working ledgers; they never enter the
// repository on the skeleton lanes. Business lanes carry their own committed
// ledgers.
import { execSync } from "node:child_process";
import { stagedFiles } from "./lib.mjs";

const branch = execSync("git branch --show-current", { encoding: "utf8" }).trim();
const onSkeletonLane = branch === "dev" || branch === "main";

if (
  onSkeletonLane &&
  stagedFiles().some((f) => f.startsWith("docs/roadmap/"))
) {
  console.error("docs/roadmap/ holds working ledgers — never committed on dev/main");
  process.exit(1);
}
