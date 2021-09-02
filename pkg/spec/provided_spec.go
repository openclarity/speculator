/*
 *
 * Copyright (c) 2020 Cisco Systems, Inc. and its affiliates.
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package spec

import (
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	oapi_spec "github.com/go-openapi/spec"
)

type ProvidedSpec struct {
	Spec *oapi_spec.Swagger
}

func (s *Spec) LoadProvidedSpec(providedSpec []byte) error {
	// Convert YAML to JSON. Since JSON is a subset of YAML, passing JSON through
	// this method should be a no-op.
	jsonSpec, err := yaml.YAMLToJSON(providedSpec)
	if err != nil {
		return fmt.Errorf("failed to convert provided spec into json: %s. %v", providedSpec, err)
	}

	if err := validateRawJsonSpec(jsonSpec); err != nil {
		return fmt.Errorf("provided spec is not valid: %s. %v", jsonSpec, err)
	}
	s.ProvidedSpec = &ProvidedSpec{
		Spec: &oapi_spec.Swagger{
			SwaggerProps: oapi_spec.SwaggerProps{
				Paths: &oapi_spec.Paths{
					Paths: map[string]oapi_spec.PathItem{},
				},
			},
		},
	}

	return json.Unmarshal(jsonSpec, s.ProvidedSpec.Spec)
}
