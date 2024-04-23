// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package validator

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      requestBody:
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

	v, _ := NewValidator(doc)
	assert.NotNil(t, v.GetParameterValidator())
	assert.NotNil(t, v.GetResponseBodyValidator())
	assert.NotNil(t, v.GetRequestBodyValidator())
}

func TestNewValidator_ValidateDocument(t *testing.T) {

	doc, _ := libopenapi.NewDocument(petstoreBytes)
	v, _ := NewValidator(doc)
	valid, errs := v.ValidateDocument()
	assert.True(t, valid)
	assert.Len(t, errs, 0)
}

func TestNewValidator_BadDoc(t *testing.T) {

	spec := `swagger: 2.0`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	_, errs := NewValidator(doc)

	assert.Len(t, errs, 1)

}

func TestNewValidator_ValidateHttpRequest_BadPath(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      requestBody:
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

	v, _ := NewValidator(doc)

	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    2,
		"vegetarian": true,
	}

	bodyBytes, _ := json.Marshal(body)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/I am a potato man",
		bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	valid, errors := v.ValidateHttpRequest(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "POST Path '/I am a potato man' not found", errors[0].Message)

}

func TestNewValidator_ValidateHttpRequest_ValidPostSimpleSchema(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      requestBody:
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

	v, _ := NewValidator(doc)

	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    2,
		"vegetarian": true,
	}

	bodyBytes, _ := json.Marshal(body)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger",
		bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	valid, errors := v.ValidateHttpRequest(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)

}

func TestNewValidator_slash_server_url(t *testing.T) {

	spec := `openapi: 3.1.0
servers:
  - url: /
paths:
  /burgers/{burgerId}/locate:
    patch:
      operationId: locateBurger
      parameters:
        - name: burgerId
          in: path
          required: true
          schema:
            type: string
            format: uuid		
`

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)

	request, _ := http.NewRequest(http.MethodPatch, "https://things.com/burgers/edd0189c-420b-489c-98f2-0facc5a26f3a/locate", nil)
	v, _ := NewValidator(doc)

	valid, errors := v.ValidateHttpRequest(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_ValidateHttpRequest_SetPath_ValidPostSimpleSchema(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      requestBody:
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

	v, _ := NewValidator(doc)

	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    2,
		"vegetarian": true,
	}

	bodyBytes, _ := json.Marshal(body)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger",
		bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	valid, errors := v.ValidateHttpRequest(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)

}

func TestNewValidator_ValidateHttpRequest_InvalidPostSchema(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      requestBody:
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

	v, _ := NewValidator(doc)

	// mix up the primitives to fire two schema violations.
	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    false, // wrong.
		"vegetarian": false,
	}

	bodyBytes, _ := json.Marshal(body)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger",
		bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	valid, errors := v.ValidateHttpRequest(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "expected integer, but got boolean", errors[0].SchemaValidationErrors[0].Reason)

}

func TestNewValidator_ValidateHttpRequest_InvalidQuery(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    parameters:
       - in: query
         name: cheese
         required: true
         schema:
           type: string
    post:
      requestBody:
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

	v, _ := NewValidator(doc)

	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    2, // wrong.
		"vegetarian": false,
	}

	bodyBytes, _ := json.Marshal(body)

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger",
		bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	valid, errors := v.ValidateHttpRequest(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Query parameter 'cheese' is missing", errors[0].Message)

}

var petstoreBytes []byte

func init() {
	petstoreBytes, _ = os.ReadFile("test_specs/petstorev3.json")
}

func TestNewValidator_PetStore_MissingContentType(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a pet
	body := map[string]interface{}{
		"id":   123,
		"name": "cotton",
		"category": map[string]interface{}{
			"id":   123,
			"name": "dogs",
		},
		"photoUrls": []string{"https://example.com"},
	}

	// marshal the body into bytes.
	bodyBytes, _ := json.Marshal(body)

	// create a new put request
	request, _ := http.NewRequest(http.MethodPut, "https://hyperspace-superherbs.com/pet",
		bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, "application/not-json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response (should be clean)
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	// should all be perfectly valid.
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "PUT / 200 operation response content type 'application/not-json' does not exist",
		errors[0].Message)

	assert.Equal(t, "The content type 'application/not-json' of the PUT response received "+
		"has not been defined, it's an unknown type",
		errors[0].Reason)

}

