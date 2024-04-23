// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package schema_validation

import (
	"encoding/json"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestLocateSchemaPropertyNodeByJSONPath(t *testing.T) {
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

	var node yaml.Node
	_ = yaml.Unmarshal([]byte(spec), &node)

	foundNode := LocateSchemaPropertyNodeByJSONPath(node.Content[0],
		"/paths/~1burgers~1createBurger/post/requestBody/content/application~1json/schema/properties/vegetarian")

	assert.Equal(t, "boolean", foundNode.Content[1].Value)

	foundNode = LocateSchemaPropertyNodeByJSONPath(node.Content[0],
		"/i/do/not/exist")

	assert.Nil(t, foundNode)
}

func TestValidateSchema_SimpleValid_String(t *testing.T) {
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

	m, _ := doc.BuildV3Model()

	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    2,
		"vegetarian": true,
	}

	bodyBytes, _ := json.Marshal(body)
	sch := m.Model.Paths.PathItems.GetOrZero("/burgers/createBurger").Post.RequestBody.Content.GetOrZero("application/json").Schema

	// create a schema validator
	v := NewSchemaValidator()

	// validate!
	valid, errors := v.ValidateSchemaString(sch.Schema(), string(bodyBytes))

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestValidateSchema_SimpleValid(t *testing.T) {
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

	m, _ := doc.BuildV3Model()

	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    2,
		"vegetarian": true,
	}

	// create a schema validator
	v := NewSchemaValidator()

	bodyBytes, _ := json.Marshal(body)
	sch := m.Model.Paths.PathItems.GetOrZero("/burgers/createBurger").Post.RequestBody.Content.GetOrZero("application/json").Schema

	// validate!
	valid, errors := v.ValidateSchemaBytes(sch.Schema(), bodyBytes)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestValidateSchema_SimpleInValid(t *testing.T) {
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

	m, _ := doc.BuildV3Model()

	body := map[string]interface{}{
		"name":       "Big Mac",
		"patties":    "I am not a number", // will fail
		"vegetarian": 23,                  // will fail
	}

	// create a schema validator
	v := NewSchemaValidator()

	bodyBytes, _ := json.Marshal(body)
	sch := m.Model.Paths.PathItems.GetOrZero("/burgers/createBurger").Post.RequestBody.Content.GetOrZero("application/json").Schema

	// validate!
	valid, errors := v.ValidateSchemaBytes(sch.Schema(), bodyBytes)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 2)
}

func TestValidateSchema_InvalidJSONType(t *testing.T) {
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

	m, _ := doc.BuildV3Model()

	body := struct{ name string }{"hello world"}

	// create a schema validator
	v := NewSchemaValidator()

	// bodyBytes, _ := json.Marshal(body)
	sch := m.Model.Paths.PathItems.GetOrZero("/burgers/createBurger").Post.RequestBody.Content.GetOrZero("application/json").Schema

	// validate!
	valid, errors := v.ValidateSchemaObject(sch.Schema(), body)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 1)
	assert.Equal(t, "jsonschema: invalid jsonType: struct { name string }", errors[0].SchemaValidationErrors[0].Reason)
}

