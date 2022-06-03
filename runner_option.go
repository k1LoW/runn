package runn

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// RunnerConfig is polymorphic config for runner
type RunnerConfig struct {
	// for httpRunner
	Endpoint             string `yaml:"endpoint,omitempty"`
	OpenApi3DocLocation  string `yaml:"openapi3,omitempty"`
	SkipValidateRequest  bool   `yaml:"skipValidateRequest,omitempty"`
	SkipValidateResponse bool   `yaml:"skipValidateResponse,omitempty"`

	openApi3Doc *openapi3.T
}

type RunnerOption func(*RunnerConfig) error

func OpenApi3(l string) RunnerOption {
	return func(c *RunnerConfig) error {
		c.OpenApi3DocLocation = l
		return nil
	}
}

func OpenApi3FromData(d []byte) RunnerOption {
	return func(c *RunnerConfig) error {
		ctx := context.Background()
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData(d)
		if err != nil {
			return err
		}
		if err := doc.Validate(ctx); err != nil {
			return fmt.Errorf("openapi document validation error: %w", err)
		}
		c.openApi3Doc = doc
		return nil
	}
}

func SkipValidateRequest(skip bool) RunnerOption {
	return func(c *RunnerConfig) error {
		c.SkipValidateRequest = skip
		return nil
	}
}

func SkipValidateResponse(skip bool) RunnerOption {
	return func(c *RunnerConfig) error {
		c.SkipValidateResponse = skip
		return nil
	}
}
