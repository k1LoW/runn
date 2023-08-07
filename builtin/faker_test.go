package builtin

import (
	"fmt"
	"testing"
)

func TestDigitN(t *testing.T) {
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
			got := faker.DigitN(tt.n)
			if len(got) != tt.wantLength {
				t.Errorf("got=%d, want=%d", len(got), tt.wantLength)
			}
		})
	}
}

func TestLetterN(t *testing.T) {
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
			got := faker.LetterN(tt.n)
			if len(got) != tt.wantLength {
				t.Errorf("got=%d, want=%d", len(got), tt.wantLength)
			}
		})
	}
}
