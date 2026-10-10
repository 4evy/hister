package document

import (
	"encoding/json"
	"os"
	"testing"
)

// These authored examples are regression fixtures, not a calibrated accuracy
// benchmark. Keep the scripts, lengths, and ambiguous inputs varied.
func TestLanguageDetectionCorpus(t *testing.T) {
	data, err := os.ReadFile("testdata/language_corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus []struct{ Name, Language, Text string }
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []struct {
		name string
		low  bool
	}{{"low", true}, {"high", false}} {
		t.Run(mode.name, func(t *testing.T) {
			d := NewLanguageDetectorFor(nil, mode.low)
			correct, abstained := 0, 0
			for _, sample := range corpus {
				got := d.DetectLanguage(sample.Text)
				if got == sample.Language {
					correct++
				}
				if got == UnknownLanguage {
					abstained++
				}
				if got != sample.Language {
					t.Errorf("%s: got %s, want %s", sample.Name, got, sample.Language)
				}
			}
			t.Logf("correct %d/%d, abstained %d/%d", correct, len(corpus), abstained, len(corpus))
		})
	}
}
