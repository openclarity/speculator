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
	"reflect"
	"sort"
	"testing"

	spec "github.com/getkin/kin-openapi/openapi3"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/assert"
	"k8s.io/utils/field"
)

func Test_merge(t *testing.T) {
	securitySchemes := spec.SecuritySchemes{}
	op := CreateTestNewOperationGenerator()
	op1, err := op.GenerateSpecOperation(&HTTPInteractionData{
		ReqBody:     req1,
		RespBody:    res1,
		ReqHeaders:  map[string]string{"X-Test-Req-1": "1", contentTypeHeaderName: mediaTypeApplicationJSON},
		RespHeaders: map[string]string{"X-Test-Res-1": "1", contentTypeHeaderName: mediaTypeApplicationJSON},
		statusCode:  200,
	}, securitySchemes)
	assert.NilError(t, err)

	op2, err := op.GenerateSpecOperation(&HTTPInteractionData{
		ReqBody:     req2,
		RespBody:    res2,
		ReqHeaders:  map[string]string{"X-Test-Req-2": "2", contentTypeHeaderName: mediaTypeApplicationJSON},
		RespHeaders: map[string]string{"X-Test-Res-2": "2", contentTypeHeaderName: mediaTypeApplicationJSON},
		statusCode:  200,
	}, securitySchemes)
	assert.NilError(t, err)

	combinedOp, err := op.GenerateSpecOperation(&HTTPInteractionData{
		ReqBody:     combinedReq,
		RespBody:    combinedRes,
		ReqHeaders:  map[string]string{"X-Test-Req-1": "1", "X-Test-Req-2": "2", contentTypeHeaderName: mediaTypeApplicationJSON},
		RespHeaders: map[string]string{"X-Test-Res-1": "1", "X-Test-Res-2": "2", contentTypeHeaderName: mediaTypeApplicationJSON},
		statusCode:  200,
	}, securitySchemes)
	assert.NilError(t, err)

	type args struct {
		operation1 *spec.Operation
		operation2 *spec.Operation
	}
	tests := []struct {
		name          string
		args          args
		want          *spec.Operation
		wantConflicts bool
	}{
		{
			name: "sanity",
			args: args{
				operation1: op1,
				operation2: op2,
			},
			want: combinedOp,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, conflicts := mergeOperation(tt.args.operation1, tt.args.operation2)
			if (len(conflicts) > 0) != tt.wantConflicts {
				t.Errorf("merge() conflicts = %v, wantConflicts %v", conflicts, tt.wantConflicts)
				return
			}
			got = sortParameters(got)
			tt.want = sortParameters(tt.want)
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreUnexported(spec.Schema{}), cmpopts.IgnoreTypes(spec.ExtensionProps{}))
		})
	}
}

