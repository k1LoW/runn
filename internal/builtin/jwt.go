package builtin

import (
	"encoding/json"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

type Jwt struct {
}

func NewJwt() *Jwt {
	return &Jwt{}
}

type JWTOptions struct {
	Secret        string         `json:"secret"`         // Required
	Algorithm     string         `json:"algorithm"`      // Optional: HS256, HS384, HS512 (default: HS256)
	Subject       string         `json:"subject"`        // Optional: sub claim
	Audience      []string       `json:"audience"`       // Optional: aud claim
	Issuer        string         `json:"issuer"`         // Optional: iss claim
	ID            string         `json:"id"`             // Optional: jti claim
	ExpiresIn     string         `json:"expires_in"`     // Optional: duration like "1h", "30m"
	NotBefore     string         `json:"not_before"`     // Optional: duration like "5m"
	PrivateClaims map[string]any `json:"private_claims"` // Optional: private claims
}

func (j *Jwt) Sign(opts map[string]any) string {
	var opt *JWTOptions
	b, err := json.Marshal(opts)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &opt)
	if err != nil {
		panic(err)
	}

	builder := jwt.NewBuilder()
	if opt.Subject != "" {
		builder.Subject(opt.Subject)
	}
	if opt.Issuer != "" {
		builder.Issuer(opt.Issuer)
	}
	if len(opt.Audience) > 0 {
		builder.Audience(opt.Audience)
	}
	if opt.ID != "" {
		builder.JwtID(opt.ID)
	}
	if opt.ExpiresIn != "" {
		duration, err := time.ParseDuration(opt.ExpiresIn)
		if err == nil {
			builder.Expiration(time.Now().Add(duration))
		}
	}
	if opt.NotBefore != "" {
		duration, err := time.ParseDuration(opt.NotBefore)
		if err == nil {
			builder.NotBefore(time.Now().Add(duration))
		}
	}
	for k, v := range opt.PrivateClaims {
		builder.Claim(k, v)
	}

	token, err := builder.Build()
	if err != nil {
		panic(err)
	}

	signed, err := jwt.Sign(token, opt.createWithKey())
	if err != nil {
		panic(err)
	}

	return string(signed)
}

func (opt *JWTOptions) createWithKey() jwt.SignEncryptParseOption {
	if opt.Algorithm == "" {
		opt.Algorithm = "HS256"
	}

	var alg jwa.SignatureAlgorithm
	switch opt.Algorithm {
	case "HS256":
		alg = jwa.HS256()
	case "HS384":
		alg = jwa.HS384()
	case "HS512":
		alg = jwa.HS512()
	default:
		panic("unsupported algorithm: " + opt.Algorithm)
	}

	return jwt.WithKey(alg, []byte(opt.Secret))
}

func (j *Jwt) Parse(serialized string, opts map[string]any) map[string]any {
	var opt *JWTOptions
	b, err := json.Marshal(opts)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &opt)
	if err != nil {
		panic(err)
	}

	token, err := jwt.ParseString(serialized, opt.createWithKey())
	if err != nil {
		panic(err)
	}

	out, err := json.Marshal(token)
	if err != nil {
		panic(err)
	}

	var payload map[string]any
	err = json.Unmarshal(out, &payload)
	if err != nil {
		panic(err)
	}

	return payload
}
