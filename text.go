// Package gocaptcha 提供生成验证码图片的功能
// 本文件实现了文本绘制相关的功能，用于在验证码中添加文本元素
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
	// ErrNilCanvas 画布为nil的错误
	ErrNilCanvas = errors.New("canvas is nil")
	// ErrNilText 文本为空的错误
	ErrNilText = errors.New("text is nil")
	// ErrInvalidCanvasSize 画布尺寸无效的错误
	ErrInvalidCanvasSize = errors.New("invalid canvas size")
	// ErrInvalidEffectSize 效果源和目标尺寸不匹配的错误
	ErrInvalidEffectSize = errors.New("effect source and destination sizes must match")
	// ErrNilEffectCanvas 效果画布为nil的错误
	ErrNilEffectCanvas = errors.New("effect canvas is nil")
)

const (
	// defaultTextDPI 默认的文本DPI
	// 影响：控制文本的清晰度，值越高文本越清晰
	defaultTextDPI = 72.0
	// defaultTwistAmplitude 默认的扭曲振幅
	// 影响：控制文本扭曲的程度，值越大扭曲效果越明显
	defaultTwistAmplitude = 2.0
	// defaultTwistFrequency 默认的扭曲频率
	// 影响：控制扭曲波纹的密度，值越大波纹越密集
	defaultTwistFrequency = 0.05
)

// TextDrawer 文本绘制器接口
// 定义了在画布上绘制文本的方法
// 方法：
// - DrawString: 在画布上绘制字符串
//   参数：
//   - canvas: 要绘制的画布
//   - text: 要绘制的文本
//   返回值：
//   - error: 绘制过程中的错误

type TextDrawer interface {
	DrawString(canvas draw.Image, text string) error
}

// TextEffect 文本效果接口
// 定义了对文本像素应用视觉效果的方法
// 方法：
// - Apply: 从源图像应用效果到目标图像
//   参数：
//   - src: 源图像
//   - dst: 目标图像
//   返回值：
//   - error: 应用过程中的错误

type TextEffect interface {
	Apply(src, dst *image.RGBA) error
}

// WaveDistortionMode 波浪扭曲模式
// 控制如何应用波浪扭曲效果

type WaveDistortionMode int

const (
	// WaveDistortionHorizontal 水平波浪扭曲
	// 文本在水平方向上产生波浪效果
	WaveDistortionHorizontal WaveDistortionMode = iota
	// WaveDistortionVertical 垂直波浪扭曲
	// 文本在垂直方向上产生波浪效果
	WaveDistortionVertical
	// WaveDistortionDual 双波浪扭曲
	// 文本在水平和垂直方向上都产生波浪效果
	WaveDistortionDual
)

// textDrawParams 文本绘制参数
// 存储文本绘制的各种参数
// 字段：
// - fontSizes: 字体大小列表
// - xPositions: X坐标位置列表
// - yPositions: Y坐标位置列表
// - fontIndexes: 字体索引列表
// - colors: 颜色列表

type textDrawParams struct {
	fontSizes   []float64
	xPositions  []int
	yPositions  []int
	fontIndexes []int
	colors      []color.RGBA
}

// textDrawer 文本绘制器
// 实现了 TextDrawer 接口，用于绘制普通文本
// 字段：
// - dpi: 每英寸点数，影响文本清晰度
// - r: 随机数生成器
// - randMu: 随机数生成器的互斥锁
// - randOnce: 确保随机数生成器只初始化一次

type textDrawer struct {
	dpi      float64
	r        *rand.Rand
	randMu   sync.Mutex
	randOnce sync.Once
}

// DrawString 在画布上绘制字符串
// 参数：
// - canvas: 要绘制的画布
// - text: 要绘制的文本
// 返回值：
// - error: 绘制过程中的错误
// 影响：在画布上绘制随机化的文本，每个字符使用不同的字体、大小和位置
func (t *textDrawer) DrawString(canvas draw.Image, text string) error {
	runes, bounds, err := validateTextDrawInput(canvas, text)
	if err != nil {
		return err
	}
	return drawRandomizedText(canvas, bounds, runes, t.dpi, t.withRand)
}

