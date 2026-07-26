# TODO List — go-policy-dsl

Short- and mid-term actionable work, bounded and estimable. Each item cites
evidence (code path or source report). When an item ships, it is **removed**
from this file and recorded in `CHANGELOG.md` — completed work never lives
here. Vague or unbounded ideas belong in `ROADMAP.md`, not here.

> **Harvest log:** items T1–T6 shipped in prior sessions (see `CHANGELOG.md`
> `[Unreleased]`). T7 below is externally blocked. T8–T16 shipped in the
> 2026-07-26 session (see `CHANGELOG.md` `[Unreleased]` for details).

---

## Ecosystem (blocked-on-external)

### [T7] Update "Consumers" section once `library-policy` migrates

README "Consumers" lists `library-policy` as the primary consumer and notes the
module is **not yet imported**. Update with the actual migration commit/PR once
it lands.

**Blocked-on-external (verified 2026-07-26):** `library-policy` does NOT yet
import `github.com/larsartmann/go-policy-dsl` — `go.mod` has no such dependency
and no `.go` file references the `policydsl.` package (checked directly). The
cutover is planned but unshipped.

This item **cannot** be completed until that cutover lands in `library-policy`.
Unblock trigger: a `library-policy` PR that adds
`require github.com/larsartmann/go-policy-dsl` to its `go.mod` and imports
`policydsl.Ban(...)`. At that point, also revisit whether any BREAKING change
(`Mode`, `[]Replacement`, `[]CVE`, panic-free removal) needs a coordinated
update on the consumer side.
_Evidence: `README.md` "Consumers"; `library-policy` `go.mod` +
`rg "github.com/larsartmann/go-policy-dsl" --include=*.go` = 0 hits._
