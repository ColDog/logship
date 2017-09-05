package filter

import (
	"bytes"
	"log"
	"sync"
	"text/template"

	"github.com/coldog/logship/input"
)

func init() {
	Register("transformer", func() Filter { return &Transformer{} })
}

// Transformer transforms records.
type Transformer struct {
	// Record is a map from the new key to a go template to encode the next value.
	Record map[string]string

	// Tag is a go template for the new tag value, null means to not change the
	// tag.
	Tag string

	tagTpl *template.Template
	tpls   map[string]*template.Template
}

// Open initializes the record templates.
func (t *Transformer) Open() (err error) {
	t.tpls = map[string]*template.Template{}
	for key, tpl := range t.Record {
		t.tpls[key], err = template.New("transformer").Parse(tpl)
		if err != nil {
			return err
		}
	}

	if t.Tag != "" {
		t.tagTpl, err = template.New("transformer").Parse(t.Tag)
		if err != nil {
			return err
		}
	}
	return nil
}

var bufferPool = &sync.Pool{
	New: func() interface{} { return bytes.NewBuffer(make([]byte, 0, 1024)) },
}

// Filter will apply the record templates
func (t *Transformer) Filter(m *input.Message) bool {
	buf := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buf)

	if t.tagTpl != nil {
		err := t.tagTpl.Execute(buf, m)
		if err != nil {
			log.Printf("[WARN] transformer: failed to execute tpl: %v", err)
			return false
		}

		m.Tag = buf.String()
		buf.Reset()
	}

	for k, tpl := range t.tpls {
		err := tpl.Execute(buf, m)
		if err != nil {
			log.Printf("[WARN] transformer: failed to execute tpl: %v", err)
			return false
		}
		m.Data[k] = buf.String()
		buf.Reset()
	}
	for k := range m.Data {
		_, ok := t.tpls[k]
		if !ok {
			delete(m.Data, k)
		}
	}
	return true
}
