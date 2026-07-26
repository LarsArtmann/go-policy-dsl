package policydsl_test

import (
	"testing"

	policydsl "github.com/larsartmann/go-policy-dsl"
)

// External test package (policydsl_test) exercises only the exported API, the
// same surface real consumers use. This keeps the library honest: if a test
// needs an unexported symbol, that is a signal the symbol leaks implementation.

func TestBan_DefaultsToCriticalSecurity(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("gorm").Spec()

	if spec.Name != "gorm" {
		t.Errorf("expected name gorm, got %q", spec.Name)
	}

	if spec.Severity != policydsl.SeverityCritical {
		t.Errorf("expected critical severity, got %q", spec.Severity)
	}

	if spec.Category != policydsl.CategorySecurity {
		t.Errorf("expected security category, got %q", spec.Category)
	}
}

func TestBuilder_FullFluentChain(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("gorm").
		Because("ORMs hide N+1 queries").
		WithSeverity(policydsl.SeverityHigh).
		WithCategory(policydsl.CategoryPerformance).
		DetectVia(policydsl.ImportPattern("gorm.io/gorm")).
		Suggest(policydsl.NewReplacement("sqlc", "type-safe SQL")).
		VersionRange("1.0.0", "2.0.0").
		Spec()

	if spec.Reason != "ORMs hide N+1 queries" {
		t.Errorf("unexpected reason: %q", spec.Reason)
	}

	if spec.Severity != policydsl.SeverityHigh {
		t.Errorf("expected high, got %q", spec.Severity)
	}

	if spec.Category != policydsl.CategoryPerformance {
		t.Errorf("expected performance, got %q", spec.Category)
	}

	if len(spec.Detection.ImportPatterns) != 1 || spec.Detection.ImportPatterns[0] != "gorm.io/gorm" {
		t.Errorf("unexpected import patterns: %v", spec.Detection.ImportPatterns)
	}

	if len(spec.Alternatives) != 1 || spec.Alternatives[0] != "sqlc" {
		t.Errorf("unexpected alternatives: %v", spec.Alternatives)
	}

	if spec.VersionMin != "1.0.0" || spec.VersionMax != "2.0.0" {
		t.Errorf("unexpected version range: min=%q max=%q", spec.VersionMin, spec.VersionMax)
	}
}

// TestBuilder_DetectVia_Replaces confirms DetectVia overwrites any prior
// detection state (set semantics), unlike the bare append helpers.
func TestBuilder_DetectVia_Replaces(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("x").
		ImportPatterns("a").
		DetectVia(policydsl.Detection{
			ImportPatterns: []string{"b"},
			GoModPatterns:  []string{"mod"},
		}).
		Spec()

	if len(spec.Detection.ImportPatterns) != 1 || spec.Detection.ImportPatterns[0] != "b" {
		t.Errorf("DetectVia should replace, got %v", spec.Detection.ImportPatterns)
	}

	if len(spec.Detection.GoModPatterns) != 1 || spec.Detection.GoModPatterns[0] != "mod" {
		t.Errorf("unexpected go.mod patterns: %v", spec.Detection.GoModPatterns)
	}
}

