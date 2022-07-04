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
	"net/http"
	"testing"

	oapi_spec "github.com/getkin/kin-openapi/openapi3"
	"gotest.tools/assert"
)

var req1 = `{"active":true,
"certificateVersion":"86eb5278-676a-3b7c-b29d-4a57007dc7be",
"controllerInstanceInfo":{"replicaId":"portshift-agent-66fc77c848-tmmk8"},
"policyAndAppVersion":1621477900361,
"version":"1.147.1"}`

var res1 = `{"cvss":[{"score":7.8,"vector":"AV:L/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H"}]}`

var req2 = `{"active":true,"statusCodes":["NO_METRICS_SERVER"],"version":"1.147.1"}`

var res2 = `{"cvss":[{"version":"3"}]}`

var combinedReq = `{"active":true,"statusCodes":["NO_METRICS_SERVER"],
"certificateVersion":"86eb5278-676a-3b7c-b29d-4a57007dc7be",
"controllerInstanceInfo":{"replicaId":"portshift-agent-66fc77c848-tmmk8"},
"policyAndAppVersion":1621477900361,
"version":"1.147.1"}`

var combinedRes = `{"cvss":[{"score":7.8,"vector":"AV:L/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H","version":"3"}]}`

type TestSpec struct {
	Doc *oapi_spec.T
}

func (t *TestSpec) WithPathItem(path string, pathItem *oapi_spec.PathItem) *TestSpec {
	t.Doc.Paths[path] = pathItem
	return t
}

type TestPathItem struct {
	PathItem oapi_spec.PathItem
}

func NewTestPathItem() *TestPathItem {
	return &TestPathItem{
		PathItem: oapi_spec.PathItem{},
	}
}

func (t *TestPathItem) WithPathParams(name string, schema *oapi_spec.Schema) *TestPathItem {
	pathParam := createPathParam(name, schema)
	t.PathItem.Parameters = append(t.PathItem.Parameters, &oapi_spec.ParameterRef{Value: pathParam.Parameter})
	return t
}

func (t *TestPathItem) WithOperation(method string, op *oapi_spec.Operation) *TestPathItem {
	switch method {
	case http.MethodGet:
		t.PathItem.Get = op
	case http.MethodDelete:
		t.PathItem.Delete = op
	case http.MethodOptions:
		t.PathItem.Options = op
	case http.MethodPatch:
		t.PathItem.Patch = op
	case http.MethodHead:
		t.PathItem.Head = op
	case http.MethodPost:
		t.PathItem.Post = op
	case http.MethodPut:
		t.PathItem.Put = op
	}
	return t
}

type TestOperation struct {
	Op *oapi_spec.Operation
}

func NewOperation(t *testing.T, data *HTTPInteractionData) *TestOperation {
	t.Helper()
	securitySchemes := oapi_spec.SecuritySchemes{}
	operation, err := CreateTestNewOperationGenerator().GenerateSpecOperation(data, securitySchemes)
	assert.NilError(t, err)
	return &TestOperation{
		Op: operation,
	}
}

func CreateTestNewOperationGenerator() *OperationGenerator {
	return NewOperationGenerator(testOperationGeneratorConfig)
}

var testOperationGeneratorConfig = OperationGeneratorConfig{
	ResponseHeadersToIgnore: []string{contentTypeHeaderName},
	RequestHeadersToIgnore:  []string{acceptTypeHeaderName, authorizationTypeHeaderName, contentTypeHeaderName},
}

func (op *TestOperation) Deprecated() *TestOperation {
	op.Op.Deprecated = true
	return op
}

func (op *TestOperation) WithResponse(status int, response *oapi_spec.Response) *TestOperation {
	op.Op.AddResponse(status, response)
	if status != 0 {
		// we don't need it to create default response in tests unless we explicitly asked for (status == 0)
		delete(op.Op.Responses, "default")
	}
	return op
}

func (op *TestOperation) WithParameter(param *oapi_spec.Parameter) *TestOperation {
	op.Op.AddParameter(param)
	return op
}

func (op *TestOperation) WithRequestBody(requestBody *oapi_spec.RequestBody) *TestOperation {
	operationSetRequestBody(op.Op, requestBody)
	return op
}

func (op *TestOperation) WithSecurityRequirement(securityRequirement oapi_spec.SecurityRequirement) *TestOperation {
	if op.Op.Security == nil {
		op.Op.Security = oapi_spec.NewSecurityRequirements()
	}
	op.Op.Security.With(securityRequirement)
	return op
}

func createTestOperation() *TestOperation {
	return &TestOperation{Op: oapi_spec.NewOperation()}
}

type TestResponse struct {
	*oapi_spec.Response
}

func createTestResponse() *TestResponse {
	return &TestResponse{
		Response: oapi_spec.NewResponse(),
	}
}

func (r *TestResponse) WithHeader(name string, schema *oapi_spec.Schema) *TestResponse {
	if r.Response.Headers == nil {
		r.Response.Headers = make(oapi_spec.Headers)
	}
	r.Response.Headers[name] = &oapi_spec.HeaderRef{
		Value: &oapi_spec.Header{
			Parameter: oapi_spec.Parameter{
				Schema: &oapi_spec.SchemaRef{
					Value: schema,
				},
			},
		},
	}
	return r
}

func (r *TestResponse) WithJSONSchema(schema *oapi_spec.Schema) *TestResponse {
	r.Response.WithJSONSchema(schema)
	return r
}

type TestResponses struct {
	oapi_spec.Responses
}

func createTestResponses() *TestResponses {
	return &TestResponses{
		Responses: oapi_spec.NewResponses(),
	}
}

func (r *TestResponses) WithResponse(code string, response *oapi_spec.Response) *TestResponses {
	r.Responses[code] = &oapi_spec.ResponseRef{
		Value: response,
	}

	return r
}
