package runn

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
)

type httpValidator interface {
	ValidateRequest(ctx context.Context, req *http.Request) error
	ValidateResponse(ctx context.Context, req *http.Request, res *http.Response) error
}

func NewHttpValidator(c *RunnerConfig) (httpValidator, error) {
	switch {
	case c.OpenApi3DocLocation != "" || c.openApi3Doc != nil:
		return NewOpenApi3Validator(c)
	default:
		if c.Endpoint == "" {
			return nil, errors.New("can not create http validator")
		}
	}
	return NewNopValidator(), nil
}

type nopValidator struct{}

func (v *nopValidator) ValidateRequest(ctx context.Context, req *http.Request) error {
	return nil
}

func (v *nopValidator) ValidateResponse(ctx context.Context, req *http.Request, res *http.Response) error {
	return nil
}

func NewNopValidator() *nopValidator {
	return &nopValidator{}
}

type openApi3Validator struct {
	prefix               string
	skipValidateRequest  bool
	skipValidateResponse bool
	router               routers.Router
}

func NewOpenApi3Validator(c *RunnerConfig) (*openApi3Validator, error) {
	if c.OpenApi3DocLocation != "" {
		l := c.OpenApi3DocLocation
		ctx := context.Background()
		loader := openapi3.NewLoader()
		var (
			doc *openapi3.T
			err error
		)
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
			doc, err = loader.LoadFromFile(l)
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
		return nil, errors.New("can not load openapi3 document")
	}
	router, err := legacyrouter.NewRouter(c.openApi3Doc)
	if err != nil {
		return nil, err
	}
	return &openApi3Validator{
		prefix:               c.Prefix,
		skipValidateRequest:  c.SkipValidateRequest,
		skipValidateResponse: c.SkipValidateResponse,
		router:               router,
	}, nil
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
		return err
	}
	return nil
}

func (v *openApi3Validator) requestInput(req *http.Request) (*openapi3filter.RequestValidationInput, error) {
	route, pathParams, err := v.router.FindRoute(req)
	if err != nil {
		return nil, err
	}
	return &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}, nil
}

func (v *openApi3Validator) responseInput(req *http.Request, res *http.Response) (*openapi3filter.ResponseValidationInput, error) {
	reqInput, err := v.requestInput(req)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	res.Body = io.NopCloser(bytes.NewBuffer(b))
	return &openapi3filter.ResponseValidationInput{
		RequestValidationInput: reqInput,
		Status:                 res.StatusCode,
		Header:                 res.Header,
		Body:                   io.NopCloser(bytes.NewBuffer(b)),
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
	if err := openapi3filter.ValidateResponse(ctx, input); err != nil {
		return err
	}
	return nil
}
