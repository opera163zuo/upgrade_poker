package gui

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	baseui "github.com/smallnest/upgrade_poker/ui"
)

type GUI struct {
	st        *state
	mouseDown bool
}

func New() *GUI {
	return &GUI{st: &state{actionCh: make(chan baseui.UIAction, 16), selected: map[int]bool{}}}
}

func (g *GUI) Init() error {
	ebiten.SetWindowSize(LogicalWidth*2, LogicalHeight*2)
	ebiten.SetWindowTitle("upgrade_poker - Ebitengine GUI")
	// 图片加载失败不退出一耐心等待assets目录
	if err := EnsureImagesLoaded(); err != nil {
		// 静默失败，使用备选纯色方块渲染
	}
	return nil
}

func (g *GUI) Run(loop func()) error {
	go loop()
	return ebiten.RunGame(g)
}

func (g *GUI) Close() error { return nil }

func (g *GUI) Render(view baseui.TableView) {
	g.st.mu.Lock()
	view.UpdatedAt = time.Now()
	g.st.view = view
	if view.SelectedIdx != nil {
		g.st.selected = map[int]bool{}
		for k, v := range view.SelectedIdx {
			g.st.selected[k] = v
		}
	}
	g.st.mu.Unlock()
}

func (g *GUI) WaitAction() baseui.UIAction {
	return <-g.st.actionCh
}

func (g *GUI) WaitActionOrTimeout(d time.Duration) (baseui.UIAction, bool) {
	select {
	case act := <-g.st.actionCh:
		return act, false
	case <-time.After(d):
		return baseui.UIAction{Type: baseui.ActionTimeout}, true
	}
}

func (g *GUI) ShowMessage(msg string, buttons []baseui.ButtonSpec) {
	g.st.mu.Lock()
	g.st.view.Message = msg
	g.st.view.Buttons = buttons
	g.st.mu.Unlock()
}

func (g *GUI) ClearMessage() {
	g.st.mu.Lock()
	g.st.view.Message = ""
	g.st.view.Buttons = nil
	g.st.mu.Unlock()
}

func (g *GUI) SetPhase(phase baseui.UIPhase) {
	g.st.mu.Lock()
	g.st.view.Phase = phase
	g.st.selected = map[int]bool{}
	g.st.mu.Unlock()
}

func (g *GUI) SleepForRedraw(d time.Duration) { time.Sleep(d) }

func (g *GUI) Update() error {
	g.updateInput()
	return nil
}

func (g *GUI) Layout(outsideWidth, outsideHeight int) (int, int) {
	return LogicalWidth, LogicalHeight
}
