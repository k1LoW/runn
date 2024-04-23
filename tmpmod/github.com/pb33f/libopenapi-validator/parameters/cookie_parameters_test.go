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

func TestNewValidator_CookieNoPath(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: number`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/I/do/not/exist", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "1"})

	valid, errors := v.ValidateCookieParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, request.Method, errors[0].RequestMethod)
	assert.Equal(t, request.URL.Path, errors[0].RequestPath)
	assert.Equal(t, "", errors[0].SpecPath)
}

func TestNewValidator_CookieParamNumberValid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: number`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "1"})

	valid, errors := v.ValidateCookieParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_CookieParamNumberValidFloat(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: number`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "123.455"})

	valid, errors := v.ValidateCookieParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_CookieParamNumberInvalid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: number`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "false"})

	valid, errors := v.ValidateCookieParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Convert the value 'false' into a number", errors[0].HowToFix)
}

func TestNewValidator_CookieParamBooleanValid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "true"})

	valid, errors := v.ValidateCookieParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_CookieParamEnumValidString(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: string
            enum:
              - beef
              - chicken
              - pea protein`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "chicken"})

	valid, errors := v.ValidateCookieParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_CookieParamEnumInvalidString(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: string
            enum:
              - beef
              - chicken
              - pea protein`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "milk"})

	valid, errors := v.ValidateCookieParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t,
		"Instead of 'milk', use one of the allowed values: 'beef, chicken, pea protein'", errors[0].HowToFix)
	assert.Equal(t, request.Method, errors[0].RequestMethod)
	assert.Equal(t, request.URL.Path, errors[0].RequestPath)
	assert.Equal(t, "/burgers/beef", errors[0].SpecPath)
}

func TestNewValidator_CookieParamBooleanInvalid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "12345"})

	valid, errors := v.ValidateCookieParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Convert the value '12345' into a true/false value", errors[0].HowToFix)
}

func TestNewValidator_CookieParamObjectValid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          explode: false
          schema:
            type: object
            properties:
              pink:
                type: boolean
              number:
                type: number
            required: [pink, number]`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "pink,true,number,2"})

	valid, errors := v.ValidateCookieParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_CookieParamObjectInvalid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          explode: false
          schema:
            type: object
            properties:
              pink:
                type: boolean
              number:
                type: number
            required: [pink, number]`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "pink,2,number,2"})

	valid, errors := v.ValidateCookieParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "expected boolean, but got number", errors[0].SchemaValidationErrors[0].Reason)
}

func TestNewValidator_CookieParamArrayValidNumber(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: array
            items:
              type: number`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "2,3,4"})

	valid, errors := v.ValidateCookieParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_CookieParamArrayInvalidNumber(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: array
            items:
              type: number`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "2,true,4,'hello'"})

	valid, errors := v.ValidateCookieParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 2)
}

func TestNewValidator_CookieParamArrayValidBoolean(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: array
            items:
              type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "true,false,true,false,true"})

	valid, errors := v.ValidateCookieParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_CookieParamArrayString(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: array
            items:
              type: string`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "true,1,hey,ho"})

	valid, errors := v.ValidateCookieParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_CookieParamArrayInvalidBoolean(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: array
            items:
              type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "true,false,pb33f,false,99.99"})

	valid, errors := v.ValidateCookieParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 2)
}

func TestNewValidator_CookieParamArrayInvalidBooleanZeroOne(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: array
            items:
              type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "true,false,0,false,1"})

	valid, errors := v.ValidateCookieParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 2)
}

func TestNewValidator_CookieParamArrayValidIntegerEnum(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: integer
            enum: [1, 2, 99]`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "2"})

	valid, errors := v.ValidateCookieParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_CookieParamArrayInvalidIntegerEnum(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: integer
            enum: [1, 2, 99]`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "2500"}) // too many dude.

	valid, errors := v.ValidateCookieParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Instead of '2500', use one of the allowed values: '1, 2, 99'", errors[0].HowToFix)
}

func TestNewValidator_PresetPath(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/beef:
    get:
      parameters:
        - name: PattyPreference
          in: cookie
          required: true
          schema:
            type: integer
            enum: [1, 2, 99]`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/beef", nil)
	request.AddCookie(&http.Cookie{Name: "PattyPreference", Value: "2500"}) // too many dude.

	// preset the path
	path, _, pv := paths.FindPath(request, &m.Model)
	v.SetPathItem(path, pv)

	valid, errors := v.ValidateCookieParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Instead of '2500', use one of the allowed values: '1, 2, 99'", errors[0].HowToFix)
}
