package ui

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/TechnoRed2026/redscope/internal/netmon"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Monitor interface {
	Snapshot(context.Context) netmon.Snapshot
}

type themeMode string

const (
	modeDark  themeMode = "dark"
	modeLight themeMode = "light"
)

type palette struct {
	bg, panel, panel2, header, line, brand, good, signal, warn, muted, text tcell.Color
}

var (
	darkPalette = palette{
		bg:     tcell.NewHexColor(0x111827),
		panel:  tcell.NewHexColor(0x182233),
		panel2: tcell.NewHexColor(0x223047),
		header: tcell.NewHexColor(0x33233a),
		line:   tcell.NewHexColor(0x3b4658),
		brand:  tcell.NewHexColor(0xf472b6),
		good:   tcell.NewHexColor(0x6ee7b7),
		signal: tcell.NewHexColor(0x93c5fd),
		warn:   tcell.NewHexColor(0xfbbf24),
		muted:  tcell.NewHexColor(0x94a3b8),
		text:   tcell.NewHexColor(0xe5e7eb),
	}
	lightPalette = palette{
		bg:     tcell.NewHexColor(0xe9eef5),
		panel:  tcell.NewHexColor(0xf8fafc),
		panel2: tcell.NewHexColor(0xe2e8f0),
		header: tcell.NewHexColor(0xf6dbea),
		line:   tcell.NewHexColor(0xcbd5e1),
		brand:  tcell.NewHexColor(0x9d174d),
		good:   tcell.NewHexColor(0x1a7f37),
		signal: tcell.NewHexColor(0x0969da),
		warn:   tcell.NewHexColor(0x9a6700),
		muted:  tcell.NewHexColor(0x57606a),
		text:   tcell.NewHexColor(0x24292f),
	}
	cBg, cPanel, cPanel2, cHeader, cLine         = darkPalette.bg, darkPalette.panel, darkPalette.panel2, darkPalette.header, darkPalette.line
	cBrand, cGood, cSignal, cWarn, cMuted, cText = darkPalette.brand, darkPalette.good, darkPalette.signal, darkPalette.warn, darkPalette.muted, darkPalette.text
)

func applyPalette(mode themeMode) {
	p := darkPalette
	if mode == modeLight {
		p = lightPalette
	}
	cBg, cPanel, cPanel2, cHeader, cLine = p.bg, p.panel, p.panel2, p.header, p.line
	cBrand, cGood, cSignal, cWarn, cMuted, cText = p.brand, p.good, p.signal, p.warn, p.muted, p.text
	tview.Styles = tview.Theme{
		PrimitiveBackgroundColor:    cBg,
		ContrastBackgroundColor:     cPanel,
		MoreContrastBackgroundColor: cPanel2,
		BorderColor:                 cLine,
		TitleColor:                  cBrand,
		GraphicsColor:               cLine,
		PrimaryTextColor:            cText,
		SecondaryTextColor:          cMuted,
		TertiaryTextColor:           cMuted,
		InverseTextColor:            cPanel,
		ContrastSecondaryTextColor:  cMuted,
	}
}

func Run(ctx context.Context, monitor Monitor, refreshEvery time.Duration) error {
	applyPalette(modeDark)
	app := tview.NewApplication()

	title := tview.NewTextView().SetDynamicColors(true)
	status := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)
	help := tview.NewTextView().SetDynamicColors(true)
	table := tview.NewTable().
		SetFixed(1, 0).
		SetSelectable(true, false).
		SetSeparator(tview.Borders.Vertical)
	filter := tview.NewInputField().
		SetLabel(" filter: ").
		SetPlaceholder("process, pid, host...")

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 2, 0, false).
		AddItem(table, 0, 1, true).
		AddItem(filter, 1, 0, false).
		AddItem(status, 1, 0, false).
		AddItem(help, 1, 0, false)
	state := &screenState{table: table, status: status, title: title, help: help, filter: filter, root: root, theme: modeDark}
	state.applyTheme()
	state.render(netmon.Snapshot{})

	filter.SetChangedFunc(func(text string) { state.render(state.last) })
	filter.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEsc || key == tcell.KeyEnter {
			app.SetFocus(table)
		}
	})

	refresh := func() {
		state.busy = true
		state.paintTitle()
		go func() {
			snap := monitor.Snapshot(ctx)
			app.QueueUpdateDraw(func() {
				state.busy = false
				state.render(snap)
			})
		}()
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if app.GetFocus() == filter {
			if event.Key() == tcell.KeyEsc {
				filter.SetText("")
				app.SetFocus(table)
				state.render(state.last)
				return nil
			}
			return event
		}

		switch event.Rune() {
		case 'q':
			app.Stop()
			return nil
		case 'r':
			refresh()
			return nil
		case '/':
			app.SetFocus(filter)
			return nil
		case 't', 'T':
			state.toggleTheme()
			return nil
		}
		if event.Key() == tcell.KeyEsc {
			filter.SetText("")
			state.render(state.last)
			return nil
		}
		return event
	})

	go func() {
		refresh()
		ticker := time.NewTicker(refreshEvery)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				app.Stop()
				return
			case <-ticker.C:
				refresh()
			}
		}
	}()

	return app.SetRoot(root, true).EnableMouse(true).Run()
}

type screenState struct {
	table  *tview.Table
	status *tview.TextView
	title  *tview.TextView
	help   *tview.TextView
	filter *tview.InputField
	root   *tview.Flex
	last   netmon.Snapshot
	theme  themeMode
	busy   bool
}

