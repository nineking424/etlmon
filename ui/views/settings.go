package views

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/etlmon/etlmon/internal/config"
	"github.com/etlmon/etlmon/ui/client"
	"github.com/etlmon/etlmon/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// settingsAPIClient interface for testing
type settingsAPIClient interface {
	GetConfig(ctx context.Context) (*config.NodeConfig, error)
	SaveConfig(ctx context.Context, cfg *config.NodeConfig) error
}

// SettingsView provides a sectioned configuration editor
type SettingsView struct {
	pages          *tview.Pages
	flex           *tview.Flex
	sectionList    *tview.List
	contentArea    *tview.Flex
	hintBar        *tview.TextView
	cfg            *config.NodeConfig
	currentSection int
	apiClient      settingsAPIClient
	tviewApp       *tview.Application
	dirty          bool

	onStatusChange func(msg string, isError bool)

	processTable *tview.Table
	logTable     *tview.Table
	pathTable    *tview.Table
}

// NewSettingsView creates a new settings view
func NewSettingsView() *SettingsView {
	v := &SettingsView{}

	// Section list (left sidebar)
	sectionList := tview.NewList().
		AddItem("Process", "Process monitoring patterns", 0, nil).
		AddItem("Logs", "Log file monitoring", 0, nil).
		AddItem("Paths", "Path scan configuration", 0, nil)
	sectionList.SetBorder(true).
		SetTitle(" Sections ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.FgLabel)
	sectionList.SetHighlightFullLine(true).
		SetSelectedBackgroundColor(theme.BgSelected).
		SetSelectedTextColor(theme.FgPrimary).
		SetMainTextColor(theme.FgSecondary).
		SetSecondaryTextColor(theme.FgMuted)
	sectionList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		v.currentSection = index
		v.showSection(index)
	})
	v.sectionList = sectionList

	// Content area
	contentArea := tview.NewFlex().SetDirection(tview.FlexRow)
	v.contentArea = contentArea

	// Create section tables
	v.processTable = v.createProcessTable()
	v.logTable = v.createLogTable()
	v.pathTable = v.createPathTable()

	// Hint bar
	hintBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	hintBar.SetText("[teal]a[silver]=add  [teal]e[silver]=edit  [teal]d[silver]=delete  [teal]s[silver]=save  [teal]Tab[silver]=switch pane  [teal]\u2191\u2193[silver]=navigate")
	hintBar.SetBackgroundColor(theme.BgStatusBar)
	v.hintBar = hintBar

	// Main layout
	mainFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(sectionList, 24, 0, true).
		AddItem(contentArea, 0, 1, false)

	v.flex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainFlex, 0, 1, true).
		AddItem(hintBar, 1, 0, false)

	// Set up input capture for main view (NOT modal)
	v.flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Pass through all keys when modal is open
		if v.IsEditing() {
			return event
		}
		// Tab: switch focus between sidebar and content
		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyBacktab {
			if v.sectionList.HasFocus() {
				v.focusContent()
			} else {
				if v.tviewApp != nil {
					v.tviewApp.SetFocus(v.sectionList)
				}
			}
			return nil
		}

		// 's' key: save (only when content table has focus)
		if event.Rune() == 's' && !v.sectionList.HasFocus() {
			v.save()
			return nil
		}

		return event
	})

	// Use Pages as root: "main" page + "modal" overlay
	v.pages = tview.NewPages().
		AddPage("main", v.flex, true, true)

	// Show initial section
	v.showSection(0)

	return v
}

// SetApp sets the tview application reference
func (v *SettingsView) SetApp(app *tview.Application) {
	v.tviewApp = app
}

// IsEditing returns true when a modal dialog is open
func (v *SettingsView) IsEditing() bool {
	return v.pages.HasPage("modal")
}

// Name returns the view name
func (v *SettingsView) Name() string {
	return "settings"
}

// Primitive returns the root tview primitive
func (v *SettingsView) Primitive() tview.Primitive {
	return v.pages
}

