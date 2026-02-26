// Package gocaptcha 提供生成验证码图片的功能
// 本文件实现了各种线条绘制器，用于在验证码中添加线条元素
package gocaptcha

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
	"time"
)

// LineDrawer 线条绘制器接口
// 定义了绘制线条的方法
// 参数：
// - canvas: 要绘制的画布
// - x: 线条起点
// - y: 线条终点
// - color: 线条颜色
// 返回值：
// - error: 绘制过程中的错误
// 实现此接口的类型可以为验证码添加不同类型的线条

type LineDrawer interface {
	DrawLine(canvas draw.Image, x image.Point, y image.Point, color color.Color) error
}

// beeline 直线绘制器
// 实现了 LineDrawer 接口，用于绘制直线

type beeline struct {
}

// NewBeeline 创建一个新的直线绘制器
// 返回值：
// - LineDrawer: 直线绘制器实例
// 示例：
// ```go
// // 创建一个直线绘制器
// lineDrawer := gocaptcha.NewBeeline()
// // 为验证码添加直线
// captcha.DrawLine(lineDrawer, color.RGBA{0, 0, 0, 255})
// ```
func NewBeeline() LineDrawer {
	return &beeline{}
}

// DrawLine 绘制一条直线
// 参数：
// - canvas: 要绘制的画布
// - x1: 线条起点
// - y1: 线条终点
// - color: 线条颜色
// 返回值：
// - error: 绘制过程中的错误
// 影响：在画布上绘制一条从起点到终点的粗直线
func (beeline) DrawLine(canvas draw.Image, x1 image.Point, y1 image.Point, color color.Color) error {
	bounds := canvas.Bounds()
	maxX := bounds.Dx()
	maxY := bounds.Dy()

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

	// 绘制粗点函数（内联，避免闭包开销）
	drawThickPoint := func(cx, cy int) {
		for _, offset := range offsets {
			nx, ny := cx+offset.dx, cy+offset.dy
			// 边界检查
			if nx >= 0 && nx < maxX && ny >= 0 && ny < maxY {
				canvas.Set(nx, ny, color)
			}
		}
	}

	// 主循环，使用Bresenham直线算法
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

// curveLine 曲线绘制器
// 实现了 LineDrawer 接口，用于绘制基于正弦函数的曲线
// 字段：
// - r: 随机数生成器

type curveLine struct {
	r *rand.Rand
}

// DrawLine 绘制基于正弦函数的曲线
// 参数：
// - canvas: 要绘制的画布
// - x: 线条起点
// - y: 线条终点
// - cl: 线条颜色
// 返回值：
// - error: 绘制过程中的错误
// 影响：在画布上绘制一条基于正弦函数的曲线，增加验证码的复杂度
func (c curveLine) DrawLine(canvas draw.Image, x image.Point, y image.Point, cl color.Color) error {
	bounds := canvas.Bounds()
	maxX := bounds.Dx()
	maxY := bounds.Dy()
	px := 0
	var py float64

	// 振幅
	amplitude := c.r.Intn(maxY / 2)

	// Y轴方向偏移量
	offsetY := Random(int64(-maxY/4), int64(maxY/4))

	// X轴方向偏移量
	frequency := Random(int64(-maxY/4), int64(maxY/4))
	// 周期
	period := 0.0
	if maxY > maxX/2 {
		period = Random(int64(maxX/2), int64(maxY))
	} else {
		period = Random(int64(maxY), int64(maxX/2))
	}
	// 相位
	phase := (2 * math.Pi) / period

	// 曲线横坐标起始位置
	px1 := 0
	px2 := int(Random(int64(float64(maxX)*0.8), int64(maxX)))

	// 预计算常数
	yOffset := float64(maxY) / 5.0
	heightDiv5 := maxY / 5

	// 绘制曲线
	for px = px1; px < px2; px++ {
		if phase != 0 {
			py = float64(amplitude)*math.Sin(phase*float64(px)+frequency) + offsetY + yOffset
			// 边界检查
			if py >= 0 && py < float64(maxY) {
				pyInt := int(py)
				// 绘制粗曲线
				for i := 0; i < heightDiv5; i++ {
					ny := pyInt + i
					if ny >= 0 && ny < maxY && px >= 0 && px < maxX {
						canvas.Set(px+i, ny, cl)
					}
				}
			}
		}
	}
	return nil
}

// NewCurveLine 创建一个基于正弦函数的曲线绘制器
// 返回值：
// - LineDrawer: 曲线绘制器实例
// 示例：
// ```go
// // 创建一个曲线绘制器
// lineDrawer := gocaptcha.NewCurveLine()
// // 为验证码添加曲线
// captcha.DrawLine(lineDrawer, color.RGBA{0, 0, 0, 255})
// ```
func NewCurveLine() LineDrawer {
	return &curveLine{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// bezierLine 贝塞尔曲线绘制器
// 实现了 LineDrawer 接口，用于绘制贝塞尔曲线
// 字段：
// - r: 随机数生成器

type bezierLine struct {
	r *rand.Rand
}

// DrawLine 绘制贝塞尔曲线
// 参数：
// - canvas: 要绘制的画布
// - p0: 线条起点（作为贝塞尔曲线的第一个控制点）
// - p2: 线条终点（作为贝塞尔曲线的第三个控制点）
// - curveColor: 线条颜色
// 返回值：
// - error: 绘制过程中的错误
// 影响：在画布上绘制一条贝塞尔曲线，增加验证码的复杂度
func (b bezierLine) DrawLine(canvas draw.Image, p0 image.Point, p2 image.Point, curveColor color.Color) error {
	bounds := canvas.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	// 随机生成4个控制点
	// p0 由参数传入
	p1 := image.Point{X: b.r.Intn(width / 2), Y: b.r.Intn(height)}
	// p2 由参数传入
	p3 := image.Point{X: width - 1, Y: b.r.Intn(height)}

	// 预计算常量系数以提高性能
	// 绘制贝塞尔曲线，减少循环步长
	for t := 0.0; t <= 1.0; t += 0.01 {
		mt := 1 - t
		c0 := mt * mt * mt
		c1 := 3 * mt * mt * t
		c2 := 3 * mt * t * t
		c3 := t * t * t

		// 计算贝塞尔曲线上的点
		x := int(c0*float64(p0.X) + c1*float64(p1.X) + c2*float64(p2.X) + c3*float64(p3.X))
		y := int(c0*float64(p0.Y) + c1*float64(p1.Y) + c2*float64(p2.Y) + c3*float64(p3.Y))

		// 添加边界检查，防止越界
		if x >= 0 && x < width && y >= 0 && y < height {
			canvas.Set(x, y, curveColor)
		}
	}
	return nil
}

// NewBezierLine 创建一个贝塞尔曲线绘制器
// 返回值：
// - LineDrawer: 贝塞尔曲线绘制器实例
// 示例：
// ```go
// // 创建一个贝塞尔曲线绘制器
// lineDrawer := gocaptcha.NewBezierLine()
// // 为验证码添加贝塞尔曲线
// captcha.DrawLine(lineDrawer, color.RGBA{0, 0, 0, 255})
// ```
func NewBezierLine() LineDrawer {
	return &bezierLine{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// bezier3DLine 3D效果贝塞尔曲线绘制器
// 实现了 LineDrawer 接口，用于绘制带有3D效果的贝塞尔曲线
// 字段：
// - r: 随机数生成器

type bezier3DLine struct {
	r *rand.Rand
}

// DrawLine 绘制3D效果的贝塞尔曲线
// 参数：
// - canvas: 要绘制的画布
// - p0: 线条起点（作为贝塞尔曲线的第一个控制点）
// - p2: 线条终点（作为贝塞尔曲线的第三个控制点）
// - cl: 线条颜色
// 返回值：
// - error: 绘制过程中的错误
// 影响：在画布上绘制一条带有3D效果的贝塞尔曲线，增加验证码的复杂度和视觉效果
func (b bezier3DLine) DrawLine(canvas draw.Image, p0 image.Point, p2 image.Point, cl color.Color) error {
	bounds := canvas.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	// 随机生成4个控制点
	// p0 由参数传入
	p1 := image.Point{X: b.r.Intn(width / 2), Y: b.r.Intn(height)}
	// p2 由参数传入
	p3 := image.Point{X: width - 1, Y: b.r.Intn(height)}

	// 绘制带宽度的点的函数
	drawPointWithWidth := func(img draw.Image, x, y int, col color.Color, width int) {
		// 优化：只在必要时执行边界检查
		minX := x - width
		maxX := x + width
		minY := y - width
		maxY := y + width

		canvasBounds := img.Bounds()
		canvasMaxX := canvasBounds.Dx()
		canvasMaxY := canvasBounds.Dy()

		// 快速越界检查
		if maxX < 0 || minX >= canvasMaxX || maxY < 0 || minY >= canvasMaxY {
			return
		}

		for dx := -width; dx <= width; dx++ {
			for dy := -width; dy <= width; dy++ {
				// 确保点在圆形范围内
				if dx*dx+dy*dy <= width*width {
					nx, ny := x+dx, y+dy
					// 边界检查
					if nx >= 0 && nx < canvasMaxX && ny >= 0 && ny < canvasMaxY {
						img.Set(nx, ny, col)
					}
				}
			}
		}
	}

	w := float64(b.r.Intn(height / 5))
	// 绘制贝塞尔曲线，模拟3D效果，减少循环次数
	for t := 0.0; t <= 1.0; t += 0.01 {
		// 计算当前点的坐标（预计算系数以优化性能）
		mt := 1 - t
		c0 := mt * mt * mt
		c1 := 3 * mt * mt * t
		c2 := 3 * mt * t * t
		c3 := t * t * t

		x := int(c0*float64(p0.X) + c1*float64(p1.X) + c2*float64(p2.X) + c3*float64(p3.X))
		y := int(c0*float64(p0.Y) + c1*float64(p1.Y) + c2*float64(p2.Y) + c3*float64(p3.Y))

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

// NewBezier3DLine 创建一个3D效果的贝塞尔曲线绘制器
// 返回值：
// - LineDrawer: 3D效果贝塞尔曲线绘制器实例
// 示例：
// ```go
// // 创建一个3D效果贝塞尔曲线绘制器
// lineDrawer := gocaptcha.NewBezier3DLine()
// // 为验证码添加3D效果贝塞尔曲线
// captcha.DrawLine(lineDrawer, color.RGBA{0, 0, 0, 255})
// ```
func NewBezier3DLine() LineDrawer {
	return &bezier3DLine{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// hollowLine 空心线绘制器
// 实现了 LineDrawer 接口，用于绘制空心线
// 字段：
// - r: 随机数生成器

type hollowLine struct {
	r *rand.Rand
}

// DrawLine 绘制空心线
// 参数：
// - canvas: 要绘制的画布
// - p0: 线条起点
// - p1: 线条终点
// - lineColor: 线条颜色
// 返回值：
// - error: 绘制过程中的错误
// 影响：在画布上绘制一条空心线，增加验证码的复杂度
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

	// 绘制空心线
	for ; x1 < x2; x1++ {
		y := math.Sin(x1*math.Pi*multiple/float64(width)) * float64(height/3)

		if multiple < 0 {
			y = y + float64(height/2)
		}

		// 确保 y 在边界内
		y = math.Max(0, math.Min(float64(height-1), y))
		py := int(y)

		// 绘制线条
		for i := 0; i <= w; i++ {
			ny := py + i
			// 添加边界检查
			if ny >= 0 && ny < height && int(x1) >= 0 && int(x1) < width {
				canvas.Set(int(x1), ny, lineColor)
			}
		}
	}
	return nil
}

// NewHollowLine 创建一个空心线绘制器
// 返回值：
// - LineDrawer: 空心线绘制器实例
// 示例：
// ```go
// // 创建一个空心线绘制器
// lineDrawer := gocaptcha.NewHollowLine()
// // 为验证码添加空心线
// captcha.DrawLine(lineDrawer, color.RGBA{0, 0, 0, 255})
// ```
func NewHollowLine() LineDrawer {
	return &hollowLine{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
