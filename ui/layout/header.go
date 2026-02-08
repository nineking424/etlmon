package layout

import (
	"fmt"

	"github.com/etlmon/etlmon/ui/theme"
	"github.com/rivo/tview"
)

// Header displays node context and resource info in a bordered box
type Header struct {
	flex     *tview.Flex
	context  *tview.TextView
	resource *tview.TextView
}

// NewHeader creates a new header component
func NewHeader() *Header {
	h := &Header{
		flex:     tview.NewFlex(),
		context:  tview.NewTextView().SetDynamicColors(true),
		resource: tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight),
	}

	// Horizontal layout: Context (left) | Resource (right)
	h.flex.SetDirection(tview.FlexColumn).
		AddItem(h.context, 0, 1, false).
		AddItem(h.resource, 0, 1, false)

	// Bordered box with ETLMON title - matches content box style
	h.flex.SetBorder(true).
		SetTitle(" ETLMON ").
		SetTitleAlign(tview.AlignLeft).
		SetTitleColor(theme.FgAccent).
		SetBorderColor(theme.FgLabel)

	return h
}

// SetContext updates the context info (node name, status)
func (h *Header) SetContext(nodeName string, status string) {
	statusColor := "[green]"
	statusIcon := "[green]\u25cf[-]"
	if status != "connected" && status != "ok" && status != "OK" {
		statusColor = "[red]"
		statusIcon = "[red]\u25cf[-]"
	}
	h.context.SetText(fmt.Sprintf(" [teal]Node:[-] [white]%s[-]  %s %s%s[-]", nodeName, statusIcon, statusColor, status))
}

// SetResource updates the resource info
func (h *Header) SetResource(info string) {
	h.resource.SetText(fmt.Sprintf("[teal]%s[-] ", info))
}

// Primitive returns the header's tview primitive
func (h *Header) Primitive() tview.Primitive {
	return h.flex
}
