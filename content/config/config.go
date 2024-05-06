package config

type Mode int

const (
	ModeTitle Mode = iota
	ModeGame
	ModeGameOver
)

const (
	ScreenWidth   = 320
	ScreenHeight  = 240
	FrameOX       = 0
	FrameOY       = 32
	FrameWidth    = 32
	FrameHeight   = 32
	FrameCount    = 8
	TitleFontSize = FontSize * 1.5
	FontSize      = 8
)

const (
	MonsterMinDistance = 10 // 怪物之间的最小距离，当两个怪物的距离大于此值，它们将趋于分离
)
