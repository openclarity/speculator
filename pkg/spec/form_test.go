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
	"testing"

	spec "github.com/getkin/kin-openapi/openapi3"
)

func newBoolSchemaWithAllowEmptyValue() *spec.Schema {
	schema := spec.NewBoolSchema()
	schema.AllowEmptyValue = true
	return schema
}

func Test_handleApplicationFormURLEncodedBody(t *testing.T) {
	type args struct {
		operation       *spec.Operation
		securitySchemes spec.SecuritySchemes
		body            string
	}
	tests := []struct {
		name    string
		args    args
		want    *spec.Operation
		want1   spec.SecuritySchemes
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				operation: spec.NewOperation(),
				body:      "name=Amy&fav_number=321.1",
			},
			want: createTestOperation().WithRequestBody(spec.NewRequestBody().WithSchema(
				spec.NewObjectSchema().WithProperties(map[string]*spec.Schema{
					"name":       spec.NewStringSchema(),
					"fav_number": spec.NewFloat64Schema(),
				}), []string{mediaTypeApplicationForm})).Op,
		},
		{
			name: "parameters without a value",
			args: args{
				operation: spec.NewOperation(),
				body:      "foo&bar&baz",
			},
			want: createTestOperation().WithRequestBody(spec.NewRequestBody().WithSchema(
				spec.NewObjectSchema().WithProperties(map[string]*spec.Schema{
					"foo": newBoolSchemaWithAllowEmptyValue(),
					"bar": newBoolSchemaWithAllowEmptyValue(),
					"baz": newBoolSchemaWithAllowEmptyValue(),
				}), []string{mediaTypeApplicationForm})).Op,
		},
		{
			name: "multiple parameter instances",
			args: args{
				operation: spec.NewOperation(),
				body:      "param=value1&param=value2&param=value3",
			},
			want: createTestOperation().WithRequestBody(spec.NewRequestBody().WithSchema(
				spec.NewObjectSchema().WithProperties(map[string]*spec.Schema{
					"param": spec.NewArraySchema().WithItems(spec.NewStringSchema()),
				}), []string{mediaTypeApplicationForm})).Op,
		},
		{
			name: "bad query",
			args: args{
				operation: spec.NewOperation(),
				body:      "name%2",
			},
			want: spec.NewOperation(),
		},
		{
			name: "OAuth2 security",
			args: args{
				operation:       spec.NewOperation(),
				body:            AccessTokenParamKey + "=token",
				securitySchemes: map[string]*spec.SecuritySchemeRef{},
			},
			want: createTestOperation().WithSecurityRequirement(map[string][]string{OAuth2SecuritySchemeKey: {}}).Op,
			want1: map[string]*spec.SecuritySchemeRef{
				OAuth2SecuritySchemeKey: {Value: NewOAuth2SecurityScheme([]string{})},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, securitySchemes, err := handleApplicationFormURLEncodedBody(tt.args.operation, tt.args.securitySchemes, tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleApplicationFormURLEncodedBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			op = sortParameters(op)
			tt.want = sortParameters(tt.want)
			if !reflect.DeepEqual(op, tt.want) {
				t.Errorf("handleApplicationFormURLEncodedBody() op = %v, want %v", op, tt.want)
			}
			if !reflect.DeepEqual(securitySchemes, tt.want1) {
				t.Errorf("handleApplicationFormURLEncodedBody() securitySchemes = %v, want %v", marshal(securitySchemes), marshal(tt.want1))
			}
		})
	}
}

var formDataBodyMultipleFileUpload = "--cdce6441022a3dcf\r\n" +
	"Content-Disposition: form-data; name=\"fileName\"; filename=\"file1.txt\"\r\n\r\n" +
	"Content-Type: text/plain\r\n\r\n" +
	"File contents go here.\r\n" +
	"--cdce6441022a3dcf\r\n" +
	"Content-Disposition: form-data; name=\"fileName\"; filename=\"file2.png\"\r\n\r\n" +
	"Content-Type: image/png\r\n\r\n" +
	"File contents go here.\r\n" +
	"--cdce6441022a3dcf\r\n" +
	"Content-Disposition: form-data; name=\"fileName\"; filename=\"file3.jpg\"\r\n\r\n" +
	"Content-Type: image/jpeg\r\n\r\n" +
	"File contents go here.\r\n" +
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
		want    *spec.Schema
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				body:   formDataBody,
				params: map[string]string{"boundary": "cdce6441022a3dcf"},
			},
			want: spec.NewObjectSchema().WithProperties(map[string]*spec.Schema{
				"upfile":                          spec.NewStringSchema().WithFormat("binary"),
				"integer":                         spec.NewInt64Schema(),
				"boolean":                         spec.NewBoolSchema(),
				"string":                          spec.NewStringSchema(),
				"array-to-ignore-expected-string": spec.NewArraySchema().WithItems(spec.NewStringSchema()),
				"boolean-empty-value":             newBoolSchemaWithAllowEmptyValue(),
			}),
			wantErr: false,
		},
		{
			name: "Multiple File Upload",
			args: args{
				body:   formDataBodyMultipleFileUpload,
				params: map[string]string{"boundary": "cdce6441022a3dcf"},
			},
			want: spec.NewObjectSchema().WithProperties(map[string]*spec.Schema{
				"fileName": spec.NewArraySchema().WithItems(spec.NewStringSchema().WithFormat("binary")),
			}),
			wantErr: false,
		},
		{
			name: "missing boundary param",
			args: args{
				body:   formDataBody,
				params: map[string]string{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMultipartFormDataSchema(tt.args.body, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMultipartFormDataSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				gotB, _ := json.Marshal(got)
				wantB, _ := json.Marshal(tt.want)
				t.Errorf("getMultipartFormDataSchema() got = %v, want %v", string(gotB), string(wantB))
			}
		})
	}
}
