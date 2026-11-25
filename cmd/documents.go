package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/asciimoo/nogon/client"
	"github.com/asciimoo/nogon/files"
	"github.com/asciimoo/nogon/server/document"
	"github.com/asciimoo/nogon/server/model"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add content to Nogon",
}

var addDocumentCmd = &cobra.Command{
	Use:   "document URL",
	Short: "Add or update a web document using text you supply",
	Long: `Supply a title, plain text, or both for an HTTP or HTTPS URL.
Use --text-file - to read standard input. Adding the same URL again replaces
its title and text for the current owner. The page itself is never fetched.`,
	Example: `  nogon add document https://example.com/notes --title "My notes" --text "Meeting agenda"
  gh-stars info asciimoo/nogon | nogon add document https://github.com/asciimoo/nogon --text-file - --label github`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawURL := args[0]
		u, err := url.Parse(rawURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || strings.ContainsFunc(rawURL, unicode.IsSpace) {
			return errors.New("expected an absolute HTTP or HTTPS document URL")
		}
		title, _ := cmd.Flags().GetString("title")
		text, _ := cmd.Flags().GetString("text")
		label, _ := cmd.Flags().GetString("label")
		c := newClient(documentSubmissionClientOptions(cmd)...)
		if cmd.Flags().Changed("text-file") {
			path, _ := cmd.Flags().GetString("text-file")
			input := cmd.InOrStdin()
			if path != "-" {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer func() { _ = file.Close() }()
				input = file
			}
			body, err := files.ReadAllBounded(input, c.MaxBatchBodyBytes())
			if err != nil {
				return fmt.Errorf("read document text: %w", err)
			}
			text = string(body)
		}
		if !utf8.ValidString(text) || !utf8.ValidString(title) || !utf8.ValidString(label) {
			return errors.New("document title, text, and label must be valid UTF-8")
		}
		if strings.TrimSpace(title) == "" && strings.TrimSpace(text) == "" {
			return errors.New("provide a nonblank title or text using --title, --text, or --text-file")
		}
		d := &client.Document{Type: client.Web, URL: rawURL, Title: title, Text: text, Label: label}
		return c.AddDocumentJSONContext(cmd.Context(), d)
	},
}

var listURLsCmd = &cobra.Command{
	Use:   "list-urls",
	Short: "List indexed URLs",
	Long:  `List all indexed URLs by fetching them from the running server`,
	Run: func(cmd *cobra.Command, _ []string) {
		offline, _ := cmd.Flags().GetBool("offline")
		if offline {
			idx, err := initIndex(model.ReadOnly)
			if err != nil {
				exit(1, err.Error())
			}
			defer idx.Close()
			idx.Iterate(func(doc *document.Document) {
				fmt.Println(doc.URL)
			})
			return
		}
		c := newClient(client.WithTimeout(0))
		pageKey := ""
		for {
			res, err := c.Search(&client.SearchQuery{Text: "*", PageKey: pageKey, Sort: "domain"})
			if err != nil {
				exit(1, "Failed to fetch URLs: "+err.Error())
			}
			for _, doc := range res.Documents {
				fmt.Println(doc.URL)
			}
			if res.PageKey == "" || len(res.Documents) == 0 {
				break
			}
			pageKey = res.PageKey
		}
	},
}

var listFilesCmd = &cobra.Command{
	Use:   "list-files",
	Short: "List files matching configured directory watch rules",
	Long: `List all files that match the configured directory watch patterns.

This lists indexing candidates, including files rejected by sensitive content,
size, or format checks. It does not report whether a file is indexed.
Rejected files remain watched and are checked again when they change.`,
	Run: func(cmd *cobra.Command, _ []string) {
		if len(cfg.Indexer.Directories) == 0 {
			exit(1, "No directories configured for watching")
		}
		relative, _ := cmd.Flags().GetBool("relative")
		for _, dir := range cfg.Indexer.Directories {
			expanded := files.ExpandHome(dir.Path)
			err := filepath.WalkDir(expanded, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					log.Warn().Err(err).Str("path", path).Msg("Error accessing path")
					return nil
				}
				if d.IsDir() {
					if path != expanded && files.ShouldSkipDir(d.Name(), dir.Excludes, dir.IncludeHidden) {
						return filepath.SkipDir
					}
					return nil
				}
				if dir.IsMatching(d.Name()) {
					outputPath := path
					if relative {
						var relErr error
						outputPath, relErr = filepath.Rel(expanded, path)
						if relErr != nil {
							log.Warn().Err(relErr).Str("path", path).Msg("Failed to make path relative")
							outputPath = path
						}
					}
					fmt.Println(outputPath)
				}
				return nil
			})
			if err != nil {
				log.Error().Err(err).Str("directory", expanded).Msg("Failed to walk directory")
			}
		}
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete QUERY",
	Short: "Remove documents from the index",
	Long: `Remove documents from the index using the search query language.

The QUERY syntax is the same as the search queries.

Examples:
  nogon delete "url:https://example.com/page"
  nogon delete "url:file:///home/user/file.pdf"
  nogon delete "domain:example.com"
  nogon delete "language:en domain:example.com"
  nogon delete --dry "updated:>90d"

Non-admin users are restricted to their own documents by the server.
Dry runs and confirmation counts use the server's deletion scope. Administrators
can delete across users; use user_id:ID to restrict the query to one owner.
Verbose output lists URLs visible in search, which may differ from the documents
selected for deletion.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		dry, _ := cmd.Flags().GetBool("dry")
		verbose, _ := cmd.Flags().GetBool("verbose")
		yes, _ := cmd.Flags().GetBool("yes")
		total, err := c.CountDocumentsForDeletion(args[0])
		if err != nil {
			return fmt.Errorf("failed to inspect documents for deletion: %w", err)
		}
		if verbose {
			cmd.PrintErrln("URLs visible in search may differ from the documents selected for deletion; the count uses deletion scope.")
			var pageKey string
			for {
				res, err := c.Search(&client.SearchQuery{Text: args[0], PageKey: pageKey, Sort: "domain"})
				if err != nil {
					return fmt.Errorf("failed to list URLs visible in search: %w", err)
				}
				for _, doc := range res.Documents {
					fmt.Println(doc.URL)
				}
				if res.PageKey == "" || len(res.Documents) == 0 {
					break
				}
				pageKey = res.PageKey
			}
		}
		if dry {
			fmt.Printf("%d document(s) would be deleted\n", total)
			return nil
		}
		if total == 0 {
			fmt.Println("No documents matched the query")
			return nil
		}
		if !yes && !yesNoPrompt(fmt.Sprintf("Delete %d document(s)?", total), false) {
			fmt.Println("Deletion cancelled")
			return nil
		}
		if err := c.DeleteDocuments(args[0]); err != nil {
			return fmt.Errorf("failed to delete: %w", err)
		}
		return nil
	},
}
