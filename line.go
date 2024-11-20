package gocaptcha

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
	"time"
)

// LineDrawer 实现划线的接口
type LineDrawer interface {
	DrawLine(canvas draw.Image, x image.Point, y image.Point, color color.Color) error
}

type beeline struct {
}

func NewBeeline() LineDrawer {
	return &beeline{}
}

// DrawLine 画一条直线
func (beeline) DrawLine(canvas draw.Image, x1 image.Point, y1 image.Point, color color.Color) error {
	dx := abs(x1.X - y1.X)
	dy := abs(y1.Y - x1.Y)

	sx, sy := 1, 1
	if x1.X >= y1.X {
		sx = -1
	}
	if x1.Y >= y1.Y {
		sy = -1
	}
	err := dx - dy
	x, y := x1.X, x1.Y

	// 预定义粗线的相对偏移
	offsets := []struct{ dx, dy int }{
		{-1, -1}, {0, -1}, {1, -1},
		{-1, 0}, {0, 0}, {1, 0},
		{-1, 1}, {0, 1}, {1, 1},
	}

	// 边界检查函数
	isValidPoint := func(cx, cy int) bool {
		return cx >= 0 && cx < canvas.Bounds().Dx() && cy >= 0 && cy < canvas.Bounds().Dy()
	}

	// 绘制粗点函数
	drawThickPoint := func(cx, cy int) {
		for _, offset := range offsets {
			nx, ny := cx+offset.dx, cy+offset.dy
			if isValidPoint(nx, ny) {
				canvas.Set(nx, ny, color)
			}
		}
	}

	// 主循环
	for {
		drawThickPoint(x, y) // 绘制粗点
		if x == y1.X && y == y1.Y {
			break // 到达终点，退出循环
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
	return nil
}

type curveLine struct {
	r *rand.Rand
}

func (c curveLine) DrawLine(canvas draw.Image, x image.Point, y image.Point, cl color.Color) error {
	px := 0
	var py float64 = 0

	//振幅
	amplitude := c.r.Intn(canvas.Bounds().Dy() / 2)

	//Y轴方向偏移量
	b := Random(int64(-canvas.Bounds().Dy()/4), int64(canvas.Bounds().Dy()/4))

	//X轴方向偏移量
	frequency := Random(int64(-canvas.Bounds().Dy()/4), int64(canvas.Bounds().Dy()/4))
	// 周期
	var t float64 = 0
	if canvas.Bounds().Dy() > canvas.Bounds().Dx()/2 {
		t = Random(int64(canvas.Bounds().Dx()/2), int64(canvas.Bounds().Dy()))
	} else {
		t = Random(int64(canvas.Bounds().Dy()), int64(canvas.Bounds().Dx()/2))
	}
	// 相位
	phase := (2 * math.Pi) / t

	// 曲线横坐标起始位置
	px1 := 0
	px2 := int(Random(int64(float64(canvas.Bounds().Dx())*0.8), int64(canvas.Bounds().Dx())))

	for px = px1; px < px2; px++ {
		if phase != 0 {
			py = float64(amplitude)*math.Sin(phase*float64(px)+frequency) + b + (float64(canvas.Bounds().Dx()) / float64(5))
			i := canvas.Bounds().Dy() / 5
			for i > 0 {
				canvas.Set(px+i, int(py), cl)
				i--
			}
		}
	}
	return nil
}

// NewCurveLine 基于正弦函数的曲线
func NewCurveLine() LineDrawer {
	return &curveLine{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

type bezierLine struct {
	r *rand.Rand
}

func (b bezierLine) DrawLine(canvas draw.Image, p0 image.Point, p2 image.Point, curveColor color.Color) error {
	width := canvas.Bounds().Dx()
	height := canvas.Bounds().Dy()
	// 随机生成4个控制点
	//p0 := image.Point{X: rand.Intn(width / 4), Y: rand.Intn(height)}
	p1 := image.Point{X: b.r.Intn(width / 2), Y: b.r.Intn(height)}
	//p2 := image.Point{X: width/2 + rand.Intn(width/4), Y: rand.Intn(height)}
	p3 := image.Point{X: width - 1, Y: b.r.Intn(height)}

	// 绘制贝塞尔曲线
	for t := 0.0; t <= 1.0; t += 0.001 {
		x := int((1-t)*(1-t)*(1-t)*float64(p0.X) + 3*(1-t)*(1-t)*t*float64(p1.X) + 3*(1-t)*t*t*float64(p2.X) + t*t*t*float64(p3.X))
		y := int((1-t)*(1-t)*(1-t)*float64(p0.Y) + 3*(1-t)*(1-t)*t*float64(p1.Y) + 3*(1-t)*t*t*float64(p2.Y) + t*t*t*float64(p3.Y))
		canvas.Set(x, y, curveColor)
	}
	return nil
}

// NewBezierLine 贝塞尔曲线
func NewBezierLine() LineDrawer {
	return &bezierLine{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

type bezier3DLine struct {
	r *rand.Rand
}

// DrawLine 绘制3D效果的贝塞尔曲线
func (b bezier3DLine) DrawLine(canvas draw.Image, p0 image.Point, p2 image.Point, cl color.Color) error {
	width := canvas.Bounds().Dx()
	height := canvas.Bounds().Dy()
	// 随机生成4个控制点
	//p0 := image.Point{X: b.r.Intn(width / 4), Y: b.r.Intn(height)}
	p1 := image.Point{X: b.r.Intn(width / 2), Y: b.r.Intn(height)}
	//p2 := image.Point{X: width/2 + b.r.Intn(width/4), Y: b.r.Intn(height)}
	p3 := image.Point{X: width - 1, Y: b.r.Intn(height)}

	drawPointWithWidth := func(img draw.Image, x, y int, col color.Color, width int) {
		for dx := -width; dx <= width; dx++ {
			for dy := -width; dy <= width; dy++ {
				// 确保点在圆形范围内
				if dx*dx+dy*dy <= width*width {
					img.Set(x+dx, y+dy, col)
				}
			}
		}
	}
	w := float64(b.r.Intn(height / 5))
	// 绘制贝塞尔曲线，模拟3D效果
	for t := 0.0; t <= 1.0; t += 0.001 {
		// 计算当前点的坐标
		x := int((1-t)*(1-t)*(1-t)*float64(p0.X) + 3*(1-t)*(1-t)*t*float64(p1.X) + 3*(1-t)*t*t*float64(p2.X) + t*t*t*float64(p3.X))
		y := int((1-t)*(1-t)*(1-t)*float64(p0.Y) + 3*(1-t)*(1-t)*t*float64(p1.Y) + 3*(1-t)*t*t*float64(p2.Y) + t*t*t*float64(p3.Y))

		// 使用 t 值调整颜色和线宽，模拟3D效果
		opacity := uint8(255 - int(t*255)) // 透明度渐变
		lineColor := color.NRGBA{
			R: uint8(250 * t), // 红色分量随 t 增加
			G: uint8(128 * (1 - t)),
			B: 255 - uint8(128*t),
			A: opacity,
		}

		// 模拟线宽，绘制当前点周围的像素
		lineWidth := int(w * (1 - t)) // 线宽随 t 减小
		drawPointWithWidth(canvas, x, y, lineColor, lineWidth)
	}
	return nil
}

// NewBezier3DLine 3D效果的贝塞尔曲线
func NewBezier3DLine() LineDrawer {
	return &bezier3DLine{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

type hollowLine struct {
	r *rand.Rand
}

// DrawLine 绘制空心线
func (h hollowLine) DrawLine(canvas draw.Image, p0 image.Point, p1 image.Point, lineColor color.Color) error {
	bounds := canvas.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	x1 := float64(p0.X)
	x2 := float64(p1.X)

	multiple := float64(h.r.Intn(5)+3) / 5.0
	if int(multiple*10)%3 == 0 {
		multiple = multiple * -1.0
	}

	w := width / 20

	for ; x1 < x2; x1++ {
		y := math.Sin(x1*math.Pi*multiple/float64(width)) * float64(height/3)

		if multiple < 0 {
			y = y + float64(height/2)
		}

		// Ensure y is within bounds
		y = math.Max(0, math.Min(float64(height-1), y))

		for i := 0; i <= w && int(y)+i < height; i++ {
			canvas.Set(int(x1), int(y)+i, lineColor)
		}
	}
	return nil
}

// NewHollowLine 空心线
func NewHollowLine() LineDrawer {
	return &hollowLine{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
