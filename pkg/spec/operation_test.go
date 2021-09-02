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
	"encoding/json"
	"net/url"
	"reflect"
	"testing"

	"gotest.tools/assert"

	"github.com/go-openapi/spec"
	"github.com/yudai/gojsondiff"
)

var agentStatusBody = `{"active":true,
"certificateVersion":"86eb5278-676a-3b7c-b29d-4a57007dc7be",
"controllerInstanceInfo":{"replicaId":"portshift-agent-66fc77c848-tmmk8"},
"policyAndAppVersion":1621477900361,
"statusCodes":["NO_METRICS_SERVER"],
"version":"1.147.1"}`

var cvssBody = `{"cvss":[{"score":7.8,"vector":"AV:L/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H","version":"3"}]}`

/*
  Test:
    type: 'object'
    additionalProperties:
      type: 'object'
      properties:
        code:
          type: 'integer'
        text:
          type: 'string'
*/
type additionalPropertiesTest map[string]additionalPropertiesTestAnon
type additionalPropertiesTestAnon struct {

	// code
	Code int64 `json:"code,omitempty"`

	// text
	Text string `json:"text,omitempty"`
}

type simpleMap map[string]string

func generateQueryParams(t *testing.T, query string) url.Values {
	parseQuery, err := url.ParseQuery(query)
	assert.NilError(t, err)
	return parseQuery
}

func TestGenerateSpecOperation(t *testing.T) {
	sd := spec.SecurityDefinitions{}
	operation, err := GenerateSpecOperation(&HTTPInteractionData{
		ReqBody:  agentStatusBody,
		RespBody: cvssBody,
		ReqHeaders: map[string]string{
			"X-Request-ID":        "77e1c83b-7bb0-437b-bc50-a7a58e5660ac",
			"X-Float-Test":        "12.2",
			"X-Collection-Test":   "a,b,c,d",
			contentTypeHeaderName: mediaTypeApplicationJson,
		},
		RespHeaders: map[string]string{
			"X-RateLimit-Limit":   "12",
			"X-RateLimit-Reset":   "2016-10-12T11:00:00Z",
			contentTypeHeaderName: mediaTypeApplicationJson,
		},
		QueryParams: generateQueryParams(t, "offset=30&limit=10"),
		statusCode:  200,
	}, sd)
	assert.NilError(t, err)

	t.Log(marshal(operation))
	t.Log(marshal(sd))
}

func validateOperation(t *testing.T, got *spec.Operation, want string) bool {
	templateB, err := json.Marshal(got)
	assert.NilError(t, err)

	differ := gojsondiff.New()
	diff, err := differ.Compare(templateB, []byte(want))
	assert.NilError(t, err)
	return diff.Modified() == false
}

