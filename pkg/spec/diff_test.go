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

	"github.com/go-openapi/spec"
	uuid "github.com/satori/go.uuid"

	"github.com/apiclarity/speculator/pkg/pathtrie"
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
		authorizationTypeHeaderName: BearerAuthPrefix + "token",
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

func createTelemetry(reqID, method, path, host, statusCode string, reqBody, respBody []byte) *SCNTelemetry {
	return &SCNTelemetry{
		RequestID: reqID,
		Scheme:    "",
		SCNTRequest: SCNTRequest{
			Method: method,
			Path:   path,
			Host:   host,
			SCNTCommon: SCNTCommon{
				Version:       "",
				Headers:       [][2]string{{contentTypeHeaderName, mediaTypeApplicationJSON}},
				Body:          reqBody,
				TruncatedBody: false,
			},
		},
		SCNTResponse: SCNTResponse{
			StatusCode: statusCode,
			SCNTCommon: SCNTCommon{
				Version:       "",
				Headers:       [][2]string{{contentTypeHeaderName, mediaTypeApplicationJSON}},
				Body:          respBody,
				TruncatedBody: false,
			},
		},
	}
}

func createTelemetryWithSecurity(reqID, method, path, host, statusCode string, reqBody, respBody []byte) *SCNTelemetry {
	telemetry := createTelemetry(reqID, method, path, host, statusCode, reqBody, respBody)
	telemetry.SCNTRequest.Headers = append(telemetry.SCNTRequest.Headers, [2]string{authorizationTypeHeaderName, BearerAuthPrefix + "token"})
	return telemetry
}

func TestSpec_DiffTelemetry_Reconstructed(t *testing.T) {
	reqID := "req-id"
	reqUUID := uuid.NewV5(uuid.Nil, reqID)
	specUUID := uuid.NewV5(uuid.Nil, "spec-id")
	type fields struct {
		ID           uuid.UUID
		ApprovedSpec *ApprovedSpec
		LearningSpec *LearningSpec
		PathTrie     pathtrie.PathTrie
	}
	type args struct {
		telemetry *SCNTelemetry
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
				PathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api", "host", "200", []byte(Data.ReqBody), []byte(Data.RespBody)),
			},
			want: &APIDiff{
				Type:             DiffTypeNoDiff,
				Path:             "/api",
				PathID:           "1",
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
				PathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetryWithSecurity(reqID, http.MethodGet, "/api", "host", "200", []byte(Data.ReqBody), []byte(Data.RespBody)),
			},
			want: &APIDiff{
				Type:   DiffTypeChanged,
				Path:   "/api",
				PathID: "1",
				// when there is no diff in response, we don’t include 'Produces' in the diff logic so we need to clear produces here
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, clearProduces(NewOperation(t, Data).Op)).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, clearProduces(NewOperation(t, DataWithAuth).Op)).PathItem,
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
				PathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api/new", "host", "200", []byte(Data.ReqBody), []byte(Data.RespBody)),
			},
			want: &APIDiff{
				Type:             DiffTypeNew,
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
				PathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodPost, "/api", "host", "200", []byte(req2), []byte(res2)),
			},
			want: &APIDiff{
				Type:             DiffTypeChanged,
				Path:             "/api",
				PathID:           "1",
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
				PathTrie: createPathTrie(map[string]string{
					"/api": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api", "host", "200", []byte(req2), []byte(res2)),
			},
			want: &APIDiff{
				Type:             DiffTypeChanged,
				Path:             "/api",
				PathID:           "1",
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
				PathTrie: createPathTrie(map[string]string{
					"/api/{my-param}": "1",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api/2", "host", "200", []byte(req2), []byte(res2)),
			},
			want: &APIDiff{
				Type:             DiffTypeChanged,
				Path:             "/api/{my-param}",
				PathID:           "1",
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
				PathTrie: createPathTrie(map[string]string{
					"/api/{my-param}": "1",
					"/api/1":          "2",
				}),
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api/1", "host", "200", []byte(req2), []byte(res2)),
			},
			want: &APIDiff{
				Type:             DiffTypeChanged,
				Path:             "/api/1",
				PathID:           "2",
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
				ID:           tt.fields.ID,
				ApprovedSpec: tt.fields.ApprovedSpec,
				LearningSpec: tt.fields.LearningSpec,
				PathTrie:     tt.fields.PathTrie,
			}
			got, err := s.DiffTelemetry(tt.args.telemetry, DiffSourceReconstructed)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiffTelemetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DiffTelemetry() got = %s, want %s", marshal(got), marshal(tt.want))
			}
		})
	}
}

