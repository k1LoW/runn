// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package responses

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/helpers"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/paths"
	"github.com/stretchr/testify/assert"
)

func TestValidateBody_MissingContentType(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
                  patties:
                    type: integer
                  vegetarian:
                    type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    false,
		"vegetarian": 2,
	}

	bodyBytes, _ := json.Marshal(body)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", bytes.NewReader(bodyBytes))
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, "cheeky/monkey")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "POST / 200 operation response content type 'cheeky/monkey' does not exist", errors[0].Message)
	assert.Equal(t, "The content type is invalid, Use one of the 1 "+
		"supported types for this operation: application/json", errors[0].HowToFix)
	assert.Equal(t, request.Method, errors[0].RequestMethod)
	assert.Equal(t, request.URL.Path, errors[0].RequestPath)
	assert.Equal(t, "/burgers/createBurger", errors[0].SpecPath)
}

func TestValidateBody_MissingPath(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
                  patties:
                    type: integer
                  vegetarian:
                    type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    false,
		"vegetarian": 2,
	}

	bodyBytes, _ := json.Marshal(body)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/I do not exist", bytes.NewReader(bodyBytes))
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, "cheeky/monkey") // won't even matter!
		w.WriteHeader(http.StatusUnprocessableEntity)              // does not matter.
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "POST Path '/I do not exist' not found", errors[0].Message)
	assert.Equal(t, request.Method, errors[0].RequestMethod)
	assert.Equal(t, request.URL.Path, errors[0].RequestPath)
	assert.Equal(t, "", errors[0].SpecPath)
}

func TestValidateBody_SetPath(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
                  patties:
                    type: integer
                  vegetarian:
                    type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    false,
		"vegetarian": 2,
	}

	bodyBytes, _ := json.Marshal(body)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/I do not exist", bytes.NewReader(bodyBytes))
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// preset the path
	path, _, pv := paths.FindPath(request, &m.Model)
	v.SetPathItem(path, pv)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, "cheeky/monkey") // won't even matter!
		w.WriteHeader(http.StatusUnprocessableEntity)              // does not matter.
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "POST Path '/I do not exist' not found", errors[0].Message)
}

func TestValidateBody_MissingStatusCode(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
                  patties:
                    type: integer
                  vegetarian:
                    type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    false,
		"vegetarian": 2,
	}

	bodyBytes, _ := json.Marshal(body)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", bytes.NewReader(bodyBytes))
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, "cheeky/monkey") // won't even matter!
		w.WriteHeader(http.StatusUnprocessableEntity)              // undefined in the spec.
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "POST operation request response code '422' does not exist", errors[0].Message)
	assert.Equal(t, "The service is responding with a code that is not defined in the spec, fix the service or add the code to the specification", errors[0].HowToFix)
}

func TestValidateBody_InvalidBasicSchema(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
                  patties:
                    type: integer
                  vegetarian:
                    type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	// mix up the primitives to fire two schema violations.
	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    false,
		"vegetarian": 2,
	}

	bodyBytes, _ := json.Marshal(body)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", bytes.NewReader(bodyBytes))
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)
	// doubletap to hit cache
	_, _ = v.ValidateResponseBody(request, response)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 2)
}

func TestValidateBody_NoBody(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
                  patties:
                    type: integer
                  vegetarian:
                    type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", http.NoBody)
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(nil)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)
	// doubletap to hit cache
	_, _ = v.ValidateResponseBody(request, response)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
	//assert.Len(t, errors[0].SchemaValidationErrors, 2)
}

func TestValidateBody_InvalidResponseBodyNil(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
                  patties:
                    type: integer
                  vegetarian:
                    type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", http.NoBody)
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// invalid response
	response := &http.Response{
		Header:     http.Header{},
		StatusCode: http.StatusOK,
		Body:       nil, // invalid response body
	}
	response.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)
	// doubletap to hit cache
	_, _ = v.ValidateResponseBody(request, response)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
}

