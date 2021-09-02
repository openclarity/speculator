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
	"net/http"
	"testing"

	"gotest.tools/assert"

	oapi_spec "github.com/go-openapi/spec"
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
	Spec *oapi_spec.Swagger
}

func NewTestSpec() *TestSpec {
	return &TestSpec{
		Spec: &oapi_spec.Swagger{
			SwaggerProps: oapi_spec.SwaggerProps{
				Paths: &oapi_spec.Paths{
					Paths: map[string]oapi_spec.PathItem{},
				},
			},
		},
	}
}

func (t *TestSpec) WithPathItem(path string, pathItem oapi_spec.PathItem) *TestSpec {
	t.Spec.Paths.Paths[path] = pathItem
	return t
}

type TestPathItem struct {
	PathItem oapi_spec.PathItem
}

func NewTestPathItem() *TestPathItem {
	return &TestPathItem{
		PathItem: oapi_spec.PathItem{
			PathItemProps: oapi_spec.PathItemProps{},
		},
	}
}

func (t *TestPathItem) WithPathParams(name, tpe, format string) *TestPathItem {
	pathParam := createPathParam(name, tpe, format)
	t.PathItem.Parameters = append(t.PathItem.Parameters, *pathParam.Parameter)
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
	sd := oapi_spec.SecurityDefinitions{}
	operation, err := GenerateSpecOperation(data, sd)
	assert.NilError(t, err)
	return &TestOperation{
		Op: operation,
	}
}
