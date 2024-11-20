package gocaptcha

import (
	"fmt"
	"math/rand"
)

// Random 生成指定大小的随机数.
func Random(min int64, max int64) float64 {
	if max <= min {
		panic(fmt.Sprintf("invalid range %d <= %d", max, min)) // 修复了错误消息的顺序
	}

	rangeSize := max - min
	var randomValue int64

	if rangeSize > 0 {
		randomValue = rand.Int63n(rangeSize) + min // 确保在[min, max)范围内
	} else { // 处理负数范围
		randomValue = rand.Int63n(-rangeSize) + min // 确保在[min, max)范围内，此时max是更小的负数
	}

	return float64(randomValue) + rand.Float64() // 添加小数部分
}
