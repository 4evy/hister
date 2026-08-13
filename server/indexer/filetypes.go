package indexer

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/asciimoo/hister/server/document"
)

type fileTypeHandler interface {
	Match(path string) bool
	Index(*Indexer, *document.Document, []byte) error
}

var fileTypeHandlers = []fileTypeHandler{
	pdfFileType{},
	docxFileType{},
	markdownFileType{},
	orgFileType{},
	plainTextFileType{},
}

func (i *Indexer) indexFileContent(path string, d *document.Document, content []byte) error {
	return fileTypeHandlerForPath(path).Index(i, d, content)
}

func fileTypeHandlerForPath(path string) fileTypeHandler {
	for _, handler := range fileTypeHandlers {
		if handler.Match(path) {
			return handler
		}
	}
	return plainTextFileType{}
}

type plainTextFileType struct{}

func (plainTextFileType) Match(_ string) bool {
	return true
}

func (plainTextFileType) Index(i *Indexer, d *document.Document, content []byte) error {
	if !utf8.Valid(content) {
		return ErrBinaryFile
	}
	if int64(len(content)) > i.maxFileSize {
		return fmt.Errorf("%w: %d bytes", ErrFileTooLarge, int64(len(content)))
	}

	d.Text = string(content)
	return i.AddDocument(d)
}

func hasExtension(path string, exts ...string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains(exts, ext)
}
