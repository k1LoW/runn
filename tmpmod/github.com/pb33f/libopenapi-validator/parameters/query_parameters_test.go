// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package parameters

import (
	"net/http"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator_QueryParamMissing(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "Query parameter 'fishy' is missing", errors[0].Message)
	assert.Equal(t, request.Method, errors[0].RequestMethod)
	assert.Equal(t, request.URL.Path, errors[0].RequestPath)
	assert.Equal(t, "/a/fishy/on/a/dishy", errors[0].SpecPath)
}

func TestNewValidator_QueryParamNotMissing(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=cod", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Nil(t, errors)
}

func TestNewValidator_QueryParamMinimum_violation(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: string
            minLength: 4
      operationId: locateFishy
`

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)
	m, errs := doc.BuildV3Model()
	require.Len(t, errs, 0)

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=cod", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "Query parameter 'fishy' failed to validate", errors[0].Message)
}

func TestNewValidator_QueryParamMinimum(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: string
            minLength: 4
      operationId: locateFishy
`

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)
	m, errs := doc.BuildV3Model()
	require.Len(t, errs, 0)

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=salmon", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Nil(t, errors)
}

func TestNewValidator_QueryParamPost(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    post:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/a/fishy/on/a/dishy?fishy=cod", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Nil(t, errors)
}

func TestNewValidator_QueryParamPut(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    put:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPut, "https://things.com/a/fishy/on/a/dishy?fishy=cod", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Nil(t, errors)
}

func TestNewValidator_QueryParamDelete(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    delete:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodDelete, "https://things.com/a/fishy/on/a/dishy?fishy=cod", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Nil(t, errors)
}

func TestNewValidator_QueryParamOptions(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    options:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodOptions, "https://things.com/a/fishy/on/a/dishy?fishy=cod", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Nil(t, errors)
}

func TestNewValidator_QueryParamHead(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    head:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodHead, "https://things.com/a/fishy/on/a/dishy?fishy=cod", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Nil(t, errors)
}

func TestNewValidator_QueryParamPatch(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    patch:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPatch, "https://things.com/a/fishy/on/a/dishy?fishy=cod", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Nil(t, errors)
}

func TestNewValidator_QueryParamTrace(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    trace:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodTrace, "https://things.com/a/fishy/on/a/dishy?fishy=cod", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Nil(t, errors)
}

func TestNewValidator_QueryParamBadPath(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: number
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/Not/Found/dishy?fishy=cod", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.NotNil(t, errors)
}

func TestNewValidator_QueryParamWrongTypeNumber(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: number
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=cod", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.NotNil(t, errors)
	assert.Equal(t, "Query parameter 'fishy' is not a valid number", errors[0].Message)
}

func TestNewValidator_QueryParamValidTypeNumber(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: number
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=123", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Nil(t, errors)
}

func TestNewValidator_QueryParamMinimumNumber(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: number
            minimum: 200
      operationId: locateFishy
`

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)
	m, errs := doc.BuildV3Model()
	require.Len(t, errs, 0)

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=300", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Nil(t, errors)
}

func TestNewValidator_QueryParamMinimumNumber_violation(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: number
            minimum: 200
      operationId: locateFishy
`

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)
	m, errs := doc.BuildV3Model()
	require.Len(t, errs, 0)

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=123", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "Query parameter 'fishy' failed to validate", errors[0].Message)
}

func TestNewValidator_QueryParamValidTypeFloat(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: number
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=123.223", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Nil(t, errors)
}

func TestNewValidator_QueryParamWrongTypeBool(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: boolean
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=cod", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.NotNil(t, errors)
	assert.Equal(t, "Query parameter 'fishy' is not a valid boolean", errors[0].Message)
}

func TestNewValidator_QueryParamValidTypeBool(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: boolean
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=true", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Nil(t, errors)
}

func TestNewValidator_QueryParamInvalidEnumString(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: string
            enum: [cod, halibut]`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=cod&fishy=haddock", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Query parameter 'fishy' does not match allowed values", errors[0].Message)
	assert.Equal(t, "Instead of 'haddock', use one of the allowed values: 'cod, halibut'", errors[0].HowToFix)
}

func TestNewValidator_QueryParamInvalidEnumNumber(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: number
            enum: [1, 99]`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=22", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Query parameter 'fishy' does not match allowed values", errors[0].Message)
	assert.Equal(t, "Instead of '22', use one of the allowed values: '1, 99'", errors[0].HowToFix)
}

