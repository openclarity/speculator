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
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	oapi_spec "github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/speculator/pkg/pathtrie"
	"github.com/openclarity/speculator/pkg/utils/errors"
)

type SpecSource string

const (
	SpecSourceReconstructed SpecSource = "RECONSTRUCTED"
	SpecSourceProvided      SpecSource = "PROVIDED"
)

type Spec struct {
	SpecInfo

	OpGenerator *OperationGenerator

	lock sync.Mutex
}

type SpecInfo struct {
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

	ApprovedPathTrie pathtrie.PathTrie
	ProvidedPathTrie pathtrie.PathTrie
}

type LearningParametrizedPaths struct {
	// map parameterized paths into a list of paths included in it.
	// e.g: /api/{param1} -> /api/1, /api/2
	// non parameterized path will map to itself
	Paths map[string]map[string]bool
}

type Telemetry struct {
	DestinationAddress   string    `json:"destinationAddress,omitempty"`
	DestinationNamespace string    `json:"destinationNamespace,omitempty"`
	Request              *Request  `json:"request,omitempty"`
	RequestID            string    `json:"requestID,omitempty"`
	Response             *Response `json:"response,omitempty"`
	Scheme               string    `json:"scheme,omitempty"`
	SourceAddress        string    `json:"sourceAddress,omitempty"`
}

type Request struct {
	Common *Common `json:"common,omitempty"`
	Host   string  `json:"host,omitempty"`
	Method string  `json:"method,omitempty"`
	Path   string  `json:"path,omitempty"`
}

type Response struct {
	Common     *Common `json:"common,omitempty"`
	StatusCode string  `json:"statusCode,omitempty"`
}

type Common struct {
	TruncatedBody bool      `json:"TruncatedBody,omitempty"`
	Body          []byte    `json:"body,omitempty"`
	Headers       []*Header `json:"headers"`
	Version       string    `json:"version,omitempty"`
}

type Header struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func (s *Spec) HasApprovedSpec() bool {
	if s.ApprovedSpec == nil || len(s.ApprovedSpec.PathItems) == 0 {
		return false
	}

	return true
}

func (s *Spec) HasProvidedSpec() bool {
	if s.ProvidedSpec == nil || s.ProvidedSpec.Doc == nil || s.ProvidedSpec.Doc.Paths == nil {
		return false
	}

	return true
}

func (s *Spec) UnsetApprovedSpec() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.ApprovedSpec = &ApprovedSpec{
		PathItems:       make(oapi_spec.Paths),
		SecuritySchemes: make(oapi_spec.SecuritySchemes),
	}
	s.LearningSpec = &LearningSpec{
		PathItems:       make(oapi_spec.Paths),
		SecuritySchemes: make(oapi_spec.SecuritySchemes),
	}
	s.ApprovedPathTrie = pathtrie.New()
}

func (s *Spec) UnsetProvidedSpec() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.ProvidedSpec = nil
	s.ProvidedPathTrie = pathtrie.New()
}

func (s *Spec) LearnTelemetry(telemetry *Telemetry) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	method := telemetry.Request.Method
	// remove query params if exists
	path, _ := GetPathAndQuery(telemetry.Request.Path)
	telemetryOp, err := s.telemetryToOperation(telemetry, s.LearningSpec.SecuritySchemes)
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

func (s *Spec) GetPathID(path string, specSource SpecSource) (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	var specID string

	switch specSource {
	case SpecSourceProvided:
		if !s.HasProvidedSpec() {
			log.Infof("No provided spec, path id will be empty")
			return "", nil
		}
		basePath := s.ProvidedSpec.GetBasePath()

		pathNoBase := trimBasePathIfNeeded(basePath, path)

		_, value, found := s.ProvidedPathTrie.GetPathAndValue(pathNoBase)
		if found {
			if pathID, ok := value.(string); !ok {
				log.Warnf("value is not a string. %v", value)
			} else {
				specID = pathID
			}
		}
	case SpecSourceReconstructed:
		if !s.HasApprovedSpec() {
			log.Infof("No approved spec. path id will be empty")
			return "", nil
		}
		log.Infof("********* %+v %+v", s.ApprovedPathTrie, s.ApprovedPathTrie.Trie)
		for k, v := range s.ApprovedPathTrie.Trie {
			log.Infof("************* %v -> %+v", k, v)
			for kk, vv := range v.Children {
				log.Infof("***************** %v -> %+v", kk, vv)
			}
			for kkk, vvv := range vv.Children {
				log.Infof("***************** %v -> %+v", kkk, vvv)
			}
		}
		_, value, found := s.ApprovedPathTrie.GetPathAndValue(path)
		if found {
			if pathID, ok := value.(string); !ok {
				log.Warnf("value is not a string. %v", value)
			} else {
				specID = pathID
			}
		}
	default:
		return "", fmt.Errorf("spec source: %v is not valid", specSource)
	}
	return specID, nil
}

