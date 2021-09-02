/*
 *
 * Copyright (c) 2020 Cisco Systems, Inc. and its affiliates.
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package speculator

import (
	"encoding/gob"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	_spec "github.com/apiclarity/speculator/pkg/spec"
)

type SpecKey string

type Speculator struct {
	Specs    map[SpecKey]*_spec.Spec `json:"specs,omitempty"`
	ApiDiffs []*_spec.ApiDiff
}

func CreateSpeculator() *Speculator {
	return &Speculator{
		Specs: make(map[SpecKey]*_spec.Spec),
	}
}

func GetSpecKey(host, port string) SpecKey {
	return SpecKey(host + ":" + port)
}

func GetHostAndPortFromSpecKey(key SpecKey) (host, port string, err error) {
	info := strings.Split(string(key), ":")
	if len(info) != 2 {
		return "", "", fmt.Errorf("invalid key: %v", key)
	}
	host = info[0]
	if len(host) == 0 {
		return "", "", fmt.Errorf("no host for key: %v", key)
	}
	port = info[1]
	if len(port) == 0 {
		return "", "", fmt.Errorf("no port for key: %v", key)
	}
	return host, port, nil
}

func (s *Speculator) SuggestedReview(specKey SpecKey) (*_spec.SuggestedSpecReview, error) {
	spec, ok := s.Specs[specKey]
	if !ok {
		return nil, fmt.Errorf("spec doesn't exist for key %v", specKey)
	}

	return spec.CreateSuggestedReview(), nil
}

type AddressInfo struct {
	IP   string
	Port string
}

func GetAddressInfoFromAddress(address string) (*AddressInfo, error) {
	addr := strings.Split(address, ":")
	if len(addr) != 2 {
		return nil, fmt.Errorf("invalid address: %v", addr)
	}

	return &AddressInfo{
		IP:   addr[0],
		Port: addr[1],
	}, nil
}

func (s *Speculator) LearnTelemetry(telemetry *_spec.SCNTelemetry) error {
	destInfo, err := GetAddressInfoFromAddress(telemetry.DestinationAddress)
	if err != nil {
		return fmt.Errorf("failed get destination info: %v", err)
	}
	specKey := GetSpecKey(telemetry.SCNTRequest.Host, destInfo.Port)
	if _, ok := s.Specs[specKey]; !ok {
		s.Specs[specKey] = _spec.CreateDefaultSpec(telemetry.SCNTRequest.Host, destInfo.Port)
	}
	spec := s.Specs[specKey]
	if err := spec.LearnTelemetry(telemetry); err != nil {
		return fmt.Errorf("failed to insert telemetry: %v. %v", telemetry, err)
	}

	return nil
}

func (s *Speculator) DiffTelemetry(telemetry *_spec.SCNTelemetry, diffSource _spec.DiffSource) (*_spec.ApiDiff, error) {
	destInfo, err := GetAddressInfoFromAddress(telemetry.DestinationAddress)
	if err != nil {
		return nil, fmt.Errorf("failed get destination info: %v", err)
	}
	specKey := GetSpecKey(telemetry.SCNTRequest.Host, destInfo.Port)
	spec, ok := s.Specs[specKey]
	if !ok {
		return nil, fmt.Errorf("no spec for key %v", specKey)
	}

	return spec.DiffTelemetry(telemetry, diffSource)
}

func (s *Speculator) HasApprovedSpec(key SpecKey) bool {
	spec, ok := s.Specs[key]
	if !ok {
		return false
	}

	return spec.HasApprovedSpec()
}

func (s *Speculator) LoadProvidedSpec(key SpecKey, providedSpec []byte) error {
	spec, ok := s.Specs[key]
	if !ok {
		return fmt.Errorf("no spec found with key: %v", key)
	}

	return spec.LoadProvidedSpec(providedSpec)
}

func (s *Speculator) HasProvidedSpec(key SpecKey) bool {
	spec, ok := s.Specs[key]
	if !ok {
		return false
	}

	return spec.HasProvidedSpec()
}

func (s *Speculator) DumpSpecs() {
	fmt.Printf("Generating Open API Specs...\n")
	for specKey, spec := range s.Specs {
		approvedYaml, err := spec.GenerateOASYaml()
		if err != nil {
			log.Errorf("failed to generate OAS yaml for %v.: %v", specKey, err)
			continue
		}
		fmt.Printf("Spec for %s:\n%s\n\n", specKey, approvedYaml)
	}
}

func (s *Speculator) ApplyApprovedReview(specKey SpecKey, approvedReview *_spec.ApprovedSpecReview) {
	s.Specs[specKey].ApplyApprovedReview(approvedReview)
}

func (s *Speculator) EncodeState(filePath string) error {
	file, err := openFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open state file: %v", err)
	}
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(s)
	if err != nil {
		return fmt.Errorf("failed to encode state: %v", err)
	}
	closeFile(file)

	return nil
}

func DecodeState(filePath string) (*Speculator, error) {
	r := &Speculator{}
	file, err := openFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file (%v): %v", filePath, err)
	}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode state: %v", err)
	}
	closeFile(file)

	return r, nil
}

func openFile(filePath string) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 400)
	if err != nil {
		return nil, fmt.Errorf("failed to open file (%v) for writing: %v", filePath, err)
	}

	return file, nil
}

func closeFile(f *os.File) {
	err := f.Close()
	if err != nil {
		log.Errorf("Failed to close file: %v", err)
	}
}
