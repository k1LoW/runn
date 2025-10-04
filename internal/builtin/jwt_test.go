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
				Secret:    "mysecret",
				Algorithm: "HS256",
			},
		},
		{
			JWTOptions{
				Secret:    "mysecret",
				Algorithm: "UNKNOWN",
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

func TestCreateWithKeyAtDefault(t *testing.T) {
	secret := "mysecret"
	original := JWTOptions{
		Secret:    secret,
		Algorithm: "HS256",
	}
	unknown := JWTOptions{
		Secret:    secret,
		Algorithm: "UNKNOWN",
	}
	none := JWTOptions{
		Secret: secret,
	}

	want := original.createWithKey()

	if diff := cmp.Diff(want, unknown.createWithKey()); diff != "" {
		t.Error(diff)
	}
	if diff := cmp.Diff(want, none.createWithKey()); diff != "" {
		t.Error(diff)
	}
}

func TestSign(t *testing.T) {
	secret := "mysecret"

	tests := []struct {
		x    JWTOptions
		want string
	}{
		{
			JWTOptions{
				Secret: secret,
			},
			"ss",
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
