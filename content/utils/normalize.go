package utils

import (
	"avoid-the-enemies/content/config"
	"math"
)

func Normalize(x, y float64) (float64, float64) {
	t := math.Sqrt(
		math.Pow(x, 2) + math.Pow(y, 2),
	)

	return x * 100 / t, y * 100 / t
}

func Normal(x, y float64) (float64, float64) {
	t := math.Sqrt(
		math.Pow(x, 2) + math.Pow(y, 2),
	)

	return x / t, y / t
}

func ReNormalize(x, y float64) (float64, float64) {
	return x * config.ScreenWidth, y * config.ScreenHeight
}
