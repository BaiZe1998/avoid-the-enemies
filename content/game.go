package main

import (
	"avoid-the-enemies/content/config"
	"bytes"
	"image"
	"image/color"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	raudio "github.com/hajimehoshi/ebiten/v2/examples/resources/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/math/f64"

	"avoid-the-enemies/content/utils"
)

type Game struct {
	mode                     config.Mode
	player                   *Player
	uniqueId                 int
	monsters                 map[int]*Player
	monsterTarget            map[int]f64.Vec2 // 记录每个怪物的目标位置
	monsterTimer             map[int]int      // 记录每个怪物的计时器
	weaponTimer              time.Time        // 武器刷新时间
	weaponPosition           map[int]f64.Vec2 // 武器位置
	weaponPositionBeenPicked map[int]bool     // 某个武器位置是否已经被某个怪物标记为了目标
	weapons                  map[int]Weapon
	suspends                 map[int]*Suspend
	hitPlayer                *audio.Player
}

func (g *Game) init() {
	g.mode = config.ModeTitle
	g.player = &Player{
		x:                 config.ScreenWidth/2 - config.FrameWidth/2,
		y:                 config.ScreenHeight/2 - config.FrameHeight/2,
		speed:             2.0, // 您可以根据需要调整这个值
		weaponX:           config.FrameWidth / 2,
		weaponY:           config.FrameHeight / 2,
		health:            100,
		lastCollisionTime: time.Now(),
		directIdx:         0,
		id:                1,
		score:             0,
		isSkill:           false,
		skillFrame:        0,
		startTime:         time.Now(),
	}
	g.monsters = make(map[int]*Player)
	g.monsterTarget = make(map[int]f64.Vec2)
	g.monsterTimer = make(map[int]int)
	g.weaponTimer = time.Now()
	g.weaponPosition = make(map[int]f64.Vec2)
	g.weaponPositionBeenPicked = make(map[int]bool)
	g.weapons = make(map[int]Weapon)
	g.suspends = make(map[int]*Suspend)
	g.uniqueId = 1

	if audioContext == nil {
		audioContext = audio.NewContext(48000)
	}
	jabD, err := wav.DecodeWithoutResampling(bytes.NewReader(raudio.Jab_wav))
	if err != nil {
		log.Fatal(err)
	}
	g.hitPlayer, err = audioContext.NewPlayer(jabD)
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Update() error {
	switch g.mode {
	case config.ModeTitle:
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.mode = config.ModeGame
			g.player.startTime = time.Now()
		}
	case config.ModeGame:
		if err := g.resolveModeGame(); err != nil {
			return err
		}
	case config.ModeGameOver:
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.init()
			g.mode = config.ModeTitle
		}
	}

	return nil
}

func (g *Game) resolveModeGame() error {
	g.player.count++

	g.resolveKeyPressed()

	// 如果 GIF 正在播放且播放完成，则停止播放
	if g.player.isSkill && time.Since(g.player.skillTime) > time.Second*3 {
		g.player.isSkill = false
	}

	// 更新所有远程武器的发射产物位置
	SuspendMove(g)

	// 生成怪物
	GenerateMonster(g)

	// 武器在地图上随机位置刷新
	GenerateWeapon(g)

	g.resolvePickWeapon()

	g.resolvePlayerWeapon()

	if err := g.resolveMonsters(); err != nil {
		return err
	}

	return nil
}

func (g *Game) resolveKeyPressed() {
	// 检查键盘输入，人物移动
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.player.Move(-g.player.speed, 0)
		g.player.directIdx = 2
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.player.Move(g.player.speed, 0)
		g.player.directIdx = 0
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.player.Move(0, -g.player.speed)
		g.player.directIdx = 3
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.player.Move(0, g.player.speed)
		g.player.directIdx = 1
	}

	// 按下 q 键可以释放技能 && 距离上一次释放技能时间大于技能冷却时间
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) && time.Since(g.player.skillTime) > time.Second*5 {
		if g.player.score >= 20 {
			g.player.isSkill = true
			g.player.skillTime = time.Now()
			g.player.score -= 20
			g.player.skillFrame = 0
		}
	}

	// 如果人物有武器，且是远程武器，按空格键开火
	if g.player.weapon != nil {
		// 检测空格键按下事件
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			// 如果上次未开火过，执行开火操作
			switch g.player.weapon.(type) {
			case *RangedWeapon:
				weapon := g.player.weapon.(*RangedWeapon)
				weapon.Fire(g, g.player)
			}
		}
	}
}

