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
	"testing"
)

func TestSpec_LearnTelemetry(t *testing.T) {
	type fields struct{}
	type args struct {
		telemetries []*SCNTelemetry
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
				telemetries: []*SCNTelemetry{
					{
						RequestID: "req-id",
						Scheme:    "http",
						SCNTRequest: SCNTRequest{
							Method: "GET",
							Path:   "/some/path",
							Host:   "www.example.com",
							SCNTCommon: SCNTCommon{
								Version:       "1",
								Headers:       [][2]string{{contentTypeHeaderName, mediaTypeApplicationJSON}},
								Body:          []byte(req1),
								TruncatedBody: false,
							},
						},
						SCNTResponse: SCNTResponse{
							StatusCode: "200",
							SCNTCommon: SCNTCommon{
								Version:       "1",
								Headers:       [][2]string{{contentTypeHeaderName, mediaTypeApplicationJSON}},
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
				telemetries: []*SCNTelemetry{
					{
						RequestID: "req-id",
						Scheme:    "http",
						SCNTRequest: SCNTRequest{
							Method: "GET",
							Path:   "/some/path",
							Host:   "www.example.com",
							SCNTCommon: SCNTCommon{
								Version:       "1",
								Body:          []byte(req1),
								Headers:       [][2]string{{"X-Test-Req-1", "req1"}, {contentTypeHeaderName, mediaTypeApplicationJSON}},
								TruncatedBody: false,
							},
						},
						SCNTResponse: SCNTResponse{
							StatusCode: "200",
							SCNTCommon: SCNTCommon{
								Version:       "1",
								Body:          []byte(res1),
								Headers:       [][2]string{{"X-Test-Res-1", "res1"}, {contentTypeHeaderName, mediaTypeApplicationJSON}},
								TruncatedBody: false,
							},
						},
					},
					{
						RequestID: "req-id",
						Scheme:    "http",
						SCNTRequest: SCNTRequest{
							Method: "GET",
							Path:   "/some/path",
							Host:   "www.example.com",
							SCNTCommon: SCNTCommon{
								Version:       "1",
								Body:          []byte(req2),
								Headers:       [][2]string{{"X-Test-Req-2", "req2"}, {contentTypeHeaderName, mediaTypeApplicationJSON}},
								TruncatedBody: false,
							},
						},
						SCNTResponse: SCNTResponse{
							StatusCode: "200",
							SCNTCommon: SCNTCommon{
								Version:       "1",
								Body:          []byte(res2),
								Headers:       [][2]string{{"X-Test-Res-2", "res2"}, {contentTypeHeaderName, mediaTypeApplicationJSON}},
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
			s := CreateDefaultSpec("host", "80")
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
