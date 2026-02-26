package gocaptcha

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
	"sync"

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

const (
	defaultSecondaryPointChance = 1.0 / 3.0
	defaultNoiseTextLength      = 1
	defaultNoiseFontJitter      = 5
)

// NoiseColorFunc customizes noise color generation.
type NoiseColorFunc func(r *rand.Rand) color.Color

func randDeepNoiseColor(r *rand.Rand) color.Color {
	return randDeepColorFrom(r)
}

// NoiseConfig controls noise generation behavior.
type NoiseConfig struct {
	Density              NoiseDensity
	PointDensityDivisor  int
	TextDensityDivisor   int
	SecondaryPointChance float64
	TextLength           int
	FontSizeJitter       int
	PointColor           NoiseColorFunc
	TextColor            NoiseColorFunc
}

// NoiseDrawer is a type that can make noise on an image
type NoiseDrawer interface {
	// DrawNoise draws noise on the image
	DrawNoise(img draw.Image, density NoiseDensity) error
}

// ConfigurableNoiseDrawer extends NoiseDrawer with configuration support.
type ConfigurableNoiseDrawer interface {
	NoiseDrawer
	DrawNoiseWithConfig(img draw.Image, config NoiseConfig) error
}

type pointNoiseDrawer struct {
	r        *rand.Rand
	randMu   sync.Mutex
	randOnce sync.Once
}

// DrawNoise draws noise on the image.
func (n *pointNoiseDrawer) DrawNoise(img draw.Image, density NoiseDensity) error {
	return n.DrawNoiseWithConfig(img, NoiseConfig{Density: density})
}

// DrawNoiseWithConfig draws configurable point noise on the image.
func (n *pointNoiseDrawer) DrawNoiseWithConfig(img draw.Image, config NoiseConfig) error {
	if img == nil {
		return ErrNilCanvas
	}
	config = normalizeNoiseConfig(config)

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil
	}

	divisor := config.PointDensityDivisor
	if divisor <= 0 {
		divisor = pointDensityDivisorFor(config.Density)
	}
	total := (width * height) / maxInt(1, divisor)
	if total <= 0 {
		return nil
	}

	secondaryChance := config.SecondaryPointChance
	pointColor := config.PointColor

	n.withRand(func(r *rand.Rand) {
		for i := 0; i < total; i++ {
			x := randIntn(r, width)
			y := randIntn(r, height)

			img.Set(bounds.Min.X+x, bounds.Min.Y+y, noiseColorFrom(r, pointColor, randColorFromRand))
			if secondaryChance > 0 && r.Float64() < secondaryChance && x+1 < width && y+1 < height {
				img.Set(bounds.Min.X+x+1, bounds.Min.Y+y+1, noiseColorFrom(r, pointColor, randColorFromRand))
			}
		}
	})
	return nil
}

func (n *pointNoiseDrawer) withRand(fn func(r *rand.Rand)) {
	withDrawerRand(&n.r, &n.randOnce, &n.randMu, fn)
}

// NewPointNoiseDrawer returns a NoiseDrawer that draws noise points.
func NewPointNoiseDrawer() NoiseDrawer {
	return &pointNoiseDrawer{
		r: newSecureSeededRand(),
	}
}

// textNoiseDrawer draws noise text.
type textNoiseDrawer struct {
	r          *rand.Rand
	randMu     sync.Mutex
	randOnce   sync.Once
	dpi        float64
	fontFamily *FontFamily
}

// DrawNoise draws noise on the image.
func (n *textNoiseDrawer) DrawNoise(img draw.Image, density NoiseDensity) error {
	return n.DrawNoiseWithConfig(img, NoiseConfig{Density: density})
}

