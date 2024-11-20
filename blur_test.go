package gocaptcha

import (
	"image"
	"image/draw"
	"testing"
)

func TestDrawBlur(t *testing.T) {
	type args struct {
		canvas     draw.Image
		kernelSize int
		sigma      float64
	}
	tests := []struct {
		name    string
		t       BlurDrawer
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			t:    NewGaussianBlur(),
			args: args{
				canvas:     image.NewRGBA(image.Rect(0, 0, 100, 100)),
				kernelSize: 5,
				sigma:      1.0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.t.DrawBlur(tt.args.canvas, tt.args.kernelSize, tt.args.sigma); (err != nil) != tt.wantErr {
				t.Errorf("textDrawer.DrawString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
