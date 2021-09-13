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

	oapi_spec "github.com/go-openapi/spec"
	"gotest.tools/assert"

	"github.com/apiclarity/speculator/pkg/pathtrie"
)

func TestSpec_LoadProvidedSpec(t *testing.T) {
	jsonSpec := "{\n  \"swagger\": \"2.0\",\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"APIClarity APIs\"\n  },\n  \"basePath\": \"/api\",\n  \"schemes\": [\n    \"http\"\n  ],\n  \"consumes\": [\n    \"application/json\"\n  ],\n  \"produces\": [\n    \"application/json\"\n  ],\n  \"paths\": {\n    \"/dashboard/apiUsage/mostUsed\": {\n      \"get\": {\n        \"summary\": \"Get most used APIs\",\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Success\",\n            \"schema\": {\n              \"type\": \"array\",\n              \"items\": {\n                \"type\": \"string\"\n              }\n            }\n          },\n          \"default\": {\n            \"$ref\": \"#/responses/UnknownError\"\n          }\n        }\n      }\n    }\n  },\n  \"definitions\": {\n    \"ApiResponse\": {\n      \"description\": \"An object that is return in all cases of failures.\",\n      \"type\": \"object\",\n      \"properties\": {\n        \"message\": {\n          \"type\": \"string\"\n        }\n      }\n    }\n  },\n  \"responses\": {\n    \"UnknownError\": {\n      \"description\": \"unknown error\",\n      \"schema\": {\n        \"$ref\": \"#/definitions/ApiResponse\"\n      }\n    }\n  }\n}"
	jsonSpecInvalid := "{\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"APIClarity APIs\"\n  },\n  \"basePath\": \"/api\",\n  \"schemes\": [\n    \"http\"\n  ],\n  \"consumes\": [\n    \"application/json\"\n  ],\n  \"produces\": [\n    \"application/json\"\n  ],\n  \"paths\": {\n    \"/dashboard/apiUsage/mostUsed\": {\n      \"get\": {\n        \"summary\": \"Get most used APIs\",\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Success\",\n            \"schema\": {\n              \"type\": \"array\",\n              \"items\": {\n                \"type\": \"string\"\n              }\n            }\n          },\n          \"default\": {\n            \"$ref\": \"#/responses/UnknownError\"\n          }\n        }\n      }\n    }\n  },\n  \"definitions\": {\n    \"ApiResponse\": {\n      \"description\": \"An object that is return in all cases of failures.\",\n      \"type\": \"object\",\n      \"properties\": {\n        \"message\": {\n          \"type\": \"string\"\n        }\n      }\n    }\n  },\n  \"responses\": {\n    \"UnknownError\": {\n      \"description\": \"unknown error\",\n      \"schema\": {\n        \"$ref\": \"#/definitions/ApiResponse\"\n      }\n    }\n  }\n}"
	yamlSpec := "---\nswagger: '2.0'\ninfo:\n  version: 1.0.0\n  title: APIClarity APIs\nbasePath: \"/api\"\nschemes:\n- http\nconsumes:\n- application/json\nproduces:\n- application/json\npaths:\n  \"/dashboard/apiUsage/mostUsed\":\n    get:\n      summary: Get most used APIs\n      responses:\n        '200':\n          description: Success\n          schema:\n            type: array\n            items:\n              type: string\n        default:\n          \"$ref\": \"#/responses/UnknownError\"\ndefinitions:\n  ApiResponse:\n    description: An object that is return in all cases of failures.\n    type: object\n    properties:\n      message:\n        type: string\nresponses:\n  UnknownError:\n    description: unknown error\n    schema:\n      \"$ref\": \"#/definitions/ApiResponse\"\n"
	wantProvidedSpec := &ProvidedSpec{
		Spec: &oapi_spec.Swagger{
			SwaggerProps: oapi_spec.SwaggerProps{
				Paths: &oapi_spec.Paths{
					Paths: map[string]oapi_spec.PathItem{},
				},
			},
		},
	}
	err := json.Unmarshal([]byte(jsonSpec), wantProvidedSpec.Spec)
	assert.NilError(t, err)

	pathToPathID := map[string]string{
		"/dashboard/apiUsage/mostUsed": "1",
	}
	wantProvidedPathTrie := createPathTrie(pathToPathID)
	emptyPathTrie := createPathTrie(nil)

	type fields struct {
		ProvidedSpec *ProvidedSpec
	}
	type args struct {
		providedSpec []byte
		pathToPathID map[string]string
	}
	tests := []struct {
		name                 string
		fields               fields
		args                 args
		wantErr              bool
		wantProvidedPathTrie pathtrie.PathTrie
	}{
		{
			name: "json spec",
			fields: fields{
				ProvidedSpec: nil,
			},
			args: args{
				providedSpec: []byte(jsonSpec),
				pathToPathID: pathToPathID,
			},
			wantErr:              false,
			wantProvidedPathTrie: wantProvidedPathTrie,
		},
		{
			name: "json spec with a missing path",
			fields: fields{
				ProvidedSpec: nil,
			},
			args: args{
				providedSpec: []byte(jsonSpec),
				pathToPathID: map[string]string{},
			},
			wantErr:              false,
			wantProvidedPathTrie: emptyPathTrie,
		},
		{
			name: "yaml spec",
			fields: fields{
				ProvidedSpec: nil,
			},
			args: args{
				providedSpec: []byte(yamlSpec),
				pathToPathID: pathToPathID,
			},
			wantErr:              false,
			wantProvidedPathTrie: wantProvidedPathTrie,
		},
		{
			name: "invalid json",
			fields: fields{
				ProvidedSpec: nil,
			},
			args: args{
				providedSpec: []byte("bad" + jsonSpec),
			},
			wantErr: true,
		},
		{
			name: "invalid spec",
			fields: fields{
				ProvidedSpec: nil,
			},
			args: args{
				providedSpec: []byte(jsonSpecInvalid),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Spec{
				SpecInfo: SpecInfo{
					ProvidedSpec: tt.fields.ProvidedSpec,
				},
			}
			if err := s.LoadProvidedSpec(tt.args.providedSpec, tt.args.pathToPathID); (err != nil) != tt.wantErr {
				t.Errorf("LoadProvidedSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(s.ProvidedSpec, wantProvidedSpec) {
					t.Errorf("LoadProvidedSpec() got = %v, want %v", marshal(s.ProvidedSpec), marshal(wantProvidedSpec))
				}
				if !reflect.DeepEqual(s.ProvidedPathTrie, tt.wantProvidedPathTrie) {
					t.Errorf("LoadProvidedSpec() got = %v, want %v", marshal(s.ProvidedPathTrie), marshal(tt.wantProvidedPathTrie))
				}
			}
		})
	}
}
