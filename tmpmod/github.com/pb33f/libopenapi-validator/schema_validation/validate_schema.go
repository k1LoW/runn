// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package schema_validation

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	liberrors "github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/errors"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/helpers"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/utils"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// SchemaValidator is an interface that defines the methods for validating a *base.Schema (V3+ Only) object.
// There are 3 methods for validating a schema:
//
//	ValidateSchemaString accepts a schema object to validate against, and a JSON/YAML blob that is defined as a string.
//	ValidateSchemaObject accepts a schema object to validate against, and an object, created from unmarshalled JSON/YAML.
//	ValidateSchemaBytes accepts a schema object to validate against, and a JSON/YAML blob that is defined as a byte array.
type SchemaValidator interface {

	// ValidateSchemaString accepts a schema object to validate against, and a JSON/YAML blob that is defined as a string.
	ValidateSchemaString(schema *base.Schema, payload string) (bool, []*liberrors.ValidationError)

	// ValidateSchemaObject accepts a schema object to validate against, and an object, created from unmarshalled JSON/YAML.
	// This is a pre-decoded object that will skip the need to unmarshal a string of JSON/YAML.
	ValidateSchemaObject(schema *base.Schema, payload interface{}) (bool, []*liberrors.ValidationError)

	// ValidateSchemaBytes accepts a schema object to validate against, and a byte slice containing a schema to
	// validate against.
	ValidateSchemaBytes(schema *base.Schema, payload []byte) (bool, []*liberrors.ValidationError)
}

var instanceLocationRegex = regexp.MustCompile(`^/(\d+)`)

type schemaValidator struct {
	logger *slog.Logger
	lock   sync.Mutex
}

// NewSchemaValidatorWithLogger will create a new SchemaValidator instance, ready to accept schemas and payloads to validate.
func NewSchemaValidatorWithLogger(logger *slog.Logger) SchemaValidator {
	return &schemaValidator{logger: logger, lock: sync.Mutex{}}
}

// NewSchemaValidator will create a new SchemaValidator instance, ready to accept schemas and payloads to validate.
func NewSchemaValidator() SchemaValidator {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	return NewSchemaValidatorWithLogger(logger)
}

func (s *schemaValidator) ValidateSchemaString(schema *base.Schema, payload string) (bool, []*liberrors.ValidationError) {
	return s.validateSchema(schema, []byte(payload), nil, s.logger)
}

func (s *schemaValidator) ValidateSchemaObject(schema *base.Schema, payload interface{}) (bool, []*liberrors.ValidationError) {
	return s.validateSchema(schema, nil, payload, s.logger)
}

func (s *schemaValidator) ValidateSchemaBytes(schema *base.Schema, payload []byte) (bool, []*liberrors.ValidationError) {
	return s.validateSchema(schema, payload, nil, s.logger)
}

