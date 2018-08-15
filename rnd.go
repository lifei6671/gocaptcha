package gocaptcha

import (
	"fmt"
	"math/rand"
)

//生成指定大小的随机数.
func Random(min int64, max int64) float64 {

	if max <= min {
		panic(fmt.Sprintf("invalid range %d >= %d", max, min))
	}
	decimal := rand.Float64()

	if max <= 0 {
		return (float64(rand.Int63n((min * -1)-(max * -1))+(max * -1)) + decimal) * -1
	}
	if min < 0 && max > 0 {
		if rand.Int()%2 == 0 {
			return float64(rand.Int63n(max)) + decimal
		} else {
			return (float64(rand.Int63n(min * -1)) + decimal) * -1
		}
	}
	return float64(rand.Int63n(max-min)+min) + decimal
}
