package gocaptcha

import (
	"image"
	"image/color"
	"image/draw"
	"sync"
	"testing"
)

var benchmarkFontOnce sync.Once

func ensureBenchmarkFonts(b *testing.B) {
	b.Helper()
	benchmarkFontOnce.Do(func() {
		if err := DefaultFontFamily.AddFontPath("./fonts"); err != nil {
			b.Fatalf("load benchmark fonts: %v", err)
		}
	})
}

func fillBenchmarkPattern(img *image.NRGBA) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r := uint8((x * 255) / maxInt(1, bounds.Dx()))
			g := uint8((y * 255) / maxInt(1, bounds.Dy()))
			b := uint8((x + y) % 256)
			img.Set(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}
}

func BenchmarkGaussianBlur(b *testing.B) {
	drawer := NewGaussianBlur()
	base := image.NewNRGBA(image.Rect(0, 0, 220, 80))
	fillBenchmarkPattern(base)
	canvas := image.NewNRGBA(base.Bounds())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.Draw(canvas, canvas.Bounds(), base, image.Point{}, draw.Src)
		if err := drawer.DrawBlur(canvas, 7, 1.2); err != nil {
			b.Fatalf("DrawBlur() failed: %v", err)
		}
	}
}

func BenchmarkTextDraw(b *testing.B) {
	ensureBenchmarkFonts(b)
	drawer := NewTextDrawer(DefaultDPI)
	canvas := image.NewNRGBA(image.Rect(0, 0, 220, 80))
	bg := image.NewUniform(color.NRGBA{R: 245, G: 245, B: 245, A: 255})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.Draw(canvas, canvas.Bounds(), bg, image.Point{}, draw.Src)
		if err := drawer.DrawString(canvas, "ABCD"); err != nil {
			b.Fatalf("DrawString() failed: %v", err)
		}
	}
}

func BenchmarkTwistTextDraw(b *testing.B) {
	ensureBenchmarkFonts(b)
	drawer := NewTwistTextDrawerWithModes(
		DefaultDPI,
		DefaultAmplitude,
		DefaultFrequency,
		WaveDistortionHorizontal,
		WaveDistortionVertical,
		WaveDistortionDual,
	)
	canvas := image.NewNRGBA(image.Rect(0, 0, 220, 80))
	bg := image.NewUniform(color.NRGBA{R: 245, G: 245, B: 245, A: 255})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.Draw(canvas, canvas.Bounds(), bg, image.Point{}, draw.Src)
		if err := drawer.DrawString(canvas, "ABCD"); err != nil {
			b.Fatalf("DrawString() failed: %v", err)
		}
	}
}

func BenchmarkPointNoiseDraw(b *testing.B) {
	drawer := NewPointNoiseDrawer()
	canvas := image.NewNRGBA(image.Rect(0, 0, 220, 80))
	bg := image.NewUniform(color.NRGBA{R: 245, G: 245, B: 245, A: 255})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.Draw(canvas, canvas.Bounds(), bg, image.Point{}, draw.Src)
		if err := drawer.DrawNoise(canvas, NoiseDensityHigh); err != nil {
			b.Fatalf("DrawNoise() failed: %v", err)
		}
	}
}

func BenchmarkPoissonNoiseDraw(b *testing.B) {
	drawer := NewPoissonPointNoiseDrawer()
	canvas := image.NewNRGBA(image.Rect(0, 0, 220, 80))
	bg := image.NewUniform(color.NRGBA{R: 245, G: 245, B: 245, A: 255})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.Draw(canvas, canvas.Bounds(), bg, image.Point{}, draw.Src)
		if err := drawer.DrawNoise(canvas, NoiseDensityHigh); err != nil {
			b.Fatalf("DrawNoise() failed: %v", err)
		}
	}
}

func BenchmarkPerlinNoiseDraw(b *testing.B) {
	drawer := NewPerlinNoiseDrawer()
	canvas := image.NewNRGBA(image.Rect(0, 0, 220, 80))
	bg := image.NewUniform(color.NRGBA{R: 245, G: 245, B: 245, A: 255})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.Draw(canvas, canvas.Bounds(), bg, image.Point{}, draw.Src)
		if err := drawer.DrawNoise(canvas, NoiseDensityMedium); err != nil {
			b.Fatalf("DrawNoise() failed: %v", err)
		}
	}
}
