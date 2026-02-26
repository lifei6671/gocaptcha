// Package gocaptcha 提供生成验证码图片的功能。
// 该包支持创建各种类型的验证码图片，包括添加噪点、线条、文字等元素。
package gocaptcha

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"sync"
)

const (
	// DefaultDPI 默认的DPI（每英寸点数）
	// 影响：控制生成图片的清晰度，值越高图片越清晰，但文件大小也会增加
	DefaultDPI = 72.0
	// DefaultBlurKernelSize 默认模糊卷积核大小
	// 影响：控制模糊效果的范围，值越大模糊范围越广
	DefaultBlurKernelSize = 2
	// DefaultBlurSigma 默认模糊sigma值
	// 影响：控制模糊的程度，值越大模糊效果越明显
	DefaultBlurSigma = 0.15
	// DefaultAmplitude 默认图片扭曲的振幅
	// 影响：控制图片扭曲的程度，值越大扭曲效果越明显
	DefaultAmplitude = 20
	// DefaultFrequency 默认图片扭曲的波频率
	// 影响：控制扭曲波纹的密度，值越大波纹越密集
	DefaultFrequency = 0.05
)

// TextCharacters 用于生成验证码的字符集
// 包含大小写字母和数字，可根据需要修改以调整验证码的复杂度
var TextCharacters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

const (
	// ImageFormatPng PNG图片格式
	ImageFormatPng ImageFormat = iota
	// ImageFormatJpeg JPEG图片格式
	ImageFormatJpeg
	// ImageFormatGif GIF图片格式
	ImageFormatGif
)

// ImageFormat 图片格式类型
// 支持PNG、JPEG和GIF三种格式
type ImageFormat int

// CaptchaImage 验证码图片结构体
// 包含图片的基本信息和绘制状态
type CaptchaImage struct {
	nrgba   *image.NRGBA // 底层的图片对象
	width   int          // 图片宽度
	height  int          // 图片高度
	Complex int          // 复杂度级别
	Error   error        // 错误信息
	r       *rand.Rand   // 随机数生成器
	randMu  sync.Mutex   // 随机数生成器的互斥锁
}

// New 创建一个新的验证码图片对象
// 参数：
// - width: 图片宽度
// - height: 图片高度
// - bgColor: 背景颜色
// 返回值：
// - *CaptchaImage: 验证码图片对象
// 示例：
// ```go
// // 创建一个100x50的验证码图片，背景色为白色
// captcha := gocaptcha.New(100, 50, color.RGBA{255, 255, 255, 255})
// ```
func New(width int, height int, bgColor color.RGBA) *CaptchaImage {
	return NewWithOptions(width, height, WithBackgroundColor(bgColor))
}

// captchaOptions 验证码图片的配置选项
// 用于存储通过功能选项模式设置的参数
type captchaOptions struct {
	bgColor    color.RGBA // 背景颜色
	bgColorSet bool       // 是否设置了背景颜色
	seedSet    bool       // 是否设置了随机种子
	seed       int64      // 随机种子值
}

// CaptchaOption 验证码图片构造的配置函数类型
// 使用功能选项模式来配置验证码图片的创建
type CaptchaOption func(*captchaOptions)

// WithBackgroundColor 设置验证码背景颜色
// 参数：
// - bgColor: 背景颜色
// 返回值：
// - CaptchaOption: 配置函数
func WithBackgroundColor(bgColor color.RGBA) CaptchaOption {
	return func(opts *captchaOptions) {
		opts.bgColor = bgColor
		opts.bgColorSet = true
	}
}

// WithRandomSeed 设置验证码的随机种子
// 参数：
// - seed: 随机种子值
// 返回值：
// - CaptchaOption: 配置函数
// 影响：设置随机种子后，生成的验证码将具有确定性，相同种子会生成相同的验证码
func WithRandomSeed(seed int64) CaptchaOption {
	return func(opts *captchaOptions) {
		opts.seed = seed
		opts.seedSet = true
	}
}

