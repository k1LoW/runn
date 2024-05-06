// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package validator

import (
	"net/http"
	"sync"

	"github.com/pb33f/libopenapi"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/errors"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/parameters"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/paths"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/requests"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/responses"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/schema_validation"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Validator provides a coarse grained interface for validating an OpenAPI 3+ documents.
// There are three primary use-cases for validation
//
// Validating *http.Request objects against and OpenAPI 3+ document
// Validating *http.Response objects against an OpenAPI 3+ document
// Validating an OpenAPI 3+ document against the OpenAPI 3+ specification
type Validator interface {

	// ValidateHttpRequest will validate an *http.Request object against an OpenAPI 3+ document.
	// The path, query, cookie and header parameters and request body are validated.
	ValidateHttpRequest(request *http.Request) (bool, []*errors.ValidationError)
	// ValidateHttpRequestSync will validate an *http.Request object against an OpenAPI 3+ document syncronously and without spawning any goroutines.
	// The path, query, cookie and header parameters and request body are validated.
	ValidateHttpRequestSync(request *http.Request) (bool, []*errors.ValidationError)

	// ValidateHttpResponse will an *http.Response object against an OpenAPI 3+ document.
	// The response body is validated. The request is only used to extract the correct reponse from the spec.
	ValidateHttpResponse(request *http.Request, response *http.Response) (bool, []*errors.ValidationError)

	// ValidateHttpRequestResponse will validate both the *http.Request and *http.Response objects against an OpenAPI 3+ document.
	// The path, query, cookie and header parameters and request and response body are validated.
	ValidateHttpRequestResponse(request *http.Request, response *http.Response) (bool, []*errors.ValidationError)

	// ValidateDocument will validate an OpenAPI 3+ document against the 3.0 or 3.1 OpenAPI 3+ specification
	ValidateDocument() (bool, []*errors.ValidationError)

	// GetParameterValidator will return a parameters.ParameterValidator instance used to validate parameters
	GetParameterValidator() parameters.ParameterValidator

	// GetRequestBodyValidator will return a parameters.RequestBodyValidator instance used to validate request bodies
	GetRequestBodyValidator() requests.RequestBodyValidator

	// GetResponseBodyValidator will return a parameters.ResponseBodyValidator instance used to validate response bodies
	GetResponseBodyValidator() responses.ResponseBodyValidator
}

// NewValidator will create a new Validator from an OpenAPI 3+ document
func NewValidator(document libopenapi.Document) (Validator, []error) {
	m, errs := document.BuildV3Model()
	if errs != nil {
		return nil, errs
	}
	v := NewValidatorFromV3Model(&m.Model)
	v.(*validator).document = document
	return v, nil
}

// NewValidatorFromV3Model will create a new Validator from an OpenAPI Model
func NewValidatorFromV3Model(m *v3.Document) Validator {
	// create a new parameter validator
	paramValidator := parameters.NewParameterValidator(m)

	// create a new request body validator
	reqBodyValidator := requests.NewRequestBodyValidator(m)

	// create a response body validator
	respBodyValidator := responses.NewResponseBodyValidator(m)

	return &validator{
		v3Model:           m,
		requestValidator:  reqBodyValidator,
		responseValidator: respBodyValidator,
		paramValidator:    paramValidator,
	}
}

func (v *validator) GetParameterValidator() parameters.ParameterValidator {
	return v.paramValidator
}
func (v *validator) GetRequestBodyValidator() requests.RequestBodyValidator {
	return v.requestValidator
}
func (v *validator) GetResponseBodyValidator() responses.ResponseBodyValidator {
	return v.responseValidator
}

func (v *validator) ValidateDocument() (bool, []*errors.ValidationError) {
	return schema_validation.ValidateOpenAPIDocument(v.document)
}

func (v *validator) ValidateHttpResponse(
	request *http.Request,
	response *http.Response) (bool, []*errors.ValidationError) {

	var pathItem *v3.PathItem
	var pathValue string
	var errs []*errors.ValidationError

	pathItem, errs, pathValue = paths.FindPath(request, v.v3Model)
	if pathItem == nil || errs != nil {
		v.errors = errs
		return false, errs
	}
	v.foundPath = pathItem
	v.foundPathValue = pathValue

	responseBodyValidator := v.responseValidator
	responseBodyValidator.SetPathItem(pathItem, pathValue)

	// validate response
	_, responseErrors := responseBodyValidator.ValidateResponseBody(request, response)

	if len(responseErrors) > 0 {
		return false, responseErrors
	}
	v.foundPath = nil
	v.foundPathValue = ""
	return true, nil
}

func (v *validator) ValidateHttpRequestResponse(
	request *http.Request,
	response *http.Response) (bool, []*errors.ValidationError) {

	var pathItem *v3.PathItem
	var pathValue string
	var errs []*errors.ValidationError

	pathItem, errs, pathValue = paths.FindPath(request, v.v3Model)
	if pathItem == nil || errs != nil {
		v.errors = errs
		return false, errs
	}
	v.foundPath = pathItem
	v.foundPathValue = pathValue

	responseBodyValidator := v.responseValidator
	responseBodyValidator.SetPathItem(pathItem, pathValue)

	// validate request and response
	_, requestErrors := v.ValidateHttpRequest(request)
	_, responseErrors := responseBodyValidator.ValidateResponseBody(request, response)

	if len(requestErrors) > 0 || len(responseErrors) > 0 {
		return false, append(requestErrors, responseErrors...)
	}
	v.foundPath = nil
	v.foundPathValue = ""
	return true, nil
}

func (v *validator) ValidateHttpRequest(request *http.Request) (bool, []*errors.ValidationError) {

	// find path
	var pathItem *v3.PathItem
	var pathValue string
	var errs []*errors.ValidationError
	if v.foundPath == nil {
		pathItem, errs, pathValue = paths.FindPath(request, v.v3Model)
		if pathItem == nil || errs != nil {
			v.errors = errs
			return false, errs
		}
		v.foundPath = pathItem
		v.foundPathValue = pathValue
	} else {
		pathItem = v.foundPath
		pathValue = v.foundPathValue
	}

	// create a new parameter validator
	paramValidator := v.paramValidator
	paramValidator.SetPathItem(pathItem, pathValue)

	// create a new request body validator
	reqBodyValidator := v.requestValidator
	reqBodyValidator.SetPathItem(pathItem, pathValue)

	// create some channels to handle async validation
	doneChan := make(chan bool)
	errChan := make(chan []*errors.ValidationError)
	controlChan := make(chan bool)

	// async param validation function.
	parameterValidationFunc := func(control chan bool, errorChan chan []*errors.ValidationError) {
		paramErrs := make(chan []*errors.ValidationError)
		paramControlChan := make(chan bool)
		paramFunctionControlChan := make(chan bool)
		var paramValidationErrors []*errors.ValidationError

		validations := []validationFunction{
			paramValidator.ValidatePathParams,
			paramValidator.ValidateCookieParams,
			paramValidator.ValidateHeaderParams,
			paramValidator.ValidateQueryParams,
			paramValidator.ValidateSecurity,
		}

		// listen for validation errors on parameters. everything will run async.
		paramListener := func(control chan bool, errorChan chan []*errors.ValidationError) {
			completedValidations := 0
			for {
				select {
				case vErrs := <-errorChan:
					paramValidationErrors = append(paramValidationErrors, vErrs...)
				case <-control:
					completedValidations++
					if completedValidations == len(validations) {
						paramFunctionControlChan <- true
						return
					}
				}
			}
		}

		validateParamFunction := func(
			control chan bool,
			errorChan chan []*errors.ValidationError,
			validatorFunc validationFunction) {
			valid, pErrs := validatorFunc(request)
			if !valid {
				errorChan <- pErrs
			}
			control <- true
		}
		go paramListener(paramControlChan, paramErrs)
		for i := range validations {
			go validateParamFunction(paramControlChan, paramErrs, validations[i])
		}

		// wait for all the validations to complete
		<-paramFunctionControlChan
		if len(paramValidationErrors) > 0 {
			errorChan <- paramValidationErrors
		}

		// let runValidation know we are done with this part.
		controlChan <- true
	}

	requestBodyValidationFunc := func(control chan bool, errorChan chan []*errors.ValidationError) {
		valid, pErrs := reqBodyValidator.ValidateRequestBody(request)
		if !valid {
			errorChan <- pErrs
		}
		control <- true
	}

	// build async functions
	asyncFunctions := []validationFunctionAsync{
		parameterValidationFunc,
		requestBodyValidationFunc,
	}

	var validationErrors []*errors.ValidationError

	// sit and wait for everything to report back.
	go runValidation(controlChan, doneChan, errChan, &validationErrors, len(asyncFunctions))

	// run async functions
	for i := range asyncFunctions {
		go asyncFunctions[i](controlChan, errChan)
	}

	// wait for all the validations to complete
	<-doneChan
	v.foundPathValue = ""
	v.foundPath = nil
	if len(validationErrors) > 0 {
		return false, validationErrors
	}
	return true, nil
}

func (v *validator) ValidateHttpRequestSync(request *http.Request) (bool, []*errors.ValidationError) {
	// find path
	var pathItem *v3.PathItem
	var pathValue string
	var errs []*errors.ValidationError
	if v.foundPath == nil {
		pathItem, errs, pathValue = paths.FindPath(request, v.v3Model)
		if pathItem == nil || errs != nil {
			v.errors = errs
			return false, errs
		}
		v.foundPath = pathItem
		v.foundPathValue = pathValue
	} else {
		pathItem = v.foundPath
		pathValue = v.foundPathValue
	}

	// create a new parameter validator
	paramValidator := v.paramValidator
	paramValidator.SetPathItem(pathItem, pathValue)

	// create a new request body validator
	reqBodyValidator := v.requestValidator
	reqBodyValidator.SetPathItem(pathItem, pathValue)

	validationErrors := make([]*errors.ValidationError, 0)

	paramValidationErrors := make([]*errors.ValidationError, 0)
	for _, validateFunc := range []validationFunction{
		paramValidator.ValidatePathParams,
		paramValidator.ValidateCookieParams,
		paramValidator.ValidateHeaderParams,
		paramValidator.ValidateQueryParams,
		paramValidator.ValidateSecurity,
	} {
		valid, pErrs := validateFunc(request)
		if !valid {
			paramValidationErrors = append(paramValidationErrors, pErrs...)
		}
	}

	valid, pErrs := reqBodyValidator.ValidateRequestBody(request)
	if !valid {
		paramValidationErrors = append(paramValidationErrors, pErrs...)
	}

	validationErrors = append(validationErrors, paramValidationErrors...)

	if len(validationErrors) > 0 {
		return false, validationErrors
	}

	return true, nil
}

type validator struct {
	v3Model           *v3.Document
	document          libopenapi.Document
	foundPath         *v3.PathItem
	foundPathValue    string
	paramValidator    parameters.ParameterValidator
	requestValidator  requests.RequestBodyValidator
	responseValidator responses.ResponseBodyValidator
	errors            []*errors.ValidationError
}

var validationLock sync.Mutex

func runValidation(control, doneChan chan bool,
	errorChan chan []*errors.ValidationError,
	validationErrors *[]*errors.ValidationError,
	total int) {
	completedValidations := 0
	for {
		select {
		case vErrs := <-errorChan:
			validationLock.Lock()
			*validationErrors = append(*validationErrors, vErrs...)
			validationLock.Unlock()
		case <-control:
			completedValidations++
			if completedValidations == total {
				doneChan <- true
				return
			}
		}
	}
}

type validationFunction func(request *http.Request) (bool, []*errors.ValidationError)
type validationFunctionAsync func(control chan bool, errorChan chan []*errors.ValidationError)