func TestNewValidator_QueryParamValidEnumNumber(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: number
            enum: [1, 99]`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=1", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidTypeArrayString(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: array
            items:
              type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=cod&fishy=haddock", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)
	assert.Nil(t, errors)
}

func TestNewValidator_QueryParamInvalidTypeArrayStringEnum(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: array
            items:
              type: string
              enum: [cod, halibut]
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=cod&fishy=haddock", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Instead of 'haddock', use one of the allowed values: 'cod, halibut'", errors[0].HowToFix)
}

func TestNewValidator_QueryParamInvalidTypeArrayNumberEnum(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: array
            items:
              type: number
              enum: [1, 99]
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=1&fishy=2", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Instead of '2', use one of the allowed values: '1, 99'", errors[0].HowToFix)
}

func TestNewValidator_QueryParamInvalidTypeArrayNumber(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         required: true
         schema:
           type: array
           items:
             type: number
     operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=cod&fishy=haddock", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.Len(t, errors, 2)
	assert.Equal(t, "Query array parameter 'fishy' is not a valid number", errors[0].Message)
	assert.Equal(t, "The query parameter (which is an array) 'fishy' is defined as being a number, "+
		"however the value 'cod' is not a valid number", errors[0].Reason)
	assert.Equal(t, "Query array parameter 'fishy' is not a valid number", errors[1].Message)
	assert.Equal(t, "The query parameter (which is an array) 'fishy' is defined as being a number, "+
		"however the value 'haddock' is not a valid number", errors[1].Reason)
}

func TestNewValidator_QueryParamValidExplodedType(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         required: true
         explode: true
         schema:
           type: array
           items:
             type: string
     operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=cod,haddock", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.Len(t, errors, 2)
}

func TestNewValidator_QueryParamInvalidExplodedArray(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         required: true
         explode: true
         schema:
           type: array
           items:
             type: number
     operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=1&fishy=2", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamInvalidExplodedArrayAndInvalidType(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         required: true
         explode: true
         schema:
           type: array
           items:
             type: number
     operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=haddock&fishy=cod", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.Len(t, errors, 2)
}

func TestNewValidator_QueryParamValidExploded(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         required: true
         explode: false
         schema:
           type: array
           items:
             type: string
     operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=cod,haddock,mackrel", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamInvalidTypeArrayBool(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         required: true
         schema:
           type: array
           items:
             type: boolean 
operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=cod&fishy=haddock", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.Len(t, errors, 2)
	assert.Equal(t, "Query array parameter 'fishy' is not a valid boolean", errors[0].Message)
	assert.Equal(t, "The query parameter (which is an array) 'fishy' is defined as being a boolean, "+
		"however the value 'cod' is not a valid true/false value", errors[0].Reason)
	assert.Equal(t, "Query array parameter 'fishy' is not a valid boolean", errors[1].Message)
	assert.Equal(t, "The query parameter (which is an array) 'fishy' is defined as being a boolean, "+
		"however the value 'haddock' is not a valid true/false value", errors[1].Reason)
}

func TestNewValidator_QueryParamInvalidTypeArrayFloat(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         required: true
         schema:
           type: array
           items:
             type: number
operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=12&fishy=12.12&fishy=1234567789.1233456657", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamInvalidTypeArrayFloatPipeDelimited(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         style: pipeDelimited
         required: true
         schema:
           type: array
           items:
             type: number
operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=12|12345.2344|22111233444.342452435", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamInvalidTypeArrayObjectPipeDelimited(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         style: pipeDelimited
         required: true
         schema:
           type: object
           properties:
             ocean:
               type: number
             silver:
               type: number
           required: [ocean, silver]
operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=ocean|12|silver|12.2345", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidTypeObject(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: array
            items:
              type: object
              properties:
                vinegar:
                  type: boolean
                chips:
                  type: number
              required:
                - vinegar
                - chips
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy={\"cod\":\"cakes\"}&fishy={\"crab\":\"legs\"}", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.Len(t, errors, 2)
	assert.Equal(t, "Query array parameter 'fishy' failed to validate", errors[0].Message)
	assert.Equal(t, "The query parameter (which is an array) 'fishy' is defined as an object, "+
		"however it failed to pass a schema validation", errors[0].Reason)
	assert.Equal(t, "missing properties: 'vinegar', 'chips'", errors[0].SchemaValidationErrors[0].Reason)
	assert.Equal(t, "/required", errors[0].SchemaValidationErrors[0].Location)
}

func TestNewValidator_QueryParamValidTypeObjectPropType_Invalid(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          content:
            application/json:
              schema:
                type: object
                properties:
                  vinegar:
                    type: boolean
                  chips:
                    type: number
                  required:
                    - vinegar
                    - chips
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy={\"vinegar\":\"cakes\",\"chips\":\"hello\"}&fishy={\"vinegar\":true,\"chips\":123.223}", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 2)
}

func TestNewValidator_QueryParamValidTypeObjectPropTypeFloat(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          content:
            application/json:
              schema:
                type: object
                properties:
                  vinegar:
                    type: boolean
                  chips:
                    type: number
              required:
                - vinegar
                - chips
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy={\"vinegar\":true,\"chips\":12}&fishy={\"vinegar\":true,\"chips\":123.333}", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamInvalidTypeObjectArrayPropType_Ref(t *testing.T) {
	spec := `openapi: 3.1.0
components:
  parameters:
    something:
      name: somethingElse
      in: query
      content:
        application/json:
          schema:
            type: array
            items:
              type: object
              properties:
                vinegar:
                  type: boolean
                chips:
                  type: number
              required:
                - vinegar
                - chips
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          content:
            $ref: "#/components/parameters/something/content"
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy={\"vinegar\":\"cakes\",\"chips\":\"hello\"}&fishy={\"vinegar\":true,\"chips\":123}", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 2)
}