// TestBuilder_AppendDetectionHelpers verifies the four append-style detection
// helpers accumulate values across calls rather than replacing.
func TestBuilder_AppendDetectionHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(*policydsl.Builder) *policydsl.Builder
		got  func(policydsl.Detection) []string
		want []string
	}{
		{
			name: "ImportPatterns",
			call: func(b *policydsl.Builder) *policydsl.Builder { return b.ImportPatterns("a", "b").ImportPatterns("c") },
			got:  func(d policydsl.Detection) []string { return d.ImportPatterns },
			want: []string{"a", "b", "c"},
		},
		{
			name: "GoModPatterns",
			call: func(b *policydsl.Builder) *policydsl.Builder { return b.GoModPatterns("m1", "m2") },
			got:  func(d policydsl.Detection) []string { return d.GoModPatterns },
			want: []string{"m1", "m2"},
		},
		{
			name: "ExcludeIfContains",
			call: func(b *policydsl.Builder) *policydsl.Builder { return b.ExcludeIfContains("nolint", "justified") },
			got:  func(d policydsl.Detection) []string { return d.ExcludeIfContains },
			want: []string{"nolint", "justified"},
		},
		{
			name: "ExcludeIfTransitiveFrom",
			call: func(b *policydsl.Builder) *policydsl.Builder { return b.ExcludeIfTransitiveFrom("ginkgo") },
			got:  func(d policydsl.Detection) []string { return d.ExcludeIfTransitiveFrom },
			want: []string{"ginkgo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.got(tt.call(policydsl.Ban("x")).Spec().Detection)
			if !equalStrings(got, tt.want) {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

// TestBuilder_Suggest_SetsDescription documents the surprising side effect:
// Suggest appends the library to Alternatives AND, when Description is empty,
// derives Description from the replacement. Intentional and documented in
// AGENTS.md; this test pins it so a silent change is caught.
func TestBuilder_Suggest_SetsDescription(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("gorm").
		Suggest(policydsl.NewReplacement("sqlc", "type-safe SQL")).
		Spec()

	if len(spec.Alternatives) != 1 || spec.Alternatives[0] != "sqlc" {
		t.Fatalf("unexpected alternatives: %v", spec.Alternatives)
	}

	const want = "Replace with sqlc: type-safe SQL"
	if spec.Description != want {
		t.Errorf("Suggest should derive Description when empty:\nwant: %q\ngot:  %q", want, spec.Description)
	}
}

// An explicit Description is never overwritten by a later Suggest.
func TestBuilder_Suggest_PreservesExplicitDescription(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("gorm").
		WithDescription("custom desc").
		Suggest(policydsl.NewReplacement("sqlc", "type-safe SQL")).
		Spec()

	if spec.Description != "custom desc" {
		t.Errorf("expected explicit description preserved, got %q", spec.Description)
	}
}

// Multiple Suggest calls append to Alternatives; only the first derives
// Description (subsequent calls see a non-empty Description and skip).
func TestBuilder_Suggest_MultipleAppendsAlternatives(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("x").
		Suggest(policydsl.NewReplacement("a", "ra")).
		Suggest(policydsl.NewReplacement("b", "rb")).
		Spec()

	if !equalStrings(spec.Alternatives, []string{"a", "b"}) {
		t.Errorf("expected [a b], got %v", spec.Alternatives)
	}

	if want := "Replace with a: ra"; spec.Description != want {
		t.Errorf("expected first suggestion's derived description %q, got %q", want, spec.Description)
	}
}

// WithAlternatives replaces the slice wholesale (set semantics), unlike Suggest
// which appends. The With- prefix denotes set/replace; bare methods append.
func TestBuilder_WithAlternatives_Replaces(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("x").
		Suggest(policydsl.NewReplacement("a", "ra")).
		WithAlternatives("p", "q").
		Spec()

	if !equalStrings(spec.Alternatives, []string{"p", "q"}) {
		t.Errorf("WithAlternatives should replace, got %v", spec.Alternatives)
	}
}

func TestBuilder_WithCVEs(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("log4j").WithCVEs("CVE-2021-44228", "CVE-2021-45046").Spec()

	if !equalStrings(spec.CVEs, []string{"CVE-2021-44228", "CVE-2021-45046"}) {
		t.Errorf("unexpected CVEs: %v", spec.CVEs)
	}
}

func TestBuilder_VersionRange(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("golog").VersionRange("", "1.99.0").Spec()

	if spec.VersionMin != "" {
		t.Errorf("expected empty min, got %q", spec.VersionMin)
	}

	if spec.VersionMax != "1.99.0" {
		t.Errorf("expected max 1.99.0, got %q", spec.VersionMax)
	}
}

func TestBuilder_RequiresCompanionAndAsCompanionOnly(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("samber/do").
		Because("DI container").
		RequiresCompanion(policydsl.Companion("samber-do-auditlog", "auditlog", "audit trail")).
		AsCompanionOnly().
		Spec()

	if !spec.CompanionOnly {
		t.Error("expected companion-only")
	}

	if len(spec.Companions) != 1 {
		t.Fatalf("expected 1 companion, got %d", len(spec.Companions))
	}

	companion := spec.Companions[0]
	if companion.Library != "samber-do-auditlog" {
		t.Errorf("unexpected companion library: %q", companion.Library)
	}

	if companion.DetectionPattern != "auditlog" {
		t.Errorf("unexpected pattern: %q", companion.DetectionPattern)
	}
}

func TestCompanion_DefaultSeverity(t *testing.T) {
	t.Parallel()

	c := policydsl.Companion("lib", "pat", "why")
	if c.Severity != policydsl.SeverityModerate {
		t.Errorf("expected moderate, got %q", c.Severity)
	}
}

func TestCompanionWithSeverity_Overrides(t *testing.T) {
	t.Parallel()

	c := policydsl.CompanionWithSeverity("lib", "pat", "why", policydsl.SeverityCritical)
	if c.Severity != policydsl.SeverityCritical {
		t.Errorf("expected critical, got %q", c.Severity)
	}
}

func TestImportPattern(t *testing.T) {
	t.Parallel()

	detection := policydsl.ImportPattern("fmt")
	if len(detection.ImportPatterns) != 1 || detection.ImportPatterns[0] != "fmt" {
		t.Errorf("unexpected: %v", detection)
	}

	if len(detection.GoModPatterns) != 0 {
		t.Errorf("expected no go.mod patterns, got %v", detection.GoModPatterns)
	}
}

func TestGoModPattern(t *testing.T) {
	t.Parallel()

	detection := policydsl.GoModPattern("github.com/foo/bar")
	if len(detection.GoModPatterns) != 1 || detection.GoModPatterns[0] != "github.com/foo/bar" {
		t.Errorf("unexpected: %v", detection)
	}

	if len(detection.ImportPatterns) != 0 {
		t.Errorf("expected no import patterns, got %v", detection.ImportPatterns)
	}
}

func TestNewReplacement(t *testing.T) {
	t.Parallel()

	replacement := policydsl.NewReplacement("github.com/x/y", "why")
	if replacement.Library != "github.com/x/y" || replacement.Reason != "why" {
		t.Errorf("unexpected replacement: %+v", replacement)
	}
}

// equalStrings is a tiny stdlib-only slice comparison (test-only) so the suite
// stays dependency-free, matching the library's zero-dep contract.
func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}

	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}
