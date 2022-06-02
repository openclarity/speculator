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
	"fmt"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/speculator/pkg/pathtrie"
)

type ProvidedSpec struct {
	Doc *openapi3.T
}

func (s *Spec) LoadProvidedSpec(providedSpec []byte, pathToPathID map[string]string) error {
	// Convert YAML to JSON. Since JSON is a subset of YAML, passing JSON through
	// this method should be a no-op.
	jsonSpec, err := yaml.YAMLToJSON(providedSpec)
	if err != nil {
		return fmt.Errorf("failed to convert provided spec into json: %s. %v", providedSpec, err)
	}

	doc, err := loadAndValidateRawJSONSpec(jsonSpec)
	if err != nil {
		log.Errorf("provided spec is not valid: %s. %v", jsonSpec, err)
		return fmt.Errorf("provided spec is not valid. %w", err)
	}

	s.ProvidedSpec.Doc = doc

	// path trie need to be repopulated from start on each new spec
	s.ProvidedPathTrie = pathtrie.New()
	for path := range s.ProvidedSpec.Doc.Paths {
		if pathID, ok := pathToPathID[path]; ok {
			s.ProvidedPathTrie.Insert(path, pathID)
		}
	}

	return nil
}

func (p *ProvidedSpec) GetPathItem(path string) *openapi3.PathItem {
	return p.Doc.Paths.Find(path)
}

func (p *ProvidedSpec) GetBasePath() string {
	for _, server := range p.Doc.Servers {
		if server.URL == "" || server.URL == "/" {
			continue
		}
		if u, err := url.Parse(server.URL); err != nil {
			log.Errorf("failed to parse server url %q: %v", server.URL, err)
		} else {
			return u.EscapedPath()
		}
	}

	return ""
}
