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
	"reflect"
	"testing"

	oapi_spec "github.com/getkin/kin-openapi/openapi3"
	uuid "github.com/satori/go.uuid"

	"github.com/openclarity/speculator/pkg/pathtrie"
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
					Doc: &oapi_spec.T{
						Info: createDefaultSwaggerInfo(),
						Paths: map[string]*oapi_spec.PathItem{
							"/api": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
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
						Doc: &oapi_spec.T{
							Info: createDefaultSwaggerInfo(),
							Paths: map[string]*oapi_spec.PathItem{
								"/api": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
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

var swaggerWithRef = `components:
  schemas:
    count_description_id_imageUrl_name_price_tag:
      properties:
        count:
          format: int64
          type: integer
        description:
          type: string
        id:
          format: uuid
          type: string
        imageUrl:
          items:
            format: json-pointer
            type: string
          type: array
        name:
          type: string
        price:
          type: number
        tag:
          items:
            type: string
          type: array
      type: object
    count_description_id_imageUrl_name_price_tag_0:
      properties:
        count:
          format: int64
          type: integer
        description:
          type: string
        id:
          type: string
        imageUrl:
          items:
            format: json-pointer
            type: string
          type: array
        name:
          type: string
        price:
          format: int64
          type: integer
        tag:
          items:
            type: string
          type: array
      type: object
    err_size:
      properties:
        err:
          type: string
        size:
          format: int64
          type: integer
      type: object
    err_tags:
      properties:
        err:
          type: string
        tags:
          items:
            type: string
          type: array
      type: object
info:
  contact:
    email: apiteam@swagger.io
  description: This is a generated Open API Spec
  license:
    name: Apache 2.0
    url: https://www.apache.org/licenses/LICENSE-2.0.html
  termsOfService: https://swagger.io/terms/
  title: Swagger
  version: 1.0.0
openapi: 3.0.3
paths:
  /catalogue:
    get:
      deprecated: true
      parameters:
        - in: query
          name: tags
          schema:
            type: string
        - in: query
          name: page
          schema:
            format: int64
            type: integer
        - in: query
          name: size
          schema:
            format: int64
            type: integer
      responses:
        '200':
          content:
            application/json:
              schema:
                items:
                  $ref: '#/components/schemas/count_description_id_imageUrl_name_price_tag'
                type: array
          description: response
        default:
          description: default
  /catalogue/size:
    get:
      parameters:
        - in: query
          name: tags
          schema:
            type: string
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/err_size'
          description: response
        default:
          description: default
  /catalogue/{param1}:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/count_description_id_imageUrl_name_price_tag_0'
          description: response
        default:
          description: default
    parameters:
      - in: path
        name: param1
        required: true
        schema:
          type: string
  /tags:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/err_tags'
          description: response
        default:
          description: default
servers:
  - url: http://catalogue.sock-shop:80
`

func TestLoadAndValidateRawJSONSpecV3(t *testing.T) {
	type args struct {
		spec []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *oapi_spec.T
		wantErr bool
	}{
		{
			name: "",
			args: args{
				spec: []byte(swaggerWithRef),
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadAndValidateRawJSONSpecV3(tt.args.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadAndValidateRawJSONSpecV3() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadAndValidateRawJSONSpecV3() got = %v, want %v", got, tt.want)
			}
		})
	}
}
