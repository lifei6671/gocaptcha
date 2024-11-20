package gocaptcha

import (
	"image"
	"image/draw"
	"math/rand"
	"testing"
)

func Test_PointNoiseDrawer(t *testing.T) {
	type args struct {
		canvas  draw.Image
		density NoiseDensity
	}
	tests := []struct {
		name    string
		t       NoiseDrawer
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			t:    &pointNoiseDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas:  image.NewRGBA(image.Rect(0, 0, 100, 100)),
				density: NoiseDensityLower,
			},
			wantErr: false,
		},
		{
			name: "test2",
			t:    &pointNoiseDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas:  image.NewRGBA(image.Rect(0, 0, 100, 100)),
				density: NoiseDensityMedium,
			},
			wantErr: false,
		},
		{
			name: "test3",
			t:    &pointNoiseDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas:  image.NewRGBA(image.Rect(0, 0, 100, 100)),
				density: NoiseDensityHigh,
			},
			wantErr: false,
		},
		{
			name: "test4",
			t:    NewPointNoiseDrawer(),
			args: args{
				canvas:  image.NewRGBA(image.Rect(0, 0, 100, 100)),
				density: NoiseDensity(4),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.t.DrawNoise(tt.args.canvas, tt.args.density); (err != nil) != tt.wantErr {
				t.Errorf("pointNoiseDrawer.DrawNoise() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_TextNoiseDrawer(t *testing.T) {
	type args struct {
		canvas  draw.Image
		density NoiseDensity
	}
	tests := []struct {
		name    string
		t       NoiseDrawer
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			t:    &textNoiseDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas:  image.NewRGBA(image.Rect(0, 0, 100, 100)),
				density: NoiseDensityLower,
			},
			wantErr: false,
		},
		{
			name: "test2",
			t:    &textNoiseDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas:  image.NewRGBA(image.Rect(0, 0, 100, 100)),
				density: NoiseDensityMedium,
			},
			wantErr: false,
		},
		{
			name: "test3",
			t:    &textNoiseDrawer{r: rand.New(rand.NewSource(1))},
			args: args{
				canvas:  image.NewRGBA(image.Rect(0, 0, 100, 100)),
				density: NoiseDensityHigh,
			},
			wantErr: false,
		},
		{
			name: "test4",
			t:    NewTextNoiseDrawer(72),
			args: args{
				canvas:  image.NewRGBA(image.Rect(0, 0, 100, 100)),
				density: NoiseDensity(4),
			},
			wantErr: false,
		},
	}

	err := DefaultFontFamily.AddFontPath("./fonts")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.t.DrawNoise(tt.args.canvas, tt.args.density); (err != nil) != tt.wantErr {
				t.Errorf("textNoiseDrawer.DrawNoise() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
