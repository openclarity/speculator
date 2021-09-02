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
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-openapi/spec"
	uuid "github.com/satori/go.uuid"
)

type PathParam struct {
	*spec.Parameter
}

func generateParamName(i int) string {
	return fmt.Sprintf("param%v", i)
}

var digitCheck = regexp.MustCompile(`^[0-9]+$`)

func createParameterizedPath(path string) string {
	var ParameterizedPathParts []string
	paramCount := 0
	pathParts := strings.Split(path, "/")

	for _, part := range pathParts {
		// if part is a suspect param, replace it with a param name, otherwise do nothing
		if isSuspectPathParam(part) {
			paramCount++
			paramName := generateParamName(paramCount)
			ParameterizedPathParts = append(ParameterizedPathParts, "{"+paramName+"}")
		} else {
			ParameterizedPathParts = append(ParameterizedPathParts, part)
		}
	}

	parameterizedPath := strings.Join(ParameterizedPathParts, "/")

	return parameterizedPath
}

type paramFormat string

const (
	paramFormatUnset  paramFormat = "paramFormatUnset"
	paramFormatNumber paramFormat = "paramFormatNumber"
	paramFormatUUID   paramFormat = "paramFormatUUID"
	paramFormatMixed  paramFormat = "paramFormatMixed"
)

// /api/1/foo, api/2/foo and index 1 will return:
// []string{1, 2}
func getOnlyIndexedPartFromPaths(paths map[string]bool, i int) []string {
	var ret []string
	for path := range paths {
		path = strings.TrimPrefix(path, "/")
		splt := strings.Split(path, "/")
		if len(splt) <= i {
			continue
		}
		ret = append(ret, splt[i])
	}
	return ret
}

// If all params in paramList can be guessed as same type and format, this type and format will be returned, otherwise,
// if there are couple of formats, type string and no format will be return
func getParamTypeAndFormat(paramsList []string) (string, string) {
	var parameterFormat = paramFormatUnset

	for _, pathPart := range paramsList {
		if isNumber(pathPart) {
			// in case there is a conflict, we will return string as the type and empty format
			if parameterFormat != paramFormatNumber && parameterFormat != paramFormatUnset {
				return schemaTypeString, ""
			}
			parameterFormat = paramFormatNumber
			continue
		}
		if isUUID(pathPart) {
			if parameterFormat != paramFormatUUID && parameterFormat != paramFormatUnset {
				return schemaTypeString, ""
			}
			parameterFormat = paramFormatUUID
			continue
		}
		if isMixed(pathPart) {
			if parameterFormat != paramFormatMixed && parameterFormat != paramFormatUnset {
				return schemaTypeString, ""
			}
			parameterFormat = paramFormatMixed
		}
	}

	switch parameterFormat {
	case paramFormatMixed:
		return schemaTypeString, ""
	case paramFormatUUID:
		return schemaTypeString, formatUUID
	case paramFormatNumber:
		return schemaTypeInteger, ""
	case paramFormatUnset:
		return schemaTypeString, ""
	}

	return schemaTypeString, ""
}

func isSuspectPathParam(pathPart string) bool {
	if isNumber(pathPart) {
		return true
	}
	if isUUID(pathPart) {
		return true
	}
	if isMixed(pathPart) {
		return true
	}
	return false
}

func isNumber(pathPart string) bool {
	return digitCheck.MatchString(pathPart)
}

func isUUID(pathPart string) bool {
	_, err := uuid.FromString(pathPart)
	return err == nil
}

// Check if a path part that is mixed from digits and chars can be considered as parameter following hard-coded heuristics.
// Temporary, we'll consider strings as parameters that are at least 8 chars longs and has at least 3 digits.
func isMixed(pathPart string) bool {
	if len(pathPart) < 8 {
		return false
	}

	return countDigitsInString(pathPart) > 2
}

func countDigitsInString(s string) int {
	count := 0
	for _, c := range s {
		if unicode.IsNumber(c) {
			count++
		}
	}
	return count
}

func createPathParam(name, tpe, format string) *PathParam {
	var pathParam PathParam

	pathParam.Parameter = spec.PathParam(name)
	pathParam.Typed(tpe, format)

	return &pathParam
}
