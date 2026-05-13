# upgrade_poker Ebitengine 图形界面技术方案

## 0. 目标与现状

目标：在 **不改坏现有 game/rules/ai 逻辑** 的前提下，为 `upgrade_poker` 增加一个基于 **Ebitengine v2** 的 2D 图形界面，并与当前 TUI 并存。

现状观察：
- 当前入口在 `main.go`，直接 `NewGame()` + `NewTUI(game)` + `tui.Run()`。
- `game.go` 直接依赖 `*TUI`，存在大量 `g.tui.SetPhase/SetMessage/WaitForAction/SleepForRedraw` 调用，并直接读写 `g.tui.selected / dealCounts / thinking / waitingForHuman / bidOptions` 等字段。
- 这意味着 **先做 UI 抽象层** 是第一优先级，否则 GUI 会被迫复制 TUI 的内部状态机，维护成本很高。

建议采用：
1. `game.go` 只依赖 `ui.GameUI` 接口。
2. TUI 退化为 `ui/tui` 的一个实现；Ebitengine GUI 作为 `ui/gui` 的另一个实现。
3. 游戏逻辑通过“**快照 + 动作**”与 UI 通信，而不是直接改 UI 内部字段。

---

## 1. 架构设计

### 1.1 抽象目标

把现在 `tui.go` 承担的三类职责拆开：

1. **输入等待**：等玩家点击/按键/确认。
2. **界面状态更新**：显示当前阶段、消息、按钮、可选操作。
3. **渲染动画**：发牌、AI 思考、出牌结果、结算停留。

`game.go` 不应知道“终端格子”或“鼠标矩形”，只应知道：
- 当前要显示什么；
- 当前允许玩家做什么；
- 用户做了什么动作。

### 1.2 推荐目录结构

```text
/home/zxyzxy/文档/upgrade_poker/
├── main.go
├── game.go
├── tui.go                  # 过渡期保留，最终迁到 ui/tui/
├── ui/
│   ├── types.go            # UIAction / UIPhase / ButtonSpec / BidChoice / Snapshot
│   ├── interface.go        # GameUI 接口
│   ├── snapshot.go         # 从 Game 生成可渲染快照的辅助函数
│   ├── tui/
│   │   ├── ui.go           # TUI 实现（从现有 tui.go 拆出）
│   │   ├── draw.go
│   │   └── input.go
│   └── gui/
│       ├── ui.go           # Ebitengine 实现，满足 GameUI
│       ├── layout.go       # 像素布局、命中测试、缩放
│       ├── render.go       # Draw 阶段
│       ├── input.go        # 鼠标/键盘/双击/拖拽
│       ├── anim.go         # Tween、发牌/出牌动画
│       ├── assets.go       # go:embed / 图片加载 / 资源表
│       ├── sprites.go      # 牌面、按钮、背景切片
│       └── state.go        # GUI 内部瞬时状态（选中牌、hover、动画队列）
├── assets/
│   ├── cards/
│   │   ├── faces/*.png
│   │   ├── backs/*.png
│   │   └── manifest.json
│   ├── bg/table.png
│   ├── ui/buttons.png
│   └── audio/*.wav
└── docs/
    └── ebitengine-plan.md
```

### 1.3 接口设计

#### 1.3.1 UI 动作模型

```go
package ui

import "time"

type UIPhase int

const (
    PhaseWelcome UIPhase = iota
    PhaseDealing
    PhaseBidding
    PhaseDiscard
    PhasePlaying
    PhaseWaitTrick
    PhaseHandResult
    PhaseGameOver
)

type ActionType string

const (
    ActionStart   ActionType = "start"
    ActionPlay    ActionType = "play"
    ActionCancel  ActionType = "cancel"
    ActionBid     ActionType = "bid"
    ActionPass    ActionType = "pass"
    ActionConfirm ActionType = "confirm"
    ActionQuit    ActionType = "quit"
)

type UIAction struct {
    Type    ActionType
    CardIdx []int
    BidType BidType
    BidSuit Suit
}

type ButtonSpec struct {
    ID      string
    Label   string
    Enabled bool
}

type BidChoice struct {
    Type BidType
    Suit Suit
    Text string
}
```

#### 1.3.2 渲染快照

`game.go` 不直接把 `Game` 整个对象暴露给 GUI，而是生成只读快照：

