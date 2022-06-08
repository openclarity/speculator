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
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/speculator/pkg/pathtrie"
)

type ProvidedSpec struct {
	Doc *openapi3.T
}

type openapi3header struct {
	OpenAPI *string `json:"openapi" yaml:"openapi"` // Required
}

type openapi2header struct {
	Swagger *string `json:"swagger" yaml:"swagger"`
}

func (s *Spec) LoadProvidedSpec(providedSpec []byte, pathToPathID map[string]string) error {
	doc, err := loadAndValidateRawJSONSpec(providedSpec)
	if err != nil {
		return fmt.Errorf("failed to load and validate spec: %w", err)
	}

	if s.ProvidedSpec == nil {
		s.ProvidedSpec = &ProvidedSpec{}
	}
	// will save doc without refs for proper diff logic
	s.ProvidedSpec.Doc = clearRefFromDoc(doc)

	// path trie need to be repopulated from start on each new spec
	s.ProvidedPathTrie = pathtrie.New()
	for path := range s.ProvidedSpec.Doc.Paths {
		if pathID, ok := pathToPathID[path]; ok {
			s.ProvidedPathTrie.Insert(path, pathID)
		}
	}

	return nil
}

func loadAndValidateRawJSONSpec(spec []byte) (*openapi3.T, error) {
	// Convert YAML to JSON. Since JSON is a subset of YAML, passing JSON through
	// this method should be a no-op.
	jsonSpec, err := yaml.YAMLToJSON(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to convert provided spec into json: %s. %v", spec, err)
	}

	var v3header openapi3header
	if err := json.Unmarshal(jsonSpec, &v3header); err != nil {
		return nil, fmt.Errorf("failed to unmarshel to v3header. %w", err)
	}

	var v2header openapi2header
	if err := json.Unmarshal(jsonSpec, &v2header); err != nil {
		return nil, fmt.Errorf("failed to unmarshel to v2header. %w", err)
	}

	var doc *openapi3.T
	if v3header.OpenAPI != nil {
		if doc, err = LoadAndValidateRawJSONSpecV3(jsonSpec); err != nil {
			log.Errorf("provided spec is not valid OpenAPI 3.0: %s. %v", jsonSpec, err)
			return nil, fmt.Errorf("provided spec is not valid OpenAPI 3.0: %v", err)
		}
	} else if v2header.Swagger != nil {
		if doc, err = LoadAndValidateRawJSONSpecV3FromV2(jsonSpec); err != nil {
			log.Errorf("provided spec is not valid OpenAPI 2.0: %s. %v", jsonSpec, err)
			return nil, fmt.Errorf("provided spec is not valid OpenAPI 2.0: %w", err)
		}
	} else {
		return nil, fmt.Errorf("provided spec missing spec header: %w", err)
	}

	return doc, nil
}

func (p *ProvidedSpec) GetPathItem(path string) *openapi3.PathItem {
	return p.Doc.Paths.Find(path)
}

func (p *ProvidedSpec) GetBasePath() string {
	for _, server := range p.Doc.Servers {
		if server.URL == "" || server.URL == "/" {
			continue
		}

		// strip scheme if exits
		urlNoScheme := server.URL
		schemeSplittedUrl := strings.Split(server.URL, "://")
		if len(schemeSplittedUrl) > 1 {
			urlNoScheme = schemeSplittedUrl[1]
		}

		// get path
		var path string
		splittedUrlNoScheme := strings.SplitN(urlNoScheme, "/", 2)
		if len(splittedUrlNoScheme) > 1 {
			path = splittedUrlNoScheme[1]
		}
		if path == "" {
			continue
		}

		return "/" + path
	}

	return ""
}

func clearRefFromDoc(doc *openapi3.T) *openapi3.T {
	if doc == nil {
		return doc
	}

	for path, item := range doc.Paths {
		doc.Paths[path] = clearRefFromPathItem(item)
	}

	return doc
}

func clearRefFromPathItem(item *openapi3.PathItem) *openapi3.PathItem {
	if item == nil {
		return item
	}

	for method, operation := range item.Operations() {
		item.SetOperation(method, clearRefFromOperation(operation))
	}

	item.Parameters = clearRefFromParameters(item.Parameters)

	item.Ref = ""

	return item
}

func clearRefFromParameters(parameters openapi3.Parameters) openapi3.Parameters {
	if len(parameters) == 0 {
		return parameters
	}

	retParameters := make(openapi3.Parameters, len(parameters))
	for i, parameterRef := range parameters {
		retParameters[i] = clearRefFromParameterRef(parameterRef)
	}

	return retParameters
}

func clearRefFromOperation(operation *openapi3.Operation) *openapi3.Operation {
	if operation == nil {
		return operation
	}

	operation.Parameters = clearRefFromParameters(operation.Parameters)
	operation.Responses = clearRefFromResponses(operation.Responses)
	operation.RequestBody = clearRefFromRequestBody(operation.RequestBody)

	return operation
}

