package ui

// locale — набор строк интерфейса. По умолчанию — готический стиль
// техножрецов (Adeptus Mechanicus), "en" — нейтральный английский.
type locale struct {
	titleTopics     func(count int) string
	loadingTopics   string
	errPrefix       string
	noTopics        string
	topicLine       func(name string, partitions int, messages int64) string
	helpTopics      string
	titleMessages   func(topic string, tail int) string
	loadingMessages string
	noMessages      string
	helpMessages    string
}

var (
	localeGothic = locale{
		titleTopics:   func(count int) string { return "ktan — Data-canticles of the Omnissiah (%d)" },
		loadingTopics: "Invoking the archives... ++machine-spirit stirring++",
		errPrefix:     "ERROR: the machine-spirit rebels — ",
		noTopics:      "The archives are silent. No data-slates found.",
		topicLine: func(name string, partitions int, messages int64) string {
			return "%s (%d sacred conduits, %s data-canticles)"
		},
		helpTopics:      "↑/↓ commune • enter invoke • r re-rite • ctrl+c sever the link",
		titleMessages:   func(topic string, tail int) string { return "ktan — Canticles of %q (last %d)" },
		loadingMessages: "Awakening the machine-spirit...",
		noMessages:      "Silence. No data-canticles recorded.",
		helpMessages:    "↑/↓ scroll the litany • r re-rite • esc retreat • ctrl+c sever",
	}
	localeEn = locale{
		titleTopics:     func(count int) string { return "ktan — Kafka topics (%d)" },
		loadingTopics:   "loading topics...",
		errPrefix:       "error: ",
		noTopics:        "no topics found",
		topicLine:       func(name string, partitions int, messages int64) string { return "%s (%d partitions, %s messages)" },
		helpTopics:      "↑/↓ move • enter open • r refresh • ctrl+c quit",
		titleMessages:   func(topic string, tail int) string { return "ktan — topic %q (tail %d)" },
		loadingMessages: "loading messages...",
		noMessages:      "no messages",
		helpMessages:    "↑/↓ scroll • r refresh • esc back • ctrl+c quit",
	}
)

func newLocale(lang string) locale {
	if lang == "en" {
		return localeEn
	}
	return localeGothic
}
