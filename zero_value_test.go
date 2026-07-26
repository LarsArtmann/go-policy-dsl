package policydsl_test

import (
	"testing"

	policydsl "github.com/larsartmann/go-policy-dsl"
)

// Zero-value tests pin the meaning of an unbuilt struct. The DSL is
// declarative: consumers may construct a PolicySpec / Detection /
// CompanionSpec / Replacement literally (not via the Builder), so the
// zero-value semantics are part of the public contract.

// TestPolicySpec_ZeroValue documents what an unbuilt PolicySpec means: it is
// a valid, empty policy with no detection, no companions, no version bounds,
// and (importantly) it Validates clean. It is NOT a meaningful ban —
// consumers must check Name/Reason/Detection before reporting.
func TestPolicySpec_ZeroValue(t *testing.T) {
	t.Parallel()

	var spec policydsl.PolicySpec

	if spec.Name != "" {
		t.Errorf("zero-value Name should be empty, got %q", spec.Name)
	}

	if spec.Severity != "" {
		t.Errorf("zero-value Severity should be empty, got %q", spec.Severity)
	}

	if spec.Category != "" {
		t.Errorf("zero-value Category should be empty, got %q", spec.Category)
	}

	if spec.Detection.ImportPatterns != nil || spec.Detection.GoModPatterns != nil {
		t.Errorf("zero-value Detection slices should be nil, got %+v", spec.Detection)
	}

	if spec.VersionMin != nil || spec.VersionMax != nil {
		t.Errorf("zero-value version bounds should be nil, got min=%v max=%v", spec.VersionMin, spec.VersionMax)
	}

	if len(spec.Companions) != 0 {
		t.Errorf("zero-value Companions should be empty, got %v", spec.Companions)
	}

	if spec.Mode != "" {
		t.Errorf("zero-value Mode should be empty (treated as ban-active), got %q", spec.Mode)
	}

	// The zero-value spec has no inverted range, so it MUST validate clean.
	if err := spec.Validate(); err != nil {
		t.Errorf("zero-value PolicySpec should Validate clean, got %v", err)
	}
}

// TestDetection_ZeroValue documents that a zero Detection matches nothing:
// no import patterns, no go.mod patterns, no suppressions. A consumer reading
// it should treat the policy as "no detection declared".
func TestDetection_ZeroValue(t *testing.T) {
	t.Parallel()

	var detection policydsl.Detection

	if len(detection.ImportPatterns) != 0 {
		t.Errorf("zero-value ImportPatterns should be empty, got %v", detection.ImportPatterns)
	}

	if len(detection.GoModPatterns) != 0 {
		t.Errorf("zero-value GoModPatterns should be empty, got %v", detection.GoModPatterns)
	}

	if len(detection.ExcludeIfContains) != 0 {
		t.Errorf("zero-value ExcludeIfContains should be empty, got %v", detection.ExcludeIfContains)
	}

	if len(detection.ExcludeIfTransitiveFrom) != 0 {
		t.Errorf("zero-value ExcludeIfTransitiveFrom should be empty, got %v", detection.ExcludeIfTransitiveFrom)
	}
}

// TestCompanionSpec_ZeroValue documents that a zero CompanionSpec has no
// library, no pattern, and crucially an EMPTY Severity (not the Moderate
// default that the Companion constructor sets). Consumers bridging Severity
// must handle the empty case.
func TestCompanionSpec_ZeroValue(t *testing.T) {
	t.Parallel()

	var companion policydsl.CompanionSpec

	if companion.Library != "" {
		t.Errorf("zero-value Library should be empty, got %q", companion.Library)
	}

	if companion.DetectionPattern != "" {
		t.Errorf("zero-value DetectionPattern should be empty, got %q", companion.DetectionPattern)
	}

	if companion.Reason != "" {
		t.Errorf("zero-value Reason should be empty, got %q", companion.Reason)
	}

	// NOTE: zero-value Severity is "" (empty), NOT SeverityModerate. Only the
	// Companion(...) constructor sets the Moderate default. This is the
	// contract a consumer bridges against.
	if companion.Severity != "" {
		t.Errorf("zero-value Severity should be empty (NOT Moderate), got %q", companion.Severity)
	}
}

// TestReplacement_ZeroValue documents that a zero Replacement recommends
// nothing (empty Library and Reason).
func TestReplacement_ZeroValue(t *testing.T) {
	t.Parallel()

	var replacement policydsl.Replacement

	if replacement.Library != "" {
		t.Errorf("zero-value Library should be empty, got %q", replacement.Library)
	}

	if replacement.Reason != "" {
		t.Errorf("zero-value Reason should be empty, got %q", replacement.Reason)
	}
}

// TestVersion_ZeroValue documents that the zero Version is the valid version
// 0.0.0 (NOT an "unbounded" sentinel). Unbounded range bounds are represented
// by nil *Version, not by the zero value.
func TestVersion_ZeroValue(t *testing.T) {
	t.Parallel()

	var v policydsl.Version

	if v.String() != "0.0.0" {
		t.Errorf("zero Version should render as 0.0.0, got %q", v.String())
	}

	// It compares equal to an explicit 0.0.0.
	explicit := policydsl.MustNewVersion(0, 0, 0)
	if !v.Equal(explicit) {
		t.Errorf("zero Version should Equal MustNewVersion(0,0,0); got %v vs %v", v, explicit)
	}
}
