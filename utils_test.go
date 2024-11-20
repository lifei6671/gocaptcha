package gocaptcha

import (
	"image/color"
	"testing"
)

func Test_abs(t *testing.T) {
	type args struct {
		a int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Positive number",
			args: args{a: 10},
			want: 10,
		},
		{
			name: "Negative number",
			args: args{a: -10},
			want: 10,
		},
		{
			name: "Zero",
			args: args{a: 0},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := abs(tt.args.a); got != tt.want {
				t.Errorf("abs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorToRGB(t *testing.T) {
	type args struct {
		colorVal int
	}
	tests := []struct {
		name string
		args args
		want color.RGBA
	}{
		{
			name: "Test 1",
			args: args{colorVal: 0xFF0000},
			want: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		},
		{
			name: "Test 2",
			args: args{colorVal: 0x00FF00},
			want: color.RGBA{R: 0, G: 255, B: 0, A: 255},
		},
		{
			name: "Test 3",
			args: args{colorVal: 0x0000FF},
			want: color.RGBA{R: 0, G: 0, B: 255, A: 255},
		},
		{
			name: "Test 4",
			args: args{colorVal: 0xFFFFFF},
			want: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		{
			name: "Test 5",
			args: args{colorVal: 0x000000},
			want: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ColorToRGB(tt.args.colorVal); got != tt.want {
				t.Errorf("ColorToRGB() = %v, want %v", got, tt.want)
			}
		})
	}
}
