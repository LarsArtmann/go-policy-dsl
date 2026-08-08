# Roadmap — go-policy-dsl

Long-term direction and raw ideas — not commitments. Items graduate to
`TODO_LIST.md` when they become bounded and estimable. For shipped work, see
`CHANGELOG.md`; for the current feature inventory, see `FEATURES.md`.

---

## The path to v1.0.0

The API is not yet stable. v1.0.0 is reached when **all** of these hold:

1. ~~**First consumer shipped.**~~ **MET (2026-08-08).** `library-policy`
   imports `github.com/larsartmann/go-policy-dsl` v0.3.0 directly (no local
   copy). The API has been exercised by a real consumer.
2. **Validation surface settled.** `Spec()` stays validation-free (returns
   exactly what was built); `Validate()` checks the version-range invariant.
   Whether `Validate()` expands to cover domain rules (non-empty `Reason`,
   detection patterns present, etc.) is the remaining open question. The
   first consumer adopted the DSL as-is, so the current validation surface
   was sufficient — but whether `library-policy` revealed unmet validation
   needs is not yet confirmed. See `FEATURES.md`
   "Domain validation in `Validate()`".
3. **No stringly-typed axes remaining.** All formerly stringly-typed fields
   have been retyped: `Version` (semver-lite, `*Version` bounds), `Mode`
   (typed enum, deny-by-default), `CVE` (branded string with format
   validation), `Alternatives` (`[]Replacement`, no information loss). No
   known stringly-typed footguns remain.

With gate 1 met, the remaining question for v1.0.0 is gate 2: does the
validation surface need to expand based on what `library-policy` revealed,
or is the current surface sufficient? Once that is answered, v1.0.0 can be
tagged.

Until v1.0.0: breaking changes are allowed in any `0.x` bump and are always
listed in `CHANGELOG.md` under `Changed (breaking)` with a `**BREAKING**`
marker.

---

## Direction: adoption

### `library-policy` (shipped consumer)

`library-policy` imports `github.com/larsartmann/go-policy-dsl` v0.3.0
and has deleted its local `domain/policy/spec.go` copy. The migration is
complete — the DSL's API was sufficient as-is, with no blocking gaps
surfaced.

### `go-linter-sdk` adoption (secondary consumer)

`go-linter-sdk` may adopt this as its rule-declaration language. No code yet.

### golangci-lint plugin template

A template repo that consumes this DSL and declares library-governance rules
as a golangci-lint plugin. Low effort once the DSL is stable; high leverage
for ecosystem adoption.

---

## Direction: the matching / execution boundary

The DSL declares _what_ a policy is, not _how_ it is detected. Every consumer
currently reinvents matching semantics for `ImportPatterns` / `GoModPatterns`
(literal? glob? regex?).

**Contract (pinned):** the DSL owns NO matching semantics — patterns are
opaque strings, stored verbatim and never interpreted. Fuzz-pinned by
`FuzzBuilder_PatternsOpaque` (any string round-trips unchanged through every
pattern entry point). Raw ideas:

- **`matcher` subpackage** — a tiny stdlib-only glob matcher on import
  paths, so consumers don't reinvent it. Decide if in-scope or left to
  consumers.
- **`ast`-based detector reference** — behind a build tag, opt-in, so a
  consumer can copy a known-correct detector instead of writing one.
- **LSP server** — reads policies and squiggles banned imports in the
  editor. High leverage, high effort; only after the DSL is stable.
- **Policy exchange format** — YAML ↔ Go round-trip for cross-org policy
  sharing. The DSL's premise is "values, not YAML", so this may belong in a
  consumer, not here.

These are explicitly **not** committed. They exist to keep the "no execution
model" line in `AGENTS.md` honest: the DSL is not expected to grow an
execution model without a deliberate decision recorded here first.

---

## Direction: type safety

### Domain validation expansion

`Spec()` performs zero validation. `Validate()` currently checks only the
version-range invariant. Domain rules — a `Ban("x")` with no `Because`, no
detection patterns, no overrides — are deliberately not enforced: the DSL
declares what a policy IS; the consumer validates domain fitness. Expanding
`Validate()` to cover domain rules is a future decision, deferred until the
first consumer migrates and the real required-field set is known. See
`FEATURES.md` "Domain validation in `Validate()`".

### Decided against: a `Require(name)` builder

A distinct "required library" policy type (`Require(...)`) was considered
and deliberately not added: no consumer needs a third policy kind beyond ban
and companion, and adding it would be speculative API surface (YAGNI). A
phantom `Require` once appeared in package docs by mistake and was removed.
It will not become real unless a concrete consumer demand appears.

### Open question: are all `Severity` values earning their keep?

The enum ships eight values (`critical` through `obsolete`). The first five
are battle-tested via consumer mappings; the last three (`recommended`,
`deprecated`, `obsolete`) have no known consumer yet. They may be
over-engineered, or they may prove useful now that `library-policy` has
adopted the DSL and `go-linter-sdk` evaluates it. Check which values
`library-policy` actually maps and prune the unused ones if warranted.

---

## Direction: release & visibility

### Go module proxy / GOPROXY visibility

The repo is **public** and all three versions (v0.1.0–v0.3.0) are confirmed
indexed on `proxy.golang.org`. pkg.go.dev renders the module page from the
**latest tag** (v0.3.0), not HEAD — so the README served on pkg.go.dev is
the v0.3.0-tagged version (includes a dead Go Report Card badge, different
structure). The current HEAD README is the rewritten public-facing landing
page. pkg.go.dev will self-correct on the next tag push; a manual re-index
request is also possible but the next release tag is the clean path.

### Release automation

Tag → pkg.go.dev update currently uses a manual tag-and-push flow. A GitHub
Action that creates a GitHub release from `CHANGELOG.md` sections (and
optionally generates an SBOM) would reduce manual steps. Now that the first
consumer has pinned v0.3.0, release automation reduces friction for future
consumer updates.

---

## Direction: documentation presence

### Public docs website

Consider the LarsArtmann Astro + Starlight + Firebase Hosting pattern for a
public docs site (sibling SDK libraries use it). Low urgency until adoption
widens; pkg.go.dev + README cover the current audience.

### `Severity` ↔ `finding.Severity` bridge examples

A runnable bridge sample makes the dependency-free `Severity` design
concrete for consumer authors (the README has one; a godoc `Example*` would
round out the surface). `library-policy` has already bridged this in
practice — the example would document the known-good pattern.

---

## Direction: CI & infrastructure

### Branch protection

**Decided against** (2026-08-08). Master will not require CI-green status
checks. The auto-git daemon commits directly to master; enforcing PRs or
status-check gates would block the daemon's workflow. The user has
disapproved branch protection. CI remains green and serves as a quality
signal even without enforcement.

### CI hardening ideas

Raw ideas, not committed: add `govulncheck` (periodic or per-PR), test
coverage reporting with a threshold gate, a Go versions matrix, and
`actionlint` to catch malformed workflow YAML. The current CI is fast (37s
quality gate via `golangci-lint-action`) and green; these are polish, not
gaps.
