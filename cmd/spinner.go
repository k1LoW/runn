package cmd

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

var _ tea.Model = (*spinnerModel)(nil)

type spinnerModel struct {
	spinner spinner.Model
}

func newSpinnerModel() *spinnerModel {
	s := spinner.New()
	return &spinnerModel{
		spinner: s,
	}
}

func (m *spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m *spinnerModel) View() string {
	return m.spinner.View()
}
