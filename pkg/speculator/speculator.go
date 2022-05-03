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
	"encoding/gob"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	_spec "github.com/openclarity/speculator/pkg/spec"
)

type SpecKey string

type Config struct {
	OperationGeneratorConfig _spec.OperationGeneratorConfig
}

type Speculator struct {
	Specs map[SpecKey]*_spec.Spec `json:"specs,omitempty"`

	// config is not exported and is not encoded part of the state
	config Config
}

func CreateSpeculator(config Config) *Speculator {
	log.Info("Creating Speculator")
	log.Debugf("Speculator Config %+v", config)
	return &Speculator{
		Specs:  make(map[SpecKey]*_spec.Spec),
		config: config,
	}
}

func GetSpecKey(host, port string) SpecKey {
	return SpecKey(host + ":" + port)
}

func GetHostAndPortFromSpecKey(key SpecKey) (host, port string, err error) {
	const hostAndPortLen = 2
	hostAndPort := strings.Split(string(key), ":")
	if len(hostAndPort) != hostAndPortLen {
		return "", "", fmt.Errorf("invalid key: %v", key)
	}
	host = hostAndPort[0]
	if len(host) == 0 {
		return "", "", fmt.Errorf("no host for key: %v", key)
	}
	port = hostAndPort[1]
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
	const addrLen = 2
	addr := strings.Split(address, ":")
	if len(addr) != addrLen {
		return nil, fmt.Errorf("invalid address: %v", addr)
	}

	return &AddressInfo{
		IP:   addr[0],
		Port: addr[1],
	}, nil
}

func (s *Speculator) InitSpec(host, port string) error {
	specKey := GetSpecKey(host, port)
	if _, ok := s.Specs[specKey]; ok {
		return fmt.Errorf("spec was already initialized using host and port: %s:%s", host, port)
	}
	s.Specs[specKey] = _spec.CreateDefaultSpec(host, port, s.config.OperationGeneratorConfig)
	return nil
}

func (s *Speculator) LearnTelemetry(telemetry *_spec.Telemetry) error {
	destInfo, err := GetAddressInfoFromAddress(telemetry.DestinationAddress)
	if err != nil {
		return fmt.Errorf("failed get destination info: %v", err)
	}
	specKey := GetSpecKey(telemetry.Request.Host, destInfo.Port)
	if _, ok := s.Specs[specKey]; !ok {
		s.Specs[specKey] = _spec.CreateDefaultSpec(telemetry.Request.Host, destInfo.Port, s.config.OperationGeneratorConfig)
	}
	spec := s.Specs[specKey]
	if err := spec.LearnTelemetry(telemetry); err != nil {
		return fmt.Errorf("failed to insert telemetry: %v. %v", telemetry, err)
	}

	return nil
}

func (s *Speculator) DiffTelemetry(telemetry *_spec.Telemetry, diffSource _spec.DiffSource) (*_spec.APIDiff, error) {
	destInfo, err := GetAddressInfoFromAddress(telemetry.DestinationAddress)
	if err != nil {
		return nil, fmt.Errorf("failed get destination info: %v", err)
	}
	specKey := GetSpecKey(telemetry.Request.Host, destInfo.Port)
	spec, ok := s.Specs[specKey]
	if !ok {
		return nil, fmt.Errorf("no spec for key %v", specKey)
	}

	apiDiff, err := spec.DiffTelemetry(telemetry, diffSource)
	if err != nil {
		return nil, fmt.Errorf("failed to run DiffTelemetry: %v", err)
	}

	return apiDiff, nil
}

func (s *Speculator) HasApprovedSpec(key SpecKey) bool {
	spec, ok := s.Specs[key]
	if !ok {
		return false
	}

	return spec.HasApprovedSpec()
}

func (s *Speculator) LoadProvidedSpec(key SpecKey, providedSpec []byte, pathToPathID map[string]string) error {
	spec, ok := s.Specs[key]
	if !ok {
		return fmt.Errorf("no spec found with key: %v", key)
	}

	if err := spec.LoadProvidedSpec(providedSpec, pathToPathID); err != nil {
		return fmt.Errorf("failed to load provided spec: %w", err)
	}

	return nil
}

func (s *Speculator) UnsetProvidedSpec(key SpecKey) error {
	spec, ok := s.Specs[key]
	if !ok {
		return fmt.Errorf("no spec found with key: %v", key)
	}
	spec.UnsetProvidedSpec()
	return nil
}

func (s *Speculator) UnsetApprovedSpec(key SpecKey) error {
	spec, ok := s.Specs[key]
	if !ok {
		return fmt.Errorf("no spec found with key: %v", key)
	}
	spec.UnsetApprovedSpec()
	return nil
}

func (s *Speculator) HasProvidedSpec(key SpecKey) bool {
	spec, ok := s.Specs[key]
	if !ok {
		return false
	}

	return spec.HasProvidedSpec()
}

func (s *Speculator) DumpSpecs() {
	log.Infof("Generating Open API Specs...\n")
	for specKey, spec := range s.Specs {
		approvedYaml, err := spec.GenerateOASYaml()
		if err != nil {
			log.Errorf("failed to generate OAS yaml for %v.: %v", specKey, err)
			continue
		}
		log.Infof("Spec for %s:\n%s\n\n", specKey, approvedYaml)
	}
}

func (s *Speculator) ApplyApprovedReview(specKey SpecKey, approvedReview *_spec.ApprovedSpecReview) error {
	if err := s.Specs[specKey].ApplyApprovedReview(approvedReview); err != nil {
		return fmt.Errorf("failed to apply approved review for spec: %v. %w", specKey, err)
	}
	return nil
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

func DecodeState(filePath string, config Config) (*Speculator, error) {
	r := &Speculator{}
	file, err := openFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file (%v): %v", filePath, err)
	}
	defer closeFile(file)

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode state: %v", err)
	}

	r.config = config

	log.Info("Speculator state was decoded")
	log.Debugf("Speculator Config %+v", config)

	return r, nil
}

func openFile(filePath string) (*os.File, error) {
	const perm = 400
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.FileMode(perm))
	if err != nil {
		return nil, fmt.Errorf("failed to open file (%v) for writing: %v", filePath, err)
	}

	return file, nil
}

func closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		log.Errorf("Failed to close file: %v", err)
	}
}