func TestGenerateSpecOperation1(t *testing.T) {
	type args struct {
		data *HTTPInteractionData
	}
	tests := []struct {
		name       string
		args       args
		want       string
		wantErr    bool
		expectedSd spec.SecurityDefinitions
	}{
		{
			name: "Basic authorization req header",
			args: args{
				data: &HTTPInteractionData{
					ReqBody:  agentStatusBody,
					RespBody: cvssBody,
					ReqHeaders: map[string]string{
						contentTypeHeaderName:       mediaTypeApplicationJson,
						authorizationTypeHeaderName: BasicAuthPrefix + "=token",
					},
					RespHeaders: map[string]string{
						contentTypeHeaderName: mediaTypeApplicationJson,
					},
					statusCode: 200,
				},
			},
			want: "{\"security\":[{\"BasicAuth\":[]}],\"consumes\":[\"application/json\"],\"produces\":[\"application/json\"],\"parameters\":[{\"name\":\"body\",\"in\":\"body\",\"schema\":{\"type\":\"object\",\"properties\":{\"active\":{\"type\":\"boolean\"},\"certificateVersion\":{\"type\":\"string\",\"format\":\"uuid\"},\"controllerInstanceInfo\":{\"type\":\"object\",\"properties\":{\"replicaId\":{\"type\":\"string\"}}},\"policyAndAppVersion\":{\"type\":\"integer\",\"format\":\"int64\"},\"statusCodes\":{\"type\":\"array\",\"items\":{\"type\":\"string\"}},\"version\":{\"type\":\"string\"}}}}],\"responses\":{\"200\":{\"description\":\"\",\"schema\":{\"type\":\"object\",\"properties\":{\"cvss\":{\"type\":\"array\",\"items\":{\"type\":\"object\",\"properties\":{\"score\":{\"type\":\"number\",\"format\":\"double\"},\"vector\":{\"type\":\"string\"},\"version\":{\"type\":\"string\"}}}}}}},\"default\":{\"description\":\"Default Response\",\"schema\":{\"type\":\"object\",\"properties\":{\"message\":{\"type\":\"string\"}}}}}}",
			expectedSd: spec.SecurityDefinitions{
				BasicAuthSecurityDefinitionKey: spec.BasicAuth(),
			},
			wantErr: false,
		},
		{
			name: "OAuth 2.0 authorization req header",
			args: args{
				data: &HTTPInteractionData{
					ReqBody:  agentStatusBody,
					RespBody: cvssBody,
					ReqHeaders: map[string]string{
						contentTypeHeaderName:       mediaTypeApplicationJson,
						authorizationTypeHeaderName: BearerAuthPrefix + "=token",
					},
					RespHeaders: map[string]string{
						contentTypeHeaderName: mediaTypeApplicationJson,
					},
					statusCode: 200,
				},
			},
			want: "{\"security\":[{\"OAuth2\":[]}],\"consumes\":[\"application/json\"],\"produces\":[\"application/json\"],\"parameters\":[{\"name\":\"body\",\"in\":\"body\",\"schema\":{\"type\":\"object\",\"properties\":{\"active\":{\"type\":\"boolean\"},\"certificateVersion\":{\"type\":\"string\",\"format\":\"uuid\"},\"controllerInstanceInfo\":{\"type\":\"object\",\"properties\":{\"replicaId\":{\"type\":\"string\"}}},\"policyAndAppVersion\":{\"type\":\"integer\",\"format\":\"int64\"},\"statusCodes\":{\"type\":\"array\",\"items\":{\"type\":\"string\"}},\"version\":{\"type\":\"string\"}}}}],\"responses\":{\"200\":{\"description\":\"\",\"schema\":{\"type\":\"object\",\"properties\":{\"cvss\":{\"type\":\"array\",\"items\":{\"type\":\"object\",\"properties\":{\"score\":{\"type\":\"number\",\"format\":\"double\"},\"vector\":{\"type\":\"string\"},\"version\":{\"type\":\"string\"}}}}}}},\"default\":{\"description\":\"Default Response\",\"schema\":{\"type\":\"object\",\"properties\":{\"message\":{\"type\":\"string\"}}}}}}",
			expectedSd: spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey: spec.OAuth2AccessToken(authorizationURL, tokenURL),
			},
			wantErr: false,
		},
		{
			name: "OAuth 2.0 URI Query Parameter",
			args: args{
				data: &HTTPInteractionData{
					ReqBody:  agentStatusBody,
					RespBody: cvssBody,
					ReqHeaders: map[string]string{
						contentTypeHeaderName: mediaTypeApplicationJson,
					},
					RespHeaders: map[string]string{
						contentTypeHeaderName: mediaTypeApplicationJson,
					},
					QueryParams: generateQueryParams(t, AccessTokenParamKey+"=token"),
					statusCode:  200,
				},
			},
			want: "{\"security\":[{\"OAuth2\":[]}],\"consumes\":[\"application/json\"],\"produces\":[\"application/json\"],\"parameters\":[{\"name\":\"body\",\"in\":\"body\",\"schema\":{\"type\":\"object\",\"properties\":{\"active\":{\"type\":\"boolean\"},\"certificateVersion\":{\"type\":\"string\",\"format\":\"uuid\"},\"controllerInstanceInfo\":{\"type\":\"object\",\"properties\":{\"replicaId\":{\"type\":\"string\"}}},\"policyAndAppVersion\":{\"type\":\"integer\",\"format\":\"int64\"},\"statusCodes\":{\"type\":\"array\",\"items\":{\"type\":\"string\"}},\"version\":{\"type\":\"string\"}}}}],\"responses\":{\"200\":{\"description\":\"\",\"schema\":{\"type\":\"object\",\"properties\":{\"cvss\":{\"type\":\"array\",\"items\":{\"type\":\"object\",\"properties\":{\"score\":{\"type\":\"number\",\"format\":\"double\"},\"vector\":{\"type\":\"string\"},\"version\":{\"type\":\"string\"}}}}}}},\"default\":{\"description\":\"Default Response\",\"schema\":{\"type\":\"object\",\"properties\":{\"message\":{\"type\":\"string\"}}}}}}",
			expectedSd: spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey: spec.OAuth2AccessToken(authorizationURL, tokenURL),
			},
			wantErr: false,
		},
		{
			name: "OAuth 2.0 Form-Encoded Body Parameter",
			args: args{
				data: &HTTPInteractionData{
					ReqBody:  AccessTokenParamKey + "=token&key=val",
					RespBody: cvssBody,
					ReqHeaders: map[string]string{
						contentTypeHeaderName: mediaTypeApplicationForm,
					},
					RespHeaders: map[string]string{
						contentTypeHeaderName: mediaTypeApplicationJson,
					},
					statusCode: 200,
				},
			},
			want: "{\"security\":[{\"OAuth2\":[]}],\"consumes\":[\"application/x-www-form-urlencoded\"],\"produces\":[\"application/json\"],\"parameters\":[{\"type\":\"string\",\"name\":\"key\",\"in\":\"formData\"}],\"responses\":{\"200\":{\"description\":\"\",\"schema\":{\"type\":\"object\",\"properties\":{\"cvss\":{\"type\":\"array\",\"items\":{\"type\":\"object\",\"properties\":{\"score\":{\"type\":\"number\",\"format\":\"double\"},\"vector\":{\"type\":\"string\"},\"version\":{\"type\":\"string\"}}}}}}},\"default\":{\"description\":\"Default Response\",\"schema\":{\"type\":\"object\",\"properties\":{\"message\":{\"type\":\"string\"}}}}}}",
			expectedSd: spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey: spec.OAuth2AccessToken(authorizationURL, tokenURL),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd := spec.SecurityDefinitions{}
			got, err := GenerateSpecOperation(tt.args.data, sd)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSpecOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !validateOperation(t, got, tt.want) {
				t.Errorf("GenerateSpecOperation() got = %v, want %v", marshal(got), marshal(tt.want))
			}

			assert.DeepEqual(t, sd, tt.expectedSd)
		})
	}
}

