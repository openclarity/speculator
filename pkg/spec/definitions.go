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

	"github.com/go-openapi/spec"
	log "github.com/sirupsen/logrus"
	"github.com/yudai/gojsondiff"
)

const (
	definitionsRefPrefix = "#/definitions/"
	maxSchemaToRefDepth  = 20
)

// will return a map of definitions and update the operation accordingly.
func updateDefinitions(definitions map[string]spec.Schema, op *spec.Operation) (retDefinitions map[string]spec.Schema, retOperation *spec.Operation) {
	if op == nil {
		return definitions, op
	}

	if op.Responses != nil {
		for i, response := range op.Responses.StatusCodeResponses {
			definitions, response.Schema = schemaToRef(definitions, response.Schema, "", 0)
			op.Responses.StatusCodeResponses[i] = response
		}
	}

	for i, parameter := range op.Parameters {
		definitions, parameter.Schema = schemaToRef(definitions, parameter.Schema, "", 0)
		op.Parameters[i] = parameter
	}

	return definitions, op
}

func schemaToRef(definitions map[string]spec.Schema, schema *spec.Schema, defNameHint string, depth int) (retDefinitions map[string]spec.Schema, retSchema *spec.Schema) {
	if schema == nil {
		return definitions, schema
	}

	if depth >= maxSchemaToRefDepth {
		log.Warnf("Maximum depth was reached")
		return definitions, schema
	}

	if schema.Type.Contains(schemaTypeArray) {
		if schema.Items == nil {
			// no need to create definition for an empty array
			return definitions, schema
		}
		// remove plural from def name hint when it's an array type (if exist)
		definitions, schema.Items.Schema = schemaToRef(definitions, schema.Items.Schema, strings.TrimSuffix(defNameHint, "s"), depth+1)
		return definitions, schema
	}

	if !schema.Type.Contains(schemaTypeObject) {
		return definitions, schema
	}

	if schema.Properties == nil {
		// no need to create definition for an empty object
		return definitions, schema
	}

	// go over all properties in the object and convert each one to ref if needed
	var propNames []string
	for propName := range schema.Properties {
		var newSchema *spec.Schema
		propSchema := schema.Properties[propName]
		definitions, newSchema = schemaToRef(definitions, &propSchema, propName, depth+1)
		schema.Properties[propName] = *newSchema
		propNames = append(propNames, propName)
	}

	// look for definition with identical schema
	defName, exist := findDefinition(definitions, schema)
	if !exist {
		// generate new definition
		defName = defNameHint
		if defName == "" {
			defName = generateDefNameFromPropNames(propNames)
		}
		if definitions == nil {
			definitions = make(map[string]spec.Schema)
		}
		if existingSchema, ok := definitions[defName]; ok {
			log.Debugf("Definition name exist with different schema. existingSchema=%+v, schema=%+v", existingSchema, schema)
			defName = getUniqueDefName(definitions, defName)
		}
		definitions[defName] = *schema
	}

	retSchema = spec.RefSchema(definitionsRefPrefix + defName)

	return definitions, retSchema
}

func generateDefNameFromPropNames(propNames []string) string {
	// generate name based on properties names when 'defNameHint' is missing
	// sort the slice to get more stable test results
	sort.Strings(propNames)
	return strings.Join(propNames, "_")
}

func getUniqueDefName(definitions map[string]spec.Schema, name string) string {
	counter := 0
	for {
		suggestedName := fmt.Sprintf("%s_%d", name, counter)
		if _, ok := definitions[suggestedName]; !ok {
			// found a unique name
			return suggestedName
		}
		// suggestedName already exist - increase counter and look again
		counter++
	}
}

// will look for identical schema definition in definitions map.
func findDefinition(definitions map[string]spec.Schema, schema *spec.Schema) (defName string, exist bool) {
	schemaBytes, _ := json.Marshal(schema)
	differ := gojsondiff.New()
	for name, defSchema := range definitions {
		defSchemaBytes, _ := json.Marshal(defSchema)
		diff, err := differ.Compare(defSchemaBytes, schemaBytes)
		if err != nil {
			log.Errorf("Failed to compare schemas: %v", err)
			continue
		}
		if !diff.Modified() {
			log.Debugf("Schema was found in definitions. schema=%+v, def name=%v", schema, name)
			return name, true
		}
	}

	log.Debugf("Schema was not found in definitions. schema=%+v", schema)
	return "", false
}
