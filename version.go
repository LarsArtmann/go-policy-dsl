package policydsl

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Version is a parsed semantic version (major.minor.patch). It is the
// version of the _library_ targeted by a policy, never the Go toolchain
// version.
//
// The zero Version is the valid version 0.0.0; an _unbounded_ bound on a
// range is represented by a nil *Version, not by the zero Version. Construct
// via NewVersion (ints) or ParseVersion (string) so components are validated.
type Version struct {
	Major int
	Minor int
	Patch int
}

// ErrInvalidVersion is returned when a version string is not exactly three
// non-negative integer components separated by dots (e.g. "1.2.3").
var ErrInvalidVersion = errors.New("policydsl: invalid version string, want \"major.minor.patch\"")

// ErrNegativeVersion is returned when a version component is negative.
var ErrNegativeVersion = errors.New("policydsl: version components must be non-negative")

// NewVersion constructs a Version from integer components, rejecting negatives.
func NewVersion(major, minor, patch int) (Version, error) {
	if major < 0 || minor < 0 || patch < 0 {
		return Version{}, ErrNegativeVersion
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// MustNewVersion is the convenience form of NewVersion for package-level
// policy initialization; it panics on a negative component.
func MustNewVersion(major, minor, patch int) Version {
	v, err := NewVersion(major, minor, patch)
	if err != nil {
		panic(err)
	}

	return v
}

// versionComponentCount is the number of dot-separated components in a full
// "major.minor.patch" version string.
const versionComponentCount = 3

// ParseVersion parses a strict "major.minor.patch" string (e.g. "1.2.3").
// Each component must be a non-negative integer with no leading sign. A
// leading "v" (as in "v1.2.3") is accepted. Anything else — "1.2", "1.2.3.4",
// "latest", "" — returns ErrInvalidVersion. Empty string is NOT a valid
// version; an unbounded range bound is represented by a nil *Version.
func ParseVersion(s string) (Version, error) {
	original := s

	trimmed := strings.TrimPrefix(s, "v")

	parts := strings.Split(trimmed, ".")
	if len(parts) != versionComponentCount {
		return Version{}, fmt.Errorf("%w: %q", ErrInvalidVersion, original)
	}

	major, err := parseVersionComponent(parts[0], "major", original)
	if err != nil {
		return Version{}, err
	}

	minor, err := parseVersionComponent(parts[1], "minor", original)
	if err != nil {
		return Version{}, err
	}

	patch, err := parseVersionComponent(parts[2], "patch", original)
	if err != nil {
		return Version{}, err
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// parseVersionComponent parses one non-negative integer component of a
// version string, producing a descriptive error on failure.
func parseVersionComponent(raw, name, original string) (int, error) {
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("%w: component %s of %q", ErrInvalidVersion, name, original)
	}

	return n, nil
}

// MustParseVersion is the convenience form of ParseVersion for package-level
// policy initialization; it panics on a parse error.
func MustParseVersion(s string) Version {
	v, err := ParseVersion(s)
	if err != nil {
		panic(err)
	}

	return v
}

// String renders the version as "major.minor.patch".
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare returns -1 if v < other, 0 if equal, +1 if v > other. The ordering
// is the natural numeric ordering on major, then minor, then patch.
func (v Version) Compare(other Version) int {
	switch {
	case v.Major != other.Major:
		return cmpSign(v.Major - other.Major)
	case v.Minor != other.Minor:
		return cmpSign(v.Minor - other.Minor)
	case v.Patch != other.Patch:
		return cmpSign(v.Patch - other.Patch)
	default:
		return 0
	}
}

// Before reports whether v is strictly less than other.
func (v Version) Before(other Version) bool { return v.Compare(other) < 0 }

// After reports whether v is strictly greater than other.
func (v Version) After(other Version) bool { return v.Compare(other) > 0 }

// Equal reports whether v and other are the same version.
func (v Version) Equal(other Version) bool { return v.Compare(other) == 0 }

// cmpSign collapses a non-zero integer difference to -1 / 0 / +1 without
// overflowing at the extremes (subtraction of two ints can overflow; the
// sign-detected branch below avoids relying on the raw difference magnitude).
func cmpSign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}
