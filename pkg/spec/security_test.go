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

func createOperationWithSecurity(sec *spec.SecurityRequirements) *spec.Operation {
	operation := spec.NewOperation()
	operation.Security = sec
	return operation
}

func Test_updateSecurityDefinitionsFromOperation(t *testing.T) {
	type args struct {
		securitySchemes spec.SecuritySchemes
		op              *spec.Operation
	}
	tests := []struct {
		name string
		args args
		want spec.SecuritySchemes
	}{
		{
			name: "OAuth2 OR BasicAuth",
			args: args{
				securitySchemes: spec.SecuritySchemes{},
				op: createOperationWithSecurity(&spec.SecurityRequirements{
					{
						OAuth2SecuritySchemeKey: {"admin"},
					},
					{
						BasicAuthSecuritySchemeKey: {},
					},
				}),
			},
			want: spec.SecuritySchemes{
				OAuth2SecuritySchemeKey:    &spec.SecuritySchemeRef{Value: NewOAuth2SecurityScheme(nil)},
				BasicAuthSecuritySchemeKey: &spec.SecuritySchemeRef{Value: NewBasicAuthSecurityScheme()},
			},
		},
		{
			name: "OAuth2 AND BasicAuth",
			args: args{
				securitySchemes: spec.SecuritySchemes{},
				op: createOperationWithSecurity(&spec.SecurityRequirements{
					{
						OAuth2SecuritySchemeKey:    {"admin"},
						BasicAuthSecuritySchemeKey: {},
					},
				}),
			},
			want: spec.SecuritySchemes{
				OAuth2SecuritySchemeKey:    &spec.SecuritySchemeRef{Value: NewOAuth2SecurityScheme(nil)},
				BasicAuthSecuritySchemeKey: &spec.SecuritySchemeRef{Value: NewBasicAuthSecurityScheme()},
			},
		},
		{
			name: "OAuth2 AND BasicAuth OR BasicAuth",
			args: args{
				securitySchemes: spec.SecuritySchemes{},
				op: createOperationWithSecurity(&spec.SecurityRequirements{
					{
						OAuth2SecuritySchemeKey:    {"admin"},
						BasicAuthSecuritySchemeKey: {},
					},
					{
						BasicAuthSecuritySchemeKey: {},
					},
				}),
			},
			want: spec.SecuritySchemes{
				OAuth2SecuritySchemeKey:    &spec.SecuritySchemeRef{Value: NewOAuth2SecurityScheme(nil)},
				BasicAuthSecuritySchemeKey: &spec.SecuritySchemeRef{Value: NewBasicAuthSecurityScheme()},
			},
		},
		{
			name: "Unsupported SecurityDefinition key - no change to securitySchemes",
			args: args{
				securitySchemes: spec.SecuritySchemes{
					OAuth2SecuritySchemeKey: &spec.SecuritySchemeRef{Value: NewOAuth2SecurityScheme(nil)},
				},
				op: createOperationWithSecurity(&spec.SecurityRequirements{
					{
						"unsupported": {"admin"},
					},
				}),
			},
			want: spec.SecuritySchemes{
				OAuth2SecuritySchemeKey: &spec.SecuritySchemeRef{Value: NewOAuth2SecurityScheme(nil)},
			},
		},
		{
			name: "nil operation - no change to securitySchemes",
			args: args{
				securitySchemes: spec.SecuritySchemes{
					OAuth2SecuritySchemeKey: &spec.SecuritySchemeRef{Value: NewOAuth2SecurityScheme(nil)},
				},
				op: nil,
			},
			want: spec.SecuritySchemes{
				OAuth2SecuritySchemeKey: &spec.SecuritySchemeRef{Value: NewOAuth2SecurityScheme(nil)},
			},
		},
		{
			name: "operation without security - no change to securitySchemes",
			args: args{
				securitySchemes: spec.SecuritySchemes{
					OAuth2SecuritySchemeKey: &spec.SecuritySchemeRef{Value: NewOAuth2SecurityScheme(nil)},
				},
				op: createOperationWithSecurity(nil),
			},
			want: spec.SecuritySchemes{
				OAuth2SecuritySchemeKey: &spec.SecuritySchemeRef{Value: NewOAuth2SecurityScheme(nil)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateSecuritySchemesFromOperation(tt.args.securitySchemes, tt.args.op); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateSecuritySchemesFromOperation() = %v, want %v", got, tt.want)
			}
		})
	}
}
