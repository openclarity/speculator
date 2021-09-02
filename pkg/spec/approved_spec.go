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

	oapi_spec "github.com/go-openapi/spec"
)

type ApprovedSpec struct {
	PathItems           map[string]*oapi_spec.PathItem
	SecurityDefinitions oapi_spec.SecurityDefinitions
}

func (a *ApprovedSpec) GetPathItem(path string) *oapi_spec.PathItem {
	if pi, ok := a.PathItems[path]; ok {
		return pi
	}
	return nil
}

func (a *ApprovedSpec) Clone() (*ApprovedSpec, error) {
	clonedApprovedSpec := new(ApprovedSpec)

	approvedSpecB, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal approved spec: %v", err)
	}

	if err := json.Unmarshal(approvedSpecB, &clonedApprovedSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal approved spec: %v", err)
	}

	return clonedApprovedSpec, nil
}
