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

	maxSize := (img.Bounds().Dy() * img.Bounds().Dx()) / densityNum

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	for i := 0; i < maxSize; i++ {
		rw := n.r.Intn(width)
		rh := n.r.Intn(height)

		img.Set(rw, rh, RandColor())
		// 优化噪声点的生成逻辑，例如可以基于一定的概率决定是否绘制额外的点
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
	if n.dpi <= 0 {
		n.dpi = 72
	}

	c.SetDPI(n.dpi)

	c.SetClip(bounds)
	c.SetDst(img)
	c.SetHinting(font.HintingFull)
	rawFontSize := float64(bounds.Dy()) / (1 + float64(n.r.Intn(7))/float64(10))

	for i := 0; i < maxSize; i++ {

		rw := n.r.Intn(bounds.Dx())
		rh := n.r.Intn(bounds.Dy())

		text := RandText(1)
		fontSize := rawFontSize/2 + float64(n.r.Intn(5))

		c.SetSrc(image.NewUniform(RandLightColor()))
		c.SetFontSize(fontSize)
		f, err := DefaultFontFamily.Random()
		if err != nil {
			return err
		}
		c.SetFont(f)
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
