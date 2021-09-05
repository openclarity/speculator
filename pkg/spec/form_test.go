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
	"reflect"
	"sort"
	"testing"

	"github.com/go-openapi/spec"
)

func Test_addApplicationFormParams(t *testing.T) {
	type args struct {
		operation *spec.Operation
		sd        spec.SecurityDefinitions
		body      string
	}
	tests := []struct {
		name  string
		args  args
		want  *spec.Operation
		want1 spec.SecurityDefinitions
	}{
		{
			name: "sanity",
			args: args{
				operation: spec.NewOperation(""),
				body:      "name=Amy&fav_number=321.1",
			},
			want: spec.NewOperation("").
				AddParam(spec.FormDataParam("name").Typed(schemaTypeString, "")).
				AddParam(spec.FormDataParam("fav_number").Typed(schemaTypeNumber, "")),
		},
		{
			name: "parameters without a value",
			args: args{
				operation: spec.NewOperation(""),
				body:      "foo&bar&baz",
			},
			want: spec.NewOperation("").
				AddParam(spec.FormDataParam("foo").Typed(schemaTypeBoolean, "").AllowsEmptyValues().AsRequired()).
				AddParam(spec.FormDataParam("bar").Typed(schemaTypeBoolean, "").AllowsEmptyValues().AsRequired()).
				AddParam(spec.FormDataParam("baz").Typed(schemaTypeBoolean, "").AllowsEmptyValues().AsRequired()),
		},
		{
			name: "multiple parameter instances",
			args: args{
				operation: spec.NewOperation(""),
				body:      "param=value1&param=value2&param=value3",
			},
			want: spec.NewOperation("").
				AddParam(spec.FormDataParam("param").CollectionOf(spec.NewItems().Typed(schemaTypeString, ""), collectionFormatMulti)),
		},
		{
			name: "bad query",
			args: args{
				operation: spec.NewOperation(""),
				body:      "name%2",
			},
			want: spec.NewOperation(""),
		},
		{
			name: "OAuth2 security",
			args: args{
				operation: spec.NewOperation(""),
				body:      AccessTokenParamKey + "=token",
				sd:        map[string]*spec.SecurityScheme{},
			},
			want: spec.NewOperation("").SecuredWith(OAuth2SecurityDefinitionKey, []string{}...),
			want1: spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey: spec.OAuth2AccessToken(authorizationURL, tknURL),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, sd := addApplicationFormParams(tt.args.operation, tt.args.sd, tt.args.body)
			sort.Slice(op.Parameters, func(i, j int) bool {
				return op.Parameters[i].Name < op.Parameters[j].Name
			})
			sort.Slice(tt.want.Parameters, func(i, j int) bool {
				return tt.want.Parameters[i].Name < tt.want.Parameters[j].Name
			})
			if !reflect.DeepEqual(op, tt.want) {
				t.Errorf("addApplicationFormParams() = %v, want %v", op, tt.want)
			}
			if !reflect.DeepEqual(sd, tt.want1) {
				t.Errorf("addApplicationFormParams() got1 = %v, want %v", marshal(sd), marshal(tt.want1))
			}
		})
	}
}

var formDataBodyMultiCollection = "--cdce6441022a3dcf\r\n" +
	"Content-Disposition: form-data; name=\"integer\"\r\n\r\n" +
	"12\r\n" +
	"--cdce6441022a3dcf\r\n" +
	"Content-Disposition: form-data; name=\"integer\"\r\n\r\n" +
	"13\r\n" +
	"--cdce6441022a3dcf--\r\n"

var formDataBody = "--cdce6441022a3dcf\r\n" +
	"Content-Disposition: form-data; name=\"upfile\"; filename=\"example.txt\"\r\n" +
	"Content-Type: text/plain\r\n\r\n" +
	"File contents go here.\r\n" +
	"--cdce6441022a3dcf\r\n" +
	"Content-Disposition: form-data; name=\"array-to-ignore-expected-string\"\r\n\r\n" +
	"1,2\r\n" +
	"--cdce6441022a3dcf\r\n" +
	"Content-Disposition: form-data; name=\"string\"\r\n\r\n" +
	"str\r\n" +
	"--cdce6441022a3dcf\r\n" +
	"Content-Disposition: form-data; name=\"integer\"\r\n\r\n" +
	"12\r\n" +
	"--cdce6441022a3dcf\r\n" +
	"Content-Disposition: form-data; name=\"boolean-empty-value\"\r\n\r\n" +
	"\r\n" +
	"--cdce6441022a3dcf\r\n" +
	"Content-Disposition: form-data; name=\"boolean\"\r\n\r\n" +
	"false\r\n" +
	"--cdce6441022a3dcf--\r\n"

func Test_addMultipartFormDataParams(t *testing.T) {
	type args struct {
		operation *spec.Operation
		body      string
		params    map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    *spec.Operation
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				operation: spec.NewOperation(""),
				body:      formDataBody,
				params:    map[string]string{"boundary": "cdce6441022a3dcf"},
			},
			want: spec.NewOperation("").
				AddParam(spec.FileParam("upfile")).
				AddParam(spec.FormDataParam("integer").Typed(schemaTypeInteger, "")).
				AddParam(spec.FormDataParam("boolean").Typed(schemaTypeBoolean, "")).
				AddParam(spec.FormDataParam("string").Typed(schemaTypeString, "")).
				AddParam(spec.FormDataParam("array-to-ignore-expected-string").Typed(schemaTypeString, "")).
				AddParam(spec.FormDataParam("boolean-empty-value").Typed(schemaTypeBoolean, "").AsRequired().AllowsEmptyValues()),
			wantErr: false,
		},
		{
			name: "multi collection format",
			args: args{
				operation: spec.NewOperation(""),
				body:      formDataBodyMultiCollection,
				params:    map[string]string{"boundary": "cdce6441022a3dcf"},
			},
			want: spec.NewOperation("").
				AddParam(spec.FormDataParam("integer").CollectionOf(spec.NewItems().Typed(schemaTypeInteger, ""), collectionFormatMulti)),
			wantErr: false,
		},
		{
			name: "missing boundary param",
			args: args{
				operation: spec.NewOperation(""),
				body:      formDataBody,
				params:    map[string]string{},
			},
			want:    spec.NewOperation(""),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := addMultipartFormDataParams(tt.args.operation, tt.args.body, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("addMultipartFormDataParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				sort.Slice(got.Parameters, func(i, j int) bool {
					return got.Parameters[i].Name < got.Parameters[j].Name
				})
			}
			sort.Slice(tt.want.Parameters, func(i, j int) bool {
				return tt.want.Parameters[i].Name < tt.want.Parameters[j].Name
			})
			if !reflect.DeepEqual(got, tt.want) {
				gotB, _ := json.Marshal(got)
				wantB, _ := json.Marshal(tt.want)
				t.Errorf("addMultipartFormDataParams() got = %v, want %v", string(gotB), string(wantB))
			}
		})
	}
}