func (g *Game) resolveImpact() {
	g.resolvePickWeapon()
}

func (g *Game) resolvePickWeapon() {
	for id, weapon := range g.weapons {
		// 玩家移动到武器位置可以获得武器
		if IsTouch(g.player.x, g.player.y, g.weaponPosition[id][0], g.weaponPosition[id][1]) {
			g.player.weapon = weapon
			delete(g.weapons, id)
			delete(g.weaponPosition, id)
			break
		}
		// 怪物移动到武器位置可以获得武器
		for _, monster := range g.monsters {
			if IsTouch(monster.x, monster.y, g.weaponPosition[id][0], g.weaponPosition[id][1]) {
				monster.weapon = weapon
				delete(g.weapons, id)
				delete(g.weaponPosition, id)
				break
			}
		}
	}
}

func (g *Game) resolvePlayerWeapon() error {
	if g.player.weapon != nil {
		switch g.player.weapon.(type) {
		case *MeleeWeapon:
			// 角色武器旋转
			weapon := g.player.weapon.(*MeleeWeapon)
			weapon.Spin()
			// 武器碰撞到敌人可以消灭敌人
			weaponCenterOffsetX := g.player.weaponX // 武器中心相对于角色中心的 X 坐标偏移
			weaponCenterOffsetY := g.player.weaponY // 武器中心相对于角色中心的 Y 坐标偏移
			// 考虑武器的旋转角度，将偏移向量旋转到合适的位置
			weaponCenterX := g.player.x + weaponCenterOffsetX + weaponCenterOffsetX*math.Cos(weapon.angle) - weaponCenterOffsetY*math.Sin(weapon.angle)
			weaponCenterY := g.player.y + weaponCenterOffsetY + weaponCenterOffsetX*math.Sin(weapon.angle) + weaponCenterOffsetY*math.Cos(weapon.angle)
			// 武器的轨迹
			weapon.Trail = append(weapon.Trail, f64.Vec2{weaponCenterX, weaponCenterY})
			if len(weapon.Trail) >= 20 {
				weapon.Trail = weapon.Trail[1:]
			}
			for id, monster := range g.monsters {
				// 怪物的中心位置
				monsterCenterX := monster.x + config.FrameWidth/2
				monsterCenterY := monster.y + config.FrameHeight/2
				if IsTouch(weaponCenterX, weaponCenterY, monsterCenterX, monsterCenterY) {
					if err := g.hitPlayer.Rewind(); err != nil {
						return err
					}
					g.hitPlayer.Play()
					g.player.score++
					delete(g.monsters, id)
					delete(g.monsterTarget, id)
					delete(g.monsterTimer, id)

					if monster.hasSteadyWeaponPosition {
						delete(g.weaponPositionBeenPicked, monster.steadyWeaponId)
						log.Println("delete picked weapon", monster.steadyWeaponId)
					}
				}
			}
		case *RangedWeapon:
			// TODO
		}
	}

	return nil
}

