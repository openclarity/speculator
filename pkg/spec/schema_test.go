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
	"reflect"
	"testing"

	spec "github.com/getkin/kin-openapi/openapi3"
)

func Test_splitByStyle(t *testing.T) {
	type args struct {
		data  string
		style string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty data",
			args: args{
				data:  "",
				style: "",
			},
			want: nil,
		},
		{
			name: "unsupported serialization style",
			args: args{
				data:  "",
				style: "Unsupported",
			},
			want: nil,
		},
		{
			name: "SerializationForm",
			args: args{
				data:  "1, 2, 3",
				style: spec.SerializationForm,
			},
			want: []string{"1", "2", "3"},
		},
		{
			name: "SerializationSimple",
			args: args{
				data:  "1, 2, 3",
				style: spec.SerializationSimple,
			},
			want: []string{"1", "2", "3"},
		},
		{
			name: "SerializationSpaceDelimited",
			args: args{
				data:  "1 2  3",
				style: spec.SerializationSpaceDelimited,
			},
			want: []string{"1", "2", "3"},
		},
		{
			name: "SerializationPipeDelimited",
			args: args{
				data:  "1|2|3",
				style: spec.SerializationPipeDelimited,
			},
			want: []string{"1", "2", "3"},
		},
		{
			name: "SerializationPipeDelimited with empty space in the middle",
			args: args{
				data:  "1| |3",
				style: spec.SerializationPipeDelimited,
			},
			want: []string{"1", "3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitByStyle(tt.args.data, tt.args.style); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitByStyle() = %v, want %v", got, tt.want)
			}
		})
	}
}
