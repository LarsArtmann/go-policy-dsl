// Package policydsl provides a fluent, compile-time-checked Go DSL for
// declaring library governance policies (bans, requirements, companions).
//
// It is the canonical policy-declaration language for the LarsArtmann ecosystem:
// library-policy is the primary consumer, and go-linter-sdk can use it as its
// rule-declaration language. Policies declared here are plain Go values — no
// YAML, no codegen, no runtime parsing — so they get full IDE support and
// compile-time validation.
//
// Example:
//
//	policydsl.Ban("gorm").
//	    Because("ORMs hide N+1 queries; use sqlc for type-safe SQL").
//	    WithSeverity(policydsl.SeverityCritical).
//	    WithCategory(policydsl.CategoryPerformance).
//	    DetectVia(policydsl.ImportPattern("gorm.io/gorm")).
//	    Suggest(policydsl.Replacement("github.com/yourorg/sqlc-queries"))
//
// The DSL is intentionally dependency-free (stdlib only) so any tool — a CLI,
// a CI check, an LSP server, a golangci-lint plugin — can adopt it without
// pulling in a heavyweight framework.
package policydsl

// Severity rates how serious a policy violation is. Maps to finding.Severity
// at the consumer boundary.
type Severity string

const (
	// SeverityCritical blocks: the violation is a production incident waiting
	// to happen (e.g. a known-vulnerable dependency, a correctness-destroying
	// pattern).
	SeverityCritical Severity = "critical"

	// SeverityHigh is a strong warning: fix before merge unless explicitly
	// justified (e.g. a performance anti-pattern with measurable impact).
	SeverityHigh Severity = "high"

	// SeverityModerate is a recommendation worth acting on but not blocking
	// (e.g. a maintainability concern, a deprecated-but-functional library).
	SeverityModerate Severity = "moderate"

	// SeverityLow is informational (e.g. a newer version available, a style
	// preference).
	SeverityLow Severity = "low"

	// SeverityInfo is a neutral observation (e.g. a detected companion that
	// should be present alongside a chosen library).
	SeverityInfo Severity = "info"
)

// Category classifies WHY a policy exists, so consumers can filter and report
// by concern (e.g. "show me all security bans").
type Category string

const (
	CategorySecurity       Category = "security"
	CategoryPerformance    Category = "performance"
	CategoryMaintainability Category = "maintainability"
	CategoryCorrectness    Category = "correctness"
	CategoryLicensing      Category = "licensing"
	CategoryCompatibility  Category = "compatibility"
	CategoryConfiguration  Category = "configuration"
)

// Detection declares how a policy violation is found. A policy may declare
// import patterns (source-level), go.mod patterns (manifest-level), or both.
type Detection struct {
	// ImportPatterns match package paths in Go source files.
	ImportPatterns []string

	// GoModPatterns match module paths in go.mod.
	GoModPatterns []string

	// ExcludeIfContains suppresses the violation when any of these strings
	// appear in the matched file or go.mod (escape hatch for justified use).
	ExcludeIfContains []string

	// ExcludeIfTransitiveFrom lists parent libraries whose direct presence
	// justifies this library appearing transitively (avoids false positives
	// on indirect dependencies).
	ExcludeIfTransitiveFrom []string
}

// Replacement recommends a swap-in alternative for a banned library.
type Replacement struct {
	// Library is the recommended module path.
	Library string

	// Reason explains why the replacement is better.
	Reason string
}

// CompanionSpec declares a library that MUST be present alongside a chosen one
// (e.g. samber-do-auditlog must accompany samber/do).
type CompanionSpec struct {
	Library          string
	DetectionPattern string
	Reason           string
	Severity         Severity
}

// PolicySpec is the declarative Go representation of a library governance
// policy. Construct via the fluent Builder (Ban/Require/Companion) so every
// field has a sensible default and the call sites read like prose.
type PolicySpec struct {
	Name     string
	Reason   string
	Severity Severity
	Category Category

	Detection Detection

	// Metadata for reporting
	Description  string
	Alternatives []string
	CVEs         []string

	// Version constraints (inclusive). Empty means unconstrained.
	GoVersionMin string
	GoVersionMax string

	// Companions that must be present when this library is used.
	Companions    []CompanionSpec
	CompanionOnly bool
}
