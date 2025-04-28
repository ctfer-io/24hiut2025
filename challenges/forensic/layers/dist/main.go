package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	flag   = "at least use multistep dockerfiles..."
	header = "POPA COLA AUTHENTICATOR"
)

var Key = ""

var (
	matrixGreen = lipgloss.Color("#00FF41")
	black       = lipgloss.Color("#000000")

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(matrixGreen).
			Foreground(matrixGreen).
			Background(black).
			Padding(1, 2).
			Width(60).
			Align(lipgloss.Center)

	titleStyle = lipgloss.NewStyle().
			Foreground(matrixGreen).
			Background(black).
			Bold(true).
			MarginBottom(1).
			Align(lipgloss.Center)

	centerStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Foreground(matrixGreen).
			Background(black)
)

type tickMsg struct{}

type model struct {
	input         textinput.Model
	resultMessage string
	state         string // "input", "result"
	width         int
	height        int
	autoDismiss   bool // only true for ACCESS DENIED
}

func newTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Enter password"
	ti.Focus()
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Prompt = "> "
	return ti
}

func initialModel() model {
	return model{
		input: newTextInput(),
		state: "input",
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		if m.state == "input" && msg.String() == "enter" {
			password := strings.TrimSpace(m.input.Value())
			if password == Key {
				m.resultMessage = fmt.Sprintf("24HIUT{%s}", flag)
				m.state = "result"
				m.autoDismiss = false
				return m, nil
			} else {
				m.resultMessage = "ACCESS DENIED"
				m.state = "result"
				m.autoDismiss = true
				return m, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
					return tickMsg{}
				})
			}
		} else if m.state == "result" && !m.autoDismiss {
			m.input = newTextInput()
			m.state = "input"
			return m, textinput.Blink
		}

	case tickMsg:
		if m.autoDismiss {
			m.input = newTextInput()
			m.state = "input"
			m.autoDismiss = false
			return m, textinput.Blink
		}
	}

	if m.state == "input" {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	var content string

	switch m.state {
	case "input":
		content = lipgloss.JoinVertical(lipgloss.Center,
			titleStyle.Width(m.width).Render(header),
			m.input.View(),
		)
	case "result":
		msg := m.resultMessage
		hint := ""
		if !m.autoDismiss {
			hint = "\nPress any key to try again..."
		}
		content = lipgloss.JoinVertical(lipgloss.Center,
			titleStyle.Width(m.width).Render(header),
			boxStyle.Render(msg),
			centerStyle.Render(hint),
		)
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func main() {
	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithContext(ctx))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
