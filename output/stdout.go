package output

import (
	"log"
	"os"
	"text/template"

	"github.com/coldog/logship/input"
)

// DefaultFormat is the default template format.
const DefaultFormat = "[{{ .Time }}] {{ .Tag }}: {{ .Data }}\n"

func init() {
	Register("stdout", func() Output { return &Stdout{} })
}

// Stdout provides a simple Stdout output plugin.
type Stdout struct {
	// Format is a go template that the message is formatted using.
	Format string

	tpl *template.Template
}

// Open is a noop.
func (o *Stdout) Open() (err error) {
	if o.Format == "" {
		o.Format = DefaultFormat
	}

	o.tpl, err = template.New("format").Parse(o.Format)
	return err
}

// Run writes each message to Stdout, formatting is not implemented.
func (o *Stdout) Run(in <-chan *input.Message) {
	for msg := range in {
		err := o.tpl.Execute(os.Stdout, msg)
		if err != nil {
			log.Printf("[WARN] stdout: failed to write msg: %v", err)
		}
		input.MessagePool.Put(msg.Clear())
	}
}
