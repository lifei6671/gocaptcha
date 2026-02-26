// Package gocaptcha 提供生成验证码图片的功能
// 本文件实现了图片模糊处理相关的功能
package gocaptcha

import (
	"image"
	"image/draw"
	"math"
	"sync"
)

const (
	// defaultGaussianKernelSize 默认的高斯模糊卷积核大小
	// 影响：控制模糊效果的范围，值越大模糊范围越广
	defaultGaussianKernelSize = 5
	// defaultGaussianSigma 默认的高斯模糊sigma值
	// 影响：控制模糊的程度，值越大模糊效果越明显
	defaultGaussianSigma = 1.0
)

// BlurDrawer 模糊效果绘制器接口
// 定义了绘制模糊效果的方法
// 参数：
// - canvas: 要绘制的画布
// - kernelSize: 模糊卷积核大小
// - sigma: 模糊sigma值
// 返回值：
// - error: 绘制过程中的错误
// 实现此接口的类型可以为验证码添加模糊效果

type BlurDrawer interface {
	DrawBlur(canvas draw.Image, kernelSize int, sigma float64) error
}

// gaussianKernelKey 高斯核缓存的键结构体
// 用于在缓存中存储和查找不同参数的高斯核
// 字段：
// - kernelSize: 卷积核大小
// - sigmaBits: sigma值的位表示

type gaussianKernelKey struct {
	kernelSize int
	sigmaBits  uint64
}

// gaussianBlur 高斯模糊实现
// 实现了 BlurDrawer 接口，使用高斯卷积核实现模糊效果
// 字段：
// - kernelCache: 高斯核缓存，避免重复计算

type gaussianBlur struct {
	kernelCache sync.Map
}

// NewGaussianBlur 创建一个新的高斯模糊绘制器
// 返回值：
// - BlurDrawer: 高斯模糊绘制器实例
// 示例：
// ```go
// // 创建一个高斯模糊绘制器
// blurDrawer := gocaptcha.NewGaussianBlur()
// // 为验证码添加模糊效果
// captcha.DrawBlur(blurDrawer, 5, 1.0)
// ```
func NewGaussianBlur() BlurDrawer {
	return &gaussianBlur{}
}

// DrawBlur 在画布上应用高斯模糊效果
// 参数：
// - canvas: 要绘制的画布
// - kernelSize: 模糊卷积核大小
// - sigma: 模糊sigma值
// 返回值：
// - error: 绘制过程中的错误
// 影响：对画布应用高斯模糊效果，增加验证码的复杂度和安全性
func (g *gaussianBlur) DrawBlur(canvas draw.Image, kernelSize int, sigma float64) error {
	if canvas == nil {
		return ErrNilCanvas
	}

	bounds := canvas.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil
	}

	// 标准化模糊参数
	kernelSize, sigma = normalizeBlurParams(kernelSize, sigma)
	// 获取或生成高斯核
	kernel := g.getKernel(kernelSize, sigma)

	// 创建临时图像
	src := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(src, src.Bounds(), canvas, bounds.Min, draw.Src)

	tmp := image.NewRGBA(src.Bounds())
	dst := image.NewRGBA(src.Bounds())

	// 应用水平和垂直卷积
	convolveHorizontal(src, tmp, kernel)
	convolveVertical(tmp, dst, kernel)

	// 将结果绘制回原画布
	draw.Draw(canvas, bounds, dst, image.Point{}, draw.Src)
	return nil
}

// getKernel 获取或生成高斯卷积核
// 参数：
// - kernelSize: 卷积核大小
// - sigma: 模糊sigma值
// 返回值：
// - []float64: 高斯卷积核
// 影响：从缓存中获取已计算的核或生成新核并缓存，提高性能
func (g *gaussianBlur) getKernel(kernelSize int, sigma float64) []float64 {
	// 创建缓存键
	key := gaussianKernelKey{
		kernelSize: kernelSize,
		sigmaBits:  math.Float64bits(sigma),
	}
	// 尝试从缓存中加载
	if cached, ok := g.kernelCache.Load(key); ok {
		return cached.([]float64)
	}

	// 生成新的高斯核
	kernel := generateGaussianKernel1D(kernelSize, sigma)
	// 存储到缓存并返回
	actual, _ := g.kernelCache.LoadOrStore(key, kernel)
	return actual.([]float64)
}

// normalizeBlurParams 标准化模糊参数
// 参数：
// - kernelSize: 卷积核大小
// - sigma: 模糊sigma值
// 返回值：
// - int: 标准化后的卷积核大小
// - float64: 标准化后的sigma值
// 影响：确保参数有效，对无效参数使用默认值
func normalizeBlurParams(kernelSize int, sigma float64) (int, float64) {
	// 确保卷积核大小为正奇数
	if kernelSize <= 0 || kernelSize%2 == 0 {
		kernelSize = defaultGaussianKernelSize
	}
	// 确保sigma值有效
	if sigma <= 0 || math.IsNaN(sigma) || math.IsInf(sigma, 0) {
		sigma = defaultGaussianSigma
	}
	return kernelSize, sigma
}

