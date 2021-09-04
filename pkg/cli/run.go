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
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/apiclarity/speculator/pkg/spec"
	"github.com/apiclarity/speculator/pkg/speculator"
)

func Run(c *cli.Context) {
	statePath := c.String("state")
	var s *speculator.Speculator
	if statePath != "" {
		var err error
		s, err = speculator.DecodeState(statePath)
		if err != nil {
			log.Fatalf("Failed to decode stored state in path %v", statePath)
		}
	} else {
		s = speculator.CreateSpeculator()
	}
	fileNames := c.StringSlice("t")

	fmt.Printf("Reading interactions from files\n")

	for _, fileName := range fileNames {
		log.Infof("Reading telemetry from %s", fileName)
		telemetryB, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Errorf("Failed to read from file: %v. %v", fileName, err)
			continue
		}
		telemetry := &spec.SCNTelemetry{}
		err = json.Unmarshal(telemetryB, telemetry)
		if err != nil {
			log.Errorf("Failed to unmarshal telemetry. %v", err)
			continue
		}
		log.Infof("Learning HTTP interaction for %v %v%v", telemetry.SCNTRequest.Method, telemetry.SCNTRequest.Host, telemetry.SCNTRequest.Path)
		err = s.LearnTelemetry(telemetry)
		if err != nil {
			log.Errorf("Failed to learn telemetry. %v", err)
			continue
		}
		log.Infof("Learned HTTP interaction for %v %v%v", telemetry.SCNTRequest.Method, telemetry.SCNTRequest.Host, telemetry.SCNTRequest.Path)
	}
	log.Infof("Generating specs")
	s.DumpSpecs()
	if c.String("save") != "" {
		if err := s.EncodeState(c.String("save")); err != nil {
			log.Fatalf("Failed to encode speculator: %v", err)
		}
	}
}
