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
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	oapi_spec "github.com/go-openapi/spec"
	uuid "github.com/satori/go.uuid"

	"github.com/apiclarity/speculator/pkg/pathtrie"
)

func TestSpec_LearnTelemetry(t *testing.T) {
	type fields struct{}
	type args struct {
		telemetries []*Telemetry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "one",
			fields: fields{},
			args: args{
				telemetries: []*Telemetry{
					{
						RequestID: "req-id",
						Scheme:    "http",
						Request: &Request{
							Method: "GET",
							Path:   "/some/path",
							Host:   "www.example.com",
							Common: &Common{
								Version: "1",
								Headers: []*Header{
									{
										Key:   contentTypeHeaderName,
										Value: mediaTypeApplicationJSON,
									},
								},
								Body:          []byte(req1),
								TruncatedBody: false,
							},
						},
						Response: &Response{
							StatusCode: "200",
							Common: &Common{
								Version: "1",
								Headers: []*Header{
									{
										Key:   contentTypeHeaderName,
										Value: mediaTypeApplicationJSON,
									},
								},
								Body:          []byte(res1),
								TruncatedBody: false,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "two",
			fields: fields{},
			args: args{
				telemetries: []*Telemetry{
					{
						RequestID: "req-id",
						Scheme:    "http",
						Request: &Request{
							Method: "GET",
							Path:   "/some/path",
							Host:   "www.example.com",
							Common: &Common{
								Version: "1",
								Body:    []byte(req1),
								Headers: []*Header{
									{
										Key:   contentTypeHeaderName,
										Value: mediaTypeApplicationJSON,
									},
									{
										Key:   "X-Test-Req-1",
										Value: "req1",
									},
								},
								TruncatedBody: false,
							},
						},
						Response: &Response{
							StatusCode: "200",
							Common: &Common{
								Version: "1",
								Body:    []byte(res1),
								Headers: []*Header{
									{
										Key:   contentTypeHeaderName,
										Value: mediaTypeApplicationJSON,
									},
									{
										Key:   "X-Test-Res-1",
										Value: "res1",
									},
								},
								TruncatedBody: false,
							},
						},
					},
					{
						RequestID: "req-id",
						Scheme:    "http",
						Request: &Request{
							Method: "GET",
							Path:   "/some/path",
							Host:   "www.example.com",
							Common: &Common{
								Version: "1",
								Body:    []byte(req2),
								Headers: []*Header{
									{
										Key:   contentTypeHeaderName,
										Value: mediaTypeApplicationJSON,
									},
									{
										Key:   "X-Test-Req-2",
										Value: "req2",
									},
								},
								TruncatedBody: false,
							},
						},
						Response: &Response{
							StatusCode: "200",
							Common: &Common{
								Version: "1",
								Body:    []byte(res2),
								Headers: []*Header{
									{
										Key:   contentTypeHeaderName,
										Value: mediaTypeApplicationJSON,
									},
									{
										Key:   "X-Test-Res-2",
										Value: "res2",
									},
								},
								TruncatedBody: false,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := CreateDefaultSpec("host", "80", testOperationGeneratorConfig)
			for _, telemetry := range tt.args.telemetries {
				// file, _ := json.MarshalIndent(telemetry, "", " ")

				//_ = ioutil.WriteFile(fmt.Sprintf("test%v.json", i), file, 0644)
				if err := s.LearnTelemetry(telemetry); (err != nil) != tt.wantErr {
					t.Errorf("LearnTelemetry() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestSpec_SpecInfoClone(t *testing.T) {
	uuidVar := uuid.NewV4()
	pathTrie := pathtrie.New()
	pathTrie.Insert("/api", 1)

	type fields struct {
		Host             string
		Port             string
		ID               uuid.UUID
		ProvidedSpec     *ProvidedSpec
		ApprovedSpec     *ApprovedSpec
		LearningSpec     *LearningSpec
		ApprovedPathTrie pathtrie.PathTrie
		ProvidedPathTrie pathtrie.PathTrie
	}
	tests := []struct {
		name    string
		fields  fields
		want    *Spec
		wantErr bool
	}{
		{
			name: "clone spec",
			fields: fields{
				Host: "host",
				Port: "80",
				ID:   uuidVar,
				ProvidedSpec: &ProvidedSpec{
					Spec: &oapi_spec.Swagger{
						SwaggerProps: oapi_spec.SwaggerProps{
							Paths: &oapi_spec.Paths{
								Paths: map[string]oapi_spec.PathItem{
									"/api": NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
								},
							},
						},
					},
				},
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*oapi_spec.PathItem{},
				},
				LearningSpec: &LearningSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"/api/1": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/api/2": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
					},
				},
				ApprovedPathTrie: pathTrie,
				ProvidedPathTrie: pathTrie,
			},
			want: &Spec{
				SpecInfo: SpecInfo{
					Host: "host",
					Port: "80",
					ID:   uuidVar,
					ProvidedSpec: &ProvidedSpec{
						Spec: &oapi_spec.Swagger{
							SwaggerProps: oapi_spec.SwaggerProps{
								Paths: &oapi_spec.Paths{
									Paths: map[string]oapi_spec.PathItem{
										"/api": NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
									},
								},
							},
						},
					},
					ApprovedSpec: &ApprovedSpec{
						PathItems: map[string]*oapi_spec.PathItem{},
					},
					LearningSpec: &LearningSpec{
						PathItems: map[string]*oapi_spec.PathItem{
							"/api/1": &NewTestPathItem().
								WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
							"/api/2": &NewTestPathItem().
								WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
						},
					},
					ApprovedPathTrie: pathTrie,
					ProvidedPathTrie: pathTrie,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Spec{
				SpecInfo: SpecInfo{
					Host:             tt.fields.Host,
					Port:             tt.fields.Port,
					ID:               tt.fields.ID,
					ProvidedSpec:     tt.fields.ProvidedSpec,
					ApprovedSpec:     tt.fields.ApprovedSpec,
					LearningSpec:     tt.fields.LearningSpec,
					ApprovedPathTrie: tt.fields.ApprovedPathTrie,
					ProvidedPathTrie: tt.fields.ProvidedPathTrie,
				},
			}
			got, err := s.SpecInfoClone()
			if (err != nil) != tt.wantErr {
				t.Errorf("SpecInfoClone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotB, _ := json.Marshal(got)
			wantB, _ := json.Marshal(tt.want)

			if !bytes.Equal(gotB, wantB) {
				t.Errorf("SpecInfoClone() got = %s, want %s", gotB, wantB)
			}
		})
	}
}
