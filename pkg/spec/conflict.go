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

	spec "github.com/getkin/kin-openapi/openapi3"
	"k8s.io/utils/field"
)

type conflict struct {
	path *field.Path
	obj1 interface{}
	obj2 interface{}
	msg  string
}

func createConflictMsg(path *field.Path, t1, t2 interface{}) string {
	return fmt.Sprintf("%s: type mismatch: %+v != %+v", path, t1, t2)
}

func createHeaderInConflictMsg(path *field.Path, in, in2 interface{}) string {
	return fmt.Sprintf("%s: header in mismatch: %+v != %+v", path, in, in2)
}

func (c conflict) String() string {
	return c.msg
}

// conflictSolver will get 2 types and returns
// -1 - types conflict can't be resolved
//
//	0 - type1 and type2 are equal
//	1 - type1 should be used
//	2 - type2 should be used
func conflictSolver(type1, type2 string) int {
	if type1 == type2 {
		return 0
	}

	if shouldPreferType(type1, type2) {
		return 1
	}

	if shouldPreferType(type2, type1) {
		return 2
	}

	return -1
}

// shouldPreferType return true if type1 should be preferred over type2
// Note: MUST be called when type1 and type2 are not identical.
func shouldPreferType(type1, type2 string) bool {
	switch type1 {
	case spec.TypeBoolean, spec.TypeObject, spec.TypeArray:
		return false
	case spec.TypeNumber:
		// Preferring number to integer type.
		return type2 == spec.TypeInteger
	case spec.TypeString:
		// Preferring string to any type.
		return true
	}

	return false
}
