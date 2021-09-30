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

	"github.com/go-openapi/spec"
	"gotest.tools/assert"
	"k8s.io/utils/field"
)

func Test_merge(t *testing.T) {
	sd := spec.SecurityDefinitions{}
	op := CreateTestNewOperationGenerator()
	op1, err := op.GenerateSpecOperation(&HTTPInteractionData{
		ReqBody:     req1,
		RespBody:    res1,
		ReqHeaders:  map[string]string{"X-Test-Req-1": "1", contentTypeHeaderName: mediaTypeApplicationJSON},
		RespHeaders: map[string]string{"X-Test-Res-1": "1", contentTypeHeaderName: mediaTypeApplicationJSON},
		statusCode:  200,
	}, sd)
	assert.NilError(t, err)
	op2, err := op.GenerateSpecOperation(&HTTPInteractionData{
		ReqBody:     req2,
		RespBody:    res2,
		ReqHeaders:  map[string]string{"X-Test-Req-2": "2", contentTypeHeaderName: mediaTypeApplicationJSON},
		RespHeaders: map[string]string{"X-Test-Res-2": "2", contentTypeHeaderName: mediaTypeApplicationJSON},
		statusCode:  200,
	}, sd)
	assert.NilError(t, err)

	combinedOp, err := op.GenerateSpecOperation(&HTTPInteractionData{
		ReqBody:     combinedReq,
		RespBody:    combinedRes,
		ReqHeaders:  map[string]string{"X-Test-Req-1": "1", "X-Test-Req-2": "2", contentTypeHeaderName: mediaTypeApplicationJSON},
		RespHeaders: map[string]string{"X-Test-Res-1": "1", "X-Test-Res-2": "2", contentTypeHeaderName: mediaTypeApplicationJSON},
		statusCode:  200,
	}, sd)
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeOperation() got = %v, want %v", marshal(got), marshal(tt.want))
			}
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
				a: spec.NewOperation(""),
				b: nil,
			},
			want:  spec.NewOperation(""),
			want1: true,
		},
		{
			name: "first nil",
			args: args{
				a: nil,
				b: spec.NewOperation(""),
			},
			want:  spec.NewOperation(""),
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
				a: spec.NewOperation(""),
				b: spec.NewOperation(""),
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
	schema := spec.Schema{}
	schema.Typed("test", "")
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
	schema := &spec.Schema{}
	schema.Typed("test", "")
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
	var emptyParameters []spec.Parameter
	parameters := []spec.Parameter{
		*spec.HeaderParam("test"),
	}
	type args struct {
		parameters  []spec.Parameter
		parameters2 []spec.Parameter
	}
	tests := []struct {
		name  string
		args  args
		want  []spec.Parameter
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
		header  spec.Header
		header2 spec.Header
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
				header:  *spec.ResponseHeader().Typed(schemaTypeString, ""),
				header2: *spec.ResponseHeader().Typed(schemaTypeString, ""),
				child:   nil,
			},
			want:  spec.ResponseHeader().Typed(schemaTypeString, ""),
			want1: nil,
		},
		{
			name: "merge string type removal",
			args: args{
				header:  *spec.ResponseHeader().Typed(schemaTypeString, formatUUID),
				header2: *spec.ResponseHeader().Typed(schemaTypeString, ""),
				child:   nil,
			},
			want:  spec.ResponseHeader().Typed(schemaTypeString, ""),
			want1: nil,
		},
		{
			name: "type conflicts",
			args: args{
				header:  *spec.ResponseHeader().Typed(schemaTypeString, ""),
				header2: *spec.ResponseHeader().CollectionOf(spec.NewItems(), ""),
				child:   field.NewPath("test"),
			},
			want: spec.ResponseHeader().Typed(schemaTypeString, ""),
			want1: []conflict{
				{
					path: field.NewPath("test"),
					obj1: spec.ResponseHeader().Typed(schemaTypeString, "").SimpleSchema,
					obj2: spec.ResponseHeader().CollectionOf(spec.NewItems(), "").SimpleSchema,
					msg:  createConflictMsg(field.NewPath("test"), schemaTypeString, schemaTypeArray),
				},
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
		headers  map[string]spec.Header
		headers2 map[string]spec.Header
		path     *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  map[string]spec.Header
		want1 []conflict
	}{
		{
			name: "first headers list empty",
			args: args{
				headers: map[string]spec.Header{},
				headers2: map[string]spec.Header{
					"test": *spec.ResponseHeader().Typed(schemaTypeString, ""),
				},
				path: nil,
			},
			want: map[string]spec.Header{
				"test": *spec.ResponseHeader().Typed(schemaTypeString, ""),
			},
			want1: nil,
		},
		{
			name: "second headers list empty",
			args: args{
				headers: map[string]spec.Header{
					"test": *spec.ResponseHeader().Typed(schemaTypeString, ""),
				},
				headers2: map[string]spec.Header{},
				path:     nil,
			},
			want: map[string]spec.Header{
				"test": *spec.ResponseHeader().Typed(schemaTypeString, ""),
			},
			want1: nil,
		},
		{
			name: "no common headers",
			args: args{
				headers: map[string]spec.Header{
					"test": *spec.ResponseHeader().Typed(schemaTypeString, ""),
				},
				headers2: map[string]spec.Header{
					"test2": *spec.ResponseHeader().Typed(schemaTypeString, ""),
				},
				path: nil,
			},
			want: map[string]spec.Header{
				"test":  *spec.ResponseHeader().Typed(schemaTypeString, ""),
				"test2": *spec.ResponseHeader().Typed(schemaTypeString, ""),
			},
			want1: nil,
		},
		{
			name: "merge mutual headers",
			args: args{
				headers: map[string]spec.Header{
					"test": *spec.ResponseHeader().Typed(schemaTypeString, ""),
				},
				headers2: map[string]spec.Header{
					"test": *spec.ResponseHeader().Typed(schemaTypeString, formatUUID),
				},
				path: nil,
			},
			want: map[string]spec.Header{
				"test": *spec.ResponseHeader().Typed(schemaTypeString, ""),
			},
			want1: nil,
		},
		{
			name: "merge mutual headers and keep non mutual",
			args: args{
				headers: map[string]spec.Header{
					"mutual":     *spec.ResponseHeader().Typed(schemaTypeString, ""),
					"nonmutual1": *spec.ResponseHeader().Typed(schemaTypeInteger, ""),
				},
				headers2: map[string]spec.Header{
					"mutual":     *spec.ResponseHeader().Typed(schemaTypeString, formatUUID),
					"nonmutual2": *spec.ResponseHeader().Typed(schemaTypeBoolean, ""),
				},
				path: nil,
			},
			want: map[string]spec.Header{
				"mutual":     *spec.ResponseHeader().Typed(schemaTypeString, ""),
				"nonmutual1": *spec.ResponseHeader().Typed(schemaTypeInteger, ""),
				"nonmutual2": *spec.ResponseHeader().Typed(schemaTypeBoolean, ""),
			},
			want1: nil,
		},
		{
			name: "merge mutual headers with conflicts",
			args: args{
				headers: map[string]spec.Header{
					"test": *spec.ResponseHeader().Typed(schemaTypeString, ""),
				},
				headers2: map[string]spec.Header{
					"test": *spec.ResponseHeader().Typed(schemaTypeBoolean, ""),
				},
				path: field.NewPath("headers"),
			},
			want: map[string]spec.Header{
				"test": *spec.ResponseHeader().Typed(schemaTypeString, ""),
			},
			want1: []conflict{
				{
					path: field.NewPath("headers").Child("test"),
					obj1: spec.ResponseHeader().Typed(schemaTypeString, "").SimpleSchema,
					obj2: spec.ResponseHeader().Typed(schemaTypeBoolean, "").SimpleSchema,
					msg:  createConflictMsg(field.NewPath("headers").Child("test"), schemaTypeString, schemaTypeBoolean),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeResponseHeader(tt.args.headers, tt.args.headers2, tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeResponseHeader() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeResponseHeader() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeResponse(t *testing.T) {
	type args struct {
		response  spec.Response
		response2 spec.Response
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
				response: spec.Response{},
				response2: *spec.NewResponse().
					WithSchema(spec.StringProperty()).
					AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, "")),
				path: nil,
			},
			want: spec.NewResponse().
				WithSchema(spec.StringProperty()).
				AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, "")),
			want1: nil,
		},
		{
			name: "second response is empty",
			args: args{
				response: *spec.NewResponse().
					WithSchema(spec.StringProperty()).
					AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, "")),
				response2: spec.Response{},
				path:      nil,
			},
			want: spec.NewResponse().
				WithSchema(spec.StringProperty()).
				AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, "")),
			want1: nil,
		},
		{
			name: "merge response schema",
			args: args{
				response: *spec.NewResponse().
					WithSchema(spec.DateTimeProperty()).
					AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, "")),
				response2: *spec.NewResponse().
					WithSchema(spec.StringProperty()).
					AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, "")),
				path: nil,
			},
			want: spec.NewResponse().
				WithSchema(spec.StringProperty()).
				AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, "")),
			want1: nil,
		},
		{
			name: "merge response header",
			args: args{
				response: *spec.NewResponse().
					WithSchema(spec.StringProperty()).
					AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
				response2: *spec.NewResponse().
					WithSchema(spec.StringProperty()).
					AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, "")),
				path: nil,
			},
			want: spec.NewResponse().
				WithSchema(spec.StringProperty()).
				AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, "")),
			want1: nil,
		},
		{
			name: "merge response header and schema",
			args: args{
				response: *spec.NewResponse().
					WithSchema(spec.DateTimeProperty()).
					AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
				response2: *spec.NewResponse().
					WithSchema(spec.StringProperty()).
					AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, "")),
				path: nil,
			},
			want: spec.NewResponse().
				WithSchema(spec.StringProperty()).
				AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, "")),
			want1: nil,
		},
		{
			name: "merge response header and schema with conflicts",
			args: args{
				response: *spec.NewResponse().
					WithSchema(spec.ArrayProperty(spec.StringProperty())).
					AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
				response2: *spec.NewResponse().
					WithSchema(spec.StringProperty()).
					AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeBoolean, "")),
				path: field.NewPath("200"),
			},
			want: spec.NewResponse().
				WithSchema(spec.ArrayProperty(spec.StringProperty())).
				AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
			want1: []conflict{
				{
					path: field.NewPath("200").Child("schema"),
					obj1: spec.ArrayProperty(spec.StringProperty()),
					obj2: spec.StringProperty(),
					msg: createConflictMsg(field.NewPath("200").Child("schema"),
						schemaTypeArray, schemaTypeString),
				},
				{
					path: field.NewPath("200").Child("headers").Child("X-Header"),
					obj1: spec.ResponseHeader().Typed(schemaTypeString, formatUUID).SimpleSchema,
					obj2: spec.ResponseHeader().Typed(schemaTypeBoolean, "").SimpleSchema,
					msg: createConflictMsg(field.NewPath("200").Child("headers").Child("X-Header"),
						schemaTypeString, schemaTypeBoolean),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeResponse(tt.args.response, tt.args.response2, tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeResponse() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeResponse() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeResponses(t *testing.T) {
	type args struct {
		responses  *spec.Responses
		responses2 *spec.Responses
		path       *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  *spec.Responses
		want1 []conflict
	}{
		{
			name: "first is nil",
			args: args{
				responses: nil,
				responses2: &spec.Responses{
					ResponsesProps: spec.ResponsesProps{
						Default: defaultResponse,
						StatusCodeResponses: map[int]spec.Response{
							200: *spec.NewResponse().
								WithSchema(spec.ArrayProperty(spec.StringProperty())).
								AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						},
					},
				},
				path: nil,
			},
			want: &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					Default: defaultResponse,
					StatusCodeResponses: map[int]spec.Response{
						200: *spec.NewResponse().
							WithSchema(spec.ArrayProperty(spec.StringProperty())).
							AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
					},
				},
			},
			want1: nil,
		},
		{
			name: "second is nil",
			args: args{
				responses: &spec.Responses{
					ResponsesProps: spec.ResponsesProps{
						Default: defaultResponse,
						StatusCodeResponses: map[int]spec.Response{
							200: *spec.NewResponse().
								WithSchema(spec.ArrayProperty(spec.StringProperty())).
								AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						},
					},
				},
				responses2: nil,
				path:       nil,
			},
			want: &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					Default: defaultResponse,
					StatusCodeResponses: map[int]spec.Response{
						200: *spec.NewResponse().
							WithSchema(spec.ArrayProperty(spec.StringProperty())).
							AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
					},
				},
			},
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
				responses: &spec.Responses{
					ResponsesProps: spec.ResponsesProps{
						Default: defaultResponse,
						StatusCodeResponses: map[int]spec.Response{
							200: *spec.NewResponse().
								WithSchema(spec.ArrayProperty(spec.StringProperty())).
								AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						},
					},
				},
				responses2: &spec.Responses{
					ResponsesProps: spec.ResponsesProps{
						Default: defaultResponse,
						StatusCodeResponses: map[int]spec.Response{
							201: *spec.NewResponse().
								WithSchema(spec.StringProperty()).
								AddHeader("X-Header2", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						},
					},
				},
				path: nil,
			},
			want: &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					Default: defaultResponse,
					StatusCodeResponses: map[int]spec.Response{
						200: *spec.NewResponse().
							WithSchema(spec.ArrayProperty(spec.StringProperty())).
							AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						201: *spec.NewResponse().
							WithSchema(spec.StringProperty()).
							AddHeader("X-Header2", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
					},
				},
			},
			want1: nil,
		},
		{
			name: "mutual response code responses",
			args: args{
				responses: &spec.Responses{
					ResponsesProps: spec.ResponsesProps{
						Default: defaultResponse,
						StatusCodeResponses: map[int]spec.Response{
							200: *spec.NewResponse().
								WithSchema(spec.ArrayProperty(spec.DateTimeProperty())).
								AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						},
					},
				},
				responses2: &spec.Responses{
					ResponsesProps: spec.ResponsesProps{
						Default: defaultResponse,
						StatusCodeResponses: map[int]spec.Response{
							200: *spec.NewResponse().
								WithSchema(spec.ArrayProperty(spec.StringProperty())).
								AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						},
					},
				},
				path: nil,
			},
			want: &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					Default: defaultResponse,
					StatusCodeResponses: map[int]spec.Response{
						200: *spec.NewResponse().
							WithSchema(spec.ArrayProperty(spec.StringProperty())).
							AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
					},
				},
			},
			want1: nil,
		},
		{
			name: "mutual and non mutual response code responses",
			args: args{
				responses: &spec.Responses{
					ResponsesProps: spec.ResponsesProps{
						Default: defaultResponse,
						StatusCodeResponses: map[int]spec.Response{
							200: *spec.NewResponse().
								WithSchema(spec.ArrayProperty(spec.DateTimeProperty())).
								AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
							201: *spec.NewResponse().
								WithSchema(spec.DateTimeProperty()).
								AddHeader("X-Header1", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						},
					},
				},
				responses2: &spec.Responses{
					ResponsesProps: spec.ResponsesProps{
						Default: defaultResponse,
						StatusCodeResponses: map[int]spec.Response{
							200: *spec.NewResponse().
								WithSchema(spec.ArrayProperty(spec.StringProperty())).
								AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
							202: *spec.NewResponse().
								WithSchema(spec.BooleanProperty()).
								AddHeader("X-Header3", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						},
					},
				},
				path: nil,
			},
			want: &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					Default: defaultResponse,
					StatusCodeResponses: map[int]spec.Response{
						200: *spec.NewResponse().
							WithSchema(spec.ArrayProperty(spec.StringProperty())).
							AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						201: *spec.NewResponse().
							WithSchema(spec.DateTimeProperty()).
							AddHeader("X-Header1", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						202: *spec.NewResponse().
							WithSchema(spec.BooleanProperty()).
							AddHeader("X-Header3", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
					},
				},
			},
			want1: nil,
		},
		{
			name: "mutual and non mutual response code responses with conflicts",
			args: args{
				responses: &spec.Responses{
					ResponsesProps: spec.ResponsesProps{
						Default: defaultResponse,
						StatusCodeResponses: map[int]spec.Response{
							200: *spec.NewResponse().
								WithSchema(spec.ArrayProperty(spec.ArrayProperty(spec.DateTimeProperty()))).
								AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
							201: *spec.NewResponse().
								WithSchema(spec.DateTimeProperty()).
								AddHeader("X-Header1", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						},
					},
				},
				responses2: &spec.Responses{
					ResponsesProps: spec.ResponsesProps{
						Default: defaultResponse,
						StatusCodeResponses: map[int]spec.Response{
							200: *spec.NewResponse().
								WithSchema(spec.ArrayProperty(spec.StringProperty())).
								AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
							202: *spec.NewResponse().
								WithSchema(spec.BooleanProperty()).
								AddHeader("X-Header3", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						},
					},
				},
				path: field.NewPath("responses"),
			},
			want: &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					Default: defaultResponse,
					StatusCodeResponses: map[int]spec.Response{
						200: *spec.NewResponse().
							WithSchema(spec.ArrayProperty(spec.ArrayProperty(spec.DateTimeProperty()))).
							AddHeader("X-Header", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						201: *spec.NewResponse().
							WithSchema(spec.DateTimeProperty()).
							AddHeader("X-Header1", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
						202: *spec.NewResponse().
							WithSchema(spec.BooleanProperty()).
							AddHeader("X-Header3", spec.ResponseHeader().Typed(schemaTypeString, formatUUID)),
					},
				},
			},
			want1: []conflict{
				{
					path: field.NewPath("responses").Child("200").Child("schema").Child("items"),
					obj1: spec.ArrayProperty(spec.DateTimeProperty()),
					obj2: spec.StringProperty(),
					msg: createConflictMsg(field.NewPath("responses").Child("200").Child("schema").Child("items"),
						schemaTypeArray, schemaTypeString),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeResponses(tt.args.responses, tt.args.responses2, tt.args.path)
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
		properties  spec.SchemaProperties
		properties2 spec.SchemaProperties
		path        *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  spec.SchemaProperties
		want1 []conflict
	}{
		{
			name: "first is nil",
			args: args{
				properties: nil,
				properties2: spec.SchemaProperties{
					"string-key": *spec.StringProperty(),
				},
				path: nil,
			},
			want: spec.SchemaProperties{
				"string-key": *spec.StringProperty(),
			},
			want1: nil,
		},
		{
			name: "second is nil",
			args: args{
				properties: spec.SchemaProperties{
					"string-key": *spec.StringProperty(),
				},
				properties2: nil,
				path:        nil,
			},
			want: spec.SchemaProperties{
				"string-key": *spec.StringProperty(),
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
			want:  make(spec.SchemaProperties),
			want1: nil,
		},
		{
			name: "non mutual properties",
			args: args{
				properties: spec.SchemaProperties{
					"string-key": *spec.StringProperty(),
				},
				properties2: spec.SchemaProperties{
					"bool-key": *spec.BooleanProperty(),
				},
				path: nil,
			},
			want: spec.SchemaProperties{
				"string-key": *spec.StringProperty(),
				"bool-key":   *spec.BooleanProperty(),
			},
			want1: nil,
		},
		{
			name: "mutual properties",
			args: args{
				properties: spec.SchemaProperties{
					"string-key": *spec.StringProperty(),
				},
				properties2: spec.SchemaProperties{
					"string-key": *spec.StringProperty(),
				},
				path: nil,
			},
			want: spec.SchemaProperties{
				"string-key": *spec.StringProperty(),
			},
			want1: nil,
		},
		{
			name: "mutual and non mutual properties",
			args: args{
				properties: spec.SchemaProperties{
					"string-key": *spec.StringProperty(),
					"bool-key":   *spec.BooleanProperty(),
				},
				properties2: spec.SchemaProperties{
					"string-key": *spec.StringProperty(),
					"int-key":    *spec.Int64Property(),
				},
				path: nil,
			},
			want: spec.SchemaProperties{
				"string-key": *spec.StringProperty(),
				"int-key":    *spec.Int64Property(),
				"bool-key":   *spec.BooleanProperty(),
			},
			want1: nil,
		},
		{
			name: "mutual and non mutual response code responses with conflicts",
			args: args{
				properties: spec.SchemaProperties{
					"conflict": *spec.BooleanProperty(),
					"bool-key": *spec.BooleanProperty(),
				},
				properties2: spec.SchemaProperties{
					"conflict": *spec.Int64Property(),
					"int-key":  *spec.Int64Property(),
				},
				path: field.NewPath("properties"),
			},
			want: spec.SchemaProperties{
				"conflict": *spec.BooleanProperty(),
				"int-key":  *spec.Int64Property(),
				"bool-key": *spec.BooleanProperty(),
			},
			want1: []conflict{
				{
					path: field.NewPath("properties").Child("conflict"),
					obj1: spec.BooleanProperty(),
					obj2: spec.Int64Property(),
					msg: createConflictMsg(field.NewPath("properties").Child("conflict"),
						schemaTypeBoolean, schemaTypeInteger),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeProperties(tt.args.properties, tt.args.properties2, tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeProperties() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeProperties() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeSchemaItems(t *testing.T) {
	type args struct {
		items  *spec.SchemaOrArray
		items2 *spec.SchemaOrArray
		path   *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  *spec.SchemaOrArray
		want1 []conflict
	}{
		{
			name: "no merge needed",
			args: args{
				items:  &spec.SchemaOrArray{Schema: spec.StringProperty()},
				items2: &spec.SchemaOrArray{Schema: spec.StringProperty()},
				path:   field.NewPath("test"),
			},
			want:  &spec.SchemaOrArray{Schema: spec.StringProperty()},
			want1: nil,
		},
		{
			name: "items with string format - format should be removed",
			args: args{
				items:  &spec.SchemaOrArray{Schema: spec.StringProperty()},
				items2: &spec.SchemaOrArray{Schema: spec.StrFmtProperty(formatUUID)},
				path:   field.NewPath("test"),
			},
			want:  &spec.SchemaOrArray{Schema: spec.StringProperty()},
			want1: nil,
		},
		{
			name: "different type of items",
			args: args{
				items:  &spec.SchemaOrArray{Schema: spec.StringProperty()},
				items2: &spec.SchemaOrArray{Schema: spec.Int64Property()},
				path:   field.NewPath("test"),
			},
			want: &spec.SchemaOrArray{Schema: spec.StringProperty()},
			want1: []conflict{
				{
					path: field.NewPath("test").Child("items"),
					obj1: spec.StringProperty(),
					obj2: spec.Int64Property(),
					msg:  createConflictMsg(field.NewPath("test").Child("items"), schemaTypeString, schemaTypeInteger),
				},
			},
		},
		{
			name: "items2 nil items - expected to get items",
			args: args{
				items:  &spec.SchemaOrArray{Schema: spec.StringProperty()},
				items2: nil,
				path:   field.NewPath("test"),
			},
			want:  &spec.SchemaOrArray{Schema: spec.StringProperty()},
			want1: nil,
		},
		{
			name: "items2 nil schema - expected to get items",
			args: args{
				items:  &spec.SchemaOrArray{Schema: spec.StringProperty()},
				items2: &spec.SchemaOrArray{Schema: nil},
				path:   field.NewPath("test"),
			},
			want:  &spec.SchemaOrArray{Schema: spec.StringProperty()},
			want1: nil,
		},
		{
			name: "items nil items - expected to get items2",
			args: args{
				items:  nil,
				items2: &spec.SchemaOrArray{Schema: spec.StringProperty()},
				path:   field.NewPath("test"),
			},
			want:  &spec.SchemaOrArray{Schema: spec.StringProperty()},
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeSchemaItems() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeSchemaItems() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeSchema(t *testing.T) {
	emptySchemaType := spec.RefSchema("test")
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
				schema:  spec.Int64Property(),
				schema2: spec.Int64Property(),
				path:    nil,
			},
			want:  spec.Int64Property(),
			want1: nil,
		},
		{
			name: "first is nil",
			args: args{
				schema:  nil,
				schema2: spec.Int64Property(),
				path:    nil,
			},
			want:  spec.Int64Property(),
			want1: nil,
		},
		{
			name: "second is nil",
			args: args{
				schema:  spec.Int64Property(),
				schema2: nil,
				path:    nil,
			},
			want:  spec.Int64Property(),
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
				schema2: spec.Int64Property(),
				path:    nil,
			},
			want:  spec.Int64Property(),
			want1: nil,
		},
		{
			name: "second has empty schema type",
			args: args{
				schema:  spec.Int64Property(),
				schema2: emptySchemaType,
				path:    nil,
			},
			want:  spec.Int64Property(),
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
				schema:  spec.Int64Property(),
				schema2: spec.BooleanProperty(),
				path:    field.NewPath("schema"),
			},
			want: spec.Int64Property(),
			want1: []conflict{
				{
					path: field.NewPath("schema"),
					obj1: spec.Int64Property(),
					obj2: spec.BooleanProperty(),
					msg:  createConflictMsg(field.NewPath("schema"), schemaTypeInteger, schemaTypeBoolean),
				},
			},
		},
		{
			name: "string type with different format - dismiss the format",
			args: args{
				schema:  spec.DateTimeProperty(),
				schema2: spec.DateProperty(),
				path:    field.NewPath("schema"),
			},
			want:  spec.StringProperty(),
			want1: nil,
		},
		{
			name: "array conflict",
			args: args{
				schema:  spec.ArrayProperty(spec.Int64Property()),
				schema2: spec.ArrayProperty(spec.Float64Property()),
				path:    field.NewPath("schema"),
			},
			want: spec.ArrayProperty(spec.Int64Property()),
			want1: []conflict{
				{
					path: field.NewPath("schema").Child("items"),
					obj1: spec.Int64Property(),
					obj2: spec.Float64Property(),
					msg:  createConflictMsg(field.NewPath("schema").Child("items"), schemaTypeInteger, schemaTypeNumber),
				},
			},
		},
		{
			name: "merge object with conflict",
			args: args{
				schema: spec.MapProperty(nil).
					SetProperty("bool", *spec.BooleanProperty()).
					SetProperty("conflict", *spec.StringProperty()),
				schema2: spec.MapProperty(nil).
					SetProperty("float", *spec.Float64Property()).
					SetProperty("conflict", *spec.Int64Property()),
				path: field.NewPath("schema"),
			},
			want: spec.MapProperty(nil).
				SetProperty("bool", *spec.BooleanProperty()).
				SetProperty("conflict", *spec.StringProperty()).
				SetProperty("float", *spec.Float64Property()),
			want1: []conflict{
				{
					path: field.NewPath("schema").Child("properties").Child("conflict"),
					obj1: spec.StringProperty(),
					obj2: spec.Int64Property(),
					msg: createConflictMsg(field.NewPath("schema").Child("properties").Child("conflict"),
						schemaTypeString, schemaTypeInteger),
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

func Test_mergeSimpleSchema(t *testing.T) {
	type args struct {
		simpleSchema  spec.SimpleSchema
		simpleSchema2 spec.SimpleSchema
		path          *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  spec.SimpleSchema
		want1 []conflict
	}{
		{
			name: "type conflict",
			args: args{
				simpleSchema:  spec.NewItems().Typed(schemaTypeInteger, "").SimpleSchema,
				simpleSchema2: spec.NewItems().Typed(schemaTypeBoolean, "").SimpleSchema,
				path:          field.NewPath("items"),
			},
			want: spec.NewItems().Typed(schemaTypeInteger, "").SimpleSchema,
			want1: []conflict{
				{
					path: field.NewPath("items"),
					obj1: spec.NewItems().Typed(schemaTypeInteger, "").SimpleSchema,
					obj2: spec.NewItems().Typed(schemaTypeBoolean, "").SimpleSchema,
					msg: createConflictMsg(field.NewPath("items"),
						schemaTypeInteger, schemaTypeBoolean),
				},
			},
		},
		{
			name: "both integer",
			args: args{
				simpleSchema:  spec.NewItems().Typed(schemaTypeInteger, "").SimpleSchema,
				simpleSchema2: spec.NewItems().Typed(schemaTypeInteger, "").SimpleSchema,
				path:          field.NewPath("items"),
			},
			want:  spec.NewItems().Typed(schemaTypeInteger, "").SimpleSchema,
			want1: nil,
		},
		{
			name: "both boolean",
			args: args{
				simpleSchema:  spec.NewItems().Typed(schemaTypeBoolean, "").SimpleSchema,
				simpleSchema2: spec.NewItems().Typed(schemaTypeBoolean, "").SimpleSchema,
				path:          field.NewPath("items"),
			},
			want:  spec.NewItems().Typed(schemaTypeBoolean, "").SimpleSchema,
			want1: nil,
		},
		{
			name: "both number",
			args: args{
				simpleSchema:  spec.NewItems().Typed(schemaTypeNumber, "").SimpleSchema,
				simpleSchema2: spec.NewItems().Typed(schemaTypeNumber, "").SimpleSchema,
				path:          field.NewPath("items"),
			},
			want:  spec.NewItems().Typed(schemaTypeNumber, "").SimpleSchema,
			want1: nil,
		},
		{
			name: "string format conflict - ignore format",
			args: args{
				simpleSchema:  spec.NewItems().Typed(schemaTypeString, "uuid").SimpleSchema,
				simpleSchema2: spec.NewItems().Typed(schemaTypeString, "date").SimpleSchema,
				path:          field.NewPath("items"),
			},
			want:  spec.NewItems().Typed(schemaTypeString, "").SimpleSchema,
			want1: nil,
		},
		{
			name: "conflict array",
			args: args{
				simpleSchema: spec.NewItems().
					CollectionOf(spec.NewItems().Typed(schemaTypeString, ""), collectionFormatComma).SimpleSchema,
				simpleSchema2: spec.NewItems().
					CollectionOf(spec.NewItems().Typed(schemaTypeBoolean, ""), collectionFormatComma).SimpleSchema,
				path: field.NewPath("items"),
			},
			want: spec.NewItems().
				CollectionOf(spec.NewItems().Typed(schemaTypeString, ""), collectionFormatComma).SimpleSchema,
			want1: []conflict{
				{
					path: field.NewPath("items").Child("items"),
					obj1: spec.NewItems().Typed(schemaTypeString, "").SimpleSchema,
					obj2: spec.NewItems().Typed(schemaTypeBoolean, "").SimpleSchema,
					msg: createConflictMsg(field.NewPath("items").Child("items"),
						schemaTypeString, schemaTypeBoolean),
				},
			},
		},
		{
			name: "not supported schema type",
			args: args{
				simpleSchema: spec.SimpleSchema{
					Type: "not supported",
				},
				simpleSchema2: spec.SimpleSchema{
					Type: "not supported",
				},
				path: field.NewPath("items"),
			},
			want: spec.SimpleSchema{
				Type: "not supported",
			},
			want1: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeSimpleSchema(tt.args.simpleSchema, tt.args.simpleSchema2, tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeSimpleSchema() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeSimpleSchema() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeSimpleSchemaItems(t *testing.T) {
	itemsInItemsSimpleSchemaInteger := spec.NewItems().Typed(schemaTypeArray, "")
	itemsInItemsSimpleSchemaInteger.Items = spec.NewItems().Typed(schemaTypeInteger, "")

	itemsInItemsSimpleSchemaBoolean := spec.NewItems().Typed(schemaTypeArray, "")
	itemsInItemsSimpleSchemaBoolean.Items = spec.NewItems().Typed(schemaTypeBoolean, "")
	type args struct {
		items  *spec.Items
		items2 *spec.Items
		path   *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  *spec.Items
		want1 []conflict
	}{
		{
			name: "no merge needed",
			args: args{
				items:  spec.NewItems().Typed(schemaTypeString, ""),
				items2: spec.NewItems().Typed(schemaTypeString, ""),
				path:   field.NewPath("test"),
			},
			want:  spec.NewItems().Typed(schemaTypeString, ""),
			want1: nil,
		},
		{
			name: "array with string format - format should be removed",
			args: args{
				items:  spec.NewItems().Typed(schemaTypeString, ""),
				items2: spec.NewItems().Typed(schemaTypeString, formatUUID),
				path:   field.NewPath("test"),
			},
			want:  spec.NewItems().Typed(schemaTypeString, ""),
			want1: nil,
		},
		{
			name: "items in items",
			args: args{
				items:  itemsInItemsSimpleSchemaInteger,
				items2: itemsInItemsSimpleSchemaBoolean,
				path:   field.NewPath("test"),
			},
			want: itemsInItemsSimpleSchemaInteger,
			want1: []conflict{
				{
					path: field.NewPath("test").Child("items").Child("items"),
					obj1: spec.SimpleSchema{
						Type: schemaTypeInteger,
					},
					obj2: spec.SimpleSchema{
						Type: schemaTypeBoolean,
					},
					msg: createConflictMsg(field.NewPath("test").Child("items").Child("items"), schemaTypeInteger, schemaTypeBoolean),
				},
			},
		},
		{
			name: "different type of items",
			args: args{
				items:  spec.NewItems().Typed(schemaTypeString, ""),
				items2: spec.NewItems().Typed(schemaTypeInteger, ""),
				path:   field.NewPath("test"),
			},
			want: spec.NewItems().Typed(schemaTypeString, ""),
			want1: []conflict{
				{
					path: field.NewPath("test").Child("items"),
					obj1: spec.SimpleSchema{
						Type: schemaTypeString,
					},
					obj2: spec.SimpleSchema{
						Type: schemaTypeInteger,
					},
					msg: createConflictMsg(field.NewPath("test").Child("items"), schemaTypeString, schemaTypeInteger),
				},
			},
		},
		{
			name: "items2 nil items - expected to get items",
			args: args{
				items:  spec.NewItems().Typed(schemaTypeInteger, ""),
				items2: nil,
				path:   field.NewPath("test"),
			},
			want:  spec.NewItems().Typed(schemaTypeInteger, ""),
			want1: nil,
		},
		{
			name: "items nil items - expected to get items2",
			args: args{
				items:  nil,
				items2: spec.NewItems().Typed(schemaTypeInteger, ""),
				path:   field.NewPath("test"),
			},
			want:  spec.NewItems().Typed(schemaTypeInteger, ""),
			want1: nil,
		},
		{
			name: "both items nil items - expected to get items",
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
			got, got1 := mergeSimpleSchemaItems(tt.args.items, tt.args.items2, tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeSimpleSchemaItems() got = %v, want %v", marshal(got), marshal(tt.want))
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeSimpleSchemaItems() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeParameter(t *testing.T) {
	type args struct {
		parameter  spec.Parameter
		parameter2 spec.Parameter
		path       *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  spec.Parameter
		want1 []conflict
	}{
		{
			name: "param type conflict",
			args: args{
				parameter:  *spec.HeaderParam("header").Typed(schemaTypeString, ""),
				parameter2: *spec.HeaderParam("header").Typed(schemaTypeBoolean, ""),
				path:       field.NewPath("param-name"),
			},
			want: *spec.HeaderParam("header").Typed(schemaTypeString, ""),
			want1: []conflict{
				{
					path: field.NewPath("param-name"),
					obj1: *spec.HeaderParam("header").Typed(schemaTypeString, ""),
					obj2: *spec.HeaderParam("header").Typed(schemaTypeBoolean, ""),
					msg:  createConflictMsg(field.NewPath("param-name"), schemaTypeString, schemaTypeBoolean),
				},
			},
		},
		{
			name: "string merge",
			args: args{
				parameter:  *spec.HeaderParam("header").Typed(schemaTypeString, ""),
				parameter2: *spec.HeaderParam("header").Typed(schemaTypeString, formatUUID),
				path:       field.NewPath("param-name"),
			},
			want:  *spec.HeaderParam("header").Typed(schemaTypeString, ""),
			want1: nil,
		},
		{
			name: "array merge with conflict",
			args: args{
				parameter:  *spec.HeaderParam("header").CollectionOf(spec.NewItems().Typed(schemaTypeString, ""), collectionFormatComma),
				parameter2: *spec.HeaderParam("header").CollectionOf(spec.NewItems().Typed(schemaTypeBoolean, ""), collectionFormatComma),
				path:       field.NewPath("param-name"),
			},
			want: *spec.HeaderParam("header").CollectionOf(spec.NewItems().Typed(schemaTypeString, ""), collectionFormatComma),
			want1: []conflict{
				{
					path: field.NewPath("param-name").Child("items"),
					obj1: spec.SimpleSchema{
						Type: schemaTypeString,
					},
					obj2: spec.SimpleSchema{
						Type: schemaTypeBoolean,
					},
					msg: createConflictMsg(field.NewPath("param-name").Child("items"), schemaTypeString, schemaTypeBoolean),
				},
			},
		},
		{
			name: "object merge",
			args: args{
				parameter:  *spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).SetProperty("string", *spec.StringProperty())),
				parameter2: *spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).SetProperty("bool", *spec.BooleanProperty())),
				path:       field.NewPath("param-name"),
			},
			want: *spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).
				SetProperty("bool", *spec.BooleanProperty()).
				SetProperty("string", *spec.StringProperty())),
			want1: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeParameter(tt.args.parameter, tt.args.parameter2, tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeParameter() got = %v, want %v", marshal(got), marshal(tt.want))
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeParameter() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_makeParametersMapByName(t *testing.T) {
	type args struct {
		parameters []spec.Parameter
	}
	tests := []struct {
		name string
		args args
		want map[string]spec.Parameter
	}{
		{
			name: "sanity",
			args: args{
				parameters: []spec.Parameter{
					*spec.HeaderParam("header"),
					*spec.HeaderParam("header2"),
					*spec.PathParam("path"),
					*spec.PathParam("path2"),
				},
			},
			want: map[string]spec.Parameter{
				"header":  *spec.HeaderParam("header"),
				"header2": *spec.HeaderParam("header2"),
				"path":    *spec.PathParam("path"),
				"path2":   *spec.PathParam("path2"),
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

func Test_mergeInBodyParameters(t *testing.T) {
	type args struct {
		parameters  []spec.Parameter
		parameters2 []spec.Parameter
		path        *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  []spec.Parameter
		want1 []conflict
	}{
		{
			name: "first is empty",
			args: args{
				parameters:  nil,
				parameters2: []spec.Parameter{*spec.BodyParam(inBodyParameterName, spec.StringProperty())},
				path:        nil,
			},
			want:  []spec.Parameter{*spec.BodyParam(inBodyParameterName, spec.StringProperty())},
			want1: nil,
		},
		{
			name: "second is empty",
			args: args{
				parameters:  []spec.Parameter{*spec.BodyParam(inBodyParameterName, spec.StringProperty())},
				parameters2: nil,
				path:        nil,
			},
			want:  []spec.Parameter{*spec.BodyParam(inBodyParameterName, spec.StringProperty())},
			want1: nil,
		},
		{
			name: "both are empty",
			args: args{
				parameters:  nil,
				parameters2: nil,
				path:        nil,
			},
			want:  nil,
			want1: nil,
		},
		{
			name: "conflict schema merge",
			args: args{
				parameters:  []spec.Parameter{*spec.BodyParam(inBodyParameterName, spec.StringProperty())},
				parameters2: []spec.Parameter{*spec.BodyParam(inBodyParameterName, spec.Int64Property())},
				path:        field.NewPath("parameters"),
			},
			want: []spec.Parameter{*spec.BodyParam(inBodyParameterName, spec.StringProperty())},
			want1: []conflict{
				{
					path: field.NewPath("parameters").Child("body").Child("schema"),
					obj1: spec.StringProperty(),
					obj2: spec.Int64Property(),
					msg: createConflictMsg(field.NewPath("parameters").Child("body").Child("schema"),
						schemaTypeString, schemaTypeInteger),
				},
			},
		},
		{
			name: "schema object merge",
			args: args{
				parameters:  []spec.Parameter{*spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).SetProperty("string", *spec.StringProperty()))},
				parameters2: []spec.Parameter{*spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).SetProperty("bool", *spec.BooleanProperty()))},
				path:        field.NewPath("parameters"),
			},
			want: []spec.Parameter{*spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).
				SetProperty("string", *spec.StringProperty()).
				SetProperty("bool", *spec.BooleanProperty()))},
			want1: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := mergeInBodyParameters(tt.args.parameters, tt.args.parameters2, tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeInBodyParameters() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("mergeInBodyParameters() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_mergeParametersByInType(t *testing.T) {
	type args struct {
		parameters  []spec.Parameter
		parameters2 []spec.Parameter
		path        *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  []spec.Parameter
		want1 []conflict
	}{
		{
			name: "first is nil",
			args: args{
				parameters:  nil,
				parameters2: []spec.Parameter{*spec.HeaderParam("h")},
				path:        nil,
			},
			want:  []spec.Parameter{*spec.HeaderParam("h")},
			want1: nil,
		},
		{
			name: "second is nil",
			args: args{
				parameters:  []spec.Parameter{*spec.HeaderParam("h")},
				parameters2: nil,
				path:        nil,
			},
			want:  []spec.Parameter{*spec.HeaderParam("h")},
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
				parameters:  []spec.Parameter{*spec.HeaderParam("X-Header-1")},
				parameters2: []spec.Parameter{*spec.HeaderParam("X-Header-2")},
				path:        nil,
			},
			want:  []spec.Parameter{*spec.HeaderParam("X-Header-1"), *spec.HeaderParam("X-Header-2")},
			want1: nil,
		},
		{
			name: "mutual parameters",
			args: args{
				parameters:  []spec.Parameter{*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "uuid")},
				parameters2: []spec.Parameter{*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "")},
				path:        nil,
			},
			want:  []spec.Parameter{*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "")},
			want1: nil,
		},
		{
			name: "mutual and non mutual parameters",
			args: args{
				parameters: []spec.Parameter{
					*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "uuid"),
					*spec.HeaderParam("X-Header-2").Typed(schemaTypeBoolean, ""),
				},
				parameters2: []spec.Parameter{
					*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, ""),
					*spec.HeaderParam("X-Header-3").Typed(schemaTypeInteger, ""),
				},
				path: nil,
			},
			want: []spec.Parameter{
				*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, ""),
				*spec.HeaderParam("X-Header-2").Typed(schemaTypeBoolean, ""),
				*spec.HeaderParam("X-Header-3").Typed(schemaTypeInteger, ""),
			},
			want1: nil,
		},
		{
			name: "mutual and non mutual parameters with conflicts",
			args: args{
				parameters: []spec.Parameter{
					*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "uuid"),
					*spec.HeaderParam("X-Header-2").Typed(schemaTypeBoolean, ""),
				},
				parameters2: []spec.Parameter{
					*spec.HeaderParam("X-Header-1").Typed(schemaTypeBoolean, ""),
					*spec.HeaderParam("X-Header-3").Typed(schemaTypeInteger, ""),
				},
				path: field.NewPath("parameters"),
			},
			want: []spec.Parameter{
				*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "uuid"),
				*spec.HeaderParam("X-Header-2").Typed(schemaTypeBoolean, ""),
				*spec.HeaderParam("X-Header-3").Typed(schemaTypeInteger, ""),
			},
			want1: []conflict{
				{
					path: field.NewPath("parameters").Child("X-Header-1"),
					obj1: *spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "uuid"),
					obj2: *spec.HeaderParam("X-Header-1").Typed(schemaTypeBoolean, ""),
					msg: createConflictMsg(field.NewPath("parameters").Child("X-Header-1"), schemaTypeString,
						schemaTypeBoolean),
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
		parameters []spec.Parameter
	}
	tests := []struct {
		name string
		args args
		want map[string][]spec.Parameter
	}{
		{
			name: "sanity",
			args: args{
				parameters: []spec.Parameter{
					*spec.HeaderParam("h1"),
					*spec.HeaderParam("h2"),
					*spec.PathParam("p1"),
					*spec.PathParam("p2"),
					*spec.BodyParam("b1", nil),
					*spec.BodyParam("b2", nil),
					*spec.FormDataParam("f1"),
					*spec.FileParam("f2"),
					*spec.QueryParam("q1"),
					*spec.QueryParam("q2"),
					{ParamProps: spec.ParamProps{Name: "foo", In: "bar"}},
				},
			},
			want: map[string][]spec.Parameter{
				parametersInBody:   {*spec.BodyParam("b1", nil), *spec.BodyParam("b2", nil)},
				parametersInHeader: {*spec.HeaderParam("h1"), *spec.HeaderParam("h2")},
				parametersInQuery:  {*spec.QueryParam("q1"), *spec.QueryParam("q2")},
				parametersInForm:   {*spec.FormDataParam("f1"), *spec.FileParam("f2")},
				parametersInPath:   {*spec.PathParam("p1"), *spec.PathParam("p2")},
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
		parameters  []spec.Parameter
		parameters2 []spec.Parameter
		path        *field.Path
	}
	tests := []struct {
		name  string
		args  args
		want  []spec.Parameter
		want1 []conflict
	}{
		{
			name: "first is nil",
			args: args{
				parameters:  nil,
				parameters2: []spec.Parameter{*spec.HeaderParam("h")},
				path:        nil,
			},
			want:  []spec.Parameter{*spec.HeaderParam("h")},
			want1: nil,
		},
		{
			name: "second is nil",
			args: args{
				parameters:  []spec.Parameter{*spec.HeaderParam("h")},
				parameters2: nil,
				path:        nil,
			},
			want:  []spec.Parameter{*spec.HeaderParam("h")},
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
				parameters: []spec.Parameter{
					*spec.HeaderParam("X-Header-1"),
					*spec.QueryParam("query-1"),
				},
				parameters2: []spec.Parameter{
					*spec.HeaderParam("X-Header-2"),
					*spec.QueryParam("query-2"),
					*spec.BodyParam(inBodyParameterName, nil),
				},
				path: nil,
			},
			want: []spec.Parameter{
				*spec.HeaderParam("X-Header-1"),
				*spec.QueryParam("query-1"),
				*spec.BodyParam(inBodyParameterName, nil),
				*spec.HeaderParam("X-Header-2"),
				*spec.QueryParam("query-2"),
			},
			want1: nil,
		},
		{
			name: "mutual parameters",
			args: args{
				parameters: []spec.Parameter{
					*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "uuid"),
					*spec.QueryParam("query-1").Typed(schemaTypeString, "uuid"),
					*spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).SetProperty("str", *spec.StringProperty())),
				},
				parameters2: []spec.Parameter{
					*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "uuid"),
					*spec.QueryParam("query-1").Typed(schemaTypeString, "uuid"),
					*spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).SetProperty("str", *spec.DateTimeProperty())),
				},
				path: nil,
			},
			want: []spec.Parameter{
				*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "uuid"),
				*spec.QueryParam("query-1").Typed(schemaTypeString, "uuid"),
				*spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).SetProperty("str", *spec.StringProperty())),
			},
			want1: nil,
		},
		{
			name: "mutual and non mutual parameters",
			args: args{
				parameters: []spec.Parameter{
					*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "uuid"),
					*spec.QueryParam("query-1").Typed(schemaTypeString, "uuid"),
					*spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).SetProperty("str", *spec.StringProperty())),
					*spec.PathParam("non-mutual-1").Typed(schemaTypeString, ""),
					*spec.FormDataParam("non-mutual-2").Typed(schemaTypeString, ""),
				},
				parameters2: []spec.Parameter{
					*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "uuid"),
					*spec.QueryParam("query-1").Typed(schemaTypeString, "uuid"),
					*spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).SetProperty("str", *spec.DateTimeProperty())),
					*spec.PathParam("non-mutual-3").Typed(schemaTypeString, ""),
					*spec.FormDataParam("non-mutual-4").Typed(schemaTypeString, ""),
				},
				path: nil,
			},
			want: []spec.Parameter{
				*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "uuid"),
				*spec.QueryParam("query-1").Typed(schemaTypeString, "uuid"),
				*spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).SetProperty("str", *spec.StringProperty())),
				*spec.PathParam("non-mutual-1").Typed(schemaTypeString, ""),
				*spec.FormDataParam("non-mutual-2").Typed(schemaTypeString, ""),
				*spec.PathParam("non-mutual-3").Typed(schemaTypeString, ""),
				*spec.FormDataParam("non-mutual-4").Typed(schemaTypeString, ""),
			},
			want1: nil,
		},
		{
			name: "mutual and non mutual parameters with conflicts",
			args: args{
				parameters: []spec.Parameter{
					*spec.HeaderParam("X-Header-1").Typed(schemaTypeBoolean, ""),
					*spec.QueryParam("query-1").Typed(schemaTypeInteger, ""),
					*spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).
						SetProperty("bool", *spec.BooleanProperty())),
					*spec.PathParam("non-mutual-1").Typed(schemaTypeString, ""),
					*spec.FormDataParam("non-mutual-2").Typed(schemaTypeString, ""),
				},
				parameters2: []spec.Parameter{
					*spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "uuid"),
					*spec.QueryParam("query-1").Typed(schemaTypeString, "uuid"),
					*spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).
						SetProperty("str", *spec.DateTimeProperty())),
					*spec.PathParam("non-mutual-3").Typed(schemaTypeString, ""),
					*spec.FormDataParam("non-mutual-4").Typed(schemaTypeString, ""),
				},
				path: field.NewPath("parameters"),
			},
			want: []spec.Parameter{
				*spec.HeaderParam("X-Header-1").Typed(schemaTypeBoolean, ""),
				*spec.QueryParam("query-1").Typed(schemaTypeInteger, ""),
				*spec.BodyParam(inBodyParameterName, spec.MapProperty(nil).
					SetProperty("str", *spec.DateTimeProperty()).
					SetProperty("bool", *spec.BooleanProperty())),
				*spec.PathParam("non-mutual-1").Typed(schemaTypeString, ""),
				*spec.FormDataParam("non-mutual-2").Typed(schemaTypeString, ""),
				*spec.PathParam("non-mutual-3").Typed(schemaTypeString, ""),
				*spec.FormDataParam("non-mutual-4").Typed(schemaTypeString, ""),
			},
			want1: []conflict{
				{
					path: field.NewPath("parameters").Child("X-Header-1"),
					obj1: *spec.HeaderParam("X-Header-1").Typed(schemaTypeBoolean, ""),
					obj2: *spec.HeaderParam("X-Header-1").Typed(schemaTypeString, "uuid"),
					msg: createConflictMsg(field.NewPath("parameters").Child("X-Header-1"), schemaTypeBoolean,
						schemaTypeString),
				},
				{
					path: field.NewPath("parameters").Child("query-1"),
					obj1: *spec.QueryParam("query-1").Typed(schemaTypeInteger, ""),
					obj2: *spec.QueryParam("query-1").Typed(schemaTypeString, "uuid"),
					msg: createConflictMsg(field.NewPath("parameters").Child("query-1"), schemaTypeInteger,
						schemaTypeString),
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

func sortParam(got []spec.Parameter) {
	sort.Slice(got, func(i, j int) bool {
		right := got[i]
		left := got[j]
		// Sibling parameters must have unique name + in values
		return right.Name+right.In < left.Name+left.In
	})
}

func Test_appendSecurityIfNeeded(t *testing.T) {
	type args struct {
		securityMap          map[string][]string
		mergedSecurity       []map[string][]string
		ignoreSecurityKeyMap map[string]bool
	}
	tests := []struct {
		name                     string
		args                     args
		wantMergedSecurity       []map[string][]string
		wantIgnoreSecurityKeyMap map[string]bool
	}{
		{
			name: "sanity",
			args: args{
				securityMap:          map[string][]string{"key": {"val1", "val2"}},
				mergedSecurity:       nil,
				ignoreSecurityKeyMap: map[string]bool{},
			},
			wantMergedSecurity:       []map[string][]string{{"key": {"val1", "val2"}}},
			wantIgnoreSecurityKeyMap: map[string]bool{"key": true},
		},
		{
			name: "key should be ignored",
			args: args{
				securityMap:          map[string][]string{"key": {"val1", "val2"}},
				mergedSecurity:       []map[string][]string{{"old-key": {}}},
				ignoreSecurityKeyMap: map[string]bool{"key": true},
			},
			wantMergedSecurity:       []map[string][]string{{"old-key": {}}},
			wantIgnoreSecurityKeyMap: map[string]bool{"key": true},
		},
		{
			name: "new key should not be ignored, old key should be ignored",
			args: args{
				securityMap:          map[string][]string{"old-key": {}, "new key": {"val1", "val2"}},
				mergedSecurity:       []map[string][]string{{"old-key": {}}},
				ignoreSecurityKeyMap: map[string]bool{"old-key": true, "key": true},
			},
			wantMergedSecurity:       []map[string][]string{{"old-key": {}}, {"new key": {"val1", "val2"}}},
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
		security  []map[string][]string
		security2 []map[string][]string
	}
	tests := []struct {
		name string
		args args
		want []map[string][]string
	}{
		{
			name: "no merge is needed",
			args: args{
				security:  []map[string][]string{{"key1": {}}, {"key2": {"val1", "val2"}}},
				security2: []map[string][]string{{"key1": {}}, {"key2": {"val1", "val2"}}},
			},
			want: []map[string][]string{{"key1": {}}, {"key2": {"val1", "val2"}}},
		},
		{
			name: "full merge",
			args: args{
				security:  []map[string][]string{{"key1": {}}},
				security2: []map[string][]string{{"key2": {"val1", "val2"}}},
			},
			want: []map[string][]string{{"key1": {}}, {"key2": {"val1", "val2"}}},
		},
		{
			name: "second list is a sub list of the first - result should be the first list",
			args: args{
				security:  []map[string][]string{{"key1": {}}, {"key2": {"val1", "val2"}}, {"key3": {}}},
				security2: []map[string][]string{{"key2": {"val1", "val2"}}},
			},
			want: []map[string][]string{{"key1": {}}, {"key2": {"val1", "val2"}}, {"key3": {}}},
		},
		{
			name: "first list is provided as an AND - output as OR",
			args: args{
				security: []map[string][]string{
					{"key1": {} /*AND*/, "key2": {"val1", "val2"}},
				},
				security2: []map[string][]string{{"key2": {"val1", "val2"}}},
			},
			want: []map[string][]string{
				{"key1": {}},
				/*OR*/
				{"key2": {"val1", "val2"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeOperationSecurity(tt.args.security, tt.args.security2)
			sort.Slice(got, func(i, j int) bool {
				_, ok := got[i]["key1"]
				return ok
			})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeOperationSecurity() = %v, want %v", got, tt.want)
			}
		})
	}
}
