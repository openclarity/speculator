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
	spec "github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
	"k8s.io/utils/field"

	"github.com/openclarity/speculator/pkg/utils"
)

var supportedParametersInTypes = []string{spec.ParameterInHeader, spec.ParameterInQuery, spec.ParameterInPath, spec.ParameterInCookie}

func mergeOperation(operation, operation2 *spec.Operation) (*spec.Operation, []conflict) {
	if op, shouldReturn := shouldReturnIfNil(operation, operation2); shouldReturn {
		return op.(*spec.Operation), nil
	}

	var requestBodyConflicts, paramConflicts, resConflicts []conflict

	ret := spec.NewOperation()

	ret.RequestBody, requestBodyConflicts = mergeRequestBody(operation.RequestBody, operation2.RequestBody,
		field.NewPath("requestBody"))
	ret.Parameters, paramConflicts = mergeParameters(operation.Parameters, operation2.Parameters,
		field.NewPath("parameters"))
	ret.Responses, resConflicts = mergeResponses(operation.Responses, operation2.Responses,
		field.NewPath("responses"))

	ret.Security = mergeOperationSecurity(operation.Security, operation2.Security)

	conflicts := append(paramConflicts, resConflicts...)
	conflicts = append(conflicts, requestBodyConflicts...)

	if len(conflicts) > 0 {
		log.Warnf("Found conflicts while merging operation: %v and operation: %v. conflicts: %v", operation, operation2, conflicts)
	}

	return ret, conflicts
}

func mergeOperationSecurity(security, security2 *spec.SecurityRequirements) *spec.SecurityRequirements {
	if s, shouldReturn := shouldReturnIfNil(security, security2); shouldReturn {
		return s.(*spec.SecurityRequirements)
	}

	var mergedSecurity spec.SecurityRequirements

	ignoreSecurityKeyMap := map[string]bool{}

	for _, securityMap := range *security {
		mergedSecurity, ignoreSecurityKeyMap = appendSecurityIfNeeded(securityMap, mergedSecurity, ignoreSecurityKeyMap)
	}
	for _, securityMap := range *security2 {
		mergedSecurity, ignoreSecurityKeyMap = appendSecurityIfNeeded(securityMap, mergedSecurity, ignoreSecurityKeyMap)
	}

	return &mergedSecurity
}

func appendSecurityIfNeeded(securityMap spec.SecurityRequirement, mergedSecurity spec.SecurityRequirements, ignoreSecurityKeyMap map[string]bool) (spec.SecurityRequirements, map[string]bool) {
	for key, values := range securityMap {
		// ignore if already appended the exact security key
		if ignoreSecurityKeyMap[key] {
			continue
		}
		// https://swagger.io/docs/specification/authentication/
		// We will treat multiple authentication types as an OR
		// (Security schemes combined via OR are alternatives – any one can be used in the given context)
		mergedSecurity = append(mergedSecurity, map[string][]string{key: values})
		ignoreSecurityKeyMap[key] = true
	}

	return mergedSecurity, ignoreSecurityKeyMap
}

func mergeRequestBody(body, body2 *spec.RequestBodyRef, path *field.Path) (*spec.RequestBodyRef, []conflict) {
	if p, shouldReturn := shouldReturnIfEmptyRequestBody(body, body2); shouldReturn {
		return p, nil
	}

	content, conflicts := mergeContent(body.Value.Content, body2.Value.Content, path.Child("content"))

	return &spec.RequestBodyRef{
		Value: spec.NewRequestBody().WithContent(content),
	}, conflicts
}

func shouldReturnIfEmptyRequestBody(body, body2 *spec.RequestBodyRef) (*spec.RequestBodyRef, bool) {
	if isEmptyRequestBody(body) {
		return body2, true
	}

	if isEmptyRequestBody(body2) {
		return body, true
	}

	return nil, false
}

func isEmptyRequestBody(body *spec.RequestBodyRef) bool {
	return body == nil || body.Value == nil || len(body.Value.Content) == 0
}

