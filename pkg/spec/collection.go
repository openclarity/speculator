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
	"strconv"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
)

var supportedCollectionFormat = []string{
	collectionFormatComma, collectionFormatSpace, collectionFormatTab, collectionFormatPipe,
}

func getCollection(value string, supportedCollectionFormats []string) (items *spec.Items, collectionFormat string) {
	for _, collectionFormat := range supportedCollectionFormats {
		splitByFormat := swag.SplitByFormat(value, collectionFormat)
		// Will create a collection only if more then a single object exists
		if len(splitByFormat) > 1 {
			// TODO: Should we look at all elements to find the common type and formant?
			tpe, format := getTypeAndFormat(splitByFormat[0])
			return spec.NewItems().Typed(tpe, format), collectionFormat
		}
	}

	return nil, ""
}

func getTypeAndFormat(value string) (tpe string, format string) {
	if _, err := swag.ConvertInt64(value); err == nil {
		return schemaTypeInteger, ""
	}

	if _, err := swag.ConvertFloat64(value); err == nil {
		return schemaTypeNumber, ""
	}

	// TODO: not sure that `strconv.ParseBool` will do the job, it depends what is considers as boolean string
	// The Go implementation for example uses `strconv.FormatBool(value)` ==> true/false
	// But if we look at swag.ConvertBool - `checked` is evaluated as true so `unchecked` will be false?
	// Also when using `strconv.ParseBool` 1 is considered as true so we must check for int before running it
	if _, err := strconv.ParseBool(value); err == nil {
		return schemaTypeBoolean, ""
	}

	return schemaTypeString, getStringFormat(value)
}
