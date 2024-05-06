package requests

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/stretchr/testify/assert"
)

func TestValidateRequestSchema(t *testing.T) {
	for name, tc := range map[string]struct {
		request                    *http.Request
		schema                     *base.Schema
		renderedSchema, jsonSchema []byte
		assertValidRequestSchema   assert.BoolAssertionFunc
		expectedErrorsCount        int
	}{
		"FailOnBooleanExclusiveMinimum": {
			request: postRequestWithBody(`{"exclusiveNumber": 13}`),
			schema: &base.Schema{
				Type: []string{"object"},
			},
			renderedSchema: []byte(`type: object
properties:
    exclusiveNumber:
        type: number
        description: This number starts its journey where most numbers are too scared to begin!
        exclusiveMinimum: true
        minimum: !!float 10`),
			jsonSchema:               []byte(`{"properties":{"exclusiveNumber":{"description":"This number starts its journey where most numbers are too scared to begin!","exclusiveMinimum":true,"minimum":10,"type":"number"}},"type":"object"}`),
			assertValidRequestSchema: assert.False,
			expectedErrorsCount:      1,
		},
		"PassWithCorrectExclusiveMinimum": {
			request: postRequestWithBody(`{"exclusiveNumber": 15}`),
			schema: &base.Schema{
				Type: []string{"object"},
			},
			renderedSchema: []byte(`type: object
properties:
    exclusiveNumber:
        type: number
        description: This number is properly constrained by a numeric exclusive minimum.
        exclusiveMinimum: 12
        minimum: 12`),
			jsonSchema:               []byte(`{"properties":{"exclusiveNumber":{"type":"number","description":"This number is properly constrained by a numeric exclusive minimum.","exclusiveMinimum":12,"minimum":12}},"type":"object"}`),
			assertValidRequestSchema: assert.True,
			expectedErrorsCount:      0,
		},
		"PassWithValidStringType": {
			request: postRequestWithBody(`{"greeting": "Hello, world!"}`),
			schema: &base.Schema{
				Type: []string{"object"},
			},
			renderedSchema: []byte(`type: object
properties:
    greeting:
        type: string
        description: A simple greeting
        example: "Hello, world!"`),
			jsonSchema:               []byte(`{"properties":{"greeting":{"type":"string","description":"A simple greeting","example":"Hello, world!"}},"type":"object"}`),
			assertValidRequestSchema: assert.True,
			expectedErrorsCount:      0,
		},
	} {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			valid, errors := ValidateRequestSchema(tc.request, tc.schema, tc.renderedSchema, tc.jsonSchema)

			tc.assertValidRequestSchema(t, valid)
			assert.Len(t, errors, tc.expectedErrorsCount)
		})
	}
}

func postRequestWithBody(payload string) *http.Request {
	return &http.Request{
		Method: http.MethodPost,
		Body:   io.NopCloser(strings.NewReader(payload)),
	}
}