```go
package ui

type PlayerView struct {
    Position    PlayerPosition
    Name        string
    IsHuman     bool
    HandCount   int
    HandCards   []Card // 仅南家暴露正面；其他玩家默认为空
    PlayedCards []Card
    IsDealer    bool
    IsThinking  bool
}

type TableView struct {
    Phase         UIPhase
    TrumpSuit     Suit
    Dealer        PlayerPosition
    DealerLevel   Rank
    OpponentLevel Rank
    TeamScore     [2]int
    TrickCount    int
    BottomCards   []Card
    CurrentTrick  map[PlayerPosition][]Card
    Message       string
    Buttons       []ButtonSpec
    BidChoices    []BidChoice
    Players       [4]PlayerView
    WaitingForHuman bool
    SelectedIdx   map[int]bool
    NeedSelect    int
    DiscardCount  int
    TrickWinner   PlayerPosition
    TrickPoints   int
}
```

#### 1.3.3 `GameUI` 接口

```go
package ui

import "time"

type GameUI interface {
    Init() error
    Run(loop func()) error
    Close() error

    // 渲染/状态
    Render(view TableView)

    // 输入
    WaitAction() UIAction
    WaitActionOrTimeout(d time.Duration) (UIAction, bool)

    // 同步提示/模态
    ShowMessage(msg string, buttons []ButtonSpec)
    ClearMessage()
    SetPhase(phase UIPhase)

    // 动画/等待
    SleepForRedraw(d time.Duration)
}
```

> 这套接口与当前 `game.go` 结构最接近，迁移成本最低。后续若要进一步解耦，可以把 `ShowMessage/SetPhase` 也折叠进单次 `Render(view)` 调用里。

### 1.4 `game.go` 需要怎样改

当前直接依赖 TUI 的代码可分 4 类：

1. **状态切换**：`SetPhase(...)`
2. **消息框**：`SetMessage(...)`
3. **阻塞等待输入**：`WaitForAction()` / `WaitForActionOrTimeout()`
4. **动画与瞬时 UI 状态**：`dealCounts`、`thinking`、`waitingForHuman`、`selected`

建议把第 4 类从“直接改 UI 字段”改为“更新快照数据 + 调用 Render”。例如：

```go
// 旧
// g.tui.thinkingPos = currentPlayer
// g.tui.thinking = true
// g.tui.SleepForRedraw(600 * time.Millisecond)
// g.tui.thinking = false

// 新
view := g.BuildTableView()
view.Players[currentPlayer].IsThinking = true
g.ui.Render(view)
g.ui.SleepForRedraw(600 * time.Millisecond)
view.Players[currentPlayer].IsThinking = false
g.ui.Render(view)
```

### 1.5 最小可执行迁移策略

不要一口气重写 `tui.go`。按下面顺序：

1. 新建 `ui/types.go`、`ui/interface.go`。
2. 把 `game.go` 的 `tui *TUI` 改成 `ui GameUI`。
3. 先写一个 **TUIAdapter**，内部仍调用旧 `TUI`，保证逻辑继续能跑。
4. 再增量开发 `ui/gui`。

这样不会在“抽象”和“GUI 实现”两个维度同时爆炸。

---

## 2. UI 布局设计（像素级）

### 2.1 参考尺寸

参考原版 WinForms：
- 窗口 ClientSize：`629 x 470`
- 菜单栏约：`24px`
- 实际牌桌绘制区接近：`590 x 400`

建议 Ebitengine 使用：

- **逻辑分辨率**：`640 x 480`
- **牌桌安全区**：`590 x 400`
- **牌桌左上角偏移**：`(25, 40)`

这样兼顾：
- 与原版布局接近；
- 留出顶部信息栏和外边距；
- 以后可以整数缩放到 1280×960 / 1920×1440。

### 2.2 牌面尺寸

原 C# 资源牌图尺寸为：`71 x 96`（已验证 `/home/zxyzxy/文档/tractor/CardSources/Resources/*.png`）。

建议 GUI 直接沿用：
- 手牌/桌面明牌：`71 x 96`
- 小按钮图标：`24 x 24`
- 牌背同尺寸：`71 x 96`

### 2.3 总体布局坐标

#### 逻辑画布基准

```text
窗口逻辑尺寸: 640 x 480
牌桌区:      x=25, y=40, w=590, h=400
```

#### 顶部信息栏

