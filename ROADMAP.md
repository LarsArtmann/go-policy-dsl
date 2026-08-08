# Roadmap — go-policy-dsl

Long-term direction and raw ideas — not commitments. Items graduate to
`TODO_LIST.md` when they become bounded and estimable. For shipped work, see
`CHANGELOG.md`; for the current feature inventory, see `FEATURES.md`.

---

## The path to v1.0.0

The API is not yet stable. v1.0.0 is reached when **all** of these hold:

1. **First consumer shipped.** `library-policy` migrates its
   `domain/policy/spec.go` onto this module and releases. This is the
   load-bearing gate — until a real consumer exercises the API, "stable" is
   a claim, not a fact.
2. **Validation surface settled.** `Spec()` stays validation-free (returns
   exactly what was built); `Validate()` checks the version-range invariant.
   Whether `Validate()` expands to cover domain rules (non-empty `Reason`,
   detection patterns present, etc.) is deferred until the first consumer
   migration reveals the real required-field set. See `FEATURES.md`
   "Domain validation in `Validate()`".
3. **No stringly-typed axes remaining.** All formerly stringly-typed fields
   have been retyped: `Version` (semver-lite, `*Version` bounds), `Mode`
   (typed enum, deny-by-default), `CVE` (branded string with format
   validation), `Alternatives` (`[]Replacement`, no information loss). No
   known stringly-typed footguns remain.

Until v1.0.0: breaking changes are allowed in any `0.x` bump and are always
listed in `CHANGELOG.md` under `Changed (breaking)` with a `**BREAKING**`
marker.

---

## Direction: adoption

### `library-policy` migration (primary consumer)

`library-policy` has its own copy of the `PolicySpec` / `Builder` surface in
`domain/policy/spec.go`. The migration target: import
`github.com/larsartmann/go-policy-dsl` directly and delete the local copy.
The `GoVersionRange` → `VersionRange` rename was verified safe against this
consumer (it does not yet import this module).

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

---

## Direction: release & visibility

### Go module proxy / GOPROXY visibility

Verify the module is visible on `proxy.golang.org` after the first tag. The
pkg.go.dev badge in README links to the module; confirm it resolves
post-publish. The repo is currently private — the proxy check is only
relevant if visibility flips to public.

### Release automation

Tag → pkg.go.dev update. Decide between a manual tag-and-push flow or a
GitHub Action. Low urgency until the first consumer is ready to pin a
version.

---

## Direction: documentation presence

### Public docs website

Consider the LarsArtmann Astro + Starlight + Firebase Hosting pattern for a
public docs site (sibling SDK libraries use it). Low urgency until adoption
widens; pkg.go.dev + README cover the current audience.

### `Severity` ↔ `finding.Severity` bridge examples

A runnable bridge sample makes the dependency-free `Severity` design
concrete for consumer authors (the README has one; a godoc `Example*` would
round out the surface). Low urgency until the first consumer cutover.
