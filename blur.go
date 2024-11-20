package gocaptcha

import (
	"image/color"
	"image/draw"
	"math"
)

type BlurDrawer interface {
	DrawBlur(canvas draw.Image, kernelSize int, sigma float64) error
}

type gaussianBlur struct {
}

func NewGaussianBlur() BlurDrawer {
	return &gaussianBlur{}
}

func (g *gaussianBlur) DrawBlur(canvas draw.Image, kernelSize int, sigma float64) error {
	kernel := g.generateGaussianKernel(kernelSize, sigma)
	bounds := canvas.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b := g.applyKernel(canvas, x, y, kernel)
			canvas.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return nil
}

func (g *gaussianBlur) generateGaussianKernel(kernelSize int, sigma float64) [][]float64 {
	kernel := make([][]float64, kernelSize)
	sum := 0.0
	mid := kernelSize / 2

	for i := 0; i < kernelSize; i++ {
		kernel[i] = make([]float64, kernelSize)
		for j := 0; j < kernelSize; j++ {
			x := float64(i - mid)
			y := float64(j - mid)
			kernel[i][j] = math.Exp(-(x*x+y*y)/(2*sigma*sigma)) / (2 * math.Pi * sigma * sigma)
			sum += kernel[i][j]
		}
	}

	// Normalize kernel
	for i := 0; i < kernelSize; i++ {
		for j := 0; j < kernelSize; j++ {
			kernel[i][j] /= sum
		}
	}

	return kernel
}

func (g *gaussianBlur) applyKernel(canvas draw.Image, x int, y int, kernel [][]float64) (uint8, uint8, uint8) {
	bounds := canvas.Bounds()
	size := len(kernel)
	mid := size / 2
	var r1, g1, b1 float64

	for ky := 0; ky < size; ky++ {
		for kx := 0; kx < size; kx++ {
			px := x + kx - mid
			py := y + ky - mid
			if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
				rr, gg, bb, _ := canvas.At(px, py).RGBA()
				k := kernel[ky][kx]
				r1 += k * float64(rr>>8)
				g1 += k * float64(gg>>8)
				b1 += k * float64(bb>>8)
			}
		}
	}

	return g.clamp(r1), g.clamp(g1), g.clamp(b1)
}

func (g *gaussianBlur) clamp(value float64) uint8 {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return uint8(value)
}
