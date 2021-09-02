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

package spec

import (
	"strings"

	"github.com/go-openapi/spec"
)

var ignoredHeaders = map[string]bool{
	contentTypeHeaderName:       true,
	acceptTypeHeaderName:        true,
	authorizationTypeHeaderName: true,
}

// TODO: Do we want to support adding ignored headers via an env var?
func shouldIgnoreHeader(headerKey string) bool {
	return ignoredHeaders[strings.ToLower(headerKey)]
}

func addResponseHeader(response *spec.Response, headerKey, headerValue string) *spec.Response {
	if shouldIgnoreHeader(headerKey) {
		return response
	}

	responseHeader := spec.ResponseHeader()

	if isDateFormat(headerValue) {
		responseHeader.Typed(schemaTypeString, "")
	} else {
		items, collectionFormat := getCollection(headerValue, supportedCollectionFormat)
		if items != nil {
			responseHeader.CollectionOf(items, collectionFormat)
		} else {
			tpe, format := getTypeAndFormat(headerValue)
			responseHeader.Typed(tpe, format)
		}
	}

	return response.AddHeader(headerKey, responseHeader)
}

func addHeaderParam(operation *spec.Operation, headerKey, headerValue string) *spec.Operation {
	if shouldIgnoreHeader(headerKey) {
		return operation
	}

	headerParam := spec.HeaderParam(headerKey)

	return operation.AddParam(populateParam(headerParam, []string{headerValue}, true))
}

func ConvertHeadersToMap(headers [][2]string) map[string]string {
	headersMap := make(map[string]string)

	for _, header := range headers {
		headersMap[strings.ToLower(header[0])] = header[1]
	}

	return headersMap
}
