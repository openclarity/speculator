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

	"github.com/go-openapi/spec"
	log "github.com/sirupsen/logrus"
	"k8s.io/utils/field"

	"github.com/apiclarity/speculator/pkg/utils"
	"github.com/apiclarity/speculator/pkg/utils/slice"
)

var supportedParametersInTypes = []string{parametersInBody, parametersInHeader, parametersInQuery, parametersInForm, parametersInPath}

func mergeOperation(operation, operation2 *spec.Operation) (*spec.Operation, []conflict) {
	if op, shouldReturn := shouldReturnIfNil(operation, operation2); shouldReturn {
		return op.(*spec.Operation), nil
	}

	var paramConflicts, resConflicts []conflict

	ret := spec.NewOperation("")

	ret.Parameters, paramConflicts = mergeParameters(operation.Parameters, operation2.Parameters,
		field.NewPath("parameters"))
	ret.Responses, resConflicts = mergeResponses(operation.Responses, operation2.Responses,
		field.NewPath("responses"))

	ret.Consumes = slice.RemoveStringDuplicates(append(operation.Consumes, operation2.Consumes...))
	ret.Produces = slice.RemoveStringDuplicates(append(operation.Produces, operation2.Produces...))

	ret.Security = mergeOperationSecurity(operation.Security, operation2.Security)

	conflicts := append(paramConflicts, resConflicts...)

	if len(conflicts) > 0 {
		log.Warnf("Found conflicts while merging operation: %v and operation: %v. conflicts: %v", operation, operation2, conflicts)
	}

	return ret, conflicts
}

func mergeOperationSecurity(security, security2 []map[string][]string) []map[string][]string {
	var mergedSecurity []map[string][]string

	ignoreSecurityKeyMap := map[string]bool{}

	for _, securityMap := range security {
		mergedSecurity, ignoreSecurityKeyMap = appendSecurityIfNeeded(securityMap, mergedSecurity, ignoreSecurityKeyMap)
	}
	for _, securityMap := range security2 {
		mergedSecurity, ignoreSecurityKeyMap = appendSecurityIfNeeded(securityMap, mergedSecurity, ignoreSecurityKeyMap)
	}

	return mergedSecurity
}

func appendSecurityIfNeeded(securityMap map[string][]string, mergedSecurity []map[string][]string, ignoreSecurityKeyMap map[string]bool) ([]map[string][]string, map[string]bool) {
	for key, values := range securityMap {
		// ignore if already appended the exact security key
		if ignoreSecurityKeyMap[key] {
			continue
		}
		// https://swagger.io/docs/specification/2-0/authentication/
		// We will treat multiple authentication types as an OR
		// (Security schemes combined via OR are alternatives – any one can be used in the given context)
		mergedSecurity = append(mergedSecurity, map[string][]string{key: values})
		ignoreSecurityKeyMap[key] = true
	}

	return mergedSecurity, ignoreSecurityKeyMap
}

func mergeParameters(parameters, parameters2 []spec.Parameter, path *field.Path) ([]spec.Parameter, []conflict) {
	if p, shouldReturn := shouldReturnIfEmptyParameters(parameters, parameters2); shouldReturn {
		return p, nil
	}

	var retParameters []spec.Parameter
	var retConflicts []conflict

	parametersByIn := getParametersByIn(parameters)
	parameters2ByIn := getParametersByIn(parameters2)
	for _, inType := range supportedParametersInTypes {
		var mergedParameters []spec.Parameter
		var conflicts []conflict

		if inType == inBodyParameterName {
			mergedParameters, conflicts = mergeInBodyParameters(parametersByIn[inType], parameters2ByIn[inType], path)
		} else {
			mergedParameters, conflicts = mergeParametersByInType(parametersByIn[inType], parameters2ByIn[inType], path)
		}
		retParameters = append(retParameters, mergedParameters...)
		retConflicts = append(retConflicts, conflicts...)
	}

	return retParameters, retConflicts
}

