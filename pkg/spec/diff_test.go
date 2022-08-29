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
	"net/http"
	"reflect"
	"testing"

	spec "github.com/getkin/kin-openapi/openapi3"
	"github.com/google/go-cmp/cmp/cmpopts"
	uuid "github.com/satori/go.uuid"
	"gotest.tools/assert"

	"github.com/openclarity/speculator/pkg/pathtrie"
)

var Data = &HTTPInteractionData{
	ReqBody:  req1,
	RespBody: res1,
	ReqHeaders: map[string]string{
		contentTypeHeaderName: mediaTypeApplicationJSON,
	},
	RespHeaders: map[string]string{
		contentTypeHeaderName: mediaTypeApplicationJSON,
	},
	statusCode: 200,
}

var DataWithAuth = &HTTPInteractionData{
	ReqBody:  req1,
	RespBody: res1,
	ReqHeaders: map[string]string{
		contentTypeHeaderName:       mediaTypeApplicationJSON,
		authorizationTypeHeaderName: BearerAuthPrefix,
	},
	RespHeaders: map[string]string{
		contentTypeHeaderName: mediaTypeApplicationJSON,
	},
	statusCode: 200,
}

var Data2 = &HTTPInteractionData{
	ReqBody:  req2,
	RespBody: res2,
	ReqHeaders: map[string]string{
		contentTypeHeaderName: mediaTypeApplicationJSON,
	},
	RespHeaders: map[string]string{
		contentTypeHeaderName: mediaTypeApplicationJSON,
	},
	statusCode: 200,
}

var DataCombined = &HTTPInteractionData{
	ReqBody:  combinedReq,
	RespBody: combinedRes,
	ReqHeaders: map[string]string{
		contentTypeHeaderName: mediaTypeApplicationJSON,
	},
	RespHeaders: map[string]string{
		contentTypeHeaderName: mediaTypeApplicationJSON,
	},
	statusCode: 200,
}

var DiffOAuthScopes = []string{"superadmin", "write:all_your_base"}

func createTelemetry(reqID, method, path, host, statusCode string, reqBody, respBody string) *Telemetry {
	return &Telemetry{
		RequestID: reqID,
		Scheme:    "",
		Request: &Request{
			Method: method,
			Path:   path,
			Host:   host,
			Common: &Common{
				Version: "",
				Headers: []*Header{
					{
						Key:   contentTypeHeaderName,
						Value: mediaTypeApplicationJSON,
					},
				},
				Body:          []byte(reqBody),
				TruncatedBody: false,
			},
		},
		Response: &Response{
			StatusCode: statusCode,
			Common: &Common{
				Version: "",
				Headers: []*Header{
					{
						Key:   contentTypeHeaderName,
						Value: mediaTypeApplicationJSON,
					},
				},
				Body:          []byte(respBody),
				TruncatedBody: false,
			},
		},
	}
}

func createTelemetryWithSecurity(reqID, method, path, host, statusCode string, reqBody, respBody string) *Telemetry {
	bearerToken, _ := generateDefaultOAuthToken(DiffOAuthScopes)

	telemetry := createTelemetry(reqID, method, path, host, statusCode, reqBody, respBody)
	telemetry.Request.Common.Headers = append(telemetry.Request.Common.Headers, &Header{
		Key:   authorizationTypeHeaderName,
		Value: BearerAuthPrefix + bearerToken,
	})
	return telemetry
}