func hex(c tcell.Color) string { return fmt.Sprintf("#%06x", c.Hex()) }

func (s *screenState) applyTheme() {
	applyPalette(s.theme)
	s.root.SetBackgroundColor(cBg)
	s.title.SetTextColor(cText)
	s.title.SetBackgroundColor(cBg)
	s.status.SetTextColor(cText)
	s.status.SetBackgroundColor(cBg)
	s.help.SetTextColor(cText)
	s.help.SetBackgroundColor(cBg)
	s.help.SetText(fmt.Sprintf(" [%s::b]/[-::-]filter  [%s::b]r[-] refresh  [%s::b]t[-] theme:%s  [%s::b]q[-] quit  [%s::b]Esc[-] clear", hex(cBrand), hex(cBrand), hex(cBrand), s.theme, hex(cBrand), hex(cBrand)))
	s.table.SetBackgroundColor(cPanel)
	s.table.SetBordersColor(cLine).
		SetSelectedStyle(tcell.StyleDefault.Foreground(cText).Background(cPanel2).Bold(true))
	s.filter.SetLabelColor(cMuted).
		SetFieldTextColor(cText).
		SetFieldBackgroundColor(cPanel2).
		SetPlaceholderTextColor(cMuted)
	s.filter.SetBackgroundColor(cBg)
}

func (s *screenState) toggleTheme() {
	if s.theme == modeDark {
		s.theme = modeLight
	} else {
		s.theme = modeDark
	}
	s.applyTheme()
	s.render(s.last)
}

func (s *screenState) paintTitle() {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(" [%s::b]» RedScope[-::-]  [%s]network radar · %s", hex(cBrand), hex(cMuted), s.theme))
	if s.busy {
		b.WriteString(fmt.Sprintf("  [%s::b]●[%s] live", hex(cSignal), hex(cMuted)))
	} else {
		b.WriteString(fmt.Sprintf("  [%s]· idle", hex(cMuted)))
	}
	s.title.SetText(b.String())
}

func (s *screenState) render(snap netmon.Snapshot) {
	s.last = snap
	s.table.Clear()

	headers := []string{"Process", "PID", "Proto", "Local", "Remote", "Host", "State"}
	widths := []int{18, 7, 5, 22, 22, 26, 16}
	for col, text := range headers {
		s.table.SetCell(0, col,
			tview.NewTableCell(pad(text, widths[col])).
				SetTextColor(cBrand).
				SetBackgroundColor(cHeader).
				SetAttributes(tcell.AttrBold).
				SetAlign(tview.AlignLeft).
				SetSelectable(false).
				SetExpansion(0).
				SetMaxWidth(widths[col]))
	}

	query := strings.ToLower(strings.TrimSpace(s.filter.GetText()))
	row := 1
	for _, e := range snap.Entries {
		if query != "" && !strings.Contains(strings.ToLower(searchText(e)), query) {
			continue
		}
		cells := []struct {
			text  string
			color tcell.Color
		}{
			{e.Process, cText},
			{fmt.Sprint(e.PID), cMuted},
			{e.Protocol, cMuted},
			{e.Local, cText},
			{fmt.Sprintf("%s:%d", e.RemoteIP, e.RemotePort), cSignal},
			{e.Host, cMuted},
			{stateLabel(e.State), stateColor(e.State)},
		}
		bg := cPanel
		if row%2 == 0 {
			bg = cPanel2
		}
		for col, c := range cells {
			s.table.SetCell(row, col,
				tview.NewTableCell(pad(c.text, widths[col])).
					SetTextColor(c.color).
					SetBackgroundColor(bg).
					SetExpansion(0).
					SetMaxWidth(widths[col]))
		}
		row++
	}

	// Status row.
	var st strings.Builder
	st.WriteString(fmt.Sprintf("[%s::b]%4d[-::-] [%s]shown   [%s::b]%4d[-] [%s]total",
		hex(cGood), row-1, hex(cMuted), hex(cMuted), len(snap.Entries), hex(cMuted)))
	if snap.Warning != "" {
		st.WriteString(fmt.Sprintf("   [%s::b]![-::-] [%s]%s", hex(cWarn), hex(cWarn), snap.Warning))
	}
	s.status.SetText(st.String())
	s.paintTitle()
}

func searchText(e netmon.Entry) string {
	return strings.Join([]string{e.Process, fmt.Sprint(e.PID), e.Protocol, e.Local, e.RemoteIP, e.Host, e.State}, " ")
}

func pad(s string, n int) string {
	if s == "" {
		s = "-"
	}
	w := utf8.RuneCountInString(s)
	if w > n {
		r := []rune(s)
		if n > 1 {
			return string(append(r[:n-1], '…'))
		}
		return string(r[:n])
	}
	return s + strings.Repeat(" ", n-w)
}

func stateLabel(state string) string {
	switch strings.ToLower(state) {
	case "established":
		return "ESTABLISHED"
	case "listen":
		return "LISTEN"
	case "time_wait":
		return "TIME-WAIT"
	case "close_wait":
		return "CLOSE-WAIT"
	case "closed":
		return "CLOSED"
	case "", "none":
		return "-"
	default:
		return strings.ToUpper(state)
	}
}

func stateColor(state string) tcell.Color {
	switch strings.ToLower(state) {
	case "established":
		return cGood
	case "listen":
		return cSignal
	case "time_wait", "close_wait", "closed":
		return cMuted
	default:
		return cText
	}
}
