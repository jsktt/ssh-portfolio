package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	wishtea "github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

// --- Keybindings ---
type keyMap struct {
	Up, Down, Enter, Back, English, Korean, Quit key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Back, k.English, k.Korean, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Back},
		{k.English, k.Korean, k.Quit},
	}
}

var keys = keyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Enter:   key.NewBinding(key.WithKeys("enter", "l"), key.WithHelp("ent/l", "select")),
	Back:    key.NewBinding(key.WithKeys("h", "backspace"), key.WithHelp("h", "back")),
	English: key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "english")),
	Korean:  key.NewBinding(key.WithKeys("k"), key.WithHelp("k", "korean")),
	Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

// --- Styling ---
var (
	purple    = lipgloss.Color("99")
	gray      = lipgloss.Color("245")
	lightGray = lipgloss.Color("241")
	green     = lipgloss.Color("#25A065")
	highlight = lipgloss.Color("#00EAD3")

	appStyle   = lipgloss.NewStyle().Padding(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(green).
			Padding(0, 1).
			Bold(true)

	leftPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(green).
			Padding(1)

	rightPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	asciiStyle = lipgloss.NewStyle().
			Foreground(gray)
)

const asciiArt = `
		__,  ,__        
	   (   )(   )
	    \ (,,) /
	    / .  . \
	   (  = ^ = )
	    )      (
	   (        )
	  ( \ \  / / )
	'----\_!!_/----'
 
`

// --- Item Definition ---
type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// --- Model ---
type model struct {
	list     list.Model
	keys     keyMap
	help     help.Model
	choice   string
	quitting bool
	width    int
	height   int
	language string
}

func initialModel() model {
	items := []list.Item{
		item{title: "🐦 About Me", desc: "Software Engineer"},
		item{title: "🐙 Projects", desc: "My Work"},
		item{title: "📧 Contact", desc: "Get in touch"},
		item{title: "📧 Blog", desc: "Read more"},
	}

	m := list.New(items, list.NewDefaultDelegate(), 0, 0)
	m.Title = "Junsung Kim"
	m.Styles.Title = titleStyle
	m.SetShowHelp(false)

	return model{
		list:     m,
		keys:     keys,
		help:     help.New(),
		language: "en",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		h, v := appStyle.GetFrameSize()
		leftWidth := (msg.Width * 2 / 3) - h - 4
		m.list.SetSize(leftWidth, msg.Height-v-6)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.English):
			m.language = "en"

		case key.Matches(msg, m.keys.Korean):
			m.language = "ko"

		case key.Matches(msg, m.keys.Enter):
			if m.choice == "" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.choice = i.title
				}
			}

		case key.Matches(msg, m.keys.Back):
			m.choice = ""
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "\n  See you later!\n\n"
	}

	totalWidth := m.width - 4
	totalHeight := m.height - 6

	leftWidth := (totalWidth * 2) / 3
	rightWidth := totalWidth - leftWidth

	// --- LEFT PANEL (2/3) ---
	leftBox := leftPanelStyle.Width(leftWidth - 2).Height(totalHeight).Render(m.list.View())

	// --- RIGHT PANEL (1/3) ---
	// 1. Language Table
	t := table.New().
		Border(lipgloss.ThickBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
		StyleFunc(func(row, col int) lipgloss.Style {
			base := lipgloss.NewStyle().Padding(0, 1)
			if row == table.HeaderRow {
				return base.Foreground(purple).Bold(true).Align(lipgloss.Center)
			}
			isEnSelected := m.language == "en" && row == 0
			isKoSelected := m.language == "ko" && row == 1
			if isEnSelected || isKoSelected {
				return base.Foreground(highlight).Background(lipgloss.Color("235")).Bold(true)
			}
			return base.Foreground(gray)
		}).
		Headers("SELECT LANGUAGE").
		Rows([][]string{{"English [e]"}, {"Korean  [k]"}}...).
		Width(rightWidth - 5)

	// 2. ASCII Art Area
	// We center the ASCII art within the remaining height of the right panel
	artView := asciiStyle.Width(rightWidth - 3).Render(asciiArt)

	// 3. Assemble Right Side
	rightContent := lipgloss.JoinVertical(lipgloss.Center, "\n"+t.String(), "\n\n", artView)
	rightBox := rightPanelStyle.Width(rightWidth - 2).Height(totalHeight).Render(rightContent)

	// --- DASHBOARD LAYOUT ---
	dashboard := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)

	// If a section is selected
	// If a section is selected
	if m.choice != "" {
		header := titleStyle.Render(m.choice)
		var content string

		// We check the choice first, then branch by language
		switch m.choice {
		case "🐦 About Me":
			if m.language == "ko" {
				content = "안녕하세요! 소프트웨어 엔지니어 김준성입니다.\n디자인의 미학과 소프트웨어의 논리를 결합하는 것을 좋아합니다."
			} else {
				content = "Hi! I'm Junsung Kim, a Software Engineer.\nI enjoy combining the aesthetics of design with the logic of software."
			}

		case "🐙 Projects":
			if m.language == "ko" {
				content = "• SSH 포트폴리오: 현재 보고 계신 이 앱입니다.\n• 포트폴리오: React + Typescript 로 만들었어요. junssung-kim.com\n• 핀트: Java17, Spring Boot 기반 이미지 공유 프랫폼\n•프롬프트 분류화: RoBERTa 모델 학습 통해 유저 프롬프트 정확하게 분류"
			} else {
				content = "• SSH Portfolio: The app you are currently viewing.\n• Portfolio: Made with React + Typescsript. junsung-kim.com.\n• Pint: Image sharing platform using Java17 and Spring Boot\n• Performant Prompt Classification: fine-tuned RoBERTa model to classify user prompts."
			}

		case "📧 Contact":
			if m.language == "ko" {
				content = "궁금한 점이 있으시면 언제든 연락주세요!\n이메일: uitomde@gmail.com\n깃허브: github.com/jsktt"
			} else {
				content = "Feel free to reach out for any inquiries!\nEmail: uitomde@gmail.com\nGithub: github.com/jsktt"
			}

		case "📧 Blog":
			if m.language == "ko" {
				content = "디테일은 재 사이트를 방문해주세요! junsung-kim.com"
			} else {
				content = "Check my site for more details! junsung-kim.com"
			}

		default:
			content = "Section under construction..."
		}

		// Combine components into the final view
		return appStyle.Render(
			header + "\n\n" +
				content + "\n\n\n" +
				m.help.View(m.keys),
		)
	}

	return appStyle.Render(dashboard + "\n" + m.help.View(m.keys))
}

func main() {
	server, _ := wish.NewServer(
		wish.WithAddress("0.0.0.0:22"),
		wish.WithHostKeyPath(".ssh/host_ed25519"),
		wish.WithMiddleware(wishtea.Middleware(teaHandler), logging.Middleware()),
	)
	fmt.Printf("SSH server running on :2222\n")
	_ = server.ListenAndServe()
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	return initialModel(), []tea.ProgramOption{tea.WithAltScreen()}
}
