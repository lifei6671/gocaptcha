package gocaptcha

import (
	"errors"
	"image"
	"image/draw"
	"math/rand"
	"testing"
)

func Test_textDrawer_DrawString(t *testing.T) {
	type args struct {
		canvas draw.Image
		text   string
	}
	tests := []struct {
		name    string
		t       *textDrawer
		args    args
		wantErr bool
	}{
		{
			name: "Successful DrawString",
			t:    &textDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: image.NewRGBA(image.Rect(0, 0, 100, 100)),
				text:   "Hello, World!",
			},
			wantErr: false,
		},
		{
			name: "DrawString with empty text",
			t:    &textDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: image.NewRGBA(image.Rect(0, 0, 100, 100)),
				text:   "",
			},
			wantErr: true,
		},
		{
			name: "DrawString with nil canvas",
			t:    &textDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: nil,
				text:   "Hello, World!",
			},
			wantErr: true,
		},
		{
			name: "DrawString with tiny canvas",
			t:    &textDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: image.NewRGBA(image.Rect(0, 0, 1, 1)),
				text:   "A",
			},
			wantErr: false,
		},
		{
			name: "DrawString with unicode text",
			t:    &textDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: image.NewRGBA(image.Rect(0, 0, 100, 100)),
				text:   "验证码A",
			},
			wantErr: false,
		},
		{
			name: "DrawString with zero size canvas",
			t:    &textDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: image.NewRGBA(image.Rect(0, 0, 0, 0)),
				text:   "A",
			},
			wantErr: true,
		},
	}
	err := DefaultFontFamily.AddFontPath("./fonts")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.t.DrawString(tt.args.canvas, tt.args.text); (err != nil) != tt.wantErr {
				t.Errorf("textDrawer.DrawString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_twistTextDrawer_DrawString(t *testing.T) {
	type args struct {
		canvas draw.Image
		text   string
	}
	tests := []struct {
		name    string
		t       *twistTextDrawer
		args    args
		wantErr bool
	}{
		{
			name: "Successful DrawString",
			t:    &twistTextDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: image.NewRGBA(image.Rect(0, 0, 100, 100)),
				text:   "Hello, World!",
			},
			wantErr: false,
		},
		{
			name: "DrawString with empty text",
			t:    &twistTextDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: image.NewRGBA(image.Rect(0, 0, 100, 100)),
				text:   "",
			},
			wantErr: true,
		},
		{
			name: "DrawString with nil canvas",
			t:    &twistTextDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: nil,
				text:   "Hello, World!",
			},
			wantErr: true,
		},
		{
			name: "DrawString with tiny canvas",
			t:    &twistTextDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: image.NewRGBA(image.Rect(0, 0, 1, 1)),
				text:   "A",
			},
			wantErr: false,
		},
		{
			name: "DrawString with unicode text",
			t:    &twistTextDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: image.NewRGBA(image.Rect(0, 0, 100, 100)),
				text:   "验证码A",
			},
			wantErr: false,
		},
		{
			name: "DrawString with zero size canvas",
			t:    &twistTextDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas: image.NewRGBA(image.Rect(0, 0, 0, 0)),
				text:   "A",
			},
			wantErr: true,
		},
	}
	err := DefaultFontFamily.AddFontPath("./fonts")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.t.DrawString(tt.args.canvas, tt.args.text); (err != nil) != tt.wantErr {
				t.Errorf("twistTextDrawer.DrawString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewTwistTextDrawerWithModes(t *testing.T) {
	err := DefaultFontFamily.AddFontPath("./fonts")
	if err != nil {
		t.Fatal(err)
	}

	canvas := image.NewRGBA(image.Rect(0, 0, 120, 50))
	drawer := NewTwistTextDrawerWithModes(72, 12, 0.05, WaveDistortionVertical)
	if err := drawer.DrawString(canvas, "Captcha"); err != nil {
		t.Fatalf("NewTwistTextDrawerWithModes DrawString() error = %v", err)
	}
}

func TestNewEffectTextDrawer(t *testing.T) {
	err := DefaultFontFamily.AddFontPath("./fonts")
	if err != nil {
		t.Fatal(err)
	}

	canvas := image.NewRGBA(image.Rect(0, 0, 120, 50))
	drawer := NewEffectTextDrawer(
		72,
		NewWaveTextEffect(10, 0.05, WaveDistortionHorizontal),
		NewWaveTextEffect(6, 0.03, WaveDistortionVertical),
	)
	if err := drawer.DrawString(canvas, "Captcha"); err != nil {
		t.Fatalf("NewEffectTextDrawer DrawString() error = %v", err)
	}
}

func TestWaveTextEffect_Apply(t *testing.T) {
	effect := NewWaveTextEffect(10, 0.05, WaveDistortionHorizontal)

	src := image.NewRGBA(image.Rect(0, 0, 20, 20))
	dst := image.NewRGBA(image.Rect(0, 0, 10, 20))

	err := effect.Apply(src, dst)
	if !errors.Is(err, ErrInvalidEffectSize) {
		t.Fatalf("wave effect Apply() error = %v, want %v", err, ErrInvalidEffectSize)
	}
}