func Test_getStringSchema(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantSchema *spec.Schema
	}{
		{
			name: "date",
			args: args{
				value: "2017-07-21",
			},
			wantSchema: spec.DateProperty(),
		},
		{
			name: "time",
			args: args{
				value: "17:32:28",
			},
			wantSchema: spec.StrFmtProperty("time"),
		},
		{
			name: "date-time",
			args: args{
				value: "2017-07-21T17:32:28Z",
			},
			wantSchema: spec.DateTimeProperty(),
		},
		{
			name: "email",
			args: args{
				value: "test@securecn.com",
			},
			wantSchema: spec.StrFmtProperty("email"),
		},
		{
			name: "ipv4",
			args: args{
				value: "1.1.1.1",
			},
			wantSchema: spec.StrFmtProperty("ipv4"),
		},
		{
			name: "ipv6",
			args: args{
				value: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			},
			wantSchema: spec.StrFmtProperty("ipv6"),
		},
		{
			name: "uuid",
			args: args{
				value: "123e4567-e89b-12d3-a456-426614174000",
			},
			wantSchema: spec.StrFmtProperty("uuid"),
		},
		{
			name: "json-pointer",
			args: args{
				value: "/k%22l",
			},
			wantSchema: spec.StrFmtProperty("json-pointer"),
		},
		{
			name: "string",
			args: args{
				value: "it is very hard to get a simple string",
			},
			wantSchema: spec.StringProperty(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSchema := getStringSchema(tt.args.value); !reflect.DeepEqual(gotSchema, tt.wantSchema) {
				t.Errorf("getStringSchema() = %v, want %v", gotSchema, tt.wantSchema)
			}
		})
	}
}

