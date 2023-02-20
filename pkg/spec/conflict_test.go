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
			name: "should not prefer - number",
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
			want: 0,
		},
		{
			name: "prefer string over anything",
			args: args{
				t1: spec.TypeString,
				t2: spec.TypeNumber,
			},
			want: 1,
		},
		{
			name: "prefer string over anything",
			args: args{
				t1: spec.TypeInteger,
				t2: spec.TypeString,
			},
			want: 2,
		},
		{
			name: "prefer number over int",
			args: args{
				t1: spec.TypeNumber,
				t2: spec.TypeInteger,
			},
			want: 1,
		},
		{
			name: "prefer number over int",
			args: args{
				t1: spec.TypeInteger,
				t2: spec.TypeNumber,
			},
			want: 2,
		},
		{
			name: "conflict - bool",
			args: args{
				t1: spec.TypeInteger,
				t2: spec.TypeBoolean,
			},
			want: -1,
		},
		{
			name: "conflict - obj",
			args: args{
				t1: spec.TypeObject,
				t2: spec.TypeBoolean,
			},
			want: -1,
		},
		{
			name: "conflict - array",
			args: args{
				t1: spec.TypeObject,
				t2: spec.TypeArray,
			},
			want: -1,
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
