package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type spinnerModel struct {
	spinner  spinner.Model
	label    string
	done     bool
	err      error
	resultFn func() error
}

type doneMsg struct{ err error }

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			return doneMsg{err: m.resultFn()}
		},
	)
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case doneMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("  %s %s\n",
		m.spinner.View(),
		StyleStep.Render(m.label),
	)
}

// RunWithSpinner exécute fn avec un spinner animé, retourne l'erreur éventuelle
func RunWithSpinner(label string, fn func() error) error {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorOrange)

	m := spinnerModel{
		spinner:  s,
		label:    label,
		resultFn: fn,
	}

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return err
	}

	if final, ok := result.(spinnerModel); ok {
		return final.err
	}
	return nil
}

// PrintStep affiche une ligne d'étape en cours
func PrintStep(label string) {
	fmt.Println(StyleStep.Render("  → " + label))
}

// PrintSuccess affiche un message de succès
func PrintSuccess(label string) {
	fmt.Println(StyleSuccess.Render("  ✓ " + label))
}

// PrintError affiche un message d'erreur
func PrintError(label string) {
	fmt.Println(StyleError.Render("  ✗ " + label))
}

// PrintDivider affiche une ligne séparatrice
func PrintDivider() {
	fmt.Println(StyleDivider.Render("  ────────────────────────────────────"))
}

// Blank ligne vide
func Blank() {
	fmt.Println()
}

// Simulate steps with a tiny pause for premium feel
func Step(label string) {
	fmt.Println(StyleStep.Render("  → " + label))
	time.Sleep(120 * time.Millisecond)
}
