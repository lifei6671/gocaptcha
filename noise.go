// Package gocaptcha 提供生成验证码图片的功能
// 本文件实现了噪声生成相关的功能，用于在验证码中添加噪声元素
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

// NoiseDensity 噪声密度类型
// 表示验证码的复杂度级别
type NoiseDensity int

const (
	// NoiseDensityLower 低噪声密度
	// 生成较少的噪声，验证码较容易识别
	NoiseDensityLower NoiseDensity = iota
	// NoiseDensityMedium 中等噪声密度
	// 生成适量的噪声，验证码难度适中
	NoiseDensityMedium
	// NoiseDensityHigh 高噪声密度
	// 生成较多的噪声，验证码较难识别
	NoiseDensityHigh
)

const (
	// defaultSecondaryPointChance 默认的次要点出现概率
	// 影响：控制次要噪声点的生成概率
	defaultSecondaryPointChance = 1.0 / 3.0
	// defaultNoiseTextLength 默认的噪声文本长度
	// 影响：控制每个噪声文本的字符数量
	defaultNoiseTextLength = 1
	// defaultNoiseFontJitter 默认的字体大小抖动
	// 影响：控制噪声文本字体大小的变化范围
	defaultNoiseFontJitter = 5
)

// NoiseColorFunc 噪声颜色生成函数类型
// 用于自定义噪声的颜色生成
// 参数：
// - r: 随机数生成器
// 返回值：
// - color.Color: 生成的颜色

type NoiseColorFunc func(r *rand.Rand) color.Color

// randDeepNoiseColor 生成深色噪声颜色
// 参数：
// - r: 随机数生成器
// 返回值：
// - color.Color: 生成的深色
func randDeepNoiseColor(r *rand.Rand) color.Color {
	return randDeepColorFrom(r)
}

// NoiseConfig 噪声生成配置
// 控制噪声生成的行为
// 字段：
// - Density: 噪声密度级别
// - PointDensityDivisor: 点噪声密度除数（值越小，噪声越多）
// - TextDensityDivisor: 文本噪声密度除数（值越小，噪声越多）
// - SecondaryPointChance: 次要点出现概率
// - TextLength: 噪声文本长度
// - FontSizeJitter: 字体大小抖动
// - PointColor: 点噪声颜色生成函数
// - TextColor: 文本噪声颜色生成函数

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

// NoiseDrawer 噪声绘制器接口
// 定义了在图像上绘制噪声的方法
// 方法：
// - DrawNoise: 在图像上绘制噪声
//   参数：
//   - img: 要绘制的图像
//   - density: 噪声密度
//   返回值：
//   - error: 绘制过程中的错误

type NoiseDrawer interface {
	// DrawNoise draws noise on the image
	DrawNoise(img draw.Image, density NoiseDensity) error
}

// ConfigurableNoiseDrawer 可配置噪声绘制器接口
// 扩展了 NoiseDrawer 接口，支持配置化的噪声生成
// 方法：
// - DrawNoiseWithConfig: 使用配置在图像上绘制噪声
//   参数：
//   - img: 要绘制的图像
//   - config: 噪声配置
//   返回值：
//   - error: 绘制过程中的错误

type ConfigurableNoiseDrawer interface {
	NoiseDrawer
	DrawNoiseWithConfig(img draw.Image, config NoiseConfig) error
}

// pointNoiseDrawer 点噪声绘制器
// 实现了 ConfigurableNoiseDrawer 接口，用于绘制点噪声
// 字段：
// - r: 随机数生成器
// - randMu: 随机数生成器的互斥锁
// - randOnce: 确保随机数生成器只初始化一次

type pointNoiseDrawer struct {
	r        *rand.Rand
	randMu   sync.Mutex
	randOnce sync.Once
}

// DrawNoise 在图像上绘制点噪声
// 参数：
// - img: 要绘制的图像
// - density: 噪声密度
// 返回值：
// - error: 绘制过程中的错误
// 影响：在图像上绘制指定密度的点噪声，增加验证码的复杂度
func (n *pointNoiseDrawer) DrawNoise(img draw.Image, density NoiseDensity) error {
	return n.DrawNoiseWithConfig(img, NoiseConfig{Density: density})
}

