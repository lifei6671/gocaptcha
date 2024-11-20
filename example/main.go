package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/lifei6671/gocaptcha"
)

const (
	dx = 180
	dy = 60
)

func main() {
	http.HandleFunc("/", Index)
	http.HandleFunc("/get/", Get)
	fmt.Println("服务已启动 -> http://127.0.0.1:8800")
	err := http.ListenAndServe(":8800", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("tpl/index.html")
	if err != nil {
		log.Fatal(err)
	}
	_ = t.Execute(w, nil)
}
func Get(w http.ResponseWriter, r *http.Request) {
	captchaImage := gocaptcha.New(dx, dy, gocaptcha.RandLightColor())
	err := captchaImage.
		DrawBorder(gocaptcha.RandDeepColor()).
		DrawNoise(gocaptcha.NoiseDensityHigh, gocaptcha.NewTextNoiseDrawer(72)).
		DrawNoise(gocaptcha.NoiseDensityLower, gocaptcha.NewPointNoiseDrawer()).
		DrawLine(gocaptcha.NewBezier3DLine(), gocaptcha.RandDeepColor()).
		DrawText(gocaptcha.NewTwistTextDrawer(gocaptcha.DefaultDPI, gocaptcha.DefaultAmplitude, gocaptcha.DefaultFrequency), gocaptcha.RandText(4)).
		DrawLine(gocaptcha.NewBeeline(), gocaptcha.RandDeepColor()).
		//DrawLine(gocaptcha.NewHollowLine(), gocaptcha.RandLightColor()).
		DrawBlur(gocaptcha.NewGaussianBlur(), gocaptcha.DefaultBlurKernelSize, gocaptcha.DefaultBlurSigma).
		Error

	if err != nil {
		fmt.Println(err)
	}

	_ = captchaImage.Encode(w, gocaptcha.ImageFormatJpeg)
}

func init() {
	err := gocaptcha.SetFontPath("../fonts/")
	if err != nil {
		panic(err)
	}
}
