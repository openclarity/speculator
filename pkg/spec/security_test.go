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

func createOperationWithSecurity(sec []map[string][]string) *spec.Operation {
	operation := spec.NewOperation("")
	operation.Security = sec
	return operation
}

func Test_updateSecurityDefinitionsFromOperation(t *testing.T) {
	type args struct {
		sd spec.SecurityDefinitions
		op *spec.Operation
	}
	tests := []struct {
		name string
		args args
		want spec.SecurityDefinitions
	}{
		{
			name: "OAuth2 OR BasicAuth",
			args: args{
				sd: spec.SecurityDefinitions{},
				op: createOperationWithSecurity([]map[string][]string{
					{
						OAuth2SecurityDefinitionKey: {"admin"},
					},
					{
						BasicAuthSecurityDefinitionKey: {},
					},
				}),
			},
			want: spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey:    spec.OAuth2AccessToken(authorizationURL, tokenURL),
				BasicAuthSecurityDefinitionKey: spec.BasicAuth(),
			},
		},
		{
			name: "OAuth2 AND BasicAuth",
			args: args{
				sd: spec.SecurityDefinitions{},
				op: createOperationWithSecurity([]map[string][]string{
					{
						OAuth2SecurityDefinitionKey:    {"admin"},
						BasicAuthSecurityDefinitionKey: {},
					},
				}),
			},
			want: spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey:    spec.OAuth2AccessToken(authorizationURL, tokenURL),
				BasicAuthSecurityDefinitionKey: spec.BasicAuth(),
			},
		},
		{
			name: "OAuth2 AND BasicAuth OR BasicAuth",
			args: args{
				sd: spec.SecurityDefinitions{},
				op: createOperationWithSecurity([]map[string][]string{
					{
						OAuth2SecurityDefinitionKey:    {"admin"},
						BasicAuthSecurityDefinitionKey: {},
					},
					{
						BasicAuthSecurityDefinitionKey: {},
					},
				}),
			},
			want: spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey:    spec.OAuth2AccessToken(authorizationURL, tokenURL),
				BasicAuthSecurityDefinitionKey: spec.BasicAuth(),
			},
		},
		{
			name: "Unsupported SecurityDefinition key - no change to sd",
			args: args{
				sd: spec.SecurityDefinitions{
					OAuth2SecurityDefinitionKey: spec.OAuth2AccessToken(authorizationURL, tokenURL),
				},
				op: createOperationWithSecurity([]map[string][]string{
					{
						"unsupported": {"admin"},
					},
				}),
			},
			want: spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey: spec.OAuth2AccessToken(authorizationURL, tokenURL),
			},
		},
		{
			name: "nil operation - no change to sd",
			args: args{
				sd: spec.SecurityDefinitions{
					OAuth2SecurityDefinitionKey: spec.OAuth2AccessToken(authorizationURL, tokenURL),
				},
				op: nil,
			},
			want: spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey: spec.OAuth2AccessToken(authorizationURL, tokenURL),
			},
		},
		{
			name: "operation without security - no change to sd",
			args: args{
				sd: spec.SecurityDefinitions{
					OAuth2SecurityDefinitionKey: spec.OAuth2AccessToken(authorizationURL, tokenURL),
				},
				op: createOperationWithSecurity(nil),
			},
			want: spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey: spec.OAuth2AccessToken(authorizationURL, tokenURL),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateSecurityDefinitionsFromOperation(tt.args.sd, tt.args.op); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateSecurityDefinitionsFromOperation() = %v, want %v", got, tt.want)
			}
		})
	}
}
