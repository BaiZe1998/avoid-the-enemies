package main

import (
	"avoid-the-enemies/content/config"
	"bytes"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/math/f64"

	raudio "avoid-the-enemies/resources/audio"
)

var (
	weaponList []Weapon
)

func InitWeapon() {
	s, err := mp3.DecodeWithoutResampling(bytes.NewReader(raudio.Shotgun_mp3))
	if err != nil {
		return
	}
	if audioContext == nil {
		audioContext = audio.NewContext(48000)
	}
	p, err := audioContext.NewPlayer(s)
	if err != nil {
		return
	}
	weaponList = append(weaponList,
		&MeleeWeapon{
			Type:  "sickle",
			Image: sickleImage,
			spin:  1.75 * math.Pi / 60, // 每帧转动的角度（弧度）
			angle: 0,
		},
		&MeleeWeapon{
			Type:  "sword",
			Image: swordImage,
			spin:  2 * math.Pi / 60, // 每帧转动的角度（弧度）
			angle: 0,
		},
		&RangedWeapon{
			Type:       "ak",
			Image:      akImage,
			bullet:     bulletImage,
			speed:      0.5,
			distance:   config.ScreenWidth,
			damage:     25,
			shotPlayer: p,
		},
	)
}

type Weapon interface {
	GetImage() *ebiten.Image
}

type MeleeWeapon struct {
	Type  string        // 武器类型
	Image *ebiten.Image // 加载武器的图片
	angle float64       // 武器的旋转角度
	spin  float64       // 武器的旋转速度
	Trail []f64.Vec2    // 武器的轨迹
}

func (w *MeleeWeapon) GetImage() *ebiten.Image {
	return w.Image
}

func (w *MeleeWeapon) Spin() {
	w.angle += w.spin
	w.angle = math.Mod(w.angle, 2*math.Pi)
}

func (w *MeleeWeapon) Copy() *MeleeWeapon {
	return &MeleeWeapon{
		Type:  w.Type,
		Image: w.Image,
		angle: w.angle,
		spin:  w.spin,
		Trail: w.Trail,
	}
}

// DrawTrail 在绘制时，绘制轨迹效果
func (w *MeleeWeapon) DrawTrail(screen *ebiten.Image) {
	// 绘制轨迹效果
	for i := 1; i < len(w.Trail); i++ {
		prevPos := w.Trail[i-1]
		currPos := w.Trail[i]
		// 绘制当前位置与前一位置之间的轨迹线段
		// 根据需要设置线段的颜色、粗细等属性
		// 例如使用 ebitenutil.DrawLine() 函数
		ebitenutil.DrawLine(screen, prevPos[0], prevPos[1], currPos[0], currPos[1], color.RGBA{255, 255, 255, 128})
	}
}

type RangedWeapon struct {
	Type         string        // 武器类型
	Image        *ebiten.Image // 加载武器的图片
	bullet       *ebiten.Image // 子弹图片
	speed        float64       // 子弹的速度
	distance     float64       // 子弹的射程
	damage       float64       // 子弹的伤害值
	LastFireTime time.Time     // 上次开火的时间
	shotPlayer   *audio.Player // 射击音效
}

func (w *RangedWeapon) GetImage() *ebiten.Image {
	return w.Image
}

func (w *RangedWeapon) Copy() *RangedWeapon {
	return &RangedWeapon{
		Type:       w.Type,
		Image:      w.Image,
		bullet:     w.bullet,
		speed:      w.speed,
		distance:   w.distance,
		damage:     w.damage,
		shotPlayer: w.shotPlayer,
	}
}

type FireOption func(s *Suspend)

func WithBulletDirection(x, y float64) FireOption {
	return func(s *Suspend) {
		s.direction = &SuspendDirection{
			x: x,
			y: y,
		}
	}
}

