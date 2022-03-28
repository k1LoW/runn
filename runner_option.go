package runn

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// RunnerConfig is polymorphic config for runner
type RunnerConfig struct {
	Endpoint             string `yaml:"endpoint,omitempty"`
	OpenApi3DocPath      string `yaml:"openapi3,omitempty"`
	Prefix               string `yaml:"prefix,omitempty"`
	SkipValidateRequest  bool   `yaml:"skipValidateRequest,omitempty"`
	SkipValidateResponse bool   `yaml:"skipValidateResponse,omitempty"`

	openApi3Doc *openapi3.T
}

type RunnerOption func(*RunnerConfig) error

func RunnerOpenApi3(path string) RunnerOption {
	return func(c *RunnerConfig) error {
		c.OpenApi3DocPath = path
		ctx := context.Background()
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromFile(path)
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

func RunnerOpenApi3FromData(d []byte) RunnerOption {
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

func RunnerPrefix(prefix string) RunnerOption {
	return func(c *RunnerConfig) error {
		c.Prefix = prefix
		return nil
	}
}

func RunnerSkipValidateRequest(skip bool) RunnerOption {
	return func(c *RunnerConfig) error {
		c.SkipValidateRequest = skip
		return nil
	}
}

func RunnerSkipValidateResponse(skip bool) RunnerOption {
	return func(c *RunnerConfig) error {
		c.SkipValidateResponse = skip
		return nil
	}
}
