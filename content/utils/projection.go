package utils

// 计算第3点在直线上的投影的位置
func GetProjection(x1, y1, x2, y2, x, y float64) (float64, float64) {
	k := (y2 - y1) / (x2 - x1)
	b := y1 - k*x1
	k1 := -1 / k
	b1 := y - k1*x
	x3 := (b1 - b) / (k - k1)
	y3 := k*x3 + b
	return x3, y3
}
