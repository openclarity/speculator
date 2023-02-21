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
	"testing"

	spec "github.com/getkin/kin-openapi/openapi3"
)

func Test_shouldPreferType(t *testing.T) {
	type args struct {
		t1 string
		t2 string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should not prefer - bool",
			args: args{
				t1: spec.TypeBoolean,
			},
			want: false,
		},
		{
			name: "should not prefer - obj",
			args: args{
				t1: spec.TypeObject,
			},
			want: false,
		},
		{
			name: "should not prefer - array",
			args: args{
				t1: spec.TypeArray,
			},
			want: false,
		},
		{
			name: "should not prefer - number over object",
			args: args{
				t1: spec.TypeNumber,
				t2: spec.TypeObject,
			},
			want: false,
		},
		{
			name: "prefer - number over int",
			args: args{
				t1: spec.TypeNumber,
				t2: spec.TypeInteger,
			},
			want: true,
		},
		{
			name: "prefer - string over anything",
			args: args{
				t1: spec.TypeString,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldPreferType(tt.args.t1, tt.args.t2); got != tt.want {
				t.Errorf("shouldPreferType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_conflictSolver(t *testing.T) {
	type args struct {
		t1 string
		t2 string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "no conflict",
			args: args{
				t1: spec.TypeNumber,
				t2: spec.TypeNumber,
			},
			want: NoConflict,
		},
		{
			name: "prefer string over anything",
			args: args{
				t1: spec.TypeString,
				t2: spec.TypeNumber,
			},
			want: PreferType1,
		},
		{
			name: "prefer string over anything",
			args: args{
				t1: spec.TypeInteger,
				t2: spec.TypeString,
			},
			want: PreferType2,
		},
		{
			name: "prefer number over int",
			args: args{
				t1: spec.TypeNumber,
				t2: spec.TypeInteger,
			},
			want: PreferType1,
		},
		{
			name: "prefer number over int",
			args: args{
				t1: spec.TypeInteger,
				t2: spec.TypeNumber,
			},
			want: PreferType2,
		},
		{
			name: "conflict - bool",
			args: args{
				t1: spec.TypeInteger,
				t2: spec.TypeBoolean,
			},
			want: ConflictUnresolved,
		},
		{
			name: "conflict - obj",
			args: args{
				t1: spec.TypeObject,
				t2: spec.TypeBoolean,
			},
			want: ConflictUnresolved,
		},
		{
			name: "conflict - array",
			args: args{
				t1: spec.TypeObject,
				t2: spec.TypeArray,
			},
			want: ConflictUnresolved,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := conflictSolver(tt.args.t1, tt.args.t2); got != tt.want {
				t.Errorf("conflictSolver() = %v, want %v", got, tt.want)
			}
		})
	}
}
