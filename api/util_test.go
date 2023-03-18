package main

import "testing"

func Test_firstN(t *testing.T) {
	type args struct {
		s string
		n int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"01", args{"hello", 2}, "he"},
		{"02", args{"hello", 5}, "hello"},
		{"02", args{"hello", 50}, "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstN(tt.args.s, tt.args.n); got != tt.want {
				t.Errorf("firstN() = %v, want %v", got, tt.want)
			}
		})
	}
}
