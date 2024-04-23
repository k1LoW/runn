// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/errors"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/helpers"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/schema_validation"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var instanceLocationRegex = regexp.MustCompile(`^/(\d+)`)

// ValidateRequestSchema will validate a http.Request pointer against a schema.
// If validation fails, it will return a list of validation errors as the second return value.
func ValidateRequestSchema(
	request *http.Request,
	schema *base.Schema,
	renderedSchema,
	jsonSchema []byte) (bool, []*errors.ValidationError) {

	var validationErrors []*errors.ValidationError

	var requestBody []byte
	if request != nil && request.Body != nil {
		requestBody, _ = io.ReadAll(request.Body)

		// close the request body, so it can be re-read later by another player in the chain
		_ = request.Body.Close()
		request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

	}

	var decodedObj interface{}

	if len(requestBody) > 0 {
		err := json.Unmarshal(requestBody, &decodedObj)

		if err != nil {
			// cannot decode the request body, so it's not valid
			violation := &errors.SchemaValidationFailure{
				Reason:          err.Error(),
				Location:        "unavailable",
				ReferenceSchema: string(renderedSchema),
				ReferenceObject: string(requestBody),
			}
			validationErrors = append(validationErrors, &errors.ValidationError{
				ValidationType:    helpers.RequestBodyValidation,
				ValidationSubType: helpers.Schema,
				Message: fmt.Sprintf("%s request body for '%s' failed to validate schema",
					request.Method, request.URL.Path),
				Reason:                 fmt.Sprintf("The request body cannot be decoded: %s", err.Error()),
				SpecLine:               1,
				SpecCol:                0,
				SchemaValidationErrors: []*errors.SchemaValidationFailure{violation},
				HowToFix:               errors.HowToFixInvalidSchema,
				Context:                string(renderedSchema), // attach the rendered schema to the error
			})
			return false, validationErrors
		}
	}

	// no request body? but we do have a schema?
	if len(requestBody) <= 0 && len(jsonSchema) > 0 {
		// cannot decode the request body, so it's not valid
		violation := &errors.SchemaValidationFailure{
			Reason:          "request body is empty, but there is a schema defined",
			ReferenceSchema: string(renderedSchema),
			ReferenceObject: string(requestBody),
		}
		validationErrors = append(validationErrors, &errors.ValidationError{
			ValidationType:    helpers.RequestBodyValidation,
			ValidationSubType: helpers.Schema,
			Message: fmt.Sprintf("%s request body is empty for '%s'",
				request.Method, request.URL.Path),
			Reason:                 "The request body is empty but there is a schema defined",
			SpecLine:               schema.GoLow().Type.KeyNode.Line,
			SpecCol:                schema.GoLow().Type.KeyNode.Line,
			SchemaValidationErrors: []*errors.SchemaValidationFailure{violation},
			HowToFix:               errors.HowToFixInvalidSchema,
			Context:                string(renderedSchema), // attach the rendered schema to the error
		})
		return false, validationErrors
	}

	compiler := jsonschema.NewCompiler()
	_ = compiler.AddResource("requestBody.json", strings.NewReader(string(jsonSchema)))
	jsch, _ := compiler.Compile("requestBody.json")

	// validate the object against the schema
	scErrs := jsch.Validate(decodedObj)
	if scErrs != nil {
		jk := scErrs.(*jsonschema.ValidationError)

		// flatten the validationErrors
		schFlatErrs := jk.BasicOutput().Errors
		var schemaValidationErrors []*errors.SchemaValidationFailure
		for q := range schFlatErrs {
			er := schFlatErrs[q]
			if er.KeywordLocation == "" || strings.HasPrefix(er.Error, "doesn't validate with") {
				continue // ignore this error, it's useless tbh, utter noise.
			}
			if er.Error != "" {

				// re-encode the schema.
				var renderedNode yaml.Node
				_ = yaml.Unmarshal(renderedSchema, &renderedNode)

				// locate the violated property in the schema
				located := schema_validation.LocateSchemaPropertyNodeByJSONPath(renderedNode.Content[0], er.KeywordLocation)

				// extract the element specified by the instance
				val := instanceLocationRegex.FindStringSubmatch(er.InstanceLocation)
				var referenceObject string

				if len(val) > 0 {
					referenceIndex, _ := strconv.Atoi(val[1])
					if reflect.ValueOf(decodedObj).Type().Kind() == reflect.Slice {
						found := decodedObj.([]any)[referenceIndex]
						recoded, _ := json.MarshalIndent(found, "", "  ")
						referenceObject = string(recoded)
					}
				}
				if referenceObject == "" {
					referenceObject = string(requestBody)
				}

				violation := &errors.SchemaValidationFailure{
					Reason:          er.Error,
					Location:        er.KeywordLocation,
					ReferenceSchema: string(renderedSchema),
					ReferenceObject: referenceObject,
					OriginalError:   jk,
				}
				// if we have a location within the schema, add it to the error
				if located != nil {

					line := located.Line
					// if the located node is a map or an array, then the actual human interpretable
					// line on which the violation occurred is the line of the key, not the value.
					if located.Kind == yaml.MappingNode || located.Kind == yaml.SequenceNode {
						if line > 0 {
							line--
						}
					}

					// location of the violation within the rendered schema.
					violation.Line = line
					violation.Column = located.Column
				}
				schemaValidationErrors = append(schemaValidationErrors, violation)
			}
		}

		line := 1
		col := 0
		if schema.GoLow().Type.KeyNode != nil {
			line = schema.GoLow().Type.KeyNode.Line
			col = schema.GoLow().Type.KeyNode.Column
		}

		// add the error to the list
		validationErrors = append(validationErrors, &errors.ValidationError{
			ValidationType:    helpers.RequestBodyValidation,
			ValidationSubType: helpers.Schema,
			Message: fmt.Sprintf("%s request body for '%s' failed to validate schema",
				request.Method, request.URL.Path),
			Reason: "The request body is defined as an object. " +
				"However, it does not meet the schema requirements of the specification",
			SpecLine:               line,
			SpecCol:                col,
			SchemaValidationErrors: schemaValidationErrors,
			HowToFix:               errors.HowToFixInvalidSchema,
			Context:                string(renderedSchema), // attach the rendered schema to the error
		})
	}
	if len(validationErrors) > 0 {
		return false, validationErrors
	}
	return true, nil
}
