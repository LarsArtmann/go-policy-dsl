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
var BannedOldGolog = policydsl.Ban("golog").
    Because("v1.x has a memory leak fixed in v2").
    VersionRangeStrings("", "1.99.0"). // ban only library versions below 2.x
    GoModPatterns("github.com/foo/golog").
    ExcludeIfTransitiveFrom("bar"). // OK if bar pulls it in directly
    Spec()
```

`VersionRangeStrings(min, max)` parses both bounds (empty = unbounded). The typed
form `VersionRange(*Version, *Version)` rejects an inverted range (`min > max`)
at construction — the footgun the old string API allowed.

---

## API

### Constructors

| Function                                               | Returns         | Purpose                                                                             |
| ------------------------------------------------------ | --------------- | ----------------------------------------------------------------------------------- |
| `Ban(name string) *Builder`                            | `*Builder`      | Start a banned-library policy (defaults to `SeverityCritical` + `CategorySecurity`) |
| `Companion(lib, pattern, reason string) CompanionSpec` | `CompanionSpec` | Build a required-companion spec (defaults to `SeverityModerate`)                    |
| `CompanionWithSeverity(...)`                           | `CompanionSpec` | Companion with custom severity                                                      |
| `ImportPattern(pattern string) Detection`              | `Detection`     | Convenience for source-import detection                                             |
| `GoModPattern(pattern string) Detection`               | `Detection`     | Convenience for go.mod-path detection                                               |
| `NewReplacement(library, reason string) Replacement`   | `Replacement`   | Build a swap-in alternative value                                                   |

### Builder methods

| Method                             | Effect                                                   |
| ---------------------------------- | -------------------------------------------------------- |
| `Because(reason)`                  | Sets the human-readable reason (required)                |
| `WithSeverity(s)`                  | Overrides the default severity                           |
| `WithCategory(c)`                  | Overrides the default category                           |
| `DetectVia(d)`                     | Sets the full `Detection` (import + go.mod + exclusions) |
| `ImportPatterns(p...)`             | Adds source-import patterns                              |
| `GoModPatterns(p...)`              | Adds go.mod-path patterns                                |
| `ExcludeIfContains(p...)`          | Suppressions: if string appears, no violation            |
| `ExcludeIfTransitiveFrom(libs...)` | Parent libs that justify transitive presence             |
| `WithDescription(desc)`            | Detailed description for reporting                       |
| `Suggest(r)`                       | Adds a recommended replacement                           |
| `WithAlternatives(alts...)`        | Sets alternative libraries directly                      |
| `WithCVEs(cves...)`                | Tags with related CVE IDs                                |
| `VersionRange(min, max)`           | Inclusive library version constraints (not Go version)   |
| `RequiresCompanion(c)`             | Adds a required companion spec                           |
| `AsCompanionOnly()`                | Never ban; only enforce companions                       |
| `Spec()`                           | Returns the finished `PolicySpec`                        |

### Types

| Type            | Purpose                                                                                                                                                                  |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `Severity`      | `SeverityCritical` / `SeverityHigh` / `SeverityModerate` / `SeverityLow` / `SeverityInfo`                                                                                |
| `Category`      | `CategorySecurity` / `CategoryPerformance` / `CategoryMaintainability` / `CategoryCorrectness` / `CategoryLicensing` / `CategoryCompatibility` / `CategoryConfiguration` |
| `Detection`     | How a violation is found: `ImportPatterns`, `GoModPatterns`, `ExcludeIfContains`, `ExcludeIfTransitiveFrom`                                                              |
| `Replacement`   | Recommended swap-in: `Library`, `Reason`                                                                                                                                 |
| `CompanionSpec` | Required companion library                                                                                                                                               |
| `PolicySpec`    | The finished declarative policy value                                                                                                                                    |

---

## Design notes

- **Stdlib only.** The DSL depends on nothing outside the standard library, so any tool — a CLI, an LSP server, a CI check, a golangci-lint plugin — can adopt it without coupling.
- **Values, not configuration files.** Policies live in your source tree as Go vars, getting full IDE refactor support, type checking, and grep-ability. YAML is supported only at consumer boundaries (e.g. `library-policy` emits YAML for backward compat).
- **The `Severity` type is a `string`, not `finding.Severity`.** This keeps the DSL dependency-free. Consumers bridge the two at the boundary.
- **No execution model.** The DSL declares _what_ a policy is; the consumer (`library-policy`, your CI check, etc.) decides _how_ to detect and report violations.

---

## Consumers

- [`library-policy`](https://github.com/LarsArtmann/library-policy) — primary consumer (governance CLI + server + golangci plugin). Its `domain/policy/spec.go` is the migration target.

Planned:

- `go-linter-sdk` — may use `go-policy-dsl` as its rule-declaration language.

## Status

Early. API is stable for the `Ban`/`Builder`/`Spec` core; the companion API may evolve. Zero consumers yet — `library-policy` migration is the first adoption target.

## License

MIT — see [LarsArtmann/template-LICENSE](https://github.com/LarsArtmann/template-LICENSE).