func (s *schemaValidator) validateSchema(schema *base.Schema, payload []byte, decodedObject interface{}, log *slog.Logger) (bool, []*liberrors.ValidationError) {

	var validationErrors []*liberrors.ValidationError

	if schema == nil {
		log.Info("schema is empty and cannot be validated. This generally means the schema is missing from the spec, or could not be read.")
		return false, validationErrors
	}

	// extract index of schema, and check the version
	//schemaIndex := schema.GoLow().Index
	var renderedSchema []byte

	// render the schema, to be used for validation, stop this from running concurrently, mutations are made to state
	// and, it will cause async issues.
	s.lock.Lock()
	renderedSchema, _ = schema.RenderInline()
	s.lock.Unlock()

	jsonSchema, _ := utils.ConvertYAMLtoJSON(renderedSchema)

	if decodedObject == nil && len(payload) > 0 {
		err := json.Unmarshal(payload, &decodedObject)

		if err != nil {
			// cannot decode the request body, so it's not valid
			violation := &liberrors.SchemaValidationFailure{
				Reason:          err.Error(),
				Location:        "unavailable",
				ReferenceSchema: string(renderedSchema),
				ReferenceObject: string(payload),
			}
			validationErrors = append(validationErrors, &liberrors.ValidationError{
				ValidationType:         helpers.RequestBodyValidation,
				ValidationSubType:      helpers.Schema,
				Message:                "schema does not pass validation",
				Reason:                 fmt.Sprintf("The schema cannot be decoded: %s", err.Error()),
				SpecLine:               1,
				SpecCol:                0,
				SchemaValidationErrors: []*liberrors.SchemaValidationFailure{violation},
				HowToFix:               liberrors.HowToFixInvalidSchema,
				Context:                string(renderedSchema), // attach the rendered schema to the error
			})
			return false, validationErrors
		}

	}
	compiler := jsonschema.NewCompiler()

	_ = compiler.AddResource("schema.json", strings.NewReader(string(jsonSchema)))
	jsch, err := compiler.Compile("schema.json")

	var schemaValidationErrors []*liberrors.SchemaValidationFailure

	// is the schema even valid? did it compile?
	if err != nil {
		var se *jsonschema.SchemaError
		if errors.As(err, &se) {
			var ve *jsonschema.ValidationError
			if errors.As(se.Err, &ve) {

				// no, this won't work, so we need to extract the errors and return them.
				basicErrors := ve.BasicOutput().Errors
				schemaValidationErrors = extractBasicErrors(basicErrors, renderedSchema, decodedObject, payload, ve, schemaValidationErrors)
				// cannot compile schema, so it's not valid
				violation := &liberrors.SchemaValidationFailure{
					Reason:          err.Error(),
					Location:        "unavailable",
					ReferenceSchema: string(renderedSchema),
					ReferenceObject: string(payload),
				}
				validationErrors = append(validationErrors, &liberrors.ValidationError{
					ValidationType:         helpers.RequestBodyValidation,
					ValidationSubType:      helpers.Schema,
					Message:                "schema does not pass validation",
					Reason:                 fmt.Sprintf("The schema cannot be decoded: %s", err.Error()),
					SpecLine:               1,
					SpecCol:                0,
					SchemaValidationErrors: []*liberrors.SchemaValidationFailure{violation},
					HowToFix:               liberrors.HowToFixInvalidSchema,
					Context:                string(renderedSchema), // attach the rendered schema to the error
				})
				return false, validationErrors
			}
		}
	}

	// 4. validate the object against the schema
	if jsch != nil && decodedObject != nil {
		scErrs := jsch.Validate(decodedObject)
		if scErrs != nil {

			// check for invalid JSON type errors.
			var invalidJSONTypeError jsonschema.InvalidJSONTypeError
			if errors.As(scErrs, &invalidJSONTypeError) {
				violation := &liberrors.SchemaValidationFailure{
					Reason:   scErrs.Error(),
					Location: "unavailable", // we don't have a location for this error, so we'll just say it's unavailable.
				}
				schemaValidationErrors = append(schemaValidationErrors, violation)
			}

			var jk *jsonschema.ValidationError
			if errors.As(scErrs, &jk) {

				// flatten the validationErrors
				schFlatErrs := jk.BasicOutput().Errors

				schemaValidationErrors = extractBasicErrors(schFlatErrs, renderedSchema, decodedObject, payload, jk, schemaValidationErrors)
			}
			line := 1
			col := 0
			if schema.GoLow().Type.KeyNode != nil {
				line = schema.GoLow().Type.KeyNode.Line
				col = schema.GoLow().Type.KeyNode.Column
			}

			// add the error to the list
			validationErrors = append(validationErrors, &liberrors.ValidationError{
				ValidationType:         helpers.Schema,
				Message:                "schema does not pass validation",
				Reason:                 "Schema failed to validate against the contract requirements",
				SpecLine:               line,
				SpecCol:                col,
				SchemaValidationErrors: schemaValidationErrors,
				HowToFix:               liberrors.HowToFixInvalidSchema,
				Context:                string(renderedSchema), // attach the rendered schema to the error
			})
		}
	}
	if len(validationErrors) > 0 {
		return false, validationErrors
	}
	return true, nil
}

func extractBasicErrors(schFlatErrs []jsonschema.BasicError,
	renderedSchema []byte, decodedObject interface{},
	payload []byte, jk *jsonschema.ValidationError,
	schemaValidationErrors []*liberrors.SchemaValidationFailure) []*liberrors.SchemaValidationFailure {
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
			located := LocateSchemaPropertyNodeByJSONPath(renderedNode.Content[0], er.KeywordLocation)

			// extract the element specified by the instance
			val := instanceLocationRegex.FindStringSubmatch(er.InstanceLocation)
			var referenceObject string

			if len(val) > 0 {
				referenceIndex, _ := strconv.Atoi(val[1])
				if reflect.ValueOf(decodedObject).Type().Kind() == reflect.Slice {
					found := decodedObject.([]any)[referenceIndex]
					recoded, _ := json.MarshalIndent(found, "", "  ")
					referenceObject = string(recoded)
				}
			}
			if referenceObject == "" {
				referenceObject = string(payload)
			}

			violation := &liberrors.SchemaValidationFailure{
				Reason:           er.Error,
				Location:         er.InstanceLocation,
				DeepLocation:     er.KeywordLocation,
				AbsoluteLocation: er.AbsoluteKeywordLocation,
				ReferenceSchema:  string(renderedSchema),
				ReferenceObject:  referenceObject,
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
	return schemaValidationErrors
}