func (w *RangedWeapon) Fire(g *Game, player *Player, options ...FireOption) {
	if err := w.shotPlayer.Rewind(); err != nil {
		return
	}
	w.shotPlayer.Play()
	// 每次开火生成一颗子弹，移动的距离为 distance 速度为 speed 图片为 bullet
	x, y := player.x+player.weaponX, player.y+player.weaponY
	weapon := player.weapon.(*RangedWeapon)
	bullet := &Suspend{
		pos:         f64.Vec2{x, y},
		rangeWeapon: weapon,
		// 子弹的起始位置
		from:        f64.Vec2{x, y},
		directIndex: player.directIdx,
		PlayerID:    player.id,
	}

	for _, option := range options {
		option(bullet)
	}

	g.uniqueId++
	g.suspends[g.uniqueId] = bullet
}

type Suspend struct {
	pos         f64.Vec2          // 当前位置
	rangeWeapon *RangedWeapon     // 武器
	from        f64.Vec2          // 发射子弹的位置
	directIndex int               // 子弹的方向
	time        int               // 子弹的生命周期
	PlayerID    int               // 子弹的拥有者
	direction   *SuspendDirection // 子弹运动的方向向量
}

type SuspendDirection struct {
	x, y float64
}

func GenerateWeapon(g *Game) {
	if time.Since(g.weaponTimer) > time.Second*5 {
		g.weaponTimer = time.Now()
		if len(g.weapons) < 2 {
			g.uniqueId++
			weapon := weaponList[rand.Intn(len(weaponList))]
			switch weapon.(type) {
			case *MeleeWeapon:
				newWeapon := weapon.(*MeleeWeapon).Copy()
				// 使用指针类型有拷贝的bug，当两个人获得同一把武器的时候，旋转会画两次，所以看起来快了一倍
				g.weapons[g.uniqueId] = newWeapon
				g.weaponPosition[g.uniqueId] = f64.Vec2{rand.Float64() * (config.ScreenWidth - config.FrameWidth/2), rand.Float64() * (config.ScreenHeight - config.FrameHeight/2)}
			case *RangedWeapon:
				newWeapon := weapon.(*RangedWeapon).Copy()
				g.weapons[g.uniqueId] = newWeapon
				g.weaponPosition[g.uniqueId] = f64.Vec2{rand.Float64() * (config.ScreenWidth - config.FrameWidth/2), rand.Float64() * (config.ScreenHeight - config.FrameHeight/2)}
			}
		}
	}
}

// SuspendMove 更新所有远程武器的发射产物位置
func SuspendMove(g *Game) {
	for id, s := range g.suspends {
		s.time++

		if s.direction != nil {
			s.pos[0] += s.direction.x * s.rangeWeapon.speed * float64(s.time)
			s.pos[1] += s.direction.y * s.rangeWeapon.speed * float64(s.time)
		} else {
			s.pos[0] += directions[s.directIndex].dx * s.rangeWeapon.speed * float64(s.time)
			s.pos[1] += directions[s.directIndex].dy * s.rangeWeapon.speed * float64(s.time)
		}

		// 如果子弹超出射程，删除子弹
		if math.Abs(s.pos[0]-s.from[0]) > s.rangeWeapon.distance || math.Abs(s.pos[1]-s.from[1]) > s.rangeWeapon.distance {
			delete(g.suspends, id)
			continue
		}
		// 如果子弹碰撞到怪物，怪物消失
		for _, m := range g.monsters {
			if m.id != s.PlayerID && IsTouch(s.pos[0], s.pos[1], m.x+config.FrameWidth/2, m.y+config.FrameHeight/2) {
				delete(g.suspends, id)
				delete(g.monsters, m.id)
				g.player.score++
				break
			}
		}
		// 如果子弹碰撞到玩家，且玩家不是无敌状态，玩家减血
		if s.PlayerID != g.player.id && !g.player.Invincible() && IsTouch(s.pos[0], s.pos[1], g.player.x+config.FrameWidth/2, g.player.y+config.FrameHeight/2) {
			g.player.health -= s.rangeWeapon.damage
			delete(g.suspends, id)
			if g.player.health <= 0 {
				g.mode = config.ModeGameOver
			}
		}
	}
}
