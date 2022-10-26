package ui

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/timer"
	"gophkeeper/cmd/cli/client"
	"gophkeeper/cmd/cli/ui/authui"
	"gophkeeper/cmd/cli/ui/creditcardui"
	"gophkeeper/cmd/cli/ui/credsui"
	"gophkeeper/cmd/cli/ui/textui"
	"gophkeeper/internal/domain"
	"strconv"
	"strings"

	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	text = iota
	card
	cred
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
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	//tableStyle = lipgloss.NewStyle().
	//	BorderStyle(lipgloss.NormalBorder()).
	//	BorderForeground(lipgloss.Color("240"))
	//isAuthorised 	= false
)

type MainModel struct {
	//titles
	titles []string

	//widgets
	auth     authui.Model
	tables   []table.Model
	editWgts []tea.Model

	//variables
	currentTable  int
	status        string
	mode          Mode
	syncDataTimer timer.Model // used without view

	//client
	client *client.GKClient
	errC   <-chan error

	//context
	ctx    context.Context
	cancel context.CancelFunc
}

func NewMainModel(addr string) MainModel {
	//init variables
	m := MainModel{
		currentTable:  0,
		mode:          ModeAuth,
		syncDataTimer: timer.NewWithInterval(3*time.Second, 3*time.Second),
	}

	//init titles
	m.titles = make([]string, 3, 3)
	m.titles[text] = "Текстовые и бинарные данные"
	m.titles[card] = "Данные банковских карт"
	m.titles[cred] = "Креды"

	//init widgets
	m.auth = authui.New()

	m.tables = make([]table.Model, 3, 3)
	m.tables[text] = createTextDataTable(nil)
	m.tables[card] = createCardDataTable(nil)
	m.tables[cred] = createCredDataTable(nil)

	m.editWgts = make([]tea.Model, 3, 3)
	m.editWgts[text] = textui.New()
	m.editWgts[card] = creditcardui.New()
	m.editWgts[cred] = credsui.New()

	//client
	m.client = client.NewGKClient(addr)

	//context
	m.ctx, m.cancel = context.WithCancel(context.Background())

	return m
}

