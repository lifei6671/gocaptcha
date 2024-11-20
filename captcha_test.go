package gocaptcha

import "testing"

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
