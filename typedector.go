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

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	purple     = lipgloss.Color("#9B59B6")
	pink       = lipgloss.Color("#FF6EB4")
	green      = lipgloss.Color("#2ECC71")
	yellow     = lipgloss.Color("#F1C40F")
	cyan       = lipgloss.Color("#00FFFF")
	orange     = lipgloss.Color("#E67E22")
	white      = lipgloss.Color("#FFFFFF")
	darkGray   = lipgloss.Color("#2D2D2D")
	mutedGray  = lipgloss.Color("#888888")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(pink).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(purple).
			Padding(0, 4).
			Align(lipgloss.Center)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedGray).
			Italic(true)

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
			Padding(0, 2).
			Width(44)

	typeColors = map[string]lipgloss.Color{
		"int":     cyan,
		"float64": yellow,
		"bool":    orange,
		"string":  green,
	}

	helpStyle = lipgloss.NewStyle().Foreground(mutedGray).Italic(true)

	counterStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true)

	doneStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(green).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(green).
			Padding(0, 4).
			Align(lipgloss.Center)

	summaryHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(pink).
				Underline(true)

	summaryRowStyle = lipgloss.NewStyle().
			Foreground(white)
)

// ── State ─────────────────────────────────────────────────────────────────────

type phase int

const (
	phaseAskCount phase = iota
	phaseTyping
	phaseDone
)

type result struct {
	input    string
	typeName string
}

type model struct {
	phase     phase
	textInput textinput.Model
	count     int
	current   int
	results   []result
	err       string
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
	ti.Placeholder = "e.g 1 , 5 , etc"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 40
	ti.PromptStyle = lipgloss.NewStyle().Foreground(purple)
	ti.TextStyle = lipgloss.NewStyle().Foreground(white)

	return model{
		phase:     phaseAskCount,
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
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			val := strings.TrimSpace(m.textInput.Value())

			if m.phase == phaseAskCount {
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
				return m, nil
			}

			if m.phase == phaseTyping {
				if val == "" {
					m.err = "⚠  Input cannot be empty"
					return m, nil
				}
				m.err = ""
				t := detectType(val)
				m.results = append(m.results, result{input: val, typeName: t})

				if m.current >= m.count {
					m.phase = phaseDone
					m.textInput.Blur()
					return m, nil
				}
				m.current++
				m.textInput.SetValue("")
				return m, nil
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	var b strings.Builder

	// Title
	b.WriteString("\n")
	b.WriteString(lipgloss.PlaceHorizontal(50, lipgloss.Center, titleStyle.Render("⚡ GO TYPE DETECTOR")))
	b.WriteString("\n")
	
	b.WriteString("\n\n")

	switch m.phase {

	case phaseAskCount:
		b.WriteString(labelStyle.Render("  How many inputs do you want to test?"))
		b.WriteString("\n\n")
		b.WriteString(inputBoxStyle.Render(m.textInput.View()))
		b.WriteString("\n\n")
		if m.err != "" {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Render(m.err))
			b.WriteString("\n")
		}
		b.WriteString("  " + helpStyle.Render("press Enter to confirm • ctrl+c to quit"))

	case phaseTyping:
		progress := fmt.Sprintf("  Input %d of %d", m.current, m.count)
		b.WriteString(counterStyle.Render(progress))
		b.WriteString("\n\n")

		// Previous results
		for _, r := range m.results {
			col, ok := typeColors[r.typeName]
			if !ok {
				col = white
			}
			typeTag := lipgloss.NewStyle().
				Bold(true).
				Foreground(darkGray).
				Background(col).
				Padding(0, 1).
				Render(r.typeName)

			line := fmt.Sprintf("  %-20s  %s", r.input, typeTag)
			b.WriteString(summaryRowStyle.Render(line))
			b.WriteString("\n")
		}

		if len(m.results) > 0 {
			b.WriteString("\n")
		}

		b.WriteString(labelStyle.Render("  Enter something:"))
		b.WriteString("\n\n")
		b.WriteString(inputBoxStyle.Render(m.textInput.View()))
		b.WriteString("\n\n")

		if m.err != "" {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Render(m.err))
			b.WriteString("\n")
		}
		b.WriteString("  " + helpStyle.Render("press Enter to submit • ctrl+c to quit"))

	case phaseDone:
		b.WriteString(lipgloss.PlaceHorizontal(50, lipgloss.Center, doneStyle.Render("✅  All done!")))
		b.WriteString("\n\n")
		b.WriteString("  " + summaryHeaderStyle.Render("Results Summary"))
		b.WriteString("\n\n")

		for i, r := range m.results {
			col, ok := typeColors[r.typeName]
			if !ok {
				col = white
			}
			typeTag := lipgloss.NewStyle().
				Bold(true).
				Foreground(darkGray).
				Background(col).
				Padding(0, 1).
				Render(r.typeName)

			num := lipgloss.NewStyle().Foreground(mutedGray).Render(fmt.Sprintf("%2d.", i+1))
			line := fmt.Sprintf("  %s  %-20s  %s", num, r.input, typeTag)
			b.WriteString(line)
			b.WriteString("\n")
		}

		b.WriteString("\n  " + helpStyle.Render("press ctrl+c to exit"))
	}

	b.WriteString("\n\n")
	return b.String()
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}