func TestSpec_DiffTelemetry_Provided(t *testing.T) {
	reqID := "req-id"
	reqUUID := uuid.NewV5(uuid.Nil, reqID)
	specUUID := uuid.NewV5(uuid.Nil, "spec-id")
	type fields struct {
		ID           uuid.UUID
		ProvidedSpec *ProvidedSpec
	}
	type args struct {
		telemetry *SCNTelemetry
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
					Spec: &spec.Swagger{
						SwaggerProps: spec.SwaggerProps{
							Paths: &spec.Paths{
								Paths: map[string]spec.PathItem{
									"/api": NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
								},
							},
						},
					},
				},
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api", "host", "200", []byte(Data.ReqBody), []byte(Data.RespBody)),
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
					Spec: &spec.Swagger{
						SwaggerProps: spec.SwaggerProps{
							Paths: &spec.Paths{
								Paths: map[string]spec.PathItem{
									"/api": NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
								},
							},
						},
					},
				},
			},
			args: args{
				telemetry: createTelemetryWithSecurity(reqID, http.MethodGet, "/api", "host", "200", []byte(Data.ReqBody), []byte(Data.RespBody)),
			},
			want: &APIDiff{
				Type: DiffTypeChanged,
				Path: "/api",
				// when there is no diff in response, we don’t include 'Produces' in the diff logic so we need to clear produces here
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, clearProduces(NewOperation(t, Data).Op)).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, clearProduces(NewOperation(t, DataWithAuth).Op)).PathItem,
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
					Spec: &spec.Swagger{
						SwaggerProps: spec.SwaggerProps{
							Paths: &spec.Paths{
								Paths: map[string]spec.PathItem{
									"/api": NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
								},
							},
						},
					},
				},
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api/new", "host", "200", []byte(Data.ReqBody), []byte(Data.RespBody)),
			},
			want: &APIDiff{
				Type:             DiffTypeNew,
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
					Spec: &spec.Swagger{
						SwaggerProps: spec.SwaggerProps{
							Paths: &spec.Paths{
								Paths: map[string]spec.PathItem{
									"/api": NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
								},
							},
						},
					},
				},
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodPost, "/api", "host", "200", []byte(req2), []byte(res2)),
			},
			want: &APIDiff{
				Type:             DiffTypeChanged,
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
					Spec: &spec.Swagger{
						SwaggerProps: spec.SwaggerProps{
							Paths: &spec.Paths{
								Paths: map[string]spec.PathItem{
									"/api": NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
								},
							},
						},
					},
				},
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api", "host", "200", []byte(req2), []byte(res2)),
			},
			want: &APIDiff{
				Type:             DiffTypeChanged,
				Path:             "/api",
				OriginalPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				ModifiedPathItem: &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
				InteractionID:    reqUUID,
				SpecID:           specUUID,
			},
			wantErr: false,
		},
		{
			name: "test remove base path",
			fields: fields{
				ID: specUUID,
				ProvidedSpec: &ProvidedSpec{
					Spec: &spec.Swagger{
						SwaggerProps: spec.SwaggerProps{
							BasePath: "/api",
							Paths: &spec.Paths{
								Paths: map[string]spec.PathItem{
									"/foo/bar": NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
								},
							},
						},
					},
				},
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/api/foo/bar", "host", "200", []byte(req2), []byte(res2)),
			},
			want: &APIDiff{
				Type:             DiffTypeChanged,
				Path:             "/api/foo/bar",
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
					Spec: &spec.Swagger{
						SwaggerProps: spec.SwaggerProps{
							BasePath: "/",
							Paths: &spec.Paths{
								Paths: map[string]spec.PathItem{
									"/foo/bar": NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
								},
							},
						},
					},
				},
			},
			args: args{
				telemetry: createTelemetry(reqID, http.MethodGet, "/foo/bar", "host", "200", []byte(req2), []byte(res2)),
			},
			want: &APIDiff{
				Type:             DiffTypeChanged,
				Path:             "/foo/bar",
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
				ID:           tt.fields.ID,
				ProvidedSpec: tt.fields.ProvidedSpec,
			}
			got, err := s.DiffTelemetry(tt.args.telemetry, DiffSourceProvided)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiffTelemetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DiffTelemetry() got = %v, want %v", marshal(got), marshal(tt.want))
			}
		})
	}
}