func getParametersByIn(parameters []spec.Parameter) map[string][]spec.Parameter {
	ret := make(map[string][]spec.Parameter)

	for i, parameter := range parameters {
		switch parameter.In {
		case parametersInBody, parametersInHeader, parametersInQuery, parametersInForm, parametersInPath:
			ret[parameter.In] = append(ret[parameter.In], parameters[i])
		default:
			log.Warnf("in parameter not supported. %v", parameter.In)
		}
	}

	return ret
}

func mergeParametersByInType(parameters, parameters2 []spec.Parameter, path *field.Path) ([]spec.Parameter, []conflict) {
	if p, shouldReturn := shouldReturnIfEmptyParameters(parameters, parameters2); shouldReturn {
		return p, nil
	}

	var retParameters []spec.Parameter
	var retConflicts []conflict

	parametersMapByName := makeParametersMapByName(parameters)
	parameters2MapByName := makeParametersMapByName(parameters2)

	// go over first parameters list
	// 1. merge mutual parameters
	// 2. add non mutual parameters
	for name, param := range parametersMapByName {
		if param2, ok := parameters2MapByName[name]; ok {
			mergedParameter, conflicts := mergeParameter(param, param2, path.Child(name))
			retConflicts = append(retConflicts, conflicts...)
			retParameters = append(retParameters, mergedParameter)
		} else {
			retParameters = append(retParameters, param)
		}
	}

	// add non mutual parameters from the second list
	for name, param := range parameters2MapByName {
		if _, ok := parametersMapByName[name]; !ok {
			retParameters = append(retParameters, param)
		}
	}

	return retParameters, retConflicts
}

func mergeInBodyParameters(parameters, parameters2 []spec.Parameter, path *field.Path) ([]spec.Parameter, []conflict) {
	if p, shouldReturn := shouldReturnIfEmptyParameters(parameters, parameters2); shouldReturn {
		return p, nil
	}

	// we can only have a single in body param named 'body' (inBodyParameterName)
	mergedSchema, conflicts := mergeSchema(parameters[0].Schema, parameters2[0].Schema,
		path.Child(parameters[0].Name, "schema"))

	return []spec.Parameter{*spec.BodyParam(inBodyParameterName, mergedSchema)}, conflicts
}

func makeParametersMapByName(parameters []spec.Parameter) map[string]spec.Parameter {
	ret := make(map[string]spec.Parameter)

	for i := range parameters {
		ret[parameters[i].Name] = parameters[i]
	}

	return ret
}

func mergeParameter(parameter, parameter2 spec.Parameter, path *field.Path) (spec.Parameter, []conflict) {
	if parameter.Type != parameter2.Type {
		return parameter, []conflict{
			{
				path: path,
				obj1: parameter,
				obj2: parameter2,
				msg:  createConflictMsg(path, parameter.Type, parameter2.Type),
			},
		}
	}

	switch parameter.Type {
	case schemaTypeBoolean, schemaTypeInteger, schemaTypeNumber, schemaTypeString:
		simpleSchema, conflicts := mergeSimpleSchema(parameter.SimpleSchema, parameter2.SimpleSchema, path)
		parameter.SimpleSchema = simpleSchema
		return parameter, conflicts
	case schemaTypeArray:
		items, conflicts := mergeSimpleSchemaItems(parameter.Items, parameter2.Items, path)
		parameter.Items = items
		return parameter, conflicts
	case "":
		// when type is missing it is probably an object - we should try and merge the parameter schema
		schema, conflicts := mergeSchema(parameter.Schema, parameter2.Schema, path.Child("schema"))
		parameter.Schema = schema
		return parameter, conflicts
	default:
		log.Warnf("not supported schema type in parameter: %v", parameter.Type)
	}

	return parameter, nil
}

func mergeSimpleSchemaItems(items, items2 *spec.Items, path *field.Path) (*spec.Items, []conflict) {
	if s, shouldReturn := shouldReturnIfNil(items, items2); shouldReturn {
		return s.(*spec.Items), nil
	}
	simpleSchema, conflicts := mergeSimpleSchema(items.SimpleSchema, items2.SimpleSchema, path.Child("items"))
	items.SimpleSchema = simpleSchema
	return items, conflicts
}

