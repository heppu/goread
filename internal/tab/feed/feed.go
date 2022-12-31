package feed

import (
	"fmt"
	"strings"

	"github.com/TypicalAM/goread/internal/backend"
	simpleList "github.com/TypicalAM/goread/internal/list"
	"github.com/TypicalAM/goread/internal/style"
	"github.com/TypicalAM/goread/internal/tab"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SelectedPane int

const (
	articlesList SelectedPane = iota
	articlesPreview
)

// RssFeedTab is a tab that displays a list of RSS feeds
type RssFeedTab struct {
	title  string
	index  int
	loaded bool

	loadingSpinner spinner.Model
	list           list.Model
	isViewportOpen bool
	viewedItem     simpleList.ListItem
	content        string
	viewport       viewport.Model
	selected       SelectedPane

	readerFunc func(string) tea.Cmd
}

// New creates a new RssFeedTab with sensible defautls
func New(title string, index int, readerFunc func(string) tea.Cmd) RssFeedTab {
	// Create a spinner for loading the data
	spin := spinner.New()
	spin.Spinner = spinner.Points
	spin.Style = lipgloss.NewStyle().Foreground(style.GlobalColorscheme.Color1)

	return RssFeedTab{
		loadingSpinner: spin,
		title:          title,
		index:          index,
		readerFunc:     readerFunc,
	}
}

// Return the title of the tab
func (r RssFeedTab) Title() string {
	return r.title
}

// Return the index of the tab
func (r RssFeedTab) Index() int {
	return r.index
}

// Show if the tab is loaded
func (r RssFeedTab) Loaded() bool {
	return r.loaded
}

// Initialize the tab
func (r RssFeedTab) Init() tea.Cmd {
	return tea.Batch(r.readerFunc(r.title), r.loadingSpinner.Tick)
}

// Update the tab
func (r RssFeedTab) Update(msg tea.Msg) (tab.Tab, tea.Cmd) {
	switch msg := msg.(type) {
	case backend.FetchSuccessMessage:
		if !r.loaded && style.WindowWidth > 0 && style.WindowHeight > 0 {
			// Set the width and the height of the components
			listWidth := style.WindowWidth / 4
			viewportWidth := style.WindowWidth - listWidth - 4 // 4 is the padding
			r.list = list.New(msg.Items, list.NewDefaultDelegate(), listWidth, style.WindowHeight-5)

			// Set some attributes for the list along with the newly fetched items
			r.list.SetShowHelp(false)
			r.list.SetShowTitle(false)
			r.list.SetShowStatusBar(false)

			// Initialize the viewport
			r.viewport = viewport.New(viewportWidth, style.WindowHeight-5)

			// We are locked and loaded
			r.loaded = true
			return r, nil
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// If the tab is not loaded, return
			if !r.loaded {
				return r, nil
			}

			// Set the content of the viewport on the selected item
			r.viewport.SetContent(r.list.SelectedItem().(simpleList.ListItem).GetContent())

			// Set the view as open if it isn't
			if !r.isViewportOpen {
				r.isViewportOpen = true
			}

			// We don't need to update the list or the viewport
			return r, nil
		case "left", "right":
			// If the viewport isn't open, don't do anything
			if !r.isViewportOpen {
				return r, nil
			}

			// If the viewport is open, switch the selected pane
			if r.selected == articlesPreview {
				r.selected = articlesList
			} else {
				r.selected = articlesPreview
			}

			// We don't need to update the list or the viewport
			return r, nil
		}
	default:
		// If the model is not loaded, update the loading spinner
		if !r.loaded {
			var cmd tea.Cmd
			r.loadingSpinner, cmd = r.loadingSpinner.Update(msg)
			return r, cmd
		}
	}

	// Update the selected item from the pane
	var cmd tea.Cmd
	if r.selected == articlesList {
		// Prevent the list from updating if we are not loaded yet
		if r.loaded {
			r.list, cmd = r.list.Update(msg)
		}
	} else {
		r.viewport, cmd = r.viewport.Update(msg)
	}

	return r, cmd
}

func (r RssFeedTab) View() string {
	if !r.loaded {
		loadingMessage := lipgloss.NewStyle().
			MarginLeft(3).
			MarginTop(1).
			Render(fmt.Sprintf("%s Loading feed %s", r.loadingSpinner.View(), r.title))

		return loadingMessage + strings.Repeat("\n", style.WindowHeight-3-lipgloss.Height(loadingMessage))
	}

	rssList := r.list.View()
	rssViewport := r.viewport.View()

	// If the view is not open show just the rss list
	if !r.isViewportOpen {
		return style.FocusedStyle.Render(rssList)
	}

	// If the viewport is open and the list is selected, highlight the list
	if r.selected == articlesList {
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			style.FocusedStyle.Render(rssList),
			style.ColumnStyle.Render(rssViewport),
		)
	}

	// Highlight the viewport
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		style.ColumnStyle.Render(rssList),
		style.FocusedStyle.Render(rssViewport),
	)
}

// Return the type of the tab
func (r RssFeedTab) Type() tab.TabType {
	return tab.Feed
}

// Set the index of the tab
func (r RssFeedTab) SetIndex(index int) tab.Tab {
	r.index = index
	return r
}
