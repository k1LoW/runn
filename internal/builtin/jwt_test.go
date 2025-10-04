package builtin

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestCreateWithKey(t *testing.T) {
	tests := []struct {
		x JWTOptions
	}{
		{
			JWTOptions{
				Secret:    "mysecret",
				Algorithm: "HS256",
			},
		},
		{
			JWTOptions{
				Secret:    "mysecret",
				Algorithm: "HS384",
			},
		},
		{
			JWTOptions{
				Secret:    "mysecret",
				Algorithm: "HS512",
			},
		},
		{
			JWTOptions{
				Secret: "mysecret",
			},
		},
	}

	for _, tt := range tests {
		_, err := tt.x.createWithKey()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestCreateWithKey_PanicsWhenAlgorithmIsUnsupported(t *testing.T) {
	unknownAlgorithm := JWTOptions{
		Secret:    "mysecret",
		Algorithm: "UNSUPPORTED",
	}

	_, err := unknownAlgorithm.createWithKey()
	if err == nil || err.Error() != "unsupported algorithm: UNSUPPORTED" {
		t.Errorf("expected error 'unsupported algorithm: UNSUPPORTED', got %v", err)
	}
}

func TestSign(t *testing.T) {
	secret := "mysecret"

	tests := []struct {
		x    map[string]any
		want string
	}{
		{
			map[string]any{
				"secret": secret,
			},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.rXr7y9H5-fBXgq0bPARRqn1uY1rEwd65regdC9TIcLI",
		},
		{ // The default algorithm is HS256. The expected value is the same as when unspecified.
			map[string]any{
				"secret":    secret,
				"algorithm": "HS256",
			},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.rXr7y9H5-fBXgq0bPARRqn1uY1rEwd65regdC9TIcLI",
		},
	}

	j := NewJwt()
	for _, tt := range tests {
		got, err := j.Sign(tt.x)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if diff := cmp.Diff(got, tt.want); diff != "" {
			t.Error(diff)
		}
	}
}

func TestParse(t *testing.T) {
	opts := map[string]any{
		"secret": "mysecret",
	}

	tests := []struct {
		x    string
		want map[string]any
	}{
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsidXNlcjEiLCJ1c2VyMiJdLCJmb28iOiJiYXIiLCJpc3MiOiJydW5uIiwianRpIjoidW5pcXVlLWlkIiwic3ViIjoiQTEyMyJ9.96_UkX2n4i_49R9qcshj6lc3WN8LqNWc0Dvdpc1FOag",
			map[string]any{
				"aud": []any{"user1", "user2"},
				"foo": "bar",
				"iss": "runn",
				"jti": "unique-id",
				"sub": "A123",
			},
		},
	}

	j := NewJwt()
	for _, tt := range tests {
		got, err := j.Parse(tt.x, opts)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if diff := cmp.Diff(got, tt.want); diff != "" {
			t.Error(diff)
		}
	}
}
