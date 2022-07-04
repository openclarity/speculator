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

	"github.com/openclarity/speculator/pkg/pathtrie"
)

func CreateDefaultSpec(host string, port string, config OperationGeneratorConfig) *Spec {
	return &Spec{
		SpecInfo: SpecInfo{
			Host: host,
			Port: port,
			LearningSpec: &LearningSpec{
				PathItems:       map[string]*spec.PathItem{},
				SecuritySchemes: spec.SecuritySchemes{},
			},
			ApprovedSpec: &ApprovedSpec{
				PathItems:       map[string]*spec.PathItem{},
				SecuritySchemes: spec.SecuritySchemes{},
			},
			ApprovedPathTrie: pathtrie.New(),
			ProvidedPathTrie: pathtrie.New(),
		},
		OpGenerator: NewOperationGenerator(config),
	}
}

func createDefaultSwaggerInfo() *spec.Info {
	return &spec.Info{
		Description:    "This is a generated Open API Spec",
		Title:          "Swagger",
		TermsOfService: "https://swagger.io/terms/",
		Contact: &spec.Contact{
			Email: "apiteam@swagger.io",
		},
		License: &spec.License{
			Name: "Apache 2.0",
			URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
		},
		Version: "1.0.0",
	}
}