// NewWithOptions 使用功能选项模式创建验证码图片
// 参数：
// - width: 图片宽度
// - height: 图片高度
// - options: 配置选项函数列表
// 返回值：
// - *CaptchaImage: 验证码图片对象
// 示例：
// ```go
// // 创建一个120x60的验证码图片，使用自定义背景色和随机种子
// captcha := gocaptcha.NewWithOptions(
//
//	120, 60,
//	gocaptcha.WithBackgroundColor(color.RGBA{240, 240, 240, 255}),
//	gocaptcha.WithRandomSeed(12345),
//
// )
// ```
func NewWithOptions(width int, height int, options ...CaptchaOption) *CaptchaImage {
	// 确保最小尺寸以避免下游操作出错
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}

	opts := captchaOptions{}
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}
	if !opts.bgColorSet {
		opts.bgColor = RandLightColor()
	}

	randomizer := newSecureSeededRand()
	if opts.seedSet {
		randomizer = rand.New(rand.NewSource(opts.seed))
	}

	m := image.NewNRGBA(image.Rect(0, 0, width, height))

	draw.Draw(m, m.Bounds(), &image.Uniform{C: opts.bgColor}, image.Point{}, draw.Src)

	return &CaptchaImage{
		nrgba:  m,
		height: height,
		width:  width,
		r:      randomizer,
	}
}

// Encode 将验证码图片编码为指定格式并写入输出流
// 参数：
// - w: 输出流
// - imageFormat: 图片格式（PNG、JPEG或GIF）
// 返回值：
// - error: 编码过程中的错误
// 示例：
// ```go
// // 将验证码编码为PNG格式并写入文件
// file, _ := os.Create("captcha.png")
// defer file.Close()
// captcha.Encode(file, gocaptcha.ImageFormatPng)
// ```
func (captcha *CaptchaImage) Encode(w io.Writer, imageFormat ImageFormat) error {

	if imageFormat == ImageFormatPng {
		return png.Encode(w, captcha.nrgba)
	}
	if imageFormat == ImageFormatJpeg {
		return jpeg.Encode(w, captcha.nrgba, &jpeg.Options{Quality: 100})
	}
	if imageFormat == ImageFormatGif {
		return gif.Encode(w, captcha.nrgba, &gif.Options{NumColors: 256})
	}

	return errors.New("not supported image format")
}

// DrawLine 在验证码图片上绘制直线
// 参数：
// - drawer: 直线绘制器
// - lineColor: 线条颜色
// 返回值：
// - *CaptchaImage: 验证码图片对象（支持链式调用）
// 影响：在图片上随机位置绘制一条直线，增加验证码的复杂度
func (captcha *CaptchaImage) DrawLine(drawer LineDrawer, lineColor color.Color) *CaptchaImage {
	if captcha.Error != nil {
		return captcha
	}
	if drawer == nil {
		captcha.Error = errors.New("nil LineDrawer")
		return captcha
	}
	y := captcha.nrgba.Bounds().Dy()
	if y <= 0 {
		captcha.Error = errors.New("image has zero height")
		return captcha
	}
	point1 := image.Point{X: captcha.nrgba.Bounds().Min.X + 1, Y: captcha.randIntn(y)}
	point2 := image.Point{X: captcha.nrgba.Bounds().Max.X - 1, Y: captcha.randIntn(y)}
	captcha.Error = drawer.DrawLine(captcha.nrgba, point1, point2, lineColor)
	return captcha
}

// DrawBorder 在验证码图片上绘制边框
// 参数：
// - borderColor: 边框颜色
// 返回值：
// - *CaptchaImage: 验证码图片对象（支持链式调用）
// 影响：在图片四周绘制边框，使验证码更加清晰可见
func (captcha *CaptchaImage) DrawBorder(borderColor color.RGBA) *CaptchaImage {
	if captcha.Error != nil {
		return captcha
	}
	for x := 0; x < captcha.width; x++ {
		captcha.nrgba.Set(x, 0, borderColor)
		captcha.nrgba.Set(x, captcha.height-1, borderColor)
	}
	for y := 0; y < captcha.height; y++ {
		captcha.nrgba.Set(0, y, borderColor)
		captcha.nrgba.Set(captcha.width-1, y, borderColor)
	}
	return captcha
}