func (m MainModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case authui.SignInMsg: //TODO get rid of code duplication
		_, err := m.client.UserSignIn(m.ctx, client.AuthInput{
			Login:    msg.Login,
			Password: msg.Password,
		})

		if err != nil {
			m.status = err.Error()
		} else {
			m.mode = ModeBrowse
			m.errC = m.client.KeepTokensFresh(m.ctx)
			cmds = append(cmds, m.syncDataTimer.Init())
		}
	case authui.SignUpMsg: //TODO get rid of code duplication
		_, err := m.client.UserSignUp(context.Background(), client.AuthInput{
			Login:    msg.Login,
			Password: msg.Password,
		})

		if err != nil {
			m.status = err.Error()
		} else {
			m.mode = ModeBrowse
			m.errC = m.client.KeepTokensFresh(m.ctx)
			cmds = append(cmds, m.syncDataTimer.Init())
		}
	case credsui.ChangedMsg:
		switch m.mode {
		case ModeEdit:
			id, _ := strconv.Atoi(m.tables[cred].SelectedRow()[0])
			err := m.client.UpdateCredData(m.ctx, domain.CredData{
				ID:       id,
				Login:    msg.Login,
				Password: msg.Password,
				Metadata: msg.Metadata,
			})
			if err != nil {
				m.status = err.Error()
			}
		case ModeAdd:
			err := m.client.CreateNewCredData(m.ctx, domain.CredData{
				Login:    msg.Login,
				Password: msg.Password,
				Metadata: msg.Metadata,
			})
			if err != nil {
				m.status = err.Error()
			}
		}
		m.mode = ModeBrowse
	case creditcardui.ChangedMsg:
		switch m.mode {
		case ModeEdit:
			id, _ := strconv.Atoi(m.tables[card].SelectedRow()[0])
			err := m.client.UpdateCardData(m.ctx, domain.CardData{
				ID:         id,
				CardNumber: msg.Number,
				ExpDate:    msg.ExpDate,
				CVV:        msg.CVV,
				Name:       msg.Name,
				Surname:    msg.Surname,
				Metadata:   msg.Metadata,
			})
			if err != nil {
				m.status = err.Error()
			}
		case ModeAdd:
			err := m.client.CreateNewCardData(m.ctx, domain.CardData{
				CardNumber: msg.Number,
				ExpDate:    msg.ExpDate,
				CVV:        msg.CVV,
				Name:       msg.Name,
				Surname:    msg.Surname,
				Metadata:   msg.Metadata,
			})
			if err != nil {
				m.status = err.Error()
			}
		}
		m.mode = ModeBrowse
	case textui.ChangedMsg:
		switch m.mode {
		case ModeEdit:
			id, _ := strconv.Atoi(m.tables[text].SelectedRow()[0])
			err := m.client.UpdateTextData(m.ctx, domain.TextData{
				ID:       id,
				Text:     msg.Text,
				Metadata: msg.Metadata,
			})
			if err != nil {
				m.status = err.Error()
			}
		case ModeAdd:
			err := m.client.CreateNewTextData(m.ctx, domain.TextData{
				Text:     msg.Text,
				Metadata: msg.Metadata,
			})
			if err != nil {
				m.status = err.Error()
			}
		}
		m.mode = ModeBrowse
	case timer.TickMsg, timer.StartStopMsg:
		m.syncDataTimer, cmd = m.syncDataTimer.Update(msg)
		cmds = append(cmds, m.syncDataTimer.Init())

		textRows, err := m.client.GetAllTextData(m.ctx)
		if err != nil {
			m.status = err.Error()
		} else {
			cursor := m.tables[text].Cursor()
			m.tables[text] = createTextDataTable(textRows)
			if cursor != -1 {
				m.tables[text].SetCursor(cursor)
			}
		}

		cardRows, err := m.client.GetAllCardData(m.ctx)
		if err != nil {
			m.status = err.Error()
		} else {
			cursor := m.tables[card].Cursor()
			m.tables[card] = createCardDataTable(cardRows)
			if cursor != -1 {
				m.tables[card].SetCursor(cursor)
			}
		}

		credsRows, err := m.client.GetAllCredsData(m.ctx)
		if err != nil {
			m.status = err.Error()
		} else {
			cursor := m.tables[cred].Cursor()
			m.tables[cred] = createCredDataTable(credsRows)
			if cursor != -1 {
				m.tables[cred].SetCursor(cursor)
			}
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.mode != ModeEdit && m.mode != ModeAdd {
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
			if m.mode != ModeBrowse {
				break
			}
			m.mode = ModeAdd

			m.editWgts[m.currentTable], cmd = m.editWgts[m.currentTable].Update(textinput.Blink)
			return m, cmd
		case "e":
			if m.mode != ModeBrowse {
				break
			}

			//take current data from selected table
			var cells []string
			if m.tables[m.currentTable].Cursor() >= 0 {
				cells = m.tables[m.currentTable].SelectedRow()
			}

			if len(cells) == 0 {
				break
			}

			switch editWgt := m.editWgts[m.currentTable].(type) {
			case textui.Model:
				//0 - data id
				//1 - text id
				//2 - metadata
				editWgt.SetData(textui.ChangedMsg{
					Text:     cells[1],
					Metadata: cells[2],
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
					Number:   cells[1],
					ExpDate:  date,
					CVV:      cells[3],
					Name:     cells[4],
					Surname:  cells[5],
					Metadata: cells[6],
				})
			case credsui.Model:
				//0 - data id
				//1 - login
				//2 - password
				//3 - metadata
				editWgt.SetData(credsui.ChangedMsg{
					Login:    cells[1],
					Password: cells[2],
					Metadata: cells[3],
				})
			}
			m.mode = ModeEdit
		}

		if m.mode != ModeAuth {
			m.tables[m.currentTable], cmd = m.tables[m.currentTable].Update(msg)
			cmds = append(cmds, cmd)
		}

		if m.mode == ModeEdit || m.mode == ModeAdd {
			m.editWgts[m.currentTable], cmd = m.editWgts[m.currentTable].Update(msg)
			cmds = append(cmds, cmd)
		}
	default:

	}

	select {
	case <-m.errC:
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

func (m MainModel) View() string {
	var s string

	//var status string
	if m.mode != ModeAuth {
		s += lipgloss.JoinHorizontal(lipgloss.Center,
			lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.titles[text])+strings.Repeat(" ", abs(m.tables[text].Width()-len(m.titles[text]))),
			lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.titles[card])+strings.Repeat(" ", abs(m.tables[card].Width()-len(m.titles[card]))),
			"                       ",
			lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.titles[cred])+strings.Repeat(" ", abs(m.tables[cred].Width()-len(m.titles[cred]))),
		)

		s += "\n"

		var line []string
		for i, tbl := range m.tables {
			if i == m.currentTable {
				if m.mode == ModeEdit || m.mode == ModeAdd {
					line = append(line, m.editWgts[m.currentTable].View())
				} else {
					line = append(line, focusedModelStyle.Render(fmt.Sprintf("%4s", tbl.View())))
				}
			} else {
				line = append(line, modelStyle.Render(tbl.View()))
			}
		}
		s += lipgloss.JoinHorizontal(lipgloss.Top, line...)
	} else {
		s += lipgloss.NewStyle().Render(m.auth.View())
	}

	if m.mode == ModeBrowse {
		help := "\ntab: focus next • q: exit • a: add new row • e: edit selected row\n"
		s += helpStyle.Render(fmt.Sprintf(help))
	}
	s += helpStyle.Render(fmt.Sprintf(m.status))

	return s
}

