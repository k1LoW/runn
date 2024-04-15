package runn

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
)

type httpValidator interface { //nostyle:ifacenames
	ValidateRequest(ctx context.Context, req *http.Request) error
	ValidateResponse(ctx context.Context, req *http.Request, res *http.Response) error
}

type UnsupportedError struct {
	Cause error
}

func (e *UnsupportedError) Error() string {
	return e.Cause.Error()
}

func (e *UnsupportedError) Unwrap() error {
	return e.Cause
}

func newHttpValidator(c *httpRunnerConfig) (httpValidator, error) {
	if c.OpenAPI3DocLocation != "" || c.openAPI3Doc != nil {
		return newOpenAPI3Validator(c)
	}
	return newNopValidator(), nil
}

type nopValidator struct{}

func (v *nopValidator) ValidateRequest(ctx context.Context, req *http.Request) error {
	return nil
}

func (v *nopValidator) ValidateResponse(ctx context.Context, req *http.Request, res *http.Response) error {
	return nil
}

func newNopValidator() *nopValidator {
	return &nopValidator{}
}

// globalOpenAPI3DocRegistory - global registory of OpenAPI3 documents.
var globalOpenAPI3DocRegistory = map[string]*openAPI3Doc{}

type openAPI3Doc struct {
	doc       *libopenapi.Document
	validator *validator.Validator
}

type openAPI3Validator struct {
	skipValidateRequest  bool
	skipValidateResponse bool
	doc                  *openAPI3Doc
}

func newOpenAPI3Validator(c *httpRunnerConfig) (*openAPI3Validator, error) {
	if c.OpenAPI3DocLocation == "" && c.openAPI3Doc == nil {
		return nil, errors.New("cannot load openapi3 document")
	}

	var hash string

	if c.OpenAPI3DocLocation != "" {
		l := c.OpenAPI3DocLocation
		var doc libopenapi.Document
		switch {
		case strings.HasPrefix(l, "https://") || strings.HasPrefix(l, "http://"):
			u, err := url.Parse(l)
			if err != nil {
				return nil, err
			}
			res, err := http.Get(u.String())
			if err != nil {
				return nil, err
			}
			defer res.Body.Close()
			b, err := io.ReadAll(res.Body)
			if err != nil {
				return nil, err
			}
			hash = hashBytes(b)
			od, ok := globalOpenAPI3DocRegistory[hash]
			if ok {
				return &openAPI3Validator{
					skipValidateRequest:  c.SkipValidateRequest,
					skipValidateResponse: c.SkipValidateResponse,
					doc:                  od,
				}, nil
			}
			doc, err = libopenapi.NewDocumentWithConfiguration(b, openAPIConfig)
			if err != nil {
				return nil, err
			}
		default:
			b, err := os.ReadFile(l)
			if err != nil {
				return nil, err
			}
			hash = hashBytes(b)
			od, ok := globalOpenAPI3DocRegistory[hash]
			if ok {
				return &openAPI3Validator{
					skipValidateRequest:  c.SkipValidateRequest,
					skipValidateResponse: c.SkipValidateResponse,
					doc:                  od,
				}, nil
			}
			openAPIConfig.BasePath = filepath.Dir(l)
			doc, err = libopenapi.NewDocumentWithConfiguration(b, openAPIConfig)
			if err != nil {
				return nil, err
			}
		}
		c.openAPI3Doc = &doc
	}

	v, errs := validator.NewValidator(*c.openAPI3Doc)
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	if _, errs := v.ValidateDocument(); len(errs) > 0 {
		var err error
		for _, e := range errs {
			err = errors.Join(err, e)
		}
		return nil, err
	}

	doc := &openAPI3Doc{
		doc:       c.openAPI3Doc,
		validator: &v,
	}

	globalOpenAPI3DocRegistory[hash] = doc

	return &openAPI3Validator{
		skipValidateRequest:  c.SkipValidateRequest,
		skipValidateResponse: c.SkipValidateResponse,
		doc:                  doc,
	}, nil
}

func (v *openAPI3Validator) ValidateRequest(ctx context.Context, req *http.Request) error {
	if v.skipValidateRequest {
		return nil
	}
	vv := *v.doc.validator
	_, errs := vv.ValidateHttpRequest(req)
	if len(errs) > 0 {
		var err error
		for _, e := range errs {
			err = errors.Join(err, e)
		}
		b, errr := httputil.DumpRequest(req, true)
		if errr != nil {
			return fmt.Errorf("runn error: %w", errr)
		}
		return fmt.Errorf("openapi3 validation error: %w\n-----START HTTP REQUEST-----\n%s\n-----END HTTP REQUEST-----\n", err, string(b))
	}
	return nil
}

func (v *openAPI3Validator) ValidateResponse(ctx context.Context, req *http.Request, res *http.Response) error {
	if v.skipValidateResponse {
		return nil
	}
	vv := *v.doc.validator
	_, errs := vv.ValidateHttpResponse(req, res)
	if len(errs) > 0 {
		var err error
		for _, e := range errs {
			err = errors.Join(err, e)
		}
		b, errr := httputil.DumpRequest(req, true)
		if errr != nil {
			return fmt.Errorf("runn error: %w", errr)
		}
		b2, errr := httputil.DumpResponse(res, true)
		if errr != nil {
			return fmt.Errorf("runn error: %w", errr)
		}
		return fmt.Errorf("openapi3 validation error: %w\n-----START HTTP REQUEST-----\n%s\n-----END HTTP REQUEST-----\n-----START HTTP RESPONSE-----\n%s\n-----END HTTP RESPONSE-----\n", err, string(b), string(b2))
	}
	return nil
}

func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
