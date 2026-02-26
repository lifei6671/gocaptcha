package gocaptcha

import (
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/golang/freetype"
	"golang.org/x/image/font"
)

var (
	ErrNilCanvas         = errors.New("canvas is nil")
	ErrNilText           = errors.New("text is nil")
	ErrInvalidCanvasSize = errors.New("invalid canvas size")
	ErrInvalidEffectSize = errors.New("effect source and destination sizes must match")
	ErrNilEffectCanvas   = errors.New("effect canvas is nil")
)

const (
	defaultTextDPI        = 72.0
	defaultTwistAmplitude = 2.0
	defaultTwistFrequency = 0.05
)

// TextDrawer is a text drawer interface.
type TextDrawer interface {
	DrawString(canvas draw.Image, text string) error
}

// TextEffect applies a visual effect on text pixels.
type TextEffect interface {
	Apply(src, dst *image.RGBA) error
}

// WaveDistortionMode controls how wave distortion is applied.
type WaveDistortionMode int

const (
	WaveDistortionHorizontal WaveDistortionMode = iota
	WaveDistortionVertical
	WaveDistortionDual
)

type textDrawParams struct {
	fontSizes   []float64
	xPositions  []int
	yPositions  []int
	fontIndexes []int
	colors      []color.RGBA
}

type textDrawer struct {
	dpi      float64
	r        *rand.Rand
	randMu   sync.Mutex
	randOnce sync.Once
}

// DrawString draws a string on the canvas.
func (t *textDrawer) DrawString(canvas draw.Image, text string) error {
	runes, bounds, err := validateTextDrawInput(canvas, text)
	if err != nil {
		return err
	}
	return drawRandomizedText(canvas, bounds, runes, t.dpi, t.withRand)
}

func (t *textDrawer) withRand(fn func(r *rand.Rand)) {
	withDrawerRand(&t.r, &t.randOnce, &t.randMu, fn)
}

// NewTextDrawer returns a new text drawer.
func NewTextDrawer(dpi float64) TextDrawer {
	return &textDrawer{
		dpi: normalizeDPI(dpi),
		r:   newSecureSeededRand(),
	}
}

type twistTextDrawer struct {
	dpi      float64
	r        *rand.Rand
	randMu   sync.Mutex
	randOnce sync.Once

	amplitude float64
	frequency float64

	effects []TextEffect
	modes   []WaveDistortionMode
}

// DrawString draws a string on the canvas.
func (t *twistTextDrawer) DrawString(canvas draw.Image, text string) error {
	runes, bounds, err := validateTextDrawInput(canvas, text)
	if err != nil {
		return err
	}

	textCanvas := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	if err := drawRandomizedText(textCanvas, textCanvas.Bounds(), runes, t.dpi, t.withRand); err != nil {
		return err
	}

	return applyTextEffects(textCanvas, canvas, t.resolveEffects())
}

func (t *twistTextDrawer) withRand(fn func(r *rand.Rand)) {
	withDrawerRand(&t.r, &t.randOnce, &t.randMu, fn)
}

func (t *twistTextDrawer) resolveEffects() []TextEffect {
	if len(t.effects) > 0 {
		out := make([]TextEffect, len(t.effects))
		copy(out, t.effects)
		return out
	}

	modes := normalizeDistortionModes(t.modes)
	mode := modes[0]
	t.withRand(func(r *rand.Rand) {
		mode = modes[randIntn(r, len(modes))]
	})

	return []TextEffect{
		NewWaveTextEffect(t.amplitude, t.frequency, mode),
	}
}

// NewTwistTextDrawer returns a new text drawer with wave effect.
// Parameters: dpi for font resolution, amplitude for wave height, frequency for wave frequency.
func NewTwistTextDrawer(dpi float64, amplitude float64, frequency float64) TextDrawer {
	return NewTwistTextDrawerWithModes(
		dpi,
		amplitude,
		frequency,
		WaveDistortionHorizontal,
		WaveDistortionVertical,
		WaveDistortionDual,
	)
}

// NewTwistTextDrawerWithModes returns a text drawer with selectable wave modes.
func NewTwistTextDrawerWithModes(dpi float64, amplitude float64, frequency float64, modes ...WaveDistortionMode) TextDrawer {
	return &twistTextDrawer{
		dpi:       normalizeDPI(dpi),
		r:         newSecureSeededRand(),
		amplitude: normalizeAmplitude(amplitude),
		frequency: normalizeFrequency(frequency),
		modes:     normalizeDistortionModes(modes),
	}
}

// NewEffectTextDrawer returns a text drawer that applies custom text effects.
func NewEffectTextDrawer(dpi float64, effects ...TextEffect) TextDrawer {
	return &twistTextDrawer{
		dpi:     normalizeDPI(dpi),
		r:       newSecureSeededRand(),
		effects: append([]TextEffect(nil), effects...),
	}
}

type waveTextEffect struct {
	amplitude float64
	frequency float64
	mode      WaveDistortionMode
}

