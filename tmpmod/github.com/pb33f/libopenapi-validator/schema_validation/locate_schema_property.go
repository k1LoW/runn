// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package schema_validation

import (
	"github.com/pb33f/libopenapi/utils"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

// LocateSchemaPropertyNodeByJSONPath will locate a schema property node by a JSONPath. It converts something like
// #/components/schemas/MySchema/properties/MyProperty to something like $.components.schemas.MySchema.properties.MyProperty
func LocateSchemaPropertyNodeByJSONPath(doc *yaml.Node, JSONPath string) *yaml.Node {
	var locatedNode *yaml.Node
	doneChan := make(chan bool)
	locatedNodeChan := make(chan *yaml.Node)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				// can't search path, too crazy.
				doneChan <- true
			}
		}()
		_, path := utils.ConvertComponentIdIntoFriendlyPathSearch(JSONPath)
		if path == "" {
			doneChan <- true
		}
		yamlPath, _ := yamlpath.NewPath(path)
		locatedNodes, _ := yamlPath.Find(doc)
		if len(locatedNodes) > 0 {
			locatedNode = locatedNodes[0]
		}
		locatedNodeChan <- locatedNode
	}()
	select {
	case locatedNode = <-locatedNodeChan:
		return locatedNode
	case <-doneChan:
		return nil
	}
}