```text
x=25, y=8, w=590, h=24
```

显示内容：
- 左：庄家、庄打等级、闲打等级
- 右：主花色、当前闲家得分、已完成轮数

#### 中央桌面区

```text
x=77, y=124, w=476, h=244
```

这组数字直接对齐原 WinForms 中央清理区，后续动画和出牌落点最稳定。

### 2.4 四个玩家区域

#### 南家（玩家自己，底部手牌）

- 区域：`x=30, y=355, w=560, h=116`
- 名称标签：`x=300, y=333`（中心对齐）
- 等待提示“等你出牌…”：`x=270, y=315`

手牌排列参数：
- 牌宽高：`71 x 96`
- 默认间距：`18px`
- 选中抬起：`20px`
- 当手牌超过 14 张时，动态把间距压缩到 `max(10, floor((560-71)/(n-1)))`

建议基准公式：

```go
handStartX := 30 + (560 - (71 + (n-1)*spacing)) / 2
handY      := 375   // 未选中
selectY    := 355   // 选中后抬起 20px
```

#### 北家（对家，顶部）

- 区域：`x=105, y=25, w=420, h=96`
- 标签：`x=300, y=8`
- 牌背从右向左错开：间距 `13px`

公式：

```go
startX := 437 - count*13
y := 25
```

#### 西家（左侧）

- 区域：`x=6, y=140, w=71, h=202`
- 标签：`x=8, y=120`
- 牌背竖向堆叠：间距 `4px`

公式：

```go
x := 6
y := 145 + i*4
```

#### 东家（右侧）

- 区域：`x=554, y=136, w=71, h=210`
- 标签：`x=560, y=120`
- 牌背竖向反向堆叠：间距 `4px`

公式：

```go
x := 554
y := 241 - i*4
```

### 2.5 底牌区域

参考原版：
- 底牌展示区建议：`x=230, y=186`
- 底牌 8 张间距：`14px`
- 第 3 张可轻微抬高强调：`y=166`

最终建议：

```text
底牌区基线: x=230, y=186
第 i 张:    x=230+i*14, y=186
高亮牌:     x=230+2*14, y=166
```

用途：
- 庄家捡底时弹窗展示
- 扣底阶段切换为可点击状态

### 2.6 出牌区（桌面中央）

#### 南家出牌区
- `x = 285 - cards*7`
- `y = 244`
- 牌间距 `14px`

#### 北家出牌区
- `x = 285 - cards*7`
- `y = 130`
- 牌间距 `14px`

#### 西家出牌区
- `x = 245 - cards*13`
- `y = 192`
- 牌间距 `13px`

#### 东家出牌区
- `x = 326`
- `y = 192`
- 牌间距 `13px`

这样可以最大程度复刻原版桌面视觉重心。

### 2.7 工具栏 / 操作按钮

原版花色工具栏位于右下桌面上方：`(415, 325, 129, 29)`。

Ebitengine 建议拆成两组：

#### 主操作按钮区

```text
出牌按钮   x=360, y=430, w=84, h=30
取消按钮   x=452, y=430, w=84, h=30
提示文字   x=160, y=434, w=190, h=22
```

#### 亮主/反主花色按钮区

```text
黑桃 x=417, y=327, w=25, h=25
红桃 x=443, y=327, w=25, h=25
方块 x=468, y=327, w=25, h=25
梅花 x=493, y=327, w=25, h=25
无主 x=518, y=327, w=25, h=25
```

如果想更适合鼠标，可把按钮扩到 `32 x 32`，但逻辑落点不变。

### 2.8 信息栏位置

建议固定：

#### 左上边栏（庄家/等级）
```text
庄家图标: x=32, y=45, 20x20
己方等级: x=46, y=68, 20x20
对方等级: x=566, y=68, 20x20
```

#### 花色显示
```text
左侧主花色:  x=43,  y=88, 25x25
右侧主花色:  x=563, y=88, 25x25
```

#### 分数
```text
左中分数框:  x=85,  y=300, 56x56
右中分数框:  x=490, y=128, 56x56
```

实际实现时可以把这些做成 HUD 常量：

```go
const (
    HUDDealerLeftX  = 32
    HUDDealerLeftY  = 45
    HUDTrumpLeftX   = 43
    HUDTrumpLeftY   = 88
    HUDScoreLeftX   = 85
    HUDScoreLeftY   = 300
)
```

