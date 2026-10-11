package logging

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestConsoleWriterSuppressesCSIForNonTerminalOutput(t *testing.T) {
	oldNoColor, hadNoColor := os.LookupEnv("NO_COLOR")
	if err := os.Unsetenv("NO_COLOR"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if hadNoColor {
			_ = os.Setenv("NO_COLOR", oldNoColor)
		} else {
			_ = os.Unsetenv("NO_COLOR")
		}
	})
	t.Setenv("CLICOLOR_FORCE", "")
	var output bytes.Buffer
	w := newConsoleWriter(&output, false)
	if !w.NoColor {
		t.Fatal("non-terminal log writer retained color")
	}
	if _, err := w.Write([]byte(`{"level":"error","time":"now","message":"broken"}` + "\n")); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "\x1b[") {
		t.Fatalf("non-terminal log output contains CSI: %q", output.String())
	}
}
