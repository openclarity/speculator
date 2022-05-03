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

	"github.com/ghodss/yaml"
	oapispec "github.com/go-openapi/spec"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/speculator/pkg/pathtrie"
)

type ProvidedSpec struct {
	Spec *oapispec.Swagger
}

func (s *Spec) LoadProvidedSpec(providedSpec []byte, pathToPathID map[string]string) error {
	// Convert YAML to JSON. Since JSON is a subset of YAML, passing JSON through
	// this method should be a no-op.
	jsonSpec, err := yaml.YAMLToJSON(providedSpec)
	if err != nil {
		return fmt.Errorf("failed to convert provided spec into json: %s. %v", providedSpec, err)
	}

	if err := validateRawJSONSpec(jsonSpec); err != nil {
		log.Errorf("provided spec is not valid: %s. %v", jsonSpec, err)
		return fmt.Errorf("provided spec is not valid. %w", err)
	}
	s.ProvidedSpec = &ProvidedSpec{
		Spec: &oapispec.Swagger{
			SwaggerProps: oapispec.SwaggerProps{
				Paths: &oapispec.Paths{
					Paths: map[string]oapispec.PathItem{},
				},
			},
		},
	}

	if err := json.Unmarshal(jsonSpec, s.ProvidedSpec.Spec); err != nil {
		return fmt.Errorf("failed to unmarshal spec: %v", err)
	}

	// path trie need to be repopulated from start on each new spec
	s.ProvidedPathTrie = pathtrie.New()
	for path := range s.ProvidedSpec.Spec.Paths.Paths {
		if pathID, ok := pathToPathID[path]; ok {
			s.ProvidedPathTrie.Insert(path, pathID)
		}
	}

	return nil
}

func (p *ProvidedSpec) GetPathItem(path string) *oapispec.PathItem {
	if pi, ok := p.Spec.Paths.Paths[path]; ok {
		return &pi
	}
	return nil
}
