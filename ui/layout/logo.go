package layout

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ASCII art logo
const LogoArt = ` _____ _____ _     __  __  ___  _   _
| ____|_   _| |   |  \/  |/ _ \| \ | |
|  _|   | | | |   | |\/| | | | |  \| |
| |___  | | | |___| |  | | |_| | |\  |
|_____| |_| |_____|_|  |_|\___/|_| \_|`

// NewLogo creates a new logo text view
func NewLogo() *tview.TextView {
	tv := tview.NewTextView().
		SetText(LogoArt).
		SetTextColor(tcell.ColorAqua).
		SetDynamicColors(false)
	return tv
}
