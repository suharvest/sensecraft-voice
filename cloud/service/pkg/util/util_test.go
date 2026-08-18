package util

import (
	"reflect"
	"testing"
)

func TestValidateStrongPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "case1",
			password: "123456",
			want:     false,
		},
		{
			name:     "case2",
			password: "12345678",
			want:     false,
		},
		{
			name:     "case3",
			password: "12345678a",
			want:     false,
		},
		{
			name:     "case4",
			password: "12345678A",
			want:     false,
		},
		{
			name:     "case5",
			password: "12345678aA",
			want:     true,
		},
		{
			name:     "case6",
			password: "123456Aa",
			want:     true,
		},
		{
			name:     "case7",
			password: "abcdefgh",
			want:     false,
		},
		{
			name:     "case8",
			password: "ABCDEFGH",
			want:     false,
		},
		{
			name:     "case9",
			password: "abcdef12",
			want:     false,
		},
		{
			name:     "case10",
			password: "ABCDEF12",
			want:     false,
		},
		{
			name:     "case11",
			password: "Abcdef12",
			want:     true,
		},
		{
			name:     "case12",
			password: "aBCDEF12",
			want:     true,
		},
		{
			name:     "case13",
			password: "$$$$$$$$",
			want:     false,
		},
		{
			name:     "case14",
			password: "$$$$$aA1",
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateStrongPassword(tt.password); got != tt.want {
				t.Errorf("ValidateStrongPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeduplicateIntSlice(t *testing.T) {
	type args struct {
		s []int64
	}
	tests := []struct {
		name string
		args args
		want []int64
	}{
		{
			name: "case 1",
			args: args{
				s: []int64{1, 2, 3, 4, 5},
			},
			want: []int64{1, 2, 3, 4, 5},
		},
		{
			name: "case 2",
			args: args{
				s: []int64{1, 1, 1, 1, 1},
			},
			want: []int64{1},
		},
		{
			name: "case 3",
			args: args{
				s: []int64{1, 1, 2, 3, 4, 5},
			},
			want: []int64{1, 2, 3, 4, 5},
		},
		{
			name: "case 4",
			args: args{
				s: []int64{1, 1, 2, 1, 3, 4, 5},
			},
			want: []int64{1, 2, 3, 4, 5},
		},
		{
			name: "case 5",
			args: args{
				s: []int64{1, 2, 3, 4, 5, 5, 5, 5},
			},
			want: []int64{1, 2, 3, 4, 5},
		},
		{
			name: "case 6",
			args: args{
				s: []int64{5, 1, 4, 2, 3, 3, 2, 4, 1, 5},
			},
			want: []int64{5, 1, 4, 2, 3},
		},
		{
			name: "case 7",
			args: args{
				s: []int64{1, 1, 1, 2, 3, 3, 4, 4, 5, 5, 5},
			},
			want: []int64{1, 2, 3, 4, 5},
		},
		{
			name: "case 8",
			args: args{
				s: []int64{1, 2, 3, 4, 5, 5, 5, 5, 4, 3, 2, 1},
			},
			want: []int64{1, 2, 3, 4, 5},
		},
		{
			name: "case 9",
			args: args{
				s: []int64{1, 2, 3, 4, 5, 5, 5, 5, 4, 3, 2, 1, 1, 1},
			},
			want: []int64{1, 2, 3, 4, 5},
		},
		{
			name: "case 10",
			args: args{
				s: []int64{1, 2, 3, 4, 5, 5, 5, 5, 4, 3, 2, 1, 1, 1, 2, 3, 4, 5},
			},
			want: []int64{1, 2, 3, 4, 5},
		},
		{
			name: "case 11",
			args: args{
				s: []int64{},
			},
			want: []int64{},
		},
		{
			name: "case 12",
			args: args{
				s: nil,
			},
			want: []int64{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeduplicateIntSlice(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeduplicateIntSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
