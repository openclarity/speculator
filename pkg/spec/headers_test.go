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

	"github.com/go-openapi/spec"
)

func Test_shouldIgnoreHeader(t *testing.T) {
	var ignoredHeaders = map[string]struct{}{
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
	op := NewOperationGenerator(&OperationGeneratorConfig{
		ResponseHeadersToIgnore:  []string{acceptTypeHeaderName},
	})
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
			want: spec.NewResponse().
				AddHeader("X-Test-Uuid", spec.ResponseHeader().Typed("string", "uuid")),
		},
		{
			name: "collection",
			args: args{
				response:    spec.NewResponse(),
				headerKey:   "X-Test-Array",
				headerValue: "1,2,3,4",
			},
			want: spec.NewResponse().
				AddHeader("X-Test-Array", spec.ResponseHeader().
					CollectionOf(spec.NewItems().Typed("integer", ""), collectionFormatComma)),
		},
		{
			name: "date",
			args: args{
				response:    spec.NewResponse(),
				headerKey:   "date",
				headerValue: "Mon, 23 Aug 2021 06:52:48 GMT",
			},
			want: spec.NewResponse().
				AddHeader("date", spec.ResponseHeader().Typed("string", "")),
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
	op := NewOperationGenerator(&OperationGeneratorConfig{
		RequestHeadersToIgnore:  []string{acceptTypeHeaderName},
	})
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
				operation:   spec.NewOperation(""),
				headerKey:   "X-Test-Uuid",
				headerValue: "77e1c83b-7bb0-437b-bc50-a7a58e5660ac",
			},
			want: spec.NewOperation("").
				AddParam(spec.HeaderParam("X-Test-Uuid").Typed("string", "uuid")),
		},
		{
			name: "collection",
			args: args{
				operation:   spec.NewOperation(""),
				headerKey:   "X-Test-Array",
				headerValue: "1,2,3,4",
			},
			want: spec.NewOperation("").AddParam(spec.HeaderParam("X-Test-Array").
				CollectionOf(spec.NewItems().Typed("integer", ""), collectionFormatComma)),
		},
		{
			name: "ignore header",
			args: args{
				operation:   spec.NewOperation(""),
				headerKey:   "Accept",
				headerValue: "",
			},
			want: spec.NewOperation(""),
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
