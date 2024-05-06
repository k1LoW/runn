// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package paths

import (
	"net/http"
	"os"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/assert"
)

func TestNewValidator_BadParam(t *testing.T) {

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/pet/doggy", nil)

	// load a doc
	b, _ := os.ReadFile("../test_specs/petstorev3.json")
	doc, _ := libopenapi.NewDocument(b)

	m, _ := doc.BuildV3Model()

	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
}

func TestNewValidator_GoodParamFloat(t *testing.T) {

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/pet/232.233", nil)

	b, _ := os.ReadFile("../test_specs/petstorev3.json")
	doc, _ := libopenapi.NewDocument(b)
	m, _ := doc.BuildV3Model()

	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
}

func TestNewValidator_GoodParamInt(t *testing.T) {

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/pet/12334", nil)

	b, _ := os.ReadFile("../test_specs/petstorev3.json")
	doc, _ := libopenapi.NewDocument(b)

	m, _ := doc.BuildV3Model()
	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
}

func TestNewValidator_FindSimpleEncodedArrayPath(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId*}/locate:
    patch:
      operationId: locateBurger
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodPatch, "https://things.com/burgers/1,2,3,4,5/locate", nil)

	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
	assert.Equal(t, "locateBurger", pathItem.Patch.OperationId)
}

func TestNewValidator_FindSimpleEncodedObjectPath(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId*}/locate:
    patch:
      operationId: locateBurger
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodPatch, "https://things.com/burgers/bish=bosh,wish=wash/locate", nil)

	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
	assert.Equal(t, "locateBurger", pathItem.Patch.OperationId)
}

func TestNewValidator_FindLabelEncodedArrayPath(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{.burgerId}/locate:
    patch:
      operationId: locateBurger
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	request, _ := http.NewRequest(http.MethodPatch, "https://things.com/burgers/.1.2.3.4.5/locate", nil)

	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
	assert.Equal(t, "locateBurger", pathItem.Patch.OperationId)
}

func TestNewValidator_FindPathPost(t *testing.T) {

	// load a doc
	b, _ := os.ReadFile("../test_specs/petstorev3.json")
	doc, _ := libopenapi.NewDocument(b)

	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/pet/12334", nil)

	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
}

func TestNewValidator_FindPathDelete(t *testing.T) {

	// load a doc
	b, _ := os.ReadFile("../test_specs/petstorev3.json")
	doc, _ := libopenapi.NewDocument(b)

	m, _ := doc.BuildV3Model()
	request, _ := http.NewRequest(http.MethodDelete, "https://things.com/pet/12334", nil)

	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
}

func TestNewValidator_FindPathPatch(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId}:
    patch:
      operationId: locateBurger
`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodPatch, "https://things.com/burgers/12345", nil)

	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
	assert.Equal(t, "locateBurger", pathItem.Patch.OperationId)

}

func TestNewValidator_FindPathOptions(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId}:
    options:
      operationId: locateBurger
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()
	request, _ := http.NewRequest(http.MethodOptions, "https://things.com/burgers/12345", nil)

	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
	assert.Equal(t, "locateBurger", pathItem.Options.OperationId)

}

func TestNewValidator_FindPathTrace(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId}:
    trace:
      operationId: locateBurger
`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodTrace, "https://things.com/burgers/12345", nil)

	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
	assert.Equal(t, "locateBurger", pathItem.Trace.OperationId)

}

func TestNewValidator_FindPathPut(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId}:
    put:
      operationId: locateBurger
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodPut, "https://things.com/burgers/12345", nil)

	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
	assert.Equal(t, "locateBurger", pathItem.Put.OperationId)

}

