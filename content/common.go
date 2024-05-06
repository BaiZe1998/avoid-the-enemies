package main

import (
	"avoid-the-enemies/content/config"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"math"
)

var (
	audioContext *audio.Context
	directions   = []struct {
		dx, dy, spin float64
	}{
		{1, 0, 0.0},              // 右
		{0, 1, math.Pi / 2},      // 下
		{-1, 0, math.Pi},         // 左
		{0, -1, math.Pi / 2 * 3}, // 上
	}
	rotateAdjust = []struct {
		dx, dy float64
	}{
		{0, 0},
		{1, 0},
		{1, 1},
		{0, 1},
	}
)

func IsTouch(x1, y1, x2, y2 float64) bool {
	return math.Abs(x1-x2) < config.FrameWidth/2 && math.Abs(y1-y2) < config.FrameHeight/2
}