func TestSpec_DiffTelemetry_Reconstructed(t *testing.T) {
	reqID := "req-id"
	reqUUID := uuid.NewV5(uuid.Nil, reqID)
	specUUID := uuid.NewV5(uuid.Nil, "spec-id")
	bearerToken, _ := generateDefaultOAuthToken(DiffOAuthScopes)
	DataWithAuth.ReqHeaders[authorizationTypeHeaderName] = BearerAuthPrefix + bearerToken
	type fields struct {
		ID               uuid.UUID
		ApprovedSpec     *ApprovedSpec
		LearningSpec     *LearningSpec
		ApprovedPathTrie pathtrie.PathTrie
	}
	type args struct {
		telemetry *Telemetry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *APIDiff
		wantErr bool
	}{
		{
			name: "No diff",
			fields: fields{
				ID: specUUID,
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*spec.PathItem{
						"/api": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					},
				},
				ApprovedPathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api", "host", "200", Data.ReqBody, Data.RespBody),
			},
			want: &APIDiff{
				Type:             DiffTypeNoDiff,
				Path:             "/api",
				OriginalPathItem: nil,
				ModifiedPathItem: nil,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "Security diff",
			fields: fields{
				ID: specUUID,
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*spec.PathItem{
						"/api": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					},
				},
				ApprovedPathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetryWithSecurity(reqID, http.MethodGet, "/api", "host", "200", Data.ReqBody, Data.RespBody),
			},
			want: &APIDiff{
				Type:             DiffTypeGeneralDiff,
				Path:             "/api",
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, DataWithAuth).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "New PathItem",
			fields: fields{
				ID: specUUID,
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*spec.PathItem{
						"/api": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					},
				},
				ApprovedPathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api/new", "host", "200", Data.ReqBody, Data.RespBody),
			},
			want: &APIDiff{
				Type:             DiffTypeShadowDiff,
				Path:             "/api/new",
				OriginalPathItem: nil,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "New Operation",
			fields: fields{
				ID: specUUID,
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*spec.PathItem{
						"/api": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					},
				},
				ApprovedPathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodPost, "/api", "host", "200", req2, res2),
			},
			want: &APIDiff{
				Type:             DiffTypeShadowDiff,
				Path:             "/api",
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).WithOperation(http.MethodPost, NewOperation(t, Data2).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "Changed Operation",
			fields: fields{
				ID: specUUID,
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*spec.PathItem{
						"/api": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					},
				},
				ApprovedPathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api", "host", "200", req2, res2),
			},
			want: &APIDiff{
				Type:             DiffTypeGeneralDiff,
				Path:             "/api",
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "Parameterized path",
			fields: fields{
				ID: specUUID,
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*spec.PathItem{
						"/api/{my-param}": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					},
				},
				ApprovedPathTrie: createPathTrie(map[string]string{
					"/api/{my-param}": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api/2", "host", "200", req2, res2),
			},
			want: &APIDiff{
				Type:             DiffTypeGeneralDiff,
				Path:             "/api/{my-param}",
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "Parameterized path but also exact path",
			fields: fields{
				ID: specUUID,
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*spec.PathItem{
						"/api/1": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					},
				},
				ApprovedPathTrie: createPathTrie(map[string]string{
					"/api/{my-param}": "1",
					"/api/1":          "2",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api/1", "host", "200", req2, res2),
			},
			want: &APIDiff{
				Type:             DiffTypeGeneralDiff,
				Path:             "/api/1",
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Spec{
				SpecInfo: SpecInfo{
					ID:               tt.fields.ID,
					ApprovedSpec:     tt.fields.ApprovedSpec,
					LearningSpec:     tt.fields.LearningSpec,
					ApprovedPathTrie: tt.fields.ApprovedPathTrie,
				},
				OpGenerator: CreateTestNewOperationGenerator(),
			}
			got, err := s.DiffTelemetry(tt.args.telemetry, SpecSourceReconstructed)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiffTelemetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreUnexported(spec.Schema{}), cmpopts.IgnoreTypes(spec.ExtensionProps{}))
		})
	}
}