// withRand 使用随机数生成器执行函数
// 参数：
// - fn: 要执行的函数
// 影响：确保在安全的情况下使用随机数生成器
func (t *textDrawer) withRand(fn func(r *rand.Rand)) {
	withDrawerRand(&t.r, &t.randOnce, &t.randMu, fn)
}

// NewTextDrawer 创建一个新的文本绘制器
// 参数：
// - dpi: 每英寸点数，影响文本清晰度
// 返回值：
// - TextDrawer: 文本绘制器实例
// 示例：
// ```go
// // 创建一个文本绘制器
// textDrawer := gocaptcha.NewTextDrawer(72.0)
// // 为验证码添加文本
// captcha.DrawText(textDrawer, "1234")
// ```
func NewTextDrawer(dpi float64) TextDrawer {
	return &textDrawer{
		dpi: normalizeDPI(dpi),
		r:   newSecureSeededRand(),
	}
}

// twistTextDrawer 扭曲文本绘制器
// 实现了 TextDrawer 接口，用于绘制带有扭曲效果的文本
// 字段：
// - dpi: 每英寸点数，影响文本清晰度
// - r: 随机数生成器
// - randMu: 随机数生成器的互斥锁
// - randOnce: 确保随机数生成器只初始化一次
// - amplitude: 扭曲振幅
// - frequency: 扭曲频率
// - effects: 文本效果列表
// - modes: 扭曲模式列表

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

// DrawString 在画布上绘制带有扭曲效果的字符串
// 参数：
// - canvas: 要绘制的画布
// - text: 要绘制的文本
// 返回值：
// - error: 绘制过程中的错误
// 影响：在画布上绘制带有扭曲效果的文本，增加验证码的复杂度
func (t *twistTextDrawer) DrawString(canvas draw.Image, text string) error {
	runes, bounds, err := validateTextDrawInput(canvas, text)
	if err != nil {
		return err
	}

	// 创建临时画布
	textCanvas := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	// 在临时画布上绘制随机化文本
	if err := drawRandomizedText(textCanvas, textCanvas.Bounds(), runes, t.dpi, t.withRand); err != nil {
		return err
	}

	// 应用文本效果
	return applyTextEffects(textCanvas, canvas, t.resolveEffects())
}

// withRand 使用随机数生成器执行函数
// 参数：
// - fn: 要执行的函数
// 影响：确保在安全的情况下使用随机数生成器
func (t *twistTextDrawer) withRand(fn func(r *rand.Rand)) {
	withDrawerRand(&t.r, &t.randOnce, &t.randMu, fn)
}

// resolveEffects 解析文本效果
// 返回值：
// - []TextEffect: 文本效果列表
// 影响：如果指定了效果，则使用指定的效果；否则根据扭曲模式创建波浪效果
func (t *twistTextDrawer) resolveEffects() []TextEffect {
	if len(t.effects) > 0 {
		out := make([]TextEffect, len(t.effects))
		copy(out, t.effects)
		return out
	}

	// 标准化扭曲模式
	modes := normalizeDistortionModes(t.modes)
	mode := modes[0]
	// 随机选择一个扭曲模式
	t.withRand(func(r *rand.Rand) {
		mode = modes[randIntn(r, len(modes))]
	})

	// 创建波浪效果
	return []TextEffect{
		NewWaveTextEffect(t.amplitude, t.frequency, mode),
	}
}

// NewTwistTextDrawer 创建一个带有波浪效果的文本绘制器
// 参数：
// - dpi: 每英寸点数，影响文本清晰度
// - amplitude: 波浪高度
// - frequency: 波浪频率
// 返回值：
// - TextDrawer: 文本绘制器实例
// 示例：
// ```go
// // 创建一个带有波浪效果的文本绘制器
// textDrawer := gocaptcha.NewTwistTextDrawer(72.0, 2.0, 0.05)
// // 为验证码添加扭曲文本
// captcha.DrawText(textDrawer, "1234")
// ```
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

