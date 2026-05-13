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
	ActionStart      ActionType = "start"
	ActionRestart    ActionType = "restart"
	ActionPlay       ActionType = "play"
	ActionCancel     ActionType = "cancel"
	ActionBid        ActionType = "bid"
	ActionPass       ActionType = "pass"
	ActionConfirm    ActionType = "confirm"
	ActionQuit       ActionType = "quit"
	ActionTimeout    ActionType = "timeout"
	ActionHint       ActionType = "hint"
	ActionSelectBid  ActionType = "select_bid"
)

type UIAction struct {
	Type    ActionType
	CardIdx []int
	BidType string
	BidSuit string
}

type ButtonSpec struct {
	ID      string
	Label   string
	Enabled bool
}

type BidChoice struct {
	Type string
	Suit string
	Text string
}

type CardView struct {
	Label         string
	Suit          string
	Rank          string
	SuitNum       int // 0=♠ 1=♥ 2=♦ 3=♣ 4=N/A
	RankNum       int // 2-14, 15=SmallJoker, 16=BigJoker
	FaceUp        bool
	Trump         bool
	EffectiveSuit string
	IsTractor     bool
}

type PlayerView struct {
	Position    string
	Name        string
	IsHuman     bool
	HandCount   int
	HandCards   []CardView
	PlayedCards []CardView
	IsDealer    bool
	IsThinking  bool
}

type TableView struct {
	Phase             UIPhase
	TrumpSuit         string
	Dealer            string
	DealerLevel       string
	OpponentLevel     string
	TeamScore         [2]int
	TrickCount        int
	BottomCards       []CardView
	CurrentTrick      map[string][]CardView
	Message           string
	Buttons           []ButtonSpec
	BidChoices        []BidChoice
	Players           [4]PlayerView
	WaitingForHuman   bool
	SelectedIdx       map[int]bool
	NeedSelect        int
	DiscardCount      int
	TrickWinner       string
	TrickPoints       int
	DealCounts        [4]int
	HintCardIdx       []int
	CanHint           bool
	UpdatedAt         time.Time
}