---

## 3. 渲染实现方案

### 3.1 依赖

`go.mod` 需要新增：

```go
require (
    github.com/hajimehoshi/ebiten/v2 v2.8.6
)
```

通常还会用到：
- `github.com/hajimehoshi/ebiten/v2/inpututil`
- `github.com/hajimehoshi/ebiten/v2/text/v2`
- `image/png`
- `embed`

### 3.2 资源加载方案

建议直接复用现有 PNG：
- 牌面：`/home/zxyzxy/文档/tractor/CardSources/Resources/0.png ~ 53.png`
- 牌背：`back.png / back2.png / back3.png`

第一阶段先复制一套到项目：

```text
assets/cards/faces/0.png ... 53.png
assets/cards/backs/back.png
```

然后用 `go:embed` 打包：

```go
package gui

import (
    "bytes"
    "embed"
    "fmt"
    "image/png"

    "github.com/hajimehoshi/ebiten/v2"
)

//go:embed ../../assets/cards/faces/*.png ../../assets/cards/backs/*.png ../../assets/bg/*.png
var assetsFS embed.FS

type AssetStore struct {
    CardFaces map[string]*ebiten.Image
    CardBack  *ebiten.Image
    TableBG   *ebiten.Image
}

func loadPNG(path string) (*ebiten.Image, error) {
    raw, err := assetsFS.ReadFile(path)
    if err != nil {
        return nil, err
    }
    img, err := png.Decode(bytes.NewReader(raw))
    if err != nil {
        return nil, err
    }
    return ebiten.NewImageFromImage(img), nil
}
```

### 3.3 牌面资源映射

当前 Go 项目 `Card` 结构为：`Suit + Rank + Copy`。

建议建立一个稳定映射函数：

```go
func SpriteKey(c Card) string {
    // 与 Copy 无关，两个副本共用同一张图片
    // 例如: spade_A, heart_10, joker_small, joker_big
}
```

再在 `assets/manifest.json` 中定义：

```json
{
  "spade_3": "0.png",
  "spade_4": "1.png",
  "...": "...",
  "joker_small": "52.png",
  "joker_big": "53.png",
  "back_default": "back.png"
}
```

> 不建议在代码里硬编码 `0.png=黑桃3` 这种魔法数字；资源映射表单独维护，后续换牌面时不用改代码。

### 3.4 每张牌的绘制参数

建议常量：

```go
const (
    CardW = 71
    CardH = 96

    HandLiftY       = 20
    HoverOutlinePx  = 2
    ShadowOffsetY   = 3
    TrickCardGap    = 14
    HandCardMinGap  = 10
    HandCardBaseGap = 18
)
```

绘制规则：
- 红桃/方块文字色：`#D83A3A`
- 黑桃/梅花文字色：`#F5F5F5`
- 主牌高亮外框：`#FFD54A`
- 当前鼠标 hover 外框：`#4FC3F7`
- 选中牌上浮：`-20px`
- 非法选择临时闪红：`#FF6B6B` 180ms

### 3.5 单张牌 Draw 示例

```go
func (g *GUI) drawCard(dst *ebiten.Image, card Card, x, y float64, selected, hover, isTrump bool) {
    img := g.assets.LookupCard(card)

    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(x, y)
    dst.DrawImage(img, op)

    if isTrump {
        g.drawRectStroke(dst, x-1, y-1, CardW+2, CardH+2, 2, colorTrump)
    }
    if hover {
        g.drawRectStroke(dst, x-3, y-3, CardW+6, CardH+6, 2, colorHover)
    }
    if selected {
        g.drawSelectionMarker(dst, x+CardW/2-6, y-14)
    }
}
```

### 3.6 背景绘制

建议背景分 3 层：

1. **纯色底**：深绿 `#0E5A2A`
2. **桌布纹理**：`assets/bg/table.png`
3. **中央高光/暗角**：用半透明渐变图覆盖

第一阶段即使没有精美背景，也可以先用：

```go
func (g *GUI) drawBackground(dst *ebiten.Image) {
    dst.Fill(color.RGBA{14, 90, 42, 255})
    if g.assets.TableBG != nil {
        op := &ebiten.DrawImageOptions{}
        op.GeoM.Translate(25, 40)
        dst.DrawImage(g.assets.TableBG, op)
    }
}
```

### 3.7 动画方案

