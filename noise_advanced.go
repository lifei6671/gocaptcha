package gocaptcha

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
	"sync"
)

const (
	defaultPoissonAttempts = 20
	defaultPerlinOctaves   = 3
	sqrtHalf               = 0.7071067811865476
)

type poissonPoint struct {
	x float64
	y float64
}

type poissonPointNoiseDrawer struct {
	r           *rand.Rand
	randMu      sync.Mutex
	randOnce    sync.Once
	minDistance float64
	attempts    int
	colorFn     NoiseColorFunc
}

// NewPoissonPointNoiseDrawer creates a Poisson disk sampling noise drawer.
// It draws stroke-like clutter to better resist OCR denoising/segmentation.
func NewPoissonPointNoiseDrawer() NoiseDrawer {
	return &poissonPointNoiseDrawer{
		r: newSecureSeededRand(),
	}
}

// NewPoissonPointNoiseDrawerWithConfig creates a Poisson drawer with custom settings.
func NewPoissonPointNoiseDrawerWithConfig(minDistance float64, attempts int, colorFn NoiseColorFunc) NoiseDrawer {
	return &poissonPointNoiseDrawer{
		r:           newSecureSeededRand(),
		minDistance: minDistance,
		attempts:    attempts,
		colorFn:     colorFn,
	}
}

func (n *poissonPointNoiseDrawer) DrawNoise(img draw.Image, density NoiseDensity) error {
	if img == nil {
		return ErrNilCanvas
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil
	}

	minDistance := n.minDistance
	if minDistance <= 0 || math.IsNaN(minDistance) || math.IsInf(minDistance, 0) {
		minDistance = poissonMinDistanceForDensity(density)
	}
	if minDistance < 1 {
		minDistance = 1
	}

	attempts := n.attempts
	if attempts <= 0 {
		attempts = defaultPoissonAttempts
	}

	n.withRand(func(r *rand.Rand) {
		points := generatePoissonPoints(width, height, minDistance, attempts, r)
		for _, point := range points {
			x := bounds.Min.X + int(point.x)
			y := bounds.Min.Y + int(point.y)
			if !image.Pt(x, y).In(bounds) {
				continue
			}

			strokeLength, strokeWidth := poissonStrokeProfile(density, r)
			primaryAngle := r.Float64() * 2 * math.Pi
			noiseColor := ocrContrastNoiseColor(img, x, y, r, n.colorFn)

			drawOrientedStroke(img, bounds, x, y, strokeLength, strokeWidth, primaryAngle, noiseColor)
			if density == NoiseDensityHigh && r.Float64() < 0.35 {
				drawOrientedStroke(img, bounds, x, y, maxInt(2, strokeLength-1), strokeWidth, primaryAngle+math.Pi/2+r.Float64()*0.4, noiseColor)
			}
			if r.Float64() < 0.20 {
				drawNoiseBlob(img, bounds, x, y, maxInt(1, strokeWidth-1), noiseColor)
			}
		}
	})

	return nil
}

func (n *poissonPointNoiseDrawer) withRand(fn func(r *rand.Rand)) {
	withDrawerRand(&n.r, &n.randOnce, &n.randMu, fn)
}

func poissonMinDistanceForDensity(density NoiseDensity) float64 {
	switch density {
	case NoiseDensityLower:
		return 9
	case NoiseDensityMedium:
		return 7
	case NoiseDensityHigh:
		return 5
	default:
		return 7
	}
}

func poissonStrokeProfile(density NoiseDensity, r *rand.Rand) (int, int) {
	switch density {
	case NoiseDensityLower:
		return 3 + randIntn(r, 3), 1
	case NoiseDensityMedium:
		return 4 + randIntn(r, 4), 1 + randIntn(r, 2)
	case NoiseDensityHigh:
		return 5 + randIntn(r, 5), 2
	default:
		return 4 + randIntn(r, 4), 1
	}
}