// Refresh loads config from the API
func (v *SettingsView) Refresh(ctx context.Context, c *client.Client) error {
	v.apiClient = c
	return v.refresh(ctx, c)
}

// Focus sets initial focus on the section list
func (v *SettingsView) Focus() {
	if v.tviewApp != nil {
		v.tviewApp.SetFocus(v.sectionList)
	}
}

// SetStatusCallback sets the status message callback
func (v *SettingsView) SetStatusCallback(cb func(msg string, isError bool)) {
	v.onStatusChange = cb
}

func (v *SettingsView) refresh(ctx context.Context, c settingsAPIClient) error {
	if v.dirty || v.IsEditing() {
		return nil
	}
	cfg, err := c.GetConfig(ctx)
	if err != nil {
		return err
	}
	v.cfg = cfg
	v.refreshProcessTable()
	v.refreshLogTable()
	v.refreshPathTable()
	return nil
}

// ----- Section display -----

func (v *SettingsView) showSection(index int) {
	v.contentArea.Clear()
	switch index {
	case 0:
		v.contentArea.AddItem(v.processTable, 0, 1, true)
	case 1:
		v.contentArea.AddItem(v.logTable, 0, 1, true)
	case 2:
		v.contentArea.AddItem(v.pathTable, 0, 1, true)
	}
}

func (v *SettingsView) focusContent() {
	if v.tviewApp == nil {
		return
	}
	switch v.currentSection {
	case 0:
		v.tviewApp.SetFocus(v.processTable)
	case 1:
		v.tviewApp.SetFocus(v.logTable)
	case 2:
		v.tviewApp.SetFocus(v.pathTable)
	}
}

// ----- Process Table -----

func (v *SettingsView) createProcessTable() *tview.Table {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)
	table.SetBorder(true).
		SetTitle(" Process Monitoring ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.FgLabel)

	table.SetCell(0, 0, tview.NewTableCell("Pattern (glob: * matches any chars)").
		SetTextColor(theme.TableHeader).
		SetAttributes(theme.TableHeaderAttr).
		SetSelectable(false).
		SetExpansion(1))

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'a':
			v.addProcessPattern()
			return nil
		case 'e':
			v.editProcessPattern()
			return nil
		case 'd':
			v.deleteProcessPattern()
			return nil
		}
		return event
	})

	return table
}

func (v *SettingsView) refreshProcessTable() {
	for i := v.processTable.GetRowCount() - 1; i > 0; i-- {
		v.processTable.RemoveRow(i)
	}

	if v.cfg == nil {
		return
	}

	if len(v.cfg.Process.Patterns) == 0 {
		v.processTable.SetCell(1, 0, tview.NewTableCell("(no patterns - monitoring all processes)").
			SetTextColor(theme.FgMuted).
			SetExpansion(1))
	} else {
		for i, pat := range v.cfg.Process.Patterns {
			v.processTable.SetCell(i+1, 0, tview.NewTableCell(pat).
				SetTextColor(theme.FgPrimary).
				SetExpansion(1))
		}
	}

	topNRow := v.processTable.GetRowCount()
	v.processTable.SetCell(topNRow, 0, tview.NewTableCell("").SetSelectable(false))
	topNRow++
	v.processTable.SetCell(topNRow, 0, tview.NewTableCell(
		fmt.Sprintf("[teal]Top N:[-] %d", v.cfg.Process.TopN)).
		SetTextColor(theme.FgSecondary).
		SetSelectable(false).
		SetExpansion(1))
}

func (v *SettingsView) addProcessPattern() {
	if v.cfg == nil || v.tviewApp == nil {
		return
	}

	form := tview.NewForm()
	v.styleForm(form, " Add Process Pattern ")

	form.AddInputField("Pattern:", "", 40, nil, nil)
	form.AddButton("Add", func() {
		if field, ok := form.GetFormItem(0).(*tview.InputField); ok {
			pattern := strings.TrimSpace(field.GetText())
			if pattern != "" {
				v.cfg.Process.Patterns = append(v.cfg.Process.Patterns, pattern)
				v.dirty = true
				v.refreshProcessTable()
				v.setStatus("Modified (press 's' to save)", false)
			}
		}
		v.dismissModal()
	})
	form.AddButton("Cancel", func() {
		v.dismissModal()
	})
	form.SetCancelFunc(func() {
		v.dismissModal()
	})

	v.showModal(form, 50, 7)
}