// DrawNoise 在验证码图片上绘制噪点
// 参数：
// - complex: 噪点密度
// - noiseDrawer: 噪点绘制器
// 返回值：
// - *CaptchaImage: 验证码图片对象（支持链式调用）
// 影响：在图片上绘制噪点，增加验证码的复杂度，提高安全性
func (captcha *CaptchaImage) DrawNoise(complex NoiseDensity, noiseDrawer NoiseDrawer) *CaptchaImage {
	if captcha.Error != nil {
		return captcha
	}
	if noiseDrawer == nil {
		captcha.Error = errors.New("nil NoiseDrawer")
		return captcha
	}
	captcha.Error = noiseDrawer.DrawNoise(captcha.nrgba, complex)
	return captcha
}

// DrawNoiseWithConfig 使用配置绘制噪点（当绘制器支持配置时）
// 参数：
// - noiseDrawer: 噪点绘制器
// - config: 噪点配置
// 返回值：
// - *CaptchaImage: 验证码图片对象（支持链式调用）
// 影响：根据配置在图片上绘制噪点，提供更灵活的噪点控制
func (captcha *CaptchaImage) DrawNoiseWithConfig(noiseDrawer NoiseDrawer, config NoiseConfig) *CaptchaImage {
	if captcha.Error != nil {
		return captcha
	}
	if noiseDrawer == nil {
		captcha.Error = errors.New("nil NoiseDrawer")
		return captcha
	}

	if configurable, ok := noiseDrawer.(ConfigurableNoiseDrawer); ok {
		captcha.Error = configurable.DrawNoiseWithConfig(captcha.nrgba, config)
		return captcha
	}

	captcha.Error = noiseDrawer.DrawNoise(captcha.nrgba, config.Density)
	return captcha
}

// DrawText 在验证码图片上绘制文字
// 参数：
// - textDrawer: 文字绘制器
// - text: 要绘制的文字
// 返回值：
// - *CaptchaImage: 验证码图片对象（支持链式调用）
// 影响：在图片上绘制指定文字，这是验证码的核心内容
func (captcha *CaptchaImage) DrawText(textDrawer TextDrawer, text string) *CaptchaImage {
	if captcha.Error != nil {
		return captcha
	}
	if textDrawer == nil {
		captcha.Error = errors.New("nil TextDrawer")
		return captcha
	}
	captcha.Error = textDrawer.DrawString(captcha.nrgba, text)
	return captcha
}

// DrawBlur 对验证码图片进行模糊处理
// 参数：
// - drawer: 模糊效果绘制器
// - kernelSize: 模糊卷积核大小
// - sigma: 模糊sigma值
// 返回值：
// - *CaptchaImage: 验证码图片对象（支持链式调用）
// 影响：对图片应用模糊效果，增加验证码的复杂度和安全性
func (captcha *CaptchaImage) DrawBlur(drawer BlurDrawer, kernelSize int, sigma float64) *CaptchaImage {
	if captcha.Error != nil {
		return captcha
	}
	if drawer == nil {
		captcha.Error = errors.New("nil BlurDrawer")
		return captcha
	}
	captcha.Error = drawer.DrawBlur(captcha.nrgba, kernelSize, sigma)
	return captcha
}

// randIntn 生成指定范围内的随机整数
// 参数：
// - n: 随机数的上界（不包含）
// 返回值：
// - int: 0到n-1之间的随机整数
// 内部方法，用于在验证码生成过程中产生随机值
func (captcha *CaptchaImage) randIntn(n int) int {
	if n <= 1 {
		return 0
	}

	captcha.randMu.Lock()
	defer captcha.randMu.Unlock()
	if captcha.r == nil {
		captcha.r = newSecureSeededRand()
	}
	return captcha.r.Intn(n)
}
