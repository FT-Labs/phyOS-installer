package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type RenderStr interface {
    View() string
}

type model struct {
    Tabs       []string
    TabContent []RenderStr
	tabCurrent *RenderStr
    tabNumber  int
}

type tableModel struct {
    table.Model
}

type tabString struct {
    content string
}

func (t tabString) View() string {
    return t.content
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
        case "tab":
            m.tabNumber = (m.tabNumber+1) % len(m.Tabs)
            m.tabCurrent = &m.TabContent[m.tabNumber]
			return m, cmd
        case "shift+tab":
            m.tabNumber = (len(m.Tabs) + m.tabNumber-1) % len(m.Tabs)
            m.tabCurrent = &m.TabContent[m.tabNumber]
			return m, cmd
		case "q", "ctrl+c":
			return m, tea.Quit
		// case "esc":
  //           switch mdl := (*m.tabCurrent).(type) {
  //           case tableModel:
  //               if mdl.Focused() {
  //                   mdl.Blur()
  //               } else {
  //                   mdl.Focus()
  //               }
  //               m.TabContent[m.tabNumber] = mdl
  //       }
        }
        switch mdl := (*m.tabCurrent).(type) {
        case tableModel:
            switch msg.String() {
                case "enter":
                    return m, tea.Batch(
                        tea.Printf("Let's go to %s!", mdl.SelectedRow()[1]), cmd)
                }
            case textInputModel:
                switch msg.String() {
                case "enter":
                if mdl.focused == len(mdl.inputs)-1 {
                    return m, tea.Quit
                } else if mdl.focused == 4 {
                    if mdl.renderList == 0 {
                        mdl.renderList = TIMEZONE
                    } else {
                        mdl.inputs[mdl.focused].SetValue(string(mdl.zoneList.SelectedItem().(item)))
                        mdl.renderList = 0
                    }
                    m.TabContent[m.tabNumber] = mdl
                } else if mdl.focused == 6 {
                    if mdl.renderList == 0 {
                        mdl.renderList = KBD
                    } else {
                        mdl.inputs[mdl.focused].SetValue(string(mdl.kbdList.SelectedItem().(item)))
                        mdl.renderList = 0
                    }
                    m.TabContent[m.tabNumber] = mdl
                }
                case "ctrl+k", "up":
                    mdl, ok := (*m.tabCurrent).(textInputModel); if ok{
                        mdl.prevInput()
                    }
                    m.TabContent[m.tabNumber] = mdl
                    return m, nil
                case "ctrl+j", "down":
                    mdl, ok := (*m.tabCurrent).(textInputModel); if ok{
                        mdl.nextInput()
                    }
                    m.TabContent[m.tabNumber] = mdl
                    return m, nil
                }
            // case listModel:
            //     switch msg.String() {
            //     case "ctrl+c":
            //         mdl.quitting = true
            //         return m, tea.Quit
            //     case "enter":
            //         i, ok := mdl.list.SelectedItem().(item)
            //         if ok {
            //             mdl.choice = string(i)
            //         }
            //         m.TabContent[m.tabNumber] = mdl
            //         return m, tea.Quit
                // }
        }
    }
    switch mdl := (*m.tabCurrent).(type) {
    case tableModel:
        mdl.Model, cmd = mdl.Update(msg)
        m.TabContent[m.tabNumber] = mdl
    case textInputModel:
        if mdl.renderList != 0 {
            if mdl.renderList == TIMEZONE {
                mdl.zoneList, cmd = mdl.zoneList.Update(msg)
            } else if mdl.renderList == KBD {
                mdl.kbdList, cmd = mdl.kbdList.Update(msg)
            }
            m.TabContent[m.tabNumber] = mdl
            return m, cmd
        }
		for i := range mdl.inputs {
			mdl.inputs[i].Blur()
		}
		mdl.inputs[mdl.focused].Focus()
        var cmds []tea.Cmd = make([]tea.Cmd, len(mdl.inputs))
        for i := range mdl.inputs {
            if i != 4 && i != 5 {
                mdl.inputs[i], cmds[i] = mdl.inputs[i].Update(msg)
            }
        }
        m.TabContent[m.tabNumber] = mdl
        return m, tea.Batch(cmds...)
    }

	return m, cmd
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Copy().Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(0, 0).Border(lipgloss.NormalBorder())
)

var borPos int = -1
func (m model) View() string {
    doc := strings.Builder{}

	var renderedTabs []string

	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == m.tabNumber
		if isActive {
			style = activeTabStyle.Copy()
		} else {
			style = inactiveTabStyle.Copy()
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
        renderedTabs = append(renderedTabs, style.Render(t))
	}
        if borPos == -1 {
            borPos = strings.LastIndex(renderedTabs[len(renderedTabs) - 1], "│")
        }
        row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
        doc.WriteString(row)
        doc.WriteString("\n")
		border, _, _, _, _ := windowStyle.GetBorder()
        border.TopLeft = "├"
        windowStyle.Border(border)
        switch v := (*m.tabCurrent).(type) {
        case tableModel:
            t := windowStyle.Render(v.View())
            t = t[:borPos - 8] + "┴─" + t[borPos - 2:]
            doc.WriteString(t)
        case textInputModel:
            t := windowStyle.Render(v.View())
            if borPos - 2 > 1  {
                t = t[:borPos - 8] + "┴─" + t[borPos - 2:]
            }
            doc.WriteString(t)
        default:
            doc.WriteString(windowStyle.Render(v.View()))
        }
        return docStyle.Render(doc.String())
}

func CreatePartitionTable() {


    colLen, lineLen, _ := term.GetSize(0)
    totSum := 0

    for _, i := range TableMaxStrLenArr {
        totSum += i
    }

    var columns []table.Column
    var toAdd int


    toAdd = (colLen - totSum) / (len(TableMaxStrLenArr)) - 5
    columns = []table.Column{
        {Title: "Partition", Width:   TableMaxStrLenArr[0] + toAdd},
        {Title: "Fssize (GB)", Width: TableMaxStrLenArr[1] + toAdd},
        {Title: "Fsavail (GB)", Width:TableMaxStrLenArr[2] + toAdd},
        {Title: "Fstype", Width: TableMaxStrLenArr[3] + toAdd},
        {Title: "Size (GB)", Width:   TableMaxStrLenArr[4] + toAdd},
    }

    if lineLen > len(TablePartitionArr) {
        lineLen = len(TablePartitionArr)
    } else {
        lineLen -= 5
    }

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(TablePartitionArr),
		table.WithFocused(true),
		table.WithHeight(lineLen),
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

    tableM := tableModel{t}


    tabs := []string{"User Info", "Partitions"}
    tabContent := []RenderStr{initialtextInputModel(), tableM}

	m := model{tabs, tabContent, &tabContent[0], 0}
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
