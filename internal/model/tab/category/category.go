package category

import (
	"github.com/TypicalAM/goread/internal/backend"
	"github.com/TypicalAM/goread/internal/colorscheme"
	"github.com/TypicalAM/goread/internal/model/simplelist"
	"github.com/TypicalAM/goread/internal/model/tab"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Keymap contains the key bindings for this tab
type Keymap struct {
	CloseTab   key.Binding
	CycleTabs  key.Binding
	SelectFeed key.Binding
	NewFeed    key.Binding
	EditFeed   key.Binding
	DeleteFeed key.Binding
}

// DefaultKeymap contains the default key bindings for this tab
var DefaultKeymap = Keymap{
	CloseTab: key.NewBinding(
		key.WithKeys("c", "ctrl+w"),
		key.WithHelp("c/ctrl+w", "Close tab"),
	),
	CycleTabs: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "Cycle tabs"),
	),
	SelectFeed: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "Open"),
	),
	NewFeed: key.NewBinding(
		key.WithKeys("n", "ctrl+n"),
		key.WithHelp("n/ctrl+n", "New"),
	),
	EditFeed: key.NewBinding(
		key.WithKeys("e", "ctrl+e"),
		key.WithHelp("e/ctrl+e", "Edit"),
	),
	DeleteFeed: key.NewBinding(
		key.WithKeys("d", "ctrl+d"),
		key.WithHelp("d/ctrl+d", "Delete"),
	),
}

// ShortHelp returns the short help for this tab
func (k Keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.CloseTab, k.CycleTabs, k.SelectFeed, k.NewFeed, k.EditFeed, k.DeleteFeed,
	}
}

// FullHelp returns the full help for this tab
func (k Keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.CloseTab, k.CycleTabs, k.SelectFeed, k.NewFeed, k.EditFeed, k.DeleteFeed},
	}
}

// Model contains the state of this tab
type Model struct {
	colors colorscheme.Colorscheme
	width  int
	height int
	title  string
	loaded bool
	list   simplelist.Model
	keymap Keymap
	help   help.Model

	// reader is a function which returns a tea.Cmd which will be executed
	// when the tab is initialized
	reader func(string) tea.Cmd
}

// New creates a new category tab with sensible defaults
func New(colors colorscheme.Colorscheme, width, height int, title string, reader func(string) tea.Cmd) Model {
	help := help.New()
	help.Styles.ShortDesc = lipgloss.NewStyle().Foreground(colors.Text)
	help.Styles.ShortKey = lipgloss.NewStyle().Foreground(colors.Text)
	help.Styles.Ellipsis = lipgloss.NewStyle().Foreground(colors.BgDark)

	return Model{
		colors: colors,
		width:  width,
		height: height,
		title:  title,
		reader: reader,
		help:   help,
		keymap: DefaultKeymap,
	}
}

// Title returns the title of the tab
func (m Model) Title() string {
	return m.title
}

// Type returns the type of the tab
func (m Model) Type() tab.Type {
	return tab.Category
}

// SetSize sets the dimensions of the tab
func (m Model) SetSize(width, height int) tab.Tab {
	m.width = width
	m.height = height
	m.list.SetHeight(m.height)
	return m
}

// ShowHelp shows the help for this tab
func (m Model) ShowHelp() string {
	return m.help.View(m.keymap)
}

// Init initializes the tab
func (m Model) Init() tea.Cmd {
	return m.reader(m.title)
}

// Update updates the variables of the tab
func (m Model) Update(msg tea.Msg) (tab.Tab, tea.Cmd) {
	switch msg := msg.(type) {
	case backend.FetchSuccessMessage:
		// The data fetch was successful
		if !m.loaded {
			m.list = simplelist.New(m.colors, m.title, m.height, false)
			m.loaded = true
		}

		// Set the items of the list
		m.list.SetItems(msg.Items)
		return m, nil

	case tea.KeyMsg:
		// If the tab is not loaded, return
		if !m.loaded {
			return m, nil
		}

		// Handle the key message
		switch {
		case key.Matches(msg, m.keymap.SelectFeed):
			if !m.list.IsEmpty() {
				return m, tab.NewTab(m.list.SelectedItem().FilterValue(), tab.Feed)
			}

			// If the list is empty, return nothing
			return m, nil

		case key.Matches(msg, m.keymap.NewFeed):
			// Add a new feed
			feedPath := []string{m.title}
			return m, backend.NewItem(backend.Feed, true, feedPath, nil)

		case key.Matches(msg, m.keymap.EditFeed):
			// If the list is empty, return nothing
			if m.list.IsEmpty() {
				return m, nil
			}

			// Edit the selected feed
			feedPath := []string{m.title, m.list.SelectedItem().FilterValue()}
			item := m.list.SelectedItem().(simplelist.Item)
			return m, backend.NewItem(
				backend.Feed,
				false,
				feedPath,
				[]string{item.FilterValue(), item.Description()},
			)

		case key.Matches(msg, m.keymap.DeleteFeed):
			if !m.list.IsEmpty() {
				// Delete the selected feed
				delItemName := m.list.SelectedItem().FilterValue()
				itemCount := len(m.list.Items())

				// Move the selection to the next item
				if itemCount == 1 {
					m.list.SetIndex(0)
				} else {
					m.list.SetIndex(m.list.Index() % (itemCount - 1))
				}

				return m, backend.DeleteItem(backend.Feed, delItemName)
			}

		default:
			// Check if we need to open a new feed
			if item, ok := m.list.GetItem(msg.String()); ok {
				return m, tab.NewTab(item.FilterValue(), tab.Feed)
			}
		}
	}

	// Update the list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View returns the view of the tab
func (m Model) View() string {
	// Check if the program is loaded, if not, return a loading message
	if !m.loaded {
		return "Loading..."
	}

	// Return the list view
	return m.list.View()
}
