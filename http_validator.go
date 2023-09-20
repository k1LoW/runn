package runn

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
)

type httpValidator interface {
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
	if c.OpenApi3DocLocation != "" || c.openApi3Doc != nil {
		return newOpenApi3Validator(c)
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

type openApi3Validator struct {
	skipValidateRequest  bool
	skipValidateResponse bool
	doc                  *openapi3.T
}

func newOpenApi3Validator(c *httpRunnerConfig) (*openApi3Validator, error) {
	if c.OpenApi3DocLocation != "" {
		l := c.OpenApi3DocLocation
		ctx := context.Background()
		loader := openapi3.NewLoader()
		var doc *openapi3.T
		switch {
		case strings.HasPrefix(l, "https://") || strings.HasPrefix(l, "http://"):
			u, err := url.Parse(l)
			if err != nil {
				return nil, err
			}
			doc, err = loader.LoadFromURI(u)
			if err != nil {
				return nil, err
			}
		default:
			b, err := readFile(l)
			if err != nil {
				return nil, err
			}
			doc, err = loader.LoadFromData(b)
			if err != nil {
				return nil, err
			}
		}

		if err := doc.Validate(ctx); err != nil {
			return nil, fmt.Errorf("openapi3 document validation error: %w", err)
		}
		c.openApi3Doc = doc
	}

	if c.openApi3Doc == nil {
		return nil, errors.New("cannot load openapi3 document")
	}
	return &openApi3Validator{
		skipValidateRequest:  c.SkipValidateRequest,
		skipValidateResponse: c.SkipValidateResponse,
		doc:                  c.openApi3Doc,
	}, nil
}

// FIXME: better to depend on any library
// currently refer to https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types
var registerBodyMimeTypes = []string{
	"text/css", "text/html", "text/csv", "text/xml", "text/javascript",
	"image/apng", "image/avif", "image/gif", "image/jpeg", "image/png", "image/svg+xml", "image/webp",
	"audio/wave", "audio/wav", "audio/x-wav", "audio/x-pn-wav", "audio/webm", "audio/ogg", "audio/mpeg", "audio/vorbis",
	"video/webm", "video/ogg", "video/mp4",
	"application/pdf", "application/pkcs8", "application/zip", "application/wasm", "application/ogg",
	"font/woff", "font/ttf", "font/otf",
	"model/3mf", "model/vrml",
}

func (v *openApi3Validator) ValidateRequest(ctx context.Context, req *http.Request) error {
	if v.skipValidateRequest {
		return nil
	}
	input, err := v.requestInput(req)
	if err != nil {
		return err
	}
	if err := openapi3filter.ValidateRequest(ctx, input); err != nil {
		b, errr := httputil.DumpRequest(req, true)
		if errr != nil {
			return fmt.Errorf("runn error: %w", errr)
		}
		return fmt.Errorf("openapi3 validation error: %w\n-----START HTTP REQUEST-----\n%s\n-----END HTTP REQUEST-----\n", err, string(b))
	}
	return nil
}

func (v *openApi3Validator) requestInput(req *http.Request) (*openapi3filter.RequestValidationInput, error) {
	// skip scheme://host:port validation
	for _, server := range v.doc.Servers {
		su, err := url.Parse(server.URL)
		if err != nil {
			return nil, err
		}
		su.Host = req.URL.Host
		su.Opaque = req.URL.Opaque
		su.Scheme = req.URL.Scheme
		server.URL = su.String()
	}
	router, err := legacyrouter.NewRouter(v.doc)
	if err != nil {
		return nil, err
	}

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		return nil, fmt.Errorf("failed to find route: %w (%s %s)", err, req.Method, req.URL.Path)
	}
	return &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		Options: &openapi3filter.Options{
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	}, nil
}

func (v *openApi3Validator) responseInput(req *http.Request, res *http.Response) (*openapi3filter.ResponseValidationInput, error) {
	reqInput, err := v.requestInput(req)
	if err != nil {
		return nil, err
	}
	var body io.ReadCloser
	if res.Body != nil {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewBuffer(b))
		body = io.NopCloser(bytes.NewBuffer(b))
	}
	return &openapi3filter.ResponseValidationInput{
		RequestValidationInput: reqInput,
		Status:                 res.StatusCode,
		Header:                 res.Header,
		Body:                   body,
		Options:                &openapi3filter.Options{IncludeResponseStatus: true},
	}, nil
}

func (v *openApi3Validator) ValidateResponse(ctx context.Context, req *http.Request, res *http.Response) error {
	if v.skipValidateResponse {
		return nil
	}
	input, err := v.responseInput(req, res)
	if err != nil {
		return err
	}

	err = openapi3filter.ValidateResponse(ctx, input)

	if err != nil {
		var parseError *openapi3filter.ParseError
		var responseError *openapi3filter.ResponseError
		switch {
		case errors.As(err, &parseError):
			if parseError.Kind == openapi3filter.KindUnsupportedFormat {
				return &UnsupportedError{Cause: err}
			}
		case errors.As(err, &responseError):
			if errors.As(responseError.Err, &parseError) {
				if parseError.Kind == openapi3filter.KindUnsupportedFormat {
					return &UnsupportedError{Cause: err}
				}
			}
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

func init() {
	for _, mime := range registerBodyMimeTypes {
		openapi3filter.RegisterBodyDecoder(mime, openapi3filter.FileBodyDecoder)
	}
}
