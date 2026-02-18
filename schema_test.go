package runn

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	goyaml "github.com/goccy/go-yaml"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaPath = "runbook.schema.yaml"

// schemaID must match the $id in the schema file.
const schemaID = "https://raw.githubusercontent.com/k1LoW/runn/main/runbook.schema.yaml"

func loadSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	doc := loadSchemaAsMap(t)
	c := jsonschema.NewCompiler()
	if err := c.AddResource(schemaID, doc); err != nil {
		t.Fatalf("failed to add schema resource: %v", err)
	}
	sch, err := c.Compile(schemaID)
	if err != nil {
		t.Fatalf("failed to compile schema: %v", err)
	}
	return sch
}

// loadSchemaAsMap reads the YAML schema file and returns it as map[string]any via JSON round-trip.
func loadSchemaAsMap(t *testing.T) map[string]any {
	t.Helper()
	b, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("failed to read schema: %v", err)
	}
	v := yamlToJSON(t, b)
	m, ok := v.(map[string]any)
	if !ok {
		t.Fatal("schema is not an object")
	}
	return m
}

// yamlToJSON converts YAML bytes to a JSON-compatible any value via goccy/go-yaml + encoding/json round-trip.
func yamlToJSON(t *testing.T, b []byte) any {
	t.Helper()
	var v any
	if err := goyaml.Unmarshal(b, &v); err != nil {
		t.Fatalf("failed to unmarshal YAML: %v", err)
	}
	// Round-trip through JSON to normalize types (e.g. map[any]any -> map[string]any)
	jb, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal to JSON: %v", err)
	}
	var jv any
	if err := json.Unmarshal(jb, &jv); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
	return jv
}

func TestSchemaValidatesAllFixtures(t *testing.T) {
	sch := loadSchema(t)

	files, err := filepath.Glob("testdata/book/*.yml")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no fixture files found")
	}

	for _, f := range files {
		name := filepath.Base(f)
		t.Run(name, func(t *testing.T) {
			b, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("failed to read %s: %v", f, err)
			}
			v := yamlToJSON(t, b)
			if err := sch.Validate(v); err != nil {
				t.Errorf("schema validation failed for %s: %v", name, err)
			}
		})
	}
}

func TestSchemaCoversReservedKeys(t *testing.T) {
	schema := loadSchemaAsMap(t)

	// Extract step property keys from schema
	defs, ok := schema["$defs"].(map[string]any)
	if !ok {
		t.Fatal("$defs not found in schema")
	}
	stepDef, ok := defs["step"].(map[string]any)
	if !ok {
		t.Fatal("step definition not found in schema")
	}
	stepProps, ok := stepDef["properties"].(map[string]any)
	if !ok {
		t.Fatal("step properties not found in schema")
	}

	// Reserved runner keys
	runnerKeys := []string{
		includeRunnerKey, // "include"
		testRunnerKey,    // "test"
		dumpRunnerKey,    // "dump"
		execRunnerKey,    // "exec"
		bindRunnerKey,    // "bind"
		runnerRunnerKey,  // "runner"
	}
	for _, k := range runnerKeys {
		if _, ok := stepProps[k]; !ok {
			t.Errorf("reserved runner key %q is not in schema step properties", k)
		}
	}

	// Reserved section keys
	sectionKeys := []string{
		ifSectionKey,    // "if"
		descSectionKey,  // "desc"
		loopSectionKey,  // "loop"
		deferSectionKey, // "defer"
		forceSectionKey, // "force"
	}
	for _, k := range sectionKeys {
		if _, ok := stepProps[k]; !ok {
			t.Errorf("reserved section key %q is not in schema step properties", k)
		}
	}
}

func TestSchemaCoversRunnerConfigFields(t *testing.T) {
	schema := loadSchemaAsMap(t)
	defs, ok := schema["$defs"].(map[string]any)
	if !ok {
		t.Fatal("$defs not found in schema")
	}

	tests := []struct {
		name     string
		config   any
		defName  string
		skipTags map[string]bool // Fields to skip (unexported, yaml:"-", etc.)
	}{
		{
			name:    "httpRunnerConfig",
			config:  httpRunnerConfig{},
			defName: "httpRunnerConfig",
			skipTags: map[string]bool{
				"openAPI3Doc": true, // unexported, no yaml tag
			},
		},
		{
			name:    "grpcRunnerConfig",
			config:  grpcRunnerConfig{},
			defName: "grpcRunnerConfig",
			skipTags: map[string]bool{
				"cacert": true, // unexported []byte field, not YAML-serializable
				"cert":   true, // unexported []byte field, not YAML-serializable
				"key":    true, // unexported []byte field, not YAML-serializable
			},
		},
		{
			name:    "dbRunnerConfig",
			config:  dbRunnerConfig{},
			defName: "dbRunnerConfig",
		},
		{
			name:    "sshRunnerConfig",
			config:  sshRunnerConfig{},
			defName: "sshRunnerConfig",
		},
		{
			name:    "cdpRunnerConfig",
			config:  cdpRunnerConfig{},
			defName: "cdpRunnerConfig",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, ok := defs[tt.defName].(map[string]any)
			if !ok {
				t.Fatalf("definition %q not found in schema", tt.defName)
			}
			props, ok := def["properties"].(map[string]any)
			if !ok {
				t.Fatalf("properties not found in definition %q", tt.defName)
			}

			rt := reflect.TypeOf(tt.config)
			for i := 0; i < rt.NumField(); i++ {
				field := rt.Field(i)
				if !field.IsExported() {
					continue
				}
				yamlTag := field.Tag.Get("yaml")
				if yamlTag == "-" {
					continue
				}
				yamlName := strings.Split(yamlTag, ",")[0]
				if yamlName == "" {
					yamlName = strings.ToLower(field.Name)
				}
				if tt.skipTags != nil && tt.skipTags[yamlName] {
					continue
				}
				if _, ok := props[yamlName]; !ok {
					t.Errorf("field %q (yaml: %q) of %s is not in schema definition %q", field.Name, yamlName, tt.name, tt.defName)
				}
			}
		})
	}
}

func TestSchemaRejectsInvalidRunbooks(t *testing.T) {
	sch := loadSchema(t)

	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "invalid labels with space",
			yaml: `
labels:
  - "has space"
steps:
  - test: true
`,
		},
		{
			name: "invalid labels with exclamation",
			yaml: `
labels:
  - "has!"
steps:
  - test: true
`,
		},
		{
			name: "invalid loop structure",
			yaml: `
steps:
  - loop:
      count: 3
      invalidField: true
    test: true
`,
		},
		{
			name: "steps empty array",
			yaml: `
steps: []
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := yamlToJSON(t, []byte(tt.yaml))
			if err := sch.Validate(v); err == nil {
				t.Errorf("expected schema validation to fail for %q, but it passed", tt.name)
			}
		})
	}
}
