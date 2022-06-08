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
	jsonSpecV2 := "{\n  \"swagger\": \"2.0\",\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"APIClarity APIs\"\n  },\n  \"basePath\": \"/api\",\n  \"schemes\": [\n    \"http\"\n  ],\n  \"consumes\": [\n    \"application/json\"\n  ],\n  \"produces\": [\n    \"application/json\"\n  ],\n  \"paths\": {\n    \"/dashboard/apiUsage/mostUsed\": {\n      \"get\": {\n        \"summary\": \"Get most used APIs\",\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Success\",\n            \"schema\": {\n              \"type\": \"array\",\n              \"items\": {\n                \"type\": \"string\"\n              }\n            }\n          },\n          \"default\": {\n            \"$ref\": \"#/responses/UnknownError\"\n          }\n        }\n      }\n    }\n  },\n  \"schemas\": {\n    \"ApiResponse\": {\n      \"description\": \"An object that is return in all cases of failures.\",\n      \"type\": \"object\",\n      \"properties\": {\n        \"message\": {\n          \"type\": \"string\"\n        }\n      }\n    }\n  },\n  \"responses\": {\n    \"UnknownError\": {\n      \"description\": \"unknown error\",\n      \"schema\": {\n        \"$ref\": \"\"\n      }\n    }\n  }\n}"
	jsonSpecV2Invalid := "{\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"APIClarity APIs\"\n  },\n  \"basePath\": \"/api\",\n  \"schemes\": [\n    \"http\"\n  ],\n  \"consumes\": [\n    \"application/json\"\n  ],\n  \"produces\": [\n    \"application/json\"\n  ],\n  \"paths\": {\n    \"/dashboard/apiUsage/mostUsed\": {\n      \"get\": {\n        \"summary\": \"Get most used APIs\",\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Success\",\n            \"schema\": {\n              \"type\": \"array\",\n              \"items\": {\n                \"type\": \"string\"\n              }\n            }\n          },\n          \"default\": {\n            \"$ref\": \"#/responses/UnknownError\"\n          }\n        }\n      }\n    }\n  },\n  \"schemas\": {\n    \"ApiResponse\": {\n      \"description\": \"An object that is return in all cases of failures.\",\n      \"type\": \"object\",\n      \"properties\": {\n        \"message\": {\n          \"type\": \"string\"\n        }\n      }\n    }\n  },\n  \"responses\": {\n    \"UnknownError\": {\n      \"description\": \"unknown error\",\n      \"schema\": {\n        \"$ref\": \"#/schemas/ApiResponse\"\n      }\n    }\n  }\n}"
	jsonSpec := "{\n  \"openapi\": \"3.0.3\",\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"Simple API\",\n    \"description\": \"A simple API to illustrate OpenAPI concepts\"\n  },\n  \"servers\": [\n    {\n      \"url\": \"https://example.io/v1\"\n    }\n  ],\n  \"security\": [\n    {\n      \"BasicAuth\": []\n    }\n  ],\n  \"paths\": {\n    \"/artists\": {\n      \"get\": {\n        \"description\": \"Returns a list of artists\",\n        \"parameters\": [\n          {\n            \"name\": \"limit\",\n            \"in\": \"query\",\n            \"description\": \"Limits the number of items on a page\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          },\n          {\n            \"name\": \"offset\",\n            \"in\": \"query\",\n            \"description\": \"Specifies the page number of the artists to be displayed\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned a list of artists\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"array\",\n                  \"items\": {\n                    \"type\": \"object\",\n                    \"required\": [\n                      \"username\"\n                    ],\n                    \"properties\": {\n                      \"artist_name\": {\n                        \"type\": \"string\"\n                      },\n                      \"artist_genre\": {\n                        \"type\": \"string\"\n                      },\n                      \"albums_recorded\": {\n                        \"type\": \"integer\"\n                      },\n                      \"username\": {\n                        \"type\": \"string\"\n                      }\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      },\n      \"post\": {\n        \"description\": \"Lets a user post a new artist\",\n        \"requestBody\": {\n          \"required\": true,\n          \"content\": {\n            \"application/json\": {\n              \"schema\": {\n                \"type\": \"array\",\n                \"items\": {\n                  \"type\": \"object\",\n                  \"required\": [\n                    \"username\"\n                  ],\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    },\n                    \"username\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        },\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully created a new artist\"\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    },\n    \"/artists/{username}\": {\n      \"get\": {\n        \"description\": \"Obtain information about an artist from his or her unique username\",\n        \"parameters\": [\n          {\n            \"name\": \"username\",\n            \"in\": \"path\",\n            \"required\": true,\n            \"schema\": {\n              \"type\": \"string\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned an artist\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    }\n  },\n  \"components\": {\n    \"securitySchemes\": {\n      \"BasicAuth\": {\n        \"type\": \"http\",\n        \"scheme\": \"basic\"\n      }\n    }\n  }\n}"
	jsonSpecWithRef := "{\n  \"openapi\": \"3.0.3\",\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"Simple API\",\n    \"description\": \"A simple API to illustrate OpenAPI concepts\"\n  },\n  \"servers\": [\n    {\n      \"url\": \"https://example.io/v1\"\n    }\n  ],\n  \"security\": [\n    {\n      \"BasicAuth\": []\n    }\n  ],\n  \"paths\": {\n    \"/artists\": {\n      \"get\": {\n        \"description\": \"Returns a list of artists\",\n        \"parameters\": [\n          {\n            \"name\": \"limit\",\n            \"in\": \"query\",\n            \"description\": \"Limits the number of items on a page\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          },\n          {\n            \"name\": \"offset\",\n            \"in\": \"query\",\n            \"description\": \"Specifies the page number of the artists to be displayed\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned a list of artists\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"array\",\n                  \"items\": {\n                    \"type\": \"object\",\n                    \"required\": [\n                      \"username\"\n                    ],\n                    \"properties\": {\n                      \"artist_name\": {\n                        \"type\": \"string\"\n                      },\n                      \"artist_genre\": {\n                        \"type\": \"string\"\n                      },\n                      \"albums_recorded\": {\n                        \"type\": \"integer\"\n                      },\n                      \"username\": {\n                        \"type\": \"string\"\n                      }\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      },\n      \"post\": {\n        \"description\": \"Lets a user post a new artist\",\n        \"requestBody\": {\n          \"required\": true,\n          \"content\": {\n            \"application/json\": {\n              \"schema\": {\n                \"type\": \"array\",\n                \"items\": {\n                  \"type\": \"object\",\n                  \"required\": [\n                    \"username\"\n                  ],\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    },\n                    \"username\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        },\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully created a new artist\"\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    },\n    \"/artists/{username}\": {\n      \"get\": {\n        \"description\": \"Obtain information about an artist from his or her unique username\",\n        \"parameters\": [\n          {\n            \"name\": \"username\",\n            \"in\": \"path\",\n            \"required\": true,\n            \"schema\": {\n              \"type\": \"string\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned an artist\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"$ref\": \"#/components/schemas/Artists\"\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    }\n  },\n  \"components\": {\n    \"schemas\": {\n      \"Artists\": {\n        \"type\": \"object\",\n        \"properties\": {\n          \"artist_name\": {\n            \"type\": \"string\"\n          },\n          \"artist_genre\": {\n            \"type\": \"string\"\n          },\n          \"albums_recorded\": {\n            \"type\": \"integer\"\n          }\n        }\n      }\n    },\n    \"securitySchemes\": {\n      \"BasicAuth\": {\n        \"type\": \"http\",\n        \"scheme\": \"basic\"\n      }\n    }\n  }\n}"
	jsonSpecWithRefAfterRemoveRef := "{\n  \"openapi\": \"3.0.3\",\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"Simple API\",\n    \"description\": \"A simple API to illustrate OpenAPI concepts\"\n  },\n  \"servers\": [\n    {\n      \"url\": \"https://example.io/v1\"\n    }\n  ],\n  \"security\": [\n    {\n      \"BasicAuth\": []\n    }\n  ],\n  \"paths\": {\n    \"/artists\": {\n      \"get\": {\n        \"description\": \"Returns a list of artists\",\n        \"parameters\": [\n          {\n            \"name\": \"limit\",\n            \"in\": \"query\",\n            \"description\": \"Limits the number of items on a page\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          },\n          {\n            \"name\": \"offset\",\n            \"in\": \"query\",\n            \"description\": \"Specifies the page number of the artists to be displayed\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned a list of artists\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"array\",\n                  \"items\": {\n                    \"type\": \"object\",\n                    \"required\": [\n                      \"username\"\n                    ],\n                    \"properties\": {\n                      \"artist_name\": {\n                        \"type\": \"string\"\n                      },\n                      \"artist_genre\": {\n                        \"type\": \"string\"\n                      },\n                      \"albums_recorded\": {\n                        \"type\": \"integer\"\n                      },\n                      \"username\": {\n                        \"type\": \"string\"\n                      }\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      },\n      \"post\": {\n        \"description\": \"Lets a user post a new artist\",\n        \"requestBody\": {\n          \"required\": true,\n          \"content\": {\n            \"application/json\": {\n              \"schema\": {\n                \"type\": \"array\",\n                \"items\": {\n                  \"type\": \"object\",\n                  \"required\": [\n                    \"username\"\n                  ],\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    },\n                    \"username\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        },\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully created a new artist\"\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    },\n    \"/artists/{username}\": {\n      \"get\": {\n        \"description\": \"Obtain information about an artist from his or her unique username\",\n        \"parameters\": [\n          {\n            \"name\": \"username\",\n            \"in\": \"path\",\n            \"required\": true,\n            \"schema\": {\n              \"type\": \"string\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned an artist\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    }\n  },\n  \"components\": {\n    \"schemas\": {\n      \"Artists\": {\n        \"type\": \"object\",\n        \"properties\": {\n          \"artist_name\": {\n            \"type\": \"string\"\n          },\n          \"artist_genre\": {\n            \"type\": \"string\"\n          },\n          \"albums_recorded\": {\n            \"type\": \"integer\"\n          }\n        }\n      }\n    },\n    \"securitySchemes\": {\n      \"BasicAuth\": {\n        \"type\": \"http\",\n        \"scheme\": \"basic\"\n      }\n    }\n  }\n}"
	jsonSpecInvalid := "{\n  \"openapi\": \"\",\n  \"info\": {\n    \"version\": \"1.0.0\",\n    \"title\": \"Simple API\",\n    \"description\": \"A simple API to illustrate OpenAPI concepts\"\n  },\n  \"servers\": [\n    {\n      \"url\": \"https://example.io/v1\"\n    }\n  ],\n  \"security\": [\n    {\n      \"BasicAuth\": []\n    }\n  ],\n  \"paths\": {\n    \"/artists\": {\n      \"get\": {\n        \"description\": \"Returns a list of artists\",\n        \"parameters\": [\n          {\n            \"name\": \"limit\",\n            \"in\": \"query\",\n            \"description\": \"Limits the number of items on a page\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          },\n          {\n            \"name\": \"offset\",\n            \"in\": \"query\",\n            \"description\": \"Specifies the page number of the artists to be displayed\",\n            \"schema\": {\n              \"type\": \"integer\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned a list of artists\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"array\",\n                  \"items\": {\n                    \"type\": \"object\",\n                    \"required\": [\n                      \"username\"\n                    ],\n                    \"properties\": {\n                      \"artist_name\": {\n                        \"type\": \"string\"\n                      },\n                      \"artist_genre\": {\n                        \"type\": \"string\"\n                      },\n                      \"albums_recorded\": {\n                        \"type\": \"integer\"\n                      },\n                      \"username\": {\n                        \"type\": \"string\"\n                      }\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      },\n      \"post\": {\n        \"description\": \"Lets a user post a new artist\",\n        \"requestBody\": {\n          \"required\": true,\n          \"content\": {\n            \"application/json\": {\n              \"schema\": {\n                \"type\": \"array\",\n                \"items\": {\n                  \"type\": \"object\",\n                  \"required\": [\n                    \"username\"\n                  ],\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    },\n                    \"username\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        },\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully created a new artist\"\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    },\n    \"/artists/{username}\": {\n      \"get\": {\n        \"description\": \"Obtain information about an artist from his or her unique username\",\n        \"parameters\": [\n          {\n            \"name\": \"username\",\n            \"in\": \"path\",\n            \"required\": true,\n            \"schema\": {\n              \"type\": \"string\"\n            }\n          }\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"Successfully returned an artist\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"artist_name\": {\n                      \"type\": \"string\"\n                    },\n                    \"artist_genre\": {\n                      \"type\": \"string\"\n                    },\n                    \"albums_recorded\": {\n                      \"type\": \"integer\"\n                    }\n                  }\n                }\n              }\n            }\n          },\n          \"400\": {\n            \"description\": \"Invalid request\",\n            \"content\": {\n              \"application/json\": {\n                \"schema\": {\n                  \"type\": \"object\",\n                  \"properties\": {\n                    \"message\": {\n                      \"type\": \"string\"\n                    }\n                  }\n                }\n              }\n            }\n          }\n        }\n      }\n    }\n  },\n  \"components\": {\n    \"securitySchemes\": {\n      \"BasicAuth\": {\n        \"type\": \"http\",\n        \"scheme\": \"basic\"\n      }\n    }\n  }\n}"
	yamlSpec := "openapi: 3.0.3\ninfo:\n  version: 1.0.0\n  title: Simple API\n  description: A simple API to illustrate OpenAPI concepts\n\nservers:\n  - url: https://example.io/v1\n\nsecurity:\n  - BasicAuth: []\n\npaths:\n  /artists:\n    get:\n      description: Returns a list of artists \n      parameters:\n        - name: limit\n          in: query\n          description: Limits the number of items on a page\n          schema:\n            type: integer\n        - name: offset\n          in: query\n          description: Specifies the page number of the artists to be displayed\n          schema:\n            type: integer\n      responses:\n        '200':\n          description: Successfully returned a list of artists\n          content:\n            application/json:\n              schema:\n                type: array\n                items:\n                  type: object\n                  required:\n                    - username\n                  properties:\n                    artist_name:\n                      type: string\n                    artist_genre:\n                        type: string\n                    albums_recorded:\n                        type: integer\n                    username:\n                        type: string\n        '400':\n          description: Invalid request\n          content:\n            application/json:\n              schema:\n                type: object \n                properties:\n                  message:\n                    type: string\n\n    post:\n      description: Lets a user post a new artist\n      requestBody:\n        required: true\n        content:\n          application/json:\n            schema:\n              type: array\n              items:\n                type: object\n                required:\n                  - username\n                properties:\n                  artist_name:\n                    type: string\n                  artist_genre:\n                      type: string\n                  albums_recorded:\n                      type: integer\n                  username:\n                      type: string\n      responses:\n        '200':\n          description: Successfully created a new artist\n        '400':\n          description: Invalid request\n          content:\n            application/json:\n              schema:\n                type: object \n                properties:\n                  message:\n                    type: string\n\n  /artists/{username}:\n    get:\n      description: Obtain information about an artist from his or her unique username\n      parameters:\n        - name: username\n          in: path\n          required: true\n          schema:\n            type: string\n          \n      responses:\n        '200':\n          description: Successfully returned an artist\n          content:\n            application/json:\n              schema:\n                type: object\n                properties:\n                  artist_name:\n                    type: string\n                  artist_genre:\n                    type: string\n                  albums_recorded:\n                    type: integer\n                \n        '400':\n          description: Invalid request\n          content:\n            application/json:\n              schema:\n                type: object \n                properties:\n                  message:\n                    type: string\n\ncomponents:\n  securitySchemes:\n    BasicAuth:\n      type: http\n      scheme: basic\n"

	v3, err := LoadAndValidateRawJSONSpecV3([]byte(jsonSpec))
	assert.NilError(t, err)
	wantProvidedSpec := &ProvidedSpec{
		Doc: v3,
	}

	v2, err := LoadAndValidateRawJSONSpecV3FromV2([]byte(jsonSpecV2))
	assert.NilError(t, err)
	wantProvidedSpecV2 := &ProvidedSpec{
		Doc: clearRefFromDoc(v2),
	}

	wantProvidedSpecWithRefAfterRemoveRef := &ProvidedSpec{
		Doc: &openapi3.T{
			Paths: openapi3.Paths{},
		},
	}
	err = json.Unmarshal([]byte(jsonSpecWithRefAfterRemoveRef), wantProvidedSpecWithRefAfterRemoveRef.Doc)
	assert.NilError(t, err)

	pathToPathID := map[string]string{
		"/artists": "1",
	}
	wantProvidedPathTrie := createPathTrie(pathToPathID)

	pathToPathIDv2 := map[string]string{
		"/dashboard/apiUsage/mostUsed": "1",
	}
	wantProvidedPathv2Trie := createPathTrie(pathToPathIDv2)
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
		wantProvidedSpec     *ProvidedSpec
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
			wantProvidedSpec:     wantProvidedSpec,
		},
		{
			name: "json spec v2",
			fields: fields{
				ProvidedSpec: nil,
			},
			args: args{
				providedSpec: []byte(jsonSpecV2),
				pathToPathID: pathToPathIDv2,
			},
			wantErr:              false,
			wantProvidedPathTrie: wantProvidedPathv2Trie,
			wantProvidedSpec:     wantProvidedSpecV2,
		},
		{
			name: "json spec with ref",
			fields: fields{
				ProvidedSpec: nil,
			},
			args: args{
				providedSpec: []byte(jsonSpecWithRef),
				pathToPathID: pathToPathID,
			},
			wantErr:              false,
			wantProvidedPathTrie: wantProvidedPathTrie,
			wantProvidedSpec:     wantProvidedSpecWithRefAfterRemoveRef,
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
			wantProvidedSpec:     wantProvidedSpec,
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
			wantProvidedSpec:     wantProvidedSpec,
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
			name: "invalid spec v3",
			fields: fields{
				ProvidedSpec: nil,
			},
			args: args{
				providedSpec: []byte(jsonSpecInvalid),
			},
			wantErr: true,
		},
		{
			name: "invalid spec v2",
			fields: fields{
				ProvidedSpec: nil,
			},
			args: args{
				providedSpec: []byte(jsonSpecV2Invalid),
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
				assert.DeepEqual(t, s.ProvidedSpec, tt.wantProvidedSpec, cmpopts.IgnoreUnexported(openapi3.Schema{}), cmpopts.IgnoreTypes(openapi3.ExtensionProps{}))
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

func Test_clearRefFromDoc(t *testing.T) {
	type args struct {
		doc *openapi3.T
	}
	tests := []struct {
		name string
		args args
		want *openapi3.T
	}{
		{
			name: "nil doc",
			args: args{
				doc: nil,
			},
			want: nil,
		},
		{
			name: "no paths",
			args: args{
				doc: &openapi3.T{
					Paths: openapi3.Paths{},
				},
			},
			want: &openapi3.T{
				Paths: openapi3.Paths{},
			},
		},
		{
			name: "multiple paths",
			args: args{
				doc: &openapi3.T{
					Paths: openapi3.Paths{
						"path1": &openapi3.PathItem{
							Get: createTestOperation().
								WithRequestBody(openapi3.NewRequestBody().
									WithJSONSchemaRef(openapi3.NewSchemaRef("array-string",
										openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema())))).Op,
						},
						"path2": &openapi3.PathItem{
							Put: createTestOperation().
								WithRequestBody(openapi3.NewRequestBody().
									WithJSONSchemaRef(openapi3.NewSchemaRef("array-int",
										openapi3.NewObjectSchema().WithItems(openapi3.NewInt64Schema())))).Op,
						},
					},
				},
			},
			want: &openapi3.T{
				Paths: openapi3.Paths{
					"path1": &openapi3.PathItem{
						Get: createTestOperation().
							WithRequestBody(openapi3.NewRequestBody().
								WithJSONSchemaRef(openapi3.NewSchemaRef("",
									openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema())))).Op,
					},
					"path2": &openapi3.PathItem{
						Put: createTestOperation().
							WithRequestBody(openapi3.NewRequestBody().
								WithJSONSchemaRef(openapi3.NewSchemaRef("",
									openapi3.NewObjectSchema().WithItems(openapi3.NewInt64Schema())))).Op,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.DeepEqual(t, clearRefFromDoc(tt.args.doc), tt.want, cmpopts.IgnoreUnexported(openapi3.Schema{}), cmpopts.IgnoreTypes(openapi3.ExtensionProps{}))
		})
	}
}

func Test_clearRefFromPathItem(t *testing.T) {
	type args struct {
		item *openapi3.PathItem
	}
	tests := []struct {
		name string
		args args
		want *openapi3.PathItem
	}{
		{
			name: "nil item",
			args: args{
				item: nil,
			},
			want: nil,
		},
		{
			name: "empty item",
			args: args{
				item: &openapi3.PathItem{},
			},
			want: &openapi3.PathItem{},
		},
		{
			name: "ref item",
			args: args{
				item: &openapi3.PathItem{
					Ref: "ref",
				},
			},
			want: &openapi3.PathItem{
				Ref: "",
			},
		},
		{
			name: "multiple operations",
			args: args{
				item: &openapi3.PathItem{
					Connect: createTestOperation().
						WithRequestBody(openapi3.NewRequestBody().
							WithJSONSchemaRef(openapi3.NewSchemaRef("array-string",
								openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema())))).Op,
					Delete: createTestOperation().
						WithRequestBody(openapi3.NewRequestBody().
							WithJSONSchemaRef(openapi3.NewSchemaRef("array-int",
								openapi3.NewObjectSchema().WithItems(openapi3.NewInt64Schema())))).Op,
				},
			},
			want: &openapi3.PathItem{
				Connect: createTestOperation().
					WithRequestBody(openapi3.NewRequestBody().
						WithJSONSchemaRef(openapi3.NewSchemaRef("",
							openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema())))).Op,
				Delete: createTestOperation().
					WithRequestBody(openapi3.NewRequestBody().
						WithJSONSchemaRef(openapi3.NewSchemaRef("",
							openapi3.NewObjectSchema().WithItems(openapi3.NewInt64Schema())))).Op,
			},
		},
		{
			name: "multiple parameters",
			args: args{
				item: &openapi3.PathItem{
					Parameters: openapi3.Parameters{
						{
							Ref:   "ref-path",
							Value: openapi3.NewPathParameter("path"),
						},
						{
							Ref:   "ref-query",
							Value: openapi3.NewQueryParameter("query"),
						},
					},
				},
			},
			want: &openapi3.PathItem{
				Parameters: openapi3.Parameters{
					{
						Ref:   "",
						Value: openapi3.NewPathParameter("path"),
					},
					{
						Ref:   "",
						Value: openapi3.NewQueryParameter("query"),
					},
				},
			},
		},
		{
			name: "multiple operations and parameters",
			args: args{
				item: &openapi3.PathItem{
					Connect: createTestOperation().
						WithRequestBody(openapi3.NewRequestBody().
							WithJSONSchemaRef(openapi3.NewSchemaRef("array-string",
								openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema())))).Op,
					Delete: createTestOperation().
						WithRequestBody(openapi3.NewRequestBody().
							WithJSONSchemaRef(openapi3.NewSchemaRef("array-int",
								openapi3.NewObjectSchema().WithItems(openapi3.NewInt64Schema())))).Op,
					Parameters: openapi3.Parameters{
						{
							Ref:   "ref-path",
							Value: openapi3.NewPathParameter("path"),
						},
						{
							Ref:   "ref-query",
							Value: openapi3.NewQueryParameter("query"),
						},
					},
				},
			},
			want: &openapi3.PathItem{
				Connect: createTestOperation().
					WithRequestBody(openapi3.NewRequestBody().
						WithJSONSchemaRef(openapi3.NewSchemaRef("",
							openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema())))).Op,
				Delete: createTestOperation().
					WithRequestBody(openapi3.NewRequestBody().
						WithJSONSchemaRef(openapi3.NewSchemaRef("",
							openapi3.NewObjectSchema().WithItems(openapi3.NewInt64Schema())))).Op,
				Parameters: openapi3.Parameters{
					{
						Ref:   "",
						Value: openapi3.NewPathParameter("path"),
					},
					{
						Ref:   "",
						Value: openapi3.NewQueryParameter("query"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.DeepEqual(t, clearRefFromPathItem(tt.args.item), tt.want, cmpopts.IgnoreUnexported(openapi3.Schema{}), cmpopts.IgnoreTypes(openapi3.ExtensionProps{}))
		})
	}
}

func Test_clearRefFromParameters(t *testing.T) {
	type args struct {
		parameters openapi3.Parameters
	}
	tests := []struct {
		name string
		args args
		want openapi3.Parameters
	}{
		{
			name: "nil parameters",
			args: args{
				parameters: nil,
			},
			want: nil,
		},
		{
			name: "empty parameters",
			args: args{
				parameters: openapi3.NewParameters(),
			},
			want: openapi3.NewParameters(),
		},
		{
			name: "multiple parameters",
			args: args{
				parameters: openapi3.Parameters{
					{
						Ref:   "ref-path",
						Value: openapi3.NewPathParameter("path"),
					},
					{
						Ref:   "ref-query",
						Value: openapi3.NewQueryParameter("query"),
					},
				},
			},
			want: openapi3.Parameters{
				{
					Ref:   "",
					Value: openapi3.NewPathParameter("path"),
				},
				{
					Ref:   "",
					Value: openapi3.NewQueryParameter("query"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.DeepEqual(t, clearRefFromParameters(tt.args.parameters), tt.want, cmpopts.IgnoreUnexported(openapi3.Schema{}), cmpopts.IgnoreTypes(openapi3.ExtensionProps{}))
		})
	}
}

func Test_clearRefFromOperation(t *testing.T) {
	type args struct {
		operation *openapi3.Operation
	}
	tests := []struct {
		name string
		args args
		want *openapi3.Operation
	}{
		{
			name: "nil operation",
			args: args{
				operation: nil,
			},
			want: nil,
		},
		{
			name: "empty operation",
			args: args{
				operation: openapi3.NewOperation(),
			},
			want: openapi3.NewOperation(),
		},
		{
			name: "multiple parameters",
			args: args{
				operation: &openapi3.Operation{
					Parameters: openapi3.Parameters{
						{
							Ref:   "ref-path",
							Value: openapi3.NewPathParameter("path"),
						},
						{
							Ref:   "ref-query",
							Value: openapi3.NewQueryParameter("query"),
						},
					},
				},
			},
			want: &openapi3.Operation{
				Parameters: openapi3.Parameters{
					{
						Ref:   "",
						Value: openapi3.NewPathParameter("path"),
					},
					{
						Ref:   "",
						Value: openapi3.NewQueryParameter("query"),
					},
				},
			},
		},
		{
			name: "multiple responses",
			args: args{
				operation: &openapi3.Operation{
					Responses: openapi3.Responses{
						"response1": {
							Ref: "ref-response1",
							Value: openapi3.NewResponse().
								WithJSONSchemaRef(openapi3.NewSchemaRef("ref-int",
									openapi3.NewObjectSchema().WithItems(openapi3.NewInt64Schema()))),
						},
						"response2": {
							Ref: "ref-response2",
							Value: openapi3.NewResponse().
								WithJSONSchemaRef(openapi3.NewSchemaRef("ref-string",
									openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema()))),
						},
					},
				},
			},
			want: &openapi3.Operation{
				Responses: openapi3.Responses{
					"response1": {
						Ref: "",
						Value: openapi3.NewResponse().
							WithJSONSchemaRef(openapi3.NewSchemaRef("",
								openapi3.NewObjectSchema().WithItems(openapi3.NewInt64Schema()))),
					},
					"response2": {
						Ref: "",
						Value: openapi3.NewResponse().
							WithJSONSchemaRef(openapi3.NewSchemaRef("",
								openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema()))),
					},
				},
			},
		},
		{
			name: "request body",
			args: args{
				operation: &openapi3.Operation{
					RequestBody: &openapi3.RequestBodyRef{
						Ref: "ref-request-body",
						Value: openapi3.NewRequestBody().
							WithJSONSchemaRef(openapi3.NewSchemaRef("",
								openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema()))),
					},
				},
			},
			want: &openapi3.Operation{
				Responses: openapi3.Responses{
					"response1": {
						Ref: "",
						Value: openapi3.NewResponse().
							WithJSONSchemaRef(openapi3.NewSchemaRef("",
								openapi3.NewObjectSchema().WithItems(openapi3.NewInt64Schema()))),
					},
					"response2": {
						Ref: "",
						Value: openapi3.NewResponse().
							WithJSONSchemaRef(openapi3.NewSchemaRef("",
								openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema()))),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.DeepEqual(t, clearRefFromOperation(tt.args.operation), tt.want, cmpopts.IgnoreUnexported(openapi3.Schema{}), cmpopts.IgnoreTypes(openapi3.ExtensionProps{}))
		})
	}
}

func Test_clearRefFromResponses(t *testing.T) {
	type args struct {
		responses openapi3.Responses
	}
	tests := []struct {
		name string
		args args
		want openapi3.Responses
	}{
		{
			name: "nil responses",
			args: args{
				responses: nil,
			},
			want: nil,
		},
		{
			name: "empty responses",
			args: args{
				responses: openapi3.NewResponses(),
			},
			want: openapi3.NewResponses(),
		},
		{
			name: "multiple responses",
			args: args{
				responses: openapi3.Responses{
					"response1": {
						Ref: "ref-response1",
						Value: openapi3.NewResponse().
							WithJSONSchemaRef(openapi3.NewSchemaRef("ref-int",
								openapi3.NewObjectSchema().WithItems(openapi3.NewInt64Schema()))),
					},
					"response2": {
						Ref: "ref-response2",
						Value: openapi3.NewResponse().
							WithJSONSchemaRef(openapi3.NewSchemaRef("ref-string",
								openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema()))),
					},
				},
			},
			want: openapi3.Responses{
				"response1": {
					Ref: "",
					Value: openapi3.NewResponse().
						WithJSONSchemaRef(openapi3.NewSchemaRef("",
							openapi3.NewObjectSchema().WithItems(openapi3.NewInt64Schema()))),
				},
				"response2": {
					Ref: "",
					Value: openapi3.NewResponse().
						WithJSONSchemaRef(openapi3.NewSchemaRef("",
							openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema()))),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.DeepEqual(t, clearRefFromResponses(tt.args.responses), tt.want, cmpopts.IgnoreUnexported(openapi3.Schema{}), cmpopts.IgnoreTypes(openapi3.ExtensionProps{}))
		})
	}
}

