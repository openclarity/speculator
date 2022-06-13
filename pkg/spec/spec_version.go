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
)

type OASVersion int64

const (
	Unknown OASVersion = iota
	OASv2
	OASv3
)

func (o OASVersion) String() string {
	switch o {
	case OASv2:
		return "OASv2"
	case OASv3:
		return "OASv3"
	}
	return "unknown"
}

type oasV3header struct {
	OpenAPI *string `json:"openapi" yaml:"openapi"` // Required
}

type oasV2header struct {
	Swagger *string `json:"swagger" yaml:"swagger"`
}

func GetJsonSpecVersion(jsonSpec []byte) (OASVersion, error) {
	var v3header oasV3header
	if err := json.Unmarshal(jsonSpec, &v3header); err != nil {
		return Unknown, fmt.Errorf("failed to unmarshel to v3header. %w", err)
	}

	var v2header oasV2header
	if err := json.Unmarshal(jsonSpec, &v2header); err != nil {
		return Unknown, fmt.Errorf("failed to unmarshel to v2header. %w", err)
	}

	// openapi field is required in the OpenAPI Specification
	if v3header.OpenAPI != nil && *v3header.OpenAPI != "" {
		return OASv3, nil
	}

	if v2header.Swagger != nil {
		return OASv2, nil
	}

	return Unknown, fmt.Errorf("provided spec missing spec header")
}
