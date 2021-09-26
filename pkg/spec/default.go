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
	"github.com/go-openapi/spec"

	"github.com/apiclarity/speculator/pkg/pathtrie"
)

func CreateDefaultSpec(host string, port string, config OperationGeneratorConfig) *Spec {
	return &Spec{
		SpecInfo: SpecInfo{
			Host: host,
			Port: port,
			LearningSpec: &LearningSpec{
				PathItems:           map[string]*spec.PathItem{},
				SecurityDefinitions: map[string]*spec.SecurityScheme{},
			},
			ApprovedSpec: &ApprovedSpec{
				PathItems:           map[string]*spec.PathItem{},
				SecurityDefinitions: map[string]*spec.SecurityScheme{},
			},
			PathTrie: pathtrie.New(),
		},
		opGenerator: NewOperationGenerator(config),
	}
}

func createDefaultSwaggerInfo() *spec.Info {
	return &spec.Info{
		InfoProps: spec.InfoProps{
			Description:    "This is a generated Open API Spec",
			Title:          "Swagger",
			TermsOfService: "http://swagger.io/terms/",
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Email: "apiteam@swagger.io",
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "Apache 2.0",
					URL:  "http://www.apache.org/licenses/LICENSE-2.0.html",
				},
			},
			Version: "1.0.0",
		},
	}
}
