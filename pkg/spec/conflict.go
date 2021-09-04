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

func (c conflict) String() string {
	return c.msg
}
