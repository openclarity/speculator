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
	"net/url"
	"reflect"
	"testing"

	"github.com/go-openapi/spec"
)

func Test_extractQueryParams(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    url.Values
		wantErr bool
	}{
		{
			name: "no query params",
			args: args{
				path: "path",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "no query params with ?",
			args: args{
				path: "path?",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "with query params",
			args: args{
				path: "path?foo=bar&foo=bar2",
			},
			want:    map[string][]string{"foo": {"bar", "bar2"}},
			wantErr: false,
		},
		{
			name: "invalid query params",
			args: args{
				path: "path?foo%2",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractQueryParams(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractQueryParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractQueryParams() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addQueryParam(t *testing.T) {
	type args struct {
		operation *spec.Operation
		key       string
		values    []string
	}
	tests := []struct {
		name string
		args args
		want *spec.Operation
	}{
		{
			name: "sanity",
			args: args{
				operation: spec.NewOperation(""),
				key:       "key",
				values:    []string{"val1"},
			},
			want: spec.NewOperation("").AddParam(spec.QueryParam("key").Typed("string", "")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addQueryParam(tt.args.operation, tt.args.key, tt.args.values); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addQueryParam() = %v, want %v", got, tt.want)
			}
		})
	}
}
