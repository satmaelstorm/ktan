package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/satmaelstorm/ktan/internal/kafka"
)

const tailLimit = 10

type messagesModel struct {
	kc       *kafka.Client
	loc      locale
	topic    string
	messages []kafka.Message
	loading  bool
	err      string
	spinner  spinner.Model
	viewport viewport.Model
	ready    bool
}

type messagesLoadedMsg struct {
	messages []kafka.Message
	err      error
}

type backMsg struct{}

func newMessagesModel(kc *kafka.Client, topic string, loc locale) messagesModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return messagesModel{kc: kc, loc: loc, topic: topic, loading: true, spinner: sp}
}

func (m messagesModel) load() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		msgs, err := m.kc.TailMessages(ctx, m.topic, tailLimit)
		return messagesLoadedMsg{messages: msgs, err: err}
	}
}

func (m *messagesModel) setSize(w, h int) {
	m.viewport.Width = w
	m.viewport.Height = max(h-3, 1)
	m.ready = true
	m.viewport.SetContent(m.renderMessages())
}

func (m messagesModel) renderMessages() string {
	if len(m.messages) == 0 {
		return statusStyle.Render(m.loc.noMessages)
	}
	lines := make([]string, 0, len(m.messages))
	for _, msg := range m.messages {
		lines = append(lines, m.formatMessage(msg))
	}
	return strings.Join(lines, "\n")
}

func (m messagesModel) formatMessage(msg kafka.Message) string {
	header := fmt.Sprintf("%s p%d #%d", msg.Timestamp.Format("15:04:05.000"), msg.Partition, msg.Offset)
	key := "-"
	if len(msg.Key) > 0 {
		key = truncate(string(msg.Key), 40)
	}
	value := truncate(string(msg.Value), max(m.viewport.Width-60, 20))
	return fmt.Sprintf("%s key=%s │ %s", header, key, value)
}

func (m messagesModel) Update(msg tea.Msg) (messagesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	case messagesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.messages = nil
		} else {
			m.err = ""
			m.messages = msg.messages
		}
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return backMsg{} }
		case "r":
			m.loading = true
			m.err = ""
			return m, tea.Batch(m.load(), m.spinner.Tick)
		}
	}
	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m messagesModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf(m.loc.titleMessages(m.topic, tailLimit), m.topic, tailLimit)))
	b.WriteString("\n\n")
	switch {
	case m.loading:
		b.WriteString(m.spinner.View() + " " + m.loc.loadingMessages)
	case m.err != "":
		b.WriteString(errStyle.Render(m.loc.errPrefix + m.err))
	default:
		b.WriteString(m.viewport.View())
	}
	b.WriteString("\n")
	b.WriteString(statusStyle.Render(m.loc.helpMessages))
	return b.String()
}
