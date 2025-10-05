package builtin

import (
	"encoding/json"
	"time"
	"errors"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// Jwt provides methods to sign and parse JSON Web Tokens (JWT).
type Jwt struct {
}

// NewJwt is mapped to the built-in function `jwt`.
func NewJwt() *Jwt {
	return &Jwt{}
}

// JWTOptions represents options for JWT signing and parsing.
// It has a structure that allows defining Registered Claim Names, some Public Claim Names, and Private Claim Names.
// During actual signing and parsing, JSON data for the key names defined in the structure tag is specified.
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

// Sign generates a signed JWT string with the specified options.
// The options are accepted as an argument in the form of a map of JSON data defined by the struct tag of the JWTOptions structure.
func (j *Jwt) Sign(opts map[string]any) (string, error) {
	var opt *JWTOptions
	b, err := json.Marshal(opts)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(b, &opt)
	if err != nil {
		return "", err
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
		return "", err
	}

	signOption, err := opt.createWithKey()
	if err != nil {
		return "", err
	}

	signed, err := jwt.Sign(token, signOption)
	if err != nil {
		return string(signed), err
	}

	return string(signed), nil
}

// createWithKey creates a jwt.SignEncryptParseOption based on the JWTOptions.
func (opt *JWTOptions) createWithKey() (jwt.SignEncryptParseOption, error) {
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
		return nil, errors.New("unsupported algorithm: " + opt.Algorithm)
	}

	return jwt.WithKey(alg, []byte(opt.Secret)), nil
}

// Parse validates and parses a JWT string serialized with the specified options.
// If the signature is correctly validated and parsed, the payload map is returned.
func (j *Jwt) Parse(serialized string, opts map[string]any) (map[string]any, error) {
	var opt *JWTOptions
	b, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &opt)
	if err != nil {
		return nil, err
	}

	parseOption, err := opt.createWithKey()
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseString(serialized, parseOption)
	if err != nil {
		return nil, err
	}

	out, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}

	var payload map[string]any
	err = json.Unmarshal(out, &payload)
	if err != nil {
		return nil, err
	}

	return payload, nil	
}
