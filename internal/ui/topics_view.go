package ui

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/satmaelstorm/ktan/internal/kafka"
)

type topicsModel struct {
	kc      *kafka.Client
	loc     locale
	topics  []kafka.TopicInfo
	cursor  int
	loading bool
	err     string
	spinner spinner.Model
}

type topicsLoadedMsg struct {
	topics []kafka.TopicInfo
	err    error
}

type topicSelectedMsg struct {
	topic string
}

// formatMessages форматирует число сообщений: -1 — неизвестно,
// большие числа — с суффиксом k/M/G.
func formatMessages(n int64) string {
	if n < 0 {
		return "?"
	}
	switch {
	case n >= 1_000_000_000:
		return trimFloat(float64(n)/1_000_000_000) + "G"
	case n >= 1_000_000:
		return trimFloat(float64(n)/1_000_000) + "M"
	case n >= 10_000:
		return trimFloat(float64(n)/1_000) + "k"
	default:
		return strconv.FormatInt(n, 10)
	}
}

func trimFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', 1, 64)
	if math.IsNaN(f) {
		return "0"
	}
	return strings.TrimSuffix(s, ".0")
}

func newTopicsModel(kc *kafka.Client, loc locale) topicsModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return topicsModel{kc: kc, loc: loc, loading: true, spinner: sp}
}

func (m topicsModel) load() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		topics, err := m.kc.ListTopics(ctx)
		return topicsLoadedMsg{topics: topics, err: err}
	}
}

func (m topicsModel) Update(msg tea.Msg) (topicsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	case topicsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.topics = nil
		} else {
			m.err = ""
			m.topics = msg.topics
			if m.cursor >= len(m.topics) {
				m.cursor = 0
			}
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.topics)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.topics) > 0 {
				topic := m.topics[m.cursor].Name
				return m, func() tea.Msg { return topicSelectedMsg{topic: topic} }
			}
		case "r":
			m.loading = true
			m.err = ""
			return m, tea.Batch(m.load(), m.spinner.Tick)
		}
	}
	return m, nil
}

func (m topicsModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf(m.loc.titleTopics(len(m.topics)), len(m.topics))))
	b.WriteString("\n\n")
	switch {
	case m.loading:
		b.WriteString(m.spinner.View() + " " + m.loc.loadingTopics)
	case m.err != "":
		b.WriteString(errStyle.Render(m.loc.errPrefix + m.err))
	case len(m.topics) == 0:
		b.WriteString(statusStyle.Render(m.loc.noTopics))
	default:
		for i, t := range m.topics {
			line := m.loc.topicLine(t.Name, t.Partitions, formatMessages(t.Messages))
			if i == m.cursor {
				b.WriteString(selectedStyle.Render("> " + line))
			} else {
				b.WriteString("  " + line)
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(statusStyle.Render(m.loc.helpTopics))
	return b.String()
}
