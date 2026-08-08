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

### Added

- **`Version.Compare`** uses stdlib `cmp.Compare` (replaces hand-rolled
  `cmpSign` sign logic, eliminates integer overflow edge case).

### Internal

- In-repo regression guard `TestNoPanicsInNonTestSource` parses every non-test
  `.go` file via `go/parser` and fails if any contains a `panic(` call
  expression.
- `erraudit ./...` reports 0 violations — panic-free contract verified.
- `testhelpers_test.go` for shared test utilities.
- `TestBan_SetsModeBan` and `TestMode_DenyByDefaultContract` pin the
  deny-by-default Mode contract; `TestBuilder_VersionRange_UnboundedMin` and
  `TestBuilder_VersionRange_InvertedDeferToValidate` pin the panic-free
  `VersionRange` behaviour.
- Documented the panic-free contract, no-struct-tags policy, and Mode
  deny-by-default contract in `AGENTS.md`.
- GitHub Actions CI fuzz job running both fuzz targets for 30s each on every
  push and PR.
- `.github/dependabot.yml` for automatic GitHub Actions version bumps.
- Removed dead `errMustCommit bool` field from the `TestNewVersion` struct.

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
- GitHub Actions CI workflow, `CODEOWNERS`, issue/PR templates.
- MIT `LICENSE`.