func clearProduces(op *spec.Operation) *spec.Operation {
	op.Produces = nil
	return op
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
		name    string
		args    args
		want    *spec.Operation
		wantErr bool
	}{
		{
			name: "keep 1 remove 1",
			args: args{
				op: spec.NewOperation("").
					RespondsWith(200, spec.ResponseRef("keep")).
					RespondsWith(300, spec.ResponseRef("delete")),
				statusCodeToKeep: "200",
			},
			want:    spec.NewOperation("").RespondsWith(200, spec.ResponseRef("keep")),
			wantErr: false,
		},
		{
			name: "status code to keep not found - remove all",
			args: args{
				op: spec.NewOperation("").
					RespondsWith(202, spec.ResponseRef("delete")).
					RespondsWith(300, spec.ResponseRef("delete")),
				statusCodeToKeep: "200",
			},
			want:    spec.NewOperation(""),
			wantErr: false,
		},
		{
			name: "status code to keep not found - remove all keep default response",
			args: args{
				op: spec.NewOperation("").
					RespondsWith(202, spec.ResponseRef("delete")).
					RespondsWith(300, spec.ResponseRef("delete")).
					WithDefaultResponse(spec.ResponseRef("keep-default")),
				statusCodeToKeep: "200",
			},
			want: spec.NewOperation("").
				WithDefaultResponse(spec.ResponseRef("keep-default")),
			wantErr: false,
		},
		{
			name: "only status code to keep is found",
			args: args{
				op:               spec.NewOperation("").RespondsWith(200, spec.ResponseRef("keep")),
				statusCodeToKeep: "200",
			},
			want:    spec.NewOperation("").RespondsWith(200, spec.ResponseRef("keep")),
			wantErr: false,
		},
		{
			name: "invalid status code",
			args: args{
				op:               spec.NewOperation("").RespondsWith(200, spec.ResponseRef("")),
				statusCodeToKeep: "invalid",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := keepResponseStatusCode(tt.args.op, tt.args.statusCodeToKeep)
			if (err != nil) != tt.wantErr {
				t.Errorf("keepResponseStatusCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("keepResponseStatusCode() got = %v, want %v", marshal(got), marshal(tt.want))
			}
		})
	}
}

