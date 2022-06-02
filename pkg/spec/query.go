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

	spec "github.com/getkin/kin-openapi/openapi3"
)

func addQueryParam(operation *spec.Operation, key string, values []string) *spec.Operation {
	var schema *spec.Schema
	if len(values) == 0 || values[0] == "" {
		schema = spec.NewBoolSchema()
		schema.AllowEmptyValue = true
	} else {
		schema = getSchemaFromValues(values, true, spec.ParameterInQuery)
	}

	operation.AddParameter(spec.NewQueryParameter(key).WithSchema(schema))
	return operation
}

func extractQueryParams(path string) (url.Values, error) {
	_, query := GetPathAndQuery(path)
	if query == "" {
		return nil, nil
	}

	values, err := url.ParseQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %v", err)
	}

	return values, nil
}
