package main

import (
	"golang.org/x/image/math/f64"
	"math/rand"
	"time"
)

type Player struct {
	id                int
	score             int // 玩家的得分
	count             int
	x, y              float64   // 人物在屏幕上的位置
	speed             float64   // 人物移动速度
	weapon            Weapon    // 武器的具体类型
	weaponX           float64   // 武器相对于人物中心的X偏移
	weaponY           float64   // 武器相对于人物中心的Y偏移
	health            float64   // 人物的生命值
	lastCollisionTime time.Time // 上次碰撞发生的时间
	directIdx         int       // 人物的方向
	isSkill           bool      // 是否释放技能
	skillFrame        int       // 技能的帧数
	skillTime         time.Time // 技能释放的时间
	startTime         time.Time // 游戏开始的时间
}

// Invincible 是否无敌
func (p *Player) Invincible() bool {
	return p.isSkill
}

func (p *Player) Move(dx, dy float64) {
	p.x += dx
	p.y += dy
	p.x = clamp(p.x, -frameWidth/2, screenWidth-frameWidth/2)
	p.y = clamp(p.y, -frameHeight/2, screenHeight-frameHeight/2)
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func GenerateMonster(g *Game) {
	// 随着时间的推移，怪物的数量会增加
	if len(g.monsters) < g.uniqueId/10+3 {
		g.uniqueId++
		g.monsters[g.uniqueId] = &Player{
			id:        g.uniqueId,
			count:     g.player.count,
			x:         rand.Float64() * (screenWidth - frameWidth/2),
			y:         rand.Float64() * (screenHeight - frameHeight/2),
			speed:     1.0 / 180,
			weaponX:   frameWidth / 2,
			weaponY:   frameHeight / 2,
			directIdx: 0,
		}
		g.monsterTimer[g.uniqueId] = 0
		// 以玩家为目标
		g.monsterTarget[g.uniqueId] = f64.Vec2{g.player.x, g.player.y}
	}
}