func Test_calculateOperationDiff(t *testing.T) {
	type args struct {
		specOp            *spec.Operation
		telemetryOp       *spec.Operation
		telemetryResponse SCNTResponse
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
				specOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("header")).
					RespondsWith(200, spec.ResponseRef("test")),
				telemetryOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("header")).
					RespondsWith(200, spec.ResponseRef("test")),
				telemetryResponse: SCNTResponse{
					StatusCode: "200",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "no diff - parameters are not sorted",
			args: args{
				specOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("header2")).
					AddParam(spec.HeaderParam("header1")).
					RespondsWith(200, spec.ResponseRef("test")),
				telemetryOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("header1")).
					AddParam(spec.HeaderParam("header2")).
					RespondsWith(200, spec.ResponseRef("test")),
				telemetryResponse: SCNTResponse{
					StatusCode: "200",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "no diff - existing response should be removed",
			args: args{
				specOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("header")).
					RespondsWith(200, spec.ResponseRef("test")).
					RespondsWith(300, spec.ResponseRef("remove")),
				telemetryOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("header")).
					RespondsWith(200, spec.ResponseRef("test")),
				telemetryResponse: SCNTResponse{
					StatusCode: "200",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "no diff - produces should be cleared",
			args: args{
				specOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("header")).
					RespondsWith(200, spec.ResponseRef("test")).
					RespondsWith(403, spec.ResponseRef("keep")).
					WithProduces("application/json"),
				telemetryOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("header")).
					RespondsWith(403, spec.ResponseRef("keep")),
				telemetryResponse: SCNTResponse{
					StatusCode: "403",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "has diff",
			args: args{
				specOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("header")).
					RespondsWith(200, spec.ResponseRef("test")),
				telemetryOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("new-header")).
					RespondsWith(200, spec.ResponseRef("test")),
				telemetryResponse: SCNTResponse{
					StatusCode: "200",
				},
			},
			want: &operationDiff{
				OriginalOperation: spec.NewOperation("").
					AddParam(spec.HeaderParam("header")).
					RespondsWith(200, spec.ResponseRef("test")),
				ModifiedOperation: spec.NewOperation("").
					AddParam(spec.HeaderParam("new-header")).
					RespondsWith(200, spec.ResponseRef("test")),
			},
			wantErr: false,
		},
		{
			name: "has diff in param and not in response - produces should be ignored",
			args: args{
				specOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("header")).
					RespondsWith(200, spec.ResponseRef("200")).
					RespondsWith(403, spec.ResponseRef("403")).
					WithProduces("application/json"),
				telemetryOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("new-header")).
					RespondsWith(200, spec.ResponseRef("200")).
					WithProduces("will-be-ignore"),
				telemetryResponse: SCNTResponse{
					StatusCode: "200",
				},
			},
			want: &operationDiff{
				OriginalOperation: spec.NewOperation("").
					AddParam(spec.HeaderParam("header")).
					RespondsWith(200, spec.ResponseRef("200")),
				ModifiedOperation: spec.NewOperation("").
					AddParam(spec.HeaderParam("new-header")).
					RespondsWith(200, spec.ResponseRef("200")),
			},
			wantErr: false,
		},
		{
			name: "has diff in response - produces should not be ignored",
			args: args{
				specOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("header")).
					RespondsWith(200, spec.ResponseRef("200")).
					RespondsWith(403, spec.ResponseRef("403")).
					WithProduces("application/json"),
				telemetryOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("new-header")).
					RespondsWith(200, spec.ResponseRef("new-200")).
					WithProduces("will-not-be-ignore"),
				telemetryResponse: SCNTResponse{
					StatusCode: "200",
				},
			},
			want: &operationDiff{
				OriginalOperation: spec.NewOperation("").
					AddParam(spec.HeaderParam("header")).
					RespondsWith(200, spec.ResponseRef("200")).
					WithProduces("application/json"),
				ModifiedOperation: spec.NewOperation("").
					AddParam(spec.HeaderParam("new-header")).
					RespondsWith(200, spec.ResponseRef("new-200")).
					WithProduces("will-not-be-ignore"),
			},
			wantErr: false,
		},
		{
			name: "invalid status code",
			args: args{
				specOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("header")).
					RespondsWith(200, spec.ResponseRef("test")),
				telemetryOp: spec.NewOperation("").
					AddParam(spec.HeaderParam("new-header")).
					RespondsWith(200, spec.ResponseRef("test")),
				telemetryResponse: SCNTResponse{
					StatusCode: "invalid",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateOperationDiff(tt.args.specOp, tt.args.telemetryOp, tt.args.telemetryResponse)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateOperationDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateOperationDiff() got = %v, want %v", marshal(got), marshal(tt.want))
			}
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
				obj1: spec.NewOperation("").AddParam(spec.HeaderParam("test")),
				obj2: spec.NewOperation("").AddParam(spec.HeaderParam("test")),
			},
			wantHasDiff: false,
			wantErr:     false,
		},
		{
			name: "has diff (compare only Responses)",
			args: args{
				obj1: spec.NewOperation("").RespondsWith(200, spec.ResponseRef("test")).Responses,
				obj2: spec.NewOperation("").RespondsWith(200, spec.ResponseRef("diff")).Responses,
			},
			wantHasDiff: true,
			wantErr:     false,
		},
		{
			name: "has diff (different objects - Operation vs Responses)",
			args: args{
				obj1: spec.NewOperation("").RespondsWith(200, spec.ResponseRef("diff")),
				obj2: spec.NewOperation("").RespondsWith(200, spec.ResponseRef("diff")).Responses,
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
				operation: spec.NewOperation("").AddParam(spec.PathParam("1")).AddParam(spec.PathParam("2")),
			},
			want: spec.NewOperation("").AddParam(spec.PathParam("1")).AddParam(spec.PathParam("2")),
		},
		{
			name: "sort is needed - sort by 'name'",
			args: args{
				operation: spec.NewOperation("").AddParam(spec.HeaderParam("3")).AddParam(spec.HeaderParam("1")).AddParam(spec.HeaderParam("2")),
			},
			want: spec.NewOperation("").AddParam(spec.HeaderParam("1")).AddParam(spec.HeaderParam("2")).AddParam(spec.HeaderParam("3")),
		},
		{
			name: "param name is the same - sort by 'in'",
			args: args{
				operation: spec.NewOperation("").AddParam(spec.PathParam("1")).AddParam(spec.BodyParam("1", nil)),
			},
			want: spec.NewOperation("").AddParam(spec.BodyParam("1", nil)).AddParam(spec.PathParam("1")),
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