func TestNewValidator_QueryParamValidTypeObjectArrayPropType_Ref(t *testing.T) {
	spec := `openapi: 3.1.0
components:
  parameters:
    something:
      name: somethingElse
      in: query
      content:
        application/json:
          schema:
            type: array
            items:
              type: object
              properties:
                vinegar:
                  type: boolean
                chips:
                  type: number
              required:
                - vinegar
                - chips
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          content:
            $ref: "#/components/parameters/something/content"
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy={\"vinegar\":false,\"chips\":999}&fishy={\"vinegar\":true,\"chips\":123}", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidTypeObjectPropType_Ref(t *testing.T) {
	spec := `openapi: 3.1.0
components:
  parameters:
    fishy:
      name: fishy
      in: query
      schema:
        type: object
        properties:
          vinegar:
            type: boolean
          chips:
            type: number
        required:
          - vinegar
          - chips
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - $ref: "#/components/parameters/fishy"
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?vinegar=true&chips=12", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidTypeObjectPropType_RefInvalid(t *testing.T) {
	spec := `openapi: 3.1.0
components:
  parameters:
    fishy:
      name: fishy
      in: query
      schema:
        type: object
        properties:
          vinegar:
            type: boolean
          chips:
            type: number
        required:
          - vinegar
          - chips
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - $ref: "#/components/parameters/fishy"
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?vinegar=true&chips=false", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "expected number, but got boolean", errors[0].SchemaValidationErrors[0].Reason)
}

func TestNewValidator_QueryParamValidTypeObjectPropType_RefViaContentWrapped(t *testing.T) {
	spec := `openapi: 3.1.0
components:
  parameters:
    fishy:
      name: fishy
      in: query
      content:
        application/json:
          schema:
            type: object
            properties:
              vinegar:
                type: boolean
              chips:
                type: number
            required:
              - vinegar
              - chips
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - $ref: "#/components/parameters/fishy"
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy={\"vinegar\":false,\"chips\":999}", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidTypeObjectPropType_RefViaContentWrappedInvalid(t *testing.T) {
	spec := `openapi: 3.1.0
components:
  parameters:
    fishy:
      name: fishy
      in: query
      content:
        application/json:
          schema:
            type: object
            properties:
              vinegar:
                type: boolean
              chips:
                type: number
            required:
              - vinegar
              - chips
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - $ref: "#/components/parameters/fishy"
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy={\"vinegar\":false,\"chips\":\"I am invalid\"}", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "expected number, but got string", errors[0].SchemaValidationErrors[0].Reason)
}

func TestNewValidator_QueryParamValidTypeObjectPropType_JSONInvalid(t *testing.T) {
	spec := `openapi: 3.1.0
components:
  parameters:
    fishy:
      name: fishy
      in: query
      content:
        application/json:
          schema:
            type: object
            properties:
              vinegar:
                type: boolean
              chips:
                type: number
            required:
              - vinegar
              - chips
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - $ref: "#/components/parameters/fishy"
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy=I am not json", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Query parameter 'fishy' is not valid JSON", errors[0].Message)
}

func TestNewValidator_QueryParamInvalidTypeObjectPropType_Ref(t *testing.T) {
	spec := `openapi: 3.1.0
components:
  schema_validation:
    chippy:
      type: object
      properties:
        vinegar:
          type: boolean
        chips:
          type: number
      required:
        - vinegar
        - chips
  parameters:
    fishy:
      name: fishy
      in: query
      content:
        application/json:
          schema:
            $ref: "#/components/schema_validation/chippy"
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - $ref: "#/components/parameters/fishy"
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/a/fishy/on/a/dishy?fishy={\"vinegar\":1234,\"chips\":false}", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 2)
}

func TestNewValidator_QueryParamValidateStyle_AllowReserved(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         required: true
         explode: true
         schema:
           type: array
           items:
             type: string
     operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=$$oh", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.Len(t, errors, 1)
	assert.Equal(t, "parameter values need to URL Encoded to ensure "+
		"reserved values are correctly encoded, for example: '%24%24oh'", errors[0].HowToFix)
}

func TestNewValidator_QueryParamValidateStyle_ValidObjectArrayNoExplode(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         required: true
         schema:
           type: array
           items:
             type: string
     operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=cod,haddock,mackrel", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidateStyle_InValidObjectArrayNoExplode(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         required: true
         explode: true
         allowReserved: true
         schema:
           type: array
           items:
             type: string
     operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=cod,haddock,mackrel", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.Len(t, errors, 1)
	assert.Equal(t, "Query parameter 'fishy' is not exploded correctly", errors[0].Message)
}

func TestNewValidator_QueryParamValidateStyle_SpaceDelimitedIncorrectlyExploded(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         required: true
         style: spaceDelimited
         schema:
           type: array
           items:
             type: string
     operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=cod&fishy=haddock&fishy=mackrel", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.Len(t, errors, 1)
	assert.Equal(t, "Query parameter 'fishy' delimited incorrectly", errors[0].Message)
}

func TestNewValidator_QueryParamValidateStyle_PipeDelimitedObjectValidExplode(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         required: true
         style: pipeDelimited
         schema:
           type: array
           items:
             type: string
     operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=cod|haddock|mackrel", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidateStyle_PipeDelimitedObjectInvalidExplode(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
 /a/fishy/on/a/dishy:
   get:
     parameters:
       - name: fishy
         in: query
         required: true
         style: pipeDelimited
         schema:
           type: array
           items:
             type: string
     operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=cod|haddock|mackrel&fishy=breaded|cooked|fried", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
}

func TestNewValidator_QueryParamValidateStyle_PipeDelimitedObjectValid(t *testing.T) {
	spec := `
openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: plate
          in: query
          required: true
          style: pipeDelimited
          schema:
            type: array
            items:
              type: string
        - name: fishy
          in: query
          required: true
          style: pipeDelimited
          schema:
            type: array
            items:
              type: string
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=cod|haddock|mackrel&plate=flat|round", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidateStyle_PipeDelimitedObjectDecode(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          style: pipeDelimited
          schema:
            type: object
            properties:
              fish:
                type: string
                enum:
                  - salmon
                  - tuna
                  - cod
              dish:
                type: string
                enum:
                 - salad
                 - soup
                 - stew
            required:
              - fish
              - dish
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=fish|salmon|dish|stew", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidateStyle_PipeDelimitedObjectDecodeInvalid(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          style: pipeDelimited
          schema:
            type: object
            properties:
              fish:
                type: string
                enum:
                  - salmon
                  - tuna
                  - cod
              dish:
                type: string
                enum:
                 - salad
                 - soup
                 - stew
            required:
              - fish
              - dish
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=fish|salmon|dish|cakes", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.Len(t, errors, 1)
	assert.Equal(t, "value must be one of \"salad\", \"soup\", \"stew\"", errors[0].SchemaValidationErrors[0].Reason)
}

func TestNewValidator_QueryParamValidateStyle_SpaceDelimitedObjectDecode(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          style: spaceDelimited
          schema:
            type: object
            properties:
              fish:
                type: string
                enum:
                  - salmon
                  - tuna
                  - cod
              dish:
                type: string
                enum:
                 - salad
                 - soup
                 - stew
            required:
              - fish
              - dish
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=fish%20salmon%20dish%20stew", nil) // dumb, don't do this.

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidateStyle_SpaceDelimitedObjectDecodeInvalid(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          style: spaceDelimited
          schema:
            type: object
            properties:
              fish:
                type: string
                enum:
                  - salmon
                  - tuna
                  - cod
              dish:
                type: string
                enum:
                 - salad
                 - soup
                 - stew
            required:
              - fish
              - dish
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=fish%20salmon%20dish%20coffee", nil) // dumb, don't do this.
	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "value must be one of \"salad\", \"soup\", \"stew\"", errors[0].SchemaValidationErrors[0].Reason)
}

func TestNewValidator_QueryParamValidateStyle_PipeDelimitedObjectInvalidMultiple(t *testing.T) {
	spec := `
openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: plate
          in: query
          required: true
          style: pipeDelimited
          schema:
            type: array
            items:
              type: string
        - name: fishy
          in: query
          required: true
          style: pipeDelimited
          schema:
            type: array
            items:
              type: string
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=cod|haddock|mackrel&plate=flat,round", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidateStyle_DeepObjectMultiValuesNoSchema(t *testing.T) {
	spec := `---
openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          style: deepObject
          schema:
            type: object
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy[ocean]=atlantic&fishy[salt]=12", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidateStyle_DeepObjectMultiValuesInvalid(t *testing.T) {
	spec := `---
openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          style: deepObject
          schema:
            type: object
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=atlantic&fishy=12", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.Len(t, errors, 2)
	assert.Equal(t, "The query parameter 'fishy' has the 'deepObject' style defined, "+
		"There are multiple values (2) supplied, instead of a single value", errors[0].Reason)
}

func TestNewValidator_QueryParamValidateStyle_FormEncoding(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: object
            properties:
              ocean:
                type: string
              fins:
                type: number
            required: [ocean, fins]
        - name: dishy
          in: query
          required: [hot, salty]
          schema:
            type: object
            properties:
              hot:
                type: boolean
              salty:
                type: boolean
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=ocean,atlantic,fins,4&dishy=hot,true,salty,true", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidateStyle_FormEncodingInvalid(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: object
            properties:
              ocean:
                type: string
              fins:
                type: number
            required: [ocean, fins]
        - name: dishy
          in: query
          schema:
            required: [hot, salty]
            type: object
            properties:
              hot:
                type: boolean
              salty:
                type: boolean
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=ocean,atlantic,fins,4&dishy=hot,no,salty,why", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.Len(t, errors, 1)
	assert.Equal(t, "expected boolean, but got string", errors[0].SchemaValidationErrors[0].Reason)
	assert.Equal(t, "expected boolean, but got string", errors[0].SchemaValidationErrors[1].Reason)
}

func TestNewValidator_QueryParamValidateStyle_FormEncodingArray(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          schema:
            type: array
            items:
              type: number
        - name: dishy
          in: query
          required: true
          schema:
            type: array
            items:
              type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=1,2,3&dishy=a,little,plate", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidateStyle_FormEncodingArrayExplode(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          explode: true
          required: true
          schema:
            type: array
            items:
              type: number
        - name: dishy
          in: query
          explode: true
          required: true
          schema:
            type: array
            items:
              type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=1&fishy=2&fishy=3&dishy=a&dishy=little&dishy=dish", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidateStyle_FormEncodingArrayExplodeInvalid(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          explode: true
          required: true
          schema:
            type: array
            items:
              type: number
        - name: dishy
          in: query
          explode: true
          required: true
          schema:
            type: array
            items:
              type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=1,2,3&dishy=little,dishy", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 4)
	assert.Equal(t, "The query parameter 'fishy' has a default or 'form' encoding defined, however the "+
		"value '1,2,3' is encoded as an object or an array using commas. "+
		"The contract defines the explode value to set to 'true'", errors[0].Reason)
}

func TestNewValidator_QueryParamValidateStyle_PipeDelimitedValid(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          style: pipeDelimited  
          schema:
            type: array
            items:
              type: number
        - name: dishy
          in: query
          style: pipeDelimited
          required: true
          schema:
            type: array
            items:
              type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=1|2|3&dishy=little|dishy", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidateStyle_SpaceDelimitedValid(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          style: spaceDelimited  
          schema:
            type: array
            items:
              type: number
        - name: dishy
          in: query
          style: spaceDelimited
          required: true
          schema:
            type: array
            items:
              type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=1%202%203&dishy=little%20dishy", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_QueryParamValidateStyle_SpaceDelimitedInvalidSchema(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          style: spaceDelimited  
          schema:
            type: array
            items:
              type: number
        - name: dishy
          in: query
          style: spaceDelimited
          required: true
          schema:
            type: array
            items:
              type: string
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy=1|%202%203&dishy=little%20dishy", nil)
	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Convert the value '1|' into a number", errors[0].HowToFix)
}

func TestNewValidator_QueryParamValidateStyle_DeepObjectMultiValuesFailedSchema(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          style: deepObject
          schema:
            type: object
            properties:
              ocean:
                type: string
              salt:
                type: boolean
            required: [ocean, salt]
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy[ocean]=atlantic&fishy[salt]=12", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.Len(t, errors, 1)
	assert.Equal(t, "expected boolean, but got number", errors[0].SchemaValidationErrors[0].Reason)
}

func TestNewValidator_QueryParamSetPath(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          style: deepObject
          schema:
            type: object
            properties:
              ocean:
                type: string
              salt:
                type: boolean
            required: [ocean, salt]
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy[ocean]=atlantic&fishy[salt]=12", nil)

	// preset the path
	path, _, pv := paths.FindPath(request, &m.Model)
	v.SetPathItem(path, pv)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.Len(t, errors, 1)
	assert.Equal(t, "expected boolean, but got number", errors[0].SchemaValidationErrors[0].Reason)
}

func TestNewValidator_QueryParamValidateStyle_DeepObjectMultiValuesFailedMultipleSchemas(t *testing.T) {
	spec := `---
openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          style: deepObject
          schema:
            type: object
            properties:
              ocean:
                type: string
              salt:
                type: boolean
            required:
              - ocean
              - salt
        - name: dishy
          in: query
          required: true
          style: deepObject
          schema:
            type: object
            properties:
              size:
                type: string
              numCracks:
                type: number
            required:
              - size
              - numCracks
        - name: cake
          in: query
          required: true
          style: deepObject
          schema:
            type: object
            properties:
              message:
                type: string
              numCandles:
                type: number
            required:
              - message
              - numCandles
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/a/fishy/on/a/dishy?fishy[ocean]=atlantic&fishy[salt]=12"+
			"&dishy[size]=big&dishy[numCracks]=false"+
			"&cake[message]=happy%20birthday&cake[numCandles]=false", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)

	assert.Len(t, errors, 3)
	assert.Equal(t, "expected boolean, but got number", errors[0].SchemaValidationErrors[0].Reason)
	assert.Equal(t, "expected number, but got boolean", errors[1].SchemaValidationErrors[0].Reason)
	assert.Equal(t, "expected number, but got boolean", errors[2].SchemaValidationErrors[0].Reason)
}