func generatePoissonPoints(width int, height int, minDistance float64, attempts int, r *rand.Rand) []poissonPoint {
	cellSize := minDistance / math.Sqrt2
	if cellSize <= 0 {
		cellSize = 1
	}

	gridW := int(math.Ceil(float64(width) / cellSize))
	gridH := int(math.Ceil(float64(height) / cellSize))
	grid := make([]int, gridW*gridH)
	for i := range grid {
		grid[i] = -1
	}

	points := make([]poissonPoint, 0, maxInt(16, (width*height)/maxInt(1, int(minDistance*minDistance))))
	active := make([]int, 0, 32)

	insertPoint := func(p poissonPoint) {
		index := len(points)
		points = append(points, p)
		gx := int(p.x / cellSize)
		gy := int(p.y / cellSize)
		if gx >= 0 && gx < gridW && gy >= 0 && gy < gridH {
			grid[gy*gridW+gx] = index
		}
		active = append(active, index)
	}

	isValid := func(p poissonPoint) bool {
		if p.x < 0 || p.x >= float64(width) || p.y < 0 || p.y >= float64(height) {
			return false
		}
		gx := int(p.x / cellSize)
		gy := int(p.y / cellSize)

		xMin := maxInt(0, gx-2)
		xMax := min(gridW-1, gx+2)
		yMin := maxInt(0, gy-2)
		yMax := min(gridH-1, gy+2)
		minDistanceSquared := minDistance * minDistance

		for yy := yMin; yy <= yMax; yy++ {
			for xx := xMin; xx <= xMax; xx++ {
				pointIndex := grid[yy*gridW+xx]
				if pointIndex < 0 {
					continue
				}
				dx := p.x - points[pointIndex].x
				dy := p.y - points[pointIndex].y
				if dx*dx+dy*dy < minDistanceSquared {
					return false
				}
			}
		}
		return true
	}

	insertPoint(poissonPoint{
		x: r.Float64() * float64(width),
		y: r.Float64() * float64(height),
	})

	for len(active) > 0 {
		activeIndex := randIntn(r, len(active))
		pointIndex := active[activeIndex]
		basePoint := points[pointIndex]

		found := false
		for i := 0; i < attempts; i++ {
			angle := r.Float64() * 2 * math.Pi
			radius := minDistance * (1 + r.Float64())
			candidate := poissonPoint{
				x: basePoint.x + math.Cos(angle)*radius,
				y: basePoint.y + math.Sin(angle)*radius,
			}
			if isValid(candidate) {
				insertPoint(candidate)
				found = true
				break
			}
		}

		if !found {
			last := len(active) - 1
			active[activeIndex] = active[last]
			active = active[:last]
		}
	}

	return points
}

type perlinNoiseDrawer struct {
	r         *rand.Rand
	randMu    sync.Mutex
	randOnce  sync.Once
	scale     float64
	octaves   int
	threshold float64
	colorFn   NoiseColorFunc
}

// NewPerlinNoiseDrawer creates a Perlin-style coherent noise drawer.
// It uses warped ridge + flow-aligned strokes for stronger OCR confusion.
func NewPerlinNoiseDrawer() NoiseDrawer {
	return &perlinNoiseDrawer{
		r: newSecureSeededRand(),
	}
}

// NewPerlinNoiseDrawerWithConfig creates a Perlin drawer with custom settings.
func NewPerlinNoiseDrawerWithConfig(scale float64, octaves int, threshold float64, colorFn NoiseColorFunc) NoiseDrawer {
	return &perlinNoiseDrawer{
		r:         newSecureSeededRand(),
		scale:     scale,
		octaves:   octaves,
		threshold: threshold,
		colorFn:   colorFn,
	}
}