func TestValidateBody_InvalidResponseBodyError(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
                  patties:
                    type: integer
                  vegetarian:
                    type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", http.NoBody)
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// invalid response
	response := &http.Response{
		Header:     http.Header{},
		StatusCode: http.StatusOK,
		Body:       &errorReader{},
	}
	response.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)
	// doubletap to hit cache
	_, _ = v.ValidateResponseBody(request, response)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
}

func TestValidateBody_InvalidBasicSchema_SetPath(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
                  patties:
                    type: integer
                  vegetarian:
                    type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	// mix up the primitives to fire two schema violations.
	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    false,
		"vegetarian": 2,
	}

	bodyBytes, _ := json.Marshal(body)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", bytes.NewReader(bodyBytes))
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// preset the path
	path, _, pv := paths.FindPath(request, &m.Model)
	v.SetPathItem(path, pv)

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 2)
	assert.Equal(t, "200 response body for '/burgers/createBurger' failed to validate schema", errors[0].Message)
}

func TestValidateBody_ValidComplexSchema(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schema_validation/TestBody' 
components:
  schema_validation:
    Uncooked:
      type: object
      required: [uncookedWeight, uncookedHeight]
      properties:
        uncookedWeight:
          type: number
        uncookedHeight:
          type: number
    Cooked:
      type: object
      required: [usedOil, usedAnimalFat]
      properties:
        usedOil:
          type: boolean
        usedAnimalFat:
          type: boolean
    Nutrients:
      type: object
      required: [fat, salt, meat]
      properties:
        fat:
          type: number
        salt:
          type: number
        meat:
          type: string
          enum:
            - beef
            - pork
            - lamb
            - vegetables      
    TestBody:
      type: object
      oneOf:
        - $ref: '#/components/schema_validation/Uncooked'
        - $ref: '#/components/schema_validation/Cooked'
      allOf:
        - $ref: '#/components/schema_validation/Nutrients'
      properties:
        name:
          type: string
        patties:
          type: integer
        vegetarian:
          type: boolean
      required: [name, patties, vegetarian]`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	body := map[string]interface{}{
		"name":          "Big Mac",
		"patties":       2,
		"vegetarian":    true,
		"fat":           10.0,
		"salt":          0.5,
		"meat":          "beef",
		"usedOil":       true,
		"usedAnimalFat": false,
	}

	bodyBytes, _ := json.Marshal(body)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", bytes.NewReader(bodyBytes))
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestValidateBody_InvalidComplexSchema(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schema_validation/TestBody' 
components:
  schema_validation:
    Uncooked:
      type: object
      required: [uncookedWeight, uncookedHeight]
      properties:
        uncookedWeight:
          type: number
        uncookedHeight:
          type: number
    Cooked:
      type: object
      required: [usedOil, usedAnimalFat]
      properties:
        usedOil:
          type: boolean
        usedAnimalFat:
          type: boolean
    Nutrients:
      type: object
      required: [fat, salt, meat]
      properties:
        fat:
          type: number
        salt:
          type: number
        meat:
          type: string
          enum:
            - beef
            - pork
            - lamb
            - vegetables      
    TestBody:
      type: object
      oneOf:
        - $ref: '#/components/schema_validation/Uncooked'
        - $ref: '#/components/schema_validation/Cooked'
      allOf:
        - $ref: '#/components/schema_validation/Nutrients'
      properties:
        name:
          type: string
        patties:
          type: integer
        vegetarian:
          type: boolean
      required: [name, patties, vegetarian]`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	body := map[string]interface{}{
		"name":          "Big Mac",
		"patties":       2,
		"vegetarian":    true,
		"fat":           10.0,
		"salt":          0.5,
		"meat":          "beef",
		"usedOil":       12345, // invalid, should be bool
		"usedAnimalFat": false,
	}

	bodyBytes, _ := json.Marshal(body)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", bytes.NewReader(bodyBytes))
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 3)
	assert.Equal(t, "expected boolean, but got number", errors[0].SchemaValidationErrors[2].Reason)
}

