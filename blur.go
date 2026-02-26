package gocaptcha

import (
	"image"
	"image/draw"
	"math"
	"sync"
)

const (
	defaultGaussianKernelSize = 5
	defaultGaussianSigma      = 1.0
)

type BlurDrawer interface {
	DrawBlur(canvas draw.Image, kernelSize int, sigma float64) error
}

type gaussianKernelKey struct {
	kernelSize int
	sigmaBits  uint64
}

type gaussianBlur struct {
	kernelCache sync.Map
}

func NewGaussianBlur() BlurDrawer {
	return &gaussianBlur{}
}

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

	kernelSize, sigma = normalizeBlurParams(kernelSize, sigma)
	kernel := g.getKernel(kernelSize, sigma)

	src := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(src, src.Bounds(), canvas, bounds.Min, draw.Src)

	tmp := image.NewRGBA(src.Bounds())
	dst := image.NewRGBA(src.Bounds())

	convolveHorizontal(src, tmp, kernel)
	convolveVertical(tmp, dst, kernel)

	draw.Draw(canvas, bounds, dst, image.Point{}, draw.Src)
	return nil
}

func (g *gaussianBlur) getKernel(kernelSize int, sigma float64) []float64 {
	key := gaussianKernelKey{
		kernelSize: kernelSize,
		sigmaBits:  math.Float64bits(sigma),
	}
	if cached, ok := g.kernelCache.Load(key); ok {
		return cached.([]float64)
	}

	kernel := generateGaussianKernel1D(kernelSize, sigma)
	actual, _ := g.kernelCache.LoadOrStore(key, kernel)
	return actual.([]float64)
}

func normalizeBlurParams(kernelSize int, sigma float64) (int, float64) {
	if kernelSize <= 0 || kernelSize%2 == 0 {
		kernelSize = defaultGaussianKernelSize
	}
	if sigma <= 0 || math.IsNaN(sigma) || math.IsInf(sigma, 0) {
		sigma = defaultGaussianSigma
	}
	return kernelSize, sigma
}

func generateGaussianKernel1D(kernelSize int, sigma float64) []float64 {
	radius := kernelSize / 2
	kernel := make([]float64, kernelSize)
	denominator := 2 * sigma * sigma

	sum := 0.0
	for i := -radius; i <= radius; i++ {
		value := math.Exp(-(float64(i * i)) / denominator)
		kernel[i+radius] = value
		sum += value
	}

	if sum > 0 {
		for i := range kernel {
			kernel[i] /= sum
		}
	}
	return kernel
}

func convolveHorizontal(src *image.RGBA, dst *image.RGBA, kernel []float64) {
	width := src.Bounds().Dx()
	height := src.Bounds().Dy()
	radius := len(kernel) / 2

	for y := 0; y < height; y++ {
		srcRowOffset := y * src.Stride
		dstRowOffset := y * dst.Stride

		for x := 0; x < width; x++ {
			var rAcc, gAcc, bAcc, aAcc float64

			for k := -radius; k <= radius; k++ {
				sampleX := mirrorIndex(x+k, width)
				weight := kernel[k+radius]
				srcOffset := srcRowOffset + sampleX*4

				rAcc += float64(src.Pix[srcOffset]) * weight
				gAcc += float64(src.Pix[srcOffset+1]) * weight
				bAcc += float64(src.Pix[srcOffset+2]) * weight
				aAcc += float64(src.Pix[srcOffset+3]) * weight
			}

			dstOffset := dstRowOffset + x*4
			dst.Pix[dstOffset] = clampToByte(rAcc)
			dst.Pix[dstOffset+1] = clampToByte(gAcc)
			dst.Pix[dstOffset+2] = clampToByte(bAcc)
			dst.Pix[dstOffset+3] = clampToByte(aAcc)
		}
	}
}

func convolveVertical(src *image.RGBA, dst *image.RGBA, kernel []float64) {
	width := src.Bounds().Dx()
	height := src.Bounds().Dy()
	radius := len(kernel) / 2

	for y := 0; y < height; y++ {
		dstRowOffset := y * dst.Stride
		for x := 0; x < width; x++ {
			var rAcc, gAcc, bAcc, aAcc float64

			for k := -radius; k <= radius; k++ {
				sampleY := mirrorIndex(y+k, height)
				weight := kernel[k+radius]
				srcOffset := sampleY*src.Stride + x*4

				rAcc += float64(src.Pix[srcOffset]) * weight
				gAcc += float64(src.Pix[srcOffset+1]) * weight
				bAcc += float64(src.Pix[srcOffset+2]) * weight
				aAcc += float64(src.Pix[srcOffset+3]) * weight
			}

			dstOffset := dstRowOffset + x*4
			dst.Pix[dstOffset] = clampToByte(rAcc)
			dst.Pix[dstOffset+1] = clampToByte(gAcc)
			dst.Pix[dstOffset+2] = clampToByte(bAcc)
			dst.Pix[dstOffset+3] = clampToByte(aAcc)
		}
	}
}

func mirrorIndex(index int, length int) int {
	if length <= 1 {
		return 0
	}
	for index < 0 || index >= length {
		if index < 0 {
			index = -index - 1
		} else {
			index = 2*length - index - 1
		}
	}
	return index
}

func clampToByte(value float64) uint8 {
	if value <= 0 {
		return 0
	}
	if value >= 255 {
		return 255
	}
	return uint8(value + 0.5)
}