func (v *SettingsView) editProcessPattern() {
	if v.cfg == nil || v.tviewApp == nil {
		return
	}
	row, _ := v.processTable.GetSelection()
	idx := row - 1
	if idx < 0 || idx >= len(v.cfg.Process.Patterns) {
		return
	}

	current := v.cfg.Process.Patterns[idx]
	form := tview.NewForm()
	v.styleForm(form, " Edit Process Pattern ")

	form.AddInputField("Pattern:", current, 40, nil, nil)
	form.AddButton("Save", func() {
		if field, ok := form.GetFormItem(0).(*tview.InputField); ok {
			pattern := strings.TrimSpace(field.GetText())
			if pattern != "" {
				v.cfg.Process.Patterns[idx] = pattern
				v.dirty = true
				v.refreshProcessTable()
				v.setStatus("Modified (press 's' to save)", false)
			}
		}
		v.dismissModal()
	})
	form.AddButton("Cancel", func() {
		v.dismissModal()
	})
	form.SetCancelFunc(func() {
		v.dismissModal()
	})

	v.showModal(form, 50, 7)
}

func (v *SettingsView) deleteProcessPattern() {
	if v.cfg == nil || len(v.cfg.Process.Patterns) == 0 {
		return
	}
	row, _ := v.processTable.GetSelection()
	idx := row - 1
	if idx < 0 || idx >= len(v.cfg.Process.Patterns) {
		return
	}
	v.cfg.Process.Patterns = append(v.cfg.Process.Patterns[:idx], v.cfg.Process.Patterns[idx+1:]...)
	v.dirty = true
	v.refreshProcessTable()
	v.setStatus("Modified (press 's' to save)", false)
}

// ----- Log Table -----

func (v *SettingsView) createLogTable() *tview.Table {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)
	table.SetBorder(true).
		SetTitle(" Log Monitoring ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.FgLabel)

	headers := []string{"Name", "Path", "MaxLines"}
	for i, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(theme.TableHeader).
			SetAttributes(theme.TableHeaderAttr).
			SetSelectable(false)
		if i == 1 {
			cell.SetExpansion(1)
		}
		table.SetCell(0, i, cell)
	}

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'a':
			v.addLogEntry()
			return nil
		case 'e':
			v.editLogEntry()
			return nil
		case 'd':
			v.deleteLogEntry()
			return nil
		}
		return event
	})

	return table
}

func (v *SettingsView) refreshLogTable() {
	for i := v.logTable.GetRowCount() - 1; i > 0; i-- {
		v.logTable.RemoveRow(i)
	}

	if v.cfg == nil {
		return
	}

	if len(v.cfg.Logs) == 0 {
		v.logTable.SetCell(1, 0, tview.NewTableCell("(no log files configured)").
			SetTextColor(theme.FgMuted).
			SetExpansion(1))
		return
	}

	for i, l := range v.cfg.Logs {
		row := i + 1
		v.logTable.SetCell(row, 0, tview.NewTableCell(l.Name).
			SetTextColor(theme.FgPrimary))
		v.logTable.SetCell(row, 1, tview.NewTableCell(l.Path).
			SetTextColor(theme.FgSecondary).
			SetExpansion(1))
		v.logTable.SetCell(row, 2, tview.NewTableCell(strconv.Itoa(l.MaxLines)).
			SetTextColor(theme.FgSecondary).
			SetAlign(tview.AlignRight))
	}
}

