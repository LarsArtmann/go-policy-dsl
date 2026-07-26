package policydsl

import "fmt"

// Builder provides a fluent API for constructing PolicySpecs. Created via
// Ban(name) or Companion(...). Each method returns the Builder so calls chain.
//
// Method-naming convention (enforced by tests and docs, not the type system):
//
//	With<X>(...)        — SET / REPLACE the field wholesale (e.g. WithSeverity,
//	                      WithDescription, WithAlternatives, WithCVEs).
//	<X>(...)  (bare)    — APPEND to a slice field (e.g. ImportPatterns,
//	                      GoModPatterns, ExcludeIfContains, Suggest,
//	                      RequiresCompanion).
//	DetectVia(d)        — REPLACE the whole Detection struct.
//	As<X>()             — SET a mode flag (e.g. AsCompanionOnly).
//	Spec()              — TERMINATE the chain, return the immutable value.
//
// When adding a new method, pick the form that matches its semantics so the
// convention stays predictable at call sites.
type Builder struct {
	spec PolicySpec
}

// Ban starts building a banned-library policy with the given name. Defaults
// to SeverityCritical + CategorySecurity — override with WithSeverity /
// WithCategory for non-security concerns.
func Ban(name string) *Builder {
	return &Builder{
		spec: PolicySpec{
			Name:     name,
			Severity: SeverityCritical,
			Category: CategorySecurity,
		},
	}
}

// Because sets the human-readable reason for the policy. Required — a policy
// without a reason is unreviewable.
func (b *Builder) Because(reason string) *Builder {
	b.spec.Reason = reason

	return b
}

// WithSeverity sets the severity level (overrides the Ban default of Critical).
func (b *Builder) WithSeverity(s Severity) *Builder {
	b.spec.Severity = s

	return b
}

// WithCategory sets the policy category (overrides the Ban default of Security).
func (b *Builder) WithCategory(c Category) *Builder {
	b.spec.Category = c

	return b
}

// DetectVia sets the detection patterns (source imports and/or go.mod matches).
// Convenience over calling ImportPatterns / GoModPatterns separately.
func (b *Builder) DetectVia(d Detection) *Builder {
	b.spec.Detection = d

	return b
}

// ImportPatterns adds import-path detection patterns (matched in Go source).
func (b *Builder) ImportPatterns(patterns ...string) *Builder {
	b.spec.Detection.ImportPatterns = append(b.spec.Detection.ImportPatterns, patterns...)

	return b
}

// GoModPatterns adds go.mod module-path detection patterns.
func (b *Builder) GoModPatterns(patterns ...string) *Builder {
	b.spec.Detection.GoModPatterns = append(b.spec.Detection.GoModPatterns, patterns...)

	return b
}

// ExcludeIfContains adds suppressions: if any string appears in the matched
// file/go.mod, the violation is suppressed (justified-use escape hatch).
func (b *Builder) ExcludeIfContains(patterns ...string) *Builder {
	b.spec.Detection.ExcludeIfContains = append(b.spec.Detection.ExcludeIfContains, patterns...)

	return b
}

// ExcludeIfTransitiveFrom lists parent libraries whose direct presence
// justifies this library appearing transitively (avoids false positives).
func (b *Builder) ExcludeIfTransitiveFrom(libraries ...string) *Builder {
	b.spec.Detection.ExcludeIfTransitiveFrom = append(b.spec.Detection.ExcludeIfTransitiveFrom, libraries...)

	return b
}

// WithDescription sets the optional detailed description for reporting.
func (b *Builder) WithDescription(desc string) *Builder {
	b.spec.Description = desc

	return b
}

// Suggest sets the recommended replacement.
func (b *Builder) Suggest(r Replacement) *Builder {
	b.spec.Alternatives = append(b.spec.Alternatives, r.Library)

	if b.spec.Description == "" && r.Reason != "" {
		b.spec.Description = "Replace with " + r.Library + ": " + r.Reason
	}

	return b
}

// WithAlternatives sets the recommended replacement libraries directly.
func (b *Builder) WithAlternatives(alts ...string) *Builder {
	b.spec.Alternatives = alts

	return b
}

// WithCVEs tags the policy with related CVE identifiers for security reporting.
func (b *Builder) WithCVEs(cves ...string) *Builder {
	b.spec.CVEs = cves

	return b
}

// VersionRange sets inclusive version constraints for the library targeted by
// this policy (NOT the Go toolchain version). nil on either side means
// unbounded. Panics if both bounds are non-nil and min > max (a nonsensical
// inverted range) — this is a programmer error that should fail fast at
// package-level policy initialization, not surface silently at runtime.
func (b *Builder) VersionRange(minVer, maxVer *Version) *Builder {
	if minVer != nil && maxVer != nil && minVer.After(*maxVer) {
		panic(fmt.Errorf("%w: min %s > max %s", ErrInvertedVersionRange, minVer, maxVer))
	}

	b.spec.VersionMin = minVer
	b.spec.VersionMax = maxVer

	return b
}

// VersionRangeStrings is the string convenience form of VersionRange: each
// bound is parsed via ParseVersion (empty string means unbounded on that
// side). Panics on a parse error or an inverted range. Intended for
// package-level policy vars where the version literals are known at compile
// time; use VersionRange with parsed Versions when the input is untrusted.
func (b *Builder) VersionRangeStrings(minVer, maxVer string) *Builder {
	return b.VersionRange(parseOptionalVersion(minVer), parseOptionalVersion(maxVer))
}

// parseOptionalVersion parses a version bound where empty string means
// unbounded (nil). Panics on a non-empty malformed string, matching the
// Must-parse convention used for package-level policy initialization.
func parseOptionalVersion(s string) *Version {
	if s == "" {
		return nil
	}

	v := MustParseVersion(s)

	return &v
}

// RequiresCompanion adds a required companion library spec.
func (b *Builder) RequiresCompanion(c CompanionSpec) *Builder {
	b.spec.Companions = append(b.spec.Companions, c)

	return b
}

// AsCompanionOnly marks this entry as companion-only: never emit a ban finding,
// only enforce that required companion libraries are present.
func (b *Builder) AsCompanionOnly() *Builder {
	b.spec.CompanionOnly = true

	return b
}

// Spec returns the finished PolicySpec.
func (b *Builder) Spec() PolicySpec {
	return b.spec
}

// ImportPattern is a convenience constructor for a Detection matching a single
// import path (the most common case for source-level bans).
func ImportPattern(pattern string) Detection {
	return Detection{ImportPatterns: []string{pattern}}
}

// GoModPattern is a convenience constructor for a Detection matching a single
// go.mod module path.
func GoModPattern(pattern string) Detection {
	return Detection{GoModPatterns: []string{pattern}}
}

// Companion creates a CompanionSpec with the default Moderate severity.
func Companion(lib, pattern, reason string) CompanionSpec {
	return CompanionSpec{
		Library:          lib,
		DetectionPattern: pattern,
		Reason:           reason,
		Severity:         SeverityModerate,
	}
}

// CompanionWithSeverity creates a CompanionSpec with a custom severity.
func CompanionWithSeverity(lib, pattern, reason string, sev Severity) CompanionSpec {
	return CompanionSpec{
		Library:          lib,
		DetectionPattern: pattern,
		Reason:           reason,
		Severity:         sev,
	}
}

// NewReplacement constructs a Replacement value (recommended swap-in alternative).
func NewReplacement(library, reason string) Replacement {
	return Replacement{Library: library, Reason: reason}
}
