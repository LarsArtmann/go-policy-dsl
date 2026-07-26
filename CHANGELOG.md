# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> **Versioning policy (pre-1.0):** the public API is not yet stable. Breaking
> changes are allowed in any `0.x` bump and are always listed under the
> **Changed** section with a `**BREAKING**` marker. The first tagged release
> will be `v0.2.0` (the rename below is breaking and the module has never been
> tagged `v0.1.0` — the entry below is the original scaffolding baseline, not a
> published release). See `ROADMAP.md` for the path to `v1.0.0`.

## [Unreleased]

### Added

- `docs/DOMAIN_LANGUAGE.md` — ubiquitous-language glossary (Ban, Companion,
  Detection, Replacement, Severity, Category, PolicySpec).
- `FEATURES.md` — honest feature inventory by status.
- `TODO_LIST.md` — short/mid-term actionable tasks.
- `ROADMAP.md` — long-term direction and the path to v1.0.0.
- Contract-locking tests for the surprising `Suggest` → `Description`
  side-effect (auto-derive when empty; preserve when explicit; append-only on
  repeated calls).
- Table-driven test covering the four append-style detection helpers
  (`ImportPatterns`, `GoModPatterns`, `ExcludeIfContains`,
  `ExcludeIfTransitiveFrom`).
- `depguard` exclusion for `_test.go` so the external test package
  (`policydsl_test`) compiles cleanly.
- `golangci-lint fmt` formatters configured (`gci`, `goimports`, `gofumpt`,
  `golines` at 120 cols).

### Changed

- **BREAKING:** renamed `GoVersionRange` → `VersionRange`,
  `GoVersionMin` → `VersionMin`, `GoVersionMax` → `VersionMax`. The old names
  lied — they constrain the version of the _library_ targeted by the policy,
  never the Go toolchain version. There are zero shipped consumers (the first
  adoption target, `library-policy`, has its own independent copy in
  `domain/policy/spec.go` and does not import this module), so the rename is
  safe.
- Rewrote the test suite as an external `policydsl_test` package with
  `t.Parallel()` on every test, so the suite exercises only the exported API
  (the same surface real consumers use) and is race-clean.
- Bumped the `golangci-lint` `go` directive `1.26.4` → `1.26.5` to match
  `go.mod`.

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
