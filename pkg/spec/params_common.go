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
	"github.com/go-openapi/spec"
	log "github.com/sirupsen/logrus"
)

func populateParam(parameter *spec.Parameter, values []string, allowCollection bool) *spec.Parameter {
	if len(values) == 0 || values[0] == "" {
		// Query string and form data parameters may only have a name and no values
		if parameter.In != parametersInQuery && parameter.In != parametersInForm {
			log.Warnf("Query string and form data parameters may only have a name and no values. type=%v", parameter.In)
			return parameter
		}
		parameter.Typed(schemaTypeBoolean, "").AllowsEmptyValues().AsRequired()
	} else if len(values) == 1 {
		if isDateFormat(values[0]) {
			parameter.Typed(schemaTypeString, "")
		} else {
			items, collectionFormat := getCollection(values[0], supportedCollectionFormat)
			if allowCollection && items != nil {
				parameter.CollectionOf(items, collectionFormat)
			} else {
				tpe, format := getTypeAndFormat(values[0])
				parameter.Typed(tpe, format)
			}
		}
	} else {
		// Multiple parameter instances rather than multiple values.
		// This is only supported for the `in: query` and `in: formData` parameters.
		// ex. foo=value&foo=another_value
		if parameter.In != parametersInQuery && parameter.In != parametersInForm {
			log.Warnf("Multiple parameter instances supported only for query and formData parameters. type=%v", parameter.In)
			return parameter
		}
		tpe, format := getTypeAndFormat(values[0])
		parameter.CollectionOf(spec.NewItems().Typed(tpe, format), collectionFormatMulti)
	}

	return parameter
}
