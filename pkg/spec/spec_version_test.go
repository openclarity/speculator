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

import "testing"

func TestGetJSONSpecVersion(t *testing.T) {
	jsonSpecV2 := "{\n  \"swagger\": \"2.0\",\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"APIClarity APIs\"\n  },\n  \"basePath\": \"/api\",\n  \"schemes\": [\n    \"http\"\n  ],\n  \"consumes\": [\n    \"application/json\"\n  ],\n  \"produces\": [\n    \"application/json\"\n  ],\n  \"paths\": {\n    \"/dashboard/apiUsage/mostUsed\": {\n      \"get\": {\n        \"summary\": \"Get most used APIs\",\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Success\",\n            \"schema\": {\n              \"type\": \"array\",\n              \"items\": {\n                \"type\": \"string\"\n              }\n            }\n          },\n          \"default\": {\n            \"$ref\": \"#/responses/UnknownError\"\n          }\n        }\n      }\n    }\n  },\n  \"schemas\": {\n    \"ApiResponse\": {\n      \"description\": \"An object that is return in all cases of failures.\",\n      \"type\": \"object\",\n      \"properties\": {\n        \"message\": {\n          \"type\": \"string\"\n        }\n      }\n    }\n  },\n  \"responses\": {\n    \"UnknownError\": {\n      \"description\": \"unknown error\",\n      \"schema\": {\n        \"$ref\": \"\"\n      }\n    }\n  }\n}"
	jsonSpecV2Invalid := "{\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"APIClarity APIs\"\n  },\n  \"basePath\": \"/api\",\n  \"schemes\": [\n    \"http\"\n  ],\n  \"consumes\": [\n    \"application/json\"\n  ],\n  \"produces\": [\n    \"application/json\"\n  ],\n  \"paths\": {\n    \"/dashboard/apiUsage/mostUsed\": {\n      \"get\": {\n        \"summary\": \"Get most used APIs\",\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Success\",\n            \"schema\": {\n              \"type\": \"array\",\n              \"items\": {\n                \"type\": \"string\"\n              }\n            }\n          },\n          \"default\": {\n            \"$ref\": \"#/responses/UnknownError\"\n          }\n        }\n      }\n    }\n  },\n  \"schemas\": {\n    \"ApiResponse\": {\n      \"description\": \"An object that is return in all cases of failures.\",\n      \"type\": \"object\",\n      \"properties\": {\n        \"message\": {\n          \"type\": \"string\"\n        }\n      }\n    }\n  },\n  \"responses\": {\n    \"UnknownError\": {\n      \"description\": \"unknown error\",\n      \"schema\": {\n        \"$ref\": \"#/schemas/ApiResponse\"\n      }\n    }\n  }\n}"
	jsonSpecV3 := "{\n  \"openapi\": \"3.0.3\",\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"Simple API\",\n    \"description\": \"A simple API to illustrate OpenAPI concepts\"\n  },\n  \"servers\": [\n    {\n      \"url\": \"https://example.io/v1\"\n    }\n  ],\n  \"security\": [\n    {\n      \"BasicAuth\": []\n    }\n  ],\n  \"paths\": {\n    \"/artists\": {\n      \"get\": {\n        \"description\": \"Returns a list of artists\",\n        \"parameters\": [\n          {\n            \"name\": \"limit\",\n            \"in\": \"query\",\n            \"description\": \"Limits the number of items on a page\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          },\n          {\n            \"name\": \"offset\",\n            \"in\": \"query\",\n            \"description\": \"Specifies the page number of the artists to be displayed\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned a list of artists\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"array\",\n                  \"items\": {\n                    \"type\": \"object\",\n                    \"required\": [\n                      \"username\"\n                    ],\n                    \"properties\": {\n                      \"artist_name\": {\n                        \"type\": \"string\"\n                      },\n                      \"artist_genre\": {\n                        \"type\": \"string\"\n                      },\n                      \"albums_recorded\": {\n                        \"type\": \"integer\"\n                      },\n                      \"username\": {\n                        \"type\": \"string\"\n                      }\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      },\n      \"post\": {\n        \"description\": \"Lets a user post a new artist\",\n        \"requestBody\": {\n          \"required\": true,\n          \"content\": {\n            \"application/json\": {\n              \"schema\": {\n                \"type\": \"array\",\n                \"items\": {\n                  \"type\": \"object\",\n                  \"required\": [\n                    \"username\"\n                  ],\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    },\n                    \"username\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        },\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully created a new artist\"\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    },\n    \"/artists/{username}\": {\n      \"get\": {\n        \"description\": \"Obtain information about an artist from his or her unique username\",\n        \"parameters\": [\n          {\n            \"name\": \"username\",\n            \"in\": \"path\",\n            \"required\": true,\n            \"schema\": {\n              \"type\": \"string\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned an artist\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    }\n  },\n  \"components\": {\n    \"securitySchemes\": {\n      \"BasicAuth\": {\n        \"type\": \"http\",\n        \"scheme\": \"basic\"\n      }\n    }\n  }\n}"
	jsonSpecV3Invalid := "{\n  \"openapi\": \"\",\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"Simple API\",\n    \"description\": \"A simple API to illustrate OpenAPI concepts\"\n  },\n  \"servers\": [\n    {\n      \"url\": \"https://example.io/v1\"\n    }\n  ],\n  \"security\": [\n    {\n      \"BasicAuth\": []\n    }\n  ],\n  \"paths\": {\n    \"/artists\": {\n      \"get\": {\n        \"description\": \"Returns a list of artists\",\n        \"parameters\": [\n          {\n            \"name\": \"limit\",\n            \"in\": \"query\",\n            \"description\": \"Limits the number of items on a page\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          },\n          {\n            \"name\": \"offset\",\n            \"in\": \"query\",\n            \"description\": \"Specifies the page number of the artists to be displayed\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned a list of artists\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"array\",\n                  \"items\": {\n                    \"type\": \"object\",\n                    \"required\": [\n                      \"username\"\n                    ],\n                    \"properties\": {\n                      \"artist_name\": {\n                        \"type\": \"string\"\n                      },\n                      \"artist_genre\": {\n                        \"type\": \"string\"\n                      },\n                      \"albums_recorded\": {\n                        \"type\": \"integer\"\n                      },\n                      \"username\": {\n                        \"type\": \"string\"\n                      }\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      },\n      \"post\": {\n        \"description\": \"Lets a user post a new artist\",\n        \"requestBody\": {\n          \"required\": true,\n          \"content\": {\n            \"application/json\": {\n              \"schema\": {\n                \"type\": \"array\",\n                \"items\": {\n                  \"type\": \"object\",\n                  \"required\": [\n                    \"username\"\n                  ],\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    },\n                    \"username\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        },\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully created a new artist\"\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    },\n    \"/artists/{username}\": {\n      \"get\": {\n        \"description\": \"Obtain information about an artist from his or her unique username\",\n        \"parameters\": [\n          {\n            \"name\": \"username\",\n            \"in\": \"path\",\n            \"required\": true,\n            \"schema\": {\n              \"type\": \"string\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned an artist\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    }\n  },\n  \"components\": {\n    \"securitySchemes\": {\n      \"BasicAuth\": {\n        \"type\": \"http\",\n        \"scheme\": \"basic\"\n      }\n    }\n  }\n}"

	type args struct {
		jsonSpec []byte
	}
	tests := []struct {
		name    string
		args    args
		want    OASVersion
		wantErr bool
	}{
		{
			name: "valid v2 spec",
			args: args{
				jsonSpec: []byte(jsonSpecV2),
			},
			want:    OASv2,
			wantErr: false,
		},
		{
			name: "invalid v2 spec",
			args: args{
				jsonSpec: []byte(jsonSpecV2Invalid),
			},
			want:    Unknown,
			wantErr: true,
		},
		{
			name: "valid v3 spec",
			args: args{
				jsonSpec: []byte(jsonSpecV3),
			},
			want:    OASv3,
			wantErr: false,
		},
		{
			name: "invalid v3 spec",
			args: args{
				jsonSpec: []byte(jsonSpecV3Invalid),
			},
			want:    Unknown,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetJSONSpecVersion(tt.args.jsonSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetJSONSpecVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetJSONSpecVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}
