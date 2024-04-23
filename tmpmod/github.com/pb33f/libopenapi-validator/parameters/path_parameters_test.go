// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package parameters

import (
	"net/http"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/k1LoW/runn/tmpmod/github.com/pb33f/libopenapi-validator/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator_SimpleArrayEncodedPath(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerIds*}/locate:
    parameters:
      - name: burgerIds
        in: path
        schema:
          type: array
          items:
            type: integer
    patch:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodPatch, "https://things.com/burgers/1,2,3,4,5/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_SimpleArrayEncodedPath_InvalidNumber(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerIds*}/locate:
    parameters:
      - name: burgerIds
        in: path
        schema:
          type: array
          items:
            type: integer
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/1,pizza,3,4,false/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 2)
	assert.Equal(t, "Path array parameter 'burgerIds' is not a valid number", errors[0].Message)
	assert.Equal(t, request.Method, errors[0].RequestMethod)
	assert.Equal(t, request.URL.Path, errors[0].RequestPath)
	assert.Equal(t, "/burgers/{burgerIds*}/locate", errors[0].SpecPath)
}

func TestNewValidator_SimpleArrayEncodedPath_InvalidBool(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerIds*}/locate:
    parameters:
      - name: burgerIds
        in: path
        schema:
          type: array
          items:
            type: boolean
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/1,true,0,frogs,false/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 3)
	assert.Equal(t, "Path array parameter 'burgerIds' is not a valid boolean", errors[0].Message)
}

func TestNewValidator_SimpleObjectEncodedPath(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burger}/locate:
    parameters:
      - name: burger
        in: path
        schema:
          type: object
          properties:
            id:
               type: integer
            vegetarian:
               type: boolean
    get:
      operationId: locateBurger`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/id,1234,vegetarian,true/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_SimpleObjectEncodedPath_Invalid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burger}/locate:
    parameters:
      - name: burger
        in: path
        schema:
          type: object
          properties:
            id:
               type: integer
            vegetarian:
               type: boolean
    get:
      operationId: locateBurger`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/id,hello,vegetarian,there/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 2)
}

func TestNewValidator_SimpleObjectEncodedPath_Exploded(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burger}/locate:
    parameters:
      - name: burger
        in: path
        explode: true
        schema:
          type: object
          properties:
            id:
               type: integer
            vegetarian:
               type: boolean
    get:
      operationId: locateBurger`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/id=1234,vegetarian=true/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_SimpleObjectEncodedPath_ExplodedInvalid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burger}/locate:
    parameters:
      - name: burger
        in: path
        explode: true
        schema:
          type: object
          properties:
            id:
               type: integer
            vegetarian:
               type: boolean
    get:
      operationId: locateBurger`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/id=toast,vegetarian=chicken/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 2)
}

func TestNewValidator_ObjectEncodedPath(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burger}/locate:
    parameters:
      - name: burger
        in: path
        schema:
          type: object
          properties:
            id:
               type: integer
            vegetarian:
               type: boolean
    get:
      operationId: locateBurger`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/id,1234,vegetarian,true/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_SimpleEncodedPath_InvalidInteger(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        schema:
          type: integer
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/hello/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' is not a valid number", errors[0].Message)
}

func TestNewValidator_SimpleEncodedPath_IntegerViolation(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        schema:
          type: integer
          minimum: 10
    get:
      operationId: locateBurgers`

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)
	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/1/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' failed to validate", errors[0].Message)
	assert.Len(t, errors[0].SchemaValidationErrors, 1)
	assert.Equal(t, "Reason: must be >= 10 but found 1, Location: /minimum", errors[0].SchemaValidationErrors[0].Error())
}

