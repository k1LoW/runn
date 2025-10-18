package builtin

import (
	"fmt"
	"regexp"
	"testing"
)

func TestDigitN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		n          int
		wantLength int
	}{
		{10, 10},
		{-1, 0},
	}
	faker := NewFaker()
	for _, tt := range tests {
		t.Run(fmt.Sprintf("n=%d", tt.n), func(t *testing.T) {
			t.Parallel()
			got := faker.DigitN(tt.n)
			if len(got) != tt.wantLength {
				t.Errorf("got=%d, want=%d", len(got), tt.wantLength)
			}
		})
	}
}

func TestLetterN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		n          int
		wantLength int
	}{
		{10, 10},
		{-1, 0},
	}
	faker := NewFaker()
	for _, tt := range tests {
		t.Run(fmt.Sprintf("n=%d", tt.n), func(t *testing.T) {
			t.Parallel()
			got := faker.LetterN(tt.n)
			if len(got) != tt.wantLength {
				t.Errorf("got=%d, want=%d", len(got), tt.wantLength)
			}
		})
	}
}

func TestRegex(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		pattern string
	}{
		// Basic character classes
		{
			name:    "lowercase letters",
			pattern: "[a-z]{5}",
		},
		{
			name:    "uppercase letters",
			pattern: "[A-Z]{3}",
		},
		{
			name:    "digits",
			pattern: "[0-9]{4}",
		},
		{
			name:    "alphanumeric",
			pattern: "[a-zA-Z0-9]{10}",
		},
		// Quantifiers
		{
			name:    "exact count",
			pattern: "[a-z]{3}",
		},
		{
			name:    "range count",
			pattern: "[0-9]{2,5}",
		},
		{
			name:    "plus quantifier",
			pattern: "[a-z]+",
		},
		{
			name:    "asterisk quantifier",
			pattern: "[0-9]*",
		},
		// Complex patterns
		{
			name:    "email-like pattern",
			pattern: "[a-z]{5,10}@[a-z]{3,7}\\.[a-z]{2,3}",
		},
		{
			name:    "phone number pattern",
			pattern: "[0-9]{3}-[0-9]{4}-[0-9]{4}",
		},
		{
			name:    "hex color code",
			pattern: "#[0-9a-fA-F]{6}",
		},
		{
			name:    "mixed alphanumeric with special chars",
			pattern: "[a-zA-Z0-9_-]{8,12}",
		},
		{
			name:    "UUID-like pattern",
			pattern: "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}",
		},
	}

	faker := NewFaker()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := faker.Regex(tt.pattern)
			if err != nil {
				t.Fatalf("got error = %v", err)
			}

			re := regexp.MustCompile("^" + tt.pattern + "$")
			if !re.MatchString(got) {
				t.Errorf("generated string %q does not match pattern %q", got, tt.pattern)
			}

			for i := range 10 {
				result, err := faker.Regex(tt.pattern)
				if err != nil {
					t.Fatalf("got error = %v", err)
				}
				if !re.MatchString(result) {
					t.Errorf("generated string %q does not match pattern %q on iteration %d", result, tt.pattern, i)
				}
			}
		})
	}
}