func TestNewValidator_ValidateEncodedObjectIsCorrect(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    get:
      parameters:
        - name: fishy
          in: query
          required: true
          style: deepObject
          schema:
            type: object
            properties:
              ocean:
                type: string
              salt:
                type: boolean
            required: [ocean, salt]
      operationId: locateFishy`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	sch := m.Model.Paths.PathItems.GetOrZero("/a/fishy/on/a/dishy").Get.Parameters[0].Schema

	s := sch.Schema()

	// this is not compatible.
	rawObject := map[int]int{
		1: 2,
	}

	errs := ValidateParameterSchema(s, rawObject, "cake", "burger", "lemons",
		"pizza", "rice", "herbs")

	assert.Len(t, errs, 1)
	assert.Equal(t, "lemons 'pizza' is defined as an object, "+
		"however it failed to be decoded as an object", errs[0].Reason)
}

// https://github.com/pb33f/wiretap/issues/82
func TestNewValidator_QueryParamValidateStyle_DeepObjectAdditionalPropertiesArray(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /anything/queryParams/deepObject/map:
    get:
      operationId: deepObjectQueryParamsMap
      parameters:
        - name: mapArrParam
          in: query
          style: deepObject
          schema:
            type: object
            additionalProperties:
              type: array
              items:
                type: string
            example: { "test": ["test", "test2"], "test2": ["test3", "test4"] }
      responses:
        "200":
          description: OK`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/anything/queryParams/deepObject/map?mapArrParam[test2]=test3&mapArrParam[test2]=test4&mapArrParam[test]=test&mapArrParam[test]=test2", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Len(t, errors, 0)
}

