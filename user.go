package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type (
	errMsg error
)

const (
        HOSTNAME = iota
        USERNAME
        PASS
        PASSCONFIRM
        TIMEZONE
        LOCALE
        KBD
    )

    const (
        hotPink  = lipgloss.Color("#FF06B7")
        darkGray = lipgloss.Color("#767676")
    )

    var (
        inputStyle    = lipgloss.NewStyle().Foreground(hotPink)
        continueStyle = lipgloss.NewStyle().Foreground(darkGray)
    )

type textInputModel struct {
	inputs  []textinput.Model
    zoneList list.Model
    kbdList  list.Model
    renderList int
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

func initialtextInputModel() textInputModel {
    cmd, _ := RunCmdOutput("timedatectl list-timezones")
    cmdArr := strings.Split(string(cmd), "\n")
    var items, itemskbd []list.Item

    for _, val := range cmdArr {
        items = append(items, item(val))
    }

    const defaultWidth = 20

    l := list.New(items, itemDelegate{}, defaultWidth, IntMin(len(items), 25))
    l.SetShowStatusBar(false)
    l.SetFilteringEnabled(false)
    l.SetShowTitle(false)
    l.Styles.PaginationStyle = paginationStyle
    l.Styles.HelpStyle = helpStyle

    cmd, _ = RunCmdOutput("timedatectl list-keymaps")
    cmdArr = strings.Split(string(cmd), "\n")

    for _, val := range cmdArr {
        itemskbd = append(itemskbd, item(val))
    }

    lk := list.New(itemskbd, itemDelegate{}, defaultWidth, IntMin(len(itemskbd), 25))
    lk.SetShowStatusBar(false)
    lk.SetFilteringEnabled(false)
    lk.SetShowTitle(false)
    lk.Styles.PaginationStyle = paginationStyle
    lk.Styles.HelpStyle = helpStyle

	var inputs []textinput.Model = make([]textinput.Model, 8)
    cmd, _ = RunCmdOutput("hostnamectl --static")
	inputs[HOSTNAME] = textinput.New()
	inputs[HOSTNAME].CharLimit = 20
	inputs[HOSTNAME].Width = 30
	inputs[HOSTNAME].SetValue(string(cmd))
	inputs[HOSTNAME].Prompt = ""

    cmd, _ = RunCmdOutput("echo -n $SUDO_USER")
	inputs[USERNAME] = textinput.New()
	inputs[USERNAME].SetValue(string(cmd))
	inputs[USERNAME].CharLimit = 13
	inputs[USERNAME].Width = 13
	inputs[USERNAME].Prompt = ""

	inputs[PASS] = textinput.New()
	inputs[PASS].Placeholder = "XXXXXX"
	inputs[PASS].CharLimit = 13
	inputs[PASS].Width = 50
	inputs[PASS].Prompt = ""
    inputs[PASS].EchoMode = textinput.EchoPassword

	inputs[PASSCONFIRM] = textinput.New()
	inputs[PASSCONFIRM].Placeholder = "confirm password"
	inputs[PASSCONFIRM].CharLimit = 50
	inputs[PASSCONFIRM].Width = 50
	inputs[PASSCONFIRM].Prompt = ""
    inputs[PASSCONFIRM].EchoMode = textinput.EchoPassword

    cmd, _ = RunCmdOutput("readlink /etc/localtime | cut -d / -f 5- | tr -d '\n'")
	inputs[TIMEZONE] = textinput.New()
	inputs[TIMEZONE].CharLimit = 50
	inputs[TIMEZONE].Width = 50
	inputs[TIMEZONE].SetValue(string(cmd))
	inputs[TIMEZONE].Prompt = ""

    cmd, _ = RunCmdOutput("localectl status | grep -Po '(?<=Keymap: ).*' ")
	inputs[KBD] = textinput.New()
	inputs[KBD].CharLimit = 50
	inputs[KBD].Width = 50
	inputs[KBD].SetValue(string(cmd))
	inputs[KBD].Prompt = ""

	return textInputModel{
		inputs:  inputs,
        zoneList: l,
        kbdList: lk,
		focused: 0,
        renderList: 0,
		err:     nil,
	}
}

func (m textInputModel) View() string {
    if m.renderList == TIMEZONE {
        return m.zoneList.View()
    } else if m.renderList == KBD {
        return m.kbdList.View()
    }
	return fmt.Sprintf(
		`

%s
 %s

%s
 %s

%s
 %s
 %s

%s
 %s

%s
 %s
%s
`,
        inputStyle.Width(50).Render("Hostname:"),
		m.inputs[HOSTNAME].View(),
        inputStyle.Width(50).Render("Username:"),
		m.inputs[USERNAME].View(),
        inputStyle.Width(50).Render("Password:"),
		m.inputs[PASS].View(),
        m.inputs[PASSCONFIRM].View(),
        inputStyle.Width(50).Render("Timezone:"),
        m.inputs[TIMEZONE].View(),
        inputStyle.Width(50).Render("Keymap:"),
        m.inputs[KBD].View(),
		continueStyle.Render("Continue ->"),
	) + "\n"
}

// nextInput focuses the next input field
func (m *textInputModel) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *textInputModel) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}
