package metrics

import (
	"testing"
	"time"
)

type blockingGaugeSource struct{ entered, release chan struct{} }

func (s *blockingGaugeSource) Total() uint64   { close(s.entered); <-s.release; return 0 }
func (s *blockingGaugeSource) DataDir() string { return "" }

func TestStopWaitsForGaugeRefresh(t *testing.T) {
	source := &blockingGaugeSource{make(chan struct{}), make(chan struct{})}
	m := New(t.Context(), source)
	<-source.entered
	stopped := make(chan struct{})
	go func() { m.Stop(); close(stopped) }()
	select {
	case <-stopped:
		close(source.release)
		t.Fatal("Stop returned while metrics still accessed the indexer")
	case <-time.After(20 * time.Millisecond):
	}
	close(source.release)
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("Stop did not finish")
	}
}