// DrawNoiseWithConfig 使用配置在图像上绘制点噪声
// 参数：
// - img: 要绘制的图像
// - config: 噪声配置
// 返回值：
// - error: 绘制过程中的错误
// 影响：根据配置在图像上绘制点噪声，提供更灵活的噪声控制
func (n *pointNoiseDrawer) DrawNoiseWithConfig(img draw.Image, config NoiseConfig) error {
	if img == nil {
		return ErrNilCanvas
	}
	// 标准化噪声配置
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
	// 计算噪声点数量
	total := (width * height) / maxInt(1, divisor)
	if total <= 0 {
		return nil
	}

	secondaryChance := config.SecondaryPointChance
	pointColor := config.PointColor

	// 使用随机数生成器绘制噪声
	n.withRand(func(r *rand.Rand) {
		for i := 0; i < total; i++ {
			x := randIntn(r, width)
			y := randIntn(r, height)

			// 绘制主要噪声点
			img.Set(bounds.Min.X+x, bounds.Min.Y+y, noiseColorFrom(r, pointColor, randColorFromRand))
			// 绘制次要噪声点
			if secondaryChance > 0 && r.Float64() < secondaryChance && x+1 < width && y+1 < height {
				img.Set(bounds.Min.X+x+1, bounds.Min.Y+y+1, noiseColorFrom(r, pointColor, randColorFromRand))
			}
		}
	})
	return nil
}

// withRand 使用随机数生成器执行函数
// 参数：
// - fn: 要执行的函数
// 影响：确保在安全的情况下使用随机数生成器
func (n *pointNoiseDrawer) withRand(fn func(r *rand.Rand)) {
	withDrawerRand(&n.r, &n.randOnce, &n.randMu, fn)
}

// NewPointNoiseDrawer 创建一个点噪声绘制器
// 返回值：
// - NoiseDrawer: 点噪声绘制器实例
// 示例：
// ```go
// // 创建一个点噪声绘制器
// noiseDrawer := gocaptcha.NewPointNoiseDrawer()
// // 为验证码添加点噪声
// captcha.DrawNoise(gocaptcha.NoiseDensityMedium, noiseDrawer)
// ```
func NewPointNoiseDrawer() NoiseDrawer {
	return &pointNoiseDrawer{
		r: newSecureSeededRand(),
	}
}

// textNoiseDrawer 文本噪声绘制器
// 实现了 ConfigurableNoiseDrawer 接口，用于绘制文本噪声
// 字段：
// - r: 随机数生成器
// - randMu: 随机数生成器的互斥锁
// - randOnce: 确保随机数生成器只初始化一次
// - dpi: 每英寸点数，影响文本清晰度
// - fontFamily: 字体家族，用于文本渲染

type textNoiseDrawer struct {
	r          *rand.Rand
	randMu     sync.Mutex
	randOnce   sync.Once
	dpi        float64
	fontFamily *FontFamily
}

// DrawNoise 在图像上绘制文本噪声
// 参数：
// - img: 要绘制的图像
// - density: 噪声密度
// 返回值：
// - error: 绘制过程中的错误
// 影响：在图像上绘制指定密度的文本噪声，增加验证码的复杂度
func (n *textNoiseDrawer) DrawNoise(img draw.Image, density NoiseDensity) error {
	return n.DrawNoiseWithConfig(img, NoiseConfig{Density: density})
}

// DrawNoiseWithConfig 使用配置在图像上绘制文本噪声
// 参数：
// - img: 要绘制的图像
// - config: 噪声配置
// 返回值：
// - error: 绘制过程中的错误
// 影响：根据配置在图像上绘制文本噪声，提供更灵活的噪声控制
func (n *textNoiseDrawer) DrawNoiseWithConfig(img draw.Image, config NoiseConfig) error {
	if img == nil {
		return ErrNilCanvas
	}
	// 标准化噪声配置
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
	// 计算噪声文本数量
	total := (width * height) / maxInt(1, divisor)
	if total <= 0 {
		return nil
	}

	// 创建freetype上下文
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
	// 获取字体
	fonts, err := fontFamily.WeightedCachedFonts()
	if err != nil {
		return err
	}

	var drawErr error
	// 使用随机数生成器绘制噪声
	n.withRand(func(r *rand.Rand) {
		// 计算基础字体大小
		rawFontSize := float64(height) / (1 + float64(randIntn(r, 7))/10.0)
		if rawFontSize < 1 {
			rawFontSize = 1
		}

		for i := 0; i < total; i++ {
			if drawErr != nil {
				return
			}

			// 随机选择字体
			c.SetFont(fonts[randIntn(r, len(fonts))])

			// 随机位置
			x := bounds.Min.X + randIntn(r, width)
			y := bounds.Min.Y + randIntn(r, height)
			// 随机字体大小
			fontSize := rawFontSize/2 + float64(randIntn(r, fontJitter))
			if fontSize < 1 {
				fontSize = 1
			}

			// 设置文本颜色和大小
			c.SetSrc(image.NewUniform(noiseColorFrom(r, textColor, randLightColorFromRand)))
			c.SetFontSize(fontSize)
			// 绘制随机文本
			if _, err := c.DrawString(randTextFromRand(r, textLength), freetype.Pt(x, y)); err != nil {
				drawErr = err
				return
			}
		}
	})

	return drawErr
}

