package textui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type ChangedMsg struct {
	Text string
}

func Change(c ChangedMsg, fn func(c ChangedMsg) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return fn(c)
	}
}

type (
	errMsg error
)

type Model struct {
	textInput textinput.Model
	err       error
}

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "La la la..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return Model{
		textInput: ti,
		err:       nil,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.change(ChangedMsg{
				Text: m.textInput.Value(),
			})
		}
	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return fmt.Sprintf(
		"Type any text here\n\n%s\n\n%s",
		m.textInput.View(),
		"(enter to apply, esc to cancel)",
	) + "\n"
}

func (m Model)SetData(data ChangedMsg) {
	m.textInput.SetValue(data.Text)
}

func (m Model) change(c ChangedMsg) tea.Cmd {
	return Change(c, func(_ ChangedMsg) tea.Msg {
		return c
	})
}