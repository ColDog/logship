package input

import (
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hpcloud/tail"
)

func init() {
	Register("tail", func() Input { return &Tail{} })
}

// Tail implements a multiple file tailer Input plugin.
type Tail struct {
	BaseInput

	Path   string `json:"path"`
	Format string `json:"format"`
	Tag    string `json:"tag"`

	files  map[string]*tail.Tail
	remove chan string
	wg     *sync.WaitGroup
}

// Open implements Open for the Input interface.
func (s *Tail) Open() error {
	if s.Tag == "" {
		s.Tag = "tail"
	}
	s.files = map[string]*tail.Tail{}
	s.wg = &sync.WaitGroup{}
	return s.BaseInput.Open()
}

// Close implements Close for the Input interface.
func (s *Tail) Close() {
	s.BaseInput.Close()
	s.wg.Wait()
}

// Run implements Run for the Input interface. This polls for new files and
// starts a tail process in a new goroutine for every new discovered file.
func (s *Tail) Run(out chan *Message) {
	defer s.Finish()

	s.walk(out)
	for {
		select {
		case <-time.After(300 * time.Millisecond):
			s.walk(out)
		case name := <-s.remove:
			delete(s.files, name)
		case <-s.Done():
			return
		}
	}
}

func (s *Tail) tail(t *tail.Tail, path string, out chan *Message) {
	defer s.wg.Done()

	log.Printf("[DEBU] tail: tailing: %s", path)
	spl := strings.Split(path, "/")
	tag := s.Tag + "." + spl[len(spl)-1]

	for {
		select {
		case line := <-t.Lines:
			if line.Err == tail.ErrStop {
				log.Printf("[DEBU] tail: removing file: %v", path)
				s.remove <- path
				return
			}
			if line.Err != nil {
				log.Printf("[WARN] tail: error tailing %s: %v", path, line.Err)
				continue
			}
			msg := MessagePool.Get().(*Message)
			msg.Time = line.Time
			msg.Data["message"] = line.Text
			msg.Tag = tag
			out <- msg
		case <-s.Done():
			return
		}
	}
}

func (s *Tail) walk(out chan *Message) {
	matches, err := filepath.Glob(s.Path)
	if err != nil {
		log.Printf("[WARN] tail: failed to find matches: %v", err)
		return
	}
	for _, path := range matches {
		if _, ok := s.files[path]; !ok {

			truePath, err := filepath.EvalSymlinks(path)
			if err != nil {
				panic(err)
			}

			t, err := tail.TailFile(truePath, tail.Config{Follow: true})
			if err != nil {
				panic(err)
			}
			s.files[path] = t

			s.wg.Add(1)
			go s.tail(t, path, out)
		}
	}
}
