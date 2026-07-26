# Features — go-policy-dsl

Honest inventory of what exists, by status. Status vocabulary:

- **FULLY_FUNCTIONAL** — code present AND verified working (tests pass).
- **PARTIALLY_FUNCTIONAL** — ships but has known gaps, edge-case bugs, or missing pieces.
- **BROKEN** — code exists but does not work / is disabled / fails.
- **PLANNED** — designed or documented but **no code exists yet**.

Never rounded up. If a feature cannot be confirmed working, it is
`PARTIALLY_FUNCTIONAL` at most.

---

## FULLY_FUNCTIONAL

### Version constraints

`VersionRange(minVer, maxVer *Version)` sets inclusive library-version
constraints (NOT Go toolchain version). `nil` on either side means unbounded.
**Inverted ranges (`min > max`) are rejected at construction** — the builder
method panics, and `PolicySpec.Validate()` returns `ErrInvertedVersionRange`
for direct-field assignment. `VersionRangeStrings(min, max string)` is the
string convenience (empty = unbounded; parses via `MustParseVersion`, panics
on parse error or inversion). The typed `Version` (`{Major, Minor, Patch}`)
is stdlib-only (no semver dependency). Evidence: `version.go`,
`builder.go:130-160`, `policy.go:130-152`, tested by `TestBuilder_VersionRange_Typed`,
`TestBuilder_VersionRangeStrings`, `TestBuilder_VersionRange_Typed_InvertedPanics`,
`TestBuilder_VersionRangeStrings_InvertedPanics`, `TestPolicySpec_Validate`,
`TestParseVersion`, `TestVersion_Compare`, `TestNewVersion`.

### Ban policy builder

`Ban(name)` starts a banned-library policy with `SeverityCritical` +
`CategorySecurity` defaults; every chainable method returns the `*Builder`;
`Spec()` returns the immutable `PolicySpec`. Evidence: `builder.go:12`,
tested by `TestBan_DefaultsToCriticalSecurity`, `TestBuilder_FullFluentChain`.

### Companion policy

`RequiresCompanion(CompanionSpec)` adds required companion libraries.
`AsCompanionOnly()` suppresses the ban so the policy only enforces companion
presence. `Companion(...)` defaults to `SeverityModerate`;
`CompanionWithSeverity(...)` overrides. Evidence: `builder.go:124-137`,
`builder.go:157-174`, tested by `TestBuilder_RequiresCompanionAndAsCompanionOnly`,
`TestCompanion_DefaultSeverity`, `TestCompanionWithSeverity_Overrides`.

### Detection model

`Detection{ImportPatterns, GoModPatterns, ExcludeIfContains,
ExcludeIfTransitiveFrom}` declares how a consumer finds a violation.
`ImportPattern(p)` / `GoModPattern(p)` are single-pattern constructors;
`DetectVia(d)` replaces the whole struct; the four bare methods
(`ImportPatterns`, `GoModPatterns`, `ExcludeIfContains`,
`ExcludeIfTransitiveFrom`) append. Evidence: `policy.go:67`,
`builder.go:44-80`, `builder.go:146-154`, tested by `TestBuilder_DetectVia_Replaces`,
`TestBuilder_AppendDetectionHelpers`, `TestImportPattern`, `TestGoModPattern`.

### Replacement suggestion

`Suggest(NewReplacement(library, reason))` appends the library to
`Alternatives` AND, when `Description` is empty, derives `Description` from
the replacement. An explicit `Description` is never overwritten; repeated
`Suggest` calls only append to `Alternatives`. `WithAlternatives(...)`
replaces the slice wholesale (set semantics). Evidence: `builder.go:90-105`,
tested by `TestBuilder_Suggest_SetsDescription`,
`TestBuilder_Suggest_PreservesExplicitDescription`,
`TestBuilder_Suggest_MultipleAppendsAlternatives`,
`TestBuilder_WithAlternatives_Replaces`.

### Severity and Category enums

`Severity` (`critical` / `high` / `moderate` / `low` / `info`) and
`Category` (`security` / `performance` / `maintainability` / `correctness` /
`licensing` / `compatibility` / `configuration`) are `string` aliases so the
DSL stays dependency-free. Consumers bridge to their own severity/category
types at the boundary. Evidence: `policy.go:26-63`.

### CVE tagging

`WithCVEs(...)` tags a policy with related CVE identifiers for security
reporting. Evidence: `builder.go:108`, tested by `TestBuilder_WithCVEs`.

---

## PLANNED

### Domain validation in `Validate()` (beyond structural)

`Spec()` performs zero validation and `Validate()` currently checks only the
structural version-range invariant. Domain rules — a `Ban("x")` with no
`Because`, no detection patterns, no overrides — are still **deliberately**
not enforced (documented in `AGENTS.md`): the DSL declares what a policy IS;
the consumer validates domain fitness. Expanding `Validate()` to cover domain
rules is a future decision, deferred until the first consumer (`library-policy`)
migrates and the real required-field set is known. Tracked in `TODO_LIST.md`.

### Ginkgo BDD suite

User-perspective behaviour tests for the fluent chain ("building a ban reads
like prose"). Sibling libraries (`go-error-family`) use Ginkgo for this.
Tracked in `TODO_LIST.md`.

### godoc `Example*` functions

`func ExampleBan()` / `func ExampleCompanion()` so pkg.go.dev renders
runnable, compile-verified examples and `go test` catches regressions in
documented usage. Tracked in `TODO_LIST.md`.
