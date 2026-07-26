# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> **Versioning policy (pre-1.0):** the public API is not yet stable. Breaking
> changes are allowed in any `0.x` bump and are always listed under the
> **Changed** section with a `**BREAKING**` marker. The first tagged release
> will be `v0.2.0` (the module has never been tagged `v0.1.0` — the entry below
> is the original scaffolding baseline, not a published release). See
> `ROADMAP.md` for the path to `v1.0.0`.

## [Unreleased]

### Changed

- **BREAKING: the library is now panic-free.** Every error condition is
  returned, never panicked. This removes the entire `Must*` family and the
  inversion panic:
  - Removed `MustCVE(id)`, `MustNewVersion(...)`, `MustParseVersion(s)`. Use the
    error-returning `NewCVE`, `NewVersion`, `ParseVersion` instead.
  - Removed `VersionRangeStrings(min, max string)`. Parse each bound with
    `ParseVersion` and pass `*Version` values to `VersionRange`. (The method hid
    parse errors behind a panic; a panic-free fluent chain cannot surface a
    parse error mid-chain, so the convenience was removed rather than silently
    dropped or re-panicked.)
  - Removed the inversion `panic` from `VersionRange(min, max *Version)`. An
    inverted range (`min > max`) is no longer rejected at construction; detect
    it via `PolicySpec.Validate()`, which returns `*InvertedVersionRangeError` —
    now the single source of truth for the inversion invariant.
  - `erraudit ./...` now reports **0 violations** (previously 3 accepted
    `Must`-panic false positives). The panic-free contract is a documented
    guarantee in `AGENTS.md`, asserted by tests and `erraudit`.
- **BREAKING:** renamed `GoVersionRange` → `VersionRange`,
  `GoVersionMin` → `VersionMin`, `GoVersionMax` → `VersionMax`. The old names
  lied — they constrain the version of the _library_ targeted by the policy,
  never the Go toolchain version. Zero shipped consumers, so the rename is safe.
- **BREAKING:** retyped `PolicySpec.VersionMin` / `VersionMax` from `string`
  to `*Version`, and `Builder.VersionRange` from `(string, string)` to
  `(*Version, *Version)`. The old string API allowed the nonsensical inverted
  range `("2.0.0", "1.0.0")` to be silently represented; the typed domain makes
  the bounds explicit. `nil` on either side means unbounded (was the empty
  string). Migrate by parsing each bound with `ParseVersion` (returns an error)
  before the chain, then passing `*Version` values to `VersionRange`.
- **BREAKING:** replaced `PolicySpec.CompanionOnly bool` with a typed `Mode`
  enum (`ModeBan` / `ModeCompanionOnly`). The old name lied — it suppressed the
  ban, not the companion. `AsCompanionOnly()` now sets `Mode = ModeCompanionOnly`;
  consumers read `spec.Mode != ModeCompanionOnly` for ban-active. The zero-value
  `Mode` (`""`) is treated as ban-active.
- **BREAKING:** retyped `PolicySpec.Alternatives` from `[]string` to
  `[]Replacement`, and `WithAlternatives(...string)` to
  `WithAlternatives(...Replacement)`. `Suggest` now appends the **full**
  `Replacement` (both `Library` and `Reason`) instead of discarding the reason.
  This eliminates silent information loss; access `alt.Library` where you read a
  bare name before.
- **BREAKING:** retyped `PolicySpec.CVEs` from `[]string` to `[]CVE`, and
  `WithCVEs(...string)` to `WithCVEs(...CVE)`. Build CVEs via `NewCVE(id)`
  (validates `CVE-YYYY-NNNN`); invalid identifiers can no longer reach a spec.
- `Validate()` now returns a concrete `*InvertedVersionRangeError` (with
  `Min`/`Max` fields) instead of a generic wrapped `error`; `errors.Is(err,
ErrInvertedVersionRange)` still works via the type's `Is` method. Callers
  using `:=` get type-safe access to the offending bounds directly.
- Rewrote the test suite as an external `policydsl_test` package with
  `t.Parallel()` on every test, so the suite exercises only the exported API
  (the same surface real consumers use) and is race-clean.
- Replaced hand-rolled `cmpSign` with stdlib `cmp.Compare` in `Version.Compare`,
  dropping ~10 lines of sign logic and eliminating the theoretical overflow
  edge case of subtracting two ints.
- Moved shared test helper `parseVersionOrFatal` from `builder_behavior_test.go`
  to a dedicated `testhelpers_test.go` (its honest home).
- Removed dead `errMustCommit bool` field from the `TestNewVersion` struct
  (declared but never read — a latent lie in the test fixtures).
- Bumped the `golangci-lint` `go` directive `1.26.4` → `1.26.5` to match
  `go.mod`.

### Added

- `TestNoPanicsInNonTestSource` — in-repo regression guard for the panic-free
  contract. Parses every non-test `.go` file via `go/parser` and fails if any
  contains a `panic(` call expression. Catches regressions at PR time
  independent of whether `erraudit` is wired into CI.
- `TestBan_SetsModeBan` — pins that `Ban(...)` explicitly sets `Mode = ModeBan`
  (not the zero-value empty string).
- `TestMode_DenyByDefaultContract` — pins the deny-by-default Mode contract:
  the ONLY value that suppresses the ban finding is `ModeCompanionOnly`; every
  other value (empty string, `ModeBan`, unknown strings) is ban-active.
