package creditcardui

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ChangedMsg struct {
	Number string
	ExpDate time.Time
	CVV	string
	Name  string
	Surname string
}

func Change(c ChangedMsg, fn func(c ChangedMsg) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return fn(c)
	}
}

type (
	errMsg error
)

const (
	ccn = iota
	exp
	cvv
	name
)

const (
	hotPink  = lipgloss.Color("#FF06B7")
	darkGray = lipgloss.Color("#767676")
)

var (
	inputStyle = lipgloss.NewStyle().Foreground(hotPink)
	applyStyle = lipgloss.NewStyle().Foreground(darkGray)
)

type Model struct {
	inputs  []textinput.Model
	focused int
	err     error
}

// Validator functions to ensure valid input
func ccnValidator(s string) error {
	// Credit Card Number should a string less than 20 digits
	// It should include 16 integers and 3 spaces
	if len(s) > 16+3 {
		return fmt.Errorf("CCN is too long")
	}

	// The last digit should be a number unless it is a multiple of 4 in which
	// case it should be a space
	if len(s)%5 == 0 && s[len(s)-1] != ' ' {
		return fmt.Errorf("CCN must separate groups with spaces")
	}
	if len(s)%5 != 0 && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
		return fmt.Errorf("CCN is invalid")
	}

	// The remaining digits should be integers
	c := strings.ReplaceAll(s, " ", "")
	_, err := strconv.ParseInt(c, 10, 64)

	return err
}

func expValidator(s string) error {
	good := regexp.MustCompile(`^(0[1-9]*|1[012]*)(/([0-9]{0,2}))?$`).MatchString(s)
	if !good {
		return errors.New("EXP is invalid")
	}

	// The 3 character should be a slash (/)
	// The rest thould be numbers
	e := strings.ReplaceAll(s, "/", "")
	_, err := strconv.ParseInt(e, 10, 64)
	if err != nil {
		return fmt.Errorf("EXP is invalid")
	}

	// There should be only one slash and it should be in the 2nd index (3rd character)
	if len(s) >= 3 && (strings.Index(s, "/") != 2 || strings.LastIndex(s, "/") != 2) {
		return fmt.Errorf("EXP is invalid")
	}

	return nil
}

func cvvValidator(s string) error {
	// The CVV should be a number of 3 digits
	// Since the input will already ensure that the CVV is a string of length 3,
	// All we need to do is check that it is a number
	_, err := strconv.ParseInt(s, 10, 64)
	return err
}

func nameValidator(s string) error {
	good := regexp.MustCompile(`^[a-zA-Z]+[\s]?[a-zA-Z]*$`).MatchString(s)
	if !good {
		return errors.New("name is invalid")
	}
	return nil
}

func New() Model {
	var inputs []textinput.Model = make([]textinput.Model, 4)
	inputs[ccn] = textinput.New()
	inputs[ccn].Placeholder = "4505 **** **** 1234"
	inputs[ccn].Focus()
	inputs[ccn].CharLimit = 20
	inputs[ccn].Width = 30
	inputs[ccn].Prompt = ""
	inputs[ccn].Validate = ccnValidator

	inputs[exp] = textinput.New()
	inputs[exp].Placeholder = "MM/YY "
	inputs[exp].CharLimit = 5
	inputs[exp].Width = 5
	inputs[exp].Prompt = ""
	inputs[exp].Validate = expValidator

	inputs[cvv] = textinput.New()
	inputs[cvv].Placeholder = "XXX"
	inputs[cvv].CharLimit = 3
	inputs[cvv].Width = 5
	inputs[cvv].Prompt = ""
	inputs[cvv].Validate = cvvValidator

	inputs[name] = textinput.New()
	inputs[name].Placeholder = "Ivan Ivanov"
	inputs[name].CharLimit = 31
	inputs[name].Width = 20
	inputs[name].Prompt = ""
	inputs[name].Validate = nameValidator

	return Model{
		inputs:  inputs,
		focused: 0,
		err:     nil,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				//TODO Again validate to escape unfilled of partly filled fields
				m.change(ChangedMsg{
					Number: strings.ReplaceAll(m.inputs[ccn].Value(), " ",""),
					ExpDate: parseExpireDate(m.inputs[exp].Value()),
					CVV: m.inputs[cvv].Value(),
					Name: parseName(m.inputs[name].Value()),
					Surname: parseSurname(m.inputs[name].Value()),
				})
				//return m, tea.Quit
			}
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return fmt.Sprintf(
		`
 %s
 %s
 %s  %s
 %s  %s
 %s
 %s
 %s
`,
		inputStyle.Width(30).Render("Card Number"),
		m.inputs[ccn].View(),
		inputStyle.Width(6).Render("EXP"),
		inputStyle.Width(6).Render("CVV"),
		m.inputs[exp].View(),
		m.inputs[cvv].View(),
		inputStyle.Width(6).Render("Name"),
		m.inputs[name].View(),
		applyStyle.Render("Apply ->"),
	) + "\n"
}

// nextInput focuses the next input field
func (m *Model) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *Model) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}

func (m Model) change(c ChangedMsg) tea.Cmd {
	return Change(c, func(_ ChangedMsg) tea.Msg {
		return c
	})
}

func (m Model)SetData(data ChangedMsg) {
	m.inputs[ccn].SetValue(ccnFormater(data.Number))
	m.inputs[exp].SetValue(data.ExpDate.Format("01/06"))
	m.inputs[cvv].SetValue(data.CVV)
	m.inputs[name].SetValue(data.Name + " " + data.Surname)
}

func parseExpireDate(exp string) time.Time {
	date, err := time.Parse("01/06", exp)
	if err != nil {
		log.Panicln(err.Error())
	}

	return date
}

func parseName(nameAndSurname string) string {
	split := strings.Split(nameAndSurname, " ")
	if len(split) >= 1 {
		return split[0]
	}
	return ""
}

func parseSurname(nameAndSurname string) string {
	split := strings.Split(nameAndSurname, " ")
	if len(split) >= 2 {
		return split[1]
	}
	return ""
}

func ccnFormater(ccn string) string{
	result := ""
	for i := 0; i < len(ccn); i++ {
		if i % 4 == 0 {
			result = result + " "
		}
		result = result + string(ccn[i])
	}
	return result
}