func (v *SettingsView) addLogEntry() {
	if v.cfg == nil || v.tviewApp == nil {
		return
	}

	form := tview.NewForm()
	v.styleForm(form, " Add Log Monitor ")

	form.AddInputField("Name:", "", 40, nil, nil)
	form.AddInputField("Path:", "", 40, nil, nil)
	form.AddInputField("Max Lines:", "1000", 10, nil, nil)
	form.AddButton("Add", func() {
		nameField, ok1 := form.GetFormItem(0).(*tview.InputField)
		pathField, ok2 := form.GetFormItem(1).(*tview.InputField)
		mlField, ok3 := form.GetFormItem(2).(*tview.InputField)
		if !ok1 || !ok2 || !ok3 {
			v.dismissModal()
			return
		}
		name := strings.TrimSpace(nameField.GetText())
		logPath := strings.TrimSpace(pathField.GetText())
		mlStr := strings.TrimSpace(mlField.GetText())
		ml, err := strconv.Atoi(mlStr)
		if err != nil || ml <= 0 {
			ml = 1000
		}
		if name != "" && logPath != "" {
			v.cfg.Logs = append(v.cfg.Logs, config.LogMonitorConfig{
				Name:     name,
				Path:     logPath,
				MaxLines: ml,
			})
			v.dirty = true
			v.refreshLogTable()
			v.setStatus("Modified (press 's' to save)", false)
		}
		v.dismissModal()
	})
	form.AddButton("Cancel", func() {
		v.dismissModal()
	})
	form.SetCancelFunc(func() {
		v.dismissModal()
	})

	v.showModal(form, 55, 11)
}

func (v *SettingsView) editLogEntry() {
	if v.cfg == nil || v.tviewApp == nil {
		return
	}
	row, _ := v.logTable.GetSelection()
	idx := row - 1
	if idx < 0 || idx >= len(v.cfg.Logs) {
		return
	}

	entry := v.cfg.Logs[idx]
	form := tview.NewForm()
	v.styleForm(form, " Edit Log Monitor ")

	form.AddInputField("Name:", entry.Name, 40, nil, nil)
	form.AddInputField("Path:", entry.Path, 40, nil, nil)
	form.AddInputField("Max Lines:", strconv.Itoa(entry.MaxLines), 10, nil, nil)
	form.AddButton("Save", func() {
		nameField, ok1 := form.GetFormItem(0).(*tview.InputField)
		pathField, ok2 := form.GetFormItem(1).(*tview.InputField)
		mlField, ok3 := form.GetFormItem(2).(*tview.InputField)
		if !ok1 || !ok2 || !ok3 {
			v.dismissModal()
			return
		}
		name := strings.TrimSpace(nameField.GetText())
		logPath := strings.TrimSpace(pathField.GetText())
		mlStr := strings.TrimSpace(mlField.GetText())
		ml, err := strconv.Atoi(mlStr)
		if err != nil || ml <= 0 {
			ml = 1000
		}
		if name != "" && logPath != "" {
			v.cfg.Logs[idx] = config.LogMonitorConfig{
				Name:     name,
				Path:     logPath,
				MaxLines: ml,
			}
			v.dirty = true
			v.refreshLogTable()
			v.setStatus("Modified (press 's' to save)", false)
		}
		v.dismissModal()
	})
	form.AddButton("Cancel", func() {
		v.dismissModal()
	})
	form.SetCancelFunc(func() {
		v.dismissModal()
	})

	v.showModal(form, 55, 11)
}

func (v *SettingsView) deleteLogEntry() {
	if v.cfg == nil || len(v.cfg.Logs) == 0 {
		return
	}
	row, _ := v.logTable.GetSelection()
	idx := row - 1
	if idx < 0 || idx >= len(v.cfg.Logs) {
		return
	}
	v.cfg.Logs = append(v.cfg.Logs[:idx], v.cfg.Logs[idx+1:]...)
	v.dirty = true
	v.refreshLogTable()
	v.setStatus("Modified (press 's' to save)", false)
}

// ----- Path Table -----

