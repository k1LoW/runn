// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package schema_validation

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocateSchemaPropertyNodeByJSONPath_BadNode(t *testing.T) {

	assert.Nil(t, LocateSchemaPropertyNodeByJSONPath(nil, ""))

}
