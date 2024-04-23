// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

// Package openapi_schemas contains the OpenAPI 3.0 and 3.1 schemas that are loaded from libopenapi, or our own
// fork of the official OpenAPI repo specifications. Using an MD5 hash, we can compare the local version against
// the remote version and determine if they differ, if they do - load the remote version.
package openapi_schemas

import (
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"io"
	"net/http"
)

var schema30, schema31 string

// LoadSchema3_0 loads the latest OpenAPI 3.0 specification. The latest version is fetched from the OpenAPI repo.
// and if there is no change in the schema, the local version is returned, otherwise the remote version is returned.
func LoadSchema3_0(schema string) string {
	if schema30 != "" {
		return schema30
	}
	remoteSpec := "https://raw.githubusercontent.com/pb33f/openapi-specification/main/schemas/v3.0/schema.json"
	schema30 = extractSchema(remoteSpec, schema)
	return schema30
}

// LoadSchema3_1 loads the latest OpenAPI 3.1 specification. The latest version is fetched from the OpenAPI repo.
// and if there is no change in the schema, the local version is returned, otherwise the remote version is returned.
func LoadSchema3_1(schema string) string {
	if schema31 != "" {
		return schema31
	}
	remoteSpec := "https://raw.githubusercontent.com/pb33f/openapi-specification/main/schemas/v3.1/schema.json"
	schema31 = extractSchema(remoteSpec, schema)
	return schema31
}

func getFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

func extractSchema(url string, local string) string {
	// check the local version against the latest version held in our repo.
	remoteVersion, err := getFile(url)
	if err != nil {
		return local
	}
	remoteHash := md5.Sum(remoteVersion)
	remoteMD5 := hex.EncodeToString(remoteHash[:])

	localHash := md5.Sum([]byte(local))
	localMD5 := hex.EncodeToString(localHash[:])

	if remoteMD5 != localMD5 {
		return string(remoteVersion)
	}
	return local
}
