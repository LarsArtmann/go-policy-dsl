# go-policy-dsl

Fluent, compile-time-checked Go DSL for declaring library governance policies — bans, requirements, companions. Pure Go values, no YAML, no codegen, no runtime parsing. **Public SDK library, stdlib-only, zero dependencies.**

**Status:** tests pass, `golangci-lint run ./...` clean (0 issues), single root package.

## Quick Start

```bash
go test ./...                 # all tests
golangci-lint run ./...       # lint (uses .golangci.yml, v2 format)
golangci-lint fmt ./...       # apply formatters (gci, goimports, gofumpt)
go build ./...                # build check
erraudit ./...                # hierarchical error analysis (0 violations — the library is panic-free)
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

## Architecture Decision: Panic-Free (Errors Returned, Never Panicked)

The library **never panics**. Every error condition — invalid CVE, invalid/negative version, malformed version string, inverted version range — is **returned** as a value, never panicked. There are deliberately no `Must*` constructors (no `MustCVE`, `MustNewVersion`, `MustParseVersion`) and no `VersionRangeStrings` convenience. Rationale: the fluent `Builder` chain cannot propagate a returned error mid-chain, so any method that can fail lives _outside_ the chain as a free function returning `(T, error)` (`NewCVE`, `NewVersion`, `ParseVersion`). The one structural invariant an assembled `PolicySpec` can still violate — an inverted version range — is checked by `PolicySpec.Validate()`, the **single source of truth** for that invariant (`*InvertedVersionRangeError`). `erraudit ./...` reports 0 violations. Do NOT re-introduce `Must*` functions or construction-time panics; the panic-free contract is now a documented guarantee, asserted by tests.

## Architecture Decision: No Struct Tags (Pure Values, Not Config)

`PolicySpec`, `Mode`, `CVE`, `Replacement`, and `Version` have **no** `json`/`yaml` struct tags. This is deliberate: the DSL declares _values_, not configuration. Adding serialization tags would couple the public API to a specific wire format (JSON field names, YAML key names) before any consumer needs it. Consumers that need to marshal/unmarshal (`library-policy` emits YAML) bridge at the boundary with an explicit field-map — this is the "values, not config" philosophy. Revisit only if multiple consumers independently request direct serialization AND agree on the field names.

## Architecture Decision: Mode Is Deny-By-Default

The `Mode` typed enum has one design rule: **the ONLY value that suppresses the ban finding is `ModeCompanionOnly`**. Every other value — the zero-value empty string, `ModeBan`, and any unknown/garbage string — is ban-active. This is a deny-by-default security property: a typo in a Mode can never silently disable enforcement. `Validate()` does NOT reject unknown Mode values (unlike inverted version ranges); it treats Mode validation as a domain rule left to the consumer. Pinned by `TestMode_DenyByDefaultContract`.

## Surprising Behaviors

- **`Suggest(r Replacement)` has a side effect beyond adding the alternative.** It appends the **full** `Replacement` (Library AND Reason) to `Alternatives` AND, if `Description` is empty, sets `Description` to `"Replace with <library>: <reason>"`. Callers that set `Description` explicitly after `Suggest` will overwrite this. **`SuggestExplicit(r)`** is the no-magic variant: it appends to `Alternatives` but never derives `Description`.
- **`Alternatives` is `[]Replacement`, not `[]string`.** Each entry keeps its `Reason` (no information loss). `WithAlternatives(...Replacement)` replaces the slice wholesale; `Suggest`/`SuggestExplicit` append.
- **`WithCVEs(...CVE)` takes validated `CVE` values.** `CVE` is a branded `string` constructed via `NewCVE(id)` (validates `CVE-YYYY-NNNN`, returns an error). A free-form `[]string` cannot reach the spec.
- **`Ban(name)` defaults to `SeverityCritical` + `CategorySecurity` + `ModeBan`.** Override with `WithSeverity` / `WithCategory` for non-security concerns, or the ban is over-rated.
- **`Companion(...)` defaults to `SeverityModerate`** (use `CompanionWithSeverity` to override).
- **`AsCompanionOnly()` sets `Mode = ModeCompanionOnly`** (the policy never emits a ban finding; it only enforces that declared companions are present). The field is the typed `Mode` enum (`ModeBan` / `ModeCompanionOnly`), NOT the former dishonest `CompanionOnly bool`. **Deny-by-default:** the zero-value `Mode` (`""`), `ModeBan`, and any unknown Mode string are ALL treated as ban-active. Only `ModeCompanionOnly` suppresses the ban. Use this for "this library is fine, but if you use it you must also use X".
- **`ExcludeIfTransitiveFrom(libs...)` is the false-positive guard** for indirect dependencies: if a listed parent library directly pulls in the banned lib, no violation fires.
- **`NewReplacement(library, reason)` is the constructor**; `Replacement` is the type. The package doc example uses `NewReplacement(...)`. (An earlier doc example wrongly called `Replacement(...)` — it would not compile.)
- **`VersionRange(minVer, maxVer *Version)` is inclusive on both ends** and constrains the version of the _library_ targeted by the policy (NOT the Go toolchain version). `nil` on either side means unconstrained. **Never panics on an inverted range** — detect `min > max` via `PolicySpec.Validate()`. Parse version strings with `ParseVersion` (returns an error) before the chain; there is no string-convenience builder method. Renamed from `GoVersionRange` (2026-07 review): the old name lied — it never constrained Go itself, only the library version.
- **`Spec()` performs no validation** — it returns exactly what was built. Structural validation lives in `PolicySpec.Validate()`, which returns a concrete `*InvertedVersionRangeError` (carrying `Min`/`Max` pointers) rather than a generic `error`. Callers using `:=` get type-safe access to the offending bounds directly; `errors.Is(err, ErrInvertedVersionRange)` still works via the type's `Is` method. Currently checks only the version-range invariant; domain rules like non-empty `Reason` remain the consumer's job.

## Conventions

- Fluent `Builder` returned by every chainable method; `Spec()` terminates the chain and returns the immutable `PolicySpec`.
- Constructor helpers (`ImportPattern`, `GoModPattern`, `Companion`, `NewReplacement`, `NewVersion`, `ParseVersion`, `NewCVE`) are package-level functions, not methods.
- File layout: `policy.go` = public types/constants/`Validate`/`Mode`/`InvertedVersionRangeError`; `builder.go` = the fluent `Builder` + constructor helpers; `version.go` = the typed `Version` domain (stdlib-only semver-lite, uses `cmp.Compare` not hand-rolled sign logic); `cve.go` = the branded `CVE` domain (`CVE-YYYY-NNNN` validation). Tests beside the code (`policy_test.go`, `version_test.go`, `cve_test.go`, `builder_behavior_test.go`, `builder_fuzz_test.go`, `zero_value_test.go`, `example_test.go`, `testhelpers_test.go` = shared test utilities, `panic_free_test.go` = regression guard for the panic-free contract).

## Consumers

- `library-policy` (github.com/LarsArtmann/library-policy) — primary consumer; its `domain/policy/spec.go` is the migration target.
- `go-linter-sdk` — may adopt this as its rule-declaration language.

Zero consumers currently shipped; `library-policy` migration is the first adoption target.

## License

MIT — matches all sibling LarsArtmann SDK libraries. See `LICENSE`.