- Documented the **Mode deny-by-default contract** in AGENTS.md: a garbage Mode
  never silently disables enforcement.
- Documented the **no-struct-tags policy** in AGENTS.md: `PolicySpec`, `Mode`,
  `CVE`, `Replacement`, and `Version` deliberately have no `json`/`yaml` struct
  tags — the DSL declares values, not config. Consumers bridge at the boundary.
- GitHub Actions CI fuzz job running both fuzz targets (`FuzzParseVersion`,
  `FuzzBuilder_PatternsOpaque`) for 30s each on every push and PR.
- `.github/dependabot.yml` for automatic GitHub Actions version bumps.
- `docs/DOMAIN_LANGUAGE.md` — ubiquitous-language glossary (Ban, Companion,
  Detection, Replacement, Severity, Category, PolicySpec, Version, Mode, CVE).
- `FEATURES.md` — honest feature inventory by status.
- `TODO_LIST.md` — short/mid-term actionable tasks.
- `ROADMAP.md` — long-term direction and the path to v1.0.0.
- `Version` type — a stdlib-only parsed semver-lite value
  (`{Major, Minor, Patch}`) with `NewVersion`, `ParseVersion`, and
  `String` / `Compare` / `Before` / `After` / `Equal`. Enables compile-time-
  checked version bounds with no semver dependency.
- `PolicySpec.Validate()` — structural validation entry point returning a
  concrete `*InvertedVersionRangeError` (carrying the offending `Min`/`Max`
  bounds). Checks only the version-range invariant; domain rules remain the
  consumer's job. `Spec()` itself stays validation-free.
- `Mode` typed enum (`ModeBan` / `ModeCompanionOnly`) replacing the former
  dishonest `CompanionOnly bool` field.
- `CVE` branded type (`cve.go`) with `NewCVE(id)` validating the canonical
  MITRE form `CVE-YYYY-NNNN`, and `ErrInvalidCVE`. An unvalidated free-form
  string can no longer reach a spec.
- `SuggestExplicit(r Replacement)` builder method — the no-magic counterpart
  to `Suggest`: appends a replacement without deriving `Description`.
- Contract-locking tests for the surprising `Suggest` → `Description`
  side-effect (auto-derive when empty; preserve when explicit; append-only on
  repeated calls).
- Table-driven test covering the four append-style detection helpers
  (`ImportPatterns`, `GoModPatterns`, `ExcludeIfContains`,
  `ExcludeIfTransitiveFrom`).
- Comprehensive `Version` type tests: `NewVersion`, `ParseVersion` (16 cases
  including sign rejection), `Compare`, relation helpers, round-trip.
- `VersionRange` typed + inversion tests, `Validate` inversion tests (including
  `TestPolicySpec_Validate_InvertedRangeErrorCarriesBounds`).
- BDD-style behaviour suite (stdlib, not Ginkgo) covering the fluent chain
  from the user perspective: `TestBehavior_BuildingABan`,
  `TestBehavior_SuggestingAReplacement`, `TestBehavior_RequiringACompanion`,
  `TestBehavior_BoundingByLibraryVersion`,
  `TestBehavior_VersionRangeInversionIsCaughtByValidate`.
- `depguard` exclusion for `_test.go` so the external test package
  (`policydsl_test`) compiles cleanly.
- `golangci-lint fmt` formatters configured (`gci`, `goimports`, `gofumpt`,
  `golines` at 120 cols).
- Zero-value tests pinning the meaning of an unbuilt `PolicySpec`,
  `Detection`, `CompanionSpec`, `Replacement`, and `Version` (the public
  contract consumers construct against literally).
- godoc `Example*` functions (`ExampleBan`, `ExampleBan_versionRange`,
  `ExampleCompanion`, `ExampleVersion`, `ExamplePolicySpec_Validate`,
  `ExampleCVE`) so pkg.go.dev renders runnable, compile-verified examples.
- `FuzzParseVersion` — fuzz target asserting the parser never panics on
  arbitrary input and every parsed version round-trips through `String`.
- `FuzzBuilder_PatternsOpaque` — fuzz target pinning the contract that
  detection patterns are opaque strings (stored verbatim, never interpreted);
  the DSL owns no matching semantics.
- CVE validation tests (`TestNewCVE_Valid`, `TestNewCVE_Invalid`,
  `TestNewCVE_InvalidWrapsSentinel`).
- GitHub Actions CI workflow (`.github/workflows/ci.yml`) enforcing the full
  gate: build, vet, test -race, `golangci-lint run`, and `golangci-lint fmt`
  as a required drift check.
- `CODEOWNERS`, issue templates, and a pull-request template.

### Fixed

- Package doc example now compiles: `policydsl.Replacement(...)` (a type, not
  callable) → `policydsl.NewReplacement(library, reason)`.
- LICENSE corrected from proprietary to **MIT**, matching the README claim and
  all sibling LarsArtmann SDK libraries.
- Removed a phantom `Require` builder referenced in the package + type docs
  that never existed in code.

## [0.1.0] - 2026-01-01

### Added

- Initial project structure: fluent `Builder` (`Ban` / `Companion`),
  `PolicySpec`, `Detection`, `Replacement`, `CompanionSpec`, `Severity`, and
  `Category` types. Stdlib-only, zero dependencies.
