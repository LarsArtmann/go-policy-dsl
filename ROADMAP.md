# Roadmap — go-policy-dsl

Long-term direction and raw ideas not yet refined into actionable tasks.
Items here are **not commitments** — they graduate to `TODO_LIST.md` when
they become bounded, estimable, and owned. For shipped work, see
`CHANGELOG.md`; for the current feature inventory, see `FEATURES.md`.

---

## The path to `v1.0.0`

The API is not yet stable. `v1.0.0` is reached when **all** of these hold:

1. **First consumer shipped.** `library-policy` migrates its
   `domain/policy/spec.go` onto this module and releases. This is the load-bearing
   gate — until a real consumer exercises the API, "stable" is a claim, not a
   fact.
2. **Validation policy decided.** `Spec()` stays validation-free by design
   (returns exactly what was built); `PolicySpec.Validate()` checks only the
   version-range invariant. Mode values are NOT validated (deny-by-default:
   only `ModeCompanionOnly` suppresses the ban; see `AGENTS.md`). Struct tags
   are deliberately absent (values, not config; see `AGENTS.md`). Both
   decisions are settled and documented.
3. **Stringly-typed axes reviewed.** The `CompanionOnly bool` → typed `Mode`
   enum landed in `[Unreleased]` (2026-07-26); `CVEs []string` → branded `[]CVE`
   with format validation landed the same session; `Alternatives []string` →
   `[]Replacement` (no information loss). The typed `Version` domain landed
   earlier — `VersionMin`/`Max` are `*Version` with inversion detected by
   `Validate()` (the library is panic-free by design).
4. **BDD + Example tests green.** User-perspective behaviour is pinned by a
   stdlib BDD-style suite (`TestBehavior_*`, deliberately NOT Ginkgo to honour
   the zero-dependency contract) and godoc `Example*` functions so refactor
   support is honest.

Until then: breaking changes are allowed in any `0.x` bump and are always
listed under `CHANGELOG.md` → `Changed` with a `**BREAKING**` marker.

---

## Direction: adoption

### `library-policy` migration (primary consumer)

`library-policy` currently has its own independent copy of the `PolicySpec` /
`Builder` surface in `domain/policy/spec.go`. The migration target is to
import `github.com/larsartmann/go-policy-dsl` directly and delete the local
copy. The rename `GoVersionRange` → `VersionRange` (and friends) was verified
safe against this consumer: it does not yet import this module.

### `go-linter-sdk` adoption (secondary consumer)

`go-linter-sdk` may adopt this as its rule-declaration language. No code yet.

### golangci-lint plugin template

A template repo that consumes this DSL and declares library-governance rules
as a golangci-lint plugin. Low effort once the DSL is stable; high leverage
for ecosystem adoption.

---

## Direction: the matching / execution boundary

The DSL intentionally declares _what_ a policy is, not _how_ it is detected.
Every consumer currently reinvents the matching semantics for
`ImportPatterns` / `GoModPatterns` (literal? glob? regex?).

**Contract (pinned):** the DSL owns NO matching semantics — patterns are
**opaque strings**, stored verbatim and never interpreted. This is fuzz-pinned
by `FuzzBuilder_PatternsOpaque` (any string round-trips unchanged through every
pattern entry point). A pattern-matching fuzz target therefore cannot live in
the DSL (there is nothing to match against); consumers own that. Raw ideas:

- **`matcher` subpackage** — a tiny stdlib-only glob matcher on import paths,
  so consumers don't reinvent it. Decide if in-scope or left to consumers.
- **`ast`-based detector reference** — behind a build tag, opt-in, so a
  consumer can copy a known-correct detector instead of writing one.
- **LSP server** — reads policies and squiggles banned imports in the editor.
  High leverage, high effort; only after the DSL is stable.
- **Policy exchange format** — YAML ↔ Go round-trip for cross-org policy
  sharing. Decide if in-scope; the DSL's whole premise is "values, not YAML",
  so this may belong in a consumer, not here.

These are explicitly **not** committed. They exist to keep the "no execution
model" line in `AGENTS.md` honest: the DSL is not expected to grow an
execution model without a deliberate decision recorded here first.

