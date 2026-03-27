package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Colors ────────────────────────────────────────────────────────────────────

var (
	purple    = lipgloss.Color("#9B59B6")
	pink      = lipgloss.Color("#FF6EB4")
	green     = lipgloss.Color("#2ECC71")
	yellow    = lipgloss.Color("#F1C40F")
	cyan      = lipgloss.Color("#00FFFF")
	orange    = lipgloss.Color("#E67E22")
	white     = lipgloss.Color("#FFFFFF")
	darkGray  = lipgloss.Color("#1A1A2E")
	mutedGray = lipgloss.Color("#888888")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(pink).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(purple).
			Padding(0, 6).
			Align(lipgloss.Center)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(cyan)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purple).
			Padding(0, 2).
			Width(44)

	resultBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
			BorderForeground(green).
			Padding(1, 4).
			Width(50)

	doneStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(green).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(green).
			Padding(0, 6)

	summaryHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(pink).
				Underline(true)

	counterStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedGray).
			Italic(true)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(darkGray).
			Background(pink).
			Padding(0, 2)

	unselectedStyle = lipgloss.NewStyle().
			Foreground(white).
			Padding(0, 2)

	typeColors = map[string]lipgloss.Color{
		"int":     cyan,
		"float64": yellow,
		"bool":    orange,
		"string":  green,
	}
)

// ── Phases ────────────────────────────────────────────────────────────────────

type phase int

const (
	phaseMenu phase = iota
	phaseAskCount
	phaseTyping
	phaseDone
)

type result struct {
	input    string
	typeName string
}

type model struct {
	phase      phase
	textInput  textinput.Model
	count      int
	current    int
	results    []result
	err        string
	width      int
	height     int
	menuCursor int
	endless    bool
}

// ── Type Detection ────────────────────────────────────────────────────────────

func detectType(s string) string {
	if _, err := strconv.Atoi(s); err == nil {
		return "int"
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return "float64"
	}
	if s == "true" || s == "false" {
		return "bool"
	}
	return "string"
}