func TestSpec_DiffTelemetry_Provided(t *testing.T) {
	reqID := "req-id"
	reqUUID := uuid.NewV5(uuid.Nil, reqID)
	specUUID := uuid.NewV5(uuid.Nil, "spec-id")
	bearerToken, _ := generateDefaultOAuthToken(DiffOAuthScopes)
	DataWithAuth.ReqHeaders[authorizationTypeHeaderName] = BearerAuthPrefix + bearerToken
	type fields struct {
		ID               uuid.UUID
		ProvidedSpec     *ProvidedSpec
		ProvidedPathTrie pathtrie.PathTrie
	}
	type args struct {
		telemetry *Telemetry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *APIDiff
		wantErr bool
	}{
		{
			name: "No diff",
			fields: fields{
				ID: specUUID,
				ProvidedSpec: &ProvidedSpec{
					Doc: &spec.T{
						Paths: map[string]*spec.PathItem{
							"/api": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						},
					},
				},
				ProvidedPathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api", "host", "200", Data.ReqBody, Data.RespBody),
			},
			want: &APIDiff{
				Type:             DiffTypeNoDiff,
				Path:             "/api",
				OriginalPathItem: nil,
				ModifiedPathItem: nil,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "Security diff",
			fields: fields{
				ID: specUUID,
				ProvidedSpec: &ProvidedSpec{
					Doc: &spec.T{
						Paths: map[string]*spec.PathItem{
							"/api": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						},
					},
				},
				ProvidedPathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetryWithSecurity(reqID, http.MethodGet, "/api", "host", "200", Data.ReqBody, Data.RespBody),
			},
			want: &APIDiff{
				Type:             DiffTypeGeneralDiff,
				Path:             "/api",
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, DataWithAuth).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "New PathItem",
			fields: fields{
				ID: specUUID,
				ProvidedSpec: &ProvidedSpec{
					Doc: &spec.T{
						Paths: map[string]*spec.PathItem{
							"/api": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						},
					},
				},
				ProvidedPathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api/new", "host", "200", Data.ReqBody, Data.RespBody),
			},
			want: &APIDiff{
				Type:             DiffTypeShadowDiff,
				Path:             "/api/new",
				OriginalPathItem: nil,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "New Operation",
			fields: fields{
				ID: specUUID,
				ProvidedSpec: &ProvidedSpec{
					Doc: &spec.T{
						Paths: map[string]*spec.PathItem{
							"/api": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						},
					},
				},
				ProvidedPathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodPost, "/api", "host", "200", req2, res2),
			},
			want: &APIDiff{
				Type:             DiffTypeShadowDiff,
				Path:             "/api",
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).WithOperation(http.MethodPost, NewOperation(t, Data2).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "Changed Operation",
			fields: fields{
				ID: specUUID,
				ProvidedSpec: &ProvidedSpec{
					Doc: &spec.T{
						Paths: map[string]*spec.PathItem{
							"/api": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						},
					},
				},
				ProvidedPathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api", "host", "200", req2, res2),
			},
			want: &APIDiff{
				Type:             DiffTypeGeneralDiff,
				Path:             "/api",
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "test remove base path + parametrized path",
			fields: fields{
				ID: specUUID,
				ProvidedSpec: &ProvidedSpec{
					Doc: &spec.T{
						Servers: spec.Servers{
							{
								URL: "https://example.com/api",
							},
						},
						Paths: map[string]*spec.PathItem{
							"/foo/{param}": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						},
					},
				},
				ProvidedPathTrie: createPathTrie(map[string]string{
					"/foo/{param}": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api/foo/bar", "host", "200", req2, res2),
			},
			want: &APIDiff{
				Type:             DiffTypeGeneralDiff,
				Path:             "/api/foo/{param}",
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "test base path = / (default)",
			fields: fields{
				ID: specUUID,
				ProvidedSpec: &ProvidedSpec{
					Doc: &spec.T{
						Servers: spec.Servers{
							{
								URL: "https://example.com/",
							},
						},
						Paths: map[string]*spec.PathItem{
							"/foo/bar": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						},
					},
				},
				ProvidedPathTrie: createPathTrie(map[string]string{
					"/foo/bar": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/foo/bar", "host", "200", req2, res2),
			},
			want: &APIDiff{
				Type:             DiffTypeGeneralDiff,
				Path:             "/foo/bar",
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "Parameterized path",
			fields: fields{
				ID: specUUID,
				ProvidedSpec: &ProvidedSpec{
					Doc: &spec.T{
						Servers: spec.Servers{
							{
								URL: "https://example.com/",
							},
						},
						Paths: map[string]*spec.PathItem{
							"/api/{my-param}": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						},
					},
				},
				ProvidedPathTrie: createPathTrie(map[string]string{
					"/api/{my-param}": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api/2", "host", "200", req2, res2),
			},
			want: &APIDiff{
				Type:             DiffTypeGeneralDiff,
				Path:             "/api/{my-param}",
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "Parameterized path but also exact path",
			fields: fields{
				ID: specUUID,
				ProvidedSpec: &ProvidedSpec{
					Doc: &spec.T{
						Servers: spec.Servers{
							{
								URL: "https://example.com/",
							},
						},
						Paths: map[string]*spec.PathItem{
							"/api/1": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						},
					},
				},
				ProvidedPathTrie: createPathTrie(map[string]string{
					"/api/{my-param}": "1",
					"/api/1":          "2",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api/1", "host", "200", req2, res2),
			},
			want: &APIDiff{
				Type:             DiffTypeGeneralDiff,
				Path:             "/api/1",
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "Deprecated API expected Zombie API diff",
			fields: fields{
				ID: specUUID,
				ProvidedSpec: &ProvidedSpec{
					Doc: &spec.T{
						Paths: map[string]*spec.PathItem{
							"/api": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Deprecated().Op).PathItem,
						},
					},
				},
				ProvidedPathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api", "host", "200", Data.ReqBody, Data.RespBody),
			},
			want: &APIDiff{
				Type:   DiffTypeZombieDiff,
				Path:   "/api",

				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Deprecated().Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "Deprecated and simple diff expected Zombie API diff",
			fields: fields{
				ID: specUUID,
				ProvidedSpec: &ProvidedSpec{
					Doc: &spec.T{
						Paths: map[string]*spec.PathItem{
							"/api": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Deprecated().Op).PathItem,
						},
					},
				},
				ProvidedPathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api", "host", "200", req2, res2),
			},
			want: &APIDiff{
				Type:             DiffTypeZombieDiff,
				Path:             "/api",
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Deprecated().Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Spec{
				SpecInfo: SpecInfo{
					ID:               tt.fields.ID,
					ProvidedSpec:     tt.fields.ProvidedSpec,
					ProvidedPathTrie: tt.fields.ProvidedPathTrie,
				},
				OpGenerator: CreateTestNewOperationGenerator(),
			}
			got, err := s.DiffTelemetry(tt.args.telemetry, SpecSourceProvided)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiffTelemetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreUnexported(spec.Schema{}), cmpopts.IgnoreTypes(spec.ExtensionProps{}))
		})
	}
}

func createPathTrie(pathToValue map[string]string) pathtrie.PathTrie {
	pt := pathtrie.New()
	for path, value := range pathToValue {
		pt.Insert(path, value)
	}
	return pt
}

func Test_keepResponseStatusCode(t *testing.T) {
	type args struct {
		op               *spec.Operation
		statusCodeToKeep string
	}
	tests := []struct {
		name string
		args args
		want *spec.Operation
	}{
		{
			name: "keep 1 remove 1",
			args: args{
				op: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("keep")).
					WithResponse(300, spec.NewResponse().WithDescription("delete")).Op,
				statusCodeToKeep: "200",
			},
			want: createTestOperation().WithResponse(200, spec.NewResponse().WithDescription("keep")).Op,
		},
		{
			name: "status code to keep not found - remove all",
			args: args{
				op: createTestOperation().
					WithResponse(202, spec.NewResponse().WithDescription("delete")).
					WithResponse(300, spec.NewResponse().WithDescription("delete")).Op,
				statusCodeToKeep: "200",
			},
			want: spec.NewOperation(),
		},
		{
			name: "status code to keep not found - remove all keep default response",
			args: args{
				op: createTestOperation().
					WithResponse(202, spec.NewResponse().WithDescription("delete")).
					WithResponse(300, spec.NewResponse().WithDescription("delete")).
					WithResponse(0, spec.NewResponse().WithDescription("keep-default")).Op,
				statusCodeToKeep: "200",
			},
			want: createTestOperation().
				WithResponse(0, spec.NewResponse().WithDescription("keep-default")).Op,
		},
		{
			name: "only status code to keep is found",
			args: args{
				op: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("keep")).Op,
				statusCodeToKeep: "200",
			},
			want: createTestOperation().
				WithResponse(200, spec.NewResponse().WithDescription("keep")).Op,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keepResponseStatusCode(tt.args.op, tt.args.statusCodeToKeep)
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreUnexported(spec.Schema{}), cmpopts.IgnoreTypes(spec.ExtensionProps{}))
		})
	}
}

func Test_calculateOperationDiff(t *testing.T) {
	type args struct {
		specOp            *spec.Operation
		telemetryOp       *spec.Operation
		telemetryResponse *Response
	}
	tests := []struct {
		name    string
		args    args
		want    *operationDiff
		wantErr bool
	}{
		{
			name: "no diff",
			args: args{
				specOp: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("test")).
					WithParameter(spec.NewHeaderParameter("header")).Op,
				telemetryOp: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("test")).
					WithParameter(spec.NewHeaderParameter("header")).Op,
				telemetryResponse: &Response{
					StatusCode: "200",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "no diff - parameters are not sorted",
			args: args{
				specOp: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("test")).
					WithParameter(spec.NewHeaderParameter("header2")).
					WithParameter(spec.NewHeaderParameter("header1")).Op,
				telemetryOp: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("test")).
					WithParameter(spec.NewHeaderParameter("header1")).
					WithParameter(spec.NewHeaderParameter("header2")).Op,
				telemetryResponse: &Response{
					StatusCode: "200",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "no diff - existing response should be removed",
			args: args{
				specOp: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("test")).
					WithResponse(300, spec.NewResponse().WithDescription("remove")).
					WithParameter(spec.NewHeaderParameter("header")).Op,
				telemetryOp: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("test")).
					WithParameter(spec.NewHeaderParameter("header")).Op,
				telemetryResponse: &Response{
					StatusCode: "200",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "no diff",
			args: args{
				specOp: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("test")).
					WithResponse(403, spec.NewResponse().WithDescription("keep")).
					WithParameter(spec.NewHeaderParameter("header")).Op,
				telemetryOp: createTestOperation().
					WithResponse(403, spec.NewResponse().WithDescription("keep")).
					WithParameter(spec.NewHeaderParameter("header")).Op,
				telemetryResponse: &Response{
					StatusCode: "403",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "has diff",
			args: args{
				specOp: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("test")).
					WithParameter(spec.NewHeaderParameter("header")).Op,
				telemetryOp: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("test")).
					WithParameter(spec.NewHeaderParameter("new-header")).Op,
				telemetryResponse: &Response{
					StatusCode: "200",
				},
			},
			want: &operationDiff{
				OriginalOperation: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("test")).
					WithParameter(spec.NewHeaderParameter("header")).Op,
				ModifiedOperation: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("test")).
					WithParameter(spec.NewHeaderParameter("new-header")).Op,
			},
			wantErr: false,
		},
		{
			name: "has diff in param and not in response",
			args: args{
				specOp: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("200")).
					WithResponse(403, spec.NewResponse().WithDescription("403")).
					WithParameter(spec.NewHeaderParameter("header")).Op,
				telemetryOp: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("200")).
					WithParameter(spec.NewHeaderParameter("new-header")).Op,
				telemetryResponse: &Response{
					StatusCode: "200",
				},
			},
			want: &operationDiff{
				OriginalOperation: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("200")).
					WithParameter(spec.NewHeaderParameter("header")).Op,
				ModifiedOperation: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("200")).
					WithParameter(spec.NewHeaderParameter("new-header")).Op,
			},
			wantErr: false,
		},
		{
			name: "has diff in response",
			args: args{
				specOp: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("200")).
					WithResponse(403, spec.NewResponse().WithDescription("403")).
					WithParameter(spec.NewHeaderParameter("header")).Op,
				telemetryOp: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("new-200")).
					WithParameter(spec.NewHeaderParameter("new-header")).Op,
				telemetryResponse: &Response{
					StatusCode: "200",
				},
			},
			want: &operationDiff{
				OriginalOperation: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("200")).
					WithParameter(spec.NewHeaderParameter("header")).Op,
				ModifiedOperation: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("new-200")).
					WithParameter(spec.NewHeaderParameter("new-header")).Op,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateOperationDiff(tt.args.specOp, tt.args.telemetryOp, tt.args.telemetryResponse)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateOperationDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreUnexported(spec.Schema{}), cmpopts.IgnoreTypes(spec.ExtensionProps{}))
		})
	}
}

