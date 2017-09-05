package input

import (
	"bufio"
	"io"
	"os"
	"time"
)

var stdin io.Reader = os.Stdin

func init() {
	Register("stdin", func() Input { return &Stdin{} })
}

// Stdin implements Input for receiving data through stdin.
type Stdin struct {
	BaseInput
	Tag string
}

// Run starts the scanner reading from stdin.
func (s *Stdin) Run(out chan *Message) {
	defer s.Finish()

	reader := bufio.NewScanner(stdin)
	for reader.Scan() {
		select {
		case <-s.Done():
			return
		default:
		}

		msg := MessagePool.Get().(*Message)

		msg.Data["message"] = reader.Text()
		msg.Time = time.Now()
		msg.Tag = s.Tag

		out <- msg
	}
}