// https://github.com/pb33f/wiretap/issues/82
func TestNewValidator_QueryParamValidateStyle_DeepObjectAdditionalPropertiesArrayTop(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /anything/queryParams/deepObject/map:
    get:
      operationId: deepObjectQueryParamsMap
      parameters:
        - name: mapArrParam
          in: query
          style: deepObject
          schema:
            type: array
            items:
              type: string
            example: { "test": ["test", "test2"], "test2": ["test3", "test4"] }
      responses:
        "200":
          description: OK`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/anything/queryParams/deepObject/map?mapArrParam[test2]=test3&mapArrParam[test2]=test4&mapArrParam[test]=test&mapArrParam[test]=test2", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)

	assert.Len(t, errors, 0)
}

// https://github.com/pb33f/wiretap/issues/82
func TestNewValidator_QueryParamValidateStyle_DeepObjectAdditionalPropertiesArray_Fail(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /anything/queryParams/deepObject/map:
    get:
      operationId: deepObjectQueryParamsMap
      parameters:
        - name: mapArrParam
          in: query
          style: deepObject
          schema:
            type: object
            additionalProperties:
              type: array
              items:
                type: string
            example: { "test": ["test", "test2"], "test2": ["test3", "test4"] }
      responses:
        "200":
          description: OK`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"https://things.com/anything/queryParams/deepObject/map?mapArrParam[test2]=23&mapArrParam[test2]=test4&mapArrParam[test]=test&mapArrParam[test]=test2", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "expected string, but got number", errors[0].SchemaValidationErrors[0].Reason)
}

