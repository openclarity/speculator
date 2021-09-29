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

package speculator

import (
	"os"
	"testing"

	uuid "github.com/satori/go.uuid"

	"github.com/apiclarity/speculator/pkg/spec"
)

func TestGetHostAndPortFromSpecKey(t *testing.T) {
	type args struct {
		key SpecKey
	}
	tests := []struct {
		name     string
		args     args
		wantHost string
		wantPort string
		wantErr  bool
	}{
		{
			name: "invalid key",
			args: args{
				key: "invalid",
			},
			wantHost: "",
			wantPort: "",
			wantErr:  true,
		},
		{
			name: "invalid:key:invalid",
			args: args{
				key: "invalid",
			},
			wantHost: "",
			wantPort: "",
			wantErr:  true,
		},
		{
			name: "invalid key - no host",
			args: args{
				key: ":8080",
			},
			wantHost: "",
			wantPort: "",
			wantErr:  true,
		},
		{
			name: "invalid key - no port",
			args: args{
				key: "host:",
			},
			wantHost: "",
			wantPort: "",
			wantErr:  true,
		},
		{
			name: "valid key",
			args: args{
				key: "host:8080",
			},
			wantHost: "host",
			wantPort: "8080",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPort, err := GetHostAndPortFromSpecKey(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHostAndPortFromSpecKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHost != tt.wantHost {
				t.Errorf("GetHostAndPortFromSpecKey() gotHost = %v, want %v", gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("GetHostAndPortFromSpecKey() gotPort = %v, want %v", gotPort, tt.wantPort)
			}
		})
	}
}

func TestDecodeState(t *testing.T) {
	testSpec := GetSpecKey("host", "port")
	testStatePath := "/tmp/" + uuid.NewV4().String() + "state.gob"
	defer func() {
		_ = os.Remove(testStatePath)
	}()

	speculatorConfig := Config{
		OperationGeneratorConfig: spec.OperationGeneratorConfig{
			ResponseHeadersToIgnore: []string{"before"},
		},
	}
	speculator := CreateSpeculator(speculatorConfig)
	speculator.Specs[testSpec] = spec.CreateDefaultSpec("host", "port", speculator.config.OperationGeneratorConfig)

	if err := speculator.EncodeState(testStatePath); err != nil {
		t.Errorf("EncodeState() error = %v", err)
		return
	}

	newSpeculatorConfig := Config{
		OperationGeneratorConfig: spec.OperationGeneratorConfig{
			ResponseHeadersToIgnore: []string{"after"},
		},
	}
	got, err := DecodeState(testStatePath, newSpeculatorConfig)
	if err != nil {
		t.Errorf("DecodeState() error = %v", err)
		return
	}

	// OpGenerator on the decoded state should hold the previous OperationGeneratorConfig
	responseHeadersToIgnore := got.Specs[testSpec].OpGenerator.ResponseHeadersToIgnore
	if _, ok := responseHeadersToIgnore["before"]; !ok {
		t.Errorf("ResponseHeadersToIgnore not as expected = %+v", responseHeadersToIgnore)
		return
	}
}
