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

import oapi_spec "github.com/go-openapi/spec"

type LearningSpec struct {
	// map parameterized path into path item
	PathItems           map[string]*oapi_spec.PathItem
	SecurityDefinitions oapi_spec.SecurityDefinitions
}

func (l *LearningSpec) AddPathItem(path string, pathItem *oapi_spec.PathItem) {
	l.PathItems[path] = pathItem
}

func (l *LearningSpec) GetPathItem(path string) *oapi_spec.PathItem {
	pi, ok := l.PathItems[path]
	if !ok {
		return nil
	}

	return pi
}
