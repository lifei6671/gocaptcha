package main

import (
	"fmt"
	"gocaptcha"
	"net/http"
	"log"
	"html/template"
	"os"
	"strings"
	"io/ioutil"
)
const (
	dx = 200
	dy = 80
)

func main() {

	//fontFils,err := ListDir("fonts",".ttf");
	//if(err != nil){
	//	fmt.Println(err);
	//	return ;
	//}
	//
	//gocaptcha.SetFontFamily(fontFils...);

	gocaptcha.SetFontFamily(
		"fonts/3Dumb.ttf",
		"fonts/DeborahFancyDress.ttf",
		"fonts/actionj.ttf",
		"fonts/chromohv.ttf",
		"fonts/D3Parallelism.ttf",
		"fonts/Flim-Flam.ttf",
		"fonts/KREMLINGEORGIANI3D.ttf",
		//"fonts/Lead.ttf",
		)


	http.HandleFunc("/", Index)
	http.HandleFunc("/get/", Get)
	fmt.Println("服务已启动...");
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

//获取指定目录下的所有文件，不进入下一级目录搜索，可以匹配后缀过滤。
func ListDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0, 10)
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}
	return files, nil
}