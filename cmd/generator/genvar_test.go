package main

import "testing"

func Test_formatVarName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple test",
			args: args{name: "buildVersion"},
			want: "Build version",
		},
		{
			name: "simple test",
			args: args{name: "buildURLAddress"},
			want: "Build URL address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatVarName(tt.args.name); got != tt.want {
				t.Errorf("formatVarName() = %v, want %v", got, tt.want)
			}
		})
	}
}
