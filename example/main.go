package main

import (
	"fmt"
	"github.com/lifei6671/gocaptcha"
	"html/template"
	"log"
	"net/http"
)

const (
	dx = 150
	dy = 50
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

	err := captchaImage.DrawNoise(gocaptcha.CaptchaComplexLower).
		DrawTextNoise(gocaptcha.CaptchaComplexLower).
		DrawText(gocaptcha.RandText(4)).
		DrawBorder(gocaptcha.ColorToRGB(0x17A7A7A)).
		DrawSineLine().Error

	if err != nil {
		fmt.Println(err)
	}

	_ = captchaImage.SaveImage(w, gocaptcha.ImageFormatJpeg)
}
