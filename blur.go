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
	// Validate parameters to prevent panic
	if kernelSize <= 0 || kernelSize%2 == 0 {
		// Ensure kernel size is positive and odd
		kernelSize = 5 // default to 5x5 kernel
	}
	if sigma <= 0 {
		sigma = 1.0 // default sigma
	}

	kernel := g.generateGaussianKernel(kernelSize, sigma)
	bounds := canvas.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, grn, b := g.applyKernel(canvas, x, y, kernel)
			canvas.Set(x, y, color.RGBA{R: r, G: grn, B: b, A: 255})
		}
	}
	return nil
}

func (g *gaussianBlur) generateGaussianKernel(kernelSize int, sigma float64) [][]float64 {
	// Guard against invalid parameters to prevent division by zero
	if sigma <= 0 {
		sigma = 1.0
	}
	if kernelSize <= 0 {
		kernelSize = 5
	}

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

	// Normalize kernel to ensure sum equals 1.0
	if sum > 0 {
		for i := 0; i < kernelSize; i++ {
			for j := 0; j < kernelSize; j++ {
				kernel[i][j] /= sum
			}
		}
	}

	return kernel
}

func (g *gaussianBlur) applyKernel(canvas draw.Image, x int, y int, kernel [][]float64) (uint8, uint8, uint8) {
	bounds := canvas.Bounds()
	size := len(kernel)
	mid := size / 2
	var rSum, gSum, bSum float64
	var weightSum float64

	for ky := 0; ky < size; ky++ {
		for kx := 0; kx < size; kx++ {
			px := x + kx - mid
			py := y + ky - mid
			if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
				// RGBA returns premultiplied alpha in 0-65535 range
				rr, gg, bb, aa := canvas.At(px, py).RGBA()
				k := kernel[ky][kx]
				// Handle alpha channel: normalize by alpha for proper transparency
				alpha := float64(aa) / 65535.0
				weight := k * alpha
				weightSum += weight
				// Convert from 16-bit to 8-bit (0-255 range)
				rSum += weight * float64(rr>>8)
				gSum += weight * float64(gg>>8)
				bSum += weight * float64(bb>>8)
			} else {
				// For out-of-bounds, use edge padding (lower weight)
				k := kernel[ky][kx]
				weightSum += k * 0.5
			}
		}
	}

	// Normalize by weight sum to handle transparency and edge artifacts
	if weightSum > 0 {
		rSum /= weightSum
		gSum /= weightSum
		bSum /= weightSum
	}

	return g.clamp(rSum), g.clamp(gSum), g.clamp(bSum)
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
