package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/lifei6671/gocaptcha"
)

const (
	dx = 180
	dy = 60
)

const defaultPreset = "classic"

type captchaPreset struct {
	ID          string
	Name        string
	Description string
}

type indexViewData struct {
	SelectedPreset string
	CurrentPreset  captchaPreset
	Presets        []captchaPreset
	NowUnixNano    int64
}

var presetList = []captchaPreset{
	{
		ID:          "classic",
		Name:        "经典",
		Description: "原始风格：文本噪声 + Poisson 点簇 + 中等扭曲与轻模糊。",
	},
	{
		ID:          "balanced",
		Name:        "均衡",
		Description: "强度均衡：点噪声 + Perlin 连贯噪声 + 中等波形扭曲。",
	},
	{
		ID:          "ocr-hard",
		Name:        "高抗 OCR",
		Description: "高强度：Poisson + Perlin 叠加、双曲线干扰、双层波形扭曲。",
	},
	{
		ID:          "light",
		Name:        "轻量",
		Description: "更易读：低密度噪声、轻扭曲与轻模糊。",
	},
}

var presetMap = func() map[string]captchaPreset {
	m := make(map[string]captchaPreset, len(presetList))
	for _, p := range presetList {
		m[p.ID] = p
	}
	return m
}()

var indexTemplate = mustParseIndexTemplate()