func TestValidateSchema_ReffyComplex_Valid(t *testing.T) {
	spec := `openapi: 3.1.0
components:
  schemas:
    Death:
      type: object
      required: [cakeOrDeath]
      properties:
        cakeOrDeath:
          type: string
          enum: [death]
    Cake:
      type: object
      required: [cakeOrDeath]
      properties:
        cakeOrDeath:
          type: string
          enum: [cake please]
    Four:
      type: object
      oneOf:
        - $ref: '#/components/schemas/Cake'
        - $ref: '#/components/schemas/Death'
    Three:
      type: object
      properties:
        name:
          type: string
        four:
          $ref: '#/components/schemas/Four'
    Two:
      type: object
      properties:
        name:
          type: string
        three:
          $ref: '#/components/schemas/Three'
    One:
      type: object
      properties:
        name:
          type: string
        two:
          $ref: '#/components/schemas/Two'
paths:
  /burgers/createBurger:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/One'`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	cakePlease := map[string]interface{}{
		"two": map[string]interface{}{
			"three": map[string]interface{}{
				"four": map[string]interface{}{
					"cakeOrDeath": "cake please",
				},
			},
		},
	}

	death := map[string]interface{}{
		"two": map[string]interface{}{
			"three": map[string]interface{}{
				"four": map[string]interface{}{
					"cakeOrDeath": "death",
				},
			},
		},
	}

	// cake? (https://www.youtube.com/watch?v=PVH0gZO5lq0)
	bodyBytes, _ := json.Marshal(cakePlease)
	sch := m.Model.Paths.PathItems.GetOrZero("/burgers/createBurger").Post.RequestBody.Content.GetOrZero("application/json").Schema

	// create a schema validator
	v := NewSchemaValidator()

	// validate!
	valid, errors := v.ValidateSchemaBytes(sch.Schema(), bodyBytes)

	assert.True(t, valid)
	assert.Len(t, errors, 0)

	// or death!
	bodyBytes, _ = json.Marshal(death)
	sch = m.Model.Paths.PathItems.GetOrZero("/burgers/createBurger").Post.RequestBody.Content.GetOrZero("application/json").Schema

	// validate!
	valid, errors = v.ValidateSchemaBytes(sch.Schema(), bodyBytes)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestValidateSchema_ReffyComplex_Invalid(t *testing.T) {
	spec := `openapi: 3.1.0
components:
  schemas:
    Death:
      type: object
      required: [cakeOrDeath]
      properties:
        cakeOrDeath:
          type: string
          enum: [death]
    Cake:
      type: object
      required: [cakeOrDeath]
      properties:
        cakeOrDeath:
          type: string
          enum: [cake please]
    Four:
      type: object
      oneOf:
        - $ref: '#/components/schemas/Cake'
        - $ref: '#/components/schemas/Death'
    Three:
      type: object
      properties:
        name:
          type: string
        four:
          $ref: '#/components/schemas/Four'
    Two:
      type: object
      properties:
        name:
          type: string
        three:
          $ref: '#/components/schemas/Three'
    One:
      type: object
      properties:
        name:
          type: string
        two:
          $ref: '#/components/schemas/Two'
paths:
  /burgers/createBurger:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/One'`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	cakePlease := map[string]interface{}{
		"two": map[string]interface{}{
			"three": map[string]interface{}{
				"four": map[string]interface{}{
					"cakeOrDeath": "no more cake? so the choice is 'or death?'",
				},
			},
		},
	}

	death := map[string]interface{}{
		"two": map[string]interface{}{
			"three": map[string]interface{}{
				"four": map[string]interface{}{
					"cakeOrDeath": "i'll have the chicken",
				},
			},
		},
	}

	// cake? (https://www.youtube.com/watch?v=PVH0gZO5lq0)
	bodyBytes, _ := json.Marshal(cakePlease)
	sch := m.Model.Paths.PathItems.GetOrZero("/burgers/createBurger").Post.RequestBody.Content.GetOrZero("application/json").Schema

	// create a schema validator
	v := NewSchemaValidator()

	// validate!
	valid, errors := v.ValidateSchemaBytes(sch.Schema(), bodyBytes)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 3)

	valid, errors = v.ValidateSchemaObject(sch.Schema(), cakePlease)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 3)

	// or death!
	bodyBytes, _ = json.Marshal(death)
	sch = m.Model.Paths.PathItems.GetOrZero("/burgers/createBurger").Post.RequestBody.Content.GetOrZero("application/json").Schema

	// validate!
	valid, errors = v.ValidateSchemaBytes(sch.Schema(), bodyBytes)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 3)

	valid, errors = v.ValidateSchemaObject(sch.Schema(), death)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 3)
}

func TestValidateSchema_EmptySchema(t *testing.T) {
	// create a schema validator
	v := NewSchemaValidator()

	// validate!
	valid, errors := v.ValidateSchemaObject(nil, nil)

	assert.False(t, valid)
	assert.Len(t, errors, 0)
}

