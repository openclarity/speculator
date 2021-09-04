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
	"time"

	"github.com/xeipuuv/gojsonschema"
)

var formats = []string{
	"date",
	"time",
	"date-time",
	"email",
	"ipv4",
	"ipv6",
	"uuid",
	"json-pointer",
	//"relative-json-pointer", // matched with "1.147.1"
	//"hostname",
	//"regex",
	//"uri",           // can be also iri
	//"uri-reference", // can be also iri-reference
	//"uri-template",
}

func getStringFormat(value interface{}) string {
	str, ok := value.(string)
	if !ok || str == "" {
		return ""
	}

	for _, format := range formats {
		if gojsonschema.FormatCheckers.IsFormat(format, value) {
			return format
		}
	}

	return ""
}

// isDateFormat checks if input is a correctly formatted date with spaces (excluding RFC3339 = "2006-01-02T15:04:05Z07:00")
// This is useful to identify date string instead of an collection
func isDateFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}
	if _, err := time.Parse(time.ANSIC, asString); err == nil {
		return true
	}
	if _, err := time.Parse(time.UnixDate, asString); err == nil {
		return true
	}
	if _, err := time.Parse(time.RubyDate, asString); err == nil {
		return true
	}
	if _, err := time.Parse(time.RFC822, asString); err == nil {
		return true
	}
	if _, err := time.Parse(time.RFC822Z, asString); err == nil {
		return true
	}
	if _, err := time.Parse(time.RFC850, asString); err == nil {
		return true
	}
	if _, err := time.Parse(time.RFC1123, asString); err == nil {
		return true
	}
	if _, err := time.Parse(time.RFC1123Z, asString); err == nil {
		return true
	}
	if _, err := time.Parse(time.Stamp, asString); err == nil {
		return true
	}
	if _, err := time.Parse(time.StampMilli, asString); err == nil {
		return true
	}
	if _, err := time.Parse(time.StampMicro, asString); err == nil {
		return true
	}
	if _, err := time.Parse(time.StampNano, asString); err == nil {
		return true
	}
	return false
}
