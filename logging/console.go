package logging

import (
	"fmt"
	"image/color"
	"io"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/rs/zerolog"
)

func newConsoleWriter(out io.Writer, noColor bool) zerolog.ConsoleWriter {
	_, noColorEnv := os.LookupEnv("NO_COLOR")
	noColor = noColor || noColorEnv || colorprofile.Detect(out, os.Environ()) <= colorprofile.NoTTY
	return zerolog.ConsoleWriter{
		Out:     out,
		NoColor: noColor,
		FormatTimestamp: func(i any) string {
			timestamp, _ := i.(string)
			return timestamp
		},
		FormatLevel: func(i any) string {
			level := strings.ToUpper(fmt.Sprintf("%-6s", i))
			if noColor {
				return fmt.Sprintf("| %s |", level)
			}
			var color color.Color
			switch i {
			case "trace":
				color = lipgloss.Color("240") // dark gray
			case "debug":
				color = lipgloss.Color("12") // bright blue
			case "info":
				color = lipgloss.Color("10") // bright green
			case "warn", "warning":
				color = lipgloss.Color("11") // bright yellow
			case "error":
				color = lipgloss.Color("9") // bright red
			case "fatal", "panic":
				color = lipgloss.Color("196") // bold red
			default:
				color = lipgloss.Color("15") // white
			}
			return fmt.Sprintf("| %s |", lipgloss.NewStyle().Foreground(color).Bold(true).Render(level))
		},
	}
}