func TestValidateBody_ValidBasicSchema(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
                  patties:
                    type: integer
                  vegetarian:
                    type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    2,
		"vegetarian": false,
	}

	bodyBytes, _ := json.Marshal(body)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", bytes.NewReader(bodyBytes))
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestValidateBody_ValidBasicSchema_WithFullContentTypeHeader(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
                  patties:
                    type: integer
                  vegetarian:
                    type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    2,
		"vegetarian": false,
	}

	bodyBytes, _ := json.Marshal(body)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", bytes.NewReader(bodyBytes))
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {

		// inject a full content type header, including charset and boundary
		w.Header().Set(helpers.ContentTypeHeader,
			fmt.Sprintf("%s; charset=utf-8; boundary=---12223344", helpers.JSONContentType))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestValidateBody_ValidBasicSchemaUsingDefault(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        default:
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
                  patties:
                    type: integer
                  vegetarian:
                    type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    2,
		"vegetarian": false,
	}

	bodyBytes, _ := json.Marshal(body)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", bytes.NewReader(bodyBytes))
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestValidateBody_InvalidBasicSchemaUsingDefault_MissingContentType(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        default:
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
                  patties:
                    type: integer
                  vegetarian:
                    type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	// primitives are now correct.
	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    2,
		"vegetarian": false,
	}

	bodyBytes, _ := json.Marshal(body)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", bytes.NewReader(bodyBytes))
	request.Header.Set(helpers.ContentTypeHeader, "chicken/nuggets;chicken=soup")

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, r.Header.Get(helpers.ContentTypeHeader))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "POST / 200 operation response content type 'chicken/nuggets' does not exist", errors[0].Message)
}

func TestValidateBody_InvalidSchemaMultiple(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  required:
                    - name
                  properties:
                    name:
                      type: string
                    patties:
                      type: integer
                    vegetarian:
                      type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	var items []map[string]interface{}
	items = append(items, map[string]interface{}{
		"patties":    1,
		"vegetarian": true,
	})
	items = append(items, map[string]interface{}{
		"name":       "Quarter Pounder",
		"patties":    true,
		"vegetarian": false,
	})
	items = append(items, map[string]interface{}{
		"name":       "Big Mac",
		"patties":    2,
		"vegetarian": false,
	})

	bodyBytes, _ := json.Marshal(items)

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", bytes.NewReader(bodyBytes))
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 2)
	assert.Equal(t, "200 response body for '/burgers/createBurger' failed to validate schema", errors[0].Message)
}

func TestValidateBody_EmptyContentType_Valid(t *testing.T) {
	spec := `openapi: "3.0.0"
info:
  title: Healthcheck
  version: '0.1.0'
paths:
  /health:
    get:
      responses:
        '200':
          description: pet response
          content: {}`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	// build a request
	request, _ := http.NewRequest(http.MethodGet, "https://things.com/health", nil)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(nil)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.True(t, valid)
	assert.Len(t, errors, 0)

}

func TestValidateBody_InvalidBodyJSON(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      responses:
        default:
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
                  patties:
                    type: integer
                  vegetarian:
                    type: boolean`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	badJson := []byte("{\"bad\": \"json\",}")

	// build a request
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", bytes.NewReader(badJson))
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, r.Header.Get(helpers.ContentTypeHeader))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(badJson)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "POST response body for '/burgers/createBurger' failed to validate schema", errors[0].Message)
	assert.Equal(t, "invalid character '}' looking for beginning of object key string", errors[0].SchemaValidationErrors[0].Reason)

}

func TestValidateBody_NoContentType_Valid(t *testing.T) {
	spec := `openapi: "3.0.0"
info:
  title: Healthcheck
  version: '0.1.0'
paths:
  /health:
    get:
      responses:
        '200':
          description: pet response`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewResponseBodyValidator(&m.Model)

	// build a request
	request, _ := http.NewRequest(http.MethodGet, "https://things.com/health", nil)

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(nil)
	}

	// fire the request
	handler(res, request)

	// record response
	response := res.Result()

	// validate!
	valid, errors := v.ValidateResponseBody(request, response)

	assert.True(t, valid)
	assert.Len(t, errors, 0)

}

type errorReader struct{}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("some io error")
}
func (er *errorReader) Close() error {
	return nil
}