func mergeSimpleSchema(simpleSchema, simpleSchema2 spec.SimpleSchema, path *field.Path) (spec.SimpleSchema, []conflict) {
	if simpleSchema.Type != simpleSchema2.Type {
		return simpleSchema, []conflict{
			{
				path: path,
				obj1: simpleSchema,
				obj2: simpleSchema2,
				msg:  createConflictMsg(path, simpleSchema.Type, simpleSchema2.Type),
			},
		}
	}

	switch simpleSchema.Type {
	case schemaTypeBoolean, schemaTypeInteger, schemaTypeNumber:
		return simpleSchema, nil
	case schemaTypeString:
		if simpleSchema.Format != simpleSchema2.Format {
			simpleSchema.Format = ""
		}
		return simpleSchema, nil
	case schemaTypeArray:
		items, conflicts := mergeSimpleSchemaItems(simpleSchema.Items, simpleSchema2.Items, path)
		simpleSchema.Items = items
		return simpleSchema, conflicts
	default:
		log.Warnf("not supported schema type in simple schema: %v", simpleSchema.Type)
	}

	return simpleSchema, nil
}

func mergeSchema(schema, schema2 *spec.Schema, path *field.Path) (*spec.Schema, []conflict) {
	if s, shouldReturn := shouldReturnIfNil(schema, schema2); shouldReturn {
		return s.(*spec.Schema), nil
	}

	if s, shouldReturn := shouldReturnIfEmptySchemaType(schema, schema2); shouldReturn {
		return s, nil
	}

	if schema.Type[0] != schema2.Type[0] {
		return schema, []conflict{
			{
				path: path,
				obj1: schema,
				obj2: schema2,
				msg:  createConflictMsg(path, schema.Type[0], schema2.Type[0]),
			},
		}
	}

	switch schema.Type[0] {
	case schemaTypeBoolean, schemaTypeInteger, schemaTypeNumber:
		return schema, nil
	case schemaTypeString:
		if schema.Format != schema2.Format {
			schema.Format = ""
		}
		return schema, nil
	case schemaTypeArray:
		items, conflicts := mergeSchemaItems(schema.Items, schema2.Items, path)
		schema.Items = items
		return schema, conflicts
	case schemaTypeObject:
		properties, conflicts := mergeProperties(schema.Properties, schema2.Properties, path.Child("properties"))
		schema.Properties = properties
		return schema, conflicts
	default:
		log.Warnf("not supported schema type in schema: %v", schema.Type[0])
	}

	return schema, nil
}

func mergeSchemaItems(items, items2 *spec.SchemaOrArray, path *field.Path) (*spec.SchemaOrArray, []conflict) {
	if s, shouldReturn := shouldReturnIfNil(items, items2); shouldReturn {
		return s.(*spec.SchemaOrArray), nil
	}

	mergedSchema, conflicts := mergeSchema(items.Schema, items2.Schema, path.Child("items"))
	items.Schema = mergedSchema
	return items, conflicts
}

func mergeProperties(properties, properties2 spec.SchemaProperties, path *field.Path) (spec.SchemaProperties, []conflict) {
	retProperties := make(spec.SchemaProperties)
	var retConflicts []conflict

	// go over first properties list
	// 1. merge mutual properties
	// 2. add non mutual properties
	for key := range properties {
		schema := properties[key]
		if schema2, ok := properties2[key]; ok {
			mergedSchema, conflicts := mergeSchema(&schema, &schema2, path.Child(key))
			retConflicts = append(retConflicts, conflicts...)
			retProperties[key] = *mergedSchema
		} else {
			retProperties[key] = schema
		}
	}

	// add non mutual properties from the second list
	for key, schema := range properties2 {
		if _, ok := properties[key]; !ok {
			retProperties[key] = schema
		}
	}

	return retProperties, retConflicts
}

