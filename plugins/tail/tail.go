package input

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/coldog/logship/input"
	"github.com/hpcloud/tail"
)

func init() {
	input.Register("tail", func() input.Input { return &Tail{} })
}

// Tail implements a multiple file tailer Input plugin.
type Tail struct {
	input.BaseInput

	Path   string `json:"path"`
	Format string `json:"format"`
	Tag    string `json:"tag"`

	files  map[string]bool
	remove chan string
}

// Open implements Open for the Input interface.
func (s *Tail) Open() error {
	if s.Tag == "" {
		s.Tag = "tail"
	}
	s.files = map[string]bool{}
	s.remove = make(chan string, 100)
	return s.BaseInput.Open()
}

// Run implements Run for the Input interface. This polls for new files and
// starts a tail process in a new goroutine for every new discovered file.
func (s *Tail) Run(out chan *input.Message) {
	s.walk(out)
MAIN:
	for {
		select {
		case <-time.After(300 * time.Millisecond):
			s.walk(out)
		case name := <-s.remove:
			delete(s.files, name)
		case <-s.Done():
			break MAIN
		}
	}

	defer s.Finish()

	if len(s.files) == 0 {
		return
	}

	log.Println("[DEBU] tail: closing")
	for name := range s.remove {
		log.Printf("[DEBU] tail: closing: %s %+v", name, s.files)
		delete(s.files, name)
		if len(s.files) == 0 {
			return
		}
	}
}

func (s *Tail) tail(path string, out chan *input.Message) {
	spl := strings.Split(path, "/")
	tag := s.Tag + "." + spl[len(spl)-1]
	log.Printf("[DEBU] tail: tailing: %s", tag)

	t, err := tail.TailFile(path, tail.Config{
		Follow: true,
		Logger: tail.DiscardingLogger,
	})
	if err != nil {
		s.remove <- path
		log.Printf("[WARN] tail: error opening tail %s: %v", tag, err)
		return
	}

	defer t.Cleanup()

	for {
		select {
		case line, ok := <-t.Lines:
			if !ok {
				s.remove <- path
				log.Printf("[DEBU] tail: closed: %s", tag)
				return
			}
			if line.Err == tail.ErrStop {
				s.remove <- path
				log.Printf("[DEBU] tail: removing file: %s", tag)
				return
			}
			if line.Err != nil {
				log.Printf("[WARN] tail: error tailing %s: %v", tag, line.Err)
				continue
			}

			// Parse the message.
			msg := input.MessagePool.Get().(*input.Message)
			msg.Time = line.Time
			msg.Data["message"] = line.Text
			msg.Tag = tag

			out <- msg
		case <-s.Done():
			s.remove <- path
			log.Printf("[DEBU] tail: exiting: %s", tag)
			return
		}
	}
}

func (s *Tail) walk(out chan *input.Message) {
	matches, err := filepath.Glob(s.Path)
	if err != nil {
		log.Printf("[WARN] tail: failed to find matches: %v", err)
		return
	}
	for _, path := range matches {
		if _, ok := s.files[path]; !ok {
			s.files[path] = true
			go s.tail(path, out)
		}
	}
}
