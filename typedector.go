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
	purple   = lipgloss.Color("#9B59B6")
	pink     = lipgloss.Color("#FF6EB4")
	green    = lipgloss.Color("#2ECC71")
	yellow   = lipgloss.Color("#F1C40F")
	cyan     = lipgloss.Color("#00FFFF")
	orange   = lipgloss.Color("#E67E22")
	white    = lipgloss.Color("#FFFFFF")
	darkGray = lipgloss.Color("#1A1A2E")
	mutedGray = lipgloss.Color("#888888")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(pink).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(purple).
			Padding(0, 6).
			Align(lipgloss.Center)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedGray).
			Italic(true).
			Align(lipgloss.Center)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(cyan).
			Align(lipgloss.Center)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purple).
			Padding(0, 2).
			Width(44).
			Align(lipgloss.Center)

	resultBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(green).
			Padding(1, 4).
			Width(50).
			Align(lipgloss.Center)

	doneStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(green).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(green).
			Padding(0, 6).
			Align(lipgloss.Center)

	summaryHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(pink).
				Underline(true).
				Align(lipgloss.Center)

	counterStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true).
			Align(lipgloss.Center)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedGray).
			Italic(true).
			Align(lipgloss.Center)

	typeColors = map[string]lipgloss.Color{
		"int":     cyan,
		"float64": yellow,
		"bool":    orange,
		"string":  green,
	}
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
	width     int
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
	case tea.WindowSizeMsg:
		m.width = msg.Width

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

// ── Center helper ─────────────────────────────────────────────────────────────

func (m model) center(s string) string {
	if m.width == 0 {
		return lipgloss.PlaceHorizontal(100, lipgloss.Center, s)
	}
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, s)
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	var b strings.Builder

	b.WriteString("\n\n")
	b.WriteString(m.center(titleStyle.Render("⚡ GO TYPE DETECTOR")))
	b.WriteString("\n\n")

	switch m.phase {

	case phaseAskCount:
		b.WriteString(m.center(labelStyle.Render("How many inputs do you want to test?")))
		b.WriteString("\n\n")
		b.WriteString(m.center(inputBoxStyle.Render(m.textInput.View())))
		b.WriteString("\n\n")
		if m.err != "" {
			errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Align(lipgloss.Center)
			b.WriteString(m.center(errStyle.Render(m.err)))
			b.WriteString("\n\n")
		}
		b.WriteString(m.center(helpStyle.Render("enter to confirm  •  ctrl+c to quit")))

	case phaseTyping:
		counter := fmt.Sprintf("Input %d of %d", m.current, m.count)
		b.WriteString(m.center(counterStyle.Render(counter)))
		b.WriteString("\n\n")

		if len(m.results) > 0 {
			var rows strings.Builder
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

				row := fmt.Sprintf("%-22s  %s", r.input, typeTag)
				rows.WriteString(lipgloss.NewStyle().Foreground(white).Render(row))
				rows.WriteString("\n")
			}
			b.WriteString(m.center(resultBoxStyle.Render(rows.String())))
			b.WriteString("\n\n")
		}

		b.WriteString(m.center(labelStyle.Render("Enter something:")))
		b.WriteString("\n\n")
		b.WriteString(m.center(inputBoxStyle.Render(m.textInput.View())))
		b.WriteString("\n\n")

		if m.err != "" {
			errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))
			b.WriteString(m.center(errStyle.Render(m.err)))
			b.WriteString("\n\n")
		}
		b.WriteString(m.center(helpStyle.Render("enter to submit  •  ctrl+c to quit")))

	case phaseDone:
		b.WriteString(m.center(doneStyle.Render("✅  All done!")))
		b.WriteString("\n\n")
		b.WriteString(m.center(summaryHeaderStyle.Render("── Results Summary ──")))
		b.WriteString("\n\n")

		var rows strings.Builder
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
			row := fmt.Sprintf("%s  %-22s  %s", num, r.input, typeTag)
			rows.WriteString(lipgloss.NewStyle().Foreground(white).Render(row))
			rows.WriteString("\n")
		}
		b.WriteString(m.center(resultBoxStyle.Render(rows.String())))
		b.WriteString("\n\n")
		b.WriteString(m.center(helpStyle.Render("ctrl+c to exit")))
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