// NewWaveTextEffect builds a sine-wave distortion effect.
func NewWaveTextEffect(amplitude float64, frequency float64, mode WaveDistortionMode) TextEffect {
	return &waveTextEffect{
		amplitude: normalizeAmplitude(amplitude),
		frequency: normalizeFrequency(frequency),
		mode:      normalizeDistortionMode(mode),
	}
}

// Apply applies the wave effect from src to dst.
func (w *waveTextEffect) Apply(src, dst *image.RGBA) error {
	if err := validateEffectBuffers(src, dst); err != nil {
		return err
	}
	clear(dst.Pix)

	amplitude := normalizeAmplitude(w.amplitude)
	frequency := normalizeFrequency(w.frequency)
	mode := normalizeDistortionMode(w.mode)

	switch mode {
	case WaveDistortionHorizontal:
		applyHorizontalWave(src, dst, precomputeWaveShifts(src.Bounds().Dy(), amplitude, frequency))
	case WaveDistortionVertical:
		applyVerticalWave(src, dst, precomputeWaveShifts(src.Bounds().Dx(), amplitude, frequency))
	case WaveDistortionDual:
		tmp := image.NewRGBA(src.Bounds())
		clear(tmp.Pix)
		applyHorizontalWave(src, tmp, precomputeWaveShifts(src.Bounds().Dy(), amplitude, frequency))
		applyVerticalWave(tmp, dst, precomputeWaveShifts(src.Bounds().Dx(), amplitude, frequency))
	}
	return nil
}

func drawRandomizedText(
	dst draw.Image,
	bounds image.Rectangle,
	runes []rune,
	dpi float64,
	withRand func(func(*rand.Rand)),
) error {
	fonts, err := DefaultFontFamily.WeightedCachedFonts()
	if err != nil {
		return err
	}

	c := freetype.NewContext()
	c.SetDPI(normalizeDPI(dpi))
	c.SetClip(bounds)
	c.SetDst(dst)
	c.SetHinting(font.HintingFull)

	params := precomputeTextDrawParams(len(runes), bounds, len(fonts), withRand)
	for i, s := range runes {
		c.SetSrc(image.NewUniform(params.colors[i]))
		c.SetFontSize(params.fontSizes[i])
		c.SetFont(fonts[params.fontIndexes[i]])

		if _, err := c.DrawString(string(s), freetype.Pt(params.xPositions[i], params.yPositions[i])); err != nil {
			return err
		}
	}
	return nil
}

func precomputeTextDrawParams(
	runeCount int,
	bounds image.Rectangle,
	fontCount int,
	withRand func(func(*rand.Rand)),
) textDrawParams {
	params := textDrawParams{
		fontSizes:   make([]float64, runeCount),
		xPositions:  make([]int, runeCount),
		yPositions:  make([]int, runeCount),
		fontIndexes: make([]int, runeCount),
		colors:      make([]color.RGBA, runeCount),
	}

	width := bounds.Dx()
	height := bounds.Dy()
	fontWidth := maxInt(1, width/maxInt(1, runeCount))
	yJitterRange := maxInt(1, height/2)

	withRand(func(r *rand.Rand) {
		for i := 0; i < runeCount; i++ {
			fontSize := float64(height) / (1 + float64(randIntn(r, 7))/9.0)
			if fontSize < 1 {
				fontSize = 1
			}
			fontSizeInt := maxInt(1, int(fontSize))

			params.fontSizes[i] = fontSize
			params.xPositions[i] = fontWidth*i + fontWidth/fontSizeInt
			params.yPositions[i] = 5 + randIntn(r, yJitterRange) + int(fontSize/2)
			params.fontIndexes[i] = randIntn(r, fontCount)
			params.colors[i] = randDeepColorFrom(r)
		}
	})

	return params
}

func applyTextEffects(src *image.RGBA, dst draw.Image, effects []TextEffect) error {
	if len(effects) == 0 {
		draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Over)
		return nil
	}

	bounds := src.Bounds()
	bufA := image.NewRGBA(bounds)
	bufB := image.NewRGBA(bounds)
	current := src
	output := bufA

	for _, effect := range effects {
		clear(output.Pix)
		if err := effect.Apply(current, output); err != nil {
			return err
		}
		current = output
		if output == bufA {
			output = bufB
		} else {
			output = bufA
		}
	}

	draw.Draw(dst, dst.Bounds(), current, current.Bounds().Min, draw.Over)
	return nil
}

func applyHorizontalWave(src, dst *image.RGBA, shifts []int) {
	width := src.Bounds().Dx()
	srcBounds := src.Bounds()
	dstBounds := dst.Bounds()

	for y, dx := range shifts {
		if dx <= -width || dx >= width {
			continue
		}

		srcStartX := 0
		dstStartX := dx
		if dx < 0 {
			srcStartX = -dx
			dstStartX = 0
		}

		rowLength := width - absInt(dx)
		if rowLength <= 0 {
			continue
		}

		srcOffset := src.PixOffset(srcBounds.Min.X+srcStartX, srcBounds.Min.Y+y)
		dstOffset := dst.PixOffset(dstBounds.Min.X+dstStartX, dstBounds.Min.Y+y)
		copy(dst.Pix[dstOffset:dstOffset+rowLength*4], src.Pix[srcOffset:srcOffset+rowLength*4])
	}
}

