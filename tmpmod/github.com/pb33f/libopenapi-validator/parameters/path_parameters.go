// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package parameters

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/errors"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/helpers"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/paths"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

func (v *paramValidator) ValidatePathParams(request *http.Request) (bool, []*errors.ValidationError) {

	// find path
	var pathItem *v3.PathItem
	var errs []*errors.ValidationError
	var foundPath string
	if v.pathItem == nil && v.pathValue == "" {
		pathItem, errs, foundPath = paths.FindPath(request, v.document)
		if pathItem == nil || errs != nil {
			v.errors = errs
			return false, errs
		}
	} else {
		pathItem = v.pathItem
		foundPath = v.pathValue
	}

	// split the path into segments
	submittedSegments := strings.Split(paths.StripRequestPath(request, v.document), helpers.Slash)
	pathSegments := strings.Split(foundPath, helpers.Slash)

	// extract params for the operation
	var params = helpers.ExtractParamsForOperation(request, pathItem)
	var validationErrors []*errors.ValidationError
	for _, p := range params {
		if p.In == helpers.Path {

			// var paramTemplate string
			for x := range pathSegments {
				if pathSegments[x] == "" { // skip empty segments
					continue
				}
				i := strings.IndexRune(pathSegments[x], '{')
				if i > -1 {
					isMatrix := false
					isLabel := false
					// isExplode := false
					isSimple := true
					paramTemplate := pathSegments[x][i+1 : len(pathSegments[x])-1]
					paramName := paramTemplate
					// check for an asterisk on the end of the parameter (explode)
					if strings.HasSuffix(paramTemplate, helpers.Asterisk) {
						// isExplode = true
						paramName = paramTemplate[:len(paramTemplate)-1]
					}
					if strings.HasPrefix(paramTemplate, helpers.Period) {
						isLabel = true
						isSimple = false
						paramName = paramName[1:]
					}
					if strings.HasPrefix(paramTemplate, helpers.SemiColon) {
						isMatrix = true
						isSimple = false
						paramName = paramName[1:]
					}

					// does this param name match the current path segment param name
					if paramName != p.Name {
						continue
					}

					paramValue := ""

					// extract the parameter value from the path.
					if x < len(submittedSegments) {
						paramValue = submittedSegments[x]
					}

					if paramValue == "" {
						// TODO: check path match issue.
						continue
					}

					// extract the schema from the parameter
					sch := p.Schema.Schema()

					// check enum (if present)
					enumCheck := func(paramValue string) {
						matchFound := false
						for _, enumVal := range sch.Enum {
							if strings.TrimSpace(paramValue) == fmt.Sprint(enumVal.Value) {
								matchFound = true
								break
							}
						}
						if !matchFound {
							validationErrors = append(validationErrors,
								errors.IncorrectPathParamEnum(p, strings.ToLower(paramValue), sch))
						}
					}

					// for each type, check the value.
					for typ := range sch.Type {

						switch sch.Type[typ] {
						case helpers.String:

							// TODO: label and matrix style validation

							// check if the param is within the enum
							if sch.Enum != nil {
								enumCheck(paramValue)
								break
							}
							validationErrors = append(validationErrors,
								ValidateSingleParameterSchema(
									sch,
									paramValue,
									"Path parameter",
									"The path parameter",
									p.Name,
									helpers.ParameterValidation,
									helpers.ParameterValidationPath,
								)...)

						case helpers.Integer, helpers.Number:
							// simple use case is already handled in find param.
							rawParamValue, paramValueParsed, err := v.resolveNumber(sch, p, isLabel, isMatrix, paramValue)
							if err != nil {
								validationErrors = append(validationErrors, err...)
								break
							}
							// check if the param is within the enum
							if sch.Enum != nil {
								enumCheck(rawParamValue)
								break
							}
							validationErrors = append(validationErrors, ValidateSingleParameterSchema(
								sch,
								paramValueParsed,
								"Path parameter",
								"The path parameter",
								p.Name,
								helpers.ParameterValidation,
								helpers.ParameterValidationPath,
							)...)

						case helpers.Boolean:
							if isLabel && p.Style == helpers.LabelStyle {
								if _, err := strconv.ParseFloat(paramValue[1:], 64); err != nil {
									validationErrors = append(validationErrors,
										errors.IncorrectPathParamBool(p, paramValue[1:], sch))
								}
							}
							if isSimple {
								if _, err := strconv.ParseBool(paramValue); err != nil {
									validationErrors = append(validationErrors,
										errors.IncorrectPathParamBool(p, paramValue, sch))
								}
							}
							if isMatrix && p.Style == helpers.MatrixStyle {
								// strip off the colon and the parameter name
								paramValue = strings.Replace(paramValue[1:], fmt.Sprintf("%s=", p.Name), "", 1)
								if _, err := strconv.ParseBool(paramValue); err != nil {
									validationErrors = append(validationErrors,
										errors.IncorrectPathParamBool(p, paramValue, sch))
								}
							}
						case helpers.Object:
							var encodedObject interface{}

							if p.IsDefaultPathEncoding() {
								encodedObject = helpers.ConstructMapFromCSV(paramValue)
							} else {
								switch p.Style {
								case helpers.LabelStyle:
									if !p.IsExploded() {
										encodedObject = helpers.ConstructMapFromCSV(paramValue[1:])
									} else {
										encodedObject = helpers.ConstructKVFromLabelEncoding(paramValue)
									}
								case helpers.MatrixStyle:
									if !p.IsExploded() {
										paramValue = strings.Replace(paramValue[1:], fmt.Sprintf("%s=", p.Name), "", 1)
										encodedObject = helpers.ConstructMapFromCSV(paramValue)
									} else {
										paramValue = strings.Replace(paramValue[1:], fmt.Sprintf("%s=", p.Name), "", 1)
										encodedObject = helpers.ConstructKVFromMatrixCSV(paramValue)
									}
								default:
									if p.IsExploded() {
										encodedObject = helpers.ConstructKVFromCSV(paramValue)
									}
								}
							}
							// if a schema was extracted
							if sch != nil {
								validationErrors = append(validationErrors,
									ValidateParameterSchema(sch,
										encodedObject,
										"",
										"Path parameter",
										"The path parameter",
										p.Name,
										helpers.ParameterValidation,
										helpers.ParameterValidationPath)...)
							}

						case helpers.Array:

							// extract the items schema in order to validate the array items.
							if sch.Items != nil && sch.Items.IsA() {
								iSch := sch.Items.A.Schema()
								for n := range iSch.Type {
									// determine how to explode the array
									var arrayValues []string
									if isSimple {
										arrayValues = strings.Split(paramValue, helpers.Comma)
									}
									if isLabel {
										if !p.IsExploded() {
											arrayValues = strings.Split(paramValue[1:], helpers.Comma)
										} else {
											arrayValues = strings.Split(paramValue[1:], helpers.Period)
										}
									}
									if isMatrix {
										if !p.IsExploded() {
											paramValue = strings.Replace(paramValue[1:], fmt.Sprintf("%s=", p.Name), "", 1)
											arrayValues = strings.Split(paramValue, helpers.Comma)
										} else {
											paramValue = strings.ReplaceAll(paramValue[1:], fmt.Sprintf("%s=", p.Name), "")
											arrayValues = strings.Split(paramValue, helpers.SemiColon)
										}
									}
									switch iSch.Type[n] {
									case helpers.Integer, helpers.Number:
										for pv := range arrayValues {
											if _, err := strconv.ParseFloat(arrayValues[pv], 64); err != nil {
												validationErrors = append(validationErrors,
													errors.IncorrectPathParamArrayNumber(p, arrayValues[pv], sch, iSch))
											}
										}
									case helpers.Boolean:
										for pv := range arrayValues {
											bc := len(validationErrors)
											if _, err := strconv.ParseBool(arrayValues[pv]); err != nil {
												validationErrors = append(validationErrors,
													errors.IncorrectPathParamArrayBoolean(p, arrayValues[pv], sch, iSch))
												continue
											}
											if len(validationErrors) == bc {
												// ParseBool will parse 0 or 1 as false/true to we
												// need to catch this edge case.
												if arrayValues[pv] == "0" || arrayValues[pv] == "1" {
													validationErrors = append(validationErrors,
														errors.IncorrectPathParamArrayBoolean(p, arrayValues[pv], sch, iSch))
													continue
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	errors.PopulateValidationErrors(validationErrors, request, foundPath)

	if len(validationErrors) > 0 {
		return false, validationErrors
	}
	return true, nil
}

func (v *paramValidator) resolveNumber(sch *base.Schema, p *v3.Parameter, isLabel bool, isMatrix bool, paramValue string) (string, float64, []*errors.ValidationError) {
	if isLabel && p.Style == helpers.LabelStyle {
		paramValueParsed, err := strconv.ParseFloat(paramValue[1:], 64)
		if err != nil {
			return "", 0, []*errors.ValidationError{errors.IncorrectPathParamNumber(p, paramValue[1:], sch)}
		}
		return paramValue[1:], paramValueParsed, nil
	}
	if isMatrix && p.Style == helpers.MatrixStyle {
		// strip off the colon and the parameter name
		paramValue = strings.Replace(paramValue[1:], fmt.Sprintf("%s=", p.Name), "", 1)
		paramValueParsed, err := strconv.ParseFloat(paramValue, 64)
		if err != nil {
			return "", 0, []*errors.ValidationError{errors.IncorrectPathParamNumber(p, paramValue[1:], sch)}
		}
		return paramValue, paramValueParsed, nil
	}
	paramValueParsed, err := strconv.ParseFloat(paramValue, 64)
	if err != nil {
		return "", 0, []*errors.ValidationError{errors.IncorrectPathParamNumber(p, paramValue[1:], sch)}
	}
	return paramValue, paramValueParsed, nil
}