func mergeResponses(responses, responses2 *spec.Responses, path *field.Path) (*spec.Responses, []conflict) {
	if r, shouldReturn := shouldReturnIfNil(responses, responses2); shouldReturn {
		return r.(*spec.Responses), nil
	}

	var retConflicts []conflict

	retResponses := &spec.Responses{
		ResponsesProps: spec.ResponsesProps{
			Default:             responses.Default, // default will be the same because we are generating it
			StatusCodeResponses: make(map[int]spec.Response),
		},
	}

	statusCodeResponses := responses.StatusCodeResponses
	statusCodeResponses2 := responses2.StatusCodeResponses

	// go over first responses list
	// 1. merge mutual response code responses
	// 2. add non mutual response code responses
	for code, response := range statusCodeResponses {
		if response2, ok := statusCodeResponses2[code]; ok {
			mergedResponse, conflicts := mergeResponse(response, response2, path.Child(strconv.Itoa(code)))
			retConflicts = append(retConflicts, conflicts...)
			retResponses.StatusCodeResponses[code] = *mergedResponse
		} else {
			retResponses.StatusCodeResponses[code] = statusCodeResponses[code]
		}
	}

	// add non mutual parameters from the second list
	for code := range statusCodeResponses2 {
		if _, ok := statusCodeResponses[code]; !ok {
			retResponses.StatusCodeResponses[code] = statusCodeResponses2[code]
		}
	}

	return retResponses, retConflicts
}

func mergeResponse(response, response2 spec.Response, path *field.Path) (*spec.Response, []conflict) {
	var retConflicts []conflict
	retResponse := spec.NewResponse()

	schema, conflicts := mergeSchema(response.Schema, response2.Schema, path.Child("schema"))
	retResponse.Schema = schema
	retConflicts = append(retConflicts, conflicts...)

	headers, conflicts := mergeResponseHeader(response.Headers, response2.Headers, path.Child("headers"))
	retResponse.Headers = headers
	retConflicts = append(retConflicts, conflicts...)

	return retResponse, retConflicts
}

func mergeResponseHeader(headers, headers2 map[string]spec.Header, path *field.Path) (map[string]spec.Header, []conflict) {
	var retConflicts []conflict
	retHeaders := make(map[string]spec.Header)

	// go over first headers list
	// 1. merge mutual headers
	// 2. add non mutual headers
	for name, header := range headers {
		if header2, ok := headers2[name]; ok {
			mergedHeader, conflicts := mergeHeader(header, header2, path.Child(name))
			retConflicts = append(retConflicts, conflicts...)
			retHeaders[name] = *mergedHeader
		} else {
			retHeaders[name] = headers[name]
		}
	}

	// add non mutual headers from the second list
	for name := range headers2 {
		if _, ok := headers[name]; !ok {
			retHeaders[name] = headers2[name]
		}
	}

	return retHeaders, retConflicts
}

func mergeHeader(header, header2 spec.Header, child *field.Path) (*spec.Header, []conflict) {
	retHeader := spec.ResponseHeader()

	simpleSchema, conflicts := mergeSimpleSchema(header.SimpleSchema, header2.SimpleSchema, child)
	retHeader.SimpleSchema = simpleSchema

	return retHeader, conflicts
}

func shouldReturnIfEmptyParameters(parameters, parameters2 []spec.Parameter) ([]spec.Parameter, bool) {
	if len(parameters) == 0 {
		return parameters2, true
	}
	if len(parameters2) == 0 {
		return parameters, true
	}
	// both are not empty
	return nil, false
}

func shouldReturnIfEmptySchemaType(s, s2 *spec.Schema) (*spec.Schema, bool) {
	if len(s.Type) == 0 {
		return s2, true
	}
	if len(s2.Type) == 0 {
		return s, true
	}
	// both are not empty
	return nil, false
}

// used only with pointers.
func shouldReturnIfNil(a, b interface{}) (interface{}, bool) {
	if utils.IsNil(a) {
		return b, true
	}
	if utils.IsNil(b) {
		return a, true
	}
	// both are not nil
	return nil, false
}