func clearRefFromResponses(responses openapi3.Responses) openapi3.Responses {
	if len(responses) == 0 {
		return responses
	}

	retResponses := make(openapi3.Responses, len(responses))
	for i, parameterRef := range responses {
		retResponses[i] = clearRefFromResponseRef(parameterRef)
	}

	return retResponses
}

func clearRefFromRequestBody(requestBodyRef *openapi3.RequestBodyRef) *openapi3.RequestBodyRef {
	if requestBodyRef == nil {
		return requestBodyRef
	}

	return &openapi3.RequestBodyRef{
		Value: clearRefFromRequestBodyRef(requestBodyRef.Value),
	}
}

func clearRefFromRequestBodyRef(requestBody *openapi3.RequestBody) *openapi3.RequestBody {
	if requestBody == nil {
		return requestBody
	}

	requestBody.Content = clearRefFromContent(requestBody.Content)

	return requestBody
}

func clearRefFromResponseRef(responseRef *openapi3.ResponseRef) *openapi3.ResponseRef {
	if responseRef == nil {
		return responseRef
	}

	return &openapi3.ResponseRef{
		Value: clearRefFromResponse(responseRef.Value),
	}
}

func clearRefFromResponse(response *openapi3.Response) *openapi3.Response {
	if response == nil {
		return response
	}

	response.Headers = clearRefFromHeaders(response.Headers)
	response.Content = clearRefFromContent(response.Content)

	return response
}

func clearRefFromHeaders(headers openapi3.Headers) openapi3.Headers {
	if len(headers) == 0 {
		return headers
	}

	retHeaders := make(openapi3.Headers, len(headers))
	for key, headerRef := range headers {
		retHeaders[key] = clearRefFromHeaderRef(headerRef)
	}
	return retHeaders
}

func clearRefFromContent(content openapi3.Content) openapi3.Content {
	if len(content) == 0 {
		return content
	}

	retContent := make(openapi3.Content, len(content))
	for key, mediaType := range content {
		retContent[key] = clearRefFromMediaType(mediaType)
	}
	return retContent
}

func clearRefFromMediaType(mediaType *openapi3.MediaType) *openapi3.MediaType {
	if mediaType == nil {
		return mediaType
	}

	mediaType.Schema = clearRefFromSchemaRef(mediaType.Schema)
	return mediaType
}

func clearRefFromHeaderRef(headerRef *openapi3.HeaderRef) *openapi3.HeaderRef {
	if headerRef == nil {
		return headerRef
	}

	return &openapi3.HeaderRef{
		Value: clearRefFromHeader(headerRef.Value),
	}
}

func clearRefFromHeader(header *openapi3.Header) *openapi3.Header {
	if header == nil {
		return header
	}

	if parameter := clearRefFromParameter(&header.Parameter); parameter != nil {
		header.Parameter = *parameter
	}

	return header
}

func clearRefFromParameterRef(parameterRef *openapi3.ParameterRef) *openapi3.ParameterRef {
	if parameterRef == nil {
		return parameterRef
	}

	return &openapi3.ParameterRef{
		Value: clearRefFromParameter(parameterRef.Value),
	}
}

func clearRefFromParameter(parameter *openapi3.Parameter) *openapi3.Parameter {
	if parameter == nil {
		return parameter
	}

	parameter.Schema = clearRefFromSchemaRef(parameter.Schema)
	return parameter
}

func clearRefFromSchemaRef(schemaRef *openapi3.SchemaRef) *openapi3.SchemaRef {
	if schemaRef == nil {
		return schemaRef
	}

	return &openapi3.SchemaRef{
		Value: clearRefFromSchema(schemaRef.Value),
	}
}

func clearRefFromSchema(schema *openapi3.Schema) *openapi3.Schema {
	if schema == nil {
		return schema
	}

	schema.OneOf = clearRefFromSchemaRefs(schema.OneOf)
	schema.AnyOf = clearRefFromSchemaRefs(schema.AnyOf)
	schema.AllOf = clearRefFromSchemaRefs(schema.AllOf)
	schema.Not = clearRefFromSchemaRef(schema.Not)
	schema.Items = clearRefFromSchemaRef(schema.Items)
	schema.Properties = clearRefFromSchemas(schema.Properties)
	schema.AdditionalProperties = clearRefFromSchemaRef(schema.AdditionalProperties)

	return schema
}

func clearRefFromSchemas(schemas openapi3.Schemas) openapi3.Schemas {
	if len(schemas) == 0 {
		return schemas
	}

	retSchemas := make(openapi3.Schemas, len(schemas))
	for key, schemaRef := range schemas {
		retSchemas[key] = clearRefFromSchemaRef(schemaRef)
	}
	return retSchemas
}

func clearRefFromSchemaRefs(schemaRefs openapi3.SchemaRefs) openapi3.SchemaRefs {
	if len(schemaRefs) == 0 {
		return schemaRefs
	}

	retSchemaRefs := make(openapi3.SchemaRefs, len(schemaRefs))
	for i, schemaRef := range schemaRefs {
		retSchemaRefs[i] = clearRefFromSchemaRef(schemaRef)
	}
	return retSchemaRefs
}
