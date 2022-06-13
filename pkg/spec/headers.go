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
	"strings"

	spec "github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
)

var defaultIgnoredHeaders = []string{
	contentTypeHeaderName,
	acceptTypeHeaderName,
	authorizationTypeHeaderName,
}

func createHeadersToIgnore(headers []string) map[string]struct{} {
	ret := make(map[string]struct{})

	for _, header := range append(defaultIgnoredHeaders, headers...) {
		ret[strings.ToLower(header)] = struct{}{}
	}

	return ret
}

func shouldIgnoreHeader(headerToIgnore map[string]struct{}, headerKey string) bool {
	_, ok := headerToIgnore[strings.ToLower(headerKey)]
	return ok
}

func (o *OperationGenerator) addResponseHeader(response *spec.Response, headerKey, headerValue string) *spec.Response {
	if shouldIgnoreHeader(o.ResponseHeadersToIgnore, headerKey) {
		return response
	}

	if response.Headers == nil {
		response.Headers = make(spec.Headers)
	}

	response.Headers[headerKey] = &spec.HeaderRef{
		Value: &spec.Header{
			Parameter: spec.Parameter{
				Schema: spec.NewSchemaRef("",
					getSchemaFromValue(headerValue, true, spec.ParameterInHeader)),
			},
		},
	}

	return response
}

// https://swagger.io/docs/specification/describing-parameters/#header-parameters
func (o *OperationGenerator) addHeaderParam(operation *spec.Operation, headerKey, headerValue string) *spec.Operation {
	if shouldIgnoreHeader(o.RequestHeadersToIgnore, headerKey) {
		return operation
	}

	headerParam := spec.NewHeaderParameter(headerKey).
		WithSchema(getSchemaFromValue(headerValue, true, spec.ParameterInHeader))
	operation.AddParameter(headerParam)

	return operation
}

// https://swagger.io/docs/specification/describing-parameters/#cookie-parameters
func (o *OperationGenerator) addCookieParam(operation *spec.Operation, headerValue string) *spec.Operation {
	// Multiple cookie parameters are sent in the same header, separated by a semicolon and space.
	for _, cookie := range strings.Split(headerValue, "; ") {
		cookieKeyAndValue := strings.Split(cookie, "=")
		if len(cookieKeyAndValue) != 2 { // nolint:gomnd
			log.Warnf("unsupported cookie param. %v", cookie)
			continue
		}
		key, value := cookieKeyAndValue[0], cookieKeyAndValue[1]
		// Cookie parameters can be primitive values, arrays and objects.
		// Arrays and objects are serialized using the form style.
		headerParam := spec.NewCookieParameter(key).WithSchema(getSchemaFromValue(value, true, spec.ParameterInCookie))
		operation.AddParameter(headerParam)
	}

	return operation
}

func ConvertHeadersToMap(headers []*Header) map[string]string {
	headersMap := make(map[string]string)

	for _, header := range headers {
		headersMap[strings.ToLower(header.Key)] = header.Value
	}

	return headersMap
}
