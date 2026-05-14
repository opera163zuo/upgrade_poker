package main

import (
	"flag"
	"fmt"
	"os"

	guipkg "github.com/smallnest/upgrade_poker/ui/gui"
)

var Version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("upgrade_poker %s\n", Version)
		os.Exit(0)
	}

	gui := guipkg.New()
	game := NewGame(gui)

	if err := gui.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "无法初始化界面: %v\n", err)
		os.Exit(1)
	}
	defer gui.Close()

	if err := gui.Run(game.Run); err != nil {
		fmt.Fprintf(os.Stderr, "界面运行失败: %v\n", err)
		os.Exit(1)
	}
}