func (g *Game) resolveMonsters() error {
	var target f64.Vec2

	// 正在追逐玩家的怪物
	chasingMonsters := make(map[int]*Player)

	// 怪物移动
	for id, monster := range g.monsters {
		if monster.hasSteadyWeaponPosition {
			_, ok := g.weaponPosition[monster.steadyWeaponId]
			if !ok {
				monster.hasSteadyWeaponPosition = false
			}
		}

		if monster.hasSteadyWeaponPosition && monster.weapon != nil {
			monster.hasSteadyWeaponPosition = false
		}

		inChasing := false

		if monster.hasSteadyWeaponPosition {
			// 当一个怪物存在一定要去的位置时
			monster.directIdx = utils.GetDirectionIdxByTargetPosition(
				monster.steadyWeaponPosition[0], monster.steadyWeaponPosition[1],
				monster.x, monster.y,
			)

			target = monster.steadyWeaponPosition
		} else if len(g.weaponPosition) == 0 {
			// 当地图上没有武器时
			monster.directIdx = utils.GetDirectionIdxByTargetPosition(g.player.x, g.player.y, monster.x, monster.y)

			target = g.monsterTarget[id]

			switch monster.weapon.(type) {
			case *RangedWeapon:
				continue
			default:
				inChasing = true
				chasingMonsters[id] = monster
			}
		} else {
			// 当地图上有武器时

			// 寻找距离最近的武器
			nearWeaponId := 0
			nearDistance := math.Inf(1)
			var nearPosition f64.Vec2
			for id, weapon := range g.weaponPosition {
				// 如果一把武器已经被某个怪物标记过了，则不再前往
				if ok := g.weaponPositionBeenPicked[id]; ok {
					continue
				}

				distance := utils.GetDistance(weapon[0], weapon[1], monster.x, monster.y)

				if distance < nearDistance {
					nearDistance = distance
					nearWeaponId = id
					nearPosition = weapon
				}
			}

			if nearWeaponId != 0 {
				monster.steadyWeaponId = nearWeaponId
				g.weaponPositionBeenPicked[nearWeaponId] = true
				monster.hasSteadyWeaponPosition = true
				monster.steadyWeaponPosition = nearPosition
			}

			if monster.hasSteadyWeaponPosition {
				monster.directIdx = utils.GetDirectionIdxByTargetPosition(
					monster.steadyWeaponPosition[0], monster.steadyWeaponPosition[1],
					monster.x, monster.y,
				)

				target = monster.steadyWeaponPosition
			} else {
				monster.directIdx = utils.GetDirectionIdxByTargetPosition(g.player.x, g.player.y, monster.x, monster.y)

				target = g.monsterTarget[id]
				switch monster.weapon.(type) {
				case *RangedWeapon:
					continue
				default:
					inChasing = true
					chasingMonsters[id] = monster
				}
			}
		}

		timer := g.monsterTimer[id]

		// 更新计时器
		timer++
		g.monsterTimer[id] = timer

		// 每隔一定时间更新一次目标位置
		if timer >= 60 {
			// 以玩家为目标
			g.monsterTarget[id] = f64.Vec2{g.player.x, g.player.y}
			g.monsterTimer[id] = 0
		}

		if inChasing {
			continue
		}

		// 计算当前位置到目标位置的方向向量
		directionX, directionY := utils.Normalize(target[0]-monster.x, target[1]-monster.y)

		if monster.weapon == nil {
			// 在移动轨迹上进行插值
			monster.Move(directionX*monster.speed, directionY*monster.speed)
		}
	}

	if len(chasingMonsters) == 0 {
		return nil
	}

	// 计算所有追逐主角的怪物的中心
	centerX := 0.0
	centerY := 0.0
	for _, monster := range chasingMonsters {
		centerX += monster.x
		centerY += monster.y
	}

	centerX /= float64(len(chasingMonsters))
	centerY /= float64(len(chasingMonsters))

	for id, monster := range chasingMonsters {
		target := g.monsterTarget[id]

		// 计算中心点在怪物和玩家之间的投影
		projectionX, projectionY := utils.GetProjection(monster.x, monster.y, g.player.x, g.player.y, centerX, centerY)
		distance2Monster := utils.GetDistance(projectionX, projectionY, monster.x, monster.y)
		distance2Player := utils.GetDistance(projectionX, projectionY, g.player.x, g.player.y)
		distance := utils.GetDistance(monster.x, monster.y, g.player.x, g.player.y)

		// 投影点在怪物之后
		var correctX, correctY float64
		if distance2Player > distance2Monster && distance2Player > distance {
			correctX, correctY = utils.Normal(monster.x-projectionX, monster.y-projectionY)
			correctX *= 100
			correctY *= 100
		}

		// 计算当前位置到目标位置的方向向量
		directionX, directionY := utils.Normalize(target[0]-monster.x, target[1]-monster.y)
		log.Println("怪物移动量", id, directionX, directionY)

		directionX += correctX
		directionY += correctY
		log.Println("怪物修正后移动量", id, directionX, directionY)

		// Flocking
		for otherId, otherMonster := range chasingMonsters {
			if id == otherId {
				continue
			}

			distance := utils.GetDistance(
				monster.x, monster.y, otherMonster.x, otherMonster.y,
			)

			if distance > config.MonsterMinDistance {
				continue
			}
		}

		if monster.weapon == nil {
			// 在移动轨迹上进行插值
			monster.Move(directionX*monster.speed, directionY*monster.speed)
		}
	}

	for id, monster := range g.monsters {
		target := g.monsterTarget[id]

		// 计算当前位置到目标位置的方向向量
		directionX, directionY := utils.Normalize(target[0]-monster.x, target[1]-monster.y)

		if monster.weapon != nil {
			switch monster.weapon.(type) {
			// 怪物武器旋转
			case *MeleeWeapon, nil:
				// 只有拿着非远程武器的怪物才会移动
				monster.Move(directionX*monster.speed, directionY*monster.speed)

				weapon := monster.weapon.(*MeleeWeapon)
				weapon.Spin()

				// 武器碰撞到玩家，降低生命值
				weaponCenterOffsetX := monster.weaponX // 武器中心相对于怪物中心的 X 坐标偏移
				weaponCenterOffsetY := monster.weaponY // 武器中心相对于怪物中心的 Y 坐标偏移

				// 考虑武器的旋转角度，将偏移向量旋转到合适的位置
				weaponCenterX := monster.x + weaponCenterOffsetX + weaponCenterOffsetX*math.Cos(weapon.angle) - weaponCenterOffsetY*math.Sin(weapon.angle)
				weaponCenterY := monster.y + weaponCenterOffsetY + weaponCenterOffsetX*math.Sin(weapon.angle) + weaponCenterOffsetY*math.Cos(weapon.angle)

				// 武器的轨迹
				weapon.Trail = append(weapon.Trail, f64.Vec2{weaponCenterX, weaponCenterY})
				if len(weapon.Trail) >= 20 {
					weapon.Trail = weapon.Trail[1:]
				}

				// 角色的中心位置
				playerCenterX := g.player.x + config.FrameWidth/2
				playerCenterY := g.player.y + config.FrameHeight/2
				// 并非无敌状态，且碰撞到角色，降低角色生命值
				if !g.player.Invincible() && IsTouch(weaponCenterX, weaponCenterY, playerCenterX, playerCenterY) {
					if time.Since(g.player.lastCollisionTime) < time.Second {
						continue
					}
					if err := g.hitPlayer.Rewind(); err != nil {
						return err
					}
					g.hitPlayer.Play()
					g.player.health -= 25
					g.player.lastCollisionTime = time.Now()
					if g.player.health <= 0 {
						g.mode = config.ModeGameOver
					}
				}
			case *RangedWeapon:
				weapon := monster.weapon.(*RangedWeapon)

				// 每秒钟发射一颗子弹
				if time.Since(weapon.LastFireTime) > time.Second {
					weapon.LastFireTime = time.Now()
					weapon.Fire(g, monster, WithBulletDirection(directionX, directionY))
				}
			}
		}

		// 并非无敌状态，怪物碰撞到人物，降低生命值
		if !g.player.Invincible() && IsTouch(g.player.x, g.player.y, monster.x, monster.y) {
			if time.Since(g.player.lastCollisionTime) < time.Second {
				continue
			}
			if err := g.hitPlayer.Rewind(); err != nil {
				return err
			}
			g.hitPlayer.Play()
			g.player.health -= 25
			g.player.lastCollisionTime = time.Now()
			if g.player.health <= 0 {
				g.mode = config.ModeGameOver
			}
		}
	}

	return nil
}

