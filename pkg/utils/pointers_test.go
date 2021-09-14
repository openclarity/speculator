package utils

import "testing"

func TestIsNil(t *testing.T) {
	integer := 1
	var intPointer *int
	type args struct {
		a interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil",
			args: args{
				a: nil,
			},
			want: true,
		},
		{
			name: "nil pointer",
			args: args{
				a: intPointer,
			},
			want: true,
		},
		{
			name: "not nil",
			args: args{
				a: &integer,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNil(tt.args.a); got != tt.want {
				t.Errorf("IsNil() = %v, want %v", got, tt.want)
			}
		})
	}
}
