# go-policy-dsl

Fluent, compile-time-checked Go DSL for declaring library governance policies — bans, requirements, companions. Pure Go values, no YAML, no codegen, no runtime parsing. **Public SDK library, stdlib-only, zero dependencies.**

**Status:** tests pass, `golangci-lint run ./...` clean (0 issues), single root package.

## Quick Start

```bash
go test ./...                 # all tests
golangci-lint run ./...       # lint (uses .golangci.yml, v2 format)
golangci-lint fmt ./...       # apply formatters (gci, goimports, gofumpt)
go build ./...                # build check
```

No `flake.nix` — this is a tiny stdlib-only library; buildflow handles CI. Module: `github.com/larsartmann/go-policy-dsl`, Go 1.26.5.

## Architecture Decision: Package Lives at Module Root

The `policydsl` package lives at the **module root** (`builder.go`, `policy.go`), NOT under `pkg/` or `internal/`. This is intentional and matches every sibling LarsArtmann public SDK library (`go-error-family`, `go-output`, `go-atomic-write`, `go-linter-sdk`, `samber-do-auditlog`).

Rationale: the module path ends in the package name, so root placement gives the cleanest import — `github.com/larsartmann/go-policy-dsl` → `policydsl.Ban(...)`. Moving to `pkg/policydsl/` would make the import path redundantly repeat `policydsl`; moving to `internal/` would make the library unimportable.

Consequence: `go-structure-linter` reports `root-package-files` (ERROR) and `internal-directory` (WARNING). These two rules are **known false-positives for public SDK libraries** and are intentionally accepted across the ecosystem. The linter's config schema (`exclude_patterns` / `enabled_rules`) cannot disable enabled-by-default rules, so the noise cannot be suppressed via config — only via the `--exclude` CLI flag, which is not wired up here. Do NOT "fix" these by moving the code; that breaks the public API.

## Architecture Decision: Stdlib-Only, String-Typed Severities

- `Severity` is a `string` alias, NOT `finding.Severity` or any consumer's type. This keeps the DSL dependency-free so any tool (CLI, LSP, golangci plugin, CI check) can adopt it. **Consumers bridge `Severity` → their own severity type at the boundary.**
- `Category` is likewise a `string` alias for the same reason.
- No execution model: the DSL declares what a policy _is_; the consumer decides how to detect and report violations. `Spec()` performs no validation — it returns exactly what was built.

## Surprising Behaviors

- **`Suggest(r Replacement)` has a side effect beyond adding the alternative.** It appends `r.Library` to `Alternatives` AND, if `Description` is empty, sets `Description` to `"Replace with <library>: <reason>"`. Callers that set `Description` explicitly after `Suggest` will overwrite this.
- **`Ban(name)` defaults to `SeverityCritical` + `CategorySecurity`.** Override with `WithSeverity` / `WithCategory` for non-security concerns, or the ban is over-rated.
- **`Companion(...)` defaults to `SeverityModerate`** (use `CompanionWithSeverity` to override).
- **`AsCompanionOnly()` suppresses the ban entirely** — the policy never emits a ban finding; it only enforces that declared companions are present. Use this for "this library is fine, but if you use it you must also use X".
- **`ExcludeIfTransitiveFrom(libs...)` is the false-positive guard** for indirect dependencies: if a listed parent library directly pulls in the banned lib, no violation fires.
- **`NewReplacement(library, reason)` is the constructor**; `Replacement` is the type. The package doc example uses `NewReplacement(...)`. (An earlier doc example wrongly called `Replacement(...)` — it would not compile.)
- **`VersionRange(minVer, maxVer *Version)` is inclusive on both ends** and constrains the version of the _library_ targeted by the policy (NOT the Go toolchain version). `nil` on either side means unconstrained. **Panics if both bounds are non-nil and min > max** (nonsensical inverted range — fail-fast at package init). `VersionRangeStrings(min, max string)` is the string convenience (empty = unbounded; parses via `MustParseVersion`, panics on parse error or inversion). Renamed from `GoVersionRange` (2026-07 review): the old name lied — it never constrained Go itself, only the library version. Typed `*Version` + `VersionRangeStrings` (2026-07-26 review): previously `string` fields where `("2.0.0","1.0.0")` was silently representable; the typed domain rejects inversion at construction.
- **`Spec()` performs no validation** — it returns exactly what was built. Structural validation lives in the separate `PolicySpec.Validate() error` method (currently checks only the version-range invariant; domain rules like non-empty `Reason` remain the consumer's job).

## Conventions

- Fluent `Builder` returned by every chainable method; `Spec()` terminates the chain and returns the immutable `PolicySpec`.
- Constructor helpers (`ImportPattern`, `GoModPattern`, `Companion`, `NewReplacement`) are package-level functions, not methods.
- File layout: `policy.go` = all public types/constants; `builder.go` = the fluent `Builder` + constructor helpers. Tests beside the code.

## Consumers

- `library-policy` (github.com/LarsArtmann/library-policy) — primary consumer; its `domain/policy/spec.go` is the migration target.
- `go-linter-sdk` — may adopt this as its rule-declaration language.

Zero consumers currently shipped; `library-policy` migration is the first adoption target.

## License

MIT — matches all sibling LarsArtmann SDK libraries. See `LICENSE`.
