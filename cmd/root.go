package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/satmaelstorm/ktan/internal/kafka"
	"github.com/satmaelstorm/ktan/internal/ui"
)

var (
	brokers string
	lang    string
)

var rootCmd = &cobra.Command{
	Use:   "ktan",
	Short: "Kafka Topic ANalyzer — TUI-инструмент для просмотра Kafka",
	RunE: func(cmd *cobra.Command, args []string) error {
		if lang == "" {
			lang = os.Getenv("KTAN_LANG")
		}

		kc, err := kafka.New(strings.Split(brokers, ","))
		if err != nil {
			return fmt.Errorf("kafka connect: %w", err)
		}
		defer kc.Close()

		p := tea.NewProgram(ui.New(kc, lang), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("tui: %w", err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&brokers, "brokers", "b", "localhost:9092", "bootstrap servers через запятую")
	rootCmd.PersistentFlags().StringVarP(&lang, "lang", "l", "", "interface language: en (default: gothic machine-cult, overrides KTAN_LANG)")
}
