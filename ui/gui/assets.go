package gui

import (
	"fmt"
	"strings"

	baseui "github.com/smallnest/upgrade_poker/ui"
)

func cardTitle(c baseui.CardView) string {
	if c.Label != "" {
		return c.Label
	}
	parts := []string{c.Suit, c.Rank}
	return strings.TrimSpace(strings.Join(parts, ""))
}

func cardDebugKey(c baseui.CardView) string {
	return fmt.Sprintf("%s|%s|%s", c.Suit, c.Rank, cardTitle(c))
}
