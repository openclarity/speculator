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
	"mime/multipart"
	"net/url"
	"strings"

	"github.com/go-openapi/spec"
	log "github.com/sirupsen/logrus"
)

const (
	// taken from net/http/request.go.
	defaultMaxMemory = 32 << 20 // 32 MB
)

func addApplicationFormParams(operation *spec.Operation, sd spec.SecurityDefinitions, body string) (*spec.Operation, spec.SecurityDefinitions) {
	values, err := url.ParseQuery(body)
	if err != nil {
		log.Warnf("failed to parse query. body=%v: %v", body, err)
		return operation, sd
	}

	for key, values := range values {
		if key == AccessTokenParamKey {
			operation = addSecurity(operation, OAuth2SecurityDefinitionKey)
			sd = updateSecurityDefinitions(sd, OAuth2SecurityDefinitionKey)
		} else {
			operation.AddParam(populateParam(spec.FormDataParam(key), values, true))
		}
	}

	return operation, sd
}

func addMultipartFormDataParams(operation *spec.Operation, body string, mediaTypeParams map[string]string) (*spec.Operation, error) {
	boundary, ok := mediaTypeParams["boundary"]
	if !ok {
		return operation, fmt.Errorf("no multipart boundary param in Content-Type")
	}

	form, err := multipart.NewReader(strings.NewReader(body), boundary).ReadForm(defaultMaxMemory)
	if err != nil {
		return operation, fmt.Errorf("failed to read form: %w", err)
	}

	// add file formData
	for key := range form.File {
		operation.AddParam(spec.FileParam(key))
	}

	// add values formData
	for key, values := range form.Value {
		// when using populateParam strings with comma it will be translated to array and it might be wrong
		// also when the array collection format is ssv the spaces will be URL encoded
		// for now we will ignore collection
		operation.AddParam(populateParam(spec.FormDataParam(key), values, false))
	}

	return operation, nil
}