// NewTwistTextDrawerWithModes 创建一个带有可选波浪模式的文本绘制器
// 参数：
// - dpi: 每英寸点数，影响文本清晰度
// - amplitude: 波浪高度
// - frequency: 波浪频率
// - modes: 波浪扭曲模式列表
// 返回值：
// - TextDrawer: 文本绘制器实例
// 示例：
// ```go
// // 创建一个带有水平和垂直波浪模式的文本绘制器
// textDrawer := gocaptcha.NewTwistTextDrawerWithModes(72.0, 2.0, 0.05, gocaptcha.WaveDistortionHorizontal, gocaptcha.WaveDistortionVertical)
// // 为验证码添加扭曲文本
// captcha.DrawText(textDrawer, "1234")
// ```
func NewTwistTextDrawerWithModes(dpi float64, amplitude float64, frequency float64, modes ...WaveDistortionMode) TextDrawer {
	return &twistTextDrawer{
		dpi:       normalizeDPI(dpi),
		r:         newSecureSeededRand(),
		amplitude: normalizeAmplitude(amplitude),
		frequency: normalizeFrequency(frequency),
		modes:     normalizeDistortionModes(modes),
	}
}

// NewEffectTextDrawer 创建一个应用自定义文本效果的文本绘制器
// 参数：
// - dpi: 每英寸点数，影响文本清晰度
// - effects: 文本效果列表
// 返回值：
// - TextDrawer: 文本绘制器实例
// 示例：
// ```go
// // 创建一个带有自定义效果的文本绘制器
// waveEffect := gocaptcha.NewWaveTextEffect(2.0, 0.05, gocaptcha.WaveDistortionDual)
// textDrawer := gocaptcha.NewEffectTextDrawer(72.0, waveEffect)
// // 为验证码添加带有自定义效果的文本
// captcha.DrawText(textDrawer, "1234")
// ```
func NewEffectTextDrawer(dpi float64, effects ...TextEffect) TextDrawer {
	return &twistTextDrawer{
		dpi:     normalizeDPI(dpi),
		r:       newSecureSeededRand(),
		effects: append([]TextEffect(nil), effects...),
	}
}

// waveTextEffect 波浪文本效果
// 实现了 TextEffect 接口，用于创建波浪扭曲效果
// 字段：
// - amplitude: 波浪振幅
// - frequency: 波浪频率
// - mode: 波浪扭曲模式

type waveTextEffect struct {
	amplitude float64
	frequency float64
	mode      WaveDistortionMode
}

// NewWaveTextEffect 创建一个正弦波浪扭曲效果
// 参数：
// - amplitude: 波浪振幅
// - frequency: 波浪频率
// - mode: 波浪扭曲模式
// 返回值：
// - TextEffect: 文本效果实例
// 示例：
// ```go
// // 创建一个双波浪扭曲效果
// waveEffect := gocaptcha.NewWaveTextEffect(2.0, 0.05, gocaptcha.WaveDistortionDual)
// ```
func NewWaveTextEffect(amplitude float64, frequency float64, mode WaveDistortionMode) TextEffect {
	return &waveTextEffect{
		amplitude: normalizeAmplitude(amplitude),
		frequency: normalizeFrequency(frequency),
		mode:      normalizeDistortionMode(mode),
	}
}

