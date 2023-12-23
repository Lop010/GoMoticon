package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gocolly/colly"
	"golang.design/x/clipboard"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		//var i string = "a"
		if os.Getenv("WAYLAND_DISPLAY") != "" {

			cmd := exec.Command("wl-copy", strings.SplitAfter(m.choice, ": ")[1])

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		} else {

			clipboard.Write(clipboard.FmtText, []byte(strings.SplitAfter(m.choice, ": ")[1]))
		}
		return quitTextStyle.Render(fmt.Sprintf("%s, is copied to clipboard :3", strings.SplitAfter(m.choice, ": ")[1]))
	}

	return "\n" + m.list.View()
}

func get_emoticons(query string) []list.Item {
	var list []list.Item
	var values []string
	var names []string

	var URL string = "https://www.fastemote.com/search?q="
	URL = strings.Join([]string{URL, query}, "")
	URL = strings.Replace(URL, " ", "+", -1)

	c := colly.NewCollector()
	c.OnHTML(".grid", func(e *colly.HTMLElement) {
		e.ForEach("a", func(_ int, a *colly.HTMLElement) {
			a.ForEach("li", func(_ int, li *colly.HTMLElement) {
				li.ForEach("div", func(i int, div *colly.HTMLElement) {
					if i == 0 {
						values = append(values, div.Text)
					} else {
						names = append(names, div.Text)
					}
				})
			})
		})
	})
	err := c.Visit(URL)
	if err != nil {
		log.Fatal(err)
		println("Maybe check your internet connection, or the website status (ᵔᴥᵔ)")
	}

	for i := 0; i < len(values); i++ {
		if names[i] != "" && values[i] != "" {
			list = append(list, item(strings.Join([]string{names[i], values[i]}, ": ")))
		}
	}

	return list
}

func main() {

	var argswithoutProg = os.Args[1:]
	if len(argswithoutProg) == 0 {
		fmt.Println("Cant search without a query :3")
		os.Exit(1)
	}
	var items = get_emoticons(argswithoutProg[0])

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Here are the emoticon i found :3"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle

	m := model{list: l}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