// withRand 使用随机数生成器执行函数
// 参数：
// - fn: 要执行的函数
// 影响：确保在安全的情况下使用随机数生成器
func (n *textNoiseDrawer) withRand(fn func(r *rand.Rand)) {
	withDrawerRand(&n.r, &n.randOnce, &n.randMu, fn)
}

// randTextFromRand 从随机数生成器生成随机文本
// 参数：
// - r: 随机数生成器
// - num: 文本长度
// 返回值：
// - string: 生成的随机文本
// 影响：生成指定长度的随机文本，用于文本噪声
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

// normalizeNoiseConfig 标准化噪声配置
// 参数：
// - config: 原始噪声配置
// 返回值：
// - NoiseConfig: 标准化后的噪声配置
// 影响：确保配置参数在有效范围内，对无效参数使用默认值
func normalizeNoiseConfig(config NoiseConfig) NoiseConfig {
	// 确保密度级别在有效范围内
	if config.Density < NoiseDensityLower || config.Density > NoiseDensityHigh {
		config.Density = NoiseDensityMedium
	}
	// 确保次要点出现概率在有效范围内
	if config.SecondaryPointChance < 0 || math.IsNaN(config.SecondaryPointChance) || math.IsInf(config.SecondaryPointChance, 0) {
		config.SecondaryPointChance = defaultSecondaryPointChance
	}
	if config.SecondaryPointChance > 1 {
		config.SecondaryPointChance = 1
	}
	// 确保文本长度在有效范围内
	if config.TextLength <= 0 {
		config.TextLength = defaultNoiseTextLength
	}
	// 确保字体大小抖动在有效范围内
	if config.FontSizeJitter <= 0 {
		config.FontSizeJitter = defaultNoiseFontJitter
	}
	return config
}

// pointDensityDivisorFor 根据密度级别获取点噪声密度除数
// 参数：
// - density: 噪声密度级别
// 返回值：
// - int: 点噪声密度除数
// 影响：密度级别越高，除数越小，噪声越多
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

// textDensityDivisorFor 根据密度级别获取文本噪声密度除数
// 参数：
// - density: 噪声密度级别
// 返回值：
// - int: 文本噪声密度除数
// 影响：密度级别越高，除数越小，噪声越多
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

// noiseColorFrom 从函数或回退函数生成噪声颜色
// 参数：
// - r: 随机数生成器
// - fn: 噪声颜色生成函数
// - fallback: 回退颜色生成函数
// 返回值：
// - color.Color: 生成的颜色
// 影响：优先使用提供的颜色生成函数，否则使用回退函数
func noiseColorFrom(r *rand.Rand, fn NoiseColorFunc, fallback func(*rand.Rand) color.Color) color.Color {
	if fn != nil {
		return fn(r)
	}
	return fallback(r)
}

// randLightColorFromRand 从随机数生成器生成亮色
// 参数：
// - r: 随机数生成器
// 返回值：
// - color.Color: 生成的亮色
// 影响：生成较亮的随机颜色，用于文本噪声
func randLightColorFromRand(r *rand.Rand) color.Color {
	return color.RGBA{
		R: uint8(randIntn(r, 128) + 128),
		G: uint8(randIntn(r, 128) + 128),
		B: uint8(randIntn(r, 128) + 128),
		A: 255,
	}
}

// randColorFromRand 从随机数生成器生成颜色
// 参数：
// - r: 随机数生成器
// 返回值：
// - color.Color: 生成的颜色
// 影响：生成随机颜色，用于点噪声
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

// NewTextNoiseDrawer 创建一个文本噪声绘制器
// 参数：
// - dpi: 每英寸点数，影响文本清晰度
// 返回值：
// - NoiseDrawer: 文本噪声绘制器实例
// 示例：
// ```go
// // 创建一个文本噪声绘制器
// noiseDrawer := gocaptcha.NewTextNoiseDrawer(72.0)
// // 为验证码添加文本噪声
// captcha.DrawNoise(gocaptcha.NoiseDensityMedium, noiseDrawer)
// ```
func NewTextNoiseDrawer(dpi float64) NoiseDrawer {
	return &textNoiseDrawer{
		r:          newSecureSeededRand(),
		dpi:        dpi,
		fontFamily: DefaultFontFamily,
	}
}
