package layout

import (
	"github.com/rivo/tview"
)

// Header displays logo, context info, and resource info
type Header struct {
	flex     *tview.Flex
	logo     *tview.TextView
	context  *tview.TextView
	resource *tview.TextView
}

// NewHeader creates a new header component
func NewHeader() *Header {
	h := &Header{
		flex:     tview.NewFlex(),
		logo:     NewLogo(),
		context:  tview.NewTextView().SetDynamicColors(true),
		resource: tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight),
	}

	// Horizontal layout: Logo | Context | Resource (all proportional for responsiveness)
	// Logo gets 2 parts, Context gets 3 parts (most flexible), Resource gets 1 part
	h.flex.SetDirection(tview.FlexColumn).
		AddItem(h.logo, 0, 2, false).
		AddItem(h.context, 0, 3, false).
		AddItem(h.resource, 0, 1, false)

	return h
}

// SetContext updates the context info (node name, status)
func (h *Header) SetContext(nodeName string, status string) {
	color := "[green]"
	if status != "connected" && status != "ok" && status != "OK" {
		color = "[red]"
	}
	h.context.SetText("  [yellow]Node:[white] " + nodeName + "  " + color + status)
}

// SetResource updates the resource info
func (h *Header) SetResource(info string) {
	h.resource.SetText(info + "  ")
}

// Primitive returns the header's tview primitive
func (h *Header) Primitive() tview.Primitive {
	return h.flex
}