func TestValidateSchema_SimpleInvalid_Multiple(t *testing.T) {
	spec := `openapi: 3.1.0
paths:
  /burgers/createBurger:
    post:
      requestBody:
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

	// create a schema validator
	v := NewSchemaValidator()

	sch := m.Model.Paths.PathItems.GetOrZero("/burgers/createBurger").Post.RequestBody.Content.GetOrZero("application/json").Schema

	// validate!
	valid, errors := v.ValidateSchemaBytes(sch.Schema(), bodyBytes)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 2)
}

func TestValidateSchema_BadJSON(t *testing.T) {
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

	m, _ := doc.BuildV3Model()

	bodyBytes := []byte("{\"bad\": \"json\",}")
	sch := m.Model.Paths.PathItems.GetOrZero("/burgers/createBurger").Post.RequestBody.Content.GetOrZero("application/json").Schema

	// create a schema validator
	v := NewSchemaValidator()

	// validate!
	valid, errors := v.ValidateSchemaString(sch.Schema(), string(bodyBytes))

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "schema does not pass validation", errors[0].Message)
	assert.Equal(t, "invalid character '}' looking for beginning of object key string", errors[0].SchemaValidationErrors[0].Reason)
}

//// https://github.com/pb33f/libopenapi-validator/issues/26
//func TestValidateSchema_v3_0_BooleanExclusiveMinimum(t *testing.T) {
//
//	spec := `openapi: 3.0.0
//paths:
//  /burgers/createBurger:
//    post:
//      requestBody:
//        content:
//          application/json:
//            schema:
//              type: object
//              properties:
//                amount:
//                  type: number
//                  minimum: 0
//                  exclusiveMinimum: true`
//
//	doc, _ := libopenapi.NewDocument([]byte(spec))
//
//	m, _ := doc.BuildV3Model()
//
//	body := map[string]interface{}{"amount": 3}
//
//	bodyBytes, _ := json.Marshal(body)
//	sch := m.Model.Paths.PathItems.GetOrZero("/burgers/createBurger"].Post.RequestBody.Content.GetOrZero("application/json"].Schema
//
//	// create a schema validator
//	v := NewSchemaValidator()
//
//	// validate!
//	valid, errors := v.ValidateSchemaString(sch.Schema(), string(bodyBytes))
//
//	assert.True(t, valid)
//	assert.Empty(t, errors)
//
//}

// https://github.com/pb33f/libopenapi-validator/issues/26
func TestValidateSchema_v3_0_NumericExclusiveMinimum(t *testing.T) {
	spec := `openapi: 3.0.0
paths:
  /burgers/createBurger:
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                amount:
                  type: number
                  exclusiveMinimum: 0`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	body := map[string]interface{}{"amount": 3}

	bodyBytes, _ := json.Marshal(body)
	sch := m.Model.Paths.PathItems.GetOrZero("/burgers/createBurger").Post.RequestBody.Content.GetOrZero("application/json").Schema

	// create a schema validator
	v := NewSchemaValidator()

	// validate!
	valid, errors := v.ValidateSchemaString(sch.Schema(), string(bodyBytes))

	assert.False(t, valid)
	assert.NotEmpty(t, errors)
}

// https://github.com/pb33f/libopenapi-validator/issues/26
func TestValidateSchema_v3_1_NumericExclusiveMinimum(t *testing.T) {
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
                amount:
                  type: number
                  exclusiveMinimum: 0`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	body := map[string]interface{}{"amount": 3}

	bodyBytes, _ := json.Marshal(body)
	sch := m.Model.Paths.PathItems.GetOrZero("/burgers/createBurger").Post.RequestBody.Content.GetOrZero("application/json").Schema

	// create a schema validator
	v := NewSchemaValidator()

	// validate!
	valid, errors := v.ValidateSchemaString(sch.Schema(), string(bodyBytes))

	assert.True(t, valid)
	assert.Empty(t, errors)
}

//func TestValidateSchema_NullableEnum(t *testing.T) {
//	spec := `openapi: 3.0.0
//paths:
//  /burgers/createBurger:
//    post:
//      requestBody:
//        content:
//          application/json:
//            schema:
//              type: object
//              required: [name]
//              properties:
//                name:
//                  type: string
//                  enum: [mcbird, mcbeef, veggie, null]
//                  nullable: true
//                patties:
//                  type: integer
//                vegetarian:
//                  type: boolean`
//
//	doc, _ := libopenapi.NewDocument([]byte(spec))
//
//	m, _ := doc.BuildV3Model()
//
//	body := map[string]interface{}{
//		"name":       nil,
//		"patties":    2,
//		"vegetarian": true,
//	}
//
//	bodyBytes, _ := json.Marshal(body)
//	sch := m.Model.Paths.PathItems.GetOrZero("/burgers/createBurger"].Post.RequestBody.Content.GetOrZero("application/json"].Schema
//
//	// create a schema validator
//	v := NewSchemaValidator()
//
//	// validate!
//	valid, errors := v.ValidateSchemaString(sch.Schema(), string(bodyBytes))
//
//	assert.True(t, valid)
//	assert.Len(t, errors, 0)
//
//}
