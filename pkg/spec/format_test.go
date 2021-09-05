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

import "testing"

// format taken from time/format.go.
func Test_isDateFormat(t *testing.T) {
	type args struct {
		input interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "RFC3339 should not match",
			args: args{
				input: "2021-08-23T06:52:48Z03:00",
			},
			want: false,
		},
		{
			name: "StampNano",
			args: args{
				input: "Aug 23 06:52:48.000000000",
			},
			want: true,
		},
		{
			name: "StampMicro",
			args: args{
				input: "Aug 23 06:52:48.000000",
			},
			want: true,
		},
		{
			name: "StampMilli",
			args: args{
				input: "Aug 23 06:52:48.000",
			},
			want: true,
		},
		{
			name: "Stamp",
			args: args{
				input: "Aug 23 06:52:48",
			},
			want: true,
		},
		{
			name: "RFC1123Z",
			args: args{
				input: "Mon, 23 Aug 2021 06:52:48 -0300",
			},
			want: true,
		},
		{
			name: "RFC1123",
			args: args{
				input: "Mon, 23 Aug 2021 06:52:48 GMT",
			},
			want: true,
		},
		{
			name: "RFC850",
			args: args{
				input: "Monday, 23-Aug-21 06:52:48 GMT",
			},
			want: true,
		},
		{
			name: "RFC822Z",
			args: args{
				input: "23 Aug 21 06:52 -0300",
			},
			want: true,
		},
		{
			name: "RFC822",
			args: args{
				input: "23 Aug 21 06:52 GMT",
			},
			want: true,
		},
		{
			name: "RubyDate",
			args: args{
				input: "Mon Aug 23 06:52:48 -0300 2021",
			},
			want: true,
		},
		{
			name: "UnixDate",
			args: args{
				input: "Mon Aug 23 06:52:48 GMT 2021",
			},
			want: true,
		},
		{
			name: "ANSIC",
			args: args{
				input: "Mon Aug 23 06:52:48 2021",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDateFormat(tt.args.input); got != tt.want {
				t.Errorf("isDateFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