func TestNewValidator_PetStore_PetPost200_Valid_PathPreset(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a pet
	body := map[string]interface{}{
		"id":   123,
		"name": "cotton",
		"category": map[string]interface{}{
			"id":   123,
			"name": "dogs",
		},
		"photoUrls": []string{"https://example.com"},
	}

	// marshal the body into bytes.
	bodyBytes, _ := json.Marshal(body)

	// create a new put request
	request, _ := http.NewRequest(http.MethodPut, "https://hyperspace-superherbs.com/pet",
		bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response (should be clean)
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	// should all be perfectly valid.
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_PetStore_PetPost200_Valid(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a pet
	body := map[string]interface{}{
		"id":   123,
		"name": "cotton",
		"category": map[string]interface{}{
			"id":   123,
			"name": "dogs",
		},
		"photoUrls": []string{"https://example.com"},
	}

	// marshal the body into bytes.
	bodyBytes, _ := json.Marshal(body)

	// create a new put request
	request, _ := http.NewRequest(http.MethodPut, "https://hyperspace-superherbs.com/pet",
		bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response (should be clean)
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	// should all be perfectly valid.
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_PetStore_PetPost200_Invalid(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a pet, but is missing the photoUrls field
	body := map[string]interface{}{
		"id":   123,
		"name": "cotton",
		"category": map[string]interface{}{
			"id":   123,
			"name": "dogs",
		},
	}

	// marshal the body into bytes.
	bodyBytes, _ := json.Marshal(body)

	// create a new put request
	request, _ := http.NewRequest(http.MethodPost, "https://hyperspace-superherbs.com/pet",
		bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusProxyAuthRequired) // this is not defined by the contract, so it should fail.
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	// we have a schema violation, and a response code violation, our validator should have picked them
	// both up.
	assert.False(t, valid)
	assert.Len(t, errors, 2)

	// check errors
	for i := range errors {
		if errors[i].SchemaValidationErrors != nil {
			assert.Equal(t, "missing properties: 'photoUrls'", errors[i].SchemaValidationErrors[0].Reason)
		} else {
			assert.Equal(t, "POST operation request response code '407' does not exist", errors[i].Message)
		}
	}
}

func TestNewValidator_PetStore_PetFindByStatusGet200_Valid(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a pet
	body := map[string]interface{}{
		"id":   123,
		"name": "cotton",
		"category": map[string]interface{}{
			"id":   123,
			"name": "dogs",
		},
		"photoUrls": []string{"https://example.com"},
	}

	// marshal the body into bytes.
	bodyBytes, _ := json.Marshal([]interface{}{body}) // operation returns an array of pets

	// create a new put request
	request, _ := http.NewRequest(http.MethodGet,
		"https://hyperspace-superherbs.com/pet/findByStatus?status=sold", nil)
	request.Header.Set("Content-Type", "application/json")

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response (should be clean)
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	// should all be perfectly valid.
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_PetStore_PetFindByStatusGet200_BadEnum(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a pet
	body := map[string]interface{}{
		"id":   123,
		"name": "cotton",
		"category": map[string]interface{}{
			"id":   123,
			"name": "dogs",
		},
		"photoUrls": []string{"https://example.com"},
	}

	// marshal the body into bytes.
	bodyBytes, _ := json.Marshal([]interface{}{body}) // operation returns an array of pets

	// create a new put request
	request, _ := http.NewRequest(http.MethodGet,
		"https://hyperspace-superherbs.com/pet/findByStatus?status=invalidEnum", nil) // enum is invalid
	request.Header.Set("Content-Type", "application/json")

	// simulate a request/response, in this case the contract returns a 200 with a pet
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.Header().Set("Herbs-And-Spice", helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response (should be clean)
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	// should all be perfectly valid.
	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Query parameter 'status' does not match allowed values", errors[0].Message)
	assert.Equal(t, "Instead of 'invalidEnum', use one of the allowed values: 'available, pending, sold'", errors[0].HowToFix)

}

func TestNewValidator_PetStore_PetFindByTagsGet200_Valid(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a pet
	body := map[string]interface{}{
		"id":   123,
		"name": "cotton",
		"category": map[string]interface{}{
			"id":   123,
			"name": "dogs",
		},
		"photoUrls": []string{"https://example.com"},
	}

	// marshal the body into bytes.
	bodyBytes, _ := json.Marshal([]interface{}{body}) // operation returns an array of pets

	// create a new put request
	request, _ := http.NewRequest(http.MethodGet,
		"https://hyperspace-superherbs.com/pet/findByTags?tags=fuzzy&tags=wuzzy", nil)
	request.Header.Set("Content-Type", "application/json")

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response (should be clean)
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	// should all be perfectly valid.
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_PetStore_PetFindByTagsGet200_InvalidExplode(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a pet
	body := map[string]interface{}{
		"id":   123,
		"name": "cotton",
		"category": map[string]interface{}{
			"id":   123,
			"name": "dogs",
		},
		"photoUrls": []string{"https://example.com"},
	}

	// marshal the body into bytes.
	bodyBytes, _ := json.Marshal([]interface{}{body}) // operation returns an array of pets

	// create a new put request
	request, _ := http.NewRequest(http.MethodGet,
		"https://hyperspace-superherbs.com/pet/findByTags?tags=fuzzy,wuzzy", nil)
	request.Header.Set("Content-Type", "application/json")

	// simulate a request/response
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	// will fail.
	assert.False(t, valid)
	assert.Len(t, errors, 2) // will fire allow reserved error, and explode error.
}

func TestNewValidator_PetStore_PetGet200_Valid(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a pet
	body := map[string]interface{}{
		"id":   123,
		"name": "cotton",
		"category": map[string]interface{}{
			"id":   123,
			"name": "dogs",
		},
		"photoUrls": []string{"https://example.com"},
	}

	// marshal the body into bytes.
	bodyBytes, _ := json.Marshal(body) // operation returns pet

	// create a new put request
	request, _ := http.NewRequest(http.MethodGet,
		"https://hyperspace-superherbs.com/pet/12345", nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("api_key", "12345")

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_PetStore_PetGet200_PathNotFound(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a pet
	body := map[string]interface{}{
		"id":   123,
		"name": "cotton",
		"category": map[string]interface{}{
			"id":   123,
			"name": "dogs",
		},
		"photoUrls": []string{"https://example.com"},
	}

	// marshal the body into bytes.
	bodyBytes, _ := json.Marshal(body) // operation returns pet

	// create a new put request
	request, _ := http.NewRequest(http.MethodGet,
		"https://hyperspace-superherbs.com/pet/IamNotANumber", nil)
	request.Header.Set("Content-Type", "application/json")

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	assert.False(t, valid)
	assert.Len(t, errors, 2)
	assert.Equal(t, "API Key api_key not found in header", errors[0].Message)
	assert.Equal(t, "Path parameter 'petId' is not a valid number", errors[1].Message)
}

func TestNewValidator_PetStore_PetGet200(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a new put request
	request, _ := http.NewRequest(http.MethodGet,
		"https://hyperspace-superherbs.com/pet/112233", nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("api_key", "12345")

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)

		// create a pet
		body := map[string]interface{}{
			"id":   123,
			"name": "cotton",
			"category": map[string]interface{}{
				"id":   123,
				"name": "dogs",
			},
			"photoUrls": []string{"https://example.com"},
		}

		// marshal the body into bytes.
		bodyBytes, _ := json.Marshal(body) // operation returns pet

		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_PetStore_PetGet200_ServerBadMediaType(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a new put request
	request, _ := http.NewRequest(http.MethodGet,
		"https://hyperspace-superherbs.com/pet/112233", nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("api_key", "12345")

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, "hot-diggity/coffee; charset=cakes") // wut?
		w.WriteHeader(http.StatusOK)

		// create a pet
		body := map[string]interface{}{
			"id":   123,
			"name": "cotton",

			"category": map[string]interface{}{
				"id":   123,
				"name": "dogs",
			},
			"photoUrls": []string{"https://example.com"},
		}

		// marshal the body into bytes.
		bodyBytes, _ := json.Marshal(body) // operation returns pet

		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "GET / 200 operation response content type 'hot-diggity/coffee' does not exist", errors[0].Message)
}

func TestNewValidator_PetStore_PetWithIdPost200_Missing200(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a new put request
	request, _ := http.NewRequest(http.MethodPost,
		"https://hyperspace-superherbs.com/pet/112233?name=peter&query=thing", nil)
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
	}

	// fire the request
	handler(res, request)

	// validate the response
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "POST operation request response code '200' does not exist", errors[0].Message)

}

func TestNewValidator_PetStore_UploadImage200_InvalidRequestBodyType(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a new put request
	request, _ := http.NewRequest(http.MethodPost,
		"https://hyperspace-superherbs.com/pet/112233/uploadImage?additionalMetadata=blem", nil)
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		// forget to write an API response
	}

	// fire the request
	handler(res, request)

	// validate the response
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	assert.False(t, valid)
	assert.Len(t, errors, 1)

}

func TestNewValidator_PetStore_UploadImage200_Valid(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a new put request
	request, _ := http.NewRequest(http.MethodPost,
		"https://hyperspace-superherbs.com/pet/112233/uploadImage?additionalMetadata=blem", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/octet-stream")

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)

		// create an API response
		body := map[string]interface{}{
			"code":    200,
			"type":    "herbs",
			"message": "smoke them every day.",
		}

		// marshal the body into bytes.
		bodyBytes, _ := json.Marshal(body) // operation returns APIResponse

		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_PetStore_UploadImage200_InvalidAPIResponse(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a new put request
	request, _ := http.NewRequest(http.MethodPost,
		"https://hyperspace-superherbs.com/pet/112233/uploadImage?additionalMetadata=blem", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/octet-stream")

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)

		// create an API response
		body := map[string]interface{}{
			"code":    200,
			"type":    false,
			"message": "smoke them every day.",
		}

		// marshal the body into bytes.
		bodyBytes, _ := json.Marshal(body) // operation returns APIResponse

		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "200 response body for '/pet/112233/uploadImage' failed to validate schema", errors[0].Message)
}

func TestNewValidator_CareRequest_WrongContentType(t *testing.T) {

	careRequestBytes, _ := os.ReadFile("test_specs/care_request.yaml")
	doc, _ := libopenapi.NewDocument(careRequestBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a new put request
	request, _ := http.NewRequest(http.MethodGet,
		"https://hyperspace-superherbs.com/requests/d4bc1a0c-c4ee-4be5-9281-26b1a041634d", nil)
	request.Header.Set("Content-Type", "application/json")

	// simulate a request/response,
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, "application/not-json")
		w.WriteHeader(http.StatusOK)

		// create a CareRequest
		body := map[string]interface{}{
			"id":     "d4bc1a0c-c4ee-4be5-9281-26b1a041634d",
			"status": "active",
		}

		// marshal the body into bytes.
		bodyBytes, _ := json.Marshal(body)

		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response
	valid, errors := v.ValidateHttpRequestResponse(request, res.Result())

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "GET / 200 operation response content type 'application/not-json' does not exist",
		errors[0].Message)

	assert.Equal(t, "The content type 'application/not-json' "+
		"of the GET response received has not been defined, it's an unknown type",
		errors[0].Reason)

}

func TestNewValidator_PetStore_InvalidPath_Response(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a new put request
	request, _ := http.NewRequest(http.MethodPost,
		"https://hyperspace-superherbs.com/missing", nil)
	request.Header.Set(helpers.ContentTypeHeader, helpers.JSONContentType)

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
	}

	// fire the request
	handler(res, request)

	// validate the response
	valid, errors := v.ValidateHttpResponse(request, res.Result())

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "POST Path '/missing' not found", errors[0].Message)
}

func TestNewValidator_PetStore_PetFindByStatusGet200_Valid_responseOnly(t *testing.T) {

	// create a new document from the petstore spec
	doc, _ := libopenapi.NewDocument(petstoreBytes)

	// create a doc
	v, _ := NewValidator(doc)

	// create a pet
	body := map[string]interface{}{
		"id":   123,
		"name": "cotton",
		"category": map[string]interface{}{
			"id":   123,
			"name": "dogs",
		},
		"photoUrls": []string{"https://example.com"},
	}

	// marshal the body into bytes.
	bodyBytes, _ := json.Marshal([]interface{}{body}) // operation returns an array of pets

	// create a new put request
	request, _ := http.NewRequest(http.MethodGet,
		"https://hyperspace-superherbs.com/pet/findByStatus?status=sold", nil)
	request.Header.Set("Content-Type", "application/json")

	// simulate a request/response, in this case the contract returns a 200 with the pet we just created.
	res := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}

	// fire the request
	handler(res, request)

	// validate the response (should be clean)
	valid, errors := v.ValidateHttpResponse(request, res.Result())

	// should all be perfectly valid.
	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_ValidateHttpResponse_RangeResponseCode(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    name:
                      type: string
                    patties:
                      type: integer
                    vegetarian:
                      type: boolean
        '4XX':
          description: Bad request
        '5XX':
          description: Server error`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	v, _ := NewValidator(doc)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers", nil)
	request.Header.Set("Content-Type", "application/json")
	response := &http.Response{
		StatusCode: 400,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
	valid, errors := v.ValidateHttpResponse(request, response)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}