func Test_compareObjects(t *testing.T) {
	type args struct {
		obj1 interface{}
		obj2 interface{}
	}
	tests := []struct {
		name        string
		args        args
		wantHasDiff bool
		wantErr     bool
	}{
		{
			name: "no diff",
			args: args{
				obj1: createTestOperation().WithParameter(spec.NewHeaderParameter("test")).Op,
				obj2: createTestOperation().WithParameter(spec.NewHeaderParameter("test")).Op,
			},
			wantHasDiff: false,
			wantErr:     false,
		},
		{
			name: "has diff (compare only Responses)",
			args: args{
				obj1: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("test")).Op.Responses,
				obj2: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("diff")).Op.Responses,
			},
			wantHasDiff: true,
			wantErr:     false,
		},
		{
			name: "has diff (different objects - Operation vs Responses)",
			args: args{
				obj1: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("test")).Op,
				obj2: createTestOperation().
					WithResponse(200, spec.NewResponse().WithDescription("diff")).Op.Responses,
			},
			wantHasDiff: true,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHasDiff, err := compareObjects(tt.args.obj1, tt.args.obj2)
			if (err != nil) != tt.wantErr {
				t.Errorf("compareObjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHasDiff != tt.wantHasDiff {
				t.Errorf("compareObjects() gotHasDiff = %v, want %v", gotHasDiff, tt.wantHasDiff)
			}
		})
	}
}

