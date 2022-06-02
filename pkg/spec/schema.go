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
	"strconv"
	"strings"

	spec "github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
)

var supportedQueryParamsSerializationStyles = []string{
	spec.SerializationForm, spec.SerializationSpaceDelimited, spec.SerializationPipeDelimited,
}

var supportedHeaderParamsSerializationStyles = []string{
	spec.SerializationSimple,
}

var supportedCookieParamsSerializationStyles = []string{
	spec.SerializationForm,
}

// splitByStyle splits a string by a known style:
//		form: comma separated value (default)
//		spaceDelimited: space separated value
//		pipeDelimited: pipe (|) separated value
// https://swagger.io/docs/specification/serialization/
func splitByStyle(data, style string) []string {
	if data == "" {
		return nil
	}
	var sep string
	switch style {
	case spec.SerializationForm, spec.SerializationSimple:
		sep = ","
	case spec.SerializationSpaceDelimited:
		sep = " "
	case spec.SerializationPipeDelimited:
		sep = "|"
	default:
		log.Warnf("Unsupported serialization style: %v", style)
		return nil
	}
	var result []string
	for _, s := range strings.Split(data, sep) {
		if ts := strings.TrimSpace(s); ts != "" {
			result = append(result, ts)
		}
	}
	return result
}

func getNewArraySchema(value string, paramInType string) (schema *spec.Schema, style string) {
	var supportedSerializationStyles []string

	switch paramInType {
	case spec.ParameterInHeader:
		supportedSerializationStyles = supportedHeaderParamsSerializationStyles
	case spec.ParameterInQuery:
		supportedSerializationStyles = supportedQueryParamsSerializationStyles
	case spec.ParameterInCookie:
		supportedSerializationStyles = supportedCookieParamsSerializationStyles
	default:
		log.Errorf("Unsupported paramInType %v", paramInType)
		return nil, ""
	}

	for _, style = range supportedSerializationStyles {
		byStyle := splitByStyle(value, style)
		// Will create an array only if more than a single element exists
		if len(byStyle) > 1 {
			return getSchemaFromValues(byStyle, false, paramInType), style
		}
	}

	return nil, ""
}

func getSchemaFromValues(values []string, shouldTryArraySchema bool, paramInType string) *spec.Schema {
	valuesLen := len(values)

	if valuesLen == 0 {
		return nil
	}

	if valuesLen == 1 {
		return getSchemaFromValue(values[0], shouldTryArraySchema, paramInType)
	}

	// find the most common schema for the items type
	return spec.NewArraySchema().WithItems(getCommonSchema(values, paramInType))
}

func getSchemaFromValue(value string, shouldTryArraySchema bool, paramInType string) *spec.Schema {
	if isDateFormat(value) {
		return spec.NewStringSchema()
	}

	if shouldTryArraySchema {
		schema, _ := getNewArraySchema(value, paramInType)
		if schema != nil {
			return schema
		}
	}

	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return spec.NewInt64Schema()
	}

	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return spec.NewFloat64Schema()
	}

	// TODO: not sure that `strconv.ParseBool` will do the job, it depends what is considers as boolean string
	// The Go implementation for example uses `strconv.FormatBool(value)` ==> true/false
	// But if we look at swag.ConvertBool - `checked` is evaluated as true so `unchecked` will be false?
	// Also when using `strconv.ParseBool` 1 is considered as true so we must check for int before running it
	if _, err := strconv.ParseBool(value); err == nil {
		return spec.NewBoolSchema()
	}

	return spec.NewStringSchema().WithFormat(getStringFormat(value))
}

func getCommonSchema(values []string, paramInType string) *spec.Schema {
	var schemaType string
	var schema *spec.Schema

	for _, value := range values {
		schema = getSchemaFromValue(value, false, paramInType)
		if schemaType == "" {
			// first value, save schema type
			schemaType = schema.Type
		} else if schemaType != schema.Type {
			// different schema type found, defaults to string schema
			return spec.NewStringSchema()
		}
	}

	// identical schema type found
	return schema
}
