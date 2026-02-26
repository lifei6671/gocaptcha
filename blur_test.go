package gocaptcha

import (
	"image"
	"image/draw"
	"testing"
)

func TestDrawBlur(t *testing.T) {
	tests := []struct {
		name       string
		canvas     draw.Image
		kernelSize int
		sigma      float64
		wantErr    bool
	}{
		{
			name:       "standard blur",
			canvas:     image.NewRGBA(image.Rect(0, 0, 100, 100)),
			kernelSize: 5,
			sigma:      1.0,
			wantErr:    false,
		},
		{
			name:       "invalid params fallback",
			canvas:     image.NewRGBA(image.Rect(0, 0, 50, 50)),
			kernelSize: 2,
			sigma:      0,
			wantErr:    false,
		},
		{
			name:       "non-zero bounds",
			canvas:     image.NewRGBA(image.Rect(10, 10, 40, 40)),
			kernelSize: 7,
			sigma:      1.4,
			wantErr:    false,
		},
		{
			name:       "nil canvas",
			canvas:     nil,
			kernelSize: 5,
			sigma:      1.0,
			wantErr:    true,
		},
	}

	drawer := NewGaussianBlur()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := drawer.DrawBlur(tt.canvas, tt.kernelSize, tt.sigma)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DrawBlur() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGaussianBlurKernelCache(t *testing.T) {
	g, ok := NewGaussianBlur().(*gaussianBlur)
	if !ok {
		t.Fatal("NewGaussianBlur() did not return *gaussianBlur")
	}

	k1 := g.getKernel(5, 1.0)
	k2 := g.getKernel(5, 1.0)

	if len(k1) == 0 || len(k2) == 0 {
		t.Fatal("cached kernel should not be empty")
	}
	if &k1[0] != &k2[0] {
		t.Fatal("expected same cached kernel instance")
	}
}
