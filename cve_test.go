package policydsl_test

import (
	"errors"
	"fmt"
	"testing"

	policydsl "github.com/larsartmann/go-policy-dsl"
)

// CVE validation tests pin the canonical MITRE format "CVE-YYYY-NNNN". The
// branded type's whole value is that an invalid identifier cannot reach a
// PolicySpec, so the accept/reject boundary is the contract.

func TestNewCVE_Valid(t *testing.T) {
	t.Parallel()

	validIdentifiers := []string{
		"CVE-2021-44228",
		"CVE-1999-0001",
		"CVE-2023-12345", // sequence longer than 4 digits is allowed by MITRE
		"CVE-2099-9999",
	}

	for _, identifier := range validIdentifiers {
		cve, err := policydsl.NewCVE(identifier)
		if err != nil {
			t.Errorf("NewCVE(%q) returned unexpected error: %v", identifier, err)
		}

		if cve.String() != identifier {
			t.Errorf("NewCVE(%q).String() = %q, want %q", identifier, cve, identifier)
		}
	}
}

func TestNewCVE_Invalid(t *testing.T) {
	t.Parallel()

	invalidIdentifiers := []string{
		"",                 // empty
		"cve-2021-44228",   // lowercase prefix
		"CVE-21-44228",     // year not 4 digits
		"CVE-2021-1",       // sequence too short
		"CVE-2021-442",     // sequence 3 digits
		"2021-44228",       // missing prefix
		"CVE-2021-44228 ",  // trailing space
		" CVE-2021-44228",  // leading space
		"CVE-2021-44228-X", // trailing junk
		"CVE–2021-44228",   // en-dash, not hyphen
		"CVE-2021-44",      // 2-digit sequence
		"GHSA-xxxx",        // wrong scheme (GitHub advisory)
	}

	for _, identifier := range invalidIdentifiers {
		if _, err := policydsl.NewCVE(identifier); err == nil {
			t.Errorf("NewCVE(%q) should return ErrInvalidCVE, got nil", identifier)
		}
	}
}

func TestNewCVE_InvalidWrapsSentinel(t *testing.T) {
	t.Parallel()

	_, err := policydsl.NewCVE("not-a-cve")
	if !errors.Is(err, policydsl.ErrInvalidCVE) {
		t.Errorf("expected ErrInvalidCVE sentinel, got %v", err)
	}
}

func ExampleCVE() {
	cve, _ := policydsl.NewCVE("CVE-2021-44228")
	fmt.Println(cve)
	// Output: CVE-2021-44228
}