// DrawNoiseWithConfig draws configurable text noise on the image.
func (n *textNoiseDrawer) DrawNoiseWithConfig(img draw.Image, config NoiseConfig) error {
	if img == nil {
		return ErrNilCanvas
	}
	config = normalizeNoiseConfig(config)

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil
	}

	divisor := config.TextDensityDivisor
	if divisor <= 0 {
		divisor = textDensityDivisorFor(config.Density)
	}
	total := (width * height) / maxInt(1, divisor)
	if total <= 0 {
		return nil
	}

	c := freetype.NewContext()
	c.SetDPI(normalizeDPI(n.dpi))
	c.SetClip(bounds)
	c.SetDst(img)
	c.SetHinting(font.HintingFull)

	textColor := config.TextColor
	textLength := config.TextLength
	fontJitter := config.FontSizeJitter
	fontFamily := n.fontFamily
	if fontFamily == nil {
		fontFamily = DefaultFontFamily
	}
	fonts, err := fontFamily.WeightedCachedFonts()
	if err != nil {
		return err
	}

	var drawErr error
	n.withRand(func(r *rand.Rand) {
		rawFontSize := float64(height) / (1 + float64(randIntn(r, 7))/10.0)
		if rawFontSize < 1 {
			rawFontSize = 1
		}

		for i := 0; i < total; i++ {
			if drawErr != nil {
				return
			}

			c.SetFont(fonts[randIntn(r, len(fonts))])

			x := bounds.Min.X + randIntn(r, width)
			y := bounds.Min.Y + randIntn(r, height)
			fontSize := rawFontSize/2 + float64(randIntn(r, fontJitter))
			if fontSize < 1 {
				fontSize = 1
			}

			c.SetSrc(image.NewUniform(noiseColorFrom(r, textColor, randLightColorFromRand)))
			c.SetFontSize(fontSize)
			if _, err := c.DrawString(randTextFromRand(r, textLength), freetype.Pt(x, y)); err != nil {
				drawErr = err
				return
			}
		}
	})

	return drawErr
}

func (n *textNoiseDrawer) withRand(fn func(r *rand.Rand)) {
	withDrawerRand(&n.r, &n.randOnce, &n.randMu, fn)
}

func randTextFromRand(r *rand.Rand, num int) string {
	if num <= 0 {
		num = defaultNoiseTextLength
	}
	if len(TextCharacters) == 0 {
		return ""
	}
	text := make([]rune, num)
	for i := 0; i < num; i++ {
		text[i] = TextCharacters[randIntn(r, len(TextCharacters))]
	}
	return string(text)
}

func normalizeNoiseConfig(config NoiseConfig) NoiseConfig {
	if config.Density < NoiseDensityLower || config.Density > NoiseDensityHigh {
		config.Density = NoiseDensityMedium
	}
	if config.SecondaryPointChance < 0 || math.IsNaN(config.SecondaryPointChance) || math.IsInf(config.SecondaryPointChance, 0) {
		config.SecondaryPointChance = defaultSecondaryPointChance
	}
	if config.SecondaryPointChance > 1 {
		config.SecondaryPointChance = 1
	}
	if config.TextLength <= 0 {
		config.TextLength = defaultNoiseTextLength
	}
	if config.FontSizeJitter <= 0 {
		config.FontSizeJitter = defaultNoiseFontJitter
	}
	return config
}

func pointDensityDivisorFor(density NoiseDensity) int {
	switch density {
	case NoiseDensityLower:
		return 28
	case NoiseDensityMedium:
		return 18
	case NoiseDensityHigh:
		return 8
	default:
		return 18
	}
}

func textDensityDivisorFor(density NoiseDensity) int {
	switch density {
	case NoiseDensityLower:
		return 2000
	case NoiseDensityMedium:
		return 1500
	case NoiseDensityHigh:
		return 1000
	default:
		return 1500
	}
}

func noiseColorFrom(r *rand.Rand, fn NoiseColorFunc, fallback func(*rand.Rand) color.Color) color.Color {
	if fn != nil {
		return fn(r)
	}
	return fallback(r)
}

func randLightColorFromRand(r *rand.Rand) color.Color {
	return color.RGBA{
		R: uint8(randIntn(r, 128) + 128),
		G: uint8(randIntn(r, 128) + 128),
		B: uint8(randIntn(r, 128) + 128),
		A: 255,
	}
}

func randColorFromRand(r *rand.Rand) color.Color {
	red := randIntn(r, 255)
	green := randIntn(r, 255)
	blue := 0
	sum := red + green
	if sum <= 400 {
		blue = min(255, 400-sum)
	}
	return color.RGBA{R: uint8(red), G: uint8(green), B: uint8(blue), A: 255}
}

func NewTextNoiseDrawer(dpi float64) NoiseDrawer {
	return &textNoiseDrawer{
		r:          newSecureSeededRand(),
		dpi:        dpi,
		fontFamily: DefaultFontFamily,
	}
}