// https://github.com/pb33f/wiretap/issues/83
func TestNewValidator_QueryParamValidateStyle_BadSchemaDeepObject(t *testing.T) {
	spec := `openapi: 3.1.0
info:
  title: Test
  version: 0.1.0
security:
  - apiKeyAuth: []
paths:
  /anything/queryParams/deepObject/obj:
    get:
      operationId: deepObjectQueryParamsObject
      parameters:
        - name: objParam
          in: query
          style: deepObject
          schema:
            $ref: "components.yaml#/components/schemas/simpleObject"
          required: true
      responses:
        "200":
          description: OK
components:
  securitySchemes:
    apiKeyAuth:
      type: apiKey
      in: header
      name: Authorization
      description: Authenticate using an API Key generated via our platform.`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, err := doc.BuildV3Model()
	assert.Len(t, err, 1) // path build will fail because of missing schema.

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"http://localhost:9090/anything/queryParams/deepObject/obj?objParam=blahdedahdedah", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

// https://github.com/pb33f/wiretap/issues/83
func TestNewValidator_QueryParamValidateStyle_BadSchemaDeepObject_Inline(t *testing.T) {
	spec := `openapi: 3.1.0
info:
  title: Test
  version: 0.1.0
security:
  - apiKeyAuth: []
paths:
  /anything/queryParams/deepObject/obj:
    get:
      operationId: deepObjectQueryParamsObject
      parameters:
        - name: objParam
          in: query
          style: deepObject
          schema:
            type: object
            properties:
              cake:
                type: string
          required: true
      responses:
        "200":
          description: OK
components:
  securitySchemes:
    apiKeyAuth:
      type: apiKey
      in: header
      name: Authorization
      description: Authenticate using an API Key generated via our platform.`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, err := doc.BuildV3Model()
	assert.Len(t, err, 0) // no patch build here

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet,
		"http://localhost:9090/anything/queryParams/deepObject/obj?objParam=blahdedahdedah", nil)

	valid, errors := v.ValidateQueryParams(request)
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "The query parameter 'objParam' is defined as an object,"+
		" however it failed to pass a schema validation", errors[0].Reason)
}
