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
	Secret        string                 `json:"secret"`         // Required
	Algorithm     string                 `json:"algorithm"`      // Optional: HS256, HS384, HS512 (default: HS256)
	Subject       string                 `json:"subject"`        // Optional: sub claim
	Audience      []string               `json:"audience"`       // Optional: aud claim
	Issuer        string                 `json:"issuer"`         // Optional: iss claim
	ID            string                 `json:"id"`             // Optional: jti claim
	ExpiresIn     string                 `json:"expires_in"`     // Optional: duration like "1h", "30m"
	NotBefore     string                 `json:"not_before"`     // Optional: duration like "5m"
	PrivateClaims map[string]interface{} `json:"private_claims"` // Optional: private claims
}

func (j *Jwt) Sign(opts map[string]interface{}) string {
	var options JWTOptions
	b, err := json.Marshal(opts)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &options)
	if err != nil {
		panic(err)
	}

	builder := jwt.NewBuilder()
	if options.Subject != "" {
		builder.Subject(options.Subject)
	}
	if options.Issuer != "" {
		builder.Issuer(options.Issuer)
	}
	if len(options.Audience) > 0 {
		builder.Audience(options.Audience)
	}
	if options.ID != "" {
		builder.JwtID(options.ID)
	}
	if options.ExpiresIn != "" {
		duration, err := time.ParseDuration(options.ExpiresIn)
		if err == nil {
			builder.Expiration(time.Now().Add(duration))
		}
	}
	if options.NotBefore != "" {
		duration, err := time.ParseDuration(options.NotBefore)
		if err == nil {
			builder.NotBefore(time.Now().Add(duration))
		}
	}
	for k, v := range options.PrivateClaims {
		builder.Claim(k, v)
	}

	token, err := builder.Build()
	if err != nil {
		panic(err)
	}

	signed, err := jwt.Sign(token, options.createWithKey())
	if err != nil {
		panic(err)
	}

	return string(signed)
}

func (options JWTOptions) createWithKey() jwt.SignEncryptParseOption {
	if options.Algorithm == "" {
		options.Algorithm = "HS256"
	}

	var alg jwa.SignatureAlgorithm
	switch options.Algorithm {
	case "HS256":
		alg = jwa.HS256()
	case "HS384":
		alg = jwa.HS384()
	case "HS512":
		alg = jwa.HS512()
	default:
		panic("unsupported algorithm: " + options.Algorithm)
	}

	return jwt.WithKey(alg, []byte(options.Secret))
}

func (j *Jwt) Parse(serialized string, opts map[string]interface{}) map[string]interface{} {
	var options JWTOptions
	b, err := json.Marshal(opts)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &options)
	if err != nil {
		panic(err)
	}

	token, err := jwt.ParseString(serialized, options.createWithKey())
	if err != nil {
		panic(err)
	}

	out, err := json.Marshal(token)
	if err != nil {
		panic(err)
	}

	var payload map[string]interface{}
	err = json.Unmarshal(out, &payload)
	if err != nil {
		panic(err)
	}

	return payload
}