// ── Init ──────────────────────────────────────────────────────────────────────

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "e.g. 42, 3.14, true, hello"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 40
	ti.PromptStyle = lipgloss.NewStyle().Foreground(purple)
	ti.TextStyle = lipgloss.NewStyle().Foreground(white)

	return model{
		phase:     phaseMenu,
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyUp, tea.KeyLeft:
			if m.phase == phaseMenu {
				m.menuCursor = 0
			}

		case tea.KeyDown, tea.KeyRight:
			if m.phase == phaseMenu {
				m.menuCursor = 1
			}

		case tea.KeyEnter:
			switch m.phase {

			case phaseMenu:
				if m.menuCursor == 0 {
					m.endless = true
					m.phase = phaseTyping
					m.current = 1
					m.results = []result{}
					m.textInput.SetValue("")
					m.textInput.Placeholder = "type anything..."
				} else {
					m.endless = false
					m.phase = phaseAskCount
					m.textInput.SetValue("")
					m.textInput.Placeholder = "enter a number..."
					m.textInput.Focus()
				}

			case phaseAskCount:
				val := strings.TrimSpace(m.textInput.Value())
				n, err := strconv.Atoi(val)
				if err != nil || n <= 0 {
					m.err = "⚠  Please enter a valid positive number"
					m.textInput.SetValue("")
					return m, nil
				}
				m.count = n
				m.current = 1
				m.results = []result{}
				m.err = ""
				m.phase = phaseTyping
				m.textInput.Placeholder = "type anything..."
				m.textInput.SetValue("")

			case phaseTyping:
				val := strings.TrimSpace(m.textInput.Value())
				if val == "" {
					m.err = "⚠  Input cannot be empty"
					return m, nil
				}
				if val == "q" || val == "quit" {
					m.phase = phaseDone
					m.textInput.Blur()
					return m, nil
				}
				m.err = ""
				t := detectType(val)
				m.results = append(m.results, result{input: val, typeName: t})

				if !m.endless && m.current >= m.count {
					m.phase = phaseDone
					m.textInput.Blur()
					return m, nil
				}
				m.current++
				m.textInput.SetValue("")
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (m model) center(s string) string {
	w := m.width
	if w == 0 {
		w = 100
	}
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, s)
}

func (m model) vcenter(s string) string {
	h := m.height
	if h == 0 {
		h = 40
	}
	contentHeight := strings.Count(s, "\n") + 1
	_ = contentHeight
	return lipgloss.PlaceVertical(h, lipgloss.Center, lipgloss.PlaceHorizontal(m.width, lipgloss.Center, s))
 }
 

// ── View ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	

	title := titleStyle.Render("⚡ GO TYPE DETECTOR")

	switch m.phase {

	case phaseMenu:
		opt1 := unselectedStyle.Render("  Endless Mode  ")
		opt2 := unselectedStyle.Render("  Set Number    ")
		if m.menuCursor == 0 {
			opt1 = selectedStyle.Render("  Endless Mode  ")
		} else {
			opt2 = selectedStyle.Render("  Set Number    ")
		}

		menu := lipgloss.JoinVertical(lipgloss.Center,
			title,
			"",
			labelStyle.Render("Choose a mode:"),
			"",
			opt1,
			"",
			opt2,
			"",
			helpStyle.Render("↑ ↓ to navigate  •  enter to select  •  ctrl+c to quit"),
		)
		return m.vcenter(menu)

	case phaseAskCount:
		block := lipgloss.JoinVertical(lipgloss.Center,
			title,
			"",
			labelStyle.Render("How many inputs do you want to test?"),
			"",
			inputBoxStyle.Render(m.textInput.View()),
			"",
			func() string {
				if m.err != "" {
					return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Render(m.err)
				}
				return helpStyle.Render("enter to confirm  •  ctrl+c to quit")
			}(),
		)
		return m.vcenter(block)

	case phaseTyping:
		var counter string
		if m.endless {
			counter = counterStyle.Render(fmt.Sprintf("Input #%d  —  type 'q' to finish", m.current))
		} else {
			counter = counterStyle.Render(fmt.Sprintf("Input %d of %d", m.current, m.count))
		}

		parts := []string{title, "", counter, ""}

		if len(m.results) > 0 {
			var rows strings.Builder
			for _, r := range m.results {
				col := typeColors[r.typeName]
				typeTag := lipgloss.NewStyle().
					Bold(true).
					Foreground(darkGray).
					Background(col).
					Padding(0, 1).
					Render(r.typeName)
				row := fmt.Sprintf("%-22s  %s", r.input, typeTag)
				rows.WriteString(lipgloss.NewStyle().Foreground(white).Render(row))
				rows.WriteString("\n")
			}
			parts = append(parts, resultBoxStyle.Render(rows.String()), "")
		}

		parts = append(parts, labelStyle.Render("Enter something:"), "")
		parts = append(parts, inputBoxStyle.Render(m.textInput.View()), "")

		if m.err != "" {
			parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Render(m.err), "")
		} else {
			if m.endless {
				parts = append(parts, helpStyle.Render("enter to submit  •  type 'q' to finish  •  ctrl+c to quit"))
			} else {
				parts = append(parts, helpStyle.Render("enter to submit  •  ctrl+c to quit"))
			}
		}

		block := lipgloss.JoinVertical(lipgloss.Center, parts...)
		return m.vcenter(block)
	
	
	case phaseDone:
		var rows strings.Builder
		for i, r := range m.results {
			col := typeColors[r.typeName]
			typeTag := lipgloss.NewStyle().
				Bold(true).
				Foreground(darkGray).
				Background(col).
				Padding(0, 1).
				Render(r.typeName)
			num := lipgloss.NewStyle().Foreground(mutedGray).Render(fmt.Sprintf("%2d.", i+1))
			row := fmt.Sprintf("%s  %-22s  %s", num, r.input, typeTag)
			rows.WriteString(lipgloss.NewStyle().Foreground(white).Render(row))
			rows.WriteString("\n")
		}

		block := lipgloss.JoinVertical(lipgloss.Center,
			doneStyle.Render("✅  All done!"),
			"",
			summaryHeaderStyle.Render("── Results Summary ──"),
			"",
			resultBoxStyle.Render(rows.String()),
			"",
			helpStyle.Render("ctrl+c to exit"),
		)
		return m.vcenter(block)
	}

	
  return ""
}




		// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}