func applyVerticalWave(src, dst *image.RGBA, shifts []int) {
	height := src.Bounds().Dy()
	srcBounds := src.Bounds()
	dstBounds := dst.Bounds()

	for x, dy := range shifts {
		if dy <= -height || dy >= height {
			continue
		}

		srcStartY := 0
		dstStartY := dy
		if dy < 0 {
			srcStartY = -dy
			dstStartY = 0
		}

		colLength := height - absInt(dy)
		if colLength <= 0 {
			continue
		}

		for i := 0; i < colLength; i++ {
			srcOffset := src.PixOffset(srcBounds.Min.X+x, srcBounds.Min.Y+srcStartY+i)
			dstOffset := dst.PixOffset(dstBounds.Min.X+x, dstBounds.Min.Y+dstStartY+i)
			copy(dst.Pix[dstOffset:dstOffset+4], src.Pix[srcOffset:srcOffset+4])
		}
	}
}

func precomputeWaveShifts(length int, amplitude float64, frequency float64) []int {
	shifts := make([]int, length)
	for i := 0; i < length; i++ {
		shifts[i] = int(amplitude * math.Sin(frequency*float64(i)))
	}
	return shifts
}

func validateTextDrawInput(canvas draw.Image, text string) ([]rune, image.Rectangle, error) {
	if canvas == nil {
		return nil, image.Rectangle{}, ErrNilCanvas
	}
	runes := []rune(text)
	if len(runes) == 0 {
		return nil, image.Rectangle{}, ErrNilText
	}

	bounds := canvas.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return nil, image.Rectangle{}, ErrInvalidCanvasSize
	}
	return runes, bounds, nil
}

func validateEffectBuffers(src, dst *image.RGBA) error {
	if src == nil || dst == nil {
		return ErrNilEffectCanvas
	}
	if src.Bounds().Dx() != dst.Bounds().Dx() || src.Bounds().Dy() != dst.Bounds().Dy() {
		return ErrInvalidEffectSize
	}
	return nil
}

func normalizeDPI(dpi float64) float64 {
	if dpi <= 0 || math.IsNaN(dpi) {
		return defaultTextDPI
	}
	return dpi
}

func normalizeAmplitude(amplitude float64) float64 {
	if amplitude <= 0 || math.IsNaN(amplitude) {
		return defaultTwistAmplitude
	}
	return amplitude
}

func normalizeFrequency(frequency float64) float64 {
	if frequency <= 0 || math.IsNaN(frequency) {
		return defaultTwistFrequency
	}
	return frequency
}

func normalizeDistortionModes(modes []WaveDistortionMode) []WaveDistortionMode {
	if len(modes) == 0 {
		return []WaveDistortionMode{WaveDistortionHorizontal}
	}
	out := make([]WaveDistortionMode, 0, len(modes))
	for _, mode := range modes {
		switch mode {
		case WaveDistortionHorizontal, WaveDistortionVertical, WaveDistortionDual:
			out = append(out, mode)
		}
	}
	if len(out) == 0 {
		return []WaveDistortionMode{WaveDistortionHorizontal}
	}
	return out
}

func normalizeDistortionMode(mode WaveDistortionMode) WaveDistortionMode {
	switch mode {
	case WaveDistortionHorizontal, WaveDistortionVertical, WaveDistortionDual:
		return mode
	default:
		return WaveDistortionHorizontal
	}
}

func randDeepColorFrom(r *rand.Rand) color.RGBA {
	const (
		maxValue = 150
		minValue = 50
	)
	return color.RGBA{
		R: uint8(randIntn(r, maxValue-minValue+1) + minValue),
		G: uint8(randIntn(r, maxValue-minValue+1) + minValue),
		B: uint8(randIntn(r, maxValue-minValue+1) + minValue),
		A: 255,
	}
}

func withDrawerRand(r **rand.Rand, once *sync.Once, mu *sync.Mutex, fn func(*rand.Rand)) {
	once.Do(func() {
		if *r == nil {
			*r = newSecureSeededRand()
		}
	})
	mu.Lock()
	defer mu.Unlock()
	fn(*r)
}

func newSecureSeededRand() *rand.Rand {
	var seedBytes [8]byte
	if _, err := crand.Read(seedBytes[:]); err != nil {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	seed := int64(binary.LittleEndian.Uint64(seedBytes[:]))
	return rand.New(rand.NewSource(seed))
}

func randIntn(r *rand.Rand, n int) int {
	if n <= 1 {
		return 0
	}
	return r.Intn(n)
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