func mergeParameters(parameters, parameters2 spec.Parameters, path *field.Path) (spec.Parameters, []conflict) {
	if p, shouldReturn := shouldReturnIfEmptyParameters(parameters, parameters2); shouldReturn {
		return p, nil
	}

	var retParameters spec.Parameters
	var retConflicts []conflict

	parametersByIn := getParametersByIn(parameters)
	parameters2ByIn := getParametersByIn(parameters2)
	for _, inType := range supportedParametersInTypes {
		mergedParameters, conflicts := mergeParametersByInType(parametersByIn[inType], parameters2ByIn[inType], path)
		retParameters = append(retParameters, mergedParameters...)
		retConflicts = append(retConflicts, conflicts...)
	}

	return retParameters, retConflicts
}

func getParametersByIn(parameters spec.Parameters) map[string]spec.Parameters {
	ret := make(map[string]spec.Parameters)

	for i, parameter := range parameters {
		if parameter.Value == nil {
			continue
		}

		switch parameter.Value.In {
		case spec.ParameterInCookie, spec.ParameterInHeader, spec.ParameterInQuery, spec.ParameterInPath:
			ret[parameter.Value.In] = append(ret[parameter.Value.In], parameters[i])
		default:
			log.Warnf("in parameter not supported. %v", parameter.Value.In)
		}
	}

	return ret
}

func mergeParametersByInType(parameters, parameters2 spec.Parameters, path *field.Path) (spec.Parameters, []conflict) {
	if p, shouldReturn := shouldReturnIfEmptyParameters(parameters, parameters2); shouldReturn {
		return p, nil
	}

	var retParameters spec.Parameters
	var retConflicts []conflict

	parametersMapByName := makeParametersMapByName(parameters)
	parameters2MapByName := makeParametersMapByName(parameters2)

	// go over first parameters list
	// 1. merge mutual parameters
	// 2. add non-mutual parameters
	for name, param := range parametersMapByName {
		if param2, ok := parameters2MapByName[name]; ok {
			mergedParameter, conflicts := mergeParameter(param.Value, param2.Value, path.Child(name))
			retConflicts = append(retConflicts, conflicts...)
			retParameters = append(retParameters, &spec.ParameterRef{Value: mergedParameter})
		} else {
			retParameters = append(retParameters, param)
		}
	}

	// add non-mutual parameters from the second list
	for name, param := range parameters2MapByName {
		if _, ok := parametersMapByName[name]; !ok {
			retParameters = append(retParameters, param)
		}
	}

	return retParameters, retConflicts
}

func makeParametersMapByName(parameters spec.Parameters) map[string]*spec.ParameterRef {
	ret := make(map[string]*spec.ParameterRef)

	for i := range parameters {
		ret[parameters[i].Value.Name] = parameters[i]
	}

	return ret
}

func mergeParameter(parameter, parameter2 *spec.Parameter, path *field.Path) (*spec.Parameter, []conflict) {
	if p, shouldReturn := shouldReturnIfEmptyParameter(parameter, parameter2); shouldReturn {
		return p, nil
	}

	type1, type2 := parameter.Schema.Value.Type, parameter2.Schema.Value.Type
	switch conflictSolver(type1, type2) {
	case 0, 1:
		// do nothing, parameter is used.
	case 2:
		// use parameter2.
		type1 = type2
		parameter = parameter2
	case -1:
		return parameter, []conflict{
			{
				path: path,
				obj1: parameter,
				obj2: parameter2,
				msg:  createConflictMsg(path, type1, type2),
			},
		}
	}

	switch type1 {
	case spec.TypeBoolean, spec.TypeInteger, spec.TypeNumber, spec.TypeString:
		schema, conflicts := mergeSchema(parameter.Schema.Value, parameter2.Schema.Value, path)
		return parameter.WithSchema(schema), conflicts
	case spec.TypeArray:
		items, conflicts := mergeSchemaItems(parameter.Schema.Value.Items, parameter2.Schema.Value.Items, path)
		return parameter.WithSchema(spec.NewArraySchema().WithItems(items.Value)), conflicts
	case spec.TypeObject, "":
		// when type is missing it is probably an object - we should try and merge the parameter schema
		schema, conflicts := mergeSchema(parameter.Schema.Value, parameter2.Schema.Value, path.Child("schema"))
		return parameter.WithSchema(schema), conflicts
	default:
		log.Warnf("not supported schema type in parameter: %v", type1)
	}

	return parameter, nil
}