#### 发牌动画

- 牌堆起点：`(284, 186)`
- 目标点：按四家手牌区域分别计算
- 单张耗时：`80ms`
- 插值：`easeOutCubic`
- 每发一张同时更新逻辑计数，渲染上显示移动中的牌

动画结构：

```go
type CardAnim struct {
    Card      Card
    FromX     float64
    FromY     float64
    ToX       float64
    ToY       float64
    StartedAt time.Time
    Duration  time.Duration
    Done      bool
}
```

#### 出牌动画

- 从玩家手牌当前位置飞向出牌区
- 单次 `120~180ms`
- 落点到达后再从手牌区删除
- 人类出牌可 120ms，AI 出牌可 250ms，保留思考感

#### 收牌/结算动画

- 一轮结束时，把中央牌按赢家方向收拢：
  - 南 `y=300`
  - 北 `y=150`
  - 西 `x=170`
  - 东 `x=470`
- 耗时 `220ms`
- 收牌后显示“X 赢得此轮 + N 分”信息框 1500~3000ms

---

## 4. 事件处理

### 4.1 鼠标点击选牌/取消选牌

每张南家手牌维护命中框：

```go
type HitBox struct {
    CardIndex int
    Rect      image.Rectangle
}
```

注意：手牌重叠时，**必须从最后绘制的一张开始逆序命中测试**，否则点击总是落到下面那张。

```go
for i := len(g.handHitBoxes)-1; i >= 0; i-- {
    hb := g.handHitBoxes[i]
    if pointInRect(mx, my, hb.Rect) {
        g.toggleSelect(hb.CardIndex)
        break
    }
}
```

### 4.2 双击确认出牌

当前 TUI 的双击逻辑是“500ms 内同一点再次点击”。GUI 建议改成更宽松：

- 时间阈值：`300ms`
- 空间阈值：`8px`
- 同一张牌即可，不要求完全同一点像素

```go
type ClickTracker struct {
    LastAt    time.Time
    LastCard  int
    LastX     int
    LastY     int
}

func (g *GUI) isDoubleClick(cardIdx, x, y int) bool {
    return time.Since(g.click.LastAt) < 300*time.Millisecond &&
        g.click.LastCard == cardIdx &&
        abs(g.click.LastX-x) <= 8 &&
        abs(g.click.LastY-y) <= 8
}
```

交互规则：
- 若当前轮需要 1 张牌，双击单牌可直接 `ActionPlay`
- 若当前轮是跟牌且需 2/3/N 张，双击只做“确认当前已选集合”，不强制只出双击那张

### 4.3 亮主/反主时的花色选择交互

当前规则只把 `validBids[0]` 暴露给玩家，GUI 阶段建议顺手补齐为“多候选可点选”：

- 中央弹出亮主框：`x=190, y=140, w=260, h=120`
- 标题：“请选择亮主方式”
- 每个 `BidChoice` 一行按钮
- 如果只有一个选项，也仍然显示，减少分支

示例：

```go
[对级牌 黑桃] [三张级牌 红桃] [对王(无主)] [不亮]
```

### 4.4 扣底牌时的牌选择交互

扣底阶段进入 `PhaseDiscard`：
- 底部手牌仍按南家规则显示
- 顶部中央显示：“请选择 8 张底牌（已选 N/8）”
- 只有当 `len(selected)==8` 时，“扣底”按钮可点击
- 超过 8 张时，禁止继续选择并播放轻微抖动/错误音效

### 4.5 键盘辅助

尽管是 GUI，仍建议保留：
- `Enter`：确认出牌/扣底
- `Esc`：取消当前选择
- `B`/`P`：亮主/不亮（兼容 TUI 用户习惯）
- `Ctrl+Q`：退出

---

## 5. 与现有 `game.go` 的集成

### 5.1 需要修改的引用点

根据当前代码，`game.go` 至少这些位置要替换：

1. `g.tui.dealCounts = [4]int{}`
2. `g.tui.SetPhase(...)`
3. `g.tui.SleepForRedraw(...)`
4. `g.tui.bidOptions = validBids`
5. `g.tui.SetMessage(...)`
6. `g.tui.WaitForAction()`
7. `g.tui.discardCount = 8`
8. `g.tui.waitingForHuman = true/false`
9. `g.tui.thinkingPos / g.tui.thinking`
10. `g.tui.selected = make(map[int]bool)`
11. `g.tui.cursorIdx = 0`
12. `g.tui.trickWinner / trickPoints`
13. `RunTUI(tui *TUI)`

