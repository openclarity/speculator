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
	"fmt"
	"sort"
	"strconv"
	"strings"

	oapi_spec "github.com/go-openapi/spec"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type DiffType string

const (
	DiffTypeNew     DiffType = "NEW"
	DiffTypeChanged DiffType = "CHANGED"
	DiffTypeNoDiff  DiffType = "NO_DIFF"
)

type DiffSource string

const (
	DiffSourceReconstructed DiffSource = "RECONSTRUCTED"
	DiffSourceProvided      DiffSource = "PROVIDED"
)

type APIDiff struct {
	Type             DiffType
	Path             string
	PathID           string
	OriginalPathItem *oapi_spec.PathItem
	ModifiedPathItem *oapi_spec.PathItem
	InteractionID    uuid.UUID
	SpecID           uuid.UUID
}

type operationDiff struct {
	OriginalOperation *oapi_spec.Operation
	ModifiedOperation *oapi_spec.Operation
}

type DiffParams struct {
	operation *oapi_spec.Operation
	method    string
	path      string
	pathID    string
	requestID string
	response  SCNTResponse
}

func createDiffParamsFromTelemetry(telemetry *SCNTelemetry) (*DiffParams, error) {
	securityDefinitions := oapi_spec.SecurityDefinitions{}

	path, _ := GetPathAndQuery(telemetry.SCNTRequest.Path)
	telemetryOp, err := telemetryToOperation(telemetry, securityDefinitions)
	if err != nil {
		return nil, fmt.Errorf("failed to convert telemetry to operation: %w", err)
	}
	return &DiffParams{
		operation: telemetryOp,
		method:    telemetry.SCNTRequest.Method,
		path:      path,
		requestID: telemetry.RequestID,
		response:  telemetry.SCNTResponse,
	}, nil
}

func (s *Spec) DiffTelemetry(telemetry *SCNTelemetry, diffSource DiffSource) (*APIDiff, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	var apiDiff *APIDiff
	var err error
	diffParams, err := createDiffParamsFromTelemetry(telemetry)
	if err != nil {
		return nil, fmt.Errorf("failed to create diff params from telemetry. %w", err)
	}

	switch diffSource {
	case DiffSourceProvided:
		if !s.HasProvidedSpec() {
			log.Infof("No provided spec to diff")
			return nil, nil
		}
		apiDiff, err = s.diffProvidedSpec(diffParams)
		if err != nil {
			return nil, fmt.Errorf("failed to diff provided spec. %w", err)
		}
	case DiffSourceReconstructed:
		if !s.HasApprovedSpec() {
			log.Infof("No approved spec to diff")
			return nil, nil
		}
		apiDiff, err = s.diffApprovedSpec(diffParams)
		if err != nil {
			return nil, fmt.Errorf("failed to diff approved spec. %w", err)
		}
	default:
		return nil, fmt.Errorf("diff source: %v is not valid", diffSource)
	}

	return apiDiff, nil
}

func (s *Spec) diffApprovedSpec(diffParams *DiffParams) (*APIDiff, error) {
	var pathItem *oapi_spec.PathItem
	pathFromTrie, value, found := s.PathTrie.GetPathAndValue(diffParams.path)
	if found {
		diffParams.path = pathFromTrie // The diff will show the parametrized path if matched and not the telemetry path
		pathItem = s.ApprovedSpec.GetPathItem(pathFromTrie)
		if pathID, ok := value.(string); !ok {
			log.Warnf("value is not a string. %v", value)
		} else {
			diffParams.pathID = pathID
		}
	}
	return s.diffPathItem(pathItem, diffParams)
}

func (s *Spec) diffProvidedSpec(diffParams *DiffParams) (*APIDiff, error) {
	var pathNoBase string
	// for path /api/foo/bar and base path of /api, the path that will be saved in paths map will be /foo/bar
	if s.ProvidedSpec.Spec.BasePath != "" {
		pathNoBase = strings.TrimPrefix(diffParams.path, s.ProvidedSpec.Spec.BasePath)
	} else {
		pathNoBase = diffParams.path
	}
	pathItem, ok := s.ProvidedSpec.Spec.Paths.Paths[pathNoBase]
	if !ok {
		return s.diffPathItem(nil, diffParams)
	}

	return s.diffPathItem(&pathItem, diffParams)
}

func (s *Spec) diffPathItem(pathItem *oapi_spec.PathItem, diffParams *DiffParams) (*APIDiff, error) {
	var apiDiff *APIDiff
	method := diffParams.method
	telemetryOp := diffParams.operation
	path := diffParams.path
	requestID := diffParams.requestID
	pathID := diffParams.pathID

	if pathItem == nil {
		apiDiff = s.createAPIDiffEvent(DiffTypeNew, nil, createPathItemFromOperation(method, telemetryOp),
			uuid.FromStringOrNil(requestID), path, pathID)
		return apiDiff, nil
	}

	specOp := GetOperationFromPathItem(pathItem, method)
	if specOp == nil {
		// new operation
		apiDiff := s.createAPIDiffEvent(DiffTypeChanged, pathItem, CopyPathItemWithNewOperation(pathItem, method, telemetryOp),
			uuid.FromStringOrNil(requestID), path, pathID)
		return apiDiff, nil
	}

	diff, err := calculateOperationDiff(specOp, telemetryOp, diffParams.response)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate operation diff: %w", err)
	}
	if diff != nil {
		apiDiff := s.createAPIDiffEvent(DiffTypeChanged, createPathItemFromOperation(method, diff.OriginalOperation),
			createPathItemFromOperation(method, diff.ModifiedOperation), uuid.FromStringOrNil(requestID), path, pathID)
		return apiDiff, nil
	}

	// no diff
	return s.createAPIDiffEvent(DiffTypeNoDiff, nil, nil, uuid.FromStringOrNil(requestID),
		path, pathID), nil
}