func mergeSchemaItems(items, items2 *spec.SchemaRef, path *field.Path) (*spec.SchemaRef, []conflict) {
	if s, shouldReturn := shouldReturnIfNil(items, items2); shouldReturn {
		return s.(*spec.SchemaRef), nil
	}
	schema, conflicts := mergeSchema(items.Value, items2.Value, path.Child("items"))
	return &spec.SchemaRef{Value: schema}, conflicts
}

func mergeSchema(schema, schema2 *spec.Schema, path *field.Path) (*spec.Schema, []conflict) {
	if s, shouldReturn := shouldReturnIfNil(schema, schema2); shouldReturn {
		return s.(*spec.Schema), nil
	}

	if s, shouldReturn := shouldReturnIfEmptySchemaType(schema, schema2); shouldReturn {
		return s, nil
	}

	switch conflictSolver(schema.Type, schema2.Type) {
	case 0, 1:
		// do nothing, schema is used.
	case 2:
		// use schema2.
		schema = schema2
	case -1:
		return schema, []conflict{
			{
				path: path,
				obj1: schema,
				obj2: schema2,
				msg:  createConflictMsg(path, schema.Type, schema2.Type),
			},
		}
	}

	switch schema.Type {
	case spec.TypeBoolean, spec.TypeInteger, spec.TypeNumber:
		return schema, nil
	case spec.TypeString:
		// Ignore format only if both schemas are string type and formats are different.
		if schema.Type == schema2.Type && schema.Format != schema2.Format {
			schema.Format = ""
		}
		return schema, nil
	case spec.TypeArray:
		items, conflicts := mergeSchemaItems(schema.Items, schema2.Items, path)
		schema.Items = items
		return schema, conflicts
	case spec.TypeObject:
		properties, conflicts := mergeProperties(schema.Properties, schema2.Properties, path.Child("properties"))
		schema.Properties = properties
		return schema, conflicts
	default:
		log.Warnf("not supported schema type in schema: %v", schema.Type)
	}

	return schema, nil
}

func mergeProperties(properties, properties2 spec.Schemas, path *field.Path) (spec.Schemas, []conflict) {
	retProperties := make(spec.Schemas)
	var retConflicts []conflict

	// go over first properties list
	// 1. merge mutual properties
	// 2. add non-mutual properties
	for key := range properties {
		schema := properties[key]
		if schema2, ok := properties2[key]; ok {
			mergedSchema, conflicts := mergeSchema(schema.Value, schema2.Value, path.Child(key))
			retConflicts = append(retConflicts, conflicts...)
			retProperties[key] = &spec.SchemaRef{Value: mergedSchema}
		} else {
			retProperties[key] = schema
		}
	}

	// add non-mutual properties from the second list
	for key, schema := range properties2 {
		if _, ok := properties[key]; !ok {
			retProperties[key] = schema
		}
	}

	return retProperties, retConflicts
}

func mergeResponses(responses, responses2 spec.Responses, path *field.Path) (spec.Responses, []conflict) {
	if r, shouldReturn := shouldReturnIfEmptyResponses(responses, responses2); shouldReturn {
		return r, nil
	}

	var retConflicts []conflict

	retResponses := spec.NewResponses()

	// go over first responses list
	// 1. merge mutual response code responses
	// 2. add non-mutual response code responses
	for code, response := range responses {
		if response2, ok := responses2[code]; ok {
			mergedResponse, conflicts := mergeResponse(response.Value, response2.Value, path.Child(code))
			retConflicts = append(retConflicts, conflicts...)
			retResponses[code] = &spec.ResponseRef{Value: mergedResponse}
		} else {
			retResponses[code] = responses[code]
		}
	}

	// add non-mutual parameters from the second list
	for code := range responses2 {
		if _, ok := responses[code]; !ok {
			retResponses[code] = responses2[code]
		}
	}

	return retResponses, retConflicts
}

