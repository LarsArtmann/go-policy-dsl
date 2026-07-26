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
Parse version strings with `ParseVersion` (which returns an error) before the
chain. **Inverted ranges (`min > max`) are not rejected at construction** — the
library never panics; `PolicySpec.Validate()` returns a concrete
`*InvertedVersionRangeError` (matched by sentinel via
`errors.Is(err, ErrInvertedVersionRange)`) for specs built any way at all. The
typed `Version` (`{Major, Minor, Patch}`) is stdlib-only (no semver
dependency). Evidence: `version.go`, `builder.go`, `policy.go`, tested by
`TestBuilder_VersionRange_Typed`, `TestBuilder_VersionRange_UnboundedMin`,
`TestBuilder_VersionRange_InvertedDeferToValidate`, `TestPolicySpec_Validate`,
`TestParseVersion`, `TestVersion_Compare`, `TestNewVersion`.

### Ban policy builder

`Ban(name)` starts a banned-library policy with `SeverityCritical` +
`CategorySecurity` + `ModeBan` defaults; every chainable method returns the
`*Builder`; `Spec()` returns the immutable `PolicySpec`. Evidence:
`builder.go`, tested by `TestBan_DefaultsToCriticalSecurity`,
`TestBan_SetsModeBan`, `TestBuilder_FullFluentChain`.

### Companion policy

`RequiresCompanion(CompanionSpec)` adds required companion libraries.
`AsCompanionOnly()` sets `Mode = ModeCompanionOnly` so the policy only
enforces companion presence (see the typed `Mode` enum below).
`Companion(...)` defaults to `SeverityModerate`; `CompanionWithSeverity(...)`
overrides. Evidence: `builder.go`, tested by
`TestBuilder_RequiresCompanionAndAsCompanionOnly`, `TestCompanion_DefaultSeverity`,
`TestCompanionWithSeverity_Overrides`.

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

`Suggest(NewReplacement(library, reason))` appends the **full** `Replacement`
(both `Library` and `Reason`) to `Alternatives` AND, when `Description` is
empty, derives `Description` from the replacement. An explicit `Description`
is never overwritten; repeated `Suggest` calls only append. `SuggestExplicit(r)`
is the no-magic variant that appends without deriving `Description`.
`WithAlternatives(...Replacement)` replaces the slice wholesale (set semantics).
`Alternatives` is typed `[]Replacement` so each entry keeps its `Reason` (no
information loss). Evidence: `builder.go`, tested by
`TestBuilder_Suggest_SetsDescription`,
`TestBuilder_Suggest_PreservesExplicitDescription`,
`TestBuilder_Suggest_MultipleAppendsAlternatives`,
`TestBuilder_SuggestExplicit_NoDescriptionDerivation`,
`TestBuilder_SuggestExplicit_MixedWithSuggest`,
`TestBuilder_WithAlternatives_Replaces`.

### Behaviour suite (BDD-style, stdlib)

User-perspective behaviour specs that read top-down as sentences
(`TestBehavior_BuildingABan` / `TestBehavior_SuggestingAReplacement` /
`TestBehavior_RequiringACompanion` / `TestBehavior_BoundingByLibraryVersion`).
Written in the stdlib `testing` package — NOT Ginkgo — because the library's
"stdlib-only, zero dependencies" contract would be contradicted by a
`require github.com/onsi/ginkgo/v2` in `go.mod` (depguard exempts `_test.go`,
but a reader of go.mod cannot see that exemption). Evidence:
`builder_behavior_test.go`.

### Severity and Category enums

`Severity` (`critical` / `high` / `moderate` / `low` / `info`) and
`Category` (`security` / `performance` / `maintainability` / `correctness` /
`licensing` / `compatibility` / `configuration`) are `string` aliases so the
DSL stays dependency-free. Consumers bridge to their own severity/category
types at the boundary. Evidence: `policy.go:26-63`.

### CVE tagging (validated)

`WithCVEs(...CVE)` tags a policy with related CVE identifiers for security
reporting. `CVE` is a branded `string` constructed via `NewCVE(id)`, which
validates the canonical MITRE form `CVE-YYYY-NNNN` (reject lowercase, wrong
digit counts, and other schemes) and returns an error on invalid input. A
free-form `[]string` cannot reach the spec. Evidence: `cve.go`, tested by
`TestBuilder_WithCVEs`, `TestNewCVE_Valid`, `TestNewCVE_Invalid`,
`TestNewCVE_InvalidWrapsSentinel`, `ExampleCVE`.

### Enforcement mode (typed `Mode` enum)

`PolicySpec.Mode` (`ModeBan` | `ModeCompanionOnly`) declares what a policy
enforces, replacing the former dishonest `CompanionOnly bool`. `Ban(...)` sets
`ModeBan` (emit ban + enforce companions); `AsCompanionOnly()` sets
`ModeCompanionOnly` (suppress ban, enforce only companions). **Deny-by-default
contract:** the ONLY mode that suppresses the ban finding is
`ModeCompanionOnly`; every other value (empty string, `ModeBan`, unknown
strings) is ban-active — a typo can never silently disable enforcement.
Evidence: `policy.go`, tested by `TestBuilder_RequiresCompanionAndAsCompanionOnly`,
`TestPolicySpec_ZeroValue`, `TestBan_SetsModeBan`,
`TestMode_DenyByDefaultContract`, `TestBehavior_RequiringACompanion`,
`ExampleCompanion`.

### Panic-free contract

The library **never panics**. Every error condition is returned, never
panicked. No `Must*` constructors exist. This is verified at two levels:
`erraudit ./...` (0 violations) and an in-repo regression test
(`TestNoPanicsInNonTestSource`) that parses every non-test `.go` file via
`go/parser` and fails if any contains a `panic(` call expression. Evidence:
`AGENTS.md` "Panic-Free" section, `panic_free_test.go`.

### Opaque-pattern contract (fuzz-pinned)

The DSL owns NO matching semantics — detection patterns are opaque strings
stored verbatim and never interpreted (literal / glob / regex is the consumer's
decision). `FuzzBuilder_PatternsOpaque` pins this: any string round-trips
unchanged through every pattern entry point. Evidence: `builder_fuzz_test.go`.

### Godoc examples

`Example*` functions (`ExampleBan`, `ExampleBan_versionRange`,
`ExampleCompanion`, `ExampleVersion`, `ExamplePolicySpec_Validate`,
`ExampleCVE`) are compile-verified by `go test` and rendered on pkg.go.dev so
documented usage cannot silently rot. Evidence: `example_test.go`, `cve_test.go`.

---

## PLANNED

### Domain validation in `Validate()` (beyond structural)

`Spec()` performs zero validation and `Validate()` currently checks only the
structural version-range invariant. Domain rules — a `Ban("x")` with no
`Because`, no detection patterns, no overrides — are still **deliberately**
not enforced (documented in `AGENTS.md`): the DSL declares what a policy IS;
the consumer validates domain fitness. Expanding `Validate()` to cover domain
rules is a future decision, deferred until the first consumer (`library-policy`)
migrates and the real required-field set is known.
