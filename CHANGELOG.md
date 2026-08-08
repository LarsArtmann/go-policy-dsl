# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> **Pre-1.0 policy:** the public API is not yet stable. Breaking changes land in
> any `0.x` bump and are marked `**BREAKING**`. See `ROADMAP.md` for the path to
> v1.0.0.

## [Unreleased]

## [0.3.0] - 2026-08-08

### Added

- **`Detection.RequireIfContains []string`** — content gate controlling WHEN a
  policy fires. When non-empty, the policy only activates if at least one
  declared string appears in the matched file content. The inverse of
  `ExcludeIfContains` (which suppresses when found). Use case: `net/http` is
  both a client and server package; `RequireIfContains` lets a policy declare
  "only fire for server code" via `http.ServeMux`, `http.ResponseWriter`, etc.
- **`Builder.RequireIfContains(patterns ...string)`** — fluent append helper
  matching the convention of `ExcludeIfContains` / `ExcludeIfTransitiveFrom`.

## [0.2.0] - 2026-07-26

### Changed (breaking)

- **The library is now panic-free.** Every error condition is returned, never
  panicked. Removed the entire `Must*` family:
  - `MustCVE(id)` → use `NewCVE(id)` (returns `(CVE, error)`)
  - `MustNewVersion(...)` → use `NewVersion(...)` (returns `(Version, error)`)
  - `MustParseVersion(s)` → use `ParseVersion(s)` (returns `(Version, error)`)
  - `VersionRangeStrings(min, max string)` → parse each bound with `ParseVersion`
    and pass `*Version` values to `VersionRange`. (The method hid parse errors
    behind a panic; a panic-free fluent chain cannot surface a parse error
    mid-chain, so the convenience was removed rather than silently dropped or
    re-panicked.)
  - `VersionRange(min, max *Version)` no longer panics on inverted ranges.
    Detect `min > max` via `PolicySpec.Validate()`, which returns
    `*InvertedVersionRangeError`.
- **Renamed** `GoVersionRange` → `VersionRange`, `GoVersionMin` → `VersionMin`,
  `GoVersionMax` → `VersionMax`. The old names lied — they constrain the
  _library_ version, not the Go toolchain.
- **Retyped** `PolicySpec.VersionMin` / `VersionMax` from `string` to `*Version`,
  and `Builder.VersionRange` from `(string, string)` to `(*Version, *Version)`.
  `nil` on either side means unbounded (was the empty string). Migrate by
  parsing each bound with `ParseVersion` before the chain.
- **Replaced** `PolicySpec.CompanionOnly bool` with a typed `Mode` enum
  (`ModeBan` / `ModeCompanionOnly`). `AsCompanionOnly()` sets
  `ModeCompanionOnly`; consumers read `spec.Mode != ModeCompanionOnly` for
  ban-active. The zero-value `Mode` (`""`) is treated as ban-active.
- **Retyped** `PolicySpec.Alternatives` from `[]string` to `[]Replacement`, and
  `WithAlternatives(...string)` to `WithAlternatives(...Replacement)`. `Suggest`
  appends the **full** `Replacement` (both `Library` and `Reason`) instead of
  discarding the reason. Access `alt.Library` where you read a bare name before.
- **Retyped** `PolicySpec.CVEs` from `[]string` to `[]CVE`, and
  `WithCVEs(...string)` to `WithCVEs(...CVE)`. Build CVEs via `NewCVE(id)`
  (validates `CVE-YYYY-NNNN`); invalid identifiers can no longer reach a spec.
- `Validate()` now returns a concrete `*InvertedVersionRangeError` (with
  `Min`/`Max` fields) instead of a generic wrapped `error`.
  `errors.Is(err, ErrInvertedVersionRange)` still works via the type's `Is`
  method.

### Added

- **`Version` type** — stdlib-only parsed semver-lite (`{Major, Minor, Patch}`)
  with `NewVersion`, `ParseVersion`, and `String` / `Compare` / `Before` /
  `After` / `Equal`. Enables compile-time-checked version bounds with no semver
  dependency.
- **`PolicySpec.Validate()`** — structural validation returning
  `*InvertedVersionRangeError`. Checks the version-range invariant; domain rules
  remain the consumer's job. `Spec()` stays validation-free.
- **`Mode` typed enum** (`ModeBan` / `ModeCompanionOnly`) — deny-by-default:
  the only value that suppresses the ban finding is `ModeCompanionOnly`; every
  other value (empty string, `ModeBan`, unknown strings) is ban-active.
- **`CVE` branded type** with `NewCVE(id)` validating `CVE-YYYY-NNNN`.
- **`SuggestExplicit(r Replacement)`** — the no-magic counterpart to `Suggest`:
  appends a replacement without deriving `Description`.