func (n *perlinNoiseDrawer) DrawNoise(img draw.Image, density NoiseDensity) error {
	if img == nil {
		return ErrNilCanvas
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil
	}

	scale := n.scale
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = perlinScaleForDensity(density)
	}
	if scale < 1 {
		scale = 1
	}

	octaves := n.octaves
	if octaves <= 0 {
		octaves = defaultPerlinOctaves
	}

	threshold := n.threshold
	if threshold <= 0 || threshold >= 1 || math.IsNaN(threshold) || math.IsInf(threshold, 0) {
		threshold = perlinThresholdForDensity(density)
	}

	step := perlinStepForDensity(density)

	n.withRand(func(r *rand.Rand) {
		seed := r.Int63()
		for y := 0; y < height; y += step {
			for x := 0; x < width; x += step {
				nx := float64(x) / scale
				ny := float64(y) / scale
				ridge := perlinWarpedRidge(nx, ny, octaves, seed)
				if ridge < threshold {
					continue
				}

				flow := perlinFBM(nx*1.2+31.7, ny*1.2-17.3, 2, seed+73421)
				angle := flow*2*math.Pi + (ridge-0.5)*0.5
				strokeLength, strokeWidth := perlinStrokeProfile(density, ridge, r)
				px := bounds.Min.X + x
				py := bounds.Min.Y + y
				noiseColor := ocrContrastNoiseColor(img, px, py, r, n.colorFn)

				drawOrientedStroke(img, bounds, px, py, strokeLength, strokeWidth, angle, noiseColor)
				if density == NoiseDensityHigh && ridge > threshold+0.08 && r.Float64() < 0.25 {
					drawOrientedStroke(img, bounds, px, py, maxInt(2, strokeLength-1), strokeWidth, angle+math.Pi/2, noiseColor)
				}
			}
		}
	})

	return nil
}

func (n *perlinNoiseDrawer) withRand(fn func(r *rand.Rand)) {
	withDrawerRand(&n.r, &n.randOnce, &n.randMu, fn)
}

func perlinScaleForDensity(density NoiseDensity) float64 {
	switch density {
	case NoiseDensityLower:
		return 30
	case NoiseDensityMedium:
		return 22
	case NoiseDensityHigh:
		return 16
	default:
		return 22
	}
}

func perlinThresholdForDensity(density NoiseDensity) float64 {
	switch density {
	case NoiseDensityLower:
		return 0.78
	case NoiseDensityMedium:
		return 0.72
	case NoiseDensityHigh:
		return 0.66
	default:
		return 0.72
	}
}

func perlinStepForDensity(density NoiseDensity) int {
	switch density {
	case NoiseDensityLower:
		return 3
	case NoiseDensityMedium:
		return 2
	case NoiseDensityHigh:
		return 1
	default:
		return 2
	}
}

func perlinWarpedRidge(x float64, y float64, octaves int, seed int64) float64 {
	warpX := perlinFBM(x*0.6+17.7, y*0.6-5.1, 2, seed+101)
	warpY := perlinFBM(x*0.6-9.3, y*0.6+11.4, 2, seed+809)
	nx := x + (warpX-0.5)*1.8
	ny := y + (warpY-0.5)*1.8

	base := perlinFBM(nx, ny, octaves, seed)
	ridge := 1 - math.Abs(2*base-1)
	if ridge < 0 {
		return 0
	}
	if ridge > 1 {
		return 1
	}
	return ridge
}

func perlinStrokeProfile(density NoiseDensity, ridge float64, r *rand.Rand) (int, int) {
	base := 2 + int(ridge*4)
	switch density {
	case NoiseDensityLower:
		return base + randIntn(r, 2), 1
	case NoiseDensityMedium:
		return base + 1 + randIntn(r, 2), 1 + randIntn(r, 2)
	case NoiseDensityHigh:
		return base + 2 + randIntn(r, 3), 2
	default:
		return base + 1, 1
	}
}

func ocrContrastNoiseColor(img draw.Image, x int, y int, r *rand.Rand, override NoiseColorFunc) color.Color {
	if override != nil {
		return override(r)
	}

	rr, gg, bb, _ := img.At(x, y).RGBA()
	luminance := 0.2126*float64(rr) + 0.7152*float64(gg) + 0.0722*float64(bb)
	if luminance > 32768 {
		if r.Float64() < 0.8 {
			return randDeepNoiseColor(r)
		}
		return randColorFromRand(r)
	}

	if r.Float64() < 0.8 {
		return randLightColorFromRand(r)
	}
	return randColorFromRand(r)
}

