package gocaptcha

import (
	"errors"
	"image"
	"image/draw"
	"math"
	"math/rand"
	"time"

	"github.com/golang/freetype"
	"golang.org/x/image/font"
)

var (
	ErrNilCanvas = errors.New("canvas is nil")
	ErrNilText   = errors.New("text is nil")
)

// TextDrawer is a text drawer interface.
type TextDrawer interface {
	DrawString(canvas draw.Image, text string) error
}

type textDrawer struct {
	dpi float64
	r   *rand.Rand
}

// DrawString draws a string on the canvas.
func (t *textDrawer) DrawString(canvas draw.Image, text string) error {
	if len(text) == 0 {
		return ErrNilText
	}
	if canvas == nil {
		return ErrNilCanvas
	}
	c := freetype.NewContext()
	if t.dpi <= 0 {
		t.dpi = 72
	}
	c.SetDPI(t.dpi)
	c.SetClip(canvas.Bounds())
	c.SetDst(canvas)
	c.SetHinting(font.HintingFull)

	fontWidth := canvas.Bounds().Dx() / len(text)

	for i, s := range text {

		fontSize := float64(canvas.Bounds().Dy()) / (1 + float64(t.r.Intn(7))/float64(9))

		c.SetSrc(image.NewUniform(RandDeepColor()))
		c.SetFontSize(fontSize)
		f, err := DefaultFontFamily.Random()

		if err != nil {
			return err
		}
		c.SetFont(f)

		x := (fontWidth)*i + (fontWidth)/int(fontSize)

		y := 5 + t.r.Intn(canvas.Bounds().Dy()/2) + int(fontSize/2)

		pt := freetype.Pt(x, y)

		_, err = c.DrawString(string(s), pt)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewTextDrawer returns a new text drawer.
func NewTextDrawer(dpi float64) TextDrawer {
	return &textDrawer{
		dpi: dpi,
		r:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

type twistTextDrawer struct {
	dpi       float64
	r         *rand.Rand
	amplitude float64
	frequency float64
}

// DrawString draws a string on the canvas.
func (t *twistTextDrawer) DrawString(canvas draw.Image, text string) error {
	if len(text) == 0 {
		return ErrNilText
	}
	if canvas == nil {
		return ErrNilCanvas
	}
	// 创建一个新的画布用于存储扭曲后的图像
	textCanvas := image.NewRGBA(image.Rect(0, 0, canvas.Bounds().Dx(), canvas.Bounds().Dy()))
	draw.Draw(textCanvas, textCanvas.Bounds(), image.Transparent, image.Point{}, draw.Src)

	c := freetype.NewContext()
	if t.dpi <= 0 {
		t.dpi = 72
	}
	c.SetDPI(t.dpi)
	c.SetClip(canvas.Bounds())
	c.SetDst(textCanvas)
	c.SetHinting(font.HintingFull)

	fontWidth := canvas.Bounds().Dx() / len(text)

	for i, s := range text {

		fontSize := float64(canvas.Bounds().Dy()) / (1 + float64(t.r.Intn(7))/float64(9))

		c.SetSrc(image.NewUniform(RandDeepColor()))
		c.SetFontSize(fontSize)
		f, err := DefaultFontFamily.Random()

		if err != nil {
			return err
		}
		c.SetFont(f)

		x := (fontWidth)*i + (fontWidth)/int(fontSize)

		y := 5 + t.r.Intn(canvas.Bounds().Dy()/2) + int(fontSize/2)

		pt := freetype.Pt(x, y)

		_, err = c.DrawString(string(s), pt)
		if err != nil {
			return err
		}
	}
	return t.twistEffect(textCanvas, canvas)
}

func (t *twistTextDrawer) twistEffect(src image.Image, dst draw.Image) error {
	width := src.Bounds().Dx()
	height := src.Bounds().Dy()

	// 遍历源图像像素
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// 计算扭曲后的坐标
			dx := int(t.amplitude * math.Sin(t.frequency*float64(y)))
			newX := x + dx
			newY := y

			// 如果新坐标在目标图像范围内，设置像素
			if newX >= 0 && newX < width && newY >= 0 && newY < height {
				_, _, _, a := src.At(x, y).RGBA()
				if a != 0 {
					dst.Set(newX, newY, src.At(x, y))
				}
			}
		}
	}
	return nil
}

// NewTwistTextDrawer returns a new text drawer with twist effect.
func NewTwistTextDrawer(dpi float64, amplitude float64, frequency float64) TextDrawer {
	return &twistTextDrawer{
		dpi:       dpi,
		r:         rand.New(rand.NewSource(time.Now().UnixNano())),
		amplitude: amplitude,
		frequency: frequency,
	}
}
