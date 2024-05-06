package utils

import "avoid-the-enemies/content/config"

func Normalize(x, y float64) (float64, float64) {
	return x / config.ScreenWidth, y / config.ScreenHeight
}

func ReNormalize(x, y float64) (float64, float64) {
	return x * config.ScreenWidth, y * config.ScreenHeight
}
