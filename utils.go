package gocaptcha

import (
	"image/color"
	"math/rand"
)

// RandDeepColor 随机生成深色系.
func RandDeepColor() color.RGBA {
	// 限制 RGB 最大值为 150 (深色系)，最小值为 50
	maxValue := 150
	minValue := 50

	r := uint8(rand.Intn(maxValue-minValue+1) + minValue)
	g := uint8(rand.Intn(maxValue-minValue+1) + minValue)
	b := uint8(rand.Intn(maxValue-minValue+1) + minValue)

	// Alpha 通道设置为完全不透明
	a := uint8(rand.Intn(256))

	return color.RGBA{R: r, G: g, B: b, A: a}
}

// RandLightColor 随机生成浅色.
func RandLightColor() color.RGBA {
	// 为每个颜色分量生成一个128到255之间的随机数
	red := rand.Intn(128) + 128
	green := rand.Intn(128) + 128
	blue := rand.Intn(128) + 128
	return color.RGBA{R: uint8(red), G: uint8(green), B: uint8(blue), A: uint8(255)}
}

// RandColor 生成随机颜色.
func RandColor() color.RGBA {
	red := rand.Intn(255)
	green := rand.Intn(255)
	var blue int

	// Calculate blue value based on the sum of red and green
	sum := red + green
	if sum > 400 {
		blue = 0
	} else {
		blueTemp := 400 - sum
		blue = int(max(0, min(255, float64(blueTemp))))
	}
	return color.RGBA{R: uint8(red), G: uint8(green), B: uint8(blue), A: 255}
}

// RandText 生成随机字体.
func RandText(num int) string {
	textNum := len(TextCharacters)
	text := make([]rune, num)
	for i := 0; i < num; i++ {
		text[i] = TextCharacters[rand.Intn(textNum)]
	}
	return string(text)
}

// ColorToRGB 颜色代码转换为RGB
// input int
// output int red, green, blue.
func ColorToRGB(colorVal int) color.RGBA {

	red := colorVal >> 16
	green := (colorVal & 0x00FF00) >> 8
	blue := colorVal & 0x0000FF

	return color.RGBA{
		R: uint8(red),
		G: uint8(green),
		B: uint8(blue),
		A: uint8(255),
	}
}

// 整数绝对值函数
func abs[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64](a T) T {
	var zero T
	if a < zero {
		return -a
	}
	return a
}
