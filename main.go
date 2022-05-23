package main

// A simple example that shows how to retrieve a value from a Bubble Tea
// program after the Bubble Tea has exited.

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

var choices = []*Config{}
var tabss = []*Tabs{}

type Config struct {
	Name           string `yaml:"name,omitempty"`
	ConfigFileName string `yaml:"config-file-name,omitempty"`
	TabID          int    `yaml:"tab-id,omitempty"`
}

type Configs struct {
	Configs []*Config `yaml:"kube-config,omitempty"`
	Tabs    []*Tabs   `yaml:"tabs,omitempty"`
}

type Tabs struct {
	ID   int    `yaml:"id"`
	Name string `yaml:"name"`
}
type model struct {
	cursor  int
	choice  string
	active  int
	actions []*Config
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "enter":
			// Send the choice on the channel and exit.
			m.choice = choices[m.cursor].Name
			return m, tea.Quit

		case "down", "j":
			m.cursor++
			if m.cursor >= len(choices) {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(choices) - 1
			}
		case "tab":
			if m.active > len(tabss)-1 {
				m.active = 1
			} else {
				m.active++

			}

		}

	}

	return m, nil
}

func (m model) View() string {
	s := strings.Builder{}

	{
		tt := []string{}
		for _, v := range tabss {
			if v.ID == m.active {
				tt = append(tt, activeTab.Render(v.Name))
			} else {
				tt = append(tt, tab.Render(v.Name))
			}
		}

		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			tt...,
		)

		gap := tabGap.Render(strings.Repeat(" ", max(0, width-lipgloss.Width(row)-1)))
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)

		s.WriteString(row + "\n\n")

	}

	targetCluster := "None"
	for i := 0; i < len(choices); i++ {
		if choices[i].TabID == m.active {
			if m.cursor == i {
				targetCluster = choices[i].Name
				s.WriteString("(🐙)")
			} else {
				s.WriteString("( ) ")
			}
			s.WriteString(choices[i].Name)
			s.WriteString("\n")
		}

	}

	{
		w := lipgloss.Width

		statusKey := statusStyle.Render("STATUS")
		encoding := encodingStyle.Render("Target Cluster")
		fishCake := fishCakeStyle.Render(targetCluster)
		statusVal := statusText.Copy().
			Width(width - w(statusKey) - w(encoding) - w(fishCake)).
			Render("choosing")

		bar := lipgloss.JoinHorizontal(lipgloss.Top,
			statusKey,
			statusVal,
			encoding,
			fishCake,
		)

		s.WriteString("\n\n" + statusBarStyle.Width(width).Render(bar))
	}

	return s.String()
}