func TestNewValidator_FindPathHead(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /burgers/{burgerId}:
    head:
      operationId: locateBurger
`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodHead, "https://things.com/burgers/12345", nil)

	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
	assert.Equal(t, "locateBurger", pathItem.Head.OperationId)

}

func TestNewValidator_FindPathWithBaseURLInServer(t *testing.T) {

	spec := `openapi: 3.1.0
servers:
  - url: https://things.com/base1
  - url: https://things.com/base2
  - url: https://things.com/base3/base4/base5/base6/
paths:
  /user:
    post:
      operationId: addUser
`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	// check against base1
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/base1/user", nil)
	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
	assert.Equal(t, "addUser", pathItem.Post.OperationId)

	// check against base2
	request, _ = http.NewRequest(http.MethodPost, "https://things.com/base2/user", nil)
	pathItem, _, _ = FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
	assert.Equal(t, "addUser", pathItem.Post.OperationId)

	// check against a deeper base
	request, _ = http.NewRequest(http.MethodPost, "https://things.com/base3/base4/base5/base6/user", nil)
	pathItem, _, _ = FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
	assert.Equal(t, "addUser", pathItem.Post.OperationId)

}

func TestNewValidator_FindPathWithBaseURLInServer_Args(t *testing.T) {

	spec := `openapi: 3.1.0
servers:
  - url: https://things.com/base3/base4/base5/base6/
paths:
  /user/{userId}/thing/{thingId}:
    post:
      operationId: addUser
`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	// check against a deeper base
	request, _ := http.NewRequest(http.MethodPost, "https://things.com/base3/base4/base5/base6/user/1234/thing/abcd", nil)
	pathItem, _, _ := FindPath(request, &m.Model)
	assert.NotNil(t, pathItem)
	assert.Equal(t, "addUser", pathItem.Post.OperationId)

}

func TestNewValidator_FindPathMissing(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /a/fishy/on/a/dishy:
    head:
      operationId: locateFishy
`

	doc, _ := libopenapi.NewDocument([]byte(spec))

	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodHead, "https://things.com/not/here", nil)

	pathItem, errs, _ := FindPath(request, &m.Model)
	assert.Nil(t, pathItem)
	assert.NotNil(t, errs)
	assert.Equal(t, "HEAD Path '/not/here' not found", errs[0].Message)

}

func TestNewValidator_GetLiteralMatch(t *testing.T) {

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/store/inventory", nil)

	// load a doc
	b, _ := os.ReadFile("../test_specs/petstorev3.json")
	doc, _ := libopenapi.NewDocument(b)

	m, _ := doc.BuildV3Model()

	_, errs, _ := FindPath(request, &m.Model)

	assert.Len(t, errs, 0)
}

func TestNewValidator_PostLiteralMatch(t *testing.T) {

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/user", nil)

	// load a doc
	b, _ := os.ReadFile("../test_specs/petstorev3.json")
	doc, _ := libopenapi.NewDocument(b)

	m, _ := doc.BuildV3Model()

	_, errs, _ := FindPath(request, &m.Model)

	assert.Len(t, errs, 0)
}

func TestNewValidator_PutLiteralMatch(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /pizza/burger:
    put:
      operationId: locateBurger`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodPut, "https://things.com/pizza/burger", nil)

	_, errs, _ := FindPath(request, &m.Model)

	assert.Len(t, errs, 0)
}

func TestNewValidator_PutMatch_Error(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /pizza/{cakes}:
    put:
      operationId: locateBurger
      parameters:
        - name: cakes
          in: path
          required: true
          schema:
            type: string`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/pizza/1234", nil)

	_, errs, _ := FindPath(request, &m.Model)

	assert.Len(t, errs, 1)
	assert.Equal(t, "POST Path '/pizza/1234' not found", errs[0].Message)
}

func TestNewValidator_OptionsMatch_Error(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /pizza/{cakes}:
    options:
      operationId: locateBurger
      parameters:
        - name: cakes
          in: path
          required: true
          schema:
            type: string`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/pizza/1234", nil)

	_, errs, _ := FindPath(request, &m.Model)

	assert.Len(t, errs, 1)
	assert.Equal(t, "POST Path '/pizza/1234' not found", errs[0].Message)
}

func TestNewValidator_PatchLiteralMatch(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /pizza/burger:
    patch:
      operationId: locateBurger`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodPatch, "https://things.com/pizza/burger", nil)

	_, errs, _ := FindPath(request, &m.Model)

	assert.Len(t, errs, 0)
}

func TestNewValidator_PatchMatch_Error(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /pizza/{cakes}:
    patch:
      operationId: locateBurger
      parameters:
        - name: cakes
          in: path
          required: true
          schema:
            type: string`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/pizza/1234", nil)

	_, errs, _ := FindPath(request, &m.Model)

	assert.Len(t, errs, 1)
	assert.Equal(t, "POST Path '/pizza/1234' not found", errs[0].Message)
}

func TestNewValidator_DeleteLiteralMatch(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /pizza/burger:
    delete:
      operationId: locateBurger`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodDelete, "https://things.com/pizza/burger", nil)

	_, errs, _ := FindPath(request, &m.Model)

	assert.Len(t, errs, 0)
}

