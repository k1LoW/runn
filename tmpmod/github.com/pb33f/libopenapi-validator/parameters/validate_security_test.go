// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package parameters

import (
	"github.com/pb33f/libopenapi"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/paths"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestParamValidator_ValidateSecurity_APIKeyHeader_NotFound(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /products:
    post:
      security:
        - ApiKeyAuth:
          - write:products
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/products", nil)

	valid, errors := v.ValidateSecurity(request)
	assert.False(t, valid)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "API Key X-API-Key not found in header", errors[0].Message)
	assert.Equal(t, request.Method, errors[0].RequestMethod)
	assert.Equal(t, request.URL.Path, errors[0].RequestPath)
	assert.Equal(t, "/products", errors[0].SpecPath)
}

func TestParamValidator_ValidateSecurity_APIKeyHeader(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /products:
    post:
      security:
        - ApiKeyAuth:
          - write:products
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/products", nil)
	request.Header.Add("X-API-Key", "1234")

	valid, errors := v.ValidateSecurity(request)
	assert.True(t, valid)
	assert.Equal(t, 0, len(errors))
}

func TestParamValidator_ValidateSecurity_APIKeyQuery_NotFound(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /products:
    post:
      security:
        - ApiKeyAuth:
          - write:products
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: query
      name: X-API-Key
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/products", nil)

	valid, errors := v.ValidateSecurity(request)
	assert.False(t, valid)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "API Key X-API-Key not found in query", errors[0].Message)
	assert.Equal(t, "Add an API Key via 'X-API-Key' to the query string of the URL, "+
		"for example 'https://things.com/products?X-API-Key=your-api-key'", errors[0].HowToFix)
	assert.Equal(t, request.Method, errors[0].RequestMethod)
	assert.Equal(t, request.URL.Path, errors[0].RequestPath)
	assert.Equal(t, "/products", errors[0].SpecPath)

}

func TestParamValidator_ValidateSecurity_APIKeyQuery(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /products:
    post:
      security:
        - ApiKeyAuth:
          - write:products
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: query
      name: X-API-Key
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/products?X-API-Key=12345", nil)

	valid, errors := v.ValidateSecurity(request)
	assert.True(t, valid)
	assert.Equal(t, 0, len(errors))
}

func TestParamValidator_ValidateSecurity_APIKeyCookie_NotFound(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /products:
    post:
      security:
        - ApiKeyAuth:
          - write:products
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: cookie
      name: X-API-Key
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/products", nil)

	valid, errors := v.ValidateSecurity(request)
	assert.False(t, valid)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "API Key X-API-Key not found in cookies", errors[0].Message)
	assert.Equal(t, request.Method, errors[0].RequestMethod)
	assert.Equal(t, request.URL.Path, errors[0].RequestPath)
	assert.Equal(t, "/products", errors[0].SpecPath)
}

func TestParamValidator_ValidateSecurity_APIKeyCookie(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /products:
    post:
      security:
        - ApiKeyAuth:
          - write:products
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: cookie
      name: X-API-Key
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/products", nil)

	request.AddCookie(&http.Cookie{
		Name:  "X-API-Key",
		Value: "1234",
	})

	valid, errors := v.ValidateSecurity(request)
	assert.True(t, valid)
	assert.Equal(t, 0, len(errors))
}

func TestParamValidator_ValidateSecurity_Basic_NotFound(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /products:
    post:
      security:
        - ApiKeyAuth:
          - write:products
components:
  securitySchemes:
    ApiKeyAuth:
      type: http
      scheme: basic
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/products", nil)

	valid, errors := v.ValidateSecurity(request)
	assert.False(t, valid)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "Authorization header for 'basic' scheme", errors[0].Message)
	assert.Equal(t, request.Method, errors[0].RequestMethod)
	assert.Equal(t, request.URL.Path, errors[0].RequestPath)
	assert.Equal(t, "/products", errors[0].SpecPath)
}

func TestParamValidator_ValidateSecurity_Basic(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /products:
    post:
      security:
        - ApiKeyAuth:
          - write:products
components:
  securitySchemes:
    ApiKeyAuth:
      type: http
      scheme: basic
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/products", nil)
	request.Header.Add("Authorization", "Basic 1234")

	valid, errors := v.ValidateSecurity(request)
	assert.True(t, valid)
	assert.Equal(t, 0, len(errors))
}

func TestParamValidator_ValidateSecurity_BadPath(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /products:
    post:
      security:
        - ApiKeyAuth:
          - write:products
components:
  securitySchemes:
    ApiKeyAuth:
      type: http
      scheme: basic
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/blimpo", nil)
	valid, errors := v.ValidateSecurity(request)
	assert.False(t, valid)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, request.Method, errors[0].RequestMethod)
	assert.Equal(t, request.URL.Path, errors[0].RequestPath)
	assert.Equal(t, "", errors[0].SpecPath)
}

func TestParamValidator_ValidateSecurity_MissingSecuritySchemes(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /products:
    post:
      security:
        - ApiKeyAuth:
          - write:products
components: {}
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/products", nil)
	valid, errors := v.ValidateSecurity(request)
	assert.False(t, valid)
	assert.Equal(t, 1, len(errors))
}

func TestParamValidator_ValidateSecurity_NoComponents(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /products:
    post:
      security:
        - ApiKeyAuth:
          - write:products
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/products", nil)
	valid, errors := v.ValidateSecurity(request)
	assert.False(t, valid)
	assert.Equal(t, 1, len(errors))
}

func TestParamValidator_ValidateSecurity_PresetPath(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /products:
    post:
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/products", nil)
	pathItem, errs, _ := paths.FindPath(request, &m.Model)
	assert.Nil(t, errs)
	v.(*paramValidator).pathItem = pathItem

	valid, errors := v.ValidateSecurity(request)
	assert.True(t, valid)
	assert.Equal(t, 0, len(errors))
}
