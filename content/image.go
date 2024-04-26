package main

import (
	"avoid-the-enemies/resources/images"
	"bytes"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"log"
)

var (
	akImage     *ebiten.Image
	bulletImage *ebiten.Image
	runnerImage *ebiten.Image
	sickleImage *ebiten.Image
	swordImage  *ebiten.Image
	skillImage  *ebiten.Image
	fireImage   *ebiten.Image
)

func InitImage() {
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		log.Fatal(err)
	}
	runnerImage = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(images.Sickle_png))
	if err != nil {
		log.Fatal(err)
	}
	sickleImage = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(images.Sword_png))
	if err != nil {
		log.Fatal(err)
	}
	swordImage = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(images.AK_png))
	if err != nil {
		log.Fatal(err)
	}
	akImage = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(images.Bullet_png))
	if err != nil {
		log.Fatal(err)
	}
	bulletImage = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(images.Skill_png))
	if err != nil {
		log.Fatal(err)
	}
	skillImage = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(images.Fire_png))
	if err != nil {
		log.Fatal(err)
	}
	fireImage = ebiten.NewImageFromImage(img)
}
