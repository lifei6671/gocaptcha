package main

import (
	"fmt"
	"gocaptcha"
	"net/http"
	"log"
	"html/template"
)
const (
	dx = 200
	dy = 80
)

func main() {

	captcha.SetFontFamily("fonts/BRACELET.ttf","fonts/ApothecaryFont.ttf");
	captcha.SetFontFamily("fonts/actionj.ttf");
	captcha.SetFontFamily("fonts/Comismsh.ttf");
	captcha.SetFontFamily("fonts/Esquisito.ttf");
	captcha.SetFontFamily("fonts/DENNEthree-dee.ttf");


	http.HandleFunc("/", Index)
	http.HandleFunc("/get/", Get)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("tpl/index.html")
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, nil)
}
func Get(w http.ResponseWriter, r *http.Request) {

	captchaImage,err := captcha.NewCaptchaImage(dx,dy,captcha.RandLightColor());

	captchaImage.Drawline(3);
	captchaImage.DrawBorder(captcha.ColorToRGB(0x17A7A7A));
	captchaImage.DrawNoise(captcha.CaptchaComplexHigh);

	captchaImage.DrawTextNoise(captcha.CaptchaComplexLower);
	captchaImage.DrawText(captcha.RandText(4));
	if err != nil {
		fmt.Println(err)
	}

	captchaImage.SaveImage(w,captcha.ImageFormatJpeg);
}