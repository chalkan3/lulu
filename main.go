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
	"gopkg.in/yaml.v3"
)

var choices = []string{}

type Config struct {
	Name           string `yaml:"name,omitempty"`
	ConfigFileName string `yaml:"config-file-name,omitempty"`
}

type Configs struct {
	Configs []*Config `yaml:"kube-config,omitempty"`
}
type model struct {
	cursor  int
	choice  string
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
			m.choice = choices[m.cursor]
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
		}

	}

	return m, nil
}

func (m model) View() string {
	s := strings.Builder{}
	s.WriteString("Bem vindo ao lulu cluster handler!  😎 escolha uma config\n\n")

	for i := 0; i < len(choices); i++ {
		if m.cursor == i {
			s.WriteString("(🤏) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(choices[i])
		s.WriteString("\n")
	}
	s.WriteString("\n(press q to quit)\n")

	return s.String()
}

func main() {
	yfile, err := ioutil.ReadFile("config/config.yml")

	if err != nil {

		log.Fatal(err)
	}

	configs := &Configs{}
	err2 := yaml.Unmarshal(yfile, &configs)

	if err2 != nil {

		log.Fatal(err2)
	}

	for _, v := range configs.Configs {
		choices = append(choices, v.Name)
	}

	mm := model{}
	mm.actions = configs.Configs
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
				stdout, err := exec.Command("rm", homePath+"/.kube/config").Output()
				fmt.Println(string(stdout))

				if err != nil {
					fmt.Println(err.Error())
					return
				}

				cmd := exec.Command("mv", homePath+"/.kube/"+v.ConfigFileName, homePath+"/.kube/config")
				stdout, err = cmd.Output()

				if err != nil {
					fmt.Println(err.Error())
					return
				}

				fmt.Println(string(stdout))
			}
		}
	}
}
