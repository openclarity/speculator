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
	"testing"

	spec "github.com/getkin/kin-openapi/openapi3"
)

func Test_shouldIgnoreHeader(t *testing.T) {
	ignoredHeaders := map[string]struct{}{
		contentTypeHeaderName:       {},
		acceptTypeHeaderName:        {},
		authorizationTypeHeaderName: {},
	}
	type args struct {
		headerKey string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should ignore",
			args: args{
				headerKey: "Accept",
			},
			want: true,
		},
		{
			name: "should ignore",
			args: args{
				headerKey: "Content-Type",
			},
			want: true,
		},
		{
			name: "should ignore",
			args: args{
				headerKey: "Authorization",
			},
			want: true,
		},
		{
			name: "should not ignore",
			args: args{
				headerKey: "X-Test",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldIgnoreHeader(ignoredHeaders, tt.args.headerKey); got != tt.want {
				t.Errorf("shouldIgnoreHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addResponseHeader(t *testing.T) {
	op := NewOperationGenerator(OperationGeneratorConfig{})
	type args struct {
		response    *spec.Response
		headerKey   string
		headerValue string
	}
	tests := []struct {
		name string
		args args
		want *spec.Response
	}{
		{
			name: "primitive",
			args: args{
				response:    spec.NewResponse(),
				headerKey:   "X-Test-Uuid",
				headerValue: "77e1c83b-7bb0-437b-bc50-a7a58e5660ac",
			},
			want: createTestResponse().
				WithHeader("X-Test-Uuid", spec.NewUUIDSchema()).Response,
		},
		{
			name: "array",
			args: args{
				response:    spec.NewResponse(),
				headerKey:   "X-Test-Array",
				headerValue: "1,2,3,4",
			},
			want: createTestResponse().
				WithHeader("X-Test-Array", spec.NewArraySchema().WithItems(spec.NewInt64Schema())).Response,
		},
		{
			name: "date",
			args: args{
				response:    spec.NewResponse(),
				headerKey:   "date",
				headerValue: "Mon, 23 Aug 2021 06:52:48 GMT",
			},
			want: createTestResponse().
				WithHeader("date", spec.NewStringSchema()).Response,
		},
		{
			name: "ignore header",
			args: args{
				response:    spec.NewResponse(),
				headerKey:   "Accept",
				headerValue: "",
			},
			want: spec.NewResponse(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := op.addResponseHeader(tt.args.response, tt.args.headerKey, tt.args.headerValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addResponseHeader() = %v, want %v", marshal(got), marshal(tt.want))
			}
		})
	}
}

func Test_addHeaderParam(t *testing.T) {
	op := NewOperationGenerator(OperationGeneratorConfig{})
	type args struct {
		operation   *spec.Operation
		headerKey   string
		headerValue string
	}
	tests := []struct {
		name string
		args args
		want *spec.Operation
	}{
		{
			name: "primitive",
			args: args{
				operation:   spec.NewOperation(),
				headerKey:   "X-Test-Uuid",
				headerValue: "77e1c83b-7bb0-437b-bc50-a7a58e5660ac",
			},
			want: createTestOperation().WithParameter(spec.NewHeaderParameter("X-Test-Uuid").
				WithSchema(spec.NewUUIDSchema())).Op,
		},
		{
			name: "array",
			args: args{
				operation:   spec.NewOperation(),
				headerKey:   "X-Test-Array",
				headerValue: "1,2,3,4",
			},
			want: createTestOperation().WithParameter(spec.NewHeaderParameter("X-Test-Array").
				WithSchema(spec.NewArraySchema().WithItems(spec.NewInt64Schema()))).Op,
		},
		{
			name: "ignore header",
			args: args{
				operation:   spec.NewOperation(),
				headerKey:   "Accept",
				headerValue: "",
			},
			want: spec.NewOperation(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := op.addHeaderParam(tt.args.operation, tt.args.headerKey, tt.args.headerValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addHeaderParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createHeadersToIgnore(t *testing.T) {
	type args struct {
		headers []string
	}
	tests := []struct {
		name string
		args args
		want map[string]struct{}
	}{
		{
			name: "only default headers",
			args: args{
				headers: nil,
			},
			want: map[string]struct{}{
				acceptTypeHeaderName:        {},
				contentTypeHeaderName:       {},
				authorizationTypeHeaderName: {},
			},
		},
		{
			name: "with custom headers",
			args: args{
				headers: []string{
					"X-H1",
					"X-H2",
				},
			},
			want: map[string]struct{}{
				acceptTypeHeaderName:        {},
				contentTypeHeaderName:       {},
				authorizationTypeHeaderName: {},
				"x-h1":                      {},
				"x-h2":                      {},
			},
		},
		{
			name: "custom headers are sub list of the default headers",
			args: args{
				headers: []string{
					acceptTypeHeaderName,
					contentTypeHeaderName,
				},
			},
			want: map[string]struct{}{
				acceptTypeHeaderName:        {},
				contentTypeHeaderName:       {},
				authorizationTypeHeaderName: {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createHeadersToIgnore(tt.args.headers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createHeadersToIgnore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperationGenerator_addCookieParam(t *testing.T) {
	op := NewOperationGenerator(OperationGeneratorConfig{})
	type args struct {
		operation   *spec.Operation
		headerValue string
	}
	tests := []struct {
		name string
		args args
		want *spec.Operation
	}{
		{
			name: "sanity",
			args: args{
				operation:   spec.NewOperation(),
				headerValue: "debug=0; csrftoken=BUSe35dohU3O1MZvDCUOJ",
			},
			want: createTestOperation().
				WithParameter(spec.NewCookieParameter("debug").WithSchema(spec.NewInt64Schema())).
				WithParameter(spec.NewCookieParameter("csrftoken").WithSchema(spec.NewStringSchema())).
				Op,
		},
		{
			name: "array",
			args: args{
				operation:   spec.NewOperation(),
				headerValue: "array=1,2,3",
			},
			want: createTestOperation().
				WithParameter(spec.NewCookieParameter("array").WithSchema(spec.NewArraySchema().WithItems(spec.NewInt64Schema()))).
				Op,
		},
		{
			name: "unsupported cookie param",
			args: args{
				operation:   spec.NewOperation(),
				headerValue: "unsupported=unsupported=unsupported; csrftoken=BUSe35dohU3O1MZvDCUOJ",
			},
			want: createTestOperation().
				WithParameter(spec.NewCookieParameter("csrftoken").WithSchema(spec.NewStringSchema())).
				Op,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := op.addCookieParam(tt.args.operation, tt.args.headerValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addCookieParam() = %v, want %v", got, tt.want)
			}
		})
	}
}
