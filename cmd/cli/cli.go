package main

import (
	"context"
	"fmt"
	"gophkeeper/cmd/cli/client"
	"gophkeeper/cmd/cli/ui/authui"
	"gophkeeper/cmd/cli/ui/creditcardui"
	"gophkeeper/cmd/cli/ui/textui"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	text = iota
	card
)

type Mode int

const (
	ModeAuth Mode = iota
	ModeBrowse
	ModeEdit
	ModeAdd
)

var (
	modelStyle = lipgloss.NewStyle().
	//	Width(40).
		Height(10).
		Align(lipgloss.Center, lipgloss.Center).
		BorderStyle(lipgloss.HiddenBorder())
	focusedModelStyle = lipgloss.NewStyle().
	//	Width(40).
		Height(10).
		Align(lipgloss.Center, lipgloss.Center).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("69"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	//tableStyle = lipgloss.NewStyle().
	//	BorderStyle(lipgloss.NormalBorder()).
	//	BorderForeground(lipgloss.Color("240"))
	//isAuthorised 	= false
)

type mainModel struct {
	//widgets
	auth 			authui.Model
	tables			[]table.Model
	editWgts		[]tea.Model

	//variables
	currentTable   	int
	status 			string
	mode			Mode

	//client
	client 			*client.GKClient
	errC			<- chan error

	//context
	ctx 			context.Context
	cancel			context.CancelFunc
}

func newModel() mainModel {
	//init variables
	m := mainModel{
		currentTable: 0,
		mode: ModeAuth,
	}

	//init widgets
	m.auth 	= authui.New(&m)

	m.tables = make([]table.Model,2,2)
	m.tables[0] = createTable("AAAAAAA")
	m.tables[1] = createTable("VVVVVVV")

	m.editWgts = make([]tea.Model,2,2)
	m.editWgts[0] = textui.New()
	m.editWgts[1] = creditcardui.New()

	//client
	m.client = client.NewGKClient("http://localhost:8081")

	//context
	m.ctx, m.cancel = context.WithCancel(context.Background())

	return m
}

func (m mainModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case authui.SignInMsg:  //TODO get rid of code duplication
		_, err := m.client.UserSignIn(m.ctx, client.AuthInput{
			Login: msg.Login,
			Password: msg.Password,
		})

		if err != nil {
			m.status = err.Error()
		} else {
			m.mode = ModeBrowse
			m.errC = m.client.KeepTokensFresh(m.ctx)
		}
	case authui.SignUpMsg: //TODO get rid of code duplication
		_, err := m.client.UserSignUp(context.Background(), client.AuthInput{
			Login: msg.Login,
			Password: msg.Password,
		})

		if err != nil {
			m.status = err.Error()
		} else {
			m.mode = ModeBrowse
			m.errC = m.client.KeepTokensFresh(m.ctx)
		}
	case creditcardui.ChangedMsg:
		m.mode = ModeBrowse
		//TODO handel creditcard changed
		//msg.Number
		//...
	case textui.ChangedMsg:
		m.mode = ModeBrowse
		//TODO handel text changed
		//msg.Text
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.mode != ModeEdit {
				m.cancel()
				return m, tea.Quit
			}
		case "esc":
			if m.mode == ModeEdit || m.mode == ModeAdd {
				m.mode = ModeBrowse
			}
		case "tab":
			if m.mode == ModeBrowse {
				m.currentTable = (m.currentTable + 1) % len(m.tables)
			}
		case "a":
			if m.mode == ModeBrowse {
				m.mode = ModeAdd
			}
		case "e":
			//take current data from selected table
			cells := m.tables[m.currentTable].SelectedRow()

			switch editWgt := m.editWgts[m.currentTable].(type) {
			case textui.Model:
				//0 - data id
				//1 - text id
				//2 - metadata
				editWgt.SetData(textui.ChangedMsg{
					Text: cells[1],
				})
			case creditcardui.Model:
				//0 - data id
				//1 - card number
				//2 - exp date
				//3 - cvv
				//4 - name
				//5 - surname
				//6 - metadata
				date, _ := time.Parse("01/06", cells[2])
				editWgt.SetData(creditcardui.ChangedMsg{
					Number: cells[0],
					ExpDate: date,
					CVV: cells[3],
					Name: cells[4],
					Surname: cells[5],
				})
			}

			if m.mode == ModeBrowse {
				m.mode = ModeEdit
			}
		}

		if m.mode != ModeAuth {
			m.tables[m.currentTable], cmd = m.tables[m.currentTable].Update(msg)
			cmds = append(cmds, cmd)
		}

		if m.mode == ModeEdit {
			m.editWgts[m.currentTable], cmd = m.editWgts[m.currentTable].Update(msg)
			cmds = append(cmds, cmd)
		}
	default:

	}

	select {
	case <- m.errC:
		m.mode = ModeAuth
	default:

	}

	if m.mode == ModeAuth {
		m.auth, cmd = m.auth.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		for _, tbl := range m.tables {
			tbl, cmd = tbl.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	var s string

	//var status string
	if m.mode != ModeAuth {
		var line []string
		for i, tbl := range m.tables {
			if i == m.currentTable {
				if m.mode == ModeEdit || m.mode == ModeAdd {
					switch m.currentTable {
					case text:
						line = append(line, m.editWgts[m.currentTable].View())
					case card:
						line = append(line, m.editWgts[m.currentTable].View())
					}
				} else {
					line = append(line, focusedModelStyle.Render(fmt.Sprintf("%4s", tbl.View())))
				}
			} else {
				line = append(line, modelStyle.Render(tbl.View()))
			}
		}
		s += lipgloss.JoinHorizontal(lipgloss.Top, line...)
		if m.mode != ModeEdit  {
			m.status = "\ntab: focus next • n: new  • q: exit\n"
		}
	} else {
		s += lipgloss.NewStyle().Render(m.auth.View())
	}

	s += helpStyle.Render(fmt.Sprintf(m.status))

	return s
}

func createTable(name string) table.Model {
	columns := []table.Column{
		{Title: name, Width: 10},
		{Title: "City", Width: 10},
		{Title: "Country", Width: 10},
		{Title: "Population", Width: 10},
	}

	rows := []table.Row{
		{"1", "Tokyo", "Japan", "37,274,000"},
		{"2", "Delhi", "India", "32,065,760"},
		{"3", "Shanghai", "China", "28,516,904"},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
		)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return t
}

func main() {
	//tea.NewProgram(creditcardui.New()).Start()
	tea.NewProgram(newModel()).Start()
}