func (s *Spec) GenerateOASYaml(version OASVersion) ([]byte, error) {
	oasJSON, err := s.GenerateOASJson(version)
	if err != nil {
		return nil, fmt.Errorf("failed to generate json spec: %w", err)
	}

	oasYaml, err := yaml.JSONToYAML(oasJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to convert json to yaml: %v", err)
	}

	return oasYaml, nil
}

func (s *Spec) GenerateOASJson(version OASVersion) ([]byte, error) {
	// yaml.Marshal does not omit empty fields
	var schemas oapi_spec.Schemas

	clonedApprovedSpec, err := s.ApprovedSpec.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone approved spec. %v", err)
	}

	clonedApprovedSpec.PathItems, schemas = reconstructObjectRefs(clonedApprovedSpec.PathItems)

	generatedSpec := &oapi_spec.T{
		OpenAPI: "3.0.3",
		Components: oapi_spec.Components{
			Schemas: schemas,
		},
		Info:  createDefaultSwaggerInfo(),
		Paths: clonedApprovedSpec.PathItems,
		Servers: oapi_spec.Servers{
			{
				// https://swagger.io/docs/specification/api-host-and-base-path/
				URL: "http://" + s.Host + ":" + s.Port,
			},
		},
	}

	var ret []byte
	if version == OASv2 {
		log.Debugf("Generating OASv2 spec")
		generatedSpecV2, err := openapi2conv.FromV3(generatedSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to convert spec from v3: %v", err)
		}

		ret, err = json.Marshal(generatedSpecV2)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal the spec. %v", err)
		}
	} else {
		log.Debugf("Generating OASv3 spec")
		ret, err = json.Marshal(generatedSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal the spec. %v", err)
		}
	}

	if _, _, err = LoadAndValidateRawJSONSpec(ret); err != nil {
		log.Errorf("Failed to validate the spec. %v\n\nspec: %s", err, ret)
		return nil, fmt.Errorf("failed to validate the spec. %w", err)
	}

	return ret, nil
}

func (s *Spec) SpecInfoClone() (*Spec, error) {
	var clonedSpecInfo SpecInfo

	specB, err := json.Marshal(s.SpecInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal spec info: %w", err)
	}

	if err := json.Unmarshal(specB, &clonedSpecInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec info: %w", err)
	}

	return &Spec{
		SpecInfo: clonedSpecInfo,
		lock:     sync.Mutex{},
	}, nil
}

func LoadAndValidateRawJSONSpecV3(spec []byte) (*oapi_spec.T, error) {
	loader := oapi_spec.NewLoader()
	loader.Context = context.TODO()

	doc, err := loader.LoadFromData(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to load data: %s. %w", spec, err)
	}

	err = doc.Validate(loader.Context)
	if err != nil {
		return nil, fmt.Errorf("spec validation failed. %v. %w", err, errors.ErrSpecValidation)
	}

	return doc, nil
}

func LoadAndValidateRawJSONSpecV3FromV2(spec []byte) (*oapi_spec.T, error) {
	loader := oapi_spec.NewLoader()
	loader.Context = context.TODO()

	var doc openapi2.T
	if err := json.Unmarshal(spec, &doc); err != nil {
		return nil, fmt.Errorf("provided spec is not valid. %w", err)
	}

	v3, err := openapi2conv.ToV3(&doc)
	if err != nil {
		return nil, fmt.Errorf("conversion to V3 failed. %w", err)
	}

	err = v3.Validate(loader.Context)
	if err != nil {
		return nil, fmt.Errorf("spec validation failed. %v. %w", err, errors.ErrSpecValidation)
	}

	return v3, nil
}

func reconstructObjectRefs(pathItems map[string]*oapi_spec.PathItem) (retPathItems map[string]*oapi_spec.PathItem, schemas oapi_spec.Schemas) {
	for _, item := range pathItems {
		schemas, item.Get = updateSchemas(schemas, item.Get)
		schemas, item.Put = updateSchemas(schemas, item.Put)
		schemas, item.Post = updateSchemas(schemas, item.Post)
		schemas, item.Delete = updateSchemas(schemas, item.Delete)
		schemas, item.Options = updateSchemas(schemas, item.Options)
		schemas, item.Head = updateSchemas(schemas, item.Head)
		schemas, item.Patch = updateSchemas(schemas, item.Patch)
	}

	return pathItems, schemas
}