建议把这些合并成对 `ui.GameUI` 的调用，再把附加显示状态放入 `TableView`。

### 5.2 核心改造示例

#### `Game` 结构

```go
type Game struct {
    Players     [4]*Player
    Deck        []Card
    BottomCards []Card
    TrumpSuit   Suit
    Level       [2]Rank
    Dealer      PlayerPosition
    CurrentBid  *Bid
    Phase       GamePhase

    CurrentTrick *Trick
    TrickCount   int
    TrickWinner  PlayerPosition
    TeamScore    [2]int

    rng       *rand.Rand
    ui        ui.GameUI
    drawOrder []Card
}
```

#### 入口改造

```go
func main() {
    uiMode := flag.String("ui", "tui", "ui mode: tui|gui")
    flag.Parse()

    game := NewGame()

    var gameUI ui.GameUI
    switch *uiMode {
    case "gui":
        gameUI = gui.New(game)
    case "tui":
        gameUI = tui.New(game)
    default:
        log.Fatalf("unknown ui mode: %s", *uiMode)
    }

    if err := gameUI.Init(); err != nil {
        log.Fatal(err)
    }
    defer gameUI.Close()

    game.Run(gameUI)
}
```

#### 新的 `Run`

```go
func (g *Game) Run(gameUI ui.GameUI) {
    g.ui = gameUI
    g.ui.SetPhase(ui.PhaseWelcome)
    g.ui.ShowMessage("升级（拖拉机）纸牌游戏\n两副牌 · 2为常主 · 4人对战",
        []ui.ButtonSpec{{ID: "start", Label: "开始游戏", Enabled: true}})
    g.ui.WaitAction()
    g.ui.ClearMessage()

    for g.Phase != PhaseGameOver {
        g.DealAnimated()
        g.RunBiddingPhase()
        g.DiscardBottom()
        g.PlayHand()
    }
}
```

### 5.3 TUI 和 GUI 并存方案

推荐使用 **运行时参数**，不要用 build tags。

理由：
- 两套 UI 都依赖同一份游戏逻辑；
- 开发 GUI 时仍需能快速回退到 TUI 验证规则；
- CI 可分别跑 `-ui=tui` 和 `-ui=gui` 冒烟测试。

命令：

```bash
go run . -ui=tui
go run . -ui=gui
```

### 5.4 编译条件建议

- **不需要** `//go:build gui` / `//go:build tui`
- 只需要按目录组织包：
  - `ui/tui`
  - `ui/gui`
- `main.go` 根据 flag 实例化对应 UI

如果未来 GUI 资源体积过大，再考虑：
- 默认二进制嵌入通用资源；
- `-tags slim` 时不打包音效和高清背景。

---

## 6. 开发路线图

### 第一阶段：接口抽象 + 静态牌桌渲染 + 鼠标选牌

目标：
- 保持规则逻辑不变
- GUI 能启动
- 能看到四家位置、南家手牌、顶部状态栏
- 鼠标可选/取消选牌

文件清单：
- 新建：`ui/types.go`
- 新建：`ui/interface.go`
- 新建：`ui/snapshot.go`
- 新建：`ui/gui/ui.go`
- 新建：`ui/gui/layout.go`
- 新建：`ui/gui/render.go`
- 新建：`ui/gui/input.go`
- 新建：`ui/gui/assets.go`
- 修改：`main.go`
- 修改：`game.go`
- 可选迁移：`tui.go -> ui/tui/*.go`

阶段验收：
- `go run . -ui=gui` 可开窗
- 手动发一副测试手牌能渲染
- 点击手牌会抬起/落下

### 第二阶段：完整流程（亮主 → 扣底 → 出牌 → 记分）

目标：
- 完整打完一局
- 亮主、反主、扣底、跟牌、错牌提示、轮次结果都可走通

文件清单：
- 修改：`game.go`
- 修改：`main.go`
- 新建：`ui/gui/state.go`
- 新建：`ui/gui/modal.go`（如果需要单独拆出）
- 修改：`ui/gui/input.go`
- 修改：`ui/gui/render.go`
- 修改：`ui/gui/layout.go`

