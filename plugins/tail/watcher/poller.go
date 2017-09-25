package watcher

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

type Poller struct {
	Path     string
	Interval time.Duration

	done  chan struct{}
	files map[string]int64
}

func (p *Poller) Close() { close(p.done) }

func (p *Poller) Run(events chan Event) {
	p.files = map[string]int64{}
	p.done = make(chan struct{})

	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			if err := p.poll(events); err != nil {
				log.Printf("[WARN] poller: failed to poll: %v", err)
			}
		}
	}
}

func (p *Poller) poll(events chan Event) error {
	matches, err := filepath.Glob(p.Path)
	if err != nil {
		return err
	}
	for _, path := range matches {
		prevSize, ok := p.files[path]
		if ok {
			s, err := os.Stat(path)
			if err != nil {
				return err
			}
			nextSize := s.Size()
			if nextSize > prevSize {
				events <- Event{Type: WriteEvent, Path: path}
				p.files[path] = nextSize
			}
		} else {
			events <- Event{Type: CreateEvent, Path: path}
			s, err := os.Stat(path)
			if err != nil {
				return err
			}
			p.files[path] = s.Size()
		}
	}

	for path := range p.files {
		if !in(path, matches) {
			delete(p.files, path)
			events <- Event{Type: RemoveEvent, Path: path}
		}
	}

	return nil
}

func in(item string, list []string) bool {
	for _, val := range list {
		if item == val {
			return true
		}
	}
	return false
}
