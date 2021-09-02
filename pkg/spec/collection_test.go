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

func Test_getCollection(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name                 string
		args                 args
		wantItems            *spec.Items
		wantCollectionFormat string
	}{
		{
			name: "collectionFormatComma",
			args: args{
				value: "1,2,3,4",
			},
			wantItems:            spec.NewItems().Typed("integer", ""),
			wantCollectionFormat: collectionFormatComma,
		},
		{
			name: "collectionFormatSpace",
			args: args{
				value: "a b c d",
			},
			wantItems:            spec.NewItems().Typed("string", ""),
			wantCollectionFormat: collectionFormatSpace,
		},
		{
			name: "collectionFormatTab",
			args: args{
				value: "true\tfalse\ttrue\tfalse",
			},
			wantItems:            spec.NewItems().Typed("boolean", ""),
			wantCollectionFormat: collectionFormatTab,
		},
		{
			name: "collectionFormatPipe",
			args: args{
				value: "14.0|12.2|13.5|15.8",
			},
			wantItems:            spec.NewItems().Typed("number", ""),
			wantCollectionFormat: collectionFormatPipe,
		},
		{
			name: "not a collection",
			args: args{
				value: "14.0",
			},
			wantItems:            nil,
			wantCollectionFormat: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotItems, gotCollectionFormat := getCollection(tt.args.value, supportedCollectionFormat)
			if !reflect.DeepEqual(gotItems, tt.wantItems) {
				t.Errorf("getCollection() gotItems = %v, want %v", gotItems, tt.wantItems)
			}
			if gotCollectionFormat != tt.wantCollectionFormat {
				t.Errorf("getCollection() gotCollectionFormat = %v, want %v", gotCollectionFormat, tt.wantCollectionFormat)
			}
		})
	}
}

func Test_getTypeAndFormat(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name       string
		args       args
		wantTpe    string
		wantFormat string
	}{
		{
			name: "boolean false",
			args: args{
				value: "false",
			},
			wantTpe:    "boolean",
			wantFormat: "",
		},
		{
			name: "boolean true",
			args: args{
				value: "true",
			},
			wantTpe:    "boolean",
			wantFormat: "",
		},
		{
			name: "integer",
			args: args{
				value: "12",
			},
			wantTpe:    "integer",
			wantFormat: "",
		},
		{
			name: "number",
			args: args{
				value: "12.2",
			},
			wantTpe:    "number",
			wantFormat: "",
		},
		{
			name: "string uuid",
			args: args{
				value: "77e1c83b-7bb0-437b-bc50-a7a58e5660ac",
			},
			wantTpe:    "string",
			wantFormat: "uuid",
		},
		{
			name: "string no format",
			args: args{
				value: "string no format",
			},
			wantTpe:    "string",
			wantFormat: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTpe, gotFormat := getTypeAndFormat(tt.args.value)
			if gotTpe != tt.wantTpe {
				t.Errorf("getTypeAndFormat() gotTpe = %v, want %v", gotTpe, tt.wantTpe)
			}
			if gotFormat != tt.wantFormat {
				t.Errorf("getTypeAndFormat() gotFormat = %v, want %v", gotFormat, tt.wantFormat)
			}
		})
	}
}
