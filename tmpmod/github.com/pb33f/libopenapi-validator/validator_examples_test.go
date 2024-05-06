// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package validator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/pb33f/libopenapi"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/helpers"
)

func ExampleNewValidator_validateDocument() {
	// 1. Load the OpenAPI 3+ spec into a byte array
	petstore, err := os.ReadFile("test_specs/invalid_31.yaml")

	if err != nil {
		panic(err)
	}

	// 2. Create a new OpenAPI document using libopenapi
	document, docErrs := libopenapi.NewDocument(petstore)

	if docErrs != nil {
		panic(docErrs)
	}

	// 3. Create a new validator
	docValidator, validatorErrs := NewValidator(document)

	if validatorErrs != nil {
		panic(validatorErrs)
	}

	// 4. Validate!
	valid, validationErrs := docValidator.ValidateDocument()

	if !valid {
		for i, e := range validationErrs {
			// 5. Handle the error
			fmt.Printf("%d: Type: %s, Failure: %s\n", i, e.ValidationType, e.Message)
			fmt.Printf("Fix: %s\n\n", e.HowToFix)
		}
	}
	// Output: 0: Type: schema, Failure: Document does not pass validation
	//Fix: Ensure that the object being submitted, matches the schema correctly
}

func ExampleNewValidator_validateHttpRequest() {
	// 1. Load the OpenAPI 3+ spec into a byte array
	petstore, err := os.ReadFile("test_specs/petstorev3.json")

	if err != nil {
		panic(err)
	}

	// 2. Create a new OpenAPI document using libopenapi
	document, docErrs := libopenapi.NewDocument(petstore)

	if docErrs != nil {
		panic(docErrs)
	}

	// 3. Create a new validator
	docValidator, validatorErrs := NewValidator(document)

	if validatorErrs != nil {
		panic(validatorErrs)
	}

	// 4. Create a new *http.Request (normally, this would be where the host application will pass in the request)
	request, _ := http.NewRequest(http.MethodGet, "/pet/NotAValidPetId", nil)

	// 5. Validate!
	valid, validationErrs := docValidator.ValidateHttpRequest(request)

	if !valid {
		for _, e := range validationErrs {
			// 5. Handle the error
			fmt.Printf("Type: %s, Failure: %s\n", e.ValidationType, e.Message)
		}
	}
	// Output: Type: security, Failure: API Key api_key not found in header
	// Type: parameter, Failure: Path parameter 'petId' is not a valid number
}

func ExampleNewValidator_validateHttpRequestSync() {
	// 1. Load the OpenAPI 3+ spec into a byte array
	petstore, err := os.ReadFile("test_specs/petstorev3.json")

	if err != nil {
		panic(err)
	}

	// 2. Create a new OpenAPI document using libopenapi
	document, docErrs := libopenapi.NewDocument(petstore)

	if docErrs != nil {
		panic(docErrs)
	}

	// 3. Create a new validator
	docValidator, validatorErrs := NewValidator(document)

	if validatorErrs != nil {
		panic(validatorErrs)
	}

	// 4. Create a new *http.Request (normally, this would be where the host application will pass in the request)
	request, _ := http.NewRequest(http.MethodGet, "/pet/NotAValidPetId", nil)

	// 5. Validate!
	valid, validationErrs := docValidator.ValidateHttpRequestSync(request)

	if !valid {
		for _, e := range validationErrs {
			// 5. Handle the error
			fmt.Printf("Type: %s, Failure: %s\n", e.ValidationType, e.Message)
		}
	}
	// Type: parameter, Failure: Path parameter 'petId' is not a valid number
	// Output: Type: security, Failure: API Key api_key not found in header
}

