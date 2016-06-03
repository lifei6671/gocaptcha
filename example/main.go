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

	gocaptcha.SetFontFamily("fonts/BRACELET.ttf","fonts/ApothecaryFont.ttf");
	gocaptcha.SetFontFamily("fonts/actionj.ttf");
	gocaptcha.SetFontFamily("fonts/Comismsh.ttf");
	gocaptcha.SetFontFamily("fonts/Esquisito.ttf");
	gocaptcha.SetFontFamily("fonts/DENNEthree-dee.ttf");


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

	captchaImage,err := gocaptcha.NewCaptchaImage(dx,dy,gocaptcha.RandLightColor());

	captchaImage.Drawline(3);
	captchaImage.DrawBorder(gocaptcha.ColorToRGB(0x17A7A7A));
	captchaImage.DrawNoise(gocaptcha.CaptchaComplexHigh);

	captchaImage.DrawTextNoise(gocaptcha.CaptchaComplexLower);
	captchaImage.DrawText(gocaptcha.RandText(4));
	if err != nil {
		fmt.Println(err)
	}

	captchaImage.SaveImage(w,gocaptcha.ImageFormatJpeg);
}