package output

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/coldog/logship/input"
	"github.com/stretchr/testify/assert"
)

func TestStdout(t *testing.T) {
	ch := make(chan *input.Message)
	std := &Stdout{}
	std.Open()
	go std.Run(ch)
	ch <- &input.Message{
		Time: time.Now(),
		Tag:  "test",
		Data: map[string]string{"message": "test"},
	}
	time.Sleep(10 * time.Millisecond)
}

func TestStdout_Deserialize(t *testing.T) {
	s := &Spec{
		Type:  "stdout",
		Match: "*",
		Spec:  json.RawMessage("{}"),
	}
	o, err := s.Output()
	assert.Nil(t, err)
	assert.IsType(t, &Stdout{}, o)
}
