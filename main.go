package main

import (
	"flag"
	"fmt"
	"os"

	guipkg "github.com/smallnest/upgrade_poker/ui/gui"
	tuipkg "github.com/smallnest/upgrade_poker/ui/tui"
)

var Version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "show version")
	uiMode := flag.String("ui", "gui", "ui mode: tui or gui")
	flag.Parse()

	if *showVersion {
		fmt.Printf("upgrade_poker %s\n", Version)
		os.Exit(0)
	}

	var game *Game
	var gameUI interface {
		Init() error
		Run(loop func()) error
		Close() error
	}

	switch *uiMode {
	case "gui":
		gui := guipkg.New()
		game = NewGame(gui)
		gameUI = gui
	case "tui":
		game = NewGame(nil)
		legacy := NewTUI(game)
		adapter := tuipkg.NewAdapter(legacy)
		game.ui = adapter
		gameUI = adapter
	default:
		fmt.Fprintf(os.Stderr, "未知 UI 模式: %s\n", *uiMode)
		os.Exit(2)
	}

	if err := gameUI.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "无法初始化界面: %v\n", err)
		os.Exit(1)
	}
	defer gameUI.Close()

	if err := gameUI.Run(game.Run); err != nil {
		fmt.Fprintf(os.Stderr, "界面运行失败: %v\n", err)
		os.Exit(1)
	}
}
