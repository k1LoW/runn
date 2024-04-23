// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package schema_validation

import (
	"errors"
	"fmt"
	"github.com/pb33f/libopenapi"
	liberrors "github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/errors"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/helpers"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"gopkg.in/yaml.v3"
	"strings"
)

// ValidateOpenAPIDocument will validate an OpenAPI document against the OpenAPI 2, 3.0 and 3.1 schemas (depending on version)
// It will return true if the document is valid, false if it is not and a slice of ValidationError pointers.
func ValidateOpenAPIDocument(doc libopenapi.Document) (bool, []*liberrors.ValidationError) {

	info := doc.GetSpecInfo()
	loadedSchema := info.APISchema
	var validationErrors []*liberrors.ValidationError
	decodedDocument := *info.SpecJSON

	compiler := jsonschema.NewCompiler()
	_ = compiler.AddResource("schema.json", strings.NewReader(loadedSchema))
	jsch, _ := compiler.Compile("schema.json")

	scErrs := jsch.Validate(decodedDocument)

	var schemaValidationErrors []*liberrors.SchemaValidationFailure

	if scErrs != nil {

		var jk *jsonschema.ValidationError
		if errors.As(scErrs, &jk) {

			// flatten the validationErrors
			schFlatErrs := jk.BasicOutput().Errors

			for q := range schFlatErrs {
				er := schFlatErrs[q]
				if er.KeywordLocation == "" ||
					strings.HasPrefix(er.Error, "oneOf failed") ||
					strings.HasPrefix(er.Error, "allOf failed") ||
					strings.HasPrefix(er.Error, "anyOf failed") ||
					strings.HasPrefix(er.Error, "if failed") ||
					strings.HasPrefix(er.Error, "else failed") ||
					strings.HasPrefix(er.Error, "doesn't validate with") {
					continue // ignore this error, it's useless tbh, utter noise.
				}
				if er.Error != "" {

					// locate the violated property in the schema
					located := LocateSchemaPropertyNodeByJSONPath(info.RootNode.Content[0], er.InstanceLocation)
					violation := &liberrors.SchemaValidationFailure{
						Reason:           er.Error,
						Location:         er.InstanceLocation,
						DeepLocation:     er.KeywordLocation,
						AbsoluteLocation: er.AbsoluteKeywordLocation,
						OriginalError:    jk,
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
		}

		// add the error to the list
		validationErrors = append(validationErrors, &liberrors.ValidationError{
			ValidationType: helpers.Schema,
			Message:        "Document does not pass validation",
			Reason: fmt.Sprintf("OpenAPI document is not valid according "+
				"to the %s specification", info.Version),
			SchemaValidationErrors: schemaValidationErrors,
			HowToFix:               liberrors.HowToFixInvalidSchema,
		})
	}
	if len(validationErrors) > 0 {
		return false, validationErrors
	}
	return true, nil
}
