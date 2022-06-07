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
	"github.com/google/go-cmp/cmp/cmpopts"
	"reflect"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"gotest.tools/assert"

	"github.com/openclarity/speculator/pkg/pathtrie"
)

func TestSpec_LoadProvidedSpec(t *testing.T) {
	//jsonSpec := "{\n  \"swagger\": \"2.0\",\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"APIClarity APIs\"\n  },\n  \"basePath\": \"/api\",\n  \"schemes\": [\n    \"http\"\n  ],\n  \"consumes\": [\n    \"application/json\"\n  ],\n  \"produces\": [\n    \"application/json\"\n  ],\n  \"paths\": {\n    \"/dashboard/apiUsage/mostUsed\": {\n      \"get\": {\n        \"summary\": \"Get most used APIs\",\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Success\",\n            \"schema\": {\n              \"type\": \"array\",\n              \"items\": {\n                \"type\": \"string\"\n              }\n            }\n          },\n          \"default\": {\n            \"$ref\": \"#/responses/UnknownError\"\n          }\n        }\n      }\n    }\n  },\n  \"schemas\": {\n    \"ApiResponse\": {\n      \"description\": \"An object that is return in all cases of failures.\",\n      \"type\": \"object\",\n      \"properties\": {\n        \"message\": {\n          \"type\": \"string\"\n        }\n      }\n    }\n  },\n  \"responses\": {\n    \"UnknownError\": {\n      \"description\": \"unknown error\",\n      \"schema\": {\n        \"$ref\": \"#/schemas/ApiResponse\"\n      }\n    }\n  }\n}"
	//jsonSpecInvalid := "{\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"APIClarity APIs\"\n  },\n  \"basePath\": \"/api\",\n  \"schemes\": [\n    \"http\"\n  ],\n  \"consumes\": [\n    \"application/json\"\n  ],\n  \"produces\": [\n    \"application/json\"\n  ],\n  \"paths\": {\n    \"/dashboard/apiUsage/mostUsed\": {\n      \"get\": {\n        \"summary\": \"Get most used APIs\",\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Success\",\n            \"schema\": {\n              \"type\": \"array\",\n              \"items\": {\n                \"type\": \"string\"\n              }\n            }\n          },\n          \"default\": {\n            \"$ref\": \"#/responses/UnknownError\"\n          }\n        }\n      }\n    }\n  },\n  \"schemas\": {\n    \"ApiResponse\": {\n      \"description\": \"An object that is return in all cases of failures.\",\n      \"type\": \"object\",\n      \"properties\": {\n        \"message\": {\n          \"type\": \"string\"\n        }\n      }\n    }\n  },\n  \"responses\": {\n    \"UnknownError\": {\n      \"description\": \"unknown error\",\n      \"schema\": {\n        \"$ref\": \"#/schemas/ApiResponse\"\n      }\n    }\n  }\n}"
	jsonSpec := "{\n  \"openapi\": \"3.0.3\",\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"Simple API\",\n    \"description\": \"A simple API to illustrate OpenAPI concepts\"\n  },\n  \"servers\": [\n    {\n      \"url\": \"https://example.io/v1\"\n    }\n  ],\n  \"security\": [\n    {\n      \"BasicAuth\": []\n    }\n  ],\n  \"paths\": {\n    \"/artists\": {\n      \"get\": {\n        \"description\": \"Returns a list of artists\",\n        \"parameters\": [\n          {\n            \"name\": \"limit\",\n            \"in\": \"query\",\n            \"description\": \"Limits the number of items on a page\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          },\n          {\n            \"name\": \"offset\",\n            \"in\": \"query\",\n            \"description\": \"Specifies the page number of the artists to be displayed\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned a list of artists\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"array\",\n                  \"items\": {\n                    \"type\": \"object\",\n                    \"required\": [\n                      \"username\"\n                    ],\n                    \"properties\": {\n                      \"artist_name\": {\n                        \"type\": \"string\"\n                      },\n                      \"artist_genre\": {\n                        \"type\": \"string\"\n                      },\n                      \"albums_recorded\": {\n                        \"type\": \"integer\"\n                      },\n                      \"username\": {\n                        \"type\": \"string\"\n                      }\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      },\n      \"post\": {\n        \"description\": \"Lets a user post a new artist\",\n        \"requestBody\": {\n          \"required\": true,\n          \"content\": {\n            \"application/json\": {\n              \"schema\": {\n                \"type\": \"array\",\n                \"items\": {\n                  \"type\": \"object\",\n                  \"required\": [\n                    \"username\"\n                  ],\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    },\n                    \"username\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        },\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully created a new artist\"\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    },\n    \"/artists/{username}\": {\n      \"get\": {\n        \"description\": \"Obtain information about an artist from his or her unique username\",\n        \"parameters\": [\n          {\n            \"name\": \"username\",\n            \"in\": \"path\",\n            \"required\": true,\n            \"schema\": {\n              \"type\": \"string\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned an artist\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    }\n  },\n  \"components\": {\n    \"securitySchemes\": {\n      \"BasicAuth\": {\n        \"type\": \"http\",\n        \"scheme\": \"basic\"\n      }\n    }\n  }\n}"
	jsonSpecInvalid := "{\n  \"openapi\": \"\",\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"Simple API\",\n    \"description\": \"A simple API to illustrate OpenAPI concepts\"\n  },\n  \"servers\": [\n    {\n      \"url\": \"https://example.io/v1\"\n    }\n  ],\n  \"security\": [\n    {\n      \"BasicAuth\": []\n    }\n  ],\n  \"paths\": {\n    \"/artists\": {\n      \"get\": {\n        \"description\": \"Returns a list of artists\",\n        \"parameters\": [\n          {\n            \"name\": \"limit\",\n            \"in\": \"query\",\n            \"description\": \"Limits the number of items on a page\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          },\n          {\n            \"name\": \"offset\",\n            \"in\": \"query\",\n            \"description\": \"Specifies the page number of the artists to be displayed\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned a list of artists\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"array\",\n                  \"items\": {\n                    \"type\": \"object\",\n                    \"required\": [\n                      \"username\"\n                    ],\n                    \"properties\": {\n                      \"artist_name\": {\n                        \"type\": \"string\"\n                      },\n                      \"artist_genre\": {\n                        \"type\": \"string\"\n                      },\n                      \"albums_recorded\": {\n                        \"type\": \"integer\"\n                      },\n                      \"username\": {\n                        \"type\": \"string\"\n                      }\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      },\n      \"post\": {\n        \"description\": \"Lets a user post a new artist\",\n        \"requestBody\": {\n          \"required\": true,\n          \"content\": {\n            \"application/json\": {\n              \"schema\": {\n                \"type\": \"array\",\n                \"items\": {\n                  \"type\": \"object\",\n                  \"required\": [\n                    \"username\"\n                  ],\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    },\n                    \"username\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        },\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully created a new artist\"\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    },\n    \"/artists/{username}\": {\n      \"get\": {\n        \"description\": \"Obtain information about an artist from his or her unique username\",\n        \"parameters\": [\n          {\n            \"name\": \"username\",\n            \"in\": \"path\",\n            \"required\": true,\n            \"schema\": {\n              \"type\": \"string\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned an artist\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    }\n  },\n  \"components\": {\n    \"securitySchemes\": {\n      \"BasicAuth\": {\n        \"type\": \"http\",\n        \"scheme\": \"basic\"\n      }\n    }\n  }\n}"
	//yamlSpec := "---\nswagger: '2.0'\ninfo:\n  version: 1.0.0\n  title: APIClarity APIs\nbasePath: \"/api\"\nschemes:\n- http\nconsumes:\n- application/json\nproduces:\n- application/json\npaths:\n  \"/dashboard/apiUsage/mostUsed\":\n    get:\n      summary: Get most used APIs\n      responses:\n        '200':\n          description: Success\n          schema:\n            type: array\n            items:\n              type: string\n        default:\n          \"$ref\": \"#/responses/UnknownError\"\nschemas:\n  ApiResponse:\n    description: An object that is return in all cases of failures.\n    type: object\n    properties:\n      message:\n        type: string\nresponses:\n  UnknownError:\n    description: unknown error\n    schema:\n      \"$ref\": \"#/schemas/ApiResponse\"\n"
	yamlSpec := "openapi: 3.0.3\ninfo:\n  version: 1.0.0\n  title: Simple API\n  description: A simple API to illustrate OpenAPI concepts\n\nservers:\n  - url: https://example.io/v1\n\nsecurity:\n  - BasicAuth: []\n\npaths:\n  /artists:\n    get:\n      description: Returns a list of artists \n      parameters:\n        - name: limit\n          in: query\n          description: Limits the number of items on a page\n          schema:\n            type: integer\n        - name: offset\n          in: query\n          description: Specifies the page number of the artists to be displayed\n          schema:\n            type: integer\n      responses:\n        '200':\n          description: Successfully returned a list of artists\n          content:\n            application/json:\n              schema:\n                type: array\n                items:\n                  type: object\n                  required:\n                    - username\n                  properties:\n                    artist_name:\n                      type: string\n                    artist_genre:\n                        type: string\n                    albums_recorded:\n                        type: integer\n                    username:\n                        type: string\n        '400':\n          description: Invalid request\n          content:\n            application/json:\n              schema:\n                type: object \n                properties:\n                  message:\n                    type: string\n\n    post:\n      description: Lets a user post a new artist\n      requestBody:\n        required: true\n        content:\n          application/json:\n            schema:\n              type: array\n              items:\n                type: object\n                required:\n                  - username\n                properties:\n                  artist_name:\n                    type: string\n                  artist_genre:\n                      type: string\n                  albums_recorded:\n                      type: integer\n                  username:\n                      type: string\n      responses:\n        '200':\n          description: Successfully created a new artist\n        '400':\n          description: Invalid request\n          content:\n            application/json:\n              schema:\n                type: object \n                properties:\n                  message:\n                    type: string\n\n  /artists/{username}:\n    get:\n      description: Obtain information about an artist from his or her unique username\n      parameters:\n        - name: username\n          in: path\n          required: true\n          schema:\n            type: string\n          \n      responses:\n        '200':\n          description: Successfully returned an artist\n          content:\n            application/json:\n              schema:\n                type: object\n                properties:\n                  artist_name:\n                    type: string\n                  artist_genre:\n                    type: string\n                  albums_recorded:\n                    type: integer\n                \n        '400':\n          description: Invalid request\n          content:\n            application/json:\n              schema:\n                type: object \n                properties:\n                  message:\n                    type: string\n\ncomponents:\n  securitySchemes:\n    BasicAuth:\n      type: http\n      scheme: basic\n"
	wantProvidedSpec := &ProvidedSpec{
		Doc: &openapi3.T{
			Paths: openapi3.Paths{},
		},
	}
	err := json.Unmarshal([]byte(jsonSpec), wantProvidedSpec.Doc)
	assert.NilError(t, err)

	pathToPathID := map[string]string{
		//"/dashboard/apiUsage/mostUsed": "1",
		"/artists": "1",
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
				assert.DeepEqual(t, s.ProvidedSpec, wantProvidedSpec, cmpopts.IgnoreUnexported(openapi3.Schema{}), cmpopts.IgnoreTypes(openapi3.ExtensionProps{}))
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

func TestProvidedSpec_GetBasePath(t *testing.T) {
	type fields struct {
		Doc *openapi3.T
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "url templating",
			fields: fields{
				Doc: &openapi3.T{
					Servers: []*openapi3.Server{
						{
							URL: "{protocol}://api.example.com/api",
						},
					},
				},
			},
			want: "/api",
		},
		{
			name: "sanity",
			fields: fields{
				Doc: &openapi3.T{
					Servers: []*openapi3.Server{
						{
							URL: "https://api.example.com:8443/v1/reports",
						},
					},
				},
			},
			want: "/v1/reports",
		},
		{
			name: "no path",
			fields: fields{
				Doc: &openapi3.T{
					Servers: []*openapi3.Server{
						{
							URL: "https://api.example.com",
						},
					},
				},
			},
			want: "",
		},
		{
			name: "no url",
			fields: fields{
				Doc: &openapi3.T{
					Servers: []*openapi3.Server{
						{
							URL: "",
						},
					},
				},
			},
			want: "",
		},
		{
			name: "only path",
			fields: fields{
				Doc: &openapi3.T{
					Servers: []*openapi3.Server{
						{
							URL: "/v1/reports",
						},
					},
				},
			},
			want: "/v1/reports",
		},
		{
			name: "root path",
			fields: fields{
				Doc: &openapi3.T{
					Servers: []*openapi3.Server{
						{
							URL: "/",
						},
					},
				},
			},
			want: "",
		},
		{
			name: "ip",
			fields: fields{
				Doc: &openapi3.T{
					Servers: []*openapi3.Server{
						{
							URL: "http://10.0.81.36/v1",
						},
					},
				},
			},
			want: "/v1",
		},
		{
			name: "bad url",
			fields: fields{
				Doc: &openapi3.T{
					Servers: []*openapi3.Server{
						{
							URL: "bad.url.dot.com.!@##",
						},
					},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProvidedSpec{
				Doc: tt.fields.Doc,
			}
			if got := p.GetBasePath(); got != tt.want {
				t.Errorf("GetBasePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