func Test_shouldReturnIfEmpty(t *testing.T) {
	type args struct {
		a interface{}
		b interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  interface{}
		want1 bool
	}{
		{
			name: "second nil",
			args: args{
				a: spec.NewOperation(),
				b: nil,
			},
			want:  spec.NewOperation(),
			want1: true,
		},
		{
			name: "first nil",
			args: args{
				a: nil,
				b: spec.NewOperation(),
			},
			want:  spec.NewOperation(),
			want1: true,
		},
		{
			name: "both nil",
			args: args{
				a: nil,
				b: nil,
			},
			want:  nil,
			want1: true,
		},
		{
			name: "not nil",
			args: args{
				a: spec.NewOperation(),
				b: spec.NewOperation(),
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := shouldReturnIfNil(tt.args.a, tt.args.b)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shouldReturnIfNil() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("shouldReturnIfNil() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_shouldReturnIfNil(t *testing.T) {
	var nilSchema *spec.Schema
	schema := spec.Schema{Type: "test"}
	type args struct {
		a interface{}
		b interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  interface{}
		want1 bool
	}{
		{
			name: "a is nil b is not",
			args: args{
				a: nilSchema,
				b: schema,
			},
			want:  schema,
			want1: true,
		},
		{
			name: "b is nil a is not",
			args: args{
				a: schema,
				b: nilSchema,
			},
			want:  schema,
			want1: true,
		},
		{
			name: "both nil",
			args: args{
				a: nilSchema,
				b: nilSchema,
			},
			want:  nilSchema,
			want1: true,
		},
		{
			name: "both not nil",
			args: args{
				a: schema,
				b: schema,
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := shouldReturnIfNil(tt.args.a, tt.args.b)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shouldReturnIfNil() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("shouldReturnIfNil() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_shouldReturnIfEmptySchemaType(t *testing.T) {
	emptySchemaType := &spec.Schema{}
	schema := &spec.Schema{Type: "test"}
	type args struct {
		s  *spec.Schema
		s2 *spec.Schema
	}
	tests := []struct {
		name  string
		args  args
		want  *spec.Schema
		want1 bool
	}{
		{
			name: "first is empty second is not",
			args: args{
				s:  emptySchemaType,
				s2: schema,
			},
			want:  schema,
			want1: true,
		},
		{
			name: "second is empty first is not",
			args: args{
				s:  schema,
				s2: emptySchemaType,
			},
			want:  schema,
			want1: true,
		},
		{
			name: "both empty",
			args: args{
				s:  emptySchemaType,
				s2: emptySchemaType,
			},
			want:  emptySchemaType,
			want1: true,
		},
		{
			name: "both not empty",
			args: args{
				s:  schema,
				s2: schema,
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := shouldReturnIfEmptySchemaType(tt.args.s, tt.args.s2)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shouldReturnIfEmptySchemaType() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("shouldReturnIfEmptySchemaType() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_shouldReturnIfEmptyParameters(t *testing.T) {
	var emptyParameters spec.Parameters
	parameters := spec.Parameters{
		&spec.ParameterRef{Value: spec.NewHeaderParameter("test")},
	}
	type args struct {
		parameters  spec.Parameters
		parameters2 spec.Parameters
	}
	tests := []struct {
		name  string
		args  args
		want  spec.Parameters
		want1 bool
	}{
		{
			name: "first is empty second is not",
			args: args{
				parameters:  emptyParameters,
				parameters2: parameters,
			},
			want:  parameters,
			want1: true,
		},
		{
			name: "second is empty first is not",
			args: args{
				parameters:  parameters,
				parameters2: emptyParameters,
			},
			want:  parameters,
			want1: true,
		},
		{
			name: "both empty",
			args: args{
				parameters:  emptyParameters,
				parameters2: emptyParameters,
			},
			want:  emptyParameters,
			want1: true,
		},
		{
			name: "both not empty",
			args: args{
				parameters:  parameters,
				parameters2: parameters,
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := shouldReturnIfEmptyParameters(tt.args.parameters, tt.args.parameters2)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shouldReturnIfEmptyParameters() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("shouldReturnIfEmptyParameters() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeHeader(t *testing.T) {
	type args struct {
		header  *spec.Header
		header2 *spec.Header
		child   *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  *spec.Header
		want1 []conflict
	}{
		{
			name: "nothing to merge",
			args: args{
				header: &spec.Header{
					Parameter: *spec.NewHeaderParameter("test").WithSchema(spec.NewStringSchema()),
				},
				header2: &spec.Header{
					Parameter: *spec.NewHeaderParameter("test").WithSchema(spec.NewStringSchema()),
				},
				child: nil,
			},
			want: &spec.Header{
				Parameter: *spec.NewHeaderParameter("test").WithSchema(spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "merge string type removal",
			args: args{
				header: &spec.Header{
					Parameter: *spec.NewHeaderParameter("test").WithSchema(spec.NewStringSchema()),
				},
				header2: &spec.Header{
					Parameter: *spec.NewHeaderParameter("test").WithSchema(spec.NewUUIDSchema()),
				},
				child: nil,
			},
			want: &spec.Header{
				Parameter: *spec.NewHeaderParameter("test").WithSchema(spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "header in conflicts",
			args: args{
				header: &spec.Header{
					Parameter: *spec.NewHeaderParameter("header").WithSchema(spec.NewStringSchema()),
				},
				header2: &spec.Header{
					Parameter: *spec.NewCookieParameter("cookie").WithSchema(spec.NewArraySchema()),
				},
				child: field.NewPath("test"),
			},
			want: &spec.Header{
				Parameter: *spec.NewHeaderParameter("header").WithSchema(spec.NewStringSchema()),
			},
			want1: []conflict{
				{
					path: field.NewPath("test"),
					obj1: &spec.Header{
						Parameter: *spec.NewHeaderParameter("header").WithSchema(spec.NewStringSchema()),
					},
					obj2: &spec.Header{
						Parameter: *spec.NewCookieParameter("cookie").WithSchema(spec.NewArraySchema()),
					},
					msg: createHeaderInConflictMsg(field.NewPath("test"), spec.ParameterInHeader, spec.ParameterInCookie),
				},
			},
		},
		{
			name: "type conflicts",
			args: args{
				header: &spec.Header{
					Parameter: *spec.NewHeaderParameter("test").WithSchema(spec.NewStringSchema()),
				},
				header2: &spec.Header{
					Parameter: *spec.NewHeaderParameter("test").WithSchema(spec.NewArraySchema()),
				},
				child: field.NewPath("test"),
			},
			want: &spec.Header{
				Parameter: *spec.NewHeaderParameter("test").WithSchema(spec.NewStringSchema()),
			},
			want1: []conflict{
				{
					path: field.NewPath("test"),
					obj1: spec.NewStringSchema(),
					obj2: spec.NewArraySchema(),
					msg:  createConflictMsg(field.NewPath("test"), spec.TypeString, spec.TypeArray),
				},
			},
		},
		{
			name: "empty header",
			args: args{
				header: &spec.Header{
					Parameter: *spec.NewHeaderParameter("empty"),
				},
				header2: &spec.Header{
					Parameter: *spec.NewHeaderParameter("test").WithSchema(spec.NewArraySchema()),
				},
				child: field.NewPath("test"),
			},
			want: &spec.Header{
				Parameter: *spec.NewHeaderParameter("test").WithSchema(spec.NewArraySchema()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeHeader(tt.args.header, tt.args.header2, tt.args.child)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeHeader() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeHeader() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeResponseHeader(t *testing.T) {
	type args struct {
		headers  spec.Headers
		headers2 spec.Headers
		path     *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  spec.Headers
		want1 []conflict
	}{
		{
			name: "first headers list empty",
			args: args{
				headers: spec.Headers{},
				headers2: spec.Headers{
					"test": createHeaderRef(spec.NewStringSchema()),
				},
				path: nil,
			},
			want: spec.Headers{
				"test": createHeaderRef(spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "second headers list empty",
			args: args{
				headers: spec.Headers{
					"test": createHeaderRef(spec.NewStringSchema()),
				},
				headers2: spec.Headers{},
				path:     nil,
			},
			want: spec.Headers{
				"test": createHeaderRef(spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "no common headers",
			args: args{
				headers: spec.Headers{
					"test": createHeaderRef(spec.NewStringSchema()),
				},
				headers2: spec.Headers{
					"test2": createHeaderRef(spec.NewStringSchema()),
				},
				path: nil,
			},
			want: spec.Headers{
				"test":  createHeaderRef(spec.NewStringSchema()),
				"test2": createHeaderRef(spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "merge mutual headers",
			args: args{
				headers: spec.Headers{
					"test": createHeaderRef(spec.NewStringSchema()),
				},
				headers2: spec.Headers{
					"test": createHeaderRef(spec.NewUUIDSchema()),
				},
				path: nil,
			},
			want: spec.Headers{
				"test": createHeaderRef(spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "merge mutual headers and keep non mutual",
			args: args{
				headers: spec.Headers{
					"mutual":     createHeaderRef(spec.NewStringSchema()),
					"nonmutual1": createHeaderRef(spec.NewInt64Schema()),
				},
				headers2: spec.Headers{
					"mutual":     createHeaderRef(spec.NewUUIDSchema()),
					"nonmutual2": createHeaderRef(spec.NewBoolSchema()),
				},
				path: nil,
			},
			want: spec.Headers{
				"mutual":     createHeaderRef(spec.NewStringSchema()),
				"nonmutual1": createHeaderRef(spec.NewInt64Schema()),
				"nonmutual2": createHeaderRef(spec.NewBoolSchema()),
			},
			want1: nil,
		},
		{
			name: "merge mutual headers with conflicts",
			args: args{
				headers: spec.Headers{
					"test": createHeaderRef(spec.NewStringSchema()),
				},
				headers2: spec.Headers{
					"test": createHeaderRef(spec.NewBoolSchema()),
				},
				path: field.NewPath("headers"),
			},
			want: spec.Headers{
				"test": createHeaderRef(spec.NewStringSchema()),
			},
			want1: []conflict{
				{
					path: field.NewPath("headers").Child("test"),
					obj1: spec.NewStringSchema(),
					obj2: spec.NewBoolSchema(),
					msg:  createConflictMsg(field.NewPath("headers").Child("test"), spec.TypeString, spec.TypeBoolean),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeResponseHeader(tt.args.headers, tt.args.headers2, tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeResponseHeader() got = %+v, want %+v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeResponseHeader() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func createHeaderRef(schema *spec.Schema) *spec.HeaderRef {
	return &spec.HeaderRef{
		Value: &spec.Header{
			Parameter: spec.Parameter{
				Schema: spec.NewSchemaRef("", schema),
			},
		},
	}
}

func Test_mergeResponse(t *testing.T) {
	type args struct {
		response  *spec.Response
		response2 *spec.Response
		path      *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  *spec.Response
		want1 []conflict
	}{
		{
			name: "first response is empty",
			args: args{
				response:  spec.NewResponse(),
				response2: createTestResponse().WithHeader("X-Header", spec.NewStringSchema()).Response,
				path:      nil,
			},
			want:  createTestResponse().WithHeader("X-Header", spec.NewStringSchema()).Response,
			want1: nil,
		},
		{
			name: "second response is empty",
			args: args{
				response:  createTestResponse().WithHeader("X-Header", spec.NewStringSchema()).Response,
				response2: spec.NewResponse(),
				path:      nil,
			},
			want:  createTestResponse().WithHeader("X-Header", spec.NewStringSchema()).Response,
			want1: nil,
		},
		{
			name: "merge response schema",
			args: args{
				response: createTestResponse().
					WithJSONSchema(spec.NewDateTimeSchema()).
					WithHeader("X-Header", spec.NewStringSchema()).Response,
				response2: createTestResponse().
					WithJSONSchema(spec.NewStringSchema()).
					WithHeader("X-Header", spec.NewStringSchema()).Response,
				path: nil,
			},
			want: createTestResponse().
				WithJSONSchema(spec.NewStringSchema()).
				WithHeader("X-Header", spec.NewStringSchema()).Response,
			want1: nil,
		},
		{
			name: "merge response header",
			args: args{
				response: createTestResponse().
					WithJSONSchema(spec.NewStringSchema()).
					WithHeader("X-Header", spec.NewUUIDSchema()).Response,
				response2: createTestResponse().
					WithJSONSchema(spec.NewStringSchema()).
					WithHeader("X-Header", spec.NewStringSchema()).Response,
				path: nil,
			},
			want: createTestResponse().
				WithJSONSchema(spec.NewStringSchema()).
				WithHeader("X-Header", spec.NewStringSchema()).Response,
			want1: nil,
		},
		{
			name: "merge response header and schema",
			args: args{
				response: createTestResponse().
					WithJSONSchema(spec.NewDateTimeSchema()).
					WithHeader("X-Header", spec.NewUUIDSchema()).Response,
				response2: createTestResponse().
					WithJSONSchema(spec.NewStringSchema()).
					WithHeader("X-Header", spec.NewStringSchema()).Response,
				path: nil,
			},
			want: createTestResponse().
				WithJSONSchema(spec.NewStringSchema()).
				WithHeader("X-Header", spec.NewStringSchema()).Response,
			want1: nil,
		},
		{
			name: "merge response header and schema with conflicts",
			args: args{
				response: createTestResponse().
					WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())).
					WithHeader("X-Header", spec.NewUUIDSchema()).Response,
				response2: createTestResponse().
					WithJSONSchema(spec.NewStringSchema()).
					WithHeader("X-Header", spec.NewBoolSchema()).Response,
				path: field.NewPath("200"),
			},
			want: createTestResponse().
				WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())).
				WithHeader("X-Header", spec.NewUUIDSchema()).Response,
			want1: []conflict{
				{
					path: field.NewPath("200").Child("content").Child("application/json"),
					obj1: spec.NewArraySchema().WithItems(spec.NewStringSchema()),
					obj2: spec.NewStringSchema(),
					msg: createConflictMsg(field.NewPath("200").Child("content").Child("application/json"),
						spec.TypeArray, spec.TypeString),
				},
				{
					path: field.NewPath("200").Child("headers").Child("X-Header"),
					obj1: spec.NewUUIDSchema(),
					obj2: spec.NewBoolSchema(),
					msg: createConflictMsg(field.NewPath("200").Child("headers").Child("X-Header"),
						spec.TypeString, spec.TypeBoolean),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeResponse(tt.args.response, tt.args.response2, tt.args.path)
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreUnexported(spec.Schema{}))
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeResponse() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeResponses(t *testing.T) {
	type args struct {
		responses  spec.Responses
		responses2 spec.Responses
		path       *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  spec.Responses
		want1 []conflict
	}{
		{
			name: "first is nil",
			args: args{
				responses: nil,
				responses2: createTestResponses().
					WithResponse("200", createTestResponse().
						WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())).
						WithHeader("X-Header", spec.NewUUIDSchema()).Response).Responses,
				path: nil,
			},
			want: createTestResponses().
				WithResponse("200", createTestResponse().
					WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())).
					WithHeader("X-Header", spec.NewUUIDSchema()).Response).Responses,
			want1: nil,
		},
		{
			name: "second is nil",
			args: args{
				responses: createTestResponses().
					WithResponse("200", createTestResponse().
						WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())).
						WithHeader("X-Header", spec.NewUUIDSchema()).Response).Responses,
				responses2: nil,
				path:       nil,
			},
			want: createTestResponses().
				WithResponse("200", createTestResponse().
					WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())).
					WithHeader("X-Header", spec.NewUUIDSchema()).Response).Responses,
			want1: nil,
		},
		{
			name: "both are nil",
			args: args{
				responses:  nil,
				responses2: nil,
				path:       nil,
			},
			want:  nil,
			want1: nil,
		},
		{
			name: "non mutual response code responses",
			args: args{
				responses: createTestResponses().
					WithResponse("200", createTestResponse().
						WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())).
						WithHeader("X-Header", spec.NewUUIDSchema()).Response).Responses,
				responses2: createTestResponses().
					WithResponse("201", createTestResponse().
						WithJSONSchema(spec.NewStringSchema()).
						WithHeader("X-Header2", spec.NewUUIDSchema()).Response).Responses,
				path: nil,
			},
			want: createTestResponses().
				WithResponse("200", createTestResponse().
					WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())).
					WithHeader("X-Header", spec.NewUUIDSchema()).Response).
				WithResponse("201", createTestResponse().
					WithJSONSchema(spec.NewStringSchema()).
					WithHeader("X-Header2", spec.NewUUIDSchema()).Response).Responses,
			want1: nil,
		},
		{
			name: "mutual response code responses",
			args: args{
				responses: createTestResponses().
					WithResponse("200", createTestResponse().
						WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewDateTimeSchema())).
						WithHeader("X-Header", spec.NewUUIDSchema()).Response).Responses,
				responses2: createTestResponses().
					WithResponse("200", createTestResponse().
						WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())).
						WithHeader("X-Header", spec.NewUUIDSchema()).Response).Responses,
				path: nil,
			},
			want: createTestResponses().
				WithResponse("200", createTestResponse().
					WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())).
					WithHeader("X-Header", spec.NewUUIDSchema()).Response).Responses,
			want1: nil,
		},
		{
			name: "mutual and non mutual response code responses",
			args: args{
				responses: createTestResponses().
					WithResponse("200", createTestResponse().
						WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewDateTimeSchema())).
						WithHeader("X-Header", spec.NewUUIDSchema()).Response).
					WithResponse("201", createTestResponse().
						WithJSONSchema(spec.NewDateTimeSchema()).
						WithHeader("X-Header1", spec.NewUUIDSchema()).Response).Responses,
				responses2: createTestResponses().
					WithResponse("200", createTestResponse().
						WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())).
						WithHeader("X-Header", spec.NewUUIDSchema()).Response).
					WithResponse("202", createTestResponse().
						WithJSONSchema(spec.NewBoolSchema()).
						WithHeader("X-Header3", spec.NewUUIDSchema()).Response).Responses,
				path: nil,
			},
			want: createTestResponses().
				WithResponse("200", createTestResponse().
					WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())).
					WithHeader("X-Header", spec.NewUUIDSchema()).Response).
				WithResponse("201", createTestResponse().
					WithJSONSchema(spec.NewDateTimeSchema()).
					WithHeader("X-Header1", spec.NewUUIDSchema()).Response).
				WithResponse("202", createTestResponse().
					WithJSONSchema(spec.NewBoolSchema()).
					WithHeader("X-Header3", spec.NewUUIDSchema()).Response).Responses,
			want1: nil,
		},
		{
			name: "mutual and non mutual response code responses with conflicts",
			args: args{
				responses: createTestResponses().
					WithResponse("200", createTestResponse().
						WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewArraySchema().WithItems(spec.NewDateTimeSchema()))).
						WithHeader("X-Header", spec.NewUUIDSchema()).Response).
					WithResponse("201", createTestResponse().
						WithJSONSchema(spec.NewDateTimeSchema()).
						WithHeader("X-Header1", spec.NewUUIDSchema()).Response).Responses,
				responses2: createTestResponses().
					WithResponse("200", createTestResponse().
						WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())).
						WithHeader("X-Header", spec.NewUUIDSchema()).Response).
					WithResponse("202", createTestResponse().
						WithJSONSchema(spec.NewBoolSchema()).
						WithHeader("X-Header3", spec.NewUUIDSchema()).Response).Responses,
				path: field.NewPath("responses"),
			},
			want: createTestResponses().
				WithResponse("200", createTestResponse().
					WithJSONSchema(spec.NewArraySchema().WithItems(spec.NewArraySchema().WithItems(spec.NewDateTimeSchema()))).
					WithHeader("X-Header", spec.NewUUIDSchema()).Response).
				WithResponse("201", createTestResponse().
					WithJSONSchema(spec.NewDateTimeSchema()).
					WithHeader("X-Header1", spec.NewUUIDSchema()).Response).
				WithResponse("202", createTestResponse().
					WithJSONSchema(spec.NewBoolSchema()).
					WithHeader("X-Header3", spec.NewUUIDSchema()).Response).Responses,
			want1: []conflict{
				{
					path: field.NewPath("responses").Child("200").Child("content").
						Child("application/json").Child("items"),
					obj1: spec.NewArraySchema().WithItems(spec.NewDateTimeSchema()),
					obj2: spec.NewStringSchema(),
					msg: createConflictMsg(field.NewPath("responses").Child("200").Child("content").
						Child("application/json").Child("items"), spec.TypeArray, spec.TypeString),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeResponses(tt.args.responses, tt.args.responses2, tt.args.path)
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreUnexported(spec.Schema{}))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeResponses() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeResponses() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeProperties(t *testing.T) {
	type args struct {
		properties  spec.Schemas
		properties2 spec.Schemas
		path        *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  spec.Schemas
		want1 []conflict
	}{
		{
			name: "first is nil",
			args: args{
				properties: nil,
				properties2: spec.Schemas{
					"string-key": spec.NewSchemaRef("", spec.NewStringSchema()),
				},
				path: nil,
			},
			want: spec.Schemas{
				"string-key": spec.NewSchemaRef("", spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "second is nil",
			args: args{
				properties: spec.Schemas{
					"string-key": spec.NewSchemaRef("", spec.NewStringSchema()),
				},
				properties2: nil,
				path:        nil,
			},
			want: spec.Schemas{
				"string-key": spec.NewSchemaRef("", spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "both are nil",
			args: args{
				properties:  nil,
				properties2: nil,
				path:        nil,
			},
			want:  make(spec.Schemas),
			want1: nil,
		},
		{
			name: "non mutual properties",
			args: args{
				properties: spec.Schemas{
					"string-key": spec.NewSchemaRef("", spec.NewStringSchema()),
				},
				properties2: spec.Schemas{
					"bool-key": spec.NewSchemaRef("", spec.NewBoolSchema()),
				},
				path: nil,
			},
			want: spec.Schemas{
				"string-key": spec.NewSchemaRef("", spec.NewStringSchema()),
				"bool-key":   spec.NewSchemaRef("", spec.NewBoolSchema()),
			},
			want1: nil,
		},
		{
			name: "mutual properties",
			args: args{
				properties: spec.Schemas{
					"string-key": spec.NewSchemaRef("", spec.NewStringSchema()),
				},
				properties2: spec.Schemas{
					"string-key": spec.NewSchemaRef("", spec.NewStringSchema()),
				},
				path: nil,
			},
			want: spec.Schemas{
				"string-key": spec.NewSchemaRef("", spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "mutual and non mutual properties",
			args: args{
				properties: spec.Schemas{
					"string-key": spec.NewSchemaRef("", spec.NewStringSchema()),
					"bool-key":   spec.NewSchemaRef("", spec.NewBoolSchema()),
				},
				properties2: spec.Schemas{
					"string-key": spec.NewSchemaRef("", spec.NewStringSchema()),
					"int-key":    spec.NewSchemaRef("", spec.NewInt64Schema()),
				},
				path: nil,
			},
			want: spec.Schemas{
				"string-key": spec.NewSchemaRef("", spec.NewStringSchema()),
				"int-key":    spec.NewSchemaRef("", spec.NewInt64Schema()),
				"bool-key":   spec.NewSchemaRef("", spec.NewBoolSchema()),
			},
			want1: nil,
		},
		{
			name: "mutual and non mutual response code responses with conflicts",
			args: args{
				properties: spec.Schemas{
					"conflict": spec.NewSchemaRef("", spec.NewBoolSchema()),
					"bool-key": spec.NewSchemaRef("", spec.NewBoolSchema()),
				},
				properties2: spec.Schemas{
					"conflict": spec.NewSchemaRef("", spec.NewInt64Schema()),
					"int-key":  spec.NewSchemaRef("", spec.NewInt64Schema()),
				},
				path: field.NewPath("properties"),
			},
			want: spec.Schemas{
				"conflict": spec.NewSchemaRef("", spec.NewBoolSchema()),
				"int-key":  spec.NewSchemaRef("", spec.NewInt64Schema()),
				"bool-key": spec.NewSchemaRef("", spec.NewBoolSchema()),
			},
			want1: []conflict{
				{
					path: field.NewPath("properties").Child("conflict"),
					obj1: spec.NewBoolSchema(),
					obj2: spec.NewInt64Schema(),
					msg: createConflictMsg(field.NewPath("properties").Child("conflict"),
						spec.TypeBoolean, spec.TypeInteger),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeProperties(tt.args.properties, tt.args.properties2, tt.args.path)
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreUnexported(spec.Schema{}))
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeProperties() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeSchemaItems(t *testing.T) {
	type args struct {
		items  *spec.SchemaRef
		items2 *spec.SchemaRef
		path   *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  *spec.SchemaRef
		want1 []conflict
	}{
		{
			name: "no merge needed",
			args: args{
				items:  spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewStringSchema())),
				items2: spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewStringSchema())),
				path:   field.NewPath("test"),
			},
			want:  spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewStringSchema())),
			want1: nil,
		},
		{
			name: "items with string format - format should be removed",
			args: args{
				items:  spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewStringSchema())),
				items2: spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewUUIDSchema())),
				path:   field.NewPath("test"),
			},
			want:  spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewStringSchema())),
			want1: nil,
		},
		{
			name: "different type of items",
			args: args{
				items:  spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewStringSchema())),
				items2: spec.NewSchemaRef("", spec.NewInt64Schema()),
				path:   field.NewPath("test"),
			},
			want: spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewStringSchema())),
			want1: []conflict{
				{
					path: field.NewPath("test").Child("items"),
					obj1: spec.NewArraySchema().WithItems(spec.NewStringSchema()),
					obj2: spec.NewInt64Schema(),
					msg:  createConflictMsg(field.NewPath("test").Child("items"), spec.TypeArray, spec.TypeInteger),
				},
			},
		},
		{
			name: "items2 nil items - expected to get items",
			args: args{
				items:  spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewStringSchema())),
				items2: nil,
				path:   field.NewPath("test"),
			},
			want:  spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewStringSchema())),
			want1: nil,
		},
		{
			name: "items2 nil schema - expected to get items",
			args: args{
				items:  spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewStringSchema())),
				items2: spec.NewSchemaRef("", spec.NewArraySchema().WithItems(nil)),
				path:   field.NewPath("test"),
			},
			want:  spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewStringSchema())),
			want1: nil,
		},
		{
			name: "items nil items - expected to get items2",
			args: args{
				items:  nil,
				items2: spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewStringSchema())),
				path:   field.NewPath("test"),
			},
			want:  spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewStringSchema())),
			want1: nil,
		},
		{
			name: "both schemas nil items - expected to get schema",
			args: args{
				items:  nil,
				items2: nil,
				path:   field.NewPath("test"),
			},
			want:  nil,
			want1: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeSchemaItems(tt.args.items, tt.args.items2, tt.args.path)
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreUnexported(spec.Schema{}))
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeSchemaItems() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeSchema(t *testing.T) {
	emptySchemaType := spec.NewSchema()
	type args struct {
		schema  *spec.Schema
		schema2 *spec.Schema
		path    *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  *spec.Schema
		want1 []conflict
	}{
		{
			name: "no merge needed",
			args: args{
				schema:  spec.NewInt64Schema(),
				schema2: spec.NewInt64Schema(),
				path:    nil,
			},
			want:  spec.NewInt64Schema(),
			want1: nil,
		},
		{
			name: "first is nil",
			args: args{
				schema:  nil,
				schema2: spec.NewInt64Schema(),
				path:    nil,
			},
			want:  spec.NewInt64Schema(),
			want1: nil,
		},
		{
			name: "second is nil",
			args: args{
				schema:  spec.NewInt64Schema(),
				schema2: nil,
				path:    nil,
			},
			want:  spec.NewInt64Schema(),
			want1: nil,
		},
		{
			name: "both are nil",
			args: args{
				schema:  nil,
				schema2: nil,
				path:    nil,
			},
			want:  nil,
			want1: nil,
		},
		{
			name: "first has empty schema type",
			args: args{
				schema:  emptySchemaType,
				schema2: spec.NewInt64Schema(),
				path:    nil,
			},
			want:  spec.NewInt64Schema(),
			want1: nil,
		},
		{
			name: "second has empty schema type",
			args: args{
				schema:  spec.NewInt64Schema(),
				schema2: emptySchemaType,
				path:    nil,
			},
			want:  spec.NewInt64Schema(),
			want1: nil,
		},
		{
			name: "both has empty schema type",
			args: args{
				schema:  emptySchemaType,
				schema2: emptySchemaType,
				path:    nil,
			},
			want:  emptySchemaType,
			want1: nil,
		},
		{
			name: "type conflict",
			args: args{
				schema:  spec.NewInt64Schema(),
				schema2: spec.NewBoolSchema(),
				path:    field.NewPath("schema"),
			},
			want: spec.NewInt64Schema(),
			want1: []conflict{
				{
					path: field.NewPath("schema"),
					obj1: spec.NewInt64Schema(),
					obj2: spec.NewBoolSchema(),
					msg:  createConflictMsg(field.NewPath("schema"), spec.TypeInteger, spec.TypeBoolean),
				},
			},
		},
		{
			name: "string type with different format - dismiss the format",
			args: args{
				schema:  spec.NewDateTimeSchema(),
				schema2: spec.NewUUIDSchema(),
				path:    field.NewPath("schema"),
			},
			want:  spec.NewStringSchema(),
			want1: nil,
		},
		{
			name: "array conflict",
			args: args{
				schema:  spec.NewArraySchema().WithItems(spec.NewInt64Schema()),
				schema2: spec.NewArraySchema().WithItems(spec.NewFloat64Schema()),
				path:    field.NewPath("schema"),
			},
			want: spec.NewArraySchema().WithItems(spec.NewInt64Schema()),
			want1: []conflict{
				{
					path: field.NewPath("schema").Child("items"),
					obj1: spec.NewInt64Schema(),
					obj2: spec.NewFloat64Schema(),
					msg:  createConflictMsg(field.NewPath("schema").Child("items"), spec.TypeInteger, spec.TypeNumber),
				},
			},
		},
		{
			name: "merge object with conflict",
			args: args{
				schema: spec.NewObjectSchema().
					WithProperty("bool", spec.NewBoolSchema()).
					WithProperty("conflict", spec.NewStringSchema()),
				schema2: spec.NewObjectSchema().
					WithProperty("float", spec.NewFloat64Schema()).
					WithProperty("conflict", spec.NewInt64Schema()),
				path: field.NewPath("schema"),
			},
			want: spec.NewObjectSchema().
				WithProperty("bool", spec.NewBoolSchema()).
				WithProperty("conflict", spec.NewStringSchema()).
				WithProperty("float", spec.NewFloat64Schema()),
			want1: []conflict{
				{
					path: field.NewPath("schema").Child("properties").Child("conflict"),
					obj1: spec.NewStringSchema(),
					obj2: spec.NewInt64Schema(),
					msg: createConflictMsg(field.NewPath("schema").Child("properties").Child("conflict"),
						spec.TypeString, spec.TypeInteger),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeSchema(tt.args.schema, tt.args.schema2, tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeSchema() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeSchema() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeParameter(t *testing.T) {
	type args struct {
		parameter  *spec.Parameter
		parameter2 *spec.Parameter
		path       *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  *spec.Parameter
		want1 []conflict
	}{
		{
			name: "param type conflict",
			args: args{
				parameter:  spec.NewHeaderParameter("header").WithSchema(spec.NewStringSchema()),
				parameter2: spec.NewHeaderParameter("header").WithSchema(spec.NewBoolSchema()),
				path:       field.NewPath("param-name"),
			},
			want: spec.NewHeaderParameter("header").WithSchema(spec.NewStringSchema()),
			want1: []conflict{
				{
					path: field.NewPath("param-name"),
					obj1: spec.NewHeaderParameter("header").WithSchema(spec.NewStringSchema()),
					obj2: spec.NewHeaderParameter("header").WithSchema(spec.NewBoolSchema()),
					msg:  createConflictMsg(field.NewPath("param-name"), spec.TypeString, spec.TypeBoolean),
				},
			},
		},
		{
			name: "string merge",
			args: args{
				parameter:  spec.NewHeaderParameter("header").WithSchema(spec.NewStringSchema()),
				parameter2: spec.NewHeaderParameter("header").WithSchema(spec.NewUUIDSchema()),
				path:       field.NewPath("param-name"),
			},
			want:  spec.NewHeaderParameter("header").WithSchema(spec.NewStringSchema()),
			want1: nil,
		},
		{
			name: "array merge with conflict",
			args: args{
				parameter:  spec.NewHeaderParameter("header").WithSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())),
				parameter2: spec.NewHeaderParameter("header").WithSchema(spec.NewArraySchema().WithItems(spec.NewBoolSchema())),
				path:       field.NewPath("param-name"),
			},
			want: spec.NewHeaderParameter("header").WithSchema(spec.NewArraySchema().WithItems(spec.NewStringSchema())),
			want1: []conflict{
				{
					path: field.NewPath("param-name").Child("items"),
					obj1: spec.NewStringSchema(),
					obj2: spec.NewBoolSchema(),
					msg:  createConflictMsg(field.NewPath("param-name").Child("items"), spec.TypeString, spec.TypeBoolean),
				},
			},
		},
		{
			name: "object merge",
			args: args{
				parameter:  spec.NewHeaderParameter("header").WithSchema(spec.NewObjectSchema().WithProperty("string", spec.NewStringSchema())),
				parameter2: spec.NewHeaderParameter("header").WithSchema(spec.NewObjectSchema().WithProperty("bool", spec.NewBoolSchema())),
				path:       field.NewPath("param-name"),
			},
			want: spec.NewHeaderParameter("header").WithSchema(spec.NewObjectSchema().
				WithProperty("bool", spec.NewBoolSchema()).
				WithProperty("string", spec.NewStringSchema())),
			want1: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeParameter(tt.args.parameter, tt.args.parameter2, tt.args.path)
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreUnexported(spec.Schema{}))
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeParameter() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_makeParametersMapByName(t *testing.T) {
	type args struct {
		parameters spec.Parameters
	}
	tests := []struct {
		name string
		args args
		want map[string]*spec.ParameterRef
	}{
		{
			name: "sanity",
			args: args{
				parameters: spec.Parameters{
					&spec.ParameterRef{Value: spec.NewHeaderParameter("header")},
					&spec.ParameterRef{Value: spec.NewHeaderParameter("header2")},
					&spec.ParameterRef{Value: spec.NewPathParameter("path")},
					&spec.ParameterRef{Value: spec.NewPathParameter("path2")},
				},
			},
			want: map[string]*spec.ParameterRef{
				"header":  {Value: spec.NewHeaderParameter("header")},
				"header2": {Value: spec.NewHeaderParameter("header2")},
				"path":    {Value: spec.NewPathParameter("path")},
				"path2":   {Value: spec.NewPathParameter("path2")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeParametersMapByName(tt.args.parameters); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeParametersMapByName() = %v, want %v", marshal(got), marshal(tt.want))
			}
		})
	}
}

func Test_mergeParametersByInType(t *testing.T) {
	type args struct {
		parameters  spec.Parameters
		parameters2 spec.Parameters
		path        *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  spec.Parameters
		want1 []conflict
	}{
		{
			name: "first is nil",
			args: args{
				parameters:  nil,
				parameters2: spec.Parameters{{Value: spec.NewHeaderParameter("h")}},
				path:        nil,
			},
			want:  spec.Parameters{{Value: spec.NewHeaderParameter("h")}},
			want1: nil,
		},
		{
			name: "second is nil",
			args: args{
				parameters:  spec.Parameters{{Value: spec.NewHeaderParameter("h")}},
				parameters2: nil,
				path:        nil,
			},
			want:  spec.Parameters{{Value: spec.NewHeaderParameter("h")}},
			want1: nil,
		},
		{
			name: "both are nil",
			args: args{
				parameters:  nil,
				parameters2: nil,
				path:        nil,
			},
			want:  nil,
			want1: nil,
		},
		{
			name: "non mutual parameters",
			args: args{
				parameters:  spec.Parameters{{Value: spec.NewHeaderParameter("X-Header-1")}},
				parameters2: spec.Parameters{{Value: spec.NewHeaderParameter("X-Header-2")}},
				path:        nil,
			},
			want:  spec.Parameters{{Value: spec.NewHeaderParameter("X-Header-1")}, {Value: spec.NewHeaderParameter("X-Header-2")}},
			want1: nil,
		},
		{
			name: "mutual parameters",
			args: args{
				parameters:  spec.Parameters{{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewUUIDSchema())}},
				parameters2: spec.Parameters{{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewStringSchema())}},
				path:        nil,
			},
			want:  spec.Parameters{{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewStringSchema())}},
			want1: nil,
		},
		{
			name: "mutual and non mutual parameters",
			args: args{
				parameters: spec.Parameters{
					{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewUUIDSchema())},
					{Value: spec.NewHeaderParameter("X-Header-2").WithSchema(spec.NewBoolSchema())},
				},
				parameters2: spec.Parameters{
					{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewStringSchema())},
					{Value: spec.NewHeaderParameter("X-Header-3").WithSchema(spec.NewInt64Schema())},
				},
				path: nil,
			},
			want: spec.Parameters{
				{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewStringSchema())},
				{Value: spec.NewHeaderParameter("X-Header-2").WithSchema(spec.NewBoolSchema())},
				{Value: spec.NewHeaderParameter("X-Header-3").WithSchema(spec.NewInt64Schema())},
			},
			want1: nil,
		},
		{
			name: "mutual and non mutual parameters with conflicts",
			args: args{
				parameters: spec.Parameters{
					{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewUUIDSchema())},
					{Value: spec.NewHeaderParameter("X-Header-2").WithSchema(spec.NewBoolSchema())},
				},
				parameters2: spec.Parameters{
					{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewBoolSchema())},
					{Value: spec.NewHeaderParameter("X-Header-3").WithSchema(spec.NewInt64Schema())},
				},
				path: field.NewPath("parameters"),
			},
			want: spec.Parameters{
				{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewUUIDSchema())},
				{Value: spec.NewHeaderParameter("X-Header-2").WithSchema(spec.NewBoolSchema())},
				{Value: spec.NewHeaderParameter("X-Header-3").WithSchema(spec.NewInt64Schema())},
			},
			want1: []conflict{
				{
					path: field.NewPath("parameters").Child("X-Header-1"),
					obj1: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewUUIDSchema()),
					obj2: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewBoolSchema()),
					msg: createConflictMsg(field.NewPath("parameters").Child("X-Header-1"), spec.TypeString,
						spec.TypeBoolean),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeParametersByInType(tt.args.parameters, tt.args.parameters2, tt.args.path)
			sortParam(got)
			sortParam(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeParametersByInType() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeParametersByInType() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getParametersByIn(t *testing.T) {
	type args struct {
		parameters spec.Parameters
	}
	tests := []struct {
		name string
		args args
		want map[string]spec.Parameters
	}{
		{
			name: "sanity",
			args: args{
				parameters: spec.Parameters{
					{Value: spec.NewHeaderParameter("h1")},
					{Value: spec.NewHeaderParameter("h2")},
					{Value: spec.NewPathParameter("p1")},
					{Value: spec.NewPathParameter("p2")},
					{Value: spec.NewQueryParameter("q1")},
					{Value: spec.NewQueryParameter("q2")},
					{Value: spec.NewCookieParameter("c1")},
					{Value: spec.NewCookieParameter("c2")},
					{Value: &spec.Parameter{In: "not-supported"}},
				},
			},
			want: map[string]spec.Parameters{
				spec.ParameterInCookie: {{Value: spec.NewCookieParameter("c1")}, {Value: spec.NewCookieParameter("c2")}},
				spec.ParameterInHeader: {{Value: spec.NewHeaderParameter("h1")}, {Value: spec.NewHeaderParameter("h2")}},
				spec.ParameterInQuery:  {{Value: spec.NewQueryParameter("q1")}, {Value: spec.NewQueryParameter("q2")}},
				spec.ParameterInPath:   {{Value: spec.NewPathParameter("p1")}, {Value: spec.NewPathParameter("p2")}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getParametersByIn(tt.args.parameters); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getParametersByIn() = %v, want %v", marshal(got), marshal(tt.want))
			}
		})
	}
}

func Test_mergeParameters(t *testing.T) {
	type args struct {
		parameters  spec.Parameters
		parameters2 spec.Parameters
		path        *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  spec.Parameters
		want1 []conflict
	}{
		{
			name: "first is nil",
			args: args{
				parameters:  nil,
				parameters2: spec.Parameters{{Value: spec.NewHeaderParameter("h")}},
				path:        nil,
			},
			want:  spec.Parameters{{Value: spec.NewHeaderParameter("h")}},
			want1: nil,
		},
		{
			name: "second is nil",
			args: args{
				parameters:  spec.Parameters{{Value: spec.NewHeaderParameter("h")}},
				parameters2: nil,
				path:        nil,
			},
			want:  spec.Parameters{{Value: spec.NewHeaderParameter("h")}},
			want1: nil,
		},
		{
			name: "both are nil",
			args: args{
				parameters:  nil,
				parameters2: nil,
				path:        nil,
			},
			want:  nil,
			want1: nil,
		},
		{
			name: "non mutual parameters",
			args: args{
				parameters: spec.Parameters{
					{Value: spec.NewHeaderParameter("X-Header-1")},
					{Value: spec.NewQueryParameter("query-1")},
				},
				parameters2: spec.Parameters{
					{Value: spec.NewHeaderParameter("X-Header-2")},
					{Value: spec.NewQueryParameter("query-2")},
					{Value: spec.NewHeaderParameter("header")},
				},
				path: nil,
			},
			want: spec.Parameters{
				{Value: spec.NewHeaderParameter("X-Header-1")},
				{Value: spec.NewQueryParameter("query-1")},
				{Value: spec.NewHeaderParameter("header")},
				{Value: spec.NewHeaderParameter("X-Header-2")},
				{Value: spec.NewQueryParameter("query-2")},
			},
			want1: nil,
		},
		{
			name: "mutual parameters",
			args: args{
				parameters: spec.Parameters{
					{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewUUIDSchema())},
					{Value: spec.NewQueryParameter("query-1").WithSchema(spec.NewUUIDSchema())},
					{Value: spec.NewHeaderParameter("header").WithSchema(spec.NewObjectSchema().WithProperty("str", spec.NewStringSchema()))},
				},
				parameters2: spec.Parameters{
					{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewUUIDSchema())},
					{Value: spec.NewQueryParameter("query-1").WithSchema(spec.NewUUIDSchema())},
					{Value: spec.NewHeaderParameter("header").WithSchema(spec.NewObjectSchema().WithProperty("str", spec.NewDateTimeSchema()))},
				},
				path: nil,
			},
			want: spec.Parameters{
				{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewUUIDSchema())},
				{Value: spec.NewQueryParameter("query-1").WithSchema(spec.NewUUIDSchema())},
				{Value: spec.NewHeaderParameter("header").WithSchema(spec.NewObjectSchema().WithProperty("str", spec.NewStringSchema()))},
			},
			want1: nil,
		},
		{
			name: "mutual and non mutual parameters",
			args: args{
				parameters: spec.Parameters{
					{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewUUIDSchema())},
					{Value: spec.NewQueryParameter("query-1").WithSchema(spec.NewUUIDSchema())},
					{Value: spec.NewHeaderParameter("header").WithSchema(spec.NewObjectSchema().WithProperty("str", spec.NewStringSchema()))},
					{Value: spec.NewPathParameter("non-mutual-1").WithSchema(spec.NewStringSchema())},
					{Value: spec.NewCookieParameter("non-mutual-2").WithSchema(spec.NewStringSchema())},
				},
				parameters2: spec.Parameters{
					{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewUUIDSchema())},
					{Value: spec.NewQueryParameter("query-1").WithSchema(spec.NewUUIDSchema())},
					{Value: spec.NewHeaderParameter("header").WithSchema(spec.NewObjectSchema().WithProperty("str", spec.NewDateTimeSchema()))},
					{Value: spec.NewPathParameter("non-mutual-3").WithSchema(spec.NewStringSchema())},
					{Value: spec.NewCookieParameter("non-mutual-4").WithSchema(spec.NewStringSchema())},
				},
				path: nil,
			},
			want: spec.Parameters{
				{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewUUIDSchema())},
				{Value: spec.NewQueryParameter("query-1").WithSchema(spec.NewUUIDSchema())},
				{Value: spec.NewHeaderParameter("header").WithSchema(spec.NewObjectSchema().WithProperty("str", spec.NewStringSchema()))},
				{Value: spec.NewPathParameter("non-mutual-1").WithSchema(spec.NewStringSchema())},
				{Value: spec.NewCookieParameter("non-mutual-2").WithSchema(spec.NewStringSchema())},
				{Value: spec.NewPathParameter("non-mutual-3").WithSchema(spec.NewStringSchema())},
				{Value: spec.NewCookieParameter("non-mutual-4").WithSchema(spec.NewStringSchema())},
			},
			want1: nil,
		},
		{
			name: "mutual and non mutual parameters with conflicts",
			args: args{
				parameters: spec.Parameters{
					{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewBoolSchema())},
					{Value: spec.NewQueryParameter("query-1").WithSchema(spec.NewInt64Schema())},
					{Value: spec.NewHeaderParameter("header").WithSchema(spec.NewObjectSchema().
						WithProperty("bool", spec.NewBoolSchema()))},
					{Value: spec.NewPathParameter("non-mutual-1").WithSchema(spec.NewStringSchema())},
					{Value: spec.NewCookieParameter("non-mutual-2").WithSchema(spec.NewStringSchema())},
				},
				parameters2: spec.Parameters{
					{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewUUIDSchema())},
					{Value: spec.NewQueryParameter("query-1").WithSchema(spec.NewUUIDSchema())},
					{Value: spec.NewHeaderParameter("header").WithSchema(spec.NewObjectSchema().
						WithProperty("str", spec.NewDateTimeSchema()))},
					{Value: spec.NewPathParameter("non-mutual-3").WithSchema(spec.NewStringSchema())},
					{Value: spec.NewCookieParameter("non-mutual-4").WithSchema(spec.NewStringSchema())},
				},
				path: field.NewPath("parameters"),
			},
			want: spec.Parameters{
				{Value: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewBoolSchema())},
				{Value: spec.NewQueryParameter("query-1").WithSchema(spec.NewInt64Schema())},
				{Value: spec.NewHeaderParameter("header").WithSchema(spec.NewObjectSchema().
					WithProperty("str", spec.NewDateTimeSchema()).
					WithProperty("bool", spec.NewBoolSchema()))},
				{Value: spec.NewPathParameter("non-mutual-1").WithSchema(spec.NewStringSchema())},
				{Value: spec.NewCookieParameter("non-mutual-2").WithSchema(spec.NewStringSchema())},
				{Value: spec.NewPathParameter("non-mutual-3").WithSchema(spec.NewStringSchema())},
				{Value: spec.NewCookieParameter("non-mutual-4").WithSchema(spec.NewStringSchema())},
			},
			want1: []conflict{
				{
					path: field.NewPath("parameters").Child("X-Header-1"),
					obj1: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewBoolSchema()),
					obj2: spec.NewHeaderParameter("X-Header-1").WithSchema(spec.NewUUIDSchema()),
					msg: createConflictMsg(field.NewPath("parameters").Child("X-Header-1"), spec.TypeBoolean,
						spec.TypeString),
				},
				{
					path: field.NewPath("parameters").Child("query-1"),
					obj1: spec.NewQueryParameter("query-1").WithSchema(spec.NewInt64Schema()),
					obj2: spec.NewQueryParameter("query-1").WithSchema(spec.NewUUIDSchema()),
					msg: createConflictMsg(field.NewPath("parameters").Child("query-1"), spec.TypeInteger,
						spec.TypeString),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeParameters(tt.args.parameters, tt.args.parameters2, tt.args.path)
			sortParam(got)
			sortParam(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeParameters() got = %v, want %v", marshal(got), marshal(tt.want))
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeParameters() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func sortParam(got spec.Parameters) {
	sort.Slice(got, func(i, j int) bool {
		right := got[i]
		left := got[j]
		// Sibling parameters must have unique name + in values
		return right.Value.Name+right.Value.In < left.Value.Name+left.Value.In
	})
}

func Test_appendSecurityIfNeeded(t *testing.T) {
	type args struct {
		securityMap          spec.SecurityRequirement
		mergedSecurity       spec.SecurityRequirements
		ignoreSecurityKeyMap map[string]bool
	}
	tests := []struct {
		name                     string
		args                     args
		wantMergedSecurity       spec.SecurityRequirements
		wantIgnoreSecurityKeyMap map[string]bool
	}{
		{
			name: "sanity",
			args: args{
				securityMap:          spec.SecurityRequirement{"key": {"val1", "val2"}},
				mergedSecurity:       nil,
				ignoreSecurityKeyMap: map[string]bool{},
			},
			wantMergedSecurity:       spec.SecurityRequirements{{"key": {"val1", "val2"}}},
			wantIgnoreSecurityKeyMap: map[string]bool{"key": true},
		},
		{
			name: "key should be ignored",
			args: args{
				securityMap:          spec.SecurityRequirement{"key": {"val1", "val2"}},
				mergedSecurity:       spec.SecurityRequirements{{"old-key": {}}},
				ignoreSecurityKeyMap: map[string]bool{"key": true},
			},
			wantMergedSecurity:       spec.SecurityRequirements{{"old-key": {}}},
			wantIgnoreSecurityKeyMap: map[string]bool{"key": true},
		},
		{
			name: "new key should not be ignored, old key should be ignored",
			args: args{
				securityMap:          spec.SecurityRequirement{"old-key": {}, "new key": {"val1", "val2"}},
				mergedSecurity:       spec.SecurityRequirements{{"old-key": {}}},
				ignoreSecurityKeyMap: map[string]bool{"old-key": true, "key": true},
			},
			wantMergedSecurity:       spec.SecurityRequirements{{"old-key": {}}, {"new key": {"val1", "val2"}}},
			wantIgnoreSecurityKeyMap: map[string]bool{"old-key": true, "key": true, "new key": true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := appendSecurityIfNeeded(tt.args.securityMap, tt.args.mergedSecurity, tt.args.ignoreSecurityKeyMap)
			if !reflect.DeepEqual(got, tt.wantMergedSecurity) {
				t.Errorf("appendSecurityIfNeeded() got = %v, want %v", got, tt.wantMergedSecurity)
			}
			if !reflect.DeepEqual(got1, tt.wantIgnoreSecurityKeyMap) {
				t.Errorf("appendSecurityIfNeeded() got1 = %v, want %v", got1, tt.wantIgnoreSecurityKeyMap)
			}
		})
	}
}

func Test_mergeOperationSecurity(t *testing.T) {
	type args struct {
		security  *spec.SecurityRequirements
		security2 *spec.SecurityRequirements
	}
	tests := []struct {
		name string
		args args
		want *spec.SecurityRequirements
	}{
		{
			name: "no merge is needed",
			args: args{
				security:  &spec.SecurityRequirements{{"key1": {}}, {"key2": {"val1", "val2"}}},
				security2: &spec.SecurityRequirements{{"key1": {}}, {"key2": {"val1", "val2"}}},
			},
			want: &spec.SecurityRequirements{{"key1": {}}, {"key2": {"val1", "val2"}}},
		},
		{
			name: "full merge",
			args: args{
				security:  &spec.SecurityRequirements{{"key1": {}}},
				security2: &spec.SecurityRequirements{{"key2": {"val1", "val2"}}},
			},
			want: &spec.SecurityRequirements{{"key1": {}}, {"key2": {"val1", "val2"}}},
		},
		{
			name: "second list is a sub list of the first - result should be the first list",
			args: args{
				security:  &spec.SecurityRequirements{{"key1": {}}, {"key2": {"val1", "val2"}}, {"key3": {}}},
				security2: &spec.SecurityRequirements{{"key2": {"val1", "val2"}}},
			},
			want: &spec.SecurityRequirements{{"key1": {}}, {"key2": {"val1", "val2"}}, {"key3": {}}},
		},
		{
			name: "first list is provided as an AND - output as OR",
			args: args{
				security: &spec.SecurityRequirements{
					{"key1": {} /*AND*/, "key2": {"val1", "val2"}},
				},
				security2: &spec.SecurityRequirements{{"key2": {"val1", "val2"}}},
			},
			want: &spec.SecurityRequirements{
				{"key1": {}},
				/*OR*/
				{"key2": {"val1", "val2"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeOperationSecurity(tt.args.security, tt.args.security2)
			sort.Slice(*got, func(i, j int) bool {
				_, ok := (*got)[i]["key1"]
				return ok
			})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeOperationSecurity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isEmptyRequestBody(t *testing.T) {
	nonEmptyContent := spec.NewContent()
	nonEmptyContent["test"] = spec.NewMediaType()
	type args struct {
		body *spec.RequestBodyRef
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "body == nil",
			args: args{
				body: nil,
			},
			want: true,
		},
		{
			name: "body.Value == nil",
			args: args{
				body: &spec.RequestBodyRef{Value: nil},
			},
			want: true,
		},
		{
			name: "len(body.Value.Content) == 0",
			args: args{
				body: &spec.RequestBodyRef{Value: spec.NewRequestBody().WithContent(nil)},
			},
			want: true,
		},
		{
			name: "len(body.Value.Content) == 0",
			args: args{
				body: &spec.RequestBodyRef{Value: spec.NewRequestBody().WithContent(spec.Content{})},
			},
			want: true,
		},
		{
			name: "not empty",
			args: args{
				body: &spec.RequestBodyRef{Value: spec.NewRequestBody().WithContent(nonEmptyContent)},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEmptyRequestBody(tt.args.body); got != tt.want {
				t.Errorf("isEmptyRequestBody() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_shouldReturnIfEmptyRequestBody(t *testing.T) {
	nonEmptyContent := spec.NewContent()
	nonEmptyContent["test"] = spec.NewMediaType()
	reqBody := &spec.RequestBodyRef{Value: spec.NewRequestBody().WithContent(nonEmptyContent)}
	type args struct {
		body  *spec.RequestBodyRef
		body2 *spec.RequestBodyRef
	}
	tests := []struct {
		name  string
		args  args
		want  *spec.RequestBodyRef
		want1 bool
	}{
		{
			name: "first body is nil",
			args: args{
				body:  nil,
				body2: reqBody,
			},
			want:  reqBody,
			want1: true,
		},
		{
			name: "second body is nil",
			args: args{
				body:  reqBody,
				body2: nil,
			},
			want:  reqBody,
			want1: true,
		},
		{
			name: "both bodies non nil",
			args: args{
				body:  reqBody,
				body2: reqBody,
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := shouldReturnIfEmptyRequestBody(tt.args.body, tt.args.body2)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shouldReturnIfEmptyRequestBody() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("shouldReturnIfEmptyRequestBody() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeRequestBody(t *testing.T) {
	requestBody := spec.NewRequestBody()
	requestBody.Content = spec.NewContent()
	requestBody.Content["application/json"] = spec.NewMediaType().WithSchema(spec.NewStringSchema())
	requestBody.Content["application/xml"] = spec.NewMediaType().WithSchema(spec.NewStringSchema())

	type args struct {
		body  *spec.RequestBodyRef
		body2 *spec.RequestBodyRef
		path  *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  *spec.RequestBodyRef
		want1 []conflict
	}{
		{
			name: "first is nil",
			args: args{
				body: nil,
				body2: &spec.RequestBodyRef{
					Value: spec.NewRequestBody().WithJSONSchema(spec.NewStringSchema()),
				},
				path: nil,
			},
			want: &spec.RequestBodyRef{
				Value: spec.NewRequestBody().WithJSONSchema(spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "second is nil",
			args: args{
				body: &spec.RequestBodyRef{
					Value: spec.NewRequestBody().WithJSONSchema(spec.NewStringSchema()),
				},
				body2: nil,
				path:  nil,
			},
			want: &spec.RequestBodyRef{
				Value: spec.NewRequestBody().WithJSONSchema(spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "both are nil",
			args: args{
				body:  nil,
				body2: nil,
				path:  nil,
			},
			want:  nil,
			want1: nil,
		},
		{
			name: "non mutual contents",
			args: args{
				body: &spec.RequestBodyRef{
					Value: spec.NewRequestBody().WithJSONSchema(spec.NewStringSchema()),
				},
				body2: &spec.RequestBodyRef{
					Value: spec.NewRequestBody().WithSchema(spec.NewStringSchema(), []string{"application/xml"}),
				},
				path: nil,
			},
			want: &spec.RequestBodyRef{
				Value: spec.NewRequestBody().
					WithSchema(spec.NewStringSchema(), []string{"application/json", "application/xml"}),
			},
			want1: nil,
		},
		{
			name: "mutual contents",
			args: args{
				body: &spec.RequestBodyRef{
					Value: spec.NewRequestBody().WithJSONSchema(spec.NewStringSchema()),
				},
				body2: &spec.RequestBodyRef{
					Value: spec.NewRequestBody().WithJSONSchema(spec.NewStringSchema()),
				},
				path: nil,
			},
			want: &spec.RequestBodyRef{
				Value: spec.NewRequestBody().
					WithJSONSchema(spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "mutual and non mutual contents",
			args: args{
				body: &spec.RequestBodyRef{
					Value: spec.NewRequestBody().WithJSONSchema(spec.NewStringSchema()),
				},
				body2: &spec.RequestBodyRef{
					Value: spec.NewRequestBody().WithSchema(spec.NewStringSchema(), []string{"application/xml", "application/json"}),
				},
				path: nil,
			},
			want: &spec.RequestBodyRef{
				Value: spec.NewRequestBody().
					WithSchema(spec.NewStringSchema(), []string{"application/xml", "application/json"}),
			},
			want1: nil,
		},
		{
			name: "non mutual contents with conflicts",
			args: args{
				body: &spec.RequestBodyRef{
					Value: spec.NewRequestBody().
						WithSchema(spec.NewStringSchema(), []string{"application/xml"}),
				},
				body2: &spec.RequestBodyRef{
					Value: spec.NewRequestBody().
						WithSchema(spec.NewInt64Schema(), []string{"application/xml"}),
				},
				path: field.NewPath("requestBody"),
			},
			want: &spec.RequestBodyRef{
				Value: spec.NewRequestBody().
					WithSchema(spec.NewStringSchema(), []string{"application/xml"}),
			},
			want1: []conflict{
				{
					path: field.NewPath("requestBody").Child("content").Child("application/xml"),
					obj1: spec.NewStringSchema(),
					obj2: spec.NewInt64Schema(),
					msg: createConflictMsg(field.NewPath("requestBody").Child("content").Child("application/xml"),
						spec.TypeString, spec.TypeInteger),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeRequestBody(tt.args.body, tt.args.body2, tt.args.path)
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreUnexported(spec.Schema{}), cmpopts.IgnoreTypes(spec.ExtensionProps{}))
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeRequestBody() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeContent(t *testing.T) {
	type args struct {
		content  spec.Content
		content2 spec.Content
		path     *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  spec.Content
		want1 []conflict
	}{
		{
			name: "first is nil",
			args: args{
				content: nil,
				content2: spec.Content{
					"json": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
				},
				path: nil,
			},
			want: spec.Content{
				"json": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "second is nil",
			args: args{
				content: spec.Content{
					"json": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
				},
				content2: nil,
				path:     nil,
			},
			want: spec.Content{
				"json": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "both are nil",
			args: args{
				content:  nil,
				content2: nil,
				path:     nil,
			},
			want:  spec.NewContent(),
			want1: nil,
		},
		{
			name: "non mutual contents",
			args: args{
				content: spec.Content{
					"json": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
				},
				content2: spec.Content{
					"xml": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
				},
				path: nil,
			},
			want: spec.Content{
				"json": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
				"xml":  spec.NewMediaType().WithSchema(spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "mutual contents",
			args: args{
				content: spec.Content{
					"json": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
				},
				content2: spec.Content{
					"json": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
				},
				path: nil,
			},
			want: spec.Content{
				"json": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "mutual and non mutual contents",
			args: args{
				content: spec.Content{
					"json": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
					"foo":  spec.NewMediaType().WithSchema(spec.NewInt64Schema()),
				},
				content2: spec.Content{
					"json": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
					"xml":  spec.NewMediaType().WithSchema(spec.NewStringSchema()),
				},
				path: nil,
			},
			want: spec.Content{
				"json": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
				"foo":  spec.NewMediaType().WithSchema(spec.NewInt64Schema()),
				"xml":  spec.NewMediaType().WithSchema(spec.NewStringSchema()),
			},
			want1: nil,
		},
		{
			name: "mutual contents with conflicts",
			args: args{
				content: spec.Content{
					"xml": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
				},
				content2: spec.Content{
					"xml": spec.NewMediaType().WithSchema(spec.NewInt64Schema()),
				},
				path: field.NewPath("start"),
			},
			want: spec.Content{
				"xml": spec.NewMediaType().WithSchema(spec.NewStringSchema()),
			},
			want1: []conflict{
				{
					path: field.NewPath("start").Child("xml"),
					obj1: spec.NewStringSchema(),
					obj2: spec.NewInt64Schema(),
					msg:  createConflictMsg(field.NewPath("start").Child("xml"), spec.TypeString, spec.TypeInteger),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeContent(tt.args.content, tt.args.content2, tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeContent() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeContent() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