阶段验收：
- `-ui=tui` 与 `-ui=gui` 规则结果一致
- 人类玩家能正常完成亮主/扣底/出牌
- 非法出牌能弹框提示

### 第三阶段：动画 + 音效 + 优化

目标：
- 发牌、出牌、收牌动画完成
- 有基础音效
- 自适应缩放与更好的资源主题切换

文件清单：
- 新建：`ui/gui/anim.go`
- 新建：`ui/gui/audio.go`
- 修改：`ui/gui/render.go`
- 修改：`ui/gui/assets.go`
- 修改：`ui/gui/ui.go`
- 新建：`assets/audio/*.wav`
- 新建：`assets/bg/table.png`
- 新建：`assets/ui/buttons.png`

阶段验收：
- 动画不影响规则正确性
- 60 FPS 基本稳定
- 无资源缺失 panic

---

## 7. 现有资源复用方案

### 7.1 牌面资源现状

实测路径：
- `/home/zxyzxy/文档/tractor/CardSources/Resources/0.png ~ 53.png`
- `/home/zxyzxy/文档/tractor/CardSources/Resources/back.png`
- `/home/zxyzxy/文档/tractor/CardSources/Resources/back2.png`
- `/home/zxyzxy/文档/tractor/CardSources/Resources/back3.png`

数量：
- 牌面 54 张
- 牌背 3 张
- 尺寸 71×96

注意：两副牌共 108 张，但 **图片只需 54 张**，因为 `Copy` 仅区分牌实例，不区分外观。

### 7.2 如何转换/加载到 Go 项目

第一阶段最省事的做法：

```bash
mkdir -p /home/zxyzxy/文档/upgrade_poker/assets/cards/faces
mkdir -p /home/zxyzxy/文档/upgrade_poker/assets/cards/backs
cp /home/zxyzxy/文档/tractor/CardSources/Resources/[0-9]*.png /home/zxyzxy/文档/upgrade_poker/assets/cards/faces/
cp /home/zxyzxy/文档/tractor/CardSources/Resources/back*.png /home/zxyzxy/文档/upgrade_poker/assets/cards/backs/
```

然后在代码里：
- 用 `go:embed` 嵌入；
- 启动时一次性加载到 `map[string]*ebiten.Image`；
- 不在 Draw 阶段重复 decode。

### 7.3 推荐增加的资源

#### 必要
- `assets/bg/table.png`：桌布背景（590×400 或 640×480）
- `assets/ui/button_play.png`
- `assets/ui/button_cancel.png`
- `assets/ui/modal_bg.png`

#### 建议
- `assets/audio/card_place.wav`
- `assets/audio/card_select.wav`
- `assets/audio/win_trick.wav`
- `assets/audio/error.wav`

#### 可后补
- 庄家标记图标
- 主花色徽章
- 队伍得分牌匾

---

## 8. 示例代码片段

### 8.1 GUI 主体结构

```go
package gui

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type GUI struct {
    game   *Game
    assets *AssetStore

    view TableView

    selected map[int]bool
    hoverIdx int

    handHitBoxes []HitBox
    actionCh     chan UIAction

    width  int
    height int
}

func New(game *Game) *GUI {
    return &GUI{
        game:      game,
        selected:  map[int]bool{},
        hoverIdx:  -1,
        actionCh:  make(chan UIAction, 16),
        width:     640,
        height:    480,
    }
}

func (g *GUI) Init() error {
    assets, err := LoadAssets()
    if err != nil {
        return err
    }
    g.assets = assets
    ebiten.SetWindowSize(1280, 960)
    ebiten.SetWindowTitle("升级（拖拉机）- Ebitengine")
    return nil
}
```

### 8.2 `Update / Draw / Layout`

```go
func (g *GUI) Update() error {
    mx, my := ebiten.CursorPosition()
    g.hoverIdx = g.hitTestHand(mx, my)

    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
        if idx := g.hitTestHand(mx, my); idx >= 0 {
            g.onCardClick(idx, mx, my)
        } else if btn := g.hitTestButton(mx, my); btn != "" {
            g.onButtonClick(btn)
        }
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
        g.submitSelection()
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
        g.clearSelection()
    }
    return nil
}

func (g *GUI) Draw(screen *ebiten.Image) {
    g.drawBackground(screen)
    g.drawHUD(screen)
    g.drawNorth(screen)
    g.drawWest(screen)
    g.drawEast(screen)
    g.drawTrickArea(screen)
    g.drawSouth(screen)
    g.drawButtons(screen)
    g.drawMessage(screen)
}

func (g *GUI) Layout(outsideWidth, outsideHeight int) (int, int) {
    return 640, 480
}
```