- **`Version.Compare`** uses stdlib `cmp.Compare` (replaces hand-rolled sign
  logic, eliminates integer overflow edge case).

### Fixed

- Package doc example now compiles: `policydsl.Replacement(...)` (a type, not
  callable) → `policydsl.NewReplacement(library, reason)`.
- LICENSE corrected from proprietary to **MIT**, matching the README and all
  sibling LarsArtmann SDK libraries.
- Removed a phantom `Require` builder referenced in docs that never existed in
  code.

### Internal

- Rewrote the test suite as an external `policydsl_test` package with
  `t.Parallel()` on every test (exercises only the exported API, race-clean).
- `erraudit ./...` reports 0 violations — panic-free contract verified.
- In-repo regression guard `TestNoPanicsInNonTestSource` parses every non-test
  `.go` file and fails if any contains a `panic(` call expression.
- GitHub Actions CI: build, vet, test -race, `golangci-lint run`,
  `golangci-lint fmt` drift check, fuzz targets (`FuzzParseVersion`,
  `FuzzBuilder_PatternsOpaque`).
- `docs/DOMAIN_LANGUAGE.md`, `FEATURES.md`, `TODO_LIST.md`, `ROADMAP.md`,
  `CODEOWNERS`, issue/PR templates, `.github/dependabot.yml`.

## [0.1.0] - 2026-07-26

### Added

- **Fluent `Builder`** (`Ban` / `Companion`) with chainable methods: `Because`,
  `WithSeverity`, `WithCategory`, `WithDescription`, `DetectVia`,
  `ImportPatterns`, `GoModPatterns`, `ExcludeIfContains`,
  `ExcludeIfTransitiveFrom`, `Suggest`, `SuggestExplicit`, `WithAlternatives`,
  `RequiresCompanion`, `WithCVEs`, `VersionRange`, `AsCompanionOnly`, and
  `Spec()` terminator. Pure Go values — no YAML, no codegen, no runtime parsing.
- **`PolicySpec`** — the immutable value returned by `Spec()`. Carries `Name`,
  `Reason`, `Severity`, `Category`, `Mode`, `Detection`, `Alternatives`,
  `Companions`, `CVEs`, `VersionMin`, `VersionMax`.
- **`Severity`** typed enum (`critical`, `high`, `moderate`, `low`, `info`,
  `recommended`, `deprecated`, `obsolete`) — `string` alias, dependency-free.
- **`Category`** typed enum (`security`, `performance`, `maintenance`,
  `deprecation`, `architecture`, `correctness`, `licensing`, `compatibility`).
- **`Mode`** typed enum (`ModeBan` / `ModeCompanionOnly`). `AsCompanionOnly()`
  sets `ModeCompanionOnly`; the zero-value `Mode` (`""`) is ban-active.
- **`Version`** type — stdlib-only semver-lite (`{Major, Minor, Patch}`) with
  `NewVersion`, `ParseVersion`, `String` / `Compare` / `Before` / `After` /
  `Equal`.
- **`VersionRangeStrings(min, max string)`** — string convenience for
  `VersionRange` (empty = unbounded; panics on parse error or inversion).
  _Removed in v0.2.0 in favour of panic-free typed `VersionRange`._
- **`PolicySpec.Validate()`** — structural validation returning
  `*InvertedVersionRangeError`.
- **`CVE`** branded type with `NewCVE(id)` / `MustCVE(id)` validating
  `CVE-YYYY-NNNN`. _`MustCVE` removed in v0.2.0._
- **`SuggestExplicit(r Replacement)`** — no-magic counterpart to `Suggest`.
- **`Replacement`** and **`CompanionSpec`** types. `Suggest` appends the full
  `Replacement` (Library + Reason) and auto-derives `Description` when empty.
- **`Detection`** struct with `ImportPatterns`, `GoModPatterns`,
  `ExcludeIfContains`, `ExcludeIfTransitiveFrom` — patterns are opaque strings
  stored verbatim; the DSL owns no matching semantics.
- Constructor helpers: `ImportPattern`, `GoModPattern`, `Companion`,
  `NewReplacement`, `NewVersion`, `ParseVersion`, `NewCVE`.
- Test suite: BDD-style behaviour specs, fuzz targets, zero-value tests,
  godoc `Example*` functions. External `policydsl_test` package with
  `t.Parallel()`.
- `docs/DOMAIN_LANGUAGE.md`, `FEATURES.md`, `TODO_LIST.md`, `ROADMAP.md`,
  `CONTRIBUTING.md`.
- GitHub Actions CI workflow, `.github/dependabot.yml`, `CODEOWNERS`, issue/PR
  templates.
- MIT `LICENSE`.
