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
		tt.x.createWithKey()
	}
}

func TestCreateWithKey_UsesDefaultAlgorithmWhenNotSpecified(t *testing.T) {
	secret := "mysecret"
	withAlgorithm := JWTOptions{
		Secret:    secret,
		Algorithm: "HS256",
	}
	withoutAlgorithm := JWTOptions{
		Secret: secret,
	}

	want := withAlgorithm.createWithKey()
	got := withoutAlgorithm.createWithKey()

	if diff := cmp.Diff(got.Option, want.Option); diff != "" {
		t.Errorf("got %v\nwant %v", got.Option, want.Option)
	}
}

func TestCreateWithKey_PanicsWhenAlgorithmIsUnsupported(t *testing.T) {
	defer func() {
		err := recover()
		if err != "illegal processing" {
			t.Errorf("got %v\nwant %v", err, "illegal processing")
		}
	}()

	secret := "mysecret"
	unknownAlgorithm := JWTOptions{
		Secret:    secret,
		Algorithm: "UNSUPPORTED",
	}

	unknownAlgorithm.createWithKey()
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
	}

	j := NewJwt()
	for _, tt := range tests {
		got := j.Sign(tt.x)
		if cmp.Diff(got, tt.want) == "" {
			t.Error(cmp.Diff(got, tt.want))
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
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.rXr7y9H5-fBXgq0bPARRqn1uY1rEwd65regdC9TIcLI",
			map[string]any{
				"alg": "HS256",
				"typ": "JWT",
			},
		},
	}

	j := NewJwt()
	for _, tt := range tests {
		got := j.Parse(tt.x, opts)
		if diff := cmp.Diff(got, tt.want); diff != "" {
			t.Error(diff)
		}
	}
}