func main() {

	doc := strings.Builder{}
	// var style = lipgloss.NewStyle().Foreground(lipgloss.Color("219"))

	// fmt.Println(style.Render("Bem vindo ao lulu cluster handler!  😎 escolha uma config\n\n"))
	desc := lipgloss.JoinVertical(lipgloss.Left,
		descStyle.Render("Lulu Multi Kube Config"),
		infoStyle.Render("Luciano"+divider+"S"+divider+"Gomes"),
	)

	row := lipgloss.JoinHorizontal(lipgloss.Top, desc)
	doc.WriteString(row + "\n\n")
	fmt.Println(doc.String())

	// usr, _ := user.Current()
	// dir := usr.HomeDir

	// yfile, err := ioutil.ReadFile(dir + "/.kube/config.yml")
	yfile, err := ioutil.ReadFile("config/config.yml")

	if err != nil {

		log.Fatal(err)
	}

	configs := &Configs{}
	err2 := yaml.Unmarshal(yfile, &configs)

	if err2 != nil {

		log.Fatal(err2)
	}

	tabss = configs.Tabs
	choices = configs.Configs

	mm := model{}
	mm.actions = configs.Configs
	mm.active = 1
	p := tea.NewProgram(mm)

	// StartReturningModel returns the model as a tea.Model.
	m, err := p.StartReturningModel()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}

	// Assert the final tea.Model to our local model and print the choice.
	if m, ok := m.(model); ok && m.choice != "" {
		for _, v := range m.actions {
			if v.Name == m.choice {
				homePath := os.ExpandEnv("$HOME")
				stdout, err := exec.Command("rm", "-rf", homePath+"/.kube/config").Output()
				fmt.Println(string(stdout))

				if err != nil {
					fmt.Println(err.Error())
					return
				}

				cmd := exec.Command("cp", homePath+"/.kube/"+v.ConfigFileName, homePath+"/.kube/config")
				stdout, err = cmd.Output()

				if err != nil {
					fmt.Println(err.Error())
					return
				}

				cm2 := exec.Command("kubectl", "config", "get-clusters")
				stdout, err = cm2.Output()

				if err != nil {
					fmt.Println(err)
					return
				}

				// fmt.Println("Cluster Ativo", v.Name)
				// fmt.Println("Executing kubectl config get-clusters\n\n")
				// fmt.Println(string(stdout))

				s := strings.Builder{}
				{
					w := lipgloss.Width

					statusKey := statusStyle.Render("STATUS")
					encoding := encodingStyle.Render("Cluster")
					fishCake := fishCakeStyle.Render(strings.TrimSpace(strings.Trim(string(stdout), "NAME")))
					statusVal := statusText.Copy().
						Width(width - w(statusKey) - w(encoding) - w(fishCake)).
						Render("kubectl config get-clusters")

					bar := lipgloss.JoinHorizontal(lipgloss.Top,
						statusKey,
						statusVal,
						encoding,
						fishCake,
					)

					s.WriteString("\n\n" + statusBarStyle.Width(width).Render(bar))
				}

				fmt.Println(s.String())

			}
		}
	}
}

const (
	// In real life situations we'd adjust the document to fit the width we've
	// detected. In the case of this example we're hardcoding the width, and
	// later using the detected width only to truncate in order to avoid jaggy
	// wrapping.
	width = 96

	columnWidth = 30
)

// Style definitions.
var (

	// General.

	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	divider = lipgloss.NewStyle().
		SetString("•").
		Padding(0, 1).
		Foreground(subtle).
		String()

	url = lipgloss.NewStyle().Foreground(special).Render

	// Tabs.

	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tab = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(highlight).
		Padding(0, 1)

	activeTab = tab.Copy().Border(activeTabBorder, true)

	tabGap = tab.Copy().
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	// Title.

	titleStyle = lipgloss.NewStyle().
			MarginLeft(1).
			MarginRight(5).
			Padding(0, 1).
			Italic(true).
			Foreground(lipgloss.Color("#FFF7DB")).
			SetString("Lip Gloss")

	descStyle = lipgloss.NewStyle().MarginTop(1)

	infoStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(subtle)

	// Dialog.

	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			Padding(0, 3).
			MarginTop(1)

	activeButtonStyle = buttonStyle.Copy().
				Foreground(lipgloss.Color("#FFF7DB")).
				Background(lipgloss.Color("#F25D94")).
				MarginRight(2).
				Underline(true)

	// List.

	list = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(subtle).
		MarginRight(2).
		Height(8).
		Width(columnWidth + 1)

	listHeader = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(subtle).
			MarginRight(2).
			Render

	listItem = lipgloss.NewStyle().PaddingLeft(2).Render

	checkMark = lipgloss.NewStyle().SetString("✓").
			Foreground(special).
			PaddingRight(1).
			String()

	listDone = func(s string) string {
		return checkMark + lipgloss.NewStyle().
			Strikethrough(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#969B86", Dark: "#696969"}).
			Render(s)
	}

	// Paragraphs/History.

	historyStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(highlight).
			Margin(1, 3, 0, 0).
			Padding(1, 2).
			Height(19).
			Width(columnWidth)

	// Status Bar.

	statusNugget = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	statusStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 1).
			MarginRight(1)

	encodingStyle = statusNugget.Copy().
			Background(lipgloss.Color("#A550DF")).
			Align(lipgloss.Right)

	statusText = lipgloss.NewStyle().Inherit(statusBarStyle)

	fishCakeStyle = statusNugget.Copy().Background(lipgloss.Color("#6124DF"))

	// Page.

	docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
