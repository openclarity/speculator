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
	"reflect"
	"testing"

	"github.com/go-openapi/spec"
)

func Test_populateParam(t *testing.T) {
	type args struct {
		parameter       *spec.Parameter
		values          []string
		allowCollection bool
	}
	tests := []struct {
		name string
		args args
		want *spec.Parameter
	}{
		{
			name: "parameters without a value",
			args: args{
				parameter:       spec.QueryParam("test"),
				values:          []string{""},
				allowCollection: true,
			},
			want: spec.QueryParam("test").Typed(schemaTypeBoolean, "").AllowsEmptyValues().AsRequired(),
		},
		{
			name: "parameters without a value - no query or form data",
			args: args{
				parameter:       spec.HeaderParam("test"),
				values:          []string{""},
				allowCollection: true,
			},
			want: spec.HeaderParam("test"),
		},
		{
			name: "Multiple parameter instances",
			args: args{
				parameter:       spec.FormDataParam("test"),
				values:          []string{"1", "2", "3"},
				allowCollection: true,
			},
			want: spec.FormDataParam("test").CollectionOf(spec.NewItems().Typed(schemaTypeInteger, ""), collectionFormatMulti),
		},
		{
			name: "Multiple parameter instances - no query or form data",
			args: args{
				parameter:       spec.HeaderParam("test"),
				values:          []string{""},
				allowCollection: true,
			},
			want: spec.HeaderParam("test"),
		},
		{
			name: "collection",
			args: args{
				parameter:       spec.HeaderParam("test"),
				values:          []string{"a b c"},
				allowCollection: true,
			},
			want: spec.HeaderParam("test").CollectionOf(spec.NewItems().Typed(schemaTypeString, ""), collectionFormatSpace),
		},
		{
			name: "collection - allowCollection == false ",
			args: args{
				parameter:       spec.HeaderParam("test"),
				values:          []string{"a b c"},
				allowCollection: false,
			},
			want: spec.HeaderParam("test").Typed(schemaTypeString, ""),
		},
		{
			name: "simple type and format",
			args: args{
				parameter:       spec.HeaderParam("test"),
				values:          []string{"a"},
				allowCollection: true,
			},
			want: spec.HeaderParam("test").Typed(schemaTypeString, ""),
		},
		{
			name: "date",
			args: args{
				parameter:       spec.HeaderParam("date"),
				values:          []string{"Mon, 23 Aug 2021 06:52:48 GMT"},
				allowCollection: true,
			},
			want: spec.HeaderParam("date").Typed(schemaTypeString, ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := populateParam(tt.args.parameter, tt.args.values, tt.args.allowCollection); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("populateParam() = %v, want %v", got, tt.want)
			}
		})
	}
}
