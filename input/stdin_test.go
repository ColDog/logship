package input

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStdin(t *testing.T) {
	stdin = bytes.NewBufferString("test\ntest\ntest\n")
	ch := make(chan *Message)
	go (&Stdin{}).Run(ch)
	m := <-ch
	assert.NotNil(t, m)
}

func TestStdin_Deserialize(t *testing.T) {
	s := &Spec{
		Type: "stdin",
		Spec: json.RawMessage("{}"),
	}
	o, err := s.Input()
	assert.Nil(t, err)
	assert.IsType(t, &Stdin{}, o)
}
