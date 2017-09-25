package tail

import (
	"bufio"
	"log"
	"os"
	"strings"
	"time"

	"github.com/coldog/logship/input"
	"github.com/coldog/logship/plugins/tail/watcher"
)

const (
	PollStrategy = "Poll"

	PollInterval = 10 * time.Millisecond
)

type file struct {
	file    *os.File
	scanner *bufio.Scanner
}

type Tail struct {
	input.BaseInput

	Path     string
	Strategy string

	events  chan watcher.Event
	files   map[string]*file
	watcher watcher.Watcher
}

func (t *Tail) Close() {
	t.BaseInput.Close()
	t.watcher.Close()
	close(t.events)

	for _, f := range t.files {
		f.file.Close()
	}
}

func (t *Tail) Open() error {
	var w watcher.Watcher

	switch t.Strategy {
	case PollStrategy:
		w = &watcher.Poller{Path: t.Path, Interval: PollInterval}
	default:
		w = &watcher.Poller{Path: t.Path, Interval: PollInterval}
	}
	t.watcher = w
	t.events = make(chan watcher.Event)
	t.files = map[string]*file{}
	go w.Run(t.events)
	return t.BaseInput.Open()
}

func (t *Tail) Run(out chan *input.Message) {
	for {
		select {
		case <-t.Done():
			t.Finish()
			return
		case evt := <-t.events:
			switch evt.Type {
			case watcher.CreateEvent:
				t.handleCreateEvent(evt, out)
			case watcher.RemoveEvent:
				t.handleRemoveEvent(evt, out)
			case watcher.WriteEvent:
				t.handleWriteEvent(evt, out)
			}
		}
	}
}

func (t *Tail) handleRemoveEvent(evt watcher.Event, out chan *input.Message) {
	delete(t.files, evt.Path)
}

func (t *Tail) handleWriteEvent(evt watcher.Event, out chan *input.Message) {
	f, ok := t.files[evt.Path]
	if !ok {
		log.Printf("[WARN] tail: file not exist: %s", evt.Path)
		return
	}
	for f.scanner.Scan() {
		msg := input.MessagePool.Get().(*input.Message)
		msg.Time = time.Now()
		msg.Data["message"] = strings.TrimSpace(f.scanner.Text())
		msg.Tag = evt.Path

		out <- msg
	}
}

func (t *Tail) handleCreateEvent(evt watcher.Event, out chan *input.Message) {
	f, err := os.Open(evt.Path)
	if err != nil {
		log.Printf("[WARN] tail: failed to create a file: %v", err)
		return
	}
	t.files[evt.Path] = &file{
		file:    f,
		scanner: bufio.NewScanner(f),
	}
}