---

## Direction: type safety

### Typed `Version` domain — LANDED

A hand-rolled semver-lite type (`Version{Major, Minor, Patch int}`) with
constructors that reject non-numeric / signed input — no semver dependency, so
the stdlib-only contract holds. Replaced the stringly-typed `VersionMin/Max`
fields with `*Version`, making the bounds explicit (an unbounded side is `nil`,
not the empty string). Inversion (`min > max`) is **not** rejected at
construction — the library is panic-free by design — but is detected by
`PolicySpec.Validate()`, which returns `*InvertedVersionRangeError`. Shipped in
`[Unreleased]` (2026-07-26); see `FEATURES.md` "Version constraints".

### Typed `Mode` enum (replace `CompanionOnly bool`) — LANDED

The policy "mode" is now a typed enum, not a boolean flag. `ModeBan` (default)
emits a ban and enforces companions; `ModeCompanionOnly` (set via
`AsCompanionOnly()`) suppresses the ban. The former `CompanionOnly bool` field
was removed — it lied (it suppressed the ban, not the companion). The contract
is deny-by-default: the ONLY mode that suppresses the ban is
`ModeCompanionOnly`; every other value (empty string, `ModeBan`, unknown
strings) is ban-active — a typo can never silently disable enforcement. Pinned
by `TestMode_DenyByDefaultContract`. Shipped in `[Unreleased]` (2026-07-26);
see `FEATURES.md` "Companion policy".

### Branded `CVE` and `[]Replacement` — LANDED

`CVEs []string` is now `[]CVE` (validated `CVE-YYYY-NNNN` via `NewCVE`), and `Alternatives []string` is now `[]Replacement` (each entry
keeps its `Reason` — no information loss). Both eliminate stringly-typed
footguns without adding a dependency. Shipped in `[Unreleased]` (2026-07-26).

### Decided against: a `Require(name)` builder

A distinct "required library" policy type (`Require(...)`) was considered and
**deliberately not added**: no consumer needs a third policy kind beyond ban
and companion, and adding it would be speculative API surface (YAGNI). A
phantom `Require` once appeared in the package docs by mistake and was
removed; it will not become real unless a concrete consumer demand appears.

---

## Direction: release & visibility

### First tagged release

The module has never been tagged. The first tag will be `v0.2.0` (the
`GoVersionRange` → `VersionRange` rename is breaking; `v0.1.0` was only ever
scaffolding, never published). Track in `TODO_LIST.md` once the v1.0.0
criteria above are closer.

### Go module proxy / GOPROXY visibility

Verify the module is visible on `proxy.golang.org` after the first tag. The
pkg.go.dev badge in README already links to the module; confirm it resolves
post-publish.

### Release automation

Tag → pkg.go.dev update. Decide between a manual tag-and-push flow or a
GitHub Action. Low urgency until the first consumer is ready to pin a version.

### Ecosystem hygiene — LANDED

`CODEOWNERS`, issue templates, a pull-request template, a GitHub Actions
CI workflow (`.github/workflows/ci.yml`) enforcing the full quality gate
(build, vet, test -race, `golangci-lint run`, `golangci-lint fmt` drift
check) shipped in `[Unreleased]` (2026-07-26). A dedicated fuzz job runs
both fuzz targets for 30s each on every push/PR. `.github/dependabot.yml`
keeps GitHub Actions versions fresh automatically. The `depguard` config keeps
the stdlib-only contract enforced at the linter level (reject any future
non-stdlib dependency).

---

## Direction: documentation presence

### Public docs website

Consider the LarsArtmann Astro + Starlight + Firebase Hosting pattern for a
public docs site (sibling SDK libraries use it). Low urgency until adoption
widens; pkg.go.dev + README cover the current audience.

### `Severity` ↔ `finding.Severity` bridge examples

A runnable bridge sample makes the dependency-free-`Severity` design concrete
for consumer authors (the README has one; a godoc `Example*` would round out
the surface). Low urgency until the first consumer cutover.
