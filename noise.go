package gocaptcha

import (
	"image"
	"image/draw"
	"math/rand"
	"time"

	"github.com/golang/freetype"
	"golang.org/x/image/font"
)

// NoiseDensity is the complexity of captcha
type NoiseDensity int

const (
	NoiseDensityLower NoiseDensity = iota
	NoiseDensityMedium
	NoiseDensityHigh
)

// NoiseDrawer is a type that can make noise on an image
type NoiseDrawer interface {
	// DrawNoise draws noise on the image
	DrawNoise(img draw.Image, density NoiseDensity) error
}

type pointNoiseDrawer struct {
	r *rand.Rand
}

// DrawNoise draws noise on the image
func (n pointNoiseDrawer) DrawNoise(img draw.Image, density NoiseDensity) error {
	var densityNum int
	switch density {
	case NoiseDensityLower:
		densityNum = 28
	case NoiseDensityMedium:
		densityNum = 18
	case NoiseDensityHigh:
		densityNum = 8
	default:
		densityNum = 18
	}

	bounds := img.Bounds()
	maxSize := (bounds.Dy() * bounds.Dx()) / densityNum

	width := bounds.Dx()
	height := bounds.Dy()

	for i := 0; i < maxSize; i++ {
		rw := n.r.Intn(width)
		rh := n.r.Intn(height)

		img.Set(rw, rh, RandColor())
		// 优化噪声点的生成逻辑，例如可以基于一定的概率决定是否绘制额外的点
		// 边界检查确保不越界
		if n.r.Intn(3) == 0 && rw+1 < width && rh+1 < height {
			img.Set(rw+1, rh+1, RandColor())
		}
	}
	return nil
}

// NewPointNoiseDrawer returns a NoiseDrawer that draws noise points
func NewPointNoiseDrawer() NoiseDrawer {
	return &pointNoiseDrawer{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// textNoiseDrawer draws noise text
type textNoiseDrawer struct {
	r   *rand.Rand
	dpi float64
}

// DrawNoise draws noise on the image
// Performance optimized: caches font to avoid repeated loading
func (n textNoiseDrawer) DrawNoise(img draw.Image, density NoiseDensity) error {
	var densityNum int
	switch density {
	case NoiseDensityLower:
		densityNum = 2000
	case NoiseDensityMedium:
		densityNum = 1500
	case NoiseDensityHigh:
		densityNum = 1000
	default:
		densityNum = 1500 // 默认值
	}
	bounds := img.Bounds()
	maxSize := (bounds.Dy() * bounds.Dx()) / densityNum
	c := freetype.NewContext()

	// 使用有效的 DPI 值，避免修改接收者字段 (race condition)
	dpi := n.dpi
	if dpi <= 0 {
		dpi = 72
	}

	c.SetDPI(dpi)

	c.SetClip(bounds)
	c.SetDst(img)
	c.SetHinting(font.HintingFull)
	rawFontSize := float64(bounds.Dy()) / (1 + float64(n.r.Intn(7))/float64(10))

	// 预加载字体一次，避免在循环中重复加载（性能优化）
	f, err := DefaultFontFamily.Random()
	if err != nil {
		return err
	}
	c.SetFont(f)

	for i := 0; i < maxSize; i++ {

		rw := n.r.Intn(bounds.Dx())
		rh := n.r.Intn(bounds.Dy())

		text := RandText(1)
		fontSize := rawFontSize/2 + float64(n.r.Intn(5))

		c.SetSrc(image.NewUniform(RandLightColor()))
		c.SetFontSize(fontSize)
		pt := freetype.Pt(rw, rh)

		_, err = c.DrawString(text, pt)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewTextNoiseDrawer(dpi float64) NoiseDrawer {
	return &textNoiseDrawer{
		r:   rand.New(rand.NewSource(time.Now().UnixNano())),
		dpi: dpi,
	}
}
