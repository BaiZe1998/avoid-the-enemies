package main

type Mode int

const (
	ModeTitle Mode = iota
	ModeGame
	ModeGameOver
)

const (
	screenWidth   = 320
	screenHeight  = 240
	frameOX       = 0
	frameOY       = 32
	frameWidth    = 32
	frameHeight   = 32
	frameCount    = 8
	titleFontSize = fontSize * 1.5
	fontSize      = 8
)
