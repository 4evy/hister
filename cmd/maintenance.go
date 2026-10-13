package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/asciimoo/hister/client"

	"github.com/spf13/cobra"
)

var reindexCmd = &cobra.Command{
	Use:   "reindex",
	Short: "Rebuild the search index",
	Long:  `Rebuild the search index from all stored documents using the running server.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		skipSensitive := false
		if b, err := cmd.Flags().GetBool("exclude-sensitive"); err == nil {
			skipSensitive = b
		}
		c := newClient(client.WithTimeout(0))
		return runMaintenance(cmd, "Reindex", "Reindexing", func() error {
			return c.Reindex(skipSensitive, cfg.Indexer.DetectLanguages)
		})
	},
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove stale local documents and orphaned data",
	Long:  `Remove indexed local documents that no longer match configured directories, and remove HTML and favicon files that are no longer referenced by any document in the index`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		c := newClient(client.WithTimeout(0))
		var result client.CleanupResult
		if err := runMaintenance(cmd, "Cleanup", "Cleaning up", func() error {
			var err error
			result, err = c.Cleanup()
			return err
		}); err != nil {
			return err
		}
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "Checked %d indexed local document(s)\nSkipped %d indexed local document(s)\nRemoved %d stale local document(s)\nRemoved %d orphaned HTML file(s)\nRemoved %d orphaned favicon file(s)\n",
			result.LocalDocumentsChecked, result.LocalDocumentsSkipped, result.LocalDocumentsRemoved, result.HTMLRemoved, result.FaviconRemoved)
		return err
	},
}

func runMaintenance(cmd *cobra.Command, name, activity string, operation func() error) error {
	cmd.SilenceUsage = true
	out := cmd.ErrOrStderr()
	stop := startMaintenanceProgress(out, activity, isTerminalWriter(out))
	err := operation()
	elapsed := stop().Round(time.Second)
	if err != nil {
		wrapped := fmt.Errorf("%s error: %w", name, err)
		if isConnectionError(err) {
			return fmt.Errorf("%w\n  Make sure the Hister server is running before executing %s", wrapped, cmd.Name())
		}
		return wrapped
	}
	_, err = fmt.Fprintf(out, "%s completed in %s.\n", name, elapsed)
	return err
}

// startMaintenanceProgress returns a function that stops and joins the timer
// before callers write their final status or results.
func startMaintenanceProgress(out io.Writer, activity string, interactive bool) func() time.Duration {
	started := time.Now()
	if !interactive {
		_, _ = fmt.Fprintf(out, "%s...\n", activity)
		return func() time.Duration { return time.Since(started) }
	}
	_, _ = fmt.Fprintf(out, "%s... 0s elapsed", activity)
	stop, done := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_, _ = fmt.Fprintf(out, "\r%s... %s elapsed", activity, time.Since(started).Truncate(time.Second))
			case <-stop:
				_, _ = fmt.Fprintln(out)
				return
			}
		}
	}()
	return func() time.Duration {
		close(stop)
		<-done
		return time.Since(started)
	}
}
