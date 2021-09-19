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
	"sync"

	"github.com/ghodss/yaml"
	"github.com/go-openapi/loads"
	oapi_spec "github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	uuid "github.com/satori/go.uuid"

	"github.com/apiclarity/speculator/pkg/pathtrie"
)

// NOTE: when adding new fields to the Spec, make sure they to copy them also in ApplyApprovedReview (if needed)
type Spec struct {
	// Host of the spec
	Host string

	Port string
	// Spec ID
	ID uuid.UUID
	// Provided Spec
	ProvidedSpec *ProvidedSpec
	// Merged & approved state (can be generated into spec YAML)
	ApprovedSpec *ApprovedSpec
	// Upon learning, this will be updated (not the ApprovedSpec field)
	LearningSpec *LearningSpec

	PathTrie pathtrie.PathTrie

	lock sync.Mutex
}

type LearningParametrizedPaths struct {
	// map parameterized paths into a list of paths included in it.
	// e.g: /api/{param1} -> /api/1, /api/2
	// non parameterized path will map to itself
	Paths map[string]map[string]bool
}

type (
	SCNTelemetry struct {
		RequestID          string       `json:"request_id"`
		Scheme             string       `json:"scheme"`
		DestinationAddress string       `json:"destination_address"`
		SourceAddress      string       `json:"source_address"`
		SCNTRequest        SCNTRequest  `json:"scnt_request"`
		SCNTResponse       SCNTResponse `json:"scnt_response"`
	}

	SCNTRequest struct {
		Method string `json:"method"`
		Path   string `json:"path"`
		Host   string `json:"host"`
		SCNTCommon
	}

	SCNTResponse struct {
		StatusCode string `json:"status_code"`
		SCNTCommon
	}

	SCNTCommon struct {
		Version       string      `json:"version"`
		Headers       [][2]string `json:"headers"`
		Body          []byte      `json:"body"`
		TruncatedBody bool        `json:"truncated_body"`
	}
)

func (s *Spec) HasApprovedSpec() bool {
	if s.ApprovedSpec == nil || len(s.ApprovedSpec.PathItems) == 0 {
		return false
	}

	return true
}

func (s *Spec) HasProvidedSpec() bool {
	if s.ProvidedSpec == nil || s.ProvidedSpec.Spec == nil || s.ProvidedSpec.Spec.Paths == nil || s.ProvidedSpec.Spec.Paths.Paths == nil {
		return false
	}

	return true
}

func (s *Spec) LearnTelemetry(telemetry *SCNTelemetry) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	method := telemetry.SCNTRequest.Method
	// remove query params if exists
	path, _ := GetPathAndQuery(telemetry.SCNTRequest.Path)
	telemetryOp, err := telemetryToOperation(telemetry, s.LearningSpec.SecurityDefinitions)
	if err != nil {
		return fmt.Errorf("failed to convert telemetry to operation. %v", err)
	}
	var existingOp *oapi_spec.Operation

	// Get existing path item or create a new one
	pathItem := s.LearningSpec.GetPathItem(path)
	if pathItem == nil {
		pathItem = &oapi_spec.PathItem{}
	}

	// Get existing operation of path item, and if exists, merge it with the operation learned from this interaction
	existingOp = GetOperationFromPathItem(pathItem, method)
	if existingOp != nil {
		telemetryOp, _ = mergeOperation(existingOp, telemetryOp)
	}

	// save Operation on the path item
	AddOperationToPathItem(pathItem, method, telemetryOp)

	// add/update this path item in the spec
	s.LearningSpec.AddPathItem(path, pathItem)

	return nil
}

func (s *Spec) GenerateOASYaml() ([]byte, error) {
	oasJSON, err := s.GenerateOASJson()
	if err != nil {
		return nil, fmt.Errorf("failed to generate json spec: %v", err)
	}

	oasYaml, err := yaml.JSONToYAML(oasJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to convert json to yaml: %v", err)
	}

	return oasYaml, nil
}

func (s *Spec) GenerateOASJson() ([]byte, error) {
	// yaml.Marshal does not omit empty fields
	var definitions oapi_spec.Definitions

	clonedApprovedSpec, err := s.ApprovedSpec.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone approved spec. %v", err)
	}

	clonedApprovedSpec.PathItems, definitions = reconstructObjectRefs(clonedApprovedSpec.PathItems)

	generatedSpec := &oapi_spec.Swagger{
		SwaggerProps: oapi_spec.SwaggerProps{
			Host:    s.Host + ":" + s.Port,
			Swagger: "2.0",
			Info:    createDefaultSwaggerInfo(),
			Paths: &oapi_spec.Paths{
				Paths: map[string]oapi_spec.PathItem{},
			},
			Definitions:         definitions,
			SecurityDefinitions: clonedApprovedSpec.SecurityDefinitions,
		},
	}

	for path, approvedPathItem := range clonedApprovedSpec.PathItems {
		generatedSpec.Paths.Paths[path] = *approvedPathItem
	}

	ret, err := json.Marshal(generatedSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the spec. %v", err)
	}
	if err := validateRawJSONSpec(ret); err != nil {
		return nil, fmt.Errorf("failed to validate the spec. %v\n\nspec: %s", err, ret)
	}

	return ret, nil
}

func (s *Spec) Clone() (*Spec, error) {
	clonedSpec := new(Spec)

	specB, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal spec: %w", err)
	}

	if err := json.Unmarshal(specB, &clonedSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
	}

	return clonedSpec, nil
}

func validateRawJSONSpec(spec []byte) error {
	doc, err := loads.Analyzed(spec, "")
	if err != nil {
		return fmt.Errorf("failed to analyze spec: %s. %v", spec, err)
	}
	err = validate.Spec(doc, strfmt.Default)
	if err != nil {
		return fmt.Errorf("spec validation failed. %v", err)
	}
	return nil
}

func reconstructObjectRefs(pathItems map[string]*oapi_spec.PathItem) (retPathItems map[string]*oapi_spec.PathItem, definitions map[string]oapi_spec.Schema) {
	for _, item := range pathItems {
		definitions, item.Get = updateDefinitions(definitions, item.Get)
		definitions, item.Put = updateDefinitions(definitions, item.Put)
		definitions, item.Post = updateDefinitions(definitions, item.Post)
		definitions, item.Delete = updateDefinitions(definitions, item.Delete)
		definitions, item.Options = updateDefinitions(definitions, item.Options)
		definitions, item.Head = updateDefinitions(definitions, item.Head)
		definitions, item.Patch = updateDefinitions(definitions, item.Patch)
	}

	return pathItems, definitions
}