func main() {
	http.HandleFunc("/", Index)
	http.HandleFunc("/get", Get)
	http.HandleFunc("/get/", Get)
	fmt.Println("服务已启动 -> http://127.0.0.1:8800")
	err := http.ListenAndServe(":8800", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	selected := normalizePreset(r.URL.Query().Get("preset"))
	data := indexViewData{
		SelectedPreset: selected,
		CurrentPreset:  presetMap[selected],
		Presets:        presetList,
		NowUnixNano:    time.Now().UnixNano(),
	}
	if err := indexTemplate.Execute(w, data); err != nil {
		log.Printf("render index failed: %v", err)
		http.Error(w, "render index failed", http.StatusInternalServerError)
		return
	}
}

func Get(w http.ResponseWriter, r *http.Request) {
	selected := normalizePreset(r.URL.Query().Get("preset"))
	captchaImage := gocaptcha.New(dx, dy, gocaptcha.RandLightColor())
	err := applyPreset(captchaImage, selected).Error
	if err != nil {
		log.Printf("generate captcha failed, preset=%s, err=%v", selected, err)
		http.Error(w, "generate captcha failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")

	if err := captchaImage.Encode(w, gocaptcha.ImageFormatJpeg); err != nil {
		log.Printf("encode captcha failed, preset=%s, err=%v", selected, err)
		http.Error(w, "encode captcha failed", http.StatusInternalServerError)
	}
}

func applyPreset(captchaImage *gocaptcha.CaptchaImage, preset string) *gocaptcha.CaptchaImage {
	switch preset {
	case "balanced":
		return captchaImage.
			DrawBorder(gocaptcha.RandDeepColor()).
			DrawNoiseWithConfig(
				gocaptcha.NewPointNoiseDrawer(),
				gocaptcha.NoiseConfig{
					Density:              gocaptcha.NoiseDensityMedium,
					PointDensityDivisor:  14,
					SecondaryPointChance: 0.42,
				},
			).
			DrawNoise(
				gocaptcha.NoiseDensityLower,
				gocaptcha.NewPerlinNoiseDrawerWithConfig(24, 3, 0.76, nil),
			).
			DrawLine(gocaptcha.NewBezierLine(), gocaptcha.RandDeepColor()).
			DrawText(
				gocaptcha.NewTwistTextDrawerWithModes(
					gocaptcha.DefaultDPI,
					8,
					0.02,
					gocaptcha.WaveDistortionHorizontal,
					gocaptcha.WaveDistortionVertical,
				),
				gocaptcha.RandText(4),
			).
			DrawBlur(gocaptcha.NewGaussianBlur(), 3, 0.80)
	case "ocr-hard":
		return captchaImage.
			DrawBorder(gocaptcha.RandDeepColor()).
			DrawNoiseWithConfig(
				gocaptcha.NewTextNoiseDrawer(gocaptcha.DefaultDPI),
				gocaptcha.NoiseConfig{
					Density:            gocaptcha.NoiseDensityHigh,
					TextDensityDivisor: 16,
					TextLength:         2,
					FontSizeJitter:     8,
				},
			).
			DrawNoise(
				gocaptcha.NoiseDensityHigh,
				gocaptcha.NewPoissonPointNoiseDrawerWithConfig(4.2, 24, nil),
			).
			DrawNoise(
				gocaptcha.NoiseDensityHigh,
				gocaptcha.NewPerlinNoiseDrawerWithConfig(14, 4, 0.62, nil),
			).
			DrawLine(gocaptcha.NewBezier3DLine(), gocaptcha.RandDeepColor()).
			DrawLine(gocaptcha.NewCurveLine(), gocaptcha.RandDeepColor()).
			DrawText(
				gocaptcha.NewEffectTextDrawer(
					gocaptcha.DefaultDPI,
					gocaptcha.NewWaveTextEffect(12, 0.03, gocaptcha.WaveDistortionDual),
					gocaptcha.NewWaveTextEffect(5, 0.08, gocaptcha.WaveDistortionHorizontal),
				),
				gocaptcha.RandText(5),
			).
			DrawBlur(gocaptcha.NewGaussianBlur(), 5, 1.10)
	case "light":
		return captchaImage.
			DrawBorder(gocaptcha.RandDeepColor()).
			DrawNoiseWithConfig(
				gocaptcha.NewTextNoiseDrawer(gocaptcha.DefaultDPI),
				gocaptcha.NoiseConfig{
					Density:            gocaptcha.NoiseDensityLower,
					TextDensityDivisor: 45,
					TextLength:         1,
					FontSizeJitter:     3,
				},
			).
			DrawText(gocaptcha.NewTextDrawer(gocaptcha.DefaultDPI), gocaptcha.RandText(4)).
			DrawLine(gocaptcha.NewBeeline(), gocaptcha.RandDeepColor()).
			DrawBlur(gocaptcha.NewGaussianBlur(), 3, 0.55)
	default:
		return captchaImage.
			DrawBorder(gocaptcha.RandDeepColor()).
			DrawNoise(gocaptcha.NoiseDensityLower, gocaptcha.NewTextNoiseDrawer(gocaptcha.DefaultDPI)).
			DrawNoise(gocaptcha.NoiseDensityLower, gocaptcha.NewPoissonPointNoiseDrawer()).
			DrawLine(gocaptcha.NewBezier3DLine(), gocaptcha.RandDeepColor()).
			DrawText(gocaptcha.NewTwistTextDrawer(gocaptcha.DefaultDPI, 15, 0.01), gocaptcha.RandText(4)).
			DrawLine(gocaptcha.NewBeeline(), gocaptcha.RandDeepColor()).
			DrawBlur(gocaptcha.NewGaussianBlur(), gocaptcha.DefaultBlurKernelSize, 0.1)
	}
}

func normalizePreset(raw string) string {
	if _, ok := presetMap[raw]; ok {
		return raw
	}
	return defaultPreset
}

func mustParseIndexTemplate() *template.Template {
	candidates := []string{"tpl/index.html", "example/tpl/index.html"}
	var lastErr error
	for _, path := range candidates {
		t, err := template.ParseFiles(path)
		if err == nil {
			return t
		}
		lastErr = err
	}
	panic(fmt.Sprintf("parse index template failed: %v", lastErr))
}

func init() {
	paths := []string{"../fonts/", "fonts/"}
	var lastErr error
	for _, path := range paths {
		err := gocaptcha.SetFontPath(path)
		if err == nil {
			return
		}
		lastErr = err
	}
	panic(lastErr)
}