func ExampleNewValidator_validateHttpRequestResponse() {
	// 1. Load the OpenAPI 3+ spec into a byte array
	petstore, err := os.ReadFile("test_specs/petstorev3.json")

	if err != nil {
		panic(err)
	}

	// 2. Create a new OpenAPI document using libopenapi
	document, docErrs := libopenapi.NewDocument(petstore)

	if docErrs != nil {
		panic(docErrs)
	}

	// 3. Create a new validator
	docValidator, validatorErrs := NewValidator(document)

	if validatorErrs != nil {
		panic(validatorErrs)
	}

	// 6. Create a new *http.Request (normally, this would be where the host application will pass in the request)
	request, _ := http.NewRequest(http.MethodGet, "/pet/findByStatus?status=sold", nil)

	// 7. Simulate a request/response, in this case the contract returns a 200 with an array of pets.
	// Normally, this would be where the host application would pass in the response.
	recorder := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {

		// set return content type.
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)

		// create a Pet
		body := map[string]interface{}{
			"id":   123,
			"name": "cotton",
			"category": map[string]interface{}{
				"id":   "NotAValidPetId", // this will fail, it should be an integer.
				"name": "dogs",
			},
			"photoUrls": []string{"https://pb33f.io"},
		}

		// marshal the request body into bytes.
		responseBodyBytes, _ := json.Marshal([]interface{}{body}) // operation returns an array of pets
		// return the response.
		_, _ = w.Write(responseBodyBytes)
	}

	// simulate request/response
	handler(recorder, request)

	// 7. Validate!
	valid, validationErrs := docValidator.ValidateHttpRequestResponse(request, recorder.Result())

	if !valid {
		for _, e := range validationErrs {
			// 5. Handle the error
			fmt.Printf("Type: %s, Failure: %s\n", e.ValidationType, e.Message)
			fmt.Printf("Schema Error: %s, Line: %d, Col: %d\n",
				e.SchemaValidationErrors[0].Reason,
				e.SchemaValidationErrors[0].Line,
				e.SchemaValidationErrors[0].Column)
		}
	}
	// Output: Type: response, Failure: 200 response body for '/pet/findByStatus' failed to validate schema
	//Schema Error: expected integer, but got string, Line: 19, Col: 27
}

func ExampleNewValidator_validateHttpResponse() {
	// 1. Load the OpenAPI 3+ spec into a byte array
	petstore, err := os.ReadFile("test_specs/petstorev3.json")

	if err != nil {
		panic(err)
	}

	// 2. Create a new OpenAPI document using libopenapi
	document, docErrs := libopenapi.NewDocument(petstore)

	if docErrs != nil {
		panic(docErrs)
	}

	// 3. Create a new validator
	docValidator, validatorErrs := NewValidator(document)

	if validatorErrs != nil {
		panic(validatorErrs)
	}

	// 6. Create a new *http.Request (normally, this would be where the host application will pass in the request)
	request, _ := http.NewRequest(http.MethodGet, "/pet/findByStatus?status=sold", nil)

	// 7. Simulate a request/response, in this case the contract returns a 200 with an array of pets.
	// Normally, this would be where the host application would pass in the response.
	recorder := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {

		// set return content type.
		w.Header().Set(helpers.ContentTypeHeader, helpers.JSONContentType)
		w.WriteHeader(http.StatusOK)

		// create a Pet
		body := map[string]interface{}{
			"id":   123,
			"name": "cotton",
			"category": map[string]interface{}{
				"id":   "NotAValidPetId", // this will fail, it should be an integer.
				"name": "dogs",
			},
			"photoUrls": []string{"https://pb33f.io"},
		}

		// marshal the request body into bytes.
		responseBodyBytes, _ := json.Marshal([]interface{}{body}) // operation returns an array of pets
		// return the response.
		_, _ = w.Write(responseBodyBytes)
	}

	// simulate request/response
	handler(recorder, request)

	// 7. Validate the response only
	valid, validationErrs := docValidator.ValidateHttpResponse(request, recorder.Result())

	if !valid {
		for _, e := range validationErrs {
			// 5. Handle the error
			fmt.Printf("Type: %s, Failure: %s\n", e.ValidationType, e.Message)
			fmt.Printf("Schema Error: %s, Line: %d, Col: %d\n",
				e.SchemaValidationErrors[0].Reason,
				e.SchemaValidationErrors[0].Line,
				e.SchemaValidationErrors[0].Column)
		}
	}
	// Output: Type: response, Failure: 200 response body for '/pet/findByStatus' failed to validate schema
	//Schema Error: expected integer, but got string, Line: 19, Col: 27
}