func Test_getNumberSchema(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantSchema *spec.Schema
	}{
		{
			name: "int",
			args: args{
				value: json.Number("85"),
			},
			wantSchema: spec.Int64Property(),
		},
		{
			name: "float",
			args: args{
				value: json.Number("85.1"),
			},
			wantSchema: spec.Float64Property(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSchema := getNumberSchema(tt.args.value); !reflect.DeepEqual(gotSchema, tt.wantSchema) {
				t.Errorf("getNumberSchema() = %v, want %v", gotSchema, tt.wantSchema)
			}
		})
	}
}

func Test_escapeString(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nothing to strip",
			args: args{
				key: "key",
			},
			want: "key",
		},
		{
			name: "escape double quotes",
			args: args{
				key: "{\"key1\":\"value1\", \"key2\":\"value2\"}",
			},
			want: "{\\\"key1\\\":\\\"value1\\\", \\\"key2\\\":\\\"value2\\\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeString(tt.args.key); got != tt.want {
				t.Errorf("stripKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCloneOperation(t *testing.T) {
	type args struct {
		op *spec.Operation
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
				op: spec.NewOperation("").
					AddParam(spec.HeaderParam("header")).
					RespondsWith(200, spec.ResponseRef("test")),
			},
			want: spec.NewOperation("").
				AddParam(spec.HeaderParam("header")).
				RespondsWith(200, spec.ResponseRef("test")),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := CloneOperation(tt.args.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("CloneOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CloneOperation() got = %v, want %v", got, tt.want)
			}
			if got != nil {
				got.Responses = nil
				if tt.args.op.Responses == nil {
					t.Errorf("CloneOperation() original object should not have been changed")
					return
				}
			}
		})
	}
}

func Test_handleAuthReqHeader(t *testing.T) {
	type args struct {
		operation *spec.Operation
		sd        spec.SecurityDefinitions
		value     string
	}
	tests := []struct {
		name   string
		args   args
		wantOp *spec.Operation
		wantSd spec.SecurityDefinitions
	}{
		{
			name: "BearerAuthPrefix",
			args: args{
				operation: spec.NewOperation(""),
				sd:        map[string]*spec.SecurityScheme{},
				value:     BearerAuthPrefix + "token",
			},
			wantOp: spec.NewOperation("").SecuredWith(OAuth2SecurityDefinitionKey, []string{}...),
			wantSd: spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey: spec.OAuth2AccessToken(authorizationURL, tokenURL),
			},
		},
		{
			name: "BasicAuthPrefix",
			args: args{
				operation: spec.NewOperation(""),
				sd:        map[string]*spec.SecurityScheme{},
				value:     BasicAuthPrefix + "token",
			},
			wantOp: spec.NewOperation("").SecuredWith(BasicAuthSecurityDefinitionKey, []string{}...),
			wantSd: spec.SecurityDefinitions{
				BasicAuthSecurityDefinitionKey: spec.BasicAuth(),
			},
		},
		{
			name: "ignoring unknown authorization header value",
			args: args{
				operation: spec.NewOperation(""),
				sd:        map[string]*spec.SecurityScheme{},
				value:     "invalid token",
			},
			wantOp: spec.NewOperation(""),
			wantSd: map[string]*spec.SecurityScheme{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := handleAuthReqHeader(tt.args.operation, tt.args.sd, tt.args.value)
			if !reflect.DeepEqual(got, tt.wantOp) {
				t.Errorf("handleAuthReqHeader() got = %v, want %v", got, tt.wantOp)
			}
			if !reflect.DeepEqual(got1, tt.wantSd) {
				t.Errorf("handleAuthReqHeader() got1 = %v, want %v", got1, tt.wantSd)
			}
		})
	}
}
