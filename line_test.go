package gocaptcha

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func Test_Beeline(t *testing.T) {
	type args struct {
		canvas    draw.Image
		x         image.Point
		y         image.Point
		lineColor color.Color
	}
	tests := []struct {
		name    string
		t       LineDrawer
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			t:    NewBeeline(),
			args: args{
				canvas:    image.NewRGBA(image.Rect(0, 0, 10, 10)),
				x:         image.Point{X: 11, Y: 5},
				y:         image.Point{X: 10, Y: 5},
				lineColor: color.RGBA{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.t.DrawLine(tt.args.canvas, tt.args.x, tt.args.y, tt.args.lineColor); (err != nil) != tt.wantErr {
				t.Errorf("beeline.DrawLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_CurveLine(t *testing.T) {
	type args struct {
		canvas    draw.Image
		x         image.Point
		y         image.Point
		lineColor color.Color
	}
	tests := []struct {
		name    string
		t       LineDrawer
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			t:    NewCurveLine(),
			args: args{
				canvas:    image.NewRGBA(image.Rect(0, 0, 10, 10)),
				x:         image.Point{X: 11, Y: 5},
				y:         image.Point{X: 10, Y: 5},
				lineColor: color.RGBA{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.t.DrawLine(tt.args.canvas, tt.args.x, tt.args.y, tt.args.lineColor); (err != nil) != tt.wantErr {
				t.Errorf("beeline.DrawLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_BezierLine(t *testing.T) {
	type args struct {
		canvas    draw.Image
		x         image.Point
		y         image.Point
		lineColor color.Color
	}
	tests := []struct {
		name    string
		t       LineDrawer
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			t:    NewBezierLine(),
			args: args{
				canvas:    image.NewRGBA(image.Rect(0, 0, 10, 10)),
				x:         image.Point{X: 11, Y: 5},
				y:         image.Point{X: 10, Y: 5},
				lineColor: color.RGBA{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.t.DrawLine(tt.args.canvas, tt.args.x, tt.args.y, tt.args.lineColor); (err != nil) != tt.wantErr {
				t.Errorf("beeline.DrawLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_HollowLine(t *testing.T) {
	type args struct {
		canvas    draw.Image
		x         image.Point
		y         image.Point
		lineColor color.Color
	}
	tests := []struct {
		name    string
		t       LineDrawer
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			t:    NewHollowLine(),
			args: args{
				canvas:    image.NewRGBA(image.Rect(0, 0, 10, 10)),
				x:         image.Point{X: 1, Y: 5},
				y:         image.Point{X: 10, Y: 5},
				lineColor: color.RGBA{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.t.DrawLine(tt.args.canvas, tt.args.x, tt.args.y, tt.args.lineColor); (err != nil) != tt.wantErr {
				t.Errorf("beeline.DrawLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
