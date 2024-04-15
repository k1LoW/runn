package runn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
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

type openAPI3Validator struct {
	skipValidateRequest  bool
	skipValidateResponse bool
	doc                  *libopenapi.Document
	validator            *validator.Validator
}

func newOpenAPI3Validator(c *httpRunnerConfig) (*openAPI3Validator, error) {
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
			doc, err = libopenapi.NewDocumentWithConfiguration(b, openAPIConfig)
			if err != nil {
				return nil, err
			}
		default:
			b, err := os.ReadFile(l)
			if err != nil {
				return nil, err
			}
			openAPIConfig.BasePath = filepath.Dir(l)
			doc, err = libopenapi.NewDocumentWithConfiguration(b, openAPIConfig)
			if err != nil {
				return nil, err
			}
		}
		v, errs := validator.NewValidator(doc)
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
		c.openAPI3Doc = &doc
		c.openAPI3Validator = &v
	}
	if c.openAPI3Doc == nil {
		return nil, errors.New("cannot load openapi3 document")
	}
	return &openAPI3Validator{
		skipValidateRequest:  c.SkipValidateRequest,
		skipValidateResponse: c.SkipValidateResponse,
		doc:                  c.openAPI3Doc,
		validator:            c.openAPI3Validator,
	}, nil
}

func (v *openAPI3Validator) ValidateRequest(ctx context.Context, req *http.Request) error {
	if v.skipValidateRequest {
		return nil
	}
	vv := *v.validator
	_, errs := vv.ValidateHttpRequest(req)
	if len(errs) > 0 {
		var err error
		for _, e := range errs {
			// nullable type workaround
			if len(e.SchemaValidationErrors) > 0 && strings.HasSuffix(e.SchemaValidationErrors[0].Reason, "but got null") && strings.HasSuffix(e.SchemaValidationErrors[0].Location, "/type") {
				if nullableType(e.SchemaValidationErrors[0].ReferenceSchema, e.SchemaValidationErrors[0].Location) {
					continue
				}
			}
			err = errors.Join(err, e)
		}
		if err == nil {
			return nil
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
	vv := *v.validator
	_, errs := vv.ValidateHttpResponse(req, res)
	if len(errs) > 0 {
		var err error
		for _, e := range errs {
			// nullable type workaround
			if len(e.SchemaValidationErrors) > 0 && strings.HasSuffix(e.SchemaValidationErrors[0].Reason, "but got null") && strings.HasSuffix(e.SchemaValidationErrors[0].Location, "/type") {
				if nullableType(e.SchemaValidationErrors[0].ReferenceSchema, e.SchemaValidationErrors[0].Location) {
					continue
				}
			}
			err = errors.Join(err, e)
		}
		if err == nil {
			return nil
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

// nullableType
func nullableType(schema, location string) bool {
	splitted := strings.Split(strings.TrimPrefix(strings.TrimSuffix(location, "/type")+"/nullable", "/"), "/")
	m := map[string]any{}
	if err := yaml.Unmarshal([]byte(schema), &m); err != nil {
		return false
	}
	v, ok := getFromMap(m, splitted...)
	if !ok {
		return false
	}
	if tf, ok := v.(bool); ok {
		return tf
	}
	return false
}

func getFromMap(m map[string]any, keys ...string) (any, bool) {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if len(keys) == 1 {
				return v, true
			}
			mm, ok := v.(map[string]any)
			if !ok {
				return nil, false
			}
			return getFromMap(mm, keys[1:]...)
		}
		return nil, false
	}
	return nil, false
}
