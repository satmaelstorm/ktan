package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/satmaelstorm/ktan/internal/kafka"
)

type screen int

const (
	screenTopics screen = iota
	screenMessages
)

type App struct {
	kc       *kafka.Client
	screen   screen
	width    int
	height   int
	topics   topicsModel
	messages messagesModel
}

func New(kc *kafka.Client) App {
	return App{
		kc:     kc,
		screen: screenTopics,
		topics: newTopicsModel(kc),
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(a.topics.load(), a.topics.spinner.Tick)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		a.messages.setSize(a.width, a.height)
		return a, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return a, tea.Quit
		}
	case topicSelectedMsg:
		a.screen = screenMessages
		a.messages = newMessagesModel(a.kc, msg.topic)
		a.messages.setSize(a.width, a.height)
		return a, tea.Batch(a.messages.load(), a.messages.spinner.Tick)
	case backMsg:
		a.screen = screenTopics
		return a, nil
	}

	var cmd tea.Cmd
	switch a.screen {
	case screenMessages:
		a.messages, cmd = a.messages.Update(msg)
	default:
		a.topics, cmd = a.topics.Update(msg)
	}
	return a, cmd
}

func (a App) View() string {
	if a.screen == screenMessages {
		return a.messages.View()
	}
	return a.topics.View()
}