// generateGaussianKernel1D 生成一维高斯卷积核
// 参数：
// - kernelSize: 卷积核大小
// - sigma: 模糊sigma值
// 返回值：
// - []float64: 一维高斯卷积核
// 影响：生成用于模糊处理的高斯核，决定模糊效果的特性
func generateGaussianKernel1D(kernelSize int, sigma float64) []float64 {
	radius := kernelSize / 2
	kernel := make([]float64, kernelSize)
	denominator := 2 * sigma * sigma

	sum := 0.0
	// 计算高斯函数值
	for i := -radius; i <= radius; i++ {
		value := math.Exp(-(float64(i * i)) / denominator)
		kernel[i+radius] = value
		sum += value
	}

	// 归一化核
	if sum > 0 {
		for i := range kernel {
			kernel[i] /= sum
		}
	}
	return kernel
}

// convolveHorizontal 水平方向卷积
// 参数：
// - src: 源图像
// - dst: 目标图像
// - kernel: 卷积核
// 影响：在水平方向应用卷积，实现水平方向的模糊效果
func convolveHorizontal(src *image.RGBA, dst *image.RGBA, kernel []float64) {
	width := src.Bounds().Dx()
	height := src.Bounds().Dy()
	radius := len(kernel) / 2

	for y := 0; y < height; y++ {
		srcRowOffset := y * src.Stride
		dstRowOffset := y * dst.Stride

		for x := 0; x < width; x++ {
			var rAcc, gAcc, bAcc, aAcc float64

			// 应用卷积核
			for k := -radius; k <= radius; k++ {
				// 镜像索引处理边界
				sampleX := mirrorIndex(x+k, width)
				weight := kernel[k+radius]
				srcOffset := srcRowOffset + sampleX*4

				// 计算加权和
				rAcc += float64(src.Pix[srcOffset]) * weight
				gAcc += float64(src.Pix[srcOffset+1]) * weight
				bAcc += float64(src.Pix[srcOffset+2]) * weight
				aAcc += float64(src.Pix[srcOffset+3]) * weight
			}

			// 写入结果
			dstOffset := dstRowOffset + x*4
			dst.Pix[dstOffset] = clampToByte(rAcc)
			dst.Pix[dstOffset+1] = clampToByte(gAcc)
			dst.Pix[dstOffset+2] = clampToByte(bAcc)
			dst.Pix[dstOffset+3] = clampToByte(aAcc)
		}
	}
}

// convolveVertical 垂直方向卷积
// 参数：
// - src: 源图像
// - dst: 目标图像
// - kernel: 卷积核
// 影响：在垂直方向应用卷积，实现垂直方向的模糊效果
func convolveVertical(src *image.RGBA, dst *image.RGBA, kernel []float64) {
	width := src.Bounds().Dx()
	height := src.Bounds().Dy()
	radius := len(kernel) / 2

	for y := 0; y < height; y++ {
		dstRowOffset := y * dst.Stride
		for x := 0; x < width; x++ {
			var rAcc, gAcc, bAcc, aAcc float64

			// 应用卷积核
			for k := -radius; k <= radius; k++ {
				// 镜像索引处理边界
				sampleY := mirrorIndex(y+k, height)
				weight := kernel[k+radius]
				srcOffset := sampleY*src.Stride + x*4

				// 计算加权和
				rAcc += float64(src.Pix[srcOffset]) * weight
				gAcc += float64(src.Pix[srcOffset+1]) * weight
				bAcc += float64(src.Pix[srcOffset+2]) * weight
				aAcc += float64(src.Pix[srcOffset+3]) * weight
			}

			// 写入结果
			dstOffset := dstRowOffset + x*4
			dst.Pix[dstOffset] = clampToByte(rAcc)
			dst.Pix[dstOffset+1] = clampToByte(gAcc)
			dst.Pix[dstOffset+2] = clampToByte(bAcc)
			dst.Pix[dstOffset+3] = clampToByte(aAcc)
		}
	}
}

// mirrorIndex 镜像索引处理
// 参数：
// - index: 原始索引
// - length: 数组长度
// 返回值：
// - int: 处理后的索引
// 影响：当索引超出范围时，通过镜像方式处理，避免边界效应
func mirrorIndex(index int, length int) int {
	if length <= 1 {
		return 0
	}
	// 镜像处理索引
	for index < 0 || index >= length {
		if index < 0 {
			index = -index - 1
		} else {
			index = 2*length - index - 1
		}
	}
	return index
}

// clampToByte 将浮点数限制在0-255范围内并转换为字节
// 参数：
// - value: 浮点数值
// 返回值：
// - uint8: 转换后的字节值
// 影响：确保颜色值在有效范围内
func clampToByte(value float64) uint8 {
	if value <= 0 {
		return 0
	}
	if value >= 255 {
		return 255
	}
	return uint8(value + 0.5)
}