### 8.3 渲染南家手牌

```go
func (g *GUI) drawSouth(screen *ebiten.Image) {
    cards := g.view.Players[PositionSouth].HandCards
    if len(cards) == 0 {
        return
    }

    spacing := HandCardBaseGap
    if len(cards) > 1 {
        maxSpacing := (560 - CardW) / (len(cards) - 1)
        if spacing > maxSpacing {
            spacing = max(HandCardMinGap, maxSpacing)
        }
    }

    startX := 30 + (560-(CardW+(len(cards)-1)*spacing))/2
    baseY := 375

    g.handHitBoxes = g.handHitBoxes[:0]
    for i, c := range cards {
        x := float64(startX + i*spacing)
        y := float64(baseY)
        if g.selected[i] {
            y -= HandLiftY
        }
        g.drawCard(screen, c, x, y, g.selected[i], g.hoverIdx == i, IsTrump(c, g.view.TrumpSuit, g.view.DealerLevel))
        g.handHitBoxes = append(g.handHitBoxes, HitBox{
            CardIndex: i,
            Rect: image.Rect(int(x), int(y), int(x)+CardW, int(y)+CardH),
        })
    }
}
```

---

## 9. 关键实现细节与坑位

### 9.1 不要让 GUI 直接操作 `player.Hand`

GUI 只能操作“选中索引”，真正出牌仍由 `game.go` 校验。
否则会把规则错误变成 UI bug，很难查。

### 9.2 `drawOrder` 必须统一

当前 `game.go` 用 `g.drawOrder[idx]` 把 UI 选中的索引映射回实际牌对象。GUI 也必须沿用同一排序规则，否则：
- 你点的是第 7 张，游戏拿到的是另一张；
- 尤其主牌/副牌重排时非常容易错。

建议把排序构造统一封装：

```go
func (g *Game) BuildDrawOrder(pos PlayerPosition) []Card
```

TUI 和 GUI 都只调用它。

### 9.3 命中测试必须逆序

重叠手牌如果正序检测，会永远选中最底下那张牌。这个坑在纸牌 GUI 里非常常见。

### 9.4 `Render` 与 `WaitAction` 分线程要谨慎

Ebitengine 的 `Update/Draw` 运行在主循环中。阻塞 `WaitAction()` 时，不要卡死窗口线程。

推荐模式：
- GUI 主循环持续跑 `Update/Draw`
- 用户操作写入 `actionCh`
- `game.go` 在 goroutine 中阻塞读 `WaitAction()`

即沿用当前 TUI 的 `actionChan` 思路，但 **不要** 在 `Update()` 里做耗时逻辑。

### 9.5 动画期间的规则状态

动画只是视觉延迟，不应推迟逻辑状态更新太久。建议：
- 先更新逻辑；
- 保存“上一帧位置”用于出牌飞行动画；
- 动画结束后清理临时 sprite。

这样最不容易和规则判定打架。

---

## 10. 最终建议：按什么顺序开始写代码

推荐实际落地顺序：

1. **先抽 `ui.GameUI` 接口**，让 `game.go` 不再直接写 `TUI` 字段。
2. **保留现有 TUI 作为第一实现**，先保证编译通过。
3. 引入 Ebitengine，先画：背景 + HUD + 四家占位 + 南家牌。
4. 实现南家鼠标选牌、取消选牌、出牌按钮。
5. 接入 `WaitAction()`，让 GUI 真正驱动一轮出牌。
6. 再补亮主、扣底、结算弹窗。
7. 最后再做动画和音效。

如果按这个顺序推进，第一周就能拿到一个“能看见、能点牌、能出牌”的 GUI 骨架，而不是陷进一次性大重构里。

---

## 11. 本方案对应的最小首批提交

建议第一批提交只做这些：

1. `feat(ui): introduce GameUI interface and TableView snapshot`
2. `refactor(game): replace direct TUI dependency with GameUI`
3. `feat(gui): add ebitengine window and static table renderer`
4. `feat(gui): support south-hand mouse selection and play/cancel buttons`

做到这里，就已经能正式进入 GUI 开发闭环。
