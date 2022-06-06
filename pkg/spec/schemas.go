// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	spec "github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
	"github.com/yudai/gojsondiff"
)

const (
	schemasRefPrefix    = "#/components/schemas/"
	maxSchemaToRefDepth = 20
)

// will return a map of SchemaRef and update the operation accordingly.
func updateSchemas(schemas spec.Schemas, op *spec.Operation) (retSchemas spec.Schemas, retOperation *spec.Operation) {
	if op == nil {
		return schemas, op
	}

	for i, response := range op.Responses {
		if response.Value == nil {
			continue
		}
		for content, mediaType := range response.Value.Content {
			
			schemas, mediaType.Schema = schemaToRef(schemas, mediaType.Schema.Value, "", 0)
			op.Responses[i].Value.Content[content] = mediaType
		}
	}

	for i, parameter := range op.Parameters {
		if parameter.Value == nil {
			continue
		}
		for content, mediaType := range parameter.Value.Content {
			schemas, mediaType.Schema = schemaToRef(schemas, mediaType.Schema.Value, "", 0)
			op.Parameters[i].Value.Content[content] = mediaType
		}
	}

	if op.RequestBody.Value != nil {
		for content, mediaType := range op.RequestBody.Value.Content {
			schemas, mediaType.Schema = schemaToRef(schemas, mediaType.Schema.Value, "", 0)
			op.RequestBody.Value.Content[content] = mediaType
		}
	}

	return schemas, op
}

func schemaToRef(schemas spec.Schemas, schema *spec.Schema, schemeNameHint string, depth int) (retSchemes spec.Schemas, schemaRef *spec.SchemaRef) {
	if schema == nil {
		return schemas, nil
	}

	if depth >= maxSchemaToRefDepth {
		log.Warnf("Maximum depth was reached")
		return schemas, spec.NewSchemaRef("", schema)
	}

	if schema.Type == spec.TypeArray {
		if schema.Items == nil {
			// no need to create definition for an empty array
			return schemas, spec.NewSchemaRef("", schema)
		}
		// remove plural from def name hint when it's an array type (if exist)
		schemas, schema.Items = schemaToRef(schemas, schema.Items.Value, strings.TrimSuffix(schemeNameHint, "s"), depth+1)
		return schemas, spec.NewSchemaRef("", schema)
	}

	if schema.Type != spec.TypeObject {
		return schemas, spec.NewSchemaRef("", schema)
	}

	if schema.Properties == nil || len(schema.Properties) == 0 {
		// no need to create ref for an empty object
		return schemas, spec.NewSchemaRef("", schema)
	}

	// go over all properties in the object and convert each one to ref if needed
	var propNames []string
	for propName := range schema.Properties {
		var ref *spec.SchemaRef
		schemas, ref = schemaToRef(schemas, schema.Properties[propName].Value, propName, depth+1)
		if ref != nil {
			schema.Properties[propName] = ref
			propNames = append(propNames, propName)
		}
	}

	// look for schema in schemas with identical schema
	schemeName, exist := findScheme(schemas, schema)
	if !exist {
		// generate new definition
		schemeName = schemeNameHint
		if schemeName == "" {
			schemeName = generateDefNameFromPropNames(propNames)
		}
		if schemas == nil {
			schemas = make(spec.Schemas)
		}
		if existingSchema, ok := schemas[schemeName]; ok {
			log.Debugf("Security scheme name exist with different schema. existingSchema=%+v, schema=%+v", existingSchema, schema)
			schemeName = getUniqueSchemeName(schemas, schemeName)
		}
		schemas[schemeName] = spec.NewSchemaRef("", schema)
	}

	return schemas, spec.NewSchemaRef(schemasRefPrefix+schemeName, nil)
}

func generateDefNameFromPropNames(propNames []string) string {
	// generate name based on properties names when 'defNameHint' is missing
	// sort the slice to get more stable test results
	sort.Strings(propNames)
	return strings.Join(propNames, "_")
}

func getUniqueSchemeName(schemes spec.Schemas, name string) string {
	counter := 0
	for {
		suggestedName := fmt.Sprintf("%s_%d", name, counter)
		if _, ok := schemes[suggestedName]; !ok {
			// found a unique name
			return suggestedName
		}
		// suggestedName already exist - increase counter and look again
		counter++
	}
}

// will look for identical scheme in schemes map.
func findScheme(schemas spec.Schemas, schema *spec.Schema) (schemeName string, exist bool) {
	schemaBytes, _ := json.Marshal(schema)
	differ := gojsondiff.New()
	for name, defSchema := range schemas {
		defSchemaBytes, _ := json.Marshal(defSchema)
		diff, err := differ.Compare(defSchemaBytes, schemaBytes)
		if err != nil {
			log.Errorf("Failed to compare schemas: %v", err)
			continue
		}
		if !diff.Modified() {
			log.Debugf("Schema was found in schemas. schema=%+v, def name=%v", schema, name)
			return name, true
		}
	}

	log.Debugf("Schema was not found in schemas. schema=%+v", schema)
	return "", false
}