func Test_sortParameters(t *testing.T) {
	type args struct {
		operation *spec.Operation
	}
	tests := []struct {
		name string
		args args
		want *spec.Operation
	}{
		{
			name: "already sorted",
			args: args{
				operation: createTestOperation().
					WithParameter(spec.NewHeaderParameter("1")).
					WithParameter(spec.NewHeaderParameter("2")).Op,
			},
			want: createTestOperation().
				WithParameter(spec.NewHeaderParameter("1")).
				WithParameter(spec.NewHeaderParameter("2")).Op,
		},
		{
			name: "sort is needed - sort by 'name'",
			args: args{
				operation: createTestOperation().
					WithParameter(spec.NewHeaderParameter("3")).
					WithParameter(spec.NewHeaderParameter("1")).
					WithParameter(spec.NewHeaderParameter("2")).Op,
			},
			want: createTestOperation().
				WithParameter(spec.NewHeaderParameter("1")).
				WithParameter(spec.NewHeaderParameter("2")).
				WithParameter(spec.NewHeaderParameter("3")).Op,
		},
		{
			name: "param name is the same - sort by 'in'",
			args: args{
				operation: createTestOperation().
					WithParameter(spec.NewHeaderParameter("1")).
					WithParameter(spec.NewCookieParameter("1")).Op,
			},
			want: createTestOperation().
				WithParameter(spec.NewCookieParameter("1")).
				WithParameter(spec.NewHeaderParameter("1")).Op,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sortParameters(tt.args.operation); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortParameters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasBasePath(t *testing.T) {
	type args struct {
		basePath string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty base path",
			args: args{
				basePath: "",
			},
			want: false,
		},
		{
			name: "slash base path",
			args: args{
				basePath: "/",
			},
			want: false,
		},
		{
			name: "base path exist",
			args: args{
				basePath: "/api",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasBasePath(tt.args.basePath); got != tt.want {
				t.Errorf("hasBasePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addBasePathIfNeeded(t *testing.T) {
	type args struct {
		basePath string
		path     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no need to add base path",
			args: args{
				basePath: "",
				path:     "/no-need",
			},
			want: "/no-need",
		},
		{
			name: "need to add base path",
			args: args{
				basePath: "/api",
				path:     "/need",
			},
			want: "/api/need",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addBasePathIfNeeded(tt.args.basePath, tt.args.path); got != tt.want {
				t.Errorf("addBasePathIfNeeded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_trimBasePathIfNeeded(t *testing.T) {
	type args struct {
		basePath string
		path     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no need to trim base path",
			args: args{
				basePath: "",
				path:     "/no-need",
			},
			want: "/no-need",
		},
		{
			name: "need to trim base path",
			args: args{
				basePath: "/api",
				path:     "/api/need",
			},
			want: "/need",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimBasePathIfNeeded(tt.args.basePath, tt.args.path); got != tt.want {
				t.Errorf("trimBasePathIfNeeded() = %v, want %v", got, tt.want)
			}
		})
	}
}
