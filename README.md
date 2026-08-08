# go-policy-dsl

[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/go-policy-dsl.svg)](https://pkg.go.dev/github.com/larsartmann/go-policy-dsl)
[![CI](https://github.com/larsartmann/go-policy-dsl/actions/workflows/ci.yml/badge.svg)](https://github.com/larsartmann/go-policy-dsl/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/larsartmann/go-policy-dsl)](https://goreportcard.com/report/github.com/larsartmann/go-policy-dsl)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Fluent, compile-time-checked Go DSL for declaring library governance policies — bans, requirements, companions. Pure Go values, no YAML, no codegen, no runtime parsing.

**Make impossible policy states unrepresentable.**

---

## The Problem

Library governance policies — "don't use gorm, use sqlc", "if you depend on samber/do you must also depend on samber-do-auditlog", "this library is banned below v2" — are typically declared in YAML and parsed at runtime. That means typos surface in CI, not in your editor; refactor support is zero; and there is no compile-time guarantee that every policy specifies the fields it needs.

| YAML policies                                                | go-policy-dsl                                                           |
| ------------------------------------------------------------ | ----------------------------------------------------------------------- |
| Typos surface at runtime (CI or worse)                       | Typos surface at compile time (your editor)                             |
| Stringly-typed fields, no validation                         | Typed fields — `CVE`, `Version`, `Mode`, `Severity` are branded types   |
| Refactor support is zero (it's just text)                    | Full IDE rename, go-to-definition, find-references                      |
| Every consumer re-implements parsing                         | Every consumer imports the same typed `PolicySpec`                      |
| Inverted version range `"2.0.0" → "1.0.0"` is silently valid | `PolicySpec.Validate()` returns a concrete `*InvertedVersionRangeError` |

`go-policy-dsl` makes policies plain Go values built via a fluent API. Every field is typed, every chain is checked by the compiler, and the result is a `PolicySpec` struct that any consumer — a CLI, a golangci-lint plugin, an LSP server, a CI check — can read.

---

## Installation

```bash
go get github.com/larsartmann/go-policy-dsl
```

Requires Go 1.26+. **Zero dependencies** — stdlib only, so any tool can adopt it without pulling in a framework.

---

## Quick Start

### Ban a library

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

### Require a companion

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

### Content-gated ban

```go
var NetHTTPServerOnly = policydsl.Ban("net/http").
    Because("server-side usage requires audit middleware").
    RequireIfContains("http.ServeMux", "http.ResponseWriter").
    Spec()
```

`RequireIfContains` is a content gate — when non-empty, the policy only fires if at least one declared string appears in the matched file content. The inverse of `ExcludeIfContains` (which suppresses when found). Use case: `net/http` is both a client and server package; `RequireIfContains` lets you declare "only fire for server code".

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
| `NewVersion(major, minor, patch int) Version`          | `Version`       | Parsed semver-lite value                                                                        |
| `ParseVersion(s string) (Version, error)`              | `Version`       | Parse a version string (returns error, never panics)                                            |

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

Method-naming convention: `With<X>` = set/replace the field wholesale; bare `<X>` = append to a slice field; `As<X>` = set a mode flag; `Spec()` = terminate the chain.

### Types

| Type            | Purpose                                                                                                                                                              |
| --------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Severity`      | `SeverityCritical` / `SeverityHigh` / `SeverityModerate` / `SeverityLow` / `SeverityInfo`                                                                            |
| `Category`      | `CategorySecurity` / `CategoryPerformance` / `CategoryMaintenance` / `CategoryCorrectness` / `CategoryLicensing` / `CategoryCompatibility` / `CategoryConfiguration` |
| `Mode`          | `ModeBan` (default: emit ban + enforce companions) / `ModeCompanionOnly` (suppress ban, enforce only companions)                                                     |
| `Detection`     | How a violation is found: `ImportPatterns`, `GoModPatterns`, `ExcludeIfContains`, `ExcludeIfTransitiveFrom`, `RequireIfContains`                                     |
| `Replacement`   | Recommended swap-in: `Library`, `Reason`                                                                                                                             |
| `CompanionSpec` | Required companion library                                                                                                                                           |
| `CVE`           | Validated `CVE-YYYY-NNNN` identifier (branded `string`)                                                                                                              |
| `Version`       | Parsed semver-lite `{Major, Minor, Patch}` for inclusive version bounds                                                                                              |
| `PolicySpec`    | The finished declarative policy value                                                                                                                                |

---

## Design Decisions

- **Stdlib only.** The DSL depends on nothing outside the standard library, so any tool — a CLI, an LSP server, a CI check, a golangci-lint plugin — can adopt it without coupling. `go.mod` has zero `require` entries (the test suite is stdlib-only too, by choice).
- **Values, not configuration files.** Policies live in your source tree as Go vars, getting full IDE refactor support, type checking, and grep-ability. The DSL declares _what_ a policy is; the consumer decides _how_ to detect and report violations.
- **`Severity` is a `string`, not your consumer's type.** This keeps the DSL dependency-free. Consumers bridge at the boundary:

  ```go
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

- **No execution model.** The matching semantics for `ImportPatterns` / `GoModPatterns` (literal substring, glob, regex?) are defined by the consumer, NOT by the DSL — patterns are opaque `[]string` so each tool picks its matcher.
- **No struct tags.** `PolicySpec`, `Mode`, `CVE`, `Replacement`, and `Version` have no `json`/`yaml` tags. The DSL declares values, not config. Consumers that need to marshal bridge at the boundary with an explicit field-map.
- **Validation is split.** `Spec()` performs no validation (returns exactly what was built). Structural validation lives in the separate `PolicySpec.Validate()` — returns a concrete `*InvertedVersionRangeError` (carrying the offending bounds; matched by sentinel via `errors.Is(err, ErrInvertedVersionRange)`). Domain rules like non-empty `Reason` remain the consumer's job.
- **Panic-free.** Every error condition — invalid CVE, invalid version, malformed version string — is returned, never panicked. There are deliberately no `Must*` constructors. The panic-free contract is asserted by an in-repo regression test (`TestNoPanicsInNonTestSource`).
- **Deny-by-default `Mode`.** The only value that suppresses the ban finding is `ModeCompanionOnly`. Every other value — the zero-value empty string, `ModeBan`, and any unknown/garbage string — is ban-active. A typo in `Mode` can never silently disable enforcement.

### Surprising behaviours (intentional, pinned by tests)

- **`Suggest(r Replacement)` derives `Description` and appends the full replacement.** It appends the **whole** `Replacement` (both `Library` and `Reason`) to `Alternatives` and, when `Description` is empty, sets it to `"Replace with <library>: <reason>"`. An explicit `Description` is never overwritten. `Alternatives` is typed `[]Replacement` (no information loss). Use `SuggestExplicit(r)` to append without deriving `Description`.
- **`Ban(name)` defaults to `SeverityCritical` + `CategorySecurity` + `ModeBan`.** Override with `WithSeverity` / `WithCategory` for non-security concerns.

The full `Builder` method convention (`With-` = set/replace vs bare = append) is documented in [`docs/DOMAIN_LANGUAGE.md`](docs/DOMAIN_LANGUAGE.md).

---

## When NOT to use

- **You need runtime-configurable policies.** This DSL is compile-time Go values. If your users need to edit policies without recompiling, use a YAML/JSON-based tool.
- **You need serialization.** `PolicySpec` has no struct tags by design. If you need JSON/YAML output, bridge at the boundary with an explicit field-map.
- **You need a detection engine.** The DSL declares _what_ a policy is, not _how_ to detect violations. You need a consumer (a CLI, a linter, an LSP server) that reads `PolicySpec` and implements the matching.

---

## Status

Pre-1.0. The API is **not yet stable** — breaking changes land in `0.x` bumps and are always marked `**BREAKING**` in [`CHANGELOG.md`](CHANGELOG.md). See [`ROADMAP.md`](ROADMAP.md) for the path to v1.0.0.

| Version    | Summary                                                                      |
| ---------- | ---------------------------------------------------------------------------- |
| **v0.3.0** | `RequireIfContains` content gate (additive, non-breaking)                    |
| v0.2.0     | Panic-free refactor + typed domain (`Mode`, `Version`, `CVE`, `Replacement`) |
| v0.1.0     | Initial tagged release — fluent builder, typed enums, version bounds         |

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Run the full quality gate locally:

```bash
go test ./...            # tests (race-clean)
golangci-lint run ./...  # lint (0 issues)
golangci-lint fmt ./...  # format check / apply
```

## License

[MIT](LICENSE)
