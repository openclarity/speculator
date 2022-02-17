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
	"fmt"
	"strconv"
	"strings"

	oapi_spec "github.com/go-openapi/spec"
)

// Note: securityDefinitions might be updated.
func (s *Spec) telemetryToOperation(telemetry *Telemetry, securityDefinitions oapi_spec.SecurityDefinitions) (*oapi_spec.Operation, error) {
	statusCode, err := strconv.Atoi(telemetry.Response.StatusCode)
	if err != nil {
		return nil, fmt.Errorf("failed to convert status code: %v. %v", statusCode, err)
	}

	queryParams, err := extractQueryParams(telemetry.Request.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to convert query params: %v", err)
	}

	if s.OpGenerator == nil {
		return nil, fmt.Errorf("operation generator was not set")
	}

	// Generate operation from telemetry
	telemetryOp, err := s.OpGenerator.GenerateSpecOperation(&HTTPInteractionData{
		ReqBody:     string(telemetry.Request.Common.Body),
		RespBody:    string(telemetry.Response.Common.Body),
		ReqHeaders:  ConvertHeadersToMap(telemetry.Request.Common.Headers),
		RespHeaders: ConvertHeadersToMap(telemetry.Response.Common.Headers),
		QueryParams: queryParams,
		statusCode:  statusCode,
	}, securityDefinitions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate spec operation. %v", err)
	}
	return telemetryOp, nil
}

// example: for "/example-path?param=value" returns "/example-path", "param=value"
func GetPathAndQuery(fullPath string) (path, query string) {
	index := strings.IndexByte(fullPath, '?')
	if index == -1 {
		return fullPath, ""
	}

	// /path?
	if index == (len(fullPath) - 1) {
		return fullPath, ""
	}

	path = fullPath[:index]
	query = fullPath[index+1:]
	return
}

func GetContentTypeWithoutParameter(contentTypeHeaderField string) string {
	// https://greenbytes.de/tech/webdav/rfc2616.html#media.types
	// remove parameters if exists
	contentTypeWithoutParams := strings.Split(contentTypeHeaderField, ";")[0]
	return strings.TrimSpace(contentTypeWithoutParams)
}
