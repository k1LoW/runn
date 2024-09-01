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
	"strconv"
	"strings"
	"sync"

	"github.com/goccy/go-yaml"
	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
	verrors "github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/libopenapi/datamodel"
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
var globalOpenAPI3DocRegistory = map[string]libopenapi.Document{}
var globalOpenAPI3DocRegistoryMu sync.RWMutex

type openAPI3Validator struct {
	skipValidateRequest  bool
	skipValidateResponse bool
	doc                  libopenapi.Document
	validator            validator.Validator
	mu                   sync.Mutex
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
			globalOpenAPI3DocRegistoryMu.RLock()
			od, ok := globalOpenAPI3DocRegistory[hash]
			globalOpenAPI3DocRegistoryMu.RUnlock()
			if ok {
				v, errs := validator.NewValidator(od)
				if len(errs) > 0 {
					return nil, errors.Join(errs...)
				}
				return &openAPI3Validator{
					skipValidateRequest:  c.SkipValidateRequest,
					skipValidateResponse: c.SkipValidateResponse,
					doc:                  od,
					validator:            v,
				}, nil
			}
			oc := &datamodel.DocumentConfiguration{
				AllowFileReferences:        true,
				AllowRemoteReferences:      true,
				SkipCircularReferenceCheck: c.SkipCircularReferenceCheck,
			}
			doc, err = libopenapi.NewDocumentWithConfiguration(b, oc)
			if err != nil {
				return nil, err
			}
		default:
			b, err := os.ReadFile(l)
			if err != nil {
				return nil, err
			}
			hash = hashBytes(b)
			globalOpenAPI3DocRegistoryMu.RLock()
			od, ok := globalOpenAPI3DocRegistory[hash]
			globalOpenAPI3DocRegistoryMu.RUnlock()
			if ok {
				v, errs := validator.NewValidator(od)
				if len(errs) > 0 {
					return nil, errors.Join(errs...)
				}
				return &openAPI3Validator{
					skipValidateRequest:  c.SkipValidateRequest,
					skipValidateResponse: c.SkipValidateResponse,
					doc:                  od,
					validator:            v,
				}, nil
			}
			oc := &datamodel.DocumentConfiguration{
				AllowFileReferences:        true,
				AllowRemoteReferences:      true,
				SkipCircularReferenceCheck: c.SkipCircularReferenceCheck,
				BasePath:                   filepath.Dir(l),
			}
			doc, err = libopenapi.NewDocumentWithConfiguration(b, oc)
			if err != nil {
				return nil, err
			}
		}
		c.openAPI3Doc = doc
	}

	v, errs := validator.NewValidator(c.openAPI3Doc)
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

	globalOpenAPI3DocRegistoryMu.Lock()
	globalOpenAPI3DocRegistory[hash] = c.openAPI3Doc
	globalOpenAPI3DocRegistoryMu.Unlock()

	return &openAPI3Validator{
		skipValidateRequest:  c.SkipValidateRequest,
		skipValidateResponse: c.SkipValidateResponse,
		doc:                  c.openAPI3Doc,
		validator:            v,
	}, nil
}

func (v *openAPI3Validator) ValidateRequest(ctx context.Context, req *http.Request) error {
	if v.skipValidateRequest {
		return nil
	}
	v.mu.Lock()
	_, errs := v.validator.ValidateHttpRequest(req)
	if len(errs) == 0 {
		v.mu.Unlock()
		return nil
	}
	{
		// renew validator (workaround)
		// ref: https://github.com/k1LoW/runn/issues/882
		vv, errrs := validator.NewValidator(v.doc)
		if len(errrs) > 0 {
			return errors.Join(errrs...)
		}
		v.validator = vv
	}
	v.mu.Unlock()
	var err error
	for _, e := range errs {
		// nullable type workaround.
		if nullableError(e) {
			continue
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

func (v *openAPI3Validator) ValidateResponse(ctx context.Context, req *http.Request, res *http.Response) error {
	if v.skipValidateResponse {
		return nil
	}
	v.mu.Lock()
	_, errs := v.validator.ValidateHttpResponse(req, res)
	if len(errs) == 0 {
		v.mu.Unlock()
		return nil
	}
	{
		// renew validator (workaround)
		// ref: https://github.com/k1LoW/runn/issues/882
		vv, errrs := validator.NewValidator(v.doc)
		if len(errrs) > 0 {
			return errors.Join(errrs...)
		}
		v.validator = vv
	}
	v.mu.Unlock()
	var err error
	for _, e := range errs {
		// nullable type workaround.
		if nullableError(e) {
			continue
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

func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// nullableTypeError returns whether the error is nullable type error or not.
func nullableError(e *verrors.ValidationError) bool {
	if len(e.SchemaValidationErrors) > 0 {
		for _, ve := range e.SchemaValidationErrors {
			if strings.HasSuffix(ve.Reason, "but got null") && strings.HasSuffix(ve.Location, "/type") {
				if nullableType(ve.ReferenceSchema, ve.Location) {
					return true
				}
			}
		}
	}
	return false
}

// nullableType returns whether the type is nullable or not.
func nullableType(schema, location string) bool {
	splitted := strings.Split(strings.TrimPrefix(strings.TrimSuffix(location, "/type")+"/nullable", "/"), "/")
	m := map[string]any{}
	if err := yaml.Unmarshal([]byte(schema), &m); err != nil {
		return false
	}
	v, ok := valueWithKeys(m, splitted...)
	if !ok {
		return false
	}
	if tf, ok := v.(bool); ok {
		return tf
	}
	return false
}

func valueWithKeys(m any, keys ...string) (any, bool) {
	if len(keys) == 0 {
		return nil, false
	}
	switch m := m.(type) {
	case map[string]any:
		if v, ok := m[keys[0]]; ok {
			if len(keys) == 1 {
				return v, true
			}
			return valueWithKeys(v, keys[1:]...)
		}
	case []any:
		i, err := strconv.Atoi(keys[0])
		if err != nil {
			return nil, false
		}
		if i < 0 || i >= len(m) {
			return nil, false
		}
		v := m[i]
		if len(keys) == 1 {
			return v, true
		}
		return valueWithKeys(v, keys[1:]...)
	default:
		return nil, false
	}
	return nil, false
}
