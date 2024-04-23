// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package requests

import (
	"net/http"
	"strings"

	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/errors"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/helpers"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/paths"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/utils"
)

func (v *requestBodyValidator) ValidateRequestBody(request *http.Request) (bool, []*errors.ValidationError) {
	// find path
	var pathItem = v.pathItem
	var foundPath string
	if v.pathItem == nil {
		var validationErrors []*errors.ValidationError
		pathItem, validationErrors, foundPath = paths.FindPath(request, v.document)
		if pathItem == nil || validationErrors != nil {
			v.errors = validationErrors
			return false, validationErrors
		}
	} else {
		foundPath = v.pathValue
	}

	operation := helpers.ExtractOperation(request, pathItem)
	if operation == nil {
		return false, []*errors.ValidationError{errors.OperationNotFound(pathItem, request, request.Method, foundPath)}
	}
	if operation.RequestBody == nil {
		return true, nil
	}

	// extract the content type from the request
	contentType := request.Header.Get(helpers.ContentTypeHeader)
	if contentType == "" {
		return false, []*errors.ValidationError{errors.RequestContentTypeNotFound(operation, request, foundPath)}
	}

	// extract the media type from the content type header.
	ct, _, _ := helpers.ExtractContentType(contentType)
	mediaType, ok := operation.RequestBody.Content.Get(ct)
	if !ok {
		return false, []*errors.ValidationError{errors.RequestContentTypeNotFound(operation, request, foundPath)}
	}

	// we currently only support JSON validation for request bodies
	// this will capture *everything* that contains some form of 'json' in the content type
	if !strings.Contains(strings.ToLower(contentType), helpers.JSONType) {
		return true, nil
	}

	// Nothing to validate
	if mediaType.Schema == nil {
		return true, nil
	}

	// extract schema from media type
	var schema *base.Schema
	var renderedInline, renderedJSON []byte

	// have we seen this schema before? let's hash it and check the cache.
	hash := mediaType.GoLow().Schema.Value.Hash()

	// perform work only once and cache the result in the validator.
	if cacheHit, ch := v.schemaCache.Load(hash); ch {
		// got a hit, use cached values
		schema = cacheHit.(*schemaCache).schema
		renderedInline = cacheHit.(*schemaCache).renderedInline
		renderedJSON = cacheHit.(*schemaCache).renderedJSON

	} else {

		// render the schema inline and perform the intensive work of rendering and converting
		// this is only performed once per schema and cached in the validator.
		schema = mediaType.Schema.Schema()
		renderedInline, _ = schema.RenderInline()
		renderedJSON, _ = utils.ConvertYAMLtoJSON(renderedInline)
		v.schemaCache.Store(hash, &schemaCache{
			schema:         schema,
			renderedInline: renderedInline,
			renderedJSON:   renderedJSON,
		})
	}

	// render the schema, to be used for validation
	validationSucceeded, validationErrors := ValidateRequestSchema(request, schema, renderedInline, renderedJSON)

	errors.PopulateValidationErrors(validationErrors, request, foundPath)

	return validationSucceeded, validationErrors
}