// Draw 每次绘制都会调用这个函数，重新设置画面元素的内容
func (g *Game) Draw(screen *ebiten.Image) {
	var titleTexts string
	var texts string
	switch g.mode {
	case config.ModeTitle:
		titleTexts = "Avoid the Enemies"
		texts = "PRESS SPACE KEY TO START"
	case config.ModeGameOver:
		titleTexts = "Game Over"
		texts = "PRESS SPACE KEY TO RESTART"
	}

	// 绘制标题
	op := &text.DrawOptions{}
	op.GeoM.Translate(config.ScreenWidth/2, 5*config.TitleFontSize)
	op.ColorScale.ScaleWithColor(color.White)
	op.LineSpacing = config.TitleFontSize
	op.PrimaryAlign = text.AlignCenter
	text.Draw(screen, titleTexts, &text.GoTextFace{
		Source: arcadeFaceSource,
		Size:   config.TitleFontSize,
	}, op)

	op = &text.DrawOptions{}
	op.GeoM.Translate(config.ScreenWidth/2, 7*config.TitleFontSize)
	op.ColorScale.ScaleWithColor(color.White)
	op.LineSpacing = config.FontSize
	op.PrimaryAlign = text.AlignCenter
	text.Draw(screen, texts, &text.GoTextFace{
		Source: arcadeFaceSource,
		Size:   config.FontSize,
	}, op)

	if g.mode == config.ModeGame {
		// 绘制分数
		op = &text.DrawOptions{}
		op.GeoM.Translate(3, 3)
		op.ColorScale.ScaleWithColor(color.White)
		op.LineSpacing = config.FontSize
		text.Draw(screen, "Score: "+strconv.Itoa(g.player.score), &text.GoTextFace{
			Source: arcadeFaceSource,
			Size:   config.FontSize,
		}, op)

		// 绘制游戏时间
		op = &text.DrawOptions{}
		op.GeoM.Translate(config.ScreenWidth/2, 3)
		op.ColorScale.ScaleWithColor(color.White)
		op.LineSpacing = config.FontSize
		text.Draw(screen, "SurvivalTime: "+strconv.Itoa(int(time.Since(g.player.startTime).Seconds()))+"s", &text.GoTextFace{
			Source: arcadeFaceSource,
			Size:   config.FontSize,
		}, op)

		// 绘制技能效果
		if g.player.isSkill {
			g.player.skillFrame++
			op := &ebiten.DrawImageOptions{}
			// 位于血条上方，血条高度为 5
			op.GeoM.Translate(g.player.x-16, g.player.y-5-16)
			//op.GeoM.Translate(g.player.x-8, g.player.y-5-36)
			i := (g.player.skillFrame / 5) % 4
			//i := (g.player.skillFrame / 5) % 90
			sx, sy := i*64, 0
			//sx, sy := i*48, 0
			screen.DrawImage(fireImage.SubImage(image.Rect(sx, sy, sx+64, sy+64)).(*ebiten.Image), op)
			//screen.DrawImage(skillImage.SubImage(image.Rect(sx, sy, sx+48, sy+36)).(*ebiten.Image), op)
		}

		// 绘制角色
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(g.player.x, g.player.y)
		i := (g.player.count / 5) % config.FrameCount
		sx, sy := config.FrameOX+i*config.FrameWidth, config.FrameOY
		screen.DrawImage(runnerImage.SubImage(image.Rect(sx, sy, sx+config.FrameWidth, sy+config.FrameHeight)).(*ebiten.Image), op)

		// 绘制角色武器
		if g.player.weapon != nil {
			switch g.player.weapon.(type) {
			case *MeleeWeapon:
				weapon := g.player.weapon.(*MeleeWeapon)
				op = &ebiten.DrawImageOptions{}
				op.GeoM.Rotate(weapon.angle)
				op.GeoM.Translate(g.player.x+g.player.weaponX, g.player.y+g.player.weaponY)
				screen.DrawImage(weapon.Image.SubImage(image.Rect(0, 0, config.FrameWidth, config.FrameHeight)).(*ebiten.Image), op)
				weapon.DrawTrail(screen)
			case *RangedWeapon:
				weapon := g.player.weapon.(*RangedWeapon)
				op = &ebiten.DrawImageOptions{}
				op.GeoM.Rotate(directions[g.player.directIdx].spin)
				op.GeoM.Translate(rotateAdjust[g.player.directIdx].dx*config.FrameWidth, rotateAdjust[g.player.directIdx].dy*config.FrameHeight)
				op.GeoM.Translate(g.player.x, g.player.y)
				screen.DrawImage(weapon.Image.SubImage(image.Rect(0, 0, config.FrameWidth, config.FrameHeight)).(*ebiten.Image), op)
			}
		}

		// 绘制武器发射产物
		for _, suspend := range g.suspends {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Rotate(directions[g.player.directIdx].spin)
			op.GeoM.Translate(suspend.pos[0], suspend.pos[1])
			screen.DrawImage(suspend.rangeWeapon.bullet.SubImage(image.Rect(0, 0, config.FrameWidth, config.FrameHeight)).(*ebiten.Image), op)
		}

		// 绘制怪物
		for _, monster := range g.monsters {
			op = &ebiten.DrawImageOptions{}
			op.GeoM.Translate(monster.x, monster.y)
			screen.DrawImage(runnerImage.SubImage(image.Rect(sx, sy, sx+config.FrameWidth, sy+config.FrameHeight)).(*ebiten.Image), op)
			// 绘制怪物武器
			if monster.weapon != nil {
				switch monster.weapon.(type) {
				case *MeleeWeapon:
					weapon := monster.weapon.(*MeleeWeapon)
					op = &ebiten.DrawImageOptions{}
					op.GeoM.Rotate(weapon.angle)
					op.GeoM.Translate(monster.x+monster.weaponX, monster.y+monster.weaponY)
					screen.DrawImage(weapon.Image.SubImage(image.Rect(0, 0, config.FrameWidth, config.FrameHeight)).(*ebiten.Image), op)
					weapon.DrawTrail(screen)
				case *RangedWeapon:
					weapon := monster.weapon.(*RangedWeapon)
					op = &ebiten.DrawImageOptions{}
					op.GeoM.Rotate(directions[monster.directIdx].spin)
					op.GeoM.Translate(rotateAdjust[monster.directIdx].dx*config.FrameWidth, rotateAdjust[monster.directIdx].dy*config.FrameHeight)
					op.GeoM.Translate(monster.x, monster.y)
					screen.DrawImage(weapon.Image.SubImage(image.Rect(0, 0, config.FrameWidth, config.FrameHeight)).(*ebiten.Image), op)
				}
			}
		}

		// 设置血条的位置和尺寸
		x := g.player.x
		y := g.player.y - 5                                         // 位于角色头顶上方
		width := float64(config.FrameWidth) * g.player.health / 100 // 血条宽度根据当前血量动态变化
		height := 5                                                 // 血条高度
		// 绘制血条底部
		ebitenutil.DrawRect(screen, x, y, float64(config.FrameWidth), float64(height), color.Gray{0x80})
		// 绘制血条
		ebitenutil.DrawRect(screen, x, y, float64(width), float64(height), color.RGBA{0xFF, 0x00, 0x00, 0xFF})

		// 地图上的武器
		for id, weapon := range g.weapons {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(g.weaponPosition[id][0], g.weaponPosition[id][1])
			screen.DrawImage(weapon.GetImage().SubImage(image.Rect(0, 0, config.FrameWidth, config.FrameHeight)).(*ebiten.Image), op)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}
