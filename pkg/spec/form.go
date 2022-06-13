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
	"mime/multipart"
	"net/url"
	"strings"

	spec "github.com/getkin/kin-openapi/openapi3"
)

const (
	// taken from net/http/request.go.
	defaultMaxMemory = 32 << 20 // 32 MB
)

func handleApplicationFormURLEncodedBody(operation *spec.Operation, securitySchemes spec.SecuritySchemes, body string) (*spec.Operation, spec.SecuritySchemes, error) {
	parseQuery, err := url.ParseQuery(body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse query. body=%v: %v", body, err)
	}

	objSchema := spec.NewObjectSchema()

	for key, values := range parseQuery {
		if key == AccessTokenParamKey {
			// https://datatracker.ietf.org/doc/html/rfc6750#section-2.2
			operation, securitySchemes = handleAuthQueryParam(operation, securitySchemes, values)
		} else {
			objSchema.WithProperty(key, getSchemaFromQueryValues(values))
		}
	}

	if len(objSchema.Properties) != 0 {
		operationSetRequestBody(operation, spec.NewRequestBody().WithContent(spec.NewContentWithSchema(objSchema, []string{mediaTypeApplicationForm})))
		// TODO: handle encoding
		// https://swagger.io/docs/specification/describing-request-body/
		// operation.RequestBody.Value.GetMediaType(mediaTypeApplicationForm).Encoding
	}

	return operation, securitySchemes, nil
}

func getMultipartFormDataSchema(body string, mediaTypeParams map[string]string) (*spec.Schema, error) {
	boundary, ok := mediaTypeParams["boundary"]
	if !ok {
		return nil, fmt.Errorf("no multipart boundary param in Content-Type")
	}

	form, err := multipart.NewReader(strings.NewReader(body), boundary).ReadForm(defaultMaxMemory)
	if err != nil {
		return nil, fmt.Errorf("failed to read form: %w", err)
	}

	schema := spec.NewObjectSchema()

	// https://swagger.io/docs/specification/describing-request-body/file-upload/
	for key, fileHeaders := range form.File {
		fileSchema := spec.NewStringSchema().WithFormat("binary")
		switch len(fileHeaders) {
		case 0:
			// do nothing
		case 1:
			// single file
			schema.WithProperty(key, fileSchema)
		default:
			// array of files
			schema.WithProperty(key, spec.NewArraySchema().WithItems(fileSchema))
		}
	}

	// add values formData
	for key, values := range form.Value {
		schema.WithProperty(key, getSchemaFromValues(values, false, ""))
	}

	return schema, nil
}
