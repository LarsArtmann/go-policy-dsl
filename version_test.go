package policydsl_test

import (
	"errors"
	"testing"

	policydsl "github.com/larsartmann/go-policy-dsl"
)

func TestNewVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		major         int
		minor         int
		patch         int
		want          policydsl.Version
		wantErr       bool
		errMustCommit bool
	}{
		{name: "all_zero", major: 0, minor: 0, patch: 0, want: policydsl.Version{}},
		{name: "happy", major: 1, minor: 2, patch: 3, want: policydsl.Version{Major: 1, Minor: 2, Patch: 3}},
		{
			name:  "large",
			major: 12,
			minor: 345,
			patch: 6789,
			want:  policydsl.Version{Major: 12, Minor: 345, Patch: 6789},
		},
		{name: "negative_major", major: -1, wantErr: true},
		{name: "negative_minor", minor: -1, wantErr: true},
		{name: "negative_patch", patch: -1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := policydsl.NewVersion(tt.major, tt.minor, tt.patch)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				if !errors.Is(err, policydsl.ErrNegativeVersion) {
					t.Errorf("expected ErrNegativeVersion, got %v", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestMustNewVersion_PanicsOnNegative(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on negative version, got none")
		}
	}()

	policydsl.MustNewVersion(-1, 0, 0)
}

func TestParseVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    policydsl.Version
		wantErr bool
	}{
		{name: "canonical", input: "1.2.3", want: policydsl.Version{Major: 1, Minor: 2, Patch: 3}},
		{name: "v_prefix", input: "v1.2.3", want: policydsl.Version{Major: 1, Minor: 2, Patch: 3}},
		{name: "all_zero", input: "0.0.0", want: policydsl.Version{}},
		{name: "large_components", input: "12.345.6789", want: policydsl.Version{Major: 12, Minor: 345, Patch: 6789}},
		{name: "empty_string", input: "", wantErr: true},
		{name: "two_components", input: "1.2", wantErr: true},
		{name: "four_components", input: "1.2.3.4", wantErr: true},
		{name: "missing_patch", input: "1.2.", wantErr: true},
		{name: "non_numeric", input: "a.b.c", wantErr: true},
		{name: "partial_non_numeric", input: "1.2.x", wantErr: true},
		{name: "negative_component", input: "1.-2.3", wantErr: true},
		{name: "plus_sign", input: "+1.2.3", wantErr: true},
		{name: "pre_release_suffix", input: "1.2.3-rc1", wantErr: true},
		{name: "build_suffix", input: "1.2.3+build", wantErr: true},
		{name: "leading_v_and_space", input: "v1.2.3 ", wantErr: true},
		{name: "named_latest", input: "latest", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := policydsl.ParseVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tt.input)
				}

				if !errors.Is(err, policydsl.ErrInvalidVersion) {
					t.Errorf("expected ErrInvalidVersion for %q, got %v", tt.input, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}

			if got != tt.want {
				t.Errorf("input %q: expected %v, got %v", tt.input, tt.want, got)
			}
		})
	}
}

func TestMustParseVersion_PanicsOnError(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on malformed version, got none")
		}
	}()

	policydsl.MustParseVersion("not-a-version")
}

func TestVersion_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		v    policydsl.Version
		want string
	}{
		{policydsl.Version{}, "0.0.0"},
		{policydsl.Version{Major: 1, Minor: 2, Patch: 3}, "1.2.3"},
		{policydsl.Version{Major: 12, Minor: 345, Patch: 6789}, "12.345.6789"},
	}

	for _, tt := range tests {
		got := tt.v.String()
		if got != tt.want {
			t.Errorf("%v.String() = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestVersion_Compare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b policydsl.Version
		want int // sign-only comparison
	}{
		{name: "equal", a: policydsl.Version{1, 2, 3}, b: policydsl.Version{1, 2, 3}, want: 0},
		{name: "major_differs_higher", a: policydsl.Version{2, 0, 0}, b: policydsl.Version{1, 9, 9}, want: 1},
		{name: "major_differs_lower", a: policydsl.Version{1, 9, 9}, b: policydsl.Version{2, 0, 0}, want: -1},
		{name: "minor_differs_higher", a: policydsl.Version{1, 3, 0}, b: policydsl.Version{1, 2, 9}, want: 1},
		{name: "minor_differs_lower", a: policydsl.Version{1, 2, 0}, b: policydsl.Version{1, 3, 0}, want: -1},
		{name: "patch_differs_higher", a: policydsl.Version{1, 2, 4}, b: policydsl.Version{1, 2, 3}, want: 1},
		{name: "patch_differs_lower", a: policydsl.Version{1, 2, 2}, b: policydsl.Version{1, 2, 3}, want: -1},
		{name: "zero_vs_one", a: policydsl.Version{}, b: policydsl.Version{1, 0, 0}, want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := signum(tt.a.Compare(tt.b))
			if got != tt.want {
				t.Errorf("%v.Compare(%v) sign = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestVersion_RelationHelpers(t *testing.T) {
	t.Parallel()

	low := policydsl.Version{1, 0, 0}
	high := policydsl.Version{2, 0, 0}
	lowCopy := low

	if !low.Before(high) {
		t.Errorf("%v.Before(%v) should be true", low, high)
	}

	if high.Before(low) {
		t.Errorf("%v.Before(%v) should be false", high, low)
	}

	if !high.After(low) {
		t.Errorf("%v.After(%v) should be true", high, low)
	}

	if low.After(high) {
		t.Errorf("%v.After(%v) should be false", low, high)
	}

	if !low.Equal(lowCopy) {
		t.Errorf("%v.Equal(%v) should be true", low, lowCopy)
	}

	if low.Equal(high) {
		t.Errorf("%v.Equal(%v) should be false", low, high)
	}
}

func TestVersion_ParseRoundTrip(t *testing.T) {
	t.Parallel()

	versionStrings := []string{"0.0.0", "1.0.0", "1.2.3", "12.345.6789"}

	for _, versionStr := range versionStrings {
		parsed, err := policydsl.ParseVersion(versionStr)
		if err != nil {
			t.Fatalf("ParseVersion(%q): %v", versionStr, err)
		}

		if rendered := parsed.String(); rendered != versionStr {
			t.Errorf("round-trip mismatch: input %q -> Version -> %q", versionStr, rendered)
		}
	}
}

// signum collapses a non-zero integer to -1 / 0 / +1 so the Compare tests
// assert only the sign, not the magnitude.
func signum(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}
