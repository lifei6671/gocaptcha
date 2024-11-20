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

func TestRandDeepColor(t *testing.T) {
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
			want: color.RGBA{R: 50, G: 50, B: 50, A: 0},
		},
		{
			name: "Test 2",
			args: args{colorVal: 0x00FF00},
			want: color.RGBA{R: 0, G: 50, B: 0, A: 0},
		},
		{
			name: "Test 3",
			args: args{colorVal: 0x0000FF},
			want: color.RGBA{R: 0, G: 0, B: 50, A: 50},
		},
		{
			name: "Test 4",
			args: args{colorVal: 0xFFFFFF},
			want: color.RGBA{R: 50, G: 50, B: 50, A: 50},
		},
		{
			name: "Test 5",
			args: args{colorVal: 0x000000},
			want: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RandDeepColor(); got.R <= tt.want.R && got.G <= tt.want.G && got.B <= tt.want.B {
				t.Errorf("RandDeepColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandLightColor(t *testing.T) {
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
			want: color.RGBA{R: 128, G: 128, B: 128, A: 255},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RandLightColor(); got.R <= tt.want.R && got.G <= tt.want.G && got.B <= tt.want.B {
				t.Errorf("RandLightColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandColor(t *testing.T) {
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
			want: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		{
			name: "Test 2",
			args: args{colorVal: 0x00FF00},
			want: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RandColor(); got.R >= tt.want.R && got.G >= tt.want.G && got.B >= tt.want.B {
				t.Errorf("RandColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandText(t *testing.T) {
	type args struct {
		num int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Test 1",
			args: args{num: 1},
			want: 1,
		},
		{
			name: "Test 2",
			args: args{num: 2},
			want: 2,
		},
		{
			name: "Test 3",
			args: args{num: 3},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RandText(tt.args.num); len(got) != tt.want {
				t.Errorf("RandText() = %v, want %v", got, tt.want)
			}
		})
	}
}
