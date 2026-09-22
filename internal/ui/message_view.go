package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/satmaelstorm/ktan/internal/kafka"
)

type messageModel struct {
	loc      locale
	topic    string
	msg      kafka.Message
	viewport viewport.Model
	ready    bool
}

type messageSelectedMsg struct {
	topic string
	msg   kafka.Message
}

type messageBackMsg struct{}

func newMessageModel(topic string, msg kafka.Message, loc locale) messageModel {
	return messageModel{loc: loc, topic: topic, msg: msg}
}

func (m *messageModel) setSize(w, h int) {
	m.viewport.Width = w
	m.viewport.Height = max(h-3, 1)
	m.ready = true
	m.viewport.SetContent(m.render())
}

// prettyJSON форматирует value как JSON с отступами, сохраняя
// исходный порядок ключей. Второе возвращаемое значение — false,
// если value не является валидным JSON.
func prettyJSON(v []byte) (string, bool) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, v, "", "  "); err != nil {
		return "", false
	}
	return buf.String(), true
}

func (m messageModel) render() string {
	var b strings.Builder
	key := "-"
	if len(m.msg.Key) > 0 {
		key = string(m.msg.Key)
	}
	b.WriteString(fmt.Sprintf("%s: %s\n", m.loc.messageTimestamp, m.msg.Timestamp.Format(time.RFC3339Nano)))
	b.WriteString(fmt.Sprintf("%s: %s\n", m.loc.messageKey, key))
	b.WriteString("\n")
	if len(m.msg.Value) == 0 {
		b.WriteString(fmt.Sprintf("%s: -", m.loc.messageValue))
		return b.String()
	}
	if pretty, ok := prettyJSON(m.msg.Value); ok {
		b.WriteString(fmt.Sprintf("%s:\n%s", m.loc.messageValueJSON, pretty))
	} else {
		b.WriteString(fmt.Sprintf("%s:\n%s", m.loc.messageValue, string(m.msg.Value)))
	}
	return b.String()
}

func (m messageModel) Update(msg tea.Msg) (messageModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return messageBackMsg{} }
		}
	}
	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m messageModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.loc.titleMessage(m.topic, m.msg.Partition, m.msg.Offset)))
	b.WriteString("\n\n")
	b.WriteString(m.viewport.View())
	b.WriteString("\n")
	b.WriteString(statusStyle.Render(m.loc.helpMessage))
	return b.String()
}
