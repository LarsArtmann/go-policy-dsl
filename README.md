# go-policy-dsl

Fluent, compile-time-checked Go DSL for declaring library governance policies — bans, requirements, companions. Pure Go values, no YAML, no codegen, no runtime parsing.

[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/go-policy-dsl.svg)](https://pkg.go.dev/github.com/larsartmann/go-policy-dsl)
[![Go Report Card](https://goreportcard.com/badge/github.com/larsartmann/go-policy-dsl)](https://goreportcard.com/report/github.com/larsartmann/go-policy-dsl)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**[pkg.go.dev](https://pkg.go.dev/github.com/larsartmann/go-policy-dsl)**

---

## Why?

Library governance policies — "don't use gorm, use sqlc", "if you depend on samber/do you must also depend on samber-do-auditlog", "this library is banned for Go ≥ 1.22" — are typically declared in YAML and parsed at runtime. That means typos surface in CI, not in your editor; refactor support is zero; and there is no compile-time guarantee that every policy specifies the fields it needs.

`go-policy-dsl` makes policies plain Go values built via a fluent API. Every field is typed, every chain is checked by the compiler, and the result is a `PolicySpec` struct that any consumer (a CLI, a golangci-lint plugin, an LSP server, a CI check) can read.

It is the canonical policy-declaration language for the LarsArtmann ecosystem: [`library-policy`](https://github.com/LarsArtmann/library-policy) is the primary consumer, and `go-linter-sdk` can use it as its rule-declaration language.

---

## Installation

```bash
go get github.com/larsartmann/go-policy-dsl
```

Requires Go 1.26+. **Zero dependencies** — stdlib only, so any tool can adopt it without pulling in a framework.

---

## Usage

### A simple ban

```go
package main

import "github.com/larsartmann/go-policy-dsl"

var Gorm = policydsl.Ban("gorm").
    Because("ORMs hide N+1 queries; use sqlc for type-safe SQL").
    WithSeverity(policydsl.SeverityCritical).
    WithCategory(policydsl.CategoryPerformance).
    DetectVia(policydsl.ImportPattern("gorm.io/gorm")).
    Suggest(policydsl.NewReplacement(
        "github.com/yourorg/sqlc-queries",
        "type-safe, compile-time-checked SQL",
    )).
    Spec()
```

### A required companion

```go
var SamberDoRequiresAuditlog = policydsl.Ban("samber/do").
    Because("DI containers without audit trails are opaque").
    RequiresCompanion(policydsl.Companion(
        "samber-do-auditlog",
        "github.com/larsartmann/samber-do-auditlog",
        "audit trail for DI lifecycle",
    )).
    AsCompanionOnly(). // never ban; only enforce companion presence
    Spec()
```

### Version-bounded ban with transitive exclusion

```go
var maxV, _ = policydsl.ParseVersion("1.99.0")

var BannedOldGolog = policydsl.Ban("golog").
    Because("v1.x has a memory leak fixed in v2").
    VersionRange(nil, &maxV). // ban only library versions below 2.x
    GoModPatterns("github.com/foo/golog").
    ExcludeIfTransitiveFrom("bar"). // OK if bar pulls it in directly
    Spec()
```

`VersionRange(min, max *Version)` sets inclusive bounds (`nil` = unbounded on
that side); parse version strings with `ParseVersion` (which returns an error)
before the chain. An inverted range (`min > max`) is not rejected at
construction — call `PolicySpec.Validate()` to detect it. The library never
panics; it returns errors.

---

## API

### Constructors

| Function                                               | Returns         | Purpose                                                                                         |
| ------------------------------------------------------ | --------------- | ----------------------------------------------------------------------------------------------- |
| `Ban(name string) *Builder`                            | `*Builder`      | Start a banned-library policy (defaults to `SeverityCritical` + `CategorySecurity` + `ModeBan`) |
| `Companion(lib, pattern, reason string) CompanionSpec` | `CompanionSpec` | Build a required-companion spec (defaults to `SeverityModerate`)                                |
| `CompanionWithSeverity(...)`                           | `CompanionSpec` | Companion with custom severity                                                                  |
| `ImportPattern(pattern string) Detection`              | `Detection`     | Convenience for source-import detection                                                         |
| `GoModPattern(pattern string) Detection`               | `Detection`     | Convenience for go.mod-path detection                                                           |
| `NewReplacement(library, reason string) Replacement`   | `Replacement`   | Build a swap-in alternative value                                                               |
| `NewCVE(id string) (CVE, error)`                       | `CVE`           | Validated CVE identifier (`CVE-YYYY-NNNN`)                                                      |

### Builder methods

| Method                             | Effect                                                           |
| ---------------------------------- | ---------------------------------------------------------------- |
| `Because(reason)`                  | Sets the human-readable reason (required)                        |
| `WithSeverity(s)`                  | Overrides the default severity                                   |
| `WithCategory(c)`                  | Overrides the default category                                   |
| `DetectVia(d)`                     | Sets the full `Detection` (import + go.mod + exclusions)         |
| `ImportPatterns(p...)`             | Adds source-import patterns                                      |
| `GoModPatterns(p...)`              | Adds go.mod-path patterns                                        |
| `ExcludeIfContains(p...)`          | Suppressions: if string appears, no violation                    |
| `RequireIfContains(p...)`          | Content gate: policy only fires when at least one string appears |
| `ExcludeIfTransitiveFrom(libs...)` | Parent libs that justify transitive presence                     |
| `WithDescription(desc)`            | Detailed description for reporting                               |
| `Suggest(r)`                       | Adds a recommended replacement (full `Replacement`)              |
| `SuggestExplicit(r)`               | Adds a replacement WITHOUT deriving `Description`                |
| `WithAlternatives(alts...)`        | Sets `[]Replacement` alternatives directly (set)                 |
| `WithCVEs(cves...)`                | Tags with validated `CVE` values                                 |
| `VersionRange(min, max)`           | Inclusive library version constraints (not Go version)           |
| `RequiresCompanion(c)`             | Adds a required companion spec                                   |
| `AsCompanionOnly()`                | Sets `Mode = ModeCompanionOnly` (never ban; enforce companions)  |
| `Spec()`                           | Returns the finished `PolicySpec`                                |

### Types

| Type            | Purpose                                                                                                                                                                  |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `Severity`      | `SeverityCritical` / `SeverityHigh` / `SeverityModerate` / `SeverityLow` / `SeverityInfo`                                                                                |
| `Category`      | `CategorySecurity` / `CategoryPerformance` / `CategoryMaintainability` / `CategoryCorrectness` / `CategoryLicensing` / `CategoryCompatibility` / `CategoryConfiguration` |
| `Mode`          | `ModeBan` (default: emit ban + enforce companions) / `ModeCompanionOnly` (suppress ban, enforce only companions)                                                         |
| `Detection`     | How a violation is found: `ImportPatterns`, `GoModPatterns`, `ExcludeIfContains`, `ExcludeIfTransitiveFrom`, `RequireIfContains`                                         |
| `Replacement`   | Recommended swap-in: `Library`, `Reason`                                                                                                                                 |
| `CompanionSpec` | Required companion library                                                                                                                                               |
| `CVE`           | Validated `CVE-YYYY-NNNN` identifier (branded `string`)                                                                                                                  |
| `Version`       | Parsed semver-lite `{Major, Minor, Patch}` for inclusive version bounds                                                                                                  |
| `PolicySpec`    | The finished declarative policy value                                                                                                                                    |

---

## Design notes

- **Stdlib only.** The DSL depends on nothing outside the standard library, so any tool — a CLI, an LSP server, a CI check, a golangci-lint plugin — can adopt it without coupling. `go.mod` has zero `require` entries (the test suite is stdlib-only too, by choice).
- **Values, not configuration files.** Policies live in your source tree as Go vars, getting full IDE refactor support, type checking, and grep-ability. YAML is supported only at consumer boundaries (e.g. `library-policy` emits YAML for backward compat).
- **The `Severity` type is a `string`, not `finding.Severity`.** This keeps the DSL dependency-free. Consumers bridge the two at the boundary:

  ```go
  import "yourorg/finding"

  func toFindingSeverity(s policydsl.Severity) finding.Severity {
      switch s {
      case policydsl.SeverityCritical: return finding.SeverityCritical
      case policydsl.SeverityHigh:     return finding.SeverityHigh
      case policydsl.SeverityModerate: return finding.SeverityModerate
      case policydsl.SeverityLow:      return finding.SeverityLow
      default:                         return finding.SeverityInfo
      }
  }
  ```

- **No execution model.** The DSL declares _what_ a policy is; the consumer (`library-policy`, your CI check, etc.) decides _how_ to detect and report violations. In particular, the matching semantics for `ImportPatterns` / `GoModPatterns` (literal substring, glob, regex?) are defined by the consumer, NOT by the DSL — the patterns are opaque `[]string` so each tool can pick its matcher.
- **Validation is split.** `Spec()` performs no validation (returns exactly what was built). Structural validation lives in the separate `PolicySpec.Validate() *InvertedVersionRangeError` (a concrete typed error carrying the offending bounds; matched by sentinel via `errors.Is(err, ErrInvertedVersionRange)`). Domain rules like non-empty `Reason` remain the consumer's job.

## Surprising behaviours (intentional, pinned by tests)

- **`Suggest(r Replacement)` derives `Description` and appends the full replacement.** It appends the **whole** `Replacement` (both `Library` and `Reason`) to `Alternatives` and, when `Description` is empty, sets it to `"Replace with <library>: <reason>"`. An explicit `Description` (set before or after) is never overwritten. `Alternatives` is typed `[]Replacement` (no information loss). Use `SuggestExplicit(r)` to append without deriving `Description`.
- **`Ban(name)` defaults to `SeverityCritical` + `CategorySecurity` + `ModeBan`.** Override with `WithSeverity` / `WithCategory` for non-security concerns.
- **`WithCVEs(...CVE)` takes validated `CVE` values** (built via `NewCVE`); a free-form `[]string` cannot reach the spec.
- **`VersionRange(min, max *Version)` never panics.** An inverted range (`min > max`) is detectable via `PolicySpec.Validate()` (returns `*InvertedVersionRangeError`); the library returns errors rather than panicking.
- **`AsCompanionOnly()` sets `Mode = ModeCompanionOnly`** — the policy never emits a ban finding; it only enforces that declared companions are present. (`Mode` is the typed enforcement enum replacing the former dishonest `CompanionOnly bool`.)

The full `Builder` method convention (`With-` = set/replace vs bare = append) is documented in [`docs/DOMAIN_LANGUAGE.md`](docs/DOMAIN_LANGUAGE.md).

---

## Consumers

- [`library-policy`](https://github.com/LarsArtmann/library-policy) — primary consumer (governance CLI + server + golangci plugin). Imports `github.com/larsartmann/go-policy-dsl` v0.2.0; `domain/policy/spec.go` re-exports the SDK types via aliases, and `policies/policies.go` declares all bans via the fluent `Builder` API.

Planned:

- `go-linter-sdk` — may use `go-policy-dsl` as its rule-declaration language.

## Status

Early but shipped. The `Ban` / `Builder` / `Spec` core, the typed `Version` domain, the companion API, and `RequireIfContains` content gates are implemented and tested. One consumer shipped — `library-policy` migrated to the SDK (v0.2.0). The API is **not yet stable** (pre-1.0; see `CHANGELOG.md` for breaking changes). See `ROADMAP.md` for the path to `v1.0.0`.

## License

MIT — see [LarsArtmann/template-LICENSE](https://github.com/LarsArtmann/template-LICENSE).