func (v *SettingsView) createPathTable() *tview.Table {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)
	table.SetBorder(true).
		SetTitle(" Path Scanning ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.FgLabel)

	headers := []string{"Path", "Interval", "MaxDepth"}
	for i, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(theme.TableHeader).
			SetAttributes(theme.TableHeaderAttr).
			SetSelectable(false)
		if i == 0 {
			cell.SetExpansion(1)
		}
		table.SetCell(0, i, cell)
	}

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'a':
			v.addPathEntry()
			return nil
		case 'e':
			v.editPathEntry()
			return nil
		case 'd':
			v.deletePathEntry()
			return nil
		}
		return event
	})

	return table
}

func (v *SettingsView) refreshPathTable() {
	for i := v.pathTable.GetRowCount() - 1; i > 0; i-- {
		v.pathTable.RemoveRow(i)
	}

	if v.cfg == nil {
		return
	}

	if len(v.cfg.Paths) == 0 {
		v.pathTable.SetCell(1, 0, tview.NewTableCell("(no paths configured)").
			SetTextColor(theme.FgMuted).
			SetExpansion(1))
		return
	}

	for i, p := range v.cfg.Paths {
		row := i + 1
		v.pathTable.SetCell(row, 0, tview.NewTableCell(p.Path).
			SetTextColor(theme.FgPrimary).
			SetExpansion(1))
		v.pathTable.SetCell(row, 1, tview.NewTableCell(formatDuration(p.ScanInterval)).
			SetTextColor(theme.FgSecondary).
			SetAlign(tview.AlignRight))
		v.pathTable.SetCell(row, 2, tview.NewTableCell(strconv.Itoa(p.MaxDepth)).
			SetTextColor(theme.FgSecondary).
			SetAlign(tview.AlignRight))
	}
}

func (v *SettingsView) addPathEntry() {
	if v.cfg == nil || v.tviewApp == nil {
		return
	}

	form := tview.NewForm()
	v.styleForm(form, " Add Path ")

	form.AddInputField("Path:", "", 40, nil, nil)
	form.AddInputField("Interval (sec):", "60", 10, nil, nil)
	form.AddInputField("Max Depth:", "10", 10, nil, nil)
	form.AddButton("Add", func() {
		pathField, ok1 := form.GetFormItem(0).(*tview.InputField)
		intervalField, ok2 := form.GetFormItem(1).(*tview.InputField)
		depthField, ok3 := form.GetFormItem(2).(*tview.InputField)
		if !ok1 || !ok2 || !ok3 {
			v.dismissModal()
			return
		}
		pathStr := strings.TrimSpace(pathField.GetText())
		interval, err := strconv.Atoi(strings.TrimSpace(intervalField.GetText()))
		if err != nil || interval <= 0 {
			interval = 60
		}
		depth, err := strconv.Atoi(strings.TrimSpace(depthField.GetText()))
		if err != nil || depth <= 0 {
			depth = 10
		}
		if pathStr != "" {
			v.cfg.Paths = append(v.cfg.Paths, config.PathConfig{
				Path:         pathStr,
				ScanInterval: time.Duration(interval) * time.Second,
				MaxDepth:     depth,
				Timeout:      30 * time.Second,
			})
			v.dirty = true
			v.refreshPathTable()
			v.setStatus("Modified (press 's' to save)", false)
		}
		v.dismissModal()
	})
	form.AddButton("Cancel", func() {
		v.dismissModal()
	})
	form.SetCancelFunc(func() {
		v.dismissModal()
	})

	v.showModal(form, 55, 11)
}

