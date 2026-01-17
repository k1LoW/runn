package builtin

import (
	"testing"
)

func TestHash_Sha256(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "byte slice",
			input:    []byte("hello"),
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name:     "number",
			input:    123,
			expected: "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
		},
	}

	h := NewHash()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Sha256(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("Sha256() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHash_Sha512(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
		},
	}

	h := NewHash()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Sha512(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("Sha512() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHash_Sha384(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "59e1748777448c69de6b800d7a33bbfb9ff1b463e44354c3553bcdb9c666fa90125a3c79f90397bdf5f6a13de828684f",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b",
		},
		{
			name:     "byte slice",
			input:    []byte("hello"),
			expected: "59e1748777448c69de6b800d7a33bbfb9ff1b463e44354c3553bcdb9c666fa90125a3c79f90397bdf5f6a13de828684f",
		},
		{
			name:     "number",
			input:    123,
			expected: "9a0a82f0c0cf31470d7affede3406cc9aa8410671520b727044eda15b4c25532a9b5cd8aaf9cec4919d76255b6bfb00f",
		},
	}

	h := NewHash()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Sha384(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("Sha384() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHash_Sha224(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "ea09ae9cc6768c50fcee903ed054556e5bfc8347907f12598aa24193",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "d14a028c2a3a2bc9476102bb288234c415a2b01f828ea62ac5b3e42f",
		},
		{
			name:     "byte slice",
			input:    []byte("hello"),
			expected: "ea09ae9cc6768c50fcee903ed054556e5bfc8347907f12598aa24193",
		},
		{
			name:     "number",
			input:    123,
			expected: "78d8045d684abd2eece923758f3cd781489df3a48e1278982466017f",
		},
	}

	h := NewHash()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Sha224(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("Sha224() = %v, want %v", got, tt.expected)
			}
		})
	}
}
