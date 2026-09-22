package ui

import (
	"embed"
	"fmt"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

//go:embed locales/*.toml
var localeFS embed.FS

// localeStrings — структура файлов локализации (locales/*.toml).
type localeStrings struct {
	Common struct {
		ErrPrefix string `toml:"err_prefix"`
	} `toml:"common"`
	Topics struct {
		Title    string `toml:"title"`
		Loading  string `toml:"loading"`
		NoTopics string `toml:"no_topics"`
		Line     string `toml:"line"`
		Help     string `toml:"help"`
	} `toml:"topics"`
	Messages struct {
		Title      string `toml:"title"`
		Loading    string `toml:"loading"`
		NoMessages string `toml:"no_messages"`
		Help       string `toml:"help"`
	} `toml:"messages"`
	Message struct {
		Title     string `toml:"title"`
		Timestamp string `toml:"timestamp"`
		Key       string `toml:"key"`
		Value     string `toml:"value"`
		ValueJSON string `toml:"value_json"`
		Help      string `toml:"help"`
	} `toml:"message"`
}

// locale — набор строк интерфейса, загруженный из embedded TOML.
// Плейсхолдеры вида {key} подставляются через t.
type locale struct {
	errPrefix        string
	topicsTitle      string
	loadingTopics    string
	noTopics         string
	topicsLine       string
	helpTopics       string
	messagesTitle    string
	loadingMessages  string
	noMessages       string
	helpMessages     string
	messageTitle     string
	messageTimestamp string
	messageKey       string
	messageValue     string
	messageValueJSON string
	helpMessage      string
}

// newLocale загружает локаль по имени; при отсутствии или ошибке
// парсинга откатывается на готический дефолт.
func newLocale(lang string) locale {
	if l, ok := loadLocale(lang); ok {
		return l
	}
	l, _ := loadLocale("gothic") // embedded-файл, обязан существовать
	return l
}

func loadLocale(lang string) (locale, bool) {
	data, err := localeFS.ReadFile("locales/" + lang + ".toml")
	if err != nil {
		return locale{}, false
	}
	var s localeStrings
	if err := toml.Unmarshal(data, &s); err != nil {
		return locale{}, false
	}
	return locale{
		errPrefix:        s.Common.ErrPrefix,
		topicsTitle:      s.Topics.Title,
		loadingTopics:    s.Topics.Loading,
		noTopics:         s.Topics.NoTopics,
		topicsLine:       s.Topics.Line,
		helpTopics:       s.Topics.Help,
		messagesTitle:    s.Messages.Title,
		loadingMessages:  s.Messages.Loading,
		noMessages:       s.Messages.NoMessages,
		helpMessages:     s.Messages.Help,
		messageTitle:     s.Message.Title,
		messageTimestamp: s.Message.Timestamp,
		messageKey:       s.Message.Key,
		messageValue:     s.Message.Value,
		messageValueJSON: s.Message.ValueJSON,
		helpMessage:      s.Message.Help,
	}, true
}

// t подставляет плейсхолдеры {key} в строку локали.
// Аргументы — последовательные пары "ключ", значение.
func (l locale) t(s string, args ...any) string {
	if len(args) < 2 {
		return s
	}
	pairs := make([]string, 0, len(args))
	for i := 0; i+1 < len(args); i += 2 {
		pairs = append(pairs, "{"+fmt.Sprint(args[i])+"}", fmt.Sprint(args[i+1]))
	}
	return strings.NewReplacer(pairs...).Replace(s)
}

func (l locale) titleTopics(count int) string {
	return l.t(l.topicsTitle, "count", count)
}

func (l locale) topicLine(name string, partitions int, messages string) string {
	return l.t(l.topicsLine, "name", name, "partitions", partitions, "messages", messages)
}

func (l locale) titleMessages(topic string, tail int) string {
	return l.t(l.messagesTitle, "topic", topic, "tail", tail)
}

func (l locale) titleMessage(topic string, partition int32, offset int64) string {
	return l.t(l.messageTitle, "topic", topic, "partition", partition, "offset", offset)
}
