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

package cli

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli"

	"github.com/apiclarity/speculator/pkg/spec"
	"github.com/apiclarity/speculator/pkg/speculator"
)

func Run(c *cli.Context) {
	statePath := c.String("state")
	var s *speculator.Speculator

	speculatorConfig := createSpeculatorConfig()
	if statePath != "" {
		var err error
		s, err = speculator.DecodeState(statePath, speculatorConfig)
		if err != nil {
			log.Fatalf("Failed to decode stored state in path %v", statePath)
		}
	} else {
		s = speculator.CreateSpeculator(speculatorConfig)
	}
	fileNames := c.StringSlice("t")

	log.Infof("Reading interactions from files...")

	for _, fileName := range fileNames {
		log.Infof("Reading telemetry from %s", fileName)
		telemetryB, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Errorf("Failed to read from file: %v. %v", fileName, err)
			continue
		}
		telemetry := &spec.Telemetry{}
		err = json.Unmarshal(telemetryB, telemetry)
		if err != nil {
			log.Errorf("Failed to unmarshal telemetry. %v", err)
			continue
		}
		log.Infof("Learning HTTP interaction for %v %v%v", telemetry.Request.Method, telemetry.Request.Host, telemetry.Request.Path)
		err = s.LearnTelemetry(telemetry)
		if err != nil {
			log.Errorf("Failed to learn telemetry. %v", err)
			continue
		}
		log.Infof("Learned HTTP interaction for %v %v%v", telemetry.Request.Method, telemetry.Request.Host, telemetry.Request.Path)
	}
	log.Infof("Generating specs")
	s.DumpSpecs()
	if c.String("save") != "" {
		if err := s.EncodeState(c.String("save")); err != nil {
			log.Fatalf("Failed to encode speculator: %v", err)
		}
	}
}

func createSpeculatorConfig() speculator.Config {
	return speculator.Config{
		OperationGeneratorConfig: spec.OperationGeneratorConfig{
			ResponseHeadersToIgnore: viper.GetStringSlice("RESPONSE_HEADERS_TO_IGNORE"),
			RequestHeadersToIgnore:  viper.GetStringSlice("REQUEST_HEADERS_TO_IGNORE"),
		},
	}
}