func (s *Spec) createAPIDiffEvent(diffType DiffType, original, modified *oapi_spec.PathItem, interactionID uuid.UUID, path, pathID string) *APIDiff {
	return &APIDiff{
		Type:             diffType,
		Path:             path,
		PathID:           pathID,
		OriginalPathItem: original,
		ModifiedPathItem: modified,
		InteractionID:    interactionID,
		SpecID:           s.ID,
	}
}

func createPathItemFromOperation(method string, operation *oapi_spec.Operation) *oapi_spec.PathItem {
	pathItem := oapi_spec.PathItem{}
	AddOperationToPathItem(&pathItem, method, operation)
	return &pathItem
}

func calculateOperationDiff(specOp, telemetryOp *oapi_spec.Operation, telemetryResponse SCNTResponse) (*operationDiff, error) {
	clonedTelemetryOp, err := CloneOperation(telemetryOp)
	if err != nil {
		return nil, fmt.Errorf("failed to clone telemetry operation: %w", err)
	}

	clonedSpecOp, err := CloneOperation(specOp)
	if err != nil {
		return nil, fmt.Errorf("failed to clone spec operation: %w", err)
	}

	clonedTelemetryOp = sortParameters(clonedTelemetryOp)
	clonedSpecOp = sortParameters(clonedSpecOp)

	// Keep only telemetry status code
	clonedSpecOp, err = keepResponseStatusCode(clonedSpecOp, telemetryResponse.StatusCode)
	if err != nil {
		return nil, err
	}

	// Check if there is a change in the response, if so, take “produces“ into account. Otherwise don’t include “produces“ in both.
	hasDiff, err := compareObjects(clonedSpecOp.Responses, clonedTelemetryOp.Responses)
	if err != nil {
		return nil, fmt.Errorf("failed to compare responses: %w", err)
	}
	if hasDiff {
		// Found a diff
		return &operationDiff{
			OriginalOperation: clonedSpecOp,
			ModifiedOperation: clonedTelemetryOp,
		}, nil
	}

	// No diff in response, don’t include 'Produces' and continue to check for diff in the overall object
	clonedTelemetryOp.Produces = nil
	clonedSpecOp.Produces = nil

	hasDiff, err = compareObjects(clonedSpecOp, clonedTelemetryOp)
	if err != nil {
		return nil, fmt.Errorf("failed to compare operations: %w", err)
	}

	if hasDiff {
		return &operationDiff{
			OriginalOperation: clonedSpecOp,
			ModifiedOperation: clonedTelemetryOp,
		}, nil
	}

	// no diff
	return nil, nil
}

func compareObjects(obj1, obj2 interface{}) (hasDiff bool, err error) {
	obj1B, err := json.Marshal(obj1)
	if err != nil {
		return false, fmt.Errorf("failed to marshal obj1: %w", err)
	}

	obj2B, err := json.Marshal(obj2)
	if err != nil {
		return false, fmt.Errorf("failed to marshal obj2: %w", err)
	}

	return !bytes.Equal(obj1B, obj2B), nil
}

// keepResponseStatusCode will remove all status codes from StatusCodeResponses map except the `statusCodeToKeep`.
func keepResponseStatusCode(op *oapi_spec.Operation, statusCodeToKeep string) (*oapi_spec.Operation, error) {
	statusCodeInt, err := strconv.Atoi(statusCodeToKeep)
	if err != nil {
		return nil, fmt.Errorf("invalid status code (%+v): %w", statusCodeToKeep, err)
	}

	// keep only the provided status code
	if op.Responses != nil {
		for code := range op.Responses.StatusCodeResponses {
			if code != statusCodeInt {
				delete(op.Responses.StatusCodeResponses, code)
			}
		}

		// clear operation Responses if needed
		if len(op.Responses.StatusCodeResponses) == 0 {
			if op.Responses.Default == nil {
				// if no `StatusCodeResponses` and no `Default` response `Responses` can be nil
				op.Responses = nil
			} else {
				// else only `StatusCodeResponses` should be nil
				op.Responses.StatusCodeResponses = nil
			}
		}
	}

	return op, nil
}

func sortParameters(operation *oapi_spec.Operation) *oapi_spec.Operation {
	sort.Slice(operation.Parameters, func(i, j int) bool {
		right := operation.Parameters[i]
		left := operation.Parameters[j]
		// Sibling parameters must have unique name + in values
		return right.Name+right.In < left.Name+left.In
	})

	return operation
}