func TestNewValidator_OptionsLiteralMatch(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /pizza/burger:
    options:
      operationId: locateBurger`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodOptions, "https://things.com/pizza/burger", nil)

	_, errs, _ := FindPath(request, &m.Model)

	assert.Len(t, errs, 0)
}

func TestNewValidator_HeadLiteralMatch(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /pizza/burger:
    head:
      operationId: locateBurger`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodHead, "https://things.com/pizza/burger", nil)

	_, errs, _ := FindPath(request, &m.Model)

	assert.Len(t, errs, 0)
}

func TestNewValidator_TraceLiteralMatch(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /pizza/burger:
    trace:
      operationId: locateBurger`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodTrace, "https://things.com/pizza/burger", nil)

	_, errs, _ := FindPath(request, &m.Model)

	assert.Len(t, errs, 0)
}

func TestNewValidator_TraceMatch_Error(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /pizza/{cakes}:
    trace:
      operationId: locateBurger
      parameters:
        - name: cakes
          in: path
          required: true
          schema:
            type: string`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/pizza/1234", nil)

	_, errs, _ := FindPath(request, &m.Model)

	assert.Len(t, errs, 1)
	assert.Equal(t, "POST Path '/pizza/1234' not found", errs[0].Message)
}

func TestNewValidator_DeleteMatch_Error(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /pizza/{cakes}:
    delete:
      operationId: locateBurger
      parameters:
        - name: cakes
          in: path
          required: true
          schema:
            type: string`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/pizza/1234", nil)

	_, errs, _ := FindPath(request, &m.Model)

	assert.Len(t, errs, 1)
	assert.Equal(t, "POST Path '/pizza/1234' not found", errs[0].Message)
}

func TestNewValidator_PostMatch_Error(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /pizza/{cakes}:
    post:
      operationId: locateBurger
      parameters:
        - name: cakes
          in: path
          required: true
          schema:
            type: string`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodPut, "https://things.com/pizza/1234", nil)

	_, errs, _ := FindPath(request, &m.Model)

	assert.Len(t, errs, 1)
	assert.Equal(t, "PUT Path '/pizza/1234' not found", errs[0].Message)
}

func TestNewValidator_FindPathWithFragment(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /hashy#one:
    post:
      operationId: one
  /hashy#two:
    post:
      operationId: two
`

	doc, _ := libopenapi.NewDocument([]byte(spec))
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodPost, "https://things.com/hashy#one", nil)

	pathItem, errs, _ := FindPath(request, &m.Model)
	assert.Len(t, errs, 0)
	assert.NotNil(t, pathItem)
	assert.Equal(t, "one", pathItem.Post.OperationId)

	request, _ = http.NewRequest(http.MethodPost, "https://things.com/hashy#two", nil)
	pathItem, errs, _ = FindPath(request, &m.Model)
	assert.Len(t, errs, 0)
	assert.NotNil(t, pathItem)
	assert.Equal(t, "two", pathItem.Post.OperationId)

}

func TestNewValidator_FindPathMissingWithBaseURLInServer(t *testing.T) {

	spec := `openapi: 3.1.0
servers:
  - url: 'https://things.com/'
paths:
  /dishy:
    get:
      operationId: one
`

	doc, err := libopenapi.NewDocument([]byte(spec))
	if err != nil {
		t.Fatal(err)
	}
	m, _ := doc.BuildV3Model()

	request, _ := http.NewRequest(http.MethodGet, "https://things.com/not_here", nil)

	_, errs, _ := FindPath(request, &m.Model)
	assert.Len(t, errs, 1)
	assert.Equal(t, "GET Path '/not_here' not found", errs[0].Message)

}

func TestGetBasePaths(t *testing.T) {
	spec := `openapi: 3.1.0
servers:
  - url: 'https://things.com/'
  - url: 'https://things.com/some/path'
  - url: 'https://things.com/more//paths//please'
  - url: 'https://{invalid}.com/'
  - url: 'https://{invalid}.com/some/path'
  - url: 'https://{invalid}.com/more//paths//please'
  - url: 'https://{invalid}.com//even//more//paths//please'
paths:
  /dishy:
    get:
      operationId: one
`

	doc, err := libopenapi.NewDocument([]byte(spec))
	if err != nil {
		t.Fatal(err)
	}
	m, _ := doc.BuildV3Model()

	basePaths := getBasePaths(&m.Model)

	expectedPaths := []string{
		"/",
		"/some/path",
		"/more//paths//please",
		"/",
		"/some/path",
		"/more//paths//please",
		"/even//more//paths//please",
	}

	assert.Equal(t, expectedPaths, basePaths)

}
