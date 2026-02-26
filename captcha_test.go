package gocaptcha

import (
	"bytes"
	"image/color"
	"testing"
)

func TestCaptchaImage_Encode(t *testing.T) {
	err := SetFontPath("./fonts")
	if err != nil {
		t.Fatal(err)
	}
	captchaImage := New(150, 20, RandLightColor())
	err = captchaImage.
		DrawBorder(RandDeepColor()).
		DrawNoise(NoiseDensityHigh, NewTextNoiseDrawer(72)).
		DrawNoise(NoiseDensityLower, NewPointNoiseDrawer()).
		DrawLine(NewBezier3DLine(), RandDeepColor()).
		DrawText(NewTwistTextDrawer(DefaultDPI, DefaultAmplitude, DefaultFrequency), RandText(4)).
		DrawLine(NewBeeline(), RandDeepColor()).
		DrawLine(NewHollowLine(), RandLightColor()).
		DrawBlur(NewGaussianBlur(), DefaultBlurKernelSize, DefaultBlurSigma).
		Error
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewWithOptions(t *testing.T) {
	bg := color.RGBA{R: 10, G: 20, B: 30, A: 255}
	captcha := NewWithOptions(120, 40, WithBackgroundColor(bg), WithRandomSeed(123))
	if captcha == nil {
		t.Fatal("NewWithOptions() returned nil")
	}
	if captcha.width != 120 || captcha.height != 40 {
		t.Fatalf("unexpected size: %dx%d", captcha.width, captcha.height)
	}
	want := color.NRGBA{R: bg.R, G: bg.G, B: bg.B, A: bg.A}
	if got := captcha.nrgba.NRGBAAt(0, 0); got != want {
		t.Fatalf("unexpected background color: got=%v want=%v", got, want)
	}
}

func TestCaptchaImage_DrawNoiseWithConfig(t *testing.T) {
	if err := SetFontPath("./fonts"); err != nil {
		t.Fatal(err)
	}

	captcha := New(150, 40, RandLightColor())
	err := captcha.DrawNoiseWithConfig(NewPointNoiseDrawer(), NoiseConfig{
		Density:              NoiseDensityHigh,
		PointDensityDivisor:  5,
		SecondaryPointChance: 0.9,
	}).Error
	if err != nil {
		t.Fatalf("DrawNoiseWithConfig(point) error = %v", err)
	}

	err = captcha.DrawNoiseWithConfig(NewTextNoiseDrawer(72), NoiseConfig{
		Density:            NoiseDensityMedium,
		TextDensityDivisor: 900,
		TextLength:         2,
		FontSizeJitter:     7,
	}).Error
	if err != nil {
		t.Fatalf("DrawNoiseWithConfig(text) error = %v", err)
	}
}

func TestCaptchaImage_WithRandomSeedForLinePointSelection(t *testing.T) {
	bg := color.RGBA{R: 240, G: 240, B: 240, A: 255}
	c1 := NewWithOptions(140, 40, WithBackgroundColor(bg), WithRandomSeed(2024))
	c2 := NewWithOptions(140, 40, WithBackgroundColor(bg), WithRandomSeed(2024))
	lineColor := color.RGBA{R: 10, G: 10, B: 10, A: 255}

	_ = c1.DrawLine(NewBeeline(), lineColor).Error
	_ = c2.DrawLine(NewBeeline(), lineColor).Error

	if !bytes.Equal(c1.nrgba.Pix, c2.nrgba.Pix) {
		t.Fatal("captchas with same random seed should produce identical line selection for deterministic line drawer")
	}
}
