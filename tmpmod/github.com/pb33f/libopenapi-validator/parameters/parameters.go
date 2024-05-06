// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package parameters

import (
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/errors"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"net/http"
)

// ParameterValidator is an interface that defines the methods for validating parameters
// There are 4 types of parameters: query, header, cookie and path.
//
//	ValidateQueryParams will validate the query parameters for the request
//	ValidateHeaderParams will validate the header parameters for the request
//	ValidateCookieParams will validate the cookie parameters for the request
//	ValidatePathParams will validate the path parameters for the request
//
// Each method accepts an *http.Request and returns true if validation passed,
// false if validation failed and a slice of ValidationError pointers.
type ParameterValidator interface {

	// SetPathItem will set the pathItem for the ParameterValidator, all validations will be performed against this pathItem
	// otherwise if not set, each validation will perform a lookup for the pathItem based on the *http.Request
	SetPathItem(path *v3.PathItem, pathValue string)

	// ValidateQueryParams accepts an *http.Request and validates the query parameters against the OpenAPI specification.
	// The method will locate the correct path, and operation, based on the verb. The parameters for the operation
	// will be matched and validated against what has been supplied in the http.Request query string.
	ValidateQueryParams(request *http.Request) (bool, []*errors.ValidationError)

	// ValidateHeaderParams validates the header parameters contained within *http.Request. It returns a boolean
	// stating true if validation passed (false for failed), and a slice of errors if validation failed.
	ValidateHeaderParams(request *http.Request) (bool, []*errors.ValidationError)

	// ValidateCookieParams validates the cookie parameters contained within *http.Request.
	// It returns a boolean stating true if validation passed (false for failed), and a slice of errors if validation failed.
	ValidateCookieParams(request *http.Request) (bool, []*errors.ValidationError)

	// ValidatePathParams validates the path parameters contained within *http.Request. It returns a boolean stating true
	// if validation passed (false for failed), and a slice of errors if validation failed.
	ValidatePathParams(request *http.Request) (bool, []*errors.ValidationError)

	// ValidateSecurity validates the security requirements for the operation. It returns a boolean stating true
	// if validation passed (false for failed), and a slice of errors if validation failed.
	ValidateSecurity(request *http.Request) (bool, []*errors.ValidationError)
}

func (v *paramValidator) SetPathItem(path *v3.PathItem, pathValue string) {
	v.pathItem = path
	v.pathValue = pathValue
}

// NewParameterValidator will create a new ParameterValidator from an OpenAPI 3+ document
func NewParameterValidator(document *v3.Document) ParameterValidator {
	return &paramValidator{document: document}
}

type paramValidator struct {
	document  *v3.Document
	pathItem  *v3.PathItem
	pathValue string
	errors    []*errors.ValidationError
}