func drawOrientedStroke(img draw.Image, bounds image.Rectangle, cx int, cy int, length int, thickness int, angle float64, col color.Color) {
	if length <= 0 {
		length = 1
	}
	if thickness <= 0 {
		thickness = 1
	}

	half := float64(length) / 2.0
	dx := int(math.Round(math.Cos(angle) * half))
	dy := int(math.Round(math.Sin(angle) * half))
	x0 := cx - dx
	y0 := cy - dy
	x1 := cx + dx
	y1 := cy + dy
	drawThickLine(img, bounds, x0, y0, x1, y1, thickness, col)
}

func drawThickLine(img draw.Image, bounds image.Rectangle, x0 int, y0 int, x1 int, y1 int, thickness int, col color.Color) {
	if thickness <= 0 {
		thickness = 1
	}
	radius := maxInt(0, thickness-1)

	dx := absInt(x1 - x0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	dy := -absInt(y1 - y0)
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy

	for {
		drawNoiseBlob(img, bounds, x0, y0, radius, col)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func drawNoiseBlob(img draw.Image, bounds image.Rectangle, cx int, cy int, radius int, col color.Color) {
	if radius <= 0 {
		if image.Pt(cx, cy).In(bounds) {
			img.Set(cx, cy, col)
		}
		return
	}

	r2 := radius * radius
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy > r2 {
				continue
			}
			x := cx + dx
			y := cy + dy
			if image.Pt(x, y).In(bounds) {
				img.Set(x, y, col)
			}
		}
	}
}

func perlinFBM(x float64, y float64, octaves int, seed int64) float64 {
	amplitude := 1.0
	frequency := 1.0
	total := 0.0
	normalization := 0.0

	for i := 0; i < octaves; i++ {
		total += amplitude * perlin2D(x*frequency, y*frequency, seed+int64(i*9973))
		normalization += amplitude
		amplitude *= 0.5
		frequency *= 2
	}

	if normalization <= 0 {
		return 0.5
	}
	value := total / normalization
	normalized := (value + 1) / 2
	if normalized < 0 {
		return 0
	}
	if normalized > 1 {
		return 1
	}
	return normalized
}

func perlin2D(x float64, y float64, seed int64) float64 {
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := x0 + 1
	y1 := y0 + 1

	sx := x - float64(x0)
	sy := y - float64(y0)

	n00 := dotGridGradient(x0, y0, x, y, seed)
	n10 := dotGridGradient(x1, y0, x, y, seed)
	n01 := dotGridGradient(x0, y1, x, y, seed)
	n11 := dotGridGradient(x1, y1, x, y, seed)

	u := perlinFade(sx)
	v := perlinFade(sy)

	ix0 := perlinLerp(n00, n10, u)
	ix1 := perlinLerp(n01, n11, u)
	return perlinLerp(ix0, ix1, v)
}

func dotGridGradient(ix int, iy int, x float64, y float64, seed int64) float64 {
	gx, gy := gradientFromHash(hash2D(ix, iy, seed))
	dx := x - float64(ix)
	dy := y - float64(iy)
	return dx*gx + dy*gy
}

func gradientFromHash(hash uint64) (float64, float64) {
	switch hash & 7 {
	case 0:
		return 1, 0
	case 1:
		return -1, 0
	case 2:
		return 0, 1
	case 3:
		return 0, -1
	case 4:
		return sqrtHalf, sqrtHalf
	case 5:
		return -sqrtHalf, sqrtHalf
	case 6:
		return sqrtHalf, -sqrtHalf
	default:
		return -sqrtHalf, -sqrtHalf
	}
}

func hash2D(x int, y int, seed int64) uint64 {
	v := uint64(uint32(x))*0x9e3779b185ebca87 ^ uint64(uint32(y))*0xc2b2ae3d27d4eb4f ^ uint64(seed)
	v ^= v >> 33
	v *= 0xff51afd7ed558ccd
	v ^= v >> 33
	v *= 0xc4ceb9fe1a85ec53
	v ^= v >> 33
	return v
}

func perlinFade(t float64) float64 {
	return t * t * t * (t*(t*6-15) + 10)
}

func perlinLerp(a float64, b float64, t float64) float64 {
	return a + t*(b-a)
}