// Apply 从源图像应用波浪效果到目标图像
// 参数：
// - src: 源图像
// - dst: 目标图像
// 返回值：
// - error: 应用过程中的错误
// 影响：根据指定的模式和参数应用波浪扭曲效果
func (w *waveTextEffect) Apply(src, dst *image.RGBA) error {
	if err := validateEffectBuffers(src, dst); err != nil {
		return err
	}
	clear(dst.Pix)

	// 标准化参数
	amplitude := normalizeAmplitude(w.amplitude)
	frequency := normalizeFrequency(w.frequency)
	mode := normalizeDistortionMode(w.mode)

	// 根据模式应用不同的波浪效果
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

// drawRandomizedText 绘制随机化的文本
// 参数：
// - dst: 目标画布
// - bounds: 边界
// - runes: 字符列表
// - dpi: 每英寸点数
// - withRand: 随机数生成器函数
// 返回值：
// - error: 绘制过程中的错误
// 影响：为每个字符使用不同的字体、大小、位置和颜色，增加验证码的复杂度
func drawRandomizedText(
	dst draw.Image,
	bounds image.Rectangle,
	runes []rune,
	dpi float64,
	withRand func(func(*rand.Rand)),
) error {
	// 获取字体
	fonts, err := DefaultFontFamily.WeightedCachedFonts()
	if err != nil {
		return err
	}

	// 创建freetype上下文
	c := freetype.NewContext()
	c.SetDPI(normalizeDPI(dpi))
	c.SetClip(bounds)
	c.SetDst(dst)
	c.SetHinting(font.HintingFull)

	// 预计算绘制参数
	params := precomputeTextDrawParams(len(runes), bounds, len(fonts), withRand)
	// 绘制每个字符
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

// precomputeTextDrawParams 预计算文本绘制参数
// 参数：
// - runeCount: 字符数量
// - bounds: 边界
// - fontCount: 字体数量
// - withRand: 随机数生成器函数
// 返回值：
// - textDrawParams: 文本绘制参数
// 影响：为每个字符生成随机的字体大小、位置、字体索引和颜色
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

	// 使用随机数生成器为每个字符生成参数
	withRand(func(r *rand.Rand) {
		for i := 0; i < runeCount; i++ {
			// 随机字体大小
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

// applyTextEffects 应用文本效果
// 参数：
// - src: 源图像
// - dst: 目标图像
// - effects: 文本效果列表
// 返回值：
// - error: 应用过程中的错误
// 影响：将多个文本效果依次应用到图像上
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

	// 依次应用每个效果
	for _, effect := range effects {
		clear(output.Pix)
		if err := effect.Apply(current, output); err != nil {
			return err
		}
		current = output
		// 切换输出缓冲区
		if output == bufA {
			output = bufB
		} else {
			output = bufA
		}
	}

	// 将结果绘制到目标图像
	draw.Draw(dst, dst.Bounds(), current, current.Bounds().Min, draw.Over)
	return nil
}

// applyHorizontalWave 应用水平波浪效果
// 参数：
// - src: 源图像
// - dst: 目标图像
// - shifts: 水平偏移量列表
// 影响：在水平方向上应用波浪扭曲效果
func applyHorizontalWave(src, dst *image.RGBA, shifts []int) {
	width := src.Bounds().Dx()
	srcBounds := src.Bounds()
	dstBounds := dst.Bounds()

	// 对每一行应用水平偏移
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

		// 复制像素
		srcOffset := src.PixOffset(srcBounds.Min.X+srcStartX, srcBounds.Min.Y+y)
		dstOffset := dst.PixOffset(dstBounds.Min.X+dstStartX, dstBounds.Min.Y+y)
		copy(dst.Pix[dstOffset:dstOffset+rowLength*4], src.Pix[srcOffset:srcOffset+rowLength*4])
	}
}

// applyVerticalWave 应用垂直波浪效果
// 参数：
// - src: 源图像
// - dst: 目标图像
// - shifts: 垂直偏移量列表
// 影响：在垂直方向上应用波浪扭曲效果
func applyVerticalWave(src, dst *image.RGBA, shifts []int) {
	height := src.Bounds().Dy()
	srcBounds := src.Bounds()
	dstBounds := dst.Bounds()

	// 对每一列应用垂直偏移
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

		// 复制像素
		for i := 0; i < colLength; i++ {
			srcOffset := src.PixOffset(srcBounds.Min.X+x, srcBounds.Min.Y+srcStartY+i)
			dstOffset := dst.PixOffset(dstBounds.Min.X+x, dstBounds.Min.Y+dstStartY+i)
			copy(dst.Pix[dstOffset:dstOffset+4], src.Pix[srcOffset:srcOffset+4])
		}
	}
}

// precomputeWaveShifts 预计算波浪偏移量
// 参数：
// - length: 长度
// - amplitude: 振幅
// - frequency: 频率
// 返回值：
// - []int: 偏移量列表
// 影响：计算正弦波的偏移量，用于波浪扭曲效果
func precomputeWaveShifts(length int, amplitude float64, frequency float64) []int {
	shifts := make([]int, length)
	// 计算每个位置的偏移量
	for i := 0; i < length; i++ {
		shifts[i] = int(amplitude * math.Sin(frequency*float64(i)))
	}
	return shifts
}

// validateTextDrawInput 验证文本绘制输入
// 参数：
// - canvas: 画布
// - text: 文本
// 返回值：
// - []rune: 字符列表
// - image.Rectangle: 边界
// - error: 错误
// 影响：验证输入是否有效，确保画布不为nil，文本不为空，画布尺寸有效
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

// validateEffectBuffers 验证效果缓冲区
// 参数：
// - src: 源图像
// - dst: 目标图像
// 返回值：
// - error: 错误
// 影响：验证源图像和目标图像是否有效，尺寸是否匹配
func validateEffectBuffers(src, dst *image.RGBA) error {
	if src == nil || dst == nil {
		return ErrNilEffectCanvas
	}
	if src.Bounds().Dx() != dst.Bounds().Dx() || src.Bounds().Dy() != dst.Bounds().Dy() {
		return ErrInvalidEffectSize
	}
	return nil
}

// normalizeDPI 标准化DPI
// 参数：
// - dpi: 每英寸点数
// 返回值：
// - float64: 标准化后的DPI
// 影响：确保DPI为正数，对无效值使用默认值
func normalizeDPI(dpi float64) float64 {
	if dpi <= 0 || math.IsNaN(dpi) {
		return defaultTextDPI
	}
	return dpi
}

// normalizeAmplitude 标准化振幅
// 参数：
// - amplitude: 振幅
// 返回值：
// - float64: 标准化后的振幅
// 影响：确保振幅为正数，对无效值使用默认值
func normalizeAmplitude(amplitude float64) float64 {
	if amplitude <= 0 || math.IsNaN(amplitude) {
		return defaultTwistAmplitude
	}
	return amplitude
}

// normalizeFrequency 标准化频率
// 参数：
// - frequency: 频率
// 返回值：
// - float64: 标准化后的频率
// 影响：确保频率为正数，对无效值使用默认值
func normalizeFrequency(frequency float64) float64 {
	if frequency <= 0 || math.IsNaN(frequency) {
		return defaultTwistFrequency
	}
	return frequency
}

// normalizeDistortionModes 标准化扭曲模式列表
// 参数：
// - modes: 扭曲模式列表
// 返回值：
// - []WaveDistortionMode: 标准化后的扭曲模式列表
// 影响：确保扭曲模式列表不为空，只包含有效的扭曲模式
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

// normalizeDistortionMode 标准化扭曲模式
// 参数：
// - mode: 扭曲模式
// 返回值：
// - WaveDistortionMode: 标准化后的扭曲模式
// 影响：确保扭曲模式有效，对无效模式使用默认值
func normalizeDistortionMode(mode WaveDistortionMode) WaveDistortionMode {
	switch mode {
	case WaveDistortionHorizontal, WaveDistortionVertical, WaveDistortionDual:
		return mode
	default:
		return WaveDistortionHorizontal
	}
}

// randDeepColorFrom 生成深色
// 参数：
// - r: 随机数生成器
// 返回值：
// - color.RGBA: 生成的颜色
// 影响：生成较深的随机颜色，用于文本
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

// withDrawerRand 使用随机数生成器执行函数
// 参数：
// - r: 随机数生成器指针的指针
// - once: 确保只初始化一次的同步对象
// - mu: 互斥锁
// - fn: 要执行的函数
// 影响：确保随机数生成器已初始化，在互斥锁保护下执行函数
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

// newSecureSeededRand 创建一个安全种子的随机数生成器
// 返回值：
// - *rand.Rand: 随机数生成器
// 影响：使用加密随机数生成种子，确保随机性
func newSecureSeededRand() *rand.Rand {
	var seedBytes [8]byte
	if _, err := crand.Read(seedBytes[:]); err != nil {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	seed := int64(binary.LittleEndian.Uint64(seedBytes[:]))
	return rand.New(rand.NewSource(seed))
}

// randIntn 生成指定范围内的随机整数
// 参数：
// - r: 随机数生成器
// - n: 随机数的上界（不包含）
// 返回值：
// - int: 0到n-1之间的随机整数
// 影响：生成指定范围内的随机整数
func randIntn(r *rand.Rand, n int) int {
	if n <= 1 {
		return 0
	}
	return r.Intn(n)
}

// maxInt 返回两个整数中的最大值
// 参数：
// - a: 第一个整数
// - b: 第二个整数
// 返回值：
// - int: 最大值
// 影响：返回两个整数中的最大值
func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// absInt 返回整数的绝对值
// 参数：
// - value: 整数
// 返回值：
// - int: 绝对值
// 影响：返回整数的绝对值
func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