func mergeResponse(response, response2 *spec.Response, path *field.Path) (*spec.Response, []conflict) {
	var retConflicts []conflict
	retResponse := spec.NewResponse()
	if response.Description != nil {
		retResponse = retResponse.WithDescription(*response.Description)
	} else if response2.Description != nil {
		retResponse = retResponse.WithDescription(*response2.Description)
	}

	content, conflicts := mergeContent(response.Content, response2.Content, path.Child("content"))
	if len(content) > 0 {
		retResponse = retResponse.WithContent(content)
	}
	retConflicts = append(retConflicts, conflicts...)

	headers, conflicts := mergeResponseHeader(response.Headers, response2.Headers, path.Child("headers"))
	if len(headers) > 0 {
		retResponse.Headers = headers
	}
	retConflicts = append(retConflicts, conflicts...)

	return retResponse, retConflicts
}

func mergeContent(content spec.Content, content2 spec.Content, path *field.Path) (spec.Content, []conflict) {
	var retConflicts []conflict
	retContent := spec.NewContent()

	// go over first content list
	// 1. merge mutual content media type
	// 2. add non-mutual content media type
	for name, mediaType := range content {
		if mediaType2, ok := content2[name]; ok {
			mergedSchema, conflicts := mergeSchema(mediaType.Schema.Value, mediaType2.Schema.Value, path.Child(name))
			// TODO: handle mediaType.Encoding
			retConflicts = append(retConflicts, conflicts...)
			retContent[name] = spec.NewMediaType().WithSchema(mergedSchema)
		} else {
			retContent[name] = content[name]
		}
	}

	// add non-mutual content media type from the second list
	for name := range content2 {
		if _, ok := content[name]; !ok {
			retContent[name] = content2[name]
		}
	}

	return retContent, retConflicts
}

func mergeResponseHeader(headers, headers2 spec.Headers, path *field.Path) (spec.Headers, []conflict) {
	var retConflicts []conflict
	retHeaders := make(spec.Headers)

	// go over first headers list
	// 1. merge mutual headers
	// 2. add non-mutual headers
	for name, header := range headers {
		if header2, ok := headers2[name]; ok {
			mergedHeader, conflicts := mergeHeader(header.Value, header2.Value, path.Child(name))
			retConflicts = append(retConflicts, conflicts...)
			retHeaders[name] = &spec.HeaderRef{Value: mergedHeader}
		} else {
			retHeaders[name] = headers[name]
		}
	}

	// add non-mutual headers from the second list
	for name := range headers2 {
		if _, ok := headers[name]; !ok {
			retHeaders[name] = headers2[name]
		}
	}

	return retHeaders, retConflicts
}

func mergeHeader(header, header2 *spec.Header, path *field.Path) (*spec.Header, []conflict) {
	if h, shouldReturn := shouldReturnIfEmptyHeader(header, header2); shouldReturn {
		return h, nil
	}

	if header.In != header2.In {
		return header, []conflict{
			{
				path: path,
				obj1: header,
				obj2: header2,
				msg:  createHeaderInConflictMsg(path, header.In, header2.In),
			},
		}
	}

	schema, conflicts := mergeSchema(header.Schema.Value, header2.Schema.Value, path)
	header.Parameter = *header.WithSchema(schema)

	return header, conflicts
}

func shouldReturnIfEmptyParameter(param, param2 *spec.Parameter) (*spec.Parameter, bool) {
	if isEmptyParameter(param) {
		return param2, true
	}

	if isEmptyParameter(param2) {
		return param, true
	}

	return nil, false
}

func isEmptyParameter(param *spec.Parameter) bool {
	return param == nil || isEmptySchemaRef(param.Schema)
}

func shouldReturnIfEmptyHeader(header, header2 *spec.Header) (*spec.Header, bool) {
	if isEmptyHeader(header) {
		return header2, true
	}

	if isEmptyHeader(header2) {
		return header, true
	}

	return nil, false
}

func isEmptyHeader(header *spec.Header) bool {
	return header == nil || isEmptySchemaRef(header.Schema)
}

func isEmptySchemaRef(schemaRef *spec.SchemaRef) bool {
	return schemaRef == nil || schemaRef.Value == nil
}

func shouldReturnIfEmptyResponses(r, r2 spec.Responses) (spec.Responses, bool) {
	if len(r) == 0 {
		return r2, true
	}
	if len(r2) == 0 {
		return r, true
	}
	// both are not empty
	return nil, false
}

func shouldReturnIfEmptyParameters(parameters, parameters2 spec.Parameters) (spec.Parameters, bool) {
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
