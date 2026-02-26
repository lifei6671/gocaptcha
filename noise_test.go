package gocaptcha

import (
	"errors"
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

func TestNoiseDrawer_DrawNoiseWithConfig(t *testing.T) {
	pointDrawer, ok := NewPointNoiseDrawer().(ConfigurableNoiseDrawer)
	if !ok {
		t.Fatal("point drawer does not implement ConfigurableNoiseDrawer")
	}
	pointCanvas := image.NewRGBA(image.Rect(0, 0, 80, 40))
	if err := pointDrawer.DrawNoiseWithConfig(pointCanvas, NoiseConfig{
		Density:              NoiseDensityHigh,
		PointDensityDivisor:  6,
		SecondaryPointChance: 0.8,
	}); err != nil {
		t.Fatalf("point DrawNoiseWithConfig() error = %v", err)
	}

	if err := DefaultFontFamily.AddFontPath("./fonts"); err != nil {
		t.Fatal(err)
	}
	textDrawer, ok := NewTextNoiseDrawer(72).(ConfigurableNoiseDrawer)
	if !ok {
		t.Fatal("text drawer does not implement ConfigurableNoiseDrawer")
	}
	textCanvas := image.NewRGBA(image.Rect(0, 0, 100, 60))
	if err := textDrawer.DrawNoiseWithConfig(textCanvas, NoiseConfig{
		Density:            NoiseDensityMedium,
		TextDensityDivisor: 700,
		TextLength:         2,
		FontSizeJitter:     8,
	}); err != nil {
		t.Fatalf("text DrawNoiseWithConfig() error = %v", err)
	}

	if err := textDrawer.DrawNoiseWithConfig(nil, NoiseConfig{Density: NoiseDensityMedium}); !errors.Is(err, ErrNilCanvas) {
		t.Fatalf("DrawNoiseWithConfig(nil) error = %v, want %v", err, ErrNilCanvas)
	}
}

func TestAdvancedNoiseDrawers(t *testing.T) {
	tests := []struct {
		name    string
		drawer  NoiseDrawer
		density NoiseDensity
		wantErr bool
	}{
		{
			name:    "poisson-default",
			drawer:  NewPoissonPointNoiseDrawer(),
			density: NoiseDensityMedium,
			wantErr: false,
		},
		{
			name:    "poisson-custom",
			drawer:  NewPoissonPointNoiseDrawerWithConfig(6, 24, nil),
			density: NoiseDensityHigh,
			wantErr: false,
		},
		{
			name:    "perlin-default",
			drawer:  NewPerlinNoiseDrawer(),
			density: NoiseDensityMedium,
			wantErr: false,
		},
		{
			name:    "perlin-custom",
			drawer:  NewPerlinNoiseDrawerWithConfig(14, 4, 0.75, nil),
			density: NoiseDensityLower,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := image.NewRGBA(image.Rect(0, 0, 120, 60))
			err := tt.drawer.DrawNoise(canvas, tt.density)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DrawNoise() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
