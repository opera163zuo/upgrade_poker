package gui

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// 牌面图片缓存
var cardImages [57]*ebiten.Image // 0-53 牌面, 54-56 牌背
var cardImagesLoaded bool

// 图片根目录：优先运行时目录，其次可执行文件同级 assets/
func assetDir() string {
	exe, _ := os.Executable()
	base := filepath.Dir(exe)
	// 尝试几个可能的位置
	candidates := []string{
		filepath.Join(base, "assets"),
		filepath.Join(base, "..", "assets"),
		"assets",
	}
	for _, d := range candidates {
		if info, err := os.Stat(d); err == nil && info.IsDir() {
			return d
		}
	}
	return "assets"
}

// EnsureImagesLoaded ensures card images are loaded
func EnsureImagesLoaded() error {
	if cardImagesLoaded {
		return nil
	}
	dir := assetDir()

	for i := 0; i < 54; i++ {
		path := filepath.Join(dir, "cards", "faces", fmt.Sprintf("%d.png", i))
		img, err := loadImage(path)
		if err != nil {
			return fmt.Errorf("load card %d: %w", i, err)
		}
		cardImages[i] = img
	}

	// 牌背
	backNames := []string{"back.png", "back2.png", "back3.png"}
	for i, name := range backNames {
		path := filepath.Join(dir, "cards", "backs", name)
		img, err := loadImage(path)
		if err != nil {
			return err
		}
		cardImages[54+i] = img
	}

	cardImagesLoaded = true
	return nil
}

func loadImage(path string) (*ebiten.Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

// CardFaceImage returns the card face image for a Go card (suit, rank)
func CardFaceImage(suit, rank int) *ebiten.Image {
	if !cardImagesLoaded {
		return nil
	}
	idx := GoCardToCSharpNumber(suit, rank)
	if idx < 0 || idx >= 54 {
		return nil
	}
	return cardImages[idx]
}

// CardBackImage returns a card back image
func CardBackImage(index int) *ebiten.Image {
	if !cardImagesLoaded || index < 0 || index > 2 {
		return cardImages[54] // default back
	}
	return cardImages[54+index]
}

// IsImageLoaded returns whether card images are loaded
func IsImageLoaded() bool {
	return cardImagesLoaded
}

// GoCardToCSharpNumber maps Go card (suit, rank) to C# image index 0-53
// C#: 0-12=♥, 13-25=♠, 26-38=♦, 39-51=♣, 52=小王, 53=大王
func GoCardToCSharpNumber(suit, rank int) int {
	if rank >= 15 {
		if rank == 15 {
			return 52
		} // small joker
		if rank == 16 {
			return 53
		} // big joker
	}
	// C# suit: 1=♥, 2=♠, 3=♦, 4=♣
	// Go suit: 0=♠, 1=♥, 2=♦, 3=♣
	csharpSuit := []int{2, 1, 3, 4}[suit]
	csharpRank := rank - 2
	return (csharpSuit-1)*13 + csharpRank
}

// FreeImages releases loaded images
func FreeImages() {
	for i := range cardImages {
		cardImages[i] = nil
	}
	cardImagesLoaded = false
}