func TestNewValidator_SimpleEncodedPath_Integer(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        schema:
          type: integer
          minimum: 10
    get:
      operationId: locateBurgers`

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)
	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/14/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Nil(t, errors)
}

func TestNewValidator_SimpleEncodedPath_InvalidBoolean(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        schema:
          type: boolean
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/hello/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' is not a valid boolean", errors[0].Message)
}

func TestNewValidator_LabelEncodedPath_InvalidInteger(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{.burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: label
        schema:
          type: integer
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/.hello/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' is not a valid number", errors[0].Message)
}

func TestNewValidator_LabelEncodedPath_IntegerViolation(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{.burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: label
        schema:
          type: integer
          minimum: 10
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/.3/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' failed to validate", errors[0].Message)
	assert.Len(t, errors[0].SchemaValidationErrors, 1)
	assert.Equal(t, "Reason: must be >= 10 but found 3, Location: /minimum", errors[0].SchemaValidationErrors[0].Error())
}

func TestNewValidator_LabelEncodedPath_InvalidBoolean(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{.burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: label
        schema:
          type: boolean
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/.hello/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' is not a valid boolean", errors[0].Message)
}

func TestNewValidator_LabelEncodedPath_ValidArray_Number(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{.burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: label
        schema:
          type: array
          items:
            type: number
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/.3,4,5,6/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_LabelEncodedPath_ValidArray_Number_Exploded(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{.burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: label
        explode: true
        schema:
          type: array
          items:
            type: number
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/.3.4.5.6/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_LabelEncodedPath_InvalidArray_Number_Exploded(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{.burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: label
        explode: true
        schema:
          type: array
          items:
            type: number
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/.3.Not a number.5.6/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path array parameter 'burgerId' is not a valid number", errors[0].Message)
}

func TestNewValidator_LabelEncodedPath_InvalidArray_Number(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{.burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: label
        schema:
          type: array
          items:
            type: number
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/.3,4,Not a number,6/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path array parameter 'burgerId' is not a valid number", errors[0].Message)
}

func TestNewValidator_LabelEncodedPath_InvalidObject(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{.burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: label
        schema:
          type: object
          properties:
            id:
              type: integer
            vegetarian:
              type: boolean
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/.id,hello,vegetarian,why/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 2)
}

func TestNewValidator_LabelEncodedPath_InvalidObject_Exploded(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{.burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: label
        explode: true
        schema:
          type: object
          properties:
            id:
              type: integer
            vegetarian:
              type: boolean
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/.id=hello.vegetarian=why/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Len(t, errors[0].SchemaValidationErrors, 2)
}

func TestNewValidator_LabelEncodedPath_ValidMultiParam(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{.burgerId}/locate/{.query}:
    parameters:
      - name: query
        in: path
        style: label
        schema:
          type: string
      - name: burgerId
        in: path
        style: label
        explode: true
        schema:
          type: object
          properties:
            id:
              type: integer
            vegetarian:
              type: boolean
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/.id=1234.vegetarian=true/locate/bigMac", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_LabelEncodedPath_InvalidMultiParam(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{.burgerId}/locate/{.query}:
    parameters:
      - name: query
        in: path
        style: label
        schema:
          type: integer
      - name: burgerId
        in: path
        style: label
        explode: true
        schema:
          type: object
          properties:
            id:
              type: integer
            vegetarian:
              type: boolean
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/.id=1234.vegetarian=true/locate/bigMac", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
}

func TestNewValidator_MatrixEncodedPath_ValidPrimitiveNumber(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: matrix
        schema:
          type: integer
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;burgerId=5/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_MatrixEncodedPath_InvalidPrimitiveNumber(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: matrix
        schema:
          type: integer
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;burgerId=I am not a number/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' is not a valid number", errors[0].Message)
}

func TestNewValidator_MatrixEncodedPath_PrimitiveNumberViolation(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: matrix
        schema:
          type: integer
          minimum: 5
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;burgerId=3/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' failed to validate", errors[0].Message)
	assert.Len(t, errors[0].SchemaValidationErrors, 1)
	assert.Equal(t, "Reason: must be >= 5 but found 3, Location: /minimum", errors[0].SchemaValidationErrors[0].Error())
}

func TestNewValidator_MatrixEncodedPath_ValidPrimitiveBoolean(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: matrix
        schema:
          type: boolean
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;burgerId=false/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_MatrixEncodedPath_InvalidPrimitiveBoolean(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: matrix
        schema:
          type: boolean
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;burgerId=I am also not a bool/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' is not a valid boolean", errors[0].Message)

}

func TestNewValidator_MatrixEncodedPath_ValidObject(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burger}/locate:
    parameters:
      - name: burger
        in: path
        style: matrix
        schema:
          type: object
          properties:
            id:
              type: integer
            vegetarian:
              type: boolean
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;burger=id,1234,vegetarian,false/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)

}

func TestNewValidator_MatrixEncodedPath_InvalidObject(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burger}/locate:
    parameters:
      - name: burger
        in: path
        style: matrix
        schema:
          type: object
          properties:
            id:
              type: integer
            vegetarian:
              type: boolean
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;burger=id,1234,vegetarian,I am not a bool/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "expected boolean, but got string", errors[0].SchemaValidationErrors[0].Reason)
}

func TestNewValidator_MatrixEncodedPath_ValidObject_Exploded(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burger*}/locate:
    parameters:
      - name: burger
        in: path
        style: matrix
        explode: true
        schema:
          type: object
          properties:
            id:
              type: integer
            vegetarian:
              type: boolean
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;id=1234;vegetarian=false/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)

}

func TestNewValidator_MatrixEncodedPath_InvalidObject_Exploded(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burger*}/locate:
    parameters:
      - name: burger
        in: path
        style: matrix
        explode: true
        schema:
          type: object
          properties:
            id:
              type: integer
            vegetarian:
              type: boolean
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;id=1234;vegetarian=I am not a boolean/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "expected boolean, but got string", errors[0].SchemaValidationErrors[0].Reason)
}

func TestNewValidator_MatrixEncodedPath_ValidArray(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burger*}/locate:
    parameters:
      - name: burger
        in: path
        style: matrix
        schema:
          type: array
          items:
            type: integer
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;burger=1,2,3,4,5/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)

}

func TestNewValidator_MatrixEncodedPath_InvalidArray(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burger*}/locate:
    parameters:
      - name: burger
        in: path
        style: matrix
        schema:
          type: array
          items:
            type: integer
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;burger=1,2,not a number,4,false/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 2)
}

func TestNewValidator_MatrixEncodedPath_ValidArray_Exploded(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burger*}/locate:
    parameters:
      - name: burger
        in: path
        style: matrix
        explode: true
        schema:
          type: array
          items:
            type: integer
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;burger=1;burger=2;burger=3/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_MatrixEncodedPath_InvalidArray_Exploded(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burger*}/locate:
    parameters:
      - name: burger
        in: path
        style: matrix
        explode: true
        schema:
          type: array
          items:
            type: integer
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;burger=1;burger=I am not an int;burger=3/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path array parameter 'burger' is not a valid number", errors[0].Message)
}

func TestNewValidator_PathParams_PathNotFound(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burger*}/locate:
    parameters:
      - name: burger
        in: path
        style: matrix
        explode: true
        schema:
          type: array
          items:
            type: integer
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/I do not exist", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
}

func TestNewValidator_PathParamStringEnumValid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        schema:
          type: string
          enum: [bigMac, whopper, mcCrispy]
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/bigMac/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)

}

func TestNewValidator_PathParamStringEnumInvalid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        schema:
          type: string
          enum: [bigMac, whopper, mcCrispy]
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/hello/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' does not match allowed values", errors[0].Message)
	assert.Equal(t, "Instead of 'hello', use one of the allowed values: 'bigMac, whopper, mcCrispy'", errors[0].HowToFix)

}

func TestNewValidator_PathParamStringViolation(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        schema:
          type: string
          minLength: 4
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/big/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' failed to validate", errors[0].Message)
	assert.Len(t, errors[0].SchemaValidationErrors, 1)
	assert.Equal(t, "Reason: length must be >= 4, but got 3, Location: /minLength", errors[0].SchemaValidationErrors[0].Error())
}

func TestNewValidator_PathParamIntegerEnumValid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        schema:
          type: number
          enum: [1,2,99,100]
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/2/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_PathParamIntegerEnumInvalid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        schema:
          type: number
          enum: [1,2,99,100]
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/3284/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' does not match allowed values", errors[0].Message)
}

func TestNewValidator_PathLabelEumValid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{.burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: label
        schema:
          type: number
          enum: [1,2,99,100]
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/.2/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}

func TestNewValidator_PathLabelEumInvalid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{.burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: label
        schema:
          type: number
          enum: [1,2,99,100]
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/.22334/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' does not match allowed values", errors[0].Message)
	assert.Equal(t, "Instead of '22334', use one of the allowed values: '1, 2, 99, 100'", errors[0].HowToFix)
}

func TestNewValidator_PathMatrixEumInvalid(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: matrix
        schema:
          type: number
          enum: [1,2,99,100]
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;burgerId=22334/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' does not match allowed values", errors[0].Message)
	assert.Equal(t, "Instead of '22334', use one of the allowed values: '1, 2, 99, 100'", errors[0].HowToFix)
}

func TestNewValidator_SetPathForPathParam(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{;burgerId}/locate:
    parameters:
      - name: burgerId
        in: path
        style: matrix
        schema:
          type: number
          enum: [1,2,99,100]
    get:
      operationId: locateBurgers`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/burgers/;burgerId=22334/locate", nil)

	// preset the path
	path, _, pv := paths.FindPath(request, &m.Model)
	v.SetPathItem(path, pv)

	valid, errors := v.ValidatePathParams(request)

	assert.False(t, valid)
	assert.Len(t, errors, 1)
	assert.Equal(t, "Path parameter 'burgerId' does not match allowed values", errors[0].Message)
	assert.Equal(t, "Instead of '22334', use one of the allowed values: '1, 2, 99, 100'", errors[0].HowToFix)
}

func TestNewValidator_ServerPathPrefixInRequestPath(t *testing.T) {

	spec := `openapi: 3.1.0
servers:
  - url: https://api.pb33f.io/lorem/ipsum
    description: Live production endpoint for general use.
paths:
  /burgers/{burger}/locate:
    parameters:
      - name: burger
        in: path
        schema:
          type: string
          format: uuid
    get:
      operationId: locateBurger`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	v := NewParameterValidator(&m.Model)

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/lorem/ipsum/burgers/d6d8d513-686c-466f-9f5a-1c051b6b4f3f/locate", nil)
	valid, errors := v.ValidatePathParams(request)

	assert.True(t, valid)
	assert.Len(t, errors, 0)
}
