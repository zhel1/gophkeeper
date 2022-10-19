package authui

// A simple example demonstrating the use of multiple text input components
// from the Bubbles component library.

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Creds struct {
	Login    string
	Password string
}

type SignInMsg Creds
type SignUpMsg Creds

func SignUp(c Creds, fn func(c Creds) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return fn(c)
	}
}

func SignIn(c Creds, fn func(c Creds) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return fn(c)
	}
}

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle.Copy()
	noStyle      = lipgloss.NewStyle()

	buttonsCount        = 2
	blurredButtons      = fmt.Sprintf("[ %s ] [ %s ]", blurredStyle.Render("Sign in"), blurredStyle.Render("Sign up"))
	focusedSignInButton = fmt.Sprintf("%s [ %s ]", focusedStyle.Copy().Render("[ Sign in ]"), blurredStyle.Render("Sign up"))
	focusedSignUpButton = fmt.Sprintf("[ %s ] %s", blurredStyle.Render("Sign in"), focusedStyle.Copy().Render("[ Sign up ]"))
)

type Model struct {
	focusIndex int
	inputs     []textinput.Model
}

func New() Model {
	m := Model{
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Login"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
			t.SetCursorMode(textinput.CursorBlink)
		case 1:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
			t.SetCursorMode(textinput.CursorBlink)
		}

		m.inputs[i] = t
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex >= len(m.inputs) {
				creds := Creds{
					Login:    m.inputs[0].Value(),
					Password: m.inputs[1].Value(),
				}

				var cmd tea.Cmd
				switch m.focusIndex {
				case len(m.inputs):
					return m, tea.Batch(m.signIn(creds), cmd)
				case len(m.inputs) + 1:
					return m, tea.Batch(m.signUp(creds), cmd)
				}
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs)+buttonsCount-1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) + buttonsCount - 1
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	buttons := &blurredButtons
	switch m.focusIndex {
	case len(m.inputs) - 1 + buttonsCount - 1:
		buttons = &focusedSignInButton
	case len(m.inputs) + buttonsCount - 1:
		buttons = &focusedSignUpButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *buttons)

	//b.WriteString(helpStyle.Render("cursor mode is "))
	//b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	//b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return b.String()
}

func (m Model) signUp(c Creds) tea.Cmd {
	return SignUp(c, func(_ Creds) tea.Msg {
		return SignUpMsg(c)
	})
}

func (m Model) signIn(c Creds) tea.Cmd {
	return SignIn(c, func(_ Creds) tea.Msg {
		return SignInMsg(c)
	})
}