func createTable(columns []table.Column, rows []table.Row) table.Model {
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

	if len(columns) == 0 || len(rows) == 0 {
		t.SetCursor(-1)
	}

	return t
}

func createTextDataTable(data []domain.TextData) table.Model {
	columns := []table.Column{
		{Title: "id", Width: 4},
		{Title: "Data", Width: 10},
		{Title: "Metadata", Width: 20},
	}

	rows := make([]table.Row, 0, len(data))
	for _, row := range data {
		rows = append(rows, table.Row{
			strconv.Itoa(row.ID), row.Text, row.Metadata,
		})
	}

	return createTable(columns, rows)
}

func createCardDataTable(data []domain.CardData) table.Model {
	columns := []table.Column{
		{Title: "id", Width: 4},
		{Title: "CardNumber", Width: 20},
		{Title: "ExpData", Width: 8},
		{Title: "CVV", Width: 4},
		{Title: "Name", Width: 10},
		{Title: "Surname", Width: 10},
		{Title: "Metadata", Width: 20},
	}

	rows := make([]table.Row, 0, len(data))
	for _, row := range data {
		rows = append(rows, table.Row{
			strconv.Itoa(row.ID), row.CardNumber, row.ExpDate.Format("01/06"), row.CVV, row.Name, row.Surname, row.Metadata,
		})
	}

	return createTable(columns, rows)
}

func createCredDataTable(data []domain.CredData) table.Model {
	columns := []table.Column{
		{Title: "id", Width: 4},
		{Title: "Login", Width: 15},
		{Title: "Password", Width: 10},
		{Title: "Metadata", Width: 20},
	}

	rows := make([]table.Row, 0, len(data))
	for _, row := range data {
		rows = append(rows, table.Row{
			strconv.Itoa(row.ID), row.Login, row.Password, row.Metadata,
		})
	}

	return createTable(columns, rows)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

//func main() {
//	tea.NewProgram(newModel()).Start()
//}
