---
paths: ["app/services/**/*.go", "shared-go/**/*.go"]
---

# Kratos law card

- Before writing: load the dispatch row for this artifact — docs/conventions/kratos/CONVENTIONS.md §0 "Execution contract"; those sections are the law for the edit.
- Before reporting done: run docs/conventions/kratos/CONVENTIONS.md §7 "Pre-completion gate" over the combined diff and emit its report table.
- `shared-go` changes emit only K-ERR and K-DEP-1 (capability self-containment, `AGENTS.md` §4).
- Procedures (new module, new service, remote dependency, operating facts): docs/conventions/kratos/RUNBOOKS.md.
