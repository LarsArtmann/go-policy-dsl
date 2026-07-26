# TODO List — go-policy-dsl

Short- and mid-term actionable work, bounded and estimable. Each item cites
evidence (code path or source report). When an item ships, it is **removed**
from this file and recorded in `CHANGELOG.md` — completed work never lives
here. Vague or unbounded ideas belong in `ROADMAP.md`, not here.

> Source reports harvested: `docs/status/2026-07-26_07-09_full-review-self-critique.md`
> (P0–P4 lists), `docs/status/2026-07-19_01-38_sdk-release-and-buildflow-coverage-fix.md`
> (§f next-50 list). Items already shipped were dropped; their entries live in
> `CHANGELOG.md` `[Unreleased]`.
>
> **2026-07-26 update:** the API-decision backlog closed — `[T1]` (opaque-pattern
> fuzz contract), `[T2]` (typed `Mode`), `[T3]` (validated `CVE`), `[T4]`
> (`[]Replacement` alternatives), `[T5]` (`SuggestExplicit`), and `[T6]`
> (`Require` decided against) all shipped/decided. See `CHANGELOG.md`
> `[Unreleased]`. Only the external-gated `[T7]` remains.

---

## Ecosystem (only after first consumer migrates)

### [T7] Add a "Consumers" section update once `library-policy` migrates

README "Consumers" lists `library-policy` as the primary consumer with
`domain/policy/spec.go` as the migration target, and already notes the module
is **not yet imported**. Update with the actual migration commit/PR once it
lands.

**Blocked-on-external (verified 2026-07-26):** `library-policy` does NOT yet
import `github.com/larsartmann/go-policy-dsl` — `go.mod` has no such dependency
and no `.go` file references the `policydsl.` package (checked directly). The
cutover is planned but unshipped: ADR `docs/adr/0006-go-dsl-cutover-requires-
semantic-audit.md` exists and several status reports
(`2026-07-25_*_dsl-cutover-analysis-and-gate-gaps.md`,
`2026-07-25_*_go-dsl-migration-and-repo-ref-cleanup.md`) track the prototype in
`policies/policies.go`, but `library-policy` still carries its own independent
`PolicySpec`/`CompanionOnly` copy in `domain/library/banned_library.go`.

This item **cannot** be completed until that cutover lands in `library-policy`;
doing so here would be fabrication. Unblock trigger: a `library-policy` PR that
adds `require github.com/larsartmann/go-policy-dsl` to its `go.mod` and imports
`policydsl.Ban(...)`. At that point, also revisit whether any BREAKING change
made this session (`Mode`, `[]Replacement`, `[]CVE`) needs a coordinated update
on the consumer side.
_Source: 2026-07-26 §f.40. Evidence: `README.md` "Consumers"; `library-policy`
`go.mod` + `rg "github.com/larsartmann/go-policy-dsl" --include=*.go`._
