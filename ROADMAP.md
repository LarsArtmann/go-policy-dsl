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
2. **Validation policy decided.** Whether `Spec()` stays validation-free or
   grows a `Validate()` is settled and documented (currently: validation-free
   by design — see `AGENTS.md`).
3. **Stringly-typed axes reviewed.** The `CompanionOnly` → `Mode` rename is
   resolved (see `TODO_LIST.md`); the typed `Version` domain landed in
   `[Unreleased]` (2026-07-26) — `VersionMin`/`Max` are now `*Version` with
   inversion rejected at construction.
4. **BDD + Example tests green.** User-perspective behaviour is pinned by a
   Ginkgo suite and godoc `Example*` functions so refactor support is honest.

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
`ImportPatterns` / `GoModPatterns` (literal? glob? regex?). Raw ideas:

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
constructors that reject inversion (`min > max`) and non-numeric / signed
input — no semver dependency, so the stdlib-only contract holds. Eliminated
the stringly-typed footgun where `VersionRange("2.0.0", "1.0.0")` was
representable. Shipped in `[Unreleased]` (2026-07-26); see `FEATURES.md`
"Version constraints".

### Typed `Mode` enum (replace `CompanionOnly bool`)

See `TODO_LIST.md` [T4]. Long-term the policy "mode" (ban / companion-only /
both) should be a typed enum, not a boolean flag. Still open.

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

### CODEOWNERS, issue/PR templates, license-scan gate

Ecosystem hygiene. Add `CODEOWNERS`, issue/PR templates, and a `depguard`
gate that keeps the stdlib-only contract enforced (reject any future non-stdlib
dependency). The depguard config already enforces this at the linter level.

---

## Direction: documentation presence

### Public docs website

Consider the LarsArtmann Astro + Starlight + Firebase Hosting pattern for a
public docs site (sibling SDK libraries use it). Low urgency until adoption
widens; pkg.go.dev + README cover the current audience.

### `Severity` ↔ `finding.Severity` bridge examples

See `TODO_LIST.md` [T9]. A runnable bridge sample makes the
dependency-free-`Severity` design concrete for consumer authors.