func (v *SettingsView) editPathEntry() {
	if v.cfg == nil || v.tviewApp == nil {
		return
	}
	row, _ := v.pathTable.GetSelection()
	idx := row - 1
	if idx < 0 || idx >= len(v.cfg.Paths) {
		return
	}

	entry := v.cfg.Paths[idx]
	form := tview.NewForm()
	v.styleForm(form, " Edit Path ")

	form.AddInputField("Path:", entry.Path, 40, nil, nil)
	form.AddInputField("Interval (sec):", strconv.Itoa(int(entry.ScanInterval.Seconds())), 10, nil, nil)
	form.AddInputField("Max Depth:", strconv.Itoa(entry.MaxDepth), 10, nil, nil)
	form.AddButton("Save", func() {
		pathField, ok1 := form.GetFormItem(0).(*tview.InputField)
		intervalField, ok2 := form.GetFormItem(1).(*tview.InputField)
		depthField, ok3 := form.GetFormItem(2).(*tview.InputField)
		if !ok1 || !ok2 || !ok3 {
			v.dismissModal()
			return
		}
		pathStr := strings.TrimSpace(pathField.GetText())
		interval, err := strconv.Atoi(strings.TrimSpace(intervalField.GetText()))
		if err != nil || interval <= 0 {
			interval = 60
		}
		depth, err := strconv.Atoi(strings.TrimSpace(depthField.GetText()))
		if err != nil || depth <= 0 {
			depth = 10
		}
		if pathStr != "" {
			v.cfg.Paths[idx] = config.PathConfig{
				Path:         pathStr,
				ScanInterval: time.Duration(interval) * time.Second,
				MaxDepth:     depth,
				Timeout:      30 * time.Second,
			}
			v.dirty = true
			v.refreshPathTable()
			v.setStatus("Modified (press 's' to save)", false)
		}
		v.dismissModal()
	})
	form.AddButton("Cancel", func() {
		v.dismissModal()
	})
	form.SetCancelFunc(func() {
		v.dismissModal()
	})

	v.showModal(form, 55, 11)
}

func (v *SettingsView) deletePathEntry() {
	if v.cfg == nil || len(v.cfg.Paths) == 0 {
		return
	}
	row, _ := v.pathTable.GetSelection()
	idx := row - 1
	if idx < 0 || idx >= len(v.cfg.Paths) {
		return
	}
	v.cfg.Paths = append(v.cfg.Paths[:idx], v.cfg.Paths[idx+1:]...)
	v.dirty = true
	v.refreshPathTable()
	v.setStatus("Modified (press 's' to save)", false)
}

// ----- Modal helpers -----

func (v *SettingsView) styleForm(form *tview.Form, title string) {
	form.SetBorder(true).
		SetTitle(title).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(theme.FgAccent)
	form.SetFieldBackgroundColor(theme.BgSelected)
	form.SetFieldTextColor(theme.FgPrimary)
	form.SetLabelColor(theme.FgLabel)
	form.SetButtonBackgroundColor(theme.BgNavBar)
	form.SetButtonTextColor(theme.FgPrimary)
}

func (v *SettingsView) showModal(form *tview.Form, width, height int) {
	if v.tviewApp == nil {
		return
	}

	// Create centered overlay
	modal := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(form, width, 0, true).
			AddItem(nil, 0, 1, false),
			height, 0, true).
		AddItem(nil, 0, 1, false)

	// Blur current content focus before adding modal overlay.
	// Without this, both "main" and "modal" pages have HasFocus()=true
	// and tview Pages routes events to "main" (added first) instead of "modal".
	v.tviewApp.SetFocus(v.pages)
	v.pages.AddPage("modal", modal, true, true)
	v.tviewApp.SetFocus(form)
}

func (v *SettingsView) dismissModal() {
	if v.pages.HasPage("modal") {
		v.pages.RemovePage("modal")
	}
	v.focusContent()
}

// ----- Save -----

func (v *SettingsView) save() {
	if v.apiClient == nil {
		v.setStatus("No API connection", true)
		return
	}
	if v.cfg == nil {
		v.setStatus("No config loaded", true)
		return
	}

	ctx := context.Background()
	if err := v.apiClient.SaveConfig(ctx, v.cfg); err != nil {
		v.setStatus(fmt.Sprintf("Save failed: %v", err), true)
		return
	}

	v.dirty = false
	v.setStatus("Settings saved and applied", false)
}

func (v *SettingsView) setStatus(msg string, isError bool) {
	if v.onStatusChange != nil {
		v.onStatusChange(msg, isError)
	}
}

// ----- Utility -----

func formatDuration(d time.Duration) string {
	if d >= time.Hour {
		return fmt.Sprintf("%.0fh", d.Hours())
	}
	if d >= time.Minute {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	return fmt.Sprintf("%.0fs", d.Seconds())
}
