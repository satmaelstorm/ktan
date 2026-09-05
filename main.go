package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/satmaelstorm/ktan/internal/kafka"
	"github.com/satmaelstorm/ktan/internal/ui"
)

func main() {
	brokers := flag.String("brokers", "localhost:9092", "bootstrap servers через запятую")
	flag.Parse()

	kc, err := kafka.New(strings.Split(*brokers, ","))
	if err != nil {
		fmt.Fprintln(os.Stderr, "kafka connect:", err)
		os.Exit(1)
	}
	defer kc.Close()

	p := tea.NewProgram(ui.New(kc), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tui:", err)
		os.Exit(1)
	}
}