func Test_clearRefFromRequestBody(t *testing.T) {
	type args struct {
		requestBodyRef *openapi3.RequestBodyRef
	}
	tests := []struct {
		name string
		args args
		want *openapi3.RequestBodyRef
	}{
		{
			name: "nil requestBodyRef",
			args: args{
				requestBodyRef: nil,
			},
			want: nil,
		},
		{
			name: "empty requestBodyRef",
			args: args{
				requestBodyRef: &openapi3.RequestBodyRef{},
			},
			want: &openapi3.RequestBodyRef{},
		},
		{
			name: "sanity requestBodyRef",
			args: args{
				requestBodyRef: &openapi3.RequestBodyRef{
					Ref: "ref",
					Value: openapi3.NewRequestBody().
						WithJSONSchemaRef(openapi3.NewSchemaRef("ref-string",
							openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema()))),
				},
			},
			want: &openapi3.RequestBodyRef{
				Ref: "",
				Value: openapi3.NewRequestBody().
					WithJSONSchemaRef(openapi3.NewSchemaRef("",
						openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema()))),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.DeepEqual(t, clearRefFromRequestBody(tt.args.requestBodyRef), tt.want, cmpopts.IgnoreUnexported(openapi3.Schema{}), cmpopts.IgnoreTypes(openapi3.ExtensionProps{}))
		})
	}
}

func Test_clearRefFromRequestBodyRef(t *testing.T) {
	type args struct {
		requestBody *openapi3.RequestBody
	}
	tests := []struct {
		name string
		args args
		want *openapi3.RequestBody
	}{
		{
			name: "nil RequestBody",
			args: args{
				requestBody: nil,
			},
			want: nil,
		},
		{
			name: "empty RequestBody",
			args: args{
				requestBody: &openapi3.RequestBody{},
			},
			want: &openapi3.RequestBody{},
		},
		{
			name: "multiple contents",
			args: args{
				requestBody: openapi3.NewRequestBody().
					WithSchemaRef(openapi3.NewSchemaRef("ref-string",
						openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema())),
						[]string{"content1", "content2"}),
			},
			want: openapi3.NewRequestBody().
				WithSchemaRef(openapi3.NewSchemaRef("",
					openapi3.NewObjectSchema().WithItems(openapi3.NewStringSchema())),
					[]string{"content1", "content2"}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.DeepEqual(t, clearRefFromRequestBodyRef(tt.args.requestBody), tt.want, cmpopts.IgnoreUnexported(openapi3.Schema{}), cmpopts.IgnoreTypes(openapi3.ExtensionProps{}))
		})
	}
